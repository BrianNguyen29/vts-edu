import { useMemo } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';
import { DEMO_ATTEMPT_ID } from '@/shared/config/demo-attempt';
import {
  useAssignedAssessments,
  useAttemptHistory,
  useStartAttempt,
} from '@/shared/api/attempts-queries';
import {
  ApiResponseError,
  type AssignedAssessment,
  type StudentAttempt,
} from '@/shared/api/attempts';
import { ErrorState } from '@/shared/components/error-state';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

function formatDateTime(iso: string | undefined | null): string {
  if (!iso) return '—';
  return new Date(iso).toLocaleString('vi-VN');
}

function statusLabel(status: StudentAttempt['status']): string {
  switch (status) {
    case 'SUBMITTED':
      return 'Đã nộp';
    case 'EXPIRED':
      return 'Hết hạn';
    case 'IN_PROGRESS':
      return 'Đang làm';
    default:
      return status;
  }
}

export function DashboardPage() {
  const auth = useAuth();
  const navigate = useNavigate();

  useDocumentTitle('Trang làm việc');

  const {
    data: assessments = [],
    isPending: assessmentsLoading,
    error: assessmentsError,
  } = useAssignedAssessments();
  const {
    data: historyData,
    isPending: historyLoading,
    error: historyError,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useAttemptHistory();
  const startAttempt = useStartAttempt();

  const history = useMemo(
    () => historyData?.pages.flatMap((page) => page.data) ?? [],
    [historyData]
  );

  const isLoading = assessmentsLoading || historyLoading;
  const error = assessmentsError || historyError;

  const startErrorMessage =
    startAttempt.error instanceof ApiResponseError &&
    startAttempt.error.code === 'attempt_limit_reached'
      ? 'Bạn đã hết số lần làm bài này.'
      : undefined;

  async function handleStart(assessment: AssignedAssessment) {
    try {
      const snapshot = await startAttempt.mutateAsync(assessment.id);
      navigate(`/exam/attempts/${snapshot.id}`);
    } catch {
      // Error is surfaced via startAttempt.error below.
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

      {(error || startAttempt.error) && (
        <ErrorState
          error={error || startAttempt.error}
          message={startErrorMessage}
          overrides={{
            404: 'Không tìm thấy bài kiểm tra.',
            409: 'Bài kiểm tra chưa mở hoặc đã hết thời gian.',
          }}
        />
      )}

      <section className="dashboard-section" aria-labelledby="assigned-heading" data-testid="assigned-assessments-section">
        <h2 id="assigned-heading">Bài kiểm tra được giao</h2>

        {isLoading && (
          <p className="dashboard-status" role="status" aria-live="polite">
            Đang tải dữ liệu…
          </p>
        )}

        {!isLoading && assessments.length === 0 && (
          <div className="dashboard-empty">
            <p>Hiện chưa có bài kiểm tra nào được giao cho bạn.</p>
            <p className="dashboard-empty-hint">
              Liên hệ giáo viên nếu bạn cho rằng đây là thiếu sót.
            </p>
          </div>
        )}

        {!isLoading && assessments.length > 0 && (
          <>
            <AssessmentGroup
              title="Đang mở"
              assessments={open}
              startingId={startAttempt.isPending ? startAttempt.variables : null}
              onStart={handleStart}
              emptyText="Không có bài kiểm tra nào đang mở."
            />
            <AssessmentGroup
              title="Sắp mở"
              assessments={upcoming}
              startingId={startAttempt.isPending ? startAttempt.variables : null}
              onStart={handleStart}
              emptyText="Không có bài kiểm tra nào sắp mở."
            />
            <AssessmentGroup
              title="Đã đóng"
              assessments={closed}
              startingId={startAttempt.isPending ? startAttempt.variables : null}
              onStart={handleStart}
              emptyText="Không có bài kiểm tra nào đã đóng."
            />
          </>
        )}
      </section>

      <section className="dashboard-section" aria-labelledby="history-heading">
        <h2 id="history-heading">Lịch sử làm bài</h2>

        {isLoading ? (
          <p className="dashboard-status" role="status" aria-live="polite">
            Đang tải lịch sử…
          </p>
        ) : history.length === 0 ? (
          <div className="dashboard-empty">
            <p>Bạn chưa có lần làm bài nào.</p>
          </div>
        ) : (
          <ul className="attempt-history-list" aria-label="Danh sách lần làm bài">
            {history.map((attempt) => (
              <li key={attempt.id} className="attempt-history-item">
                <div className="attempt-history-info">
                  <h3>{attempt.assessment_title}</h3>
                  <p className="attempt-history-meta">
                    <span
                      className={`attempt-status status-${attempt.status.toLowerCase()}`}
                    >
                      {statusLabel(attempt.status)}
                    </span>
                    <span aria-hidden="true">·</span>
                    <span>Bắt đầu: {formatDateTime(attempt.started_at)}</span>
                    {attempt.submitted_at && (
                      <>
                        <span aria-hidden="true">·</span>
                        <span>Nộp: {formatDateTime(attempt.submitted_at)}</span>
                      </>
                    )}
                    {attempt.score !== undefined &&
                      attempt.score !== null &&
                      attempt.max_score !== undefined &&
                      attempt.max_score !== null && (
                        <>
                          <span aria-hidden="true">·</span>
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
                      aria-label={`Xem kết quả bài ${attempt.assessment_title}`}
                    >
                      Xem kết quả
                    </Link>
                  )}
                </div>
              </li>
            ))}
          </ul>
        )}

        {hasNextPage && (
          <div className="load-more">
            <button
              type="button"
              onClick={() => fetchNextPage()}
              disabled={isFetchingNextPage}
              aria-busy={isFetchingNextPage}
              data-testid="load-more-history"
            >
              {isFetchingNextPage ? 'Đang tải…' : 'Tải thêm lịch sử'}
            </button>
          </div>
        )}
      </section>

      <section className="dashboard-section" aria-labelledby="tools-heading">
        <h2 id="tools-heading">Tiện ích</h2>
        <div className="dashboard-cards">
          <article className="dashboard-card">
            <h3>Cài đặt</h3>
            <p>Quản lý thông tin cá nhân và bảo mật.</p>
          </article>
          <article className="dashboard-card">
            <h3>Thi thử demo</h3>
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
  startingId: string | null | undefined;
  onStart: (assessment: AssignedAssessment) => void;
  emptyText: string;
}) {
  return (
    <div className="assessment-group">
      <h3 className="assessment-group-title">{title}</h3>
      {assessments.length === 0 ? (
        <p className="dashboard-status">{emptyText}</p>
      ) : (
        <ul className="assessment-list" aria-label={title}>
          {assessments.map((assessment) => (
            <li key={assessment.id} className="assessment-list-item">
              <div className="assessment-info">
                <h4 className="assessment-info-title">{assessment.title}</h4>
                <p className="assessment-meta">
                  <span>{assessment.duration_minutes} phút</span>
                  <span aria-hidden="true">·</span>
                  <span>Tối đa {assessment.max_attempts} lần làm</span>
                </p>
              </div>
              <button
                type="button"
                className="primary"
                onClick={() => onStart(assessment)}
                disabled={
                  startingId === assessment.id || assessment.status !== 'OPEN'
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
