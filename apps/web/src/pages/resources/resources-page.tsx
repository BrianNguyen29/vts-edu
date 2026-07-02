import { useEffect, useMemo, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useAuth } from '@/app/providers/auth-provider';
import { useClasses } from '@/shared/api/academics-queries';
import {
  useArchiveResourceMutation,
  useCreateResourceMutation,
  useDownloadResourceMutation,
  usePublishResourceMutation,
  useResourceFilesQuery,
  useResourcesQuery,
} from '@/shared/api/resources-queries';
import {
  uploadResourceFilesWithProgress,
  type FileUploadProgress,
  type ResourceFile,
} from '@/shared/api/resources';
import { loadRuntimeConfig } from '@/shared/config/runtime-config';
import { getAccessToken } from '@/shared/auth/auth-session-store';
import { ErrorState } from '@/shared/components/error-state';
import {
  getApiErrorDetails,
  formatFriendlyError,
} from '@/shared/api/api-error';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

function isManager(roles: string[]): boolean {
  return roles.includes('teacher') || roles.includes('admin');
}

const PREVIEWABLE_TYPES = new Set([
  'image/png',
  'image/jpeg',
  'image/gif',
  'image/webp',
  'image/svg+xml',
  'application/pdf',
  'text/plain',
  'text/csv',
  'text/markdown',
]);

function isPreviewable(contentType: string): boolean {
  return PREVIEWABLE_TYPES.has(contentType);
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function InlinePreviewModal({
  resourceId,
  file,
  onClose,
}: {
  resourceId: string;
  file: ResourceFile;
  onClose: () => void;
}) {
  const [objectUrl, setObjectUrl] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const dialogRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    let cancelled = false;
    let url: string | null = null;
    (async () => {
      try {
        const [config, token] = await Promise.all([
          loadRuntimeConfig(),
          Promise.resolve(getAccessToken()),
        ]);
        const u = new URL(
          `${config.apiBaseUrl}/resources/${resourceId}/download?file_id=${file.id}&disposition=inline`
        );
        const headers = new Headers();
        if (token) headers.set('Authorization', `Bearer ${token}`);
        const r = await fetch(u.toString(), { headers, credentials: 'include' });
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        const blob = await r.blob();
        if (cancelled) return;
        url = URL.createObjectURL(blob);
        setObjectUrl(url);
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : String(e));
        }
      }
    })();
    return () => {
      cancelled = true;
      if (url) URL.revokeObjectURL(url);
    };
  }, [resourceId, file.id]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [onClose]);

  return (
    <div
      className="preview-modal-backdrop"
      role="dialog"
      aria-modal="true"
      aria-labelledby="preview-title"
      onClick={onClose}
      data-testid="preview-modal"
    >
      <div
        className="preview-modal"
        ref={dialogRef}
        onClick={(e) => e.stopPropagation()}
      >
        <header className="preview-modal-header">
          <h2 id="preview-title">{file.original_name}</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Đóng cửa sổ xem trước"
            data-testid="preview-modal-close"
          >
            Đóng
          </button>
        </header>
        <div className="preview-modal-body">
          {error && <p className="form-error" role="alert">{error}</p>}
          {!error && !objectUrl && (
            <p role="status" aria-live="polite">Đang tải bản xem trước…</p>
          )}
          {objectUrl && file.content_type.startsWith('image/') && (
            <img
              src={objectUrl}
              alt={`Bản xem trước ${file.original_name}`}
              data-testid="preview-image"
            />
          )}
          {objectUrl && file.content_type === 'application/pdf' && (
            <iframe
              src={objectUrl}
              title={`Bản xem trước ${file.original_name}`}
              data-testid="preview-pdf"
            />
          )}
          {objectUrl && file.content_type.startsWith('text/') && (
            <iframe
              src={objectUrl}
              title={`Bản xem trước ${file.original_name}`}
              data-testid="preview-text"
            />
          )}
        </div>
      </div>
    </div>
  );
}

function UploadProgressList({ progress }: { progress: FileUploadProgress[] }) {
  if (progress.length === 0) return null;
  return (
    <div className="upload-progress" role="region" aria-label="Tiến trình tải tệp">
      <p role="status" aria-live="polite" className="visually-hidden">
        {progress.filter((p) => p.status === 'uploading').length > 0
          ? `Đang tải ${progress.filter((p) => p.status === 'uploading').length} tệp`
          : progress.some((p) => p.status === 'error')
            ? 'Một số tệp tải lên thất bại'
            : 'Đã tải xong tất cả tệp'}
      </p>
      <ul>
        {progress.map((p) => (
          <li key={p.fileName} className="upload-progress-item" data-testid={`upload-progress-${p.fileName}`}>
            <span className="upload-progress-name">{p.fileName}</span>
            <progress
              max={100}
              value={p.progress}
              aria-label={`Tiến trình tải ${p.fileName}`}
              data-testid={`upload-progress-bar-${p.fileName}`}
            />
            <span className="upload-progress-status">
              {p.status === 'error'
                ? `Lỗi: ${p.error ?? 'không rõ'}`
                : p.status === 'success'
                  ? 'Xong'
                  : `${p.progress}%`}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}

function ResourceRow({
  resource,
  isManager,
  canPreview,
  onPublish,
  onArchive,
  onUpload,
  onDownload,
  onPreview,
  onRefresh,
}: {
  resource: { data: { id: string; title: string; description: string; status: string; updated_at: string; context_type: string; context_id: string } };
  isManager: boolean;
  canPreview: boolean;
  onPublish: (id: string) => void;
  onArchive: (id: string) => void;
  onUpload: (id: string, files: FileList) => void;
  onDownload: (id: string, file: ResourceFile) => void;
  onPreview: (id: string, file: ResourceFile) => void;
  onRefresh: (id: string) => void;
}) {
  const filesQuery = useResourceFilesQuery(canPreview ? resource.data.id : null);
  const item = resource.data;
  return (
    <li className="resource-card" data-testid={`resource-card-${item.id}`}>
      <header className="resource-card-header">
        <div>
          <h3>
            {item.title}{' '}
            <span
              className={`status-pill status-${item.status.toLowerCase()}`}
              aria-label={`Trạng thái ${item.status}`}
            >
              {item.status}
            </span>
          </h3>
          {item.description && <p className="muted">{item.description}</p>}
        </div>
        <div className="resource-card-actions">
          {isManager && item.status === 'DRAFT' && (
            <button
              type="button"
              onClick={() => onPublish(item.id)}
              data-testid={`publish-${item.id}`}
            >
              Xuất bản
            </button>
          )}
          {isManager && item.status !== 'ARCHIVED' && (
            <button
              type="button"
              className="danger"
              onClick={() => onArchive(item.id)}
              data-testid={`archive-${item.id}`}
            >
              Lưu trữ
            </button>
          )}
        </div>
      </header>
      <div className="resource-card-body">
        {canPreview && (
          <>
            <div className="resource-card-files" data-testid={`resource-files-${item.id}`}>
              <h4>Tệp đính kèm</h4>
              {filesQuery.isPending && <p role="status" aria-live="polite">Đang tải…</p>}
              {filesQuery.error && (
                <ErrorState
                  error={filesQuery.error}
                  onRetry={() => onRefresh(item.id)}
                />
              )}
              {filesQuery.data && filesQuery.data.length === 0 && (
                <p className="muted">Chưa có tệp nào.</p>
              )}
              {filesQuery.data && filesQuery.data.length > 0 && (
                <ul className="resource-file-list">
                  {filesQuery.data.map((file) => (
                    <li key={file.id} className="resource-file-item" data-testid={`resource-file-${file.id}`}>
                      <div className="resource-file-meta">
                        <strong>{file.original_name}</strong>
                        <span className="muted">
                          {file.content_type} · {formatBytes(file.size_bytes)} ·{' '}
                          {new Date(file.created_at).toLocaleString('vi-VN')}
                        </span>
                      </div>
                      <div className="resource-file-actions">
                        {isPreviewable(file.content_type) && (
                          <button
                            type="button"
                            onClick={() => onPreview(item.id, file)}
                            data-testid={`preview-${file.id}`}
                          >
                            Xem trước
                          </button>
                        )}
                        <button
                          type="button"
                          onClick={() => onDownload(item.id, file)}
                          data-testid={`download-${file.id}`}
                        >
                          Tải về
                        </button>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>
            {isManager && item.status !== 'ARCHIVED' && (
              <label className="resource-upload" data-testid={`upload-label-${item.id}`}>
                <span className="visually-hidden">{`Tải tệp lên cho tài liệu ${item.title}`}</span>
                <input
                  type="file"
                  multiple
                  data-testid={`upload-${item.id}`}
                  onChange={(e) => {
                    if (e.target.files && e.target.files.length > 0) {
                      onUpload(item.id, e.target.files);
                      e.target.value = '';
                    }
                  }}
                />
              </label>
            )}
          </>
        )}
      </div>
    </li>
  );
}

export function ResourcesPage() {
  const auth = useAuth();
  const manager = isManager(auth.actor?.roles ?? []);
  useDocumentTitle('Tài liệu');

  const [scope, setScope] = useState<'all' | 'organization' | 'class'>('all');
  const [classId, setClassId] = useState<string>('');

  const classesQuery = useClasses();
  const classes = useMemo(() => {
    const list = (classesQuery.data?.data ?? []) as Array<{ id: string; name: string }>;
    return list;
  }, [classesQuery.data]);

  const listFilter =
    scope === 'class' && classId
      ? { contextType: 'class' as const, contextID: classId }
      : scope === 'organization'
        ? { contextType: 'organization' as const, contextID: auth.actor?.organizationId ?? '' }
        : {};
  const { data, isPending, error, refetch } = useResourcesQuery(listFilter);
  const createMutation = useCreateResourceMutation();
  const publishMutation = usePublishResourceMutation();
  const archiveMutation = useArchiveResourceMutation();
  const downloadMutation = useDownloadResourceMutation();
  const queryClient = useQueryClient();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [createScope, setCreateScope] = useState<'organization' | 'class'>('organization');
  const [createClassId, setCreateClassId] = useState<string>('');
  const [uploadProgress, setUploadProgress] = useState<FileUploadProgress[]>([]);
  const [preview, setPreview] = useState<{ resourceId: string; file: ResourceFile } | null>(null);

  if (auth.status !== 'authenticated' || !auth.actor) {
    return <div>Bạn cần đăng nhập để xem tài liệu.</div>;
  }

  const resources = data?.data ?? [];
  const createError = createMutation.error
    ? formatFriendlyError(createMutation.error)
    : null;
  const hasAnyUploadError = uploadProgress.some((p) => p.status === 'error');

  const handleUpload = async (resourceId: string, files: FileList) => {
    const fileArray = Array.from(files);
    if (fileArray.length === 0) return;
    setUploadProgress(
      fileArray.map((file) => ({
        file,
        fileName: file.name,
        size: file.size,
        loaded: 0,
        progress: 0,
        status: 'pending' as const,
      }))
    );
    const result = await uploadResourceFilesWithProgress(
      resourceId,
      fileArray,
      setUploadProgress
    );
    setUploadProgress(result);
    void queryClient.invalidateQueries({ queryKey: ['resources'] });
    // Clear progress after a short delay so the user can see the result.
    setTimeout(() => setUploadProgress((current) => (current === result ? [] : current)), 3000);
  };

  return (
    <section className="resources-page" aria-labelledby="resources-heading">
      <h1 id="resources-heading">Tài liệu</h1>
      <p className="muted">
        Tài liệu được tạo bởi giáo viên và quản trị viên trong tổ chức của bạn.
        Học sinh chỉ thấy tài liệu đã xuất bản.
      </p>

      <div className="resources-filters" role="region" aria-label="Bộ lọc tài liệu">
        <fieldset>
          <legend>Phạm vi</legend>
          <label>
            <input
              type="radio"
              name="scope"
              value="all"
              checked={scope === 'all'}
              onChange={() => setScope('all')}
              data-testid="scope-all"
            />
            Tất cả
          </label>
          <label>
            <input
              type="radio"
              name="scope"
              value="organization"
              checked={scope === 'organization'}
              onChange={() => setScope('organization')}
              data-testid="scope-organization"
            />
            Tổ chức
          </label>
          {manager && (
            <label>
              <input
                type="radio"
                name="scope"
                value="class"
                checked={scope === 'class'}
                onChange={() => setScope('class')}
                data-testid="scope-class"
              />
              Theo lớp
            </label>
          )}
        </fieldset>
        {scope === 'class' && (
          <label>
            <span>Chọn lớp</span>
            <select
              value={classId}
              onChange={(e) => setClassId(e.target.value)}
              data-testid="class-selector"
            >
              <option value="">-- Chọn --</option>
              {classes.map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          </label>
        )}
      </div>

      {manager && (
        <form
          className="resources-create"
          aria-labelledby="resources-create-heading"
          data-testid="resources-create"
          onSubmit={(e) => {
            e.preventDefault();
            if (!auth.actor) return;
            const orgId = auth.actor.organizationId ?? '';
            const contextId = createScope === 'class' ? createClassId : orgId;
            if (createScope === 'class' && !createClassId) return;
            createMutation.mutate(
              {
                title: title.trim(),
                description: description.trim(),
                context_type: createScope,
                context_id: contextId,
              },
              {
                onSuccess: () => {
                  setTitle('');
                  setDescription('');
                },
              }
            );
          }}
        >
          <h2 id="resources-create-heading">Tạo tài liệu mới</h2>
          <div className="field">
            <label htmlFor="resource-title">Tiêu đề</label>
            <input
              id="resource-title"
              type="text"
              required
              minLength={1}
              maxLength={255}
              value={title}
              data-testid="resource-title"
              onChange={(e) => setTitle(e.target.value)}
            />
          </div>
          <div className="field">
            <label htmlFor="resource-description">Mô tả (tuỳ chọn)</label>
            <textarea
              id="resource-description"
              maxLength={2000}
              value={description}
              data-testid="resource-description"
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <div className="field">
            <label htmlFor="resource-create-scope">Phạm vi</label>
            <select
              id="resource-create-scope"
              value={createScope}
              onChange={(e) => setCreateScope(e.target.value as 'organization' | 'class')}
              data-testid="resource-create-scope"
            >
              <option value="organization">Tổ chức</option>
              <option value="class">Theo lớp</option>
            </select>
          </div>
          {createScope === 'class' && (
            <div className="field">
              <label htmlFor="resource-create-class">Lớp phụ trách</label>
              <select
                id="resource-create-class"
                value={createClassId}
                onChange={(e) => setCreateClassId(e.target.value)}
                required
                data-testid="resource-create-class"
              >
                <option value="">-- Chọn --</option>
                {classes.map((c) => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>
          )}
          <button
            type="submit"
            disabled={createMutation.isPending || !title.trim() || (createScope === 'class' && !createClassId)}
            aria-busy={createMutation.isPending}
            data-testid="resource-create-submit"
          >
            {createMutation.isPending ? 'Đang tạo…' : 'Tạo tài liệu'}
          </button>
          {createError && <p className="form-error" role="alert">{createError}</p>}
        </form>
      )}

      {uploadProgress.length > 0 && (
        <UploadProgressList progress={uploadProgress} />
      )}
      {hasAnyUploadError && (
        <p className="form-error" role="alert">Một số tệp không tải lên được. Vui lòng thử lại.</p>
      )}

      {isPending && (
        <p role="status" aria-live="polite">Đang tải tài liệu…</p>
      )}
      {error && <ErrorState error={error} onRetry={() => void refetch()} />}

      {!isPending && !error && (
        <ul className="resource-list" data-testid="resources-list">
          {resources.length === 0 ? (
            <li className="empty">Chưa có tài liệu nào.</li>
          ) : (
            resources.map((r) => (
              <ResourceRow
                key={r.data.id}
                resource={r as Parameters<typeof ResourceRow>[0]['resource']}
                isManager={manager}
                canPreview={true}
                onPublish={(id) => publishMutation.mutate(id)}
                onArchive={(id) => archiveMutation.mutate(id)}
                onUpload={handleUpload}
                onDownload={async (id, file) => {
                  try {
                    await downloadMutation.mutateAsync({
                      resourceId: id,
                      filename: file.original_name,
                      fileId: file.id,
                    });
                  } catch (err) {
                    // eslint-disable-next-line no-console
                    console.error('download failed', getApiErrorDetails(err));
                  }
                }}
                onPreview={(id, file) => setPreview({ resourceId: id, file })}
                onRefresh={() => void refetch()}
              />
            ))
          )}
        </ul>
      )}

      {preview && (
        <InlinePreviewModal
          resourceId={preview.resourceId}
          file={preview.file}
          onClose={() => setPreview(null)}
        />
      )}
    </section>
  );
}
