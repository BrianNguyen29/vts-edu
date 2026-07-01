import { useEffect, useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';
import { DEMO_ATTEMPT_ID } from '@/shared/config/demo-attempt';
import {
  ApiResponseError,
  listAssignedAssessments,
  listAttemptHistory,
  startAttempt,
  type AssignedAssessment,
  type StudentAttempt,
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
          : 'Bài kiểm tra chưa mở hoặc đã hết thời gian.';
      default:
        return err.body.error.message || 'Không thể tải danh sách bài kiểm tra.';
    }
  }
  if (err instanceof Error && err.message === 'network') {
    return 'Không thể kết nối đến máy chủ. Vui lòng thử lại.';
  }
  return 'Đã xảy ra lỗi không mong muốn.';
}

function formatDateTime(iso: string | undefined | null): string {
  if (!iso) return '—';
  return new Date(iso).toLocaleString('vi-VN');
}

export function DashboardPage() {
  const auth = useAuth();
  const navigate = useNavigate();

  const [status, setStatus] = useState<DashboardStatus>({ type: 'loading' });
  const [assessments, setAssessments] = useState<AssignedAssessment[]>([]);
  const [history, setHistory] = useState<StudentAttempt[]>([]);
  const [startingId, setStartingId] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const [assessmentData, historyData] = await Promise.all([
          listAssignedAssessments(),
          listAttemptHistory(),
        ]);
        if (cancelled) return;
        setAssessments(assessmentData);
        setHistory(historyData);
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

  const { open, upcoming, closed } = useMemo(() => {
    const historyByAssessment = new Map<string, StudentAttempt[]>();
    for (const attempt of history) {
      const list = historyByAssessment.get(attempt.assessment_id) ?? [];
      list.push(attempt);
      historyByAssessment.set(attempt.assessment_id, list);
    }

    const open: AssignedAssessment[] = [];
    const upcoming: AssignedAssessment[] = [];
    const closed: AssignedAssessment[] = [];

    for (const assessment of assessments) {
      const attempts = historyByAssessment.get(assessment.id) ?? [];
      const submittedOrExpired = attempts.filter(
        (a) => a.status === 'SUBMITTED' || a.status === 'EXPIRED'
      ).length;

      if (assessment.status === 'OPEN') {
        if (submittedOrExpired >= assessment.max_attempts) {
          closed.push(assessment);
        } else {
          open.push(assessment);
        }
      } else if (assessment.status === 'PUBLISHED') {
        if (submittedOrExpired >= assessment.max_attempts) {
          closed.push(assessment);
        } else {
          upcoming.push(assessment);
        }
      } else {
        closed.push(assessment);
      }
    }

    return { open, upcoming, closed };
  }, [assessments, history]);

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

      <section className="dashboard-section" aria-labelledby="assigned-heading" data-testid="assigned-assessments-section">
        <h2 id="assigned-heading">Bài kiểm tra được giao</h2>

        {status.type === 'loading' && (
          <p className="dashboard-status">Đang tải dữ liệu…</p>
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
          <>
            <AssessmentGroup
              title="Đang mở"
              assessments={open}
              startingId={startingId}
              onStart={handleStart}
              emptyText="Không có bài kiểm tra nào đang mở."
            />
            <AssessmentGroup
              title="Sắp mở"
              assessments={upcoming}
              startingId={startingId}
              onStart={handleStart}
              emptyText="Không có bài kiểm tra nào sắp mở."
            />
            <AssessmentGroup
              title="Đã đóng"
              assessments={closed}
              startingId={startingId}
              onStart={handleStart}
              emptyText="Không có bài kiểm tra nào đã đóng."
            />
          </>
        )}
      </section>

      <section className="dashboard-section" aria-labelledby="history-heading">
        <h2 id="history-heading">Lịch sử làm bài</h2>

        {status.type === 'loading' ? (
          <p className="dashboard-status">Đang tải lịch sử…</p>
        ) : history.length === 0 ? (
          <div className="dashboard-empty">
            <p>Bạn chưa có lần làm bài nào.</p>
          </div>
        ) : (
          <ul className="attempt-history-list">
            {history.map((attempt) => (
              <li key={attempt.id} className="attempt-history-item">
                <div className="attempt-history-info">
                  <h3>{attempt.assessment_title}</h3>
                  <p className="attempt-history-meta">
                    <span
                      className={`attempt-status status-${attempt.status.toLowerCase()}`}
                    >
                      {attempt.status === 'SUBMITTED'
                        ? 'Đã nộp'
                        : attempt.status === 'EXPIRED'
                          ? 'Hết hạn'
                          : attempt.status === 'IN_PROGRESS'
                            ? 'Đang làm'
                            : attempt.status}
                    </span>
                    <span>·</span>
                    <span>Bắt đầu: {formatDateTime(attempt.started_at)}</span>
                    {attempt.submitted_at && (
                      <>
                        <span>·</span>
                        <span>Nộp: {formatDateTime(attempt.submitted_at)}</span>
                      </>
                    )}
                    {attempt.score !== undefined &&
                      attempt.score !== null &&
                      attempt.max_score !== undefined &&
                      attempt.max_score !== null && (
                        <>
                          <span>·</span>
                          <span>
                            Điểm: {attempt.score} / {attempt.max_score}
                          </span>
                        </>
                      )}
                  </p>
                </div>
                <div className="attempt-history-actions">
                  {attempt.status === 'IN_PROGRESS' && (
                    <button
                      type="button"
                      className="primary"
                      onClick={() => navigate(`/exam/attempts/${attempt.id}`)}
                    >
                      Làm tiếp
                    </button>
                  )}
                  {(attempt.status === 'SUBMITTED' ||
                    attempt.status === 'EXPIRED') && (
                    <Link
                      to={`/attempts/${attempt.id}/result`}
                      className="button-link"
                    >
                      Xem kết quả
                    </Link>
                  )}
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="dashboard-section" aria-labelledby="tools-heading">
        <h2 id="tools-heading">Tiện ích</h2>
        <div className="dashboard-cards">
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

function AssessmentGroup({
  title,
  assessments,
  startingId,
  onStart,
  emptyText,
}: {
  title: string;
  assessments: AssignedAssessment[];
  startingId: string | null;
  onStart: (assessment: AssignedAssessment) => void;
  emptyText: string;
}) {
  return (
    <div className="assessment-group">
      <h3 className="assessment-group-title">{title}</h3>
      {assessments.length === 0 ? (
        <p className="dashboard-status">{emptyText}</p>
      ) : (
        <ul className="assessment-list">
          {assessments.map((assessment) => (
            <li key={assessment.id} className="assessment-list-item">
              <div className="assessment-info">
                <h3>{assessment.title}</h3>
                <p className="assessment-meta">
                  <span>{assessment.duration_minutes} phút</span>
                  <span>·</span>
                  <span>Tối đa {assessment.max_attempts} lần làm</span>
                </p>
              </div>
              <button
                type="button"
                className="primary"
                onClick={() => onStart(assessment)}
                disabled={
                  startingId === assessment.id ||
                  assessment.status !== 'OPEN'
                }
                aria-busy={startingId === assessment.id}
                data-testid="start-assessment-button"
              >
                {startingId === assessment.id
                  ? 'Đang bắt đầu…'
                  : assessment.status === 'OPEN'
                    ? 'Bắt đầu làm bài'
                    : 'Chưa mở'}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
