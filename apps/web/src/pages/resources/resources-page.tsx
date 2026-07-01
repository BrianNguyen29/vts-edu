import { useState } from 'react';
import { useAuth } from '@/app/providers/auth-provider';
import {
  useArchiveResourceMutation,
  useCreateResourceMutation,
  usePublishResourceMutation,
  useResourcesQuery,
  useUploadResourceFileMutation,
} from '@/shared/api/resources-queries';
import { fetchResourceDownload } from '@/shared/api/resources';
import { ErrorState } from '@/shared/components/error-state';
import type { ResourceEnvelope } from '@/shared/api/resources';
import { getApiErrorDetails } from '@/shared/api/api-error';
import { formatFriendlyError } from '@/shared/api/api-error';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

function isManager(roles: string[]): boolean {
  return roles.includes('teacher') || roles.includes('admin');
}

function ResourceRow({
  resource,
  isManager: canManage,
  onPublish,
  onArchive,
  onUpload,
  onDownload,
}: {
  resource: ResourceEnvelope;
  isManager: boolean;
  onPublish: (id: string) => void;
  onArchive: (id: string) => void;
  onUpload: (id: string, file: File) => void;
  onDownload: (id: string, filename: string) => void;
}) {
  const [fileName, setFileName] = useState('');
  const item = resource.data;
  return (
    <tr>
      <td>
        <div className="resource-title">
          <strong>{item.title}</strong>
          {item.description && (
            <p className="resource-description">{item.description}</p>
          )}
        </div>
      </td>
      <td>
        <span
          className={`status-pill status-${item.status.toLowerCase()}`}
          aria-label={`Trạng thái ${item.status}`}
        >
          {item.status}
        </span>
      </td>
      <td>
        <time dateTime={item.updated_at}>
          {new Date(item.updated_at).toLocaleString('vi-VN')}
        </time>
      </td>
      <td className="resource-actions">
        {canManage && item.status === 'DRAFT' && (
          <button
            type="button"
            onClick={() => onPublish(item.id)}
            disabled={!fileName}
            aria-label={`Xuất bản tài liệu ${item.title}`}
            data-testid={`publish-${item.id}`}
          >
            Xuất bản
          </button>
        )}
        {canManage && (
          <label className="resource-upload">
            <span className="visually-hidden">{`Tải tệp lên cho tài liệu ${item.title}`}</span>
            <input
              type="file"
              data-testid={`upload-${item.id}`}
              onChange={(e) => {
                const f = e.target.files?.[0];
                if (f) {
                  setFileName(f.name);
                  onUpload(item.id, f);
                }
              }}
            />
          </label>
        )}
        {item.status === 'PUBLISHED' && (
          <button
            type="button"
            onClick={() => onDownload(item.id, item.title)}
            aria-label={`Tải về tài liệu ${item.title}`}
            data-testid={`download-${item.id}`}
          >
            Tải về
          </button>
        )}
        {canManage && item.status !== 'ARCHIVED' && (
          <button
            type="button"
            className="danger"
            onClick={() => onArchive(item.id)}
            aria-label={`Lưu trữ tài liệu ${item.title}`}
            data-testid={`archive-${item.id}`}
          >
            Lưu trữ
          </button>
        )}
      </td>
    </tr>
  );
}

export function ResourcesPage() {
  const auth = useAuth();
  const manager = isManager(auth.actor?.roles ?? []);

  useDocumentTitle('Tài liệu');

  const { data, isPending, error, refetch } = useResourcesQuery();
  const createMutation = useCreateResourceMutation();
  const publishMutation = usePublishResourceMutation();
  const archiveMutation = useArchiveResourceMutation();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [activeResourceId, setActiveResourceId] = useState<string | null>(null);
  const uploadMutation = useUploadResourceFileMutation(activeResourceId ?? '');

  if (auth.status !== 'authenticated' || !auth.actor) {
    return <div>Bạn cần đăng nhập để xem tài liệu.</div>;
  }

  const resources = data?.data ?? [];
  const createError = createMutation.error
    ? formatFriendlyError(createMutation.error)
    : null;
  const uploadError =
    activeResourceId && uploadMutation.error
      ? formatFriendlyError(uploadMutation.error)
      : null;

  return (
    <section className="resources-page" aria-labelledby="resources-heading">
      <h1 id="resources-heading">Tài liệu</h1>
      <p className="muted">
        Tài liệu được tạo bởi giáo viên và quản trị viên trong tổ chức của bạn.
        Học sinh chỉ thấy tài liệu đã xuất bản.
      </p>

      {manager && (
        <form
          className="resources-create"
          aria-labelledby="resources-create-heading"
          data-testid="resources-create"
          onSubmit={(e) => {
            e.preventDefault();
            if (!auth.actor) return;
            const orgId = auth.actor.organizationId ?? '';
            createMutation.mutate(
              {
                title: title.trim(),
                description: description.trim(),
                context_type: 'organization',
                context_id: orgId,
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
          <button
            type="submit"
            disabled={createMutation.isPending || !title.trim()}
            aria-busy={createMutation.isPending}
            data-testid="resource-create-submit"
          >
            {createMutation.isPending ? 'Đang tạo…' : 'Tạo tài liệu'}
          </button>
          {createError && <p className="form-error" role="alert">{createError}</p>}
        </form>
      )}

      {uploadError && <p className="form-error" role="alert">{uploadError}</p>}

      {isPending && (
        <p role="status" aria-live="polite">Đang tải tài liệu…</p>
      )}
      {error && <ErrorState error={error} onRetry={refetch} />}

      {!isPending && !error && (
        <table className="resources-table" data-testid="resources-table">
          <caption className="visually-hidden">
            Danh sách tài liệu của tổ chức
          </caption>
          <thead>
            <tr>
              <th scope="col">Tài liệu</th>
              <th scope="col">Trạng thái</th>
              <th scope="col">Cập nhật</th>
              <th scope="col">Hành động</th>
            </tr>
          </thead>
          <tbody>
            {resources.length === 0 ? (
              <tr>
                <td colSpan={4} className="empty">
                  Chưa có tài liệu nào.
                </td>
              </tr>
            ) : (
              resources.map((r) => (
                <ResourceRow
                  key={r.data.id}
                  resource={r}
                  isManager={manager}
                  onPublish={(id) => publishMutation.mutate(id)}
                  onArchive={(id) => archiveMutation.mutate(id)}
                  onUpload={(id, file) => {
                    setActiveResourceId(id);
                    uploadMutation.mutate(file);
                  }}
                  onDownload={async (id, name) => {
                    try {
                      await fetchResourceDownload(id, name);
                    } catch (err) {
                      const details = getApiErrorDetails(err);
                      // Surface to console; UI may add toast later.
                      // eslint-disable-next-line no-console
                      console.error('download failed', details);
                    }
                  }}
                />
              ))
            )}
          </tbody>
        </table>
      )}
    </section>
  );
}
