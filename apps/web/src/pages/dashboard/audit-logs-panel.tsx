import { useEffect, useState } from 'react';
import { listAuditLogs, type AuditLog } from '@/shared/api/admin';
import { ApiResponseError } from '@/shared/api/attempts';

const ACTION_OPTIONS = [
  { value: '', label: 'Tất cả hành động' },
  { value: 'user.create', label: 'Tạo người dùng' },
  { value: 'user.update_roles', label: 'Cập nhật vai trò' },
  { value: 'user.reset_password', label: 'Đặt lại mật khẩu' },
  { value: 'organization.update', label: 'Cập nhật tổ chức' },
];

function formatFriendlyError(err: unknown): string {
  if (err instanceof ApiResponseError) {
    switch (err.status) {
      case 401:
        return 'Phiên làm việc đã hết hạn. Vui lòng đăng nhập lại.';
      case 403:
        return 'Bạn không có quyền xem nhật ký hoạt động.';
      default:
        return err.body.error.message || 'Không thể tải nhật ký hoạt động.';
    }
  }
  if (err instanceof Error && err.message === 'network') {
    return 'Không thể kết nối đến máy chủ. Vui lòng thử lại.';
  }
  return 'Đã xảy ra lỗi không mong muốn.';
}

function formatDateTime(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  return date.toLocaleString('vi-VN');
}

function summarizeJson(value: unknown): string {
  if (value === null || value === undefined) return '—';
  if (typeof value === 'object') {
    const entries = Object.entries(value as Record<string, unknown>);
    if (entries.length === 0) return '{}';
    return entries
      .map(([k, v]) => `${k}: ${JSON.stringify(v)}`)
      .join(', ');
  }
  return String(value);
}

export function AuditLogsPanel() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [auditCursor, setAuditCursor] = useState<string | undefined>();
  const [auditHasMore, setAuditHasMore] = useState(false);
  const [isLoadingMoreAudit, setIsLoadingMoreAudit] = useState(false);

  const [action, setAction] = useState('');
  const [actorInput, setActorInput] = useState('');
  const [actorUserId, setActorUserId] = useState('');
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');

  useEffect(() => {
    const timer = setTimeout(() => {
      setActorUserId(actorInput.trim());
    }, 300);
    return () => clearTimeout(timer);
  }, [actorInput]);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      setAuditCursor(undefined);
      setAuditHasMore(false);

      try {
        const opts: {
          action?: string;
          actor_user_id?: string;
          from?: string;
          to?: string;
          limit: number;
        } = { limit: 10 };
        if (action) opts.action = action;
        if (actorUserId) opts.actor_user_id = actorUserId;
        if (fromDate) opts.from = new Date(fromDate).toISOString();
        if (toDate) opts.to = new Date(toDate).toISOString();

        const response = await listAuditLogs(opts);
        if (cancelled) return;
        setLogs(response.data);
        setAuditCursor(response.page?.next_cursor ?? undefined);
        setAuditHasMore(response.page?.has_more ?? false);
      } catch (err) {
        if (cancelled) return;
        setError(formatFriendlyError(err));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    void load();

    return () => {
      cancelled = true;
    };
  }, [action, actorUserId, fromDate, toDate]);

  async function loadMoreAuditLogs() {
    if (!auditHasMore || !auditCursor || isLoadingMoreAudit) return;
    setIsLoadingMoreAudit(true);
    try {
      const opts: {
        action?: string;
        actor_user_id?: string;
        from?: string;
        to?: string;
        limit: number;
        cursor: string;
      } = { limit: 10, cursor: auditCursor };
      if (action) opts.action = action;
      if (actorUserId) opts.actor_user_id = actorUserId;
      if (fromDate) opts.from = new Date(fromDate).toISOString();
      if (toDate) opts.to = new Date(toDate).toISOString();

      const response = await listAuditLogs(opts);
      setLogs((prev) => [...prev, ...response.data]);
      setAuditCursor(response.page?.next_cursor ?? undefined);
      setAuditHasMore(response.page?.has_more ?? false);
    } catch (err) {
      setError(formatFriendlyError(err));
    } finally {
      setIsLoadingMoreAudit(false);
    }
  }

  return (
    <section className="admin-section">
      <h2>Nhật ký hoạt động</h2>

      <div className="audit-filters">
        <div className="field">
          <label htmlFor="audit-action">Hành động</label>
          <select
            id="audit-action"
            value={action}
            onChange={(e) => setAction(e.target.value)}
          >
            {ACTION_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        <div className="field">
          <label htmlFor="audit-actor">Người thực hiện (ID)</label>
          <input
            id="audit-actor"
            type="search"
            placeholder="Nhập ID người dùng…"
            value={actorInput}
            onChange={(e) => setActorInput(e.target.value)}
          />
        </div>

        <div className="field">
          <label htmlFor="audit-from">Từ ngày</label>
          <input
            id="audit-from"
            type="date"
            value={fromDate}
            onChange={(e) => setFromDate(e.target.value)}
          />
        </div>

        <div className="field">
          <label htmlFor="audit-to">Đến ngày</label>
          <input
            id="audit-to"
            type="date"
            value={toDate}
            onChange={(e) => setToDate(e.target.value)}
          />
        </div>
      </div>

      {loading && <p className="dashboard-status">Đang tải nhật ký…</p>}

      {error && (
        <div className="error-banner" role="alert">
          {error}
        </div>
      )}

      {!loading && !error && logs.length === 0 && (
        <p className="dashboard-status">Không có nhật ký nào.</p>
      )}

      {!loading && !error && logs.length > 0 && (
        <>
          <div className="audit-logs-table-wrapper">
            <table className="audit-logs-table">
              <thead>
                <tr>
                  <th>Thời gian</th>
                  <th>Hành động</th>
                  <th>Người thực hiện</th>
                  <th>Đối tượng</th>
                  <th>Thông tin thêm</th>
                </tr>
              </thead>
              <tbody>
                {logs.map((log) => (
                  <tr key={log.id}>
                    <td>{formatDateTime(log.created_at)}</td>
                    <td>{log.action}</td>
                    <td>{log.actor_user_id}</td>
                    <td>
                      {log.resource_type}:{log.resource_id}
                    </td>
                    <td className="audit-meta">
                      {summarizeJson(log.metadata)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {auditHasMore && (
            <div className="load-more">
              <button
                type="button"
                onClick={loadMoreAuditLogs}
                disabled={isLoadingMoreAudit}
              >
                {isLoadingMoreAudit ? 'Đang tải…' : 'Tải thêm'}
              </button>
            </div>
          )}
        </>
      )}
    </section>
  );
}
