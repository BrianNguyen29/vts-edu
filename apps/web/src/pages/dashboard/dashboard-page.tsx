import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';
import { DEMO_ATTEMPT_ID } from '@/shared/config/demo-attempt';
import {
  ApiResponseError,
  listAssignedAssessments,
  startAttempt,
  type AssignedAssessment,
} from '@/shared/api/attempts';

type DashboardStatus =
  | { type: 'loading' }
  | { type: 'error'; message: string }
  | { type: 'ready' };

function formatFriendlyError(err: unknown): string {
  if (err instanceof ApiResponseError) {
    switch (err.status) {
      case 401:
        return 'Phiên làm việc đã hết hạn. Vui lòng đăng nhập lại.';
      case 403:
        return 'Không có quyền truy cập.';
      case 404:
        return 'Không tìm thấy bài kiểm tra.';
      case 409:
        return err.body.error.code === 'attempt_limit_reached'
          ? 'Bạn đã hết số lần làm bài này.'
          : 'Bài kiểm tra chưa mở hoặc đã hết thởi gian.';
      default:
        return err.body.error.message || 'Không thể tải danh sách bài kiểm tra.';
    }
  }
  if (err instanceof Error && err.message === 'network') {
    return 'Không thể kết nối đến máy chủ. Vui lòng thử lại.';
  }
  return 'Đã xảy ra lỗi không mong muốn.';
}

export function DashboardPage() {
  const auth = useAuth();
  const navigate = useNavigate();

  const [status, setStatus] = useState<DashboardStatus>({ type: 'loading' });
  const [assessments, setAssessments] = useState<AssignedAssessment[]>([]);
  const [startingId, setStartingId] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const data = await listAssignedAssessments();
        if (cancelled) return;
        setAssessments(data);
        setStatus({ type: 'ready' });
      } catch (err) {
        if (cancelled) return;
        setStatus({ type: 'error', message: formatFriendlyError(err) });
      }
    }

    void load();

    return () => {
      cancelled = true;
    };
  }, []);

  async function handleStart(assessment: AssignedAssessment) {
    if (startingId) return;
    setStartingId(assessment.id);

    try {
      const snapshot = await startAttempt(assessment.id);
      navigate(`/exam/attempts/${snapshot.id}`);
    } catch (err) {
      setStatus({ type: 'error', message: formatFriendlyError(err) });
      setStartingId(null);
    }
  }

  const isOpen = (assessment: AssignedAssessment) => assessment.status === 'OPEN';

  return (
    <div className="dashboard-page">
      <h1>Trang làm việc</h1>
      <p>
        Xin chào, <strong>{auth.actor?.displayName ?? 'bạn'}</strong>.
      </p>

      {status.type === 'error' && (
        <div className="error-banner" role="alert">
          {status.message}
        </div>
      )}

      <section className="dashboard-section" aria-labelledby="assigned-heading">
        <h2 id="assigned-heading">Bài kiểm tra được giao</h2>

        {status.type === 'loading' && (
          <p className="dashboard-status">Đang tải danh sách bài kiểm tra…</p>
        )}

        {status.type === 'ready' && assessments.length === 0 && (
          <div className="dashboard-empty">
            <p>Hiện chưa có bài kiểm tra nào được giao cho bạn.</p>
            <p className="dashboard-empty-hint">
              Liên hệ giáo viên nếu bạn cho rằng đây là thiếu sót.
            </p>
          </div>
        )}

        {status.type === 'ready' && assessments.length > 0 && (
          <ul className="assessment-list">
            {assessments.map((assessment) => (
              <li key={assessment.id} className="assessment-list-item">
                <div className="assessment-info">
                  <h3>{assessment.title}</h3>
                  <p className="assessment-meta">
                    <span
                      className={`assessment-status ${isOpen(assessment) ? 'open' : 'closed'}`}
                    >
                      {isOpen(assessment) ? 'Đang mở' : 'Chưa mở'}
                    </span>
                    <span>·</span>
                    <span>{assessment.duration_minutes} phút</span>
                    <span>·</span>
                    <span>Tối đa {assessment.max_attempts} lần làm</span>
                  </p>
                </div>
                <button
                  type="button"
                  className="primary"
                  onClick={() => handleStart(assessment)}
                  disabled={startingId === assessment.id || !isOpen(assessment)}
                  aria-busy={startingId === assessment.id}
                >
                  {startingId === assessment.id
                    ? 'Đang bắt đầu…'
                    : isOpen(assessment)
                      ? 'Bắt đầu làm bài'
                      : 'Chưa mở'}
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="dashboard-section" aria-labelledby="tools-heading">
        <h2 id="tools-heading">Tiện ích</h2>
        <div className="dashboard-cards">
          <article className="dashboard-card">
            <h2>Kết quả</h2>
            <p>Xem kết quả và nhận xét.</p>
          </article>
          <article className="dashboard-card">
            <h2>Cài đặt</h2>
            <p>Quản lý thông tin cá nhân và bảo mật.</p>
          </article>
          <article className="dashboard-card">
            <h2>Thi thử demo</h2>
            <p>Làm bài thử nghiệm để làm quen giao diện.</p>
            <Link to={`/exam/attempts/${DEMO_ATTEMPT_ID}`} className="card-link">
              Thi thử demo
            </Link>
          </article>
        </div>
      </section>
    </div>
  );
}
