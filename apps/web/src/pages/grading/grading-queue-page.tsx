import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAssessments } from '@/shared/api/assessments-queries';
import { useReviewQueue } from '@/shared/api/grading-queries';
import type { AssessmentListItem } from '@/shared/api/assessments';
import { ErrorState } from '@/shared/components/error-state';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

export function GradingQueuePage() {
  useDocumentTitle('Chấm bài thủ công');

  const { data: assessmentsData, isPending: assessmentsLoading, error: assessmentsError } = useAssessments({ limit: 100 });
  const assessments = (assessmentsData?.data ?? []) as AssessmentListItem[];

  const [selectedAssessmentId, setSelectedAssessmentId] = useState<string>('');

  useEffect(() => {
    if (!selectedAssessmentId && assessments.length > 0) {
      setSelectedAssessmentId(assessments[0].id);
    }
  }, [assessments, selectedAssessmentId]);

  const {
    data: queue = [],
    isPending: queueLoading,
    error: queueError,
  } = useReviewQueue(selectedAssessmentId || undefined);

  if (assessmentsLoading) {
    return <p className="loading">Đang tải danh sách bài kiểm tra…</p>;
  }
  if (assessmentsError) {
    return (
      <ErrorState
        error={assessmentsError}
        title="Không tải được danh sách bài kiểm tra"
      />
    );
  }
  if (assessments.length === 0) {
    return (
      <div className="dashboard-page">
        <h1>Chấm bài thủ công</h1>
        <p>Chưa có bài kiểm tra nào trong tổ chức. Hãy tạo bài kiểm tra trước.</p>
      </div>
    );
  }

  return (
    <div className="dashboard-page">
      <h1>Chấm bài thủ công</h1>
      <p className="dashboard-hint">
        Danh sách các bài làm cần giáo viên chấm cho câu tự luận / trả lời ngắn.
      </p>

      <div className="dashboard-toolbar">
        <label htmlFor="grading-assessment">Bài kiểm tra</label>
        <select
          id="grading-assessment"
          data-testid="grading-assessment-select"
          value={selectedAssessmentId}
          onChange={(e) => setSelectedAssessmentId(e.target.value)}
        >
          {assessments.map((a) => (
            <option key={a.id} value={a.id}>
              {a.title}
            </option>
          ))}
        </select>
      </div>

      {queueError ? (
        <ErrorState
          error={queueError}
          title="Không tải được hàng chờ chấm"
        />
      ) : queueLoading ? (
        <p className="loading">Đang tải hàng chờ…</p>
      ) : queue.length === 0 ? (
        <p className="empty-state">Không có bài làm nào đang chờ chấm.</p>
      ) : (
        <table className="data-table" data-testid="grading-queue-table">
          <thead>
            <tr>
              <th scope="col">Học sinh</th>
              <th scope="col">Trạng thái</th>
              <th scope="col">Nộp lúc</th>
              <th scope="col">Chờ chấm</th>
              <th scope="col">Hành động</th>
            </tr>
          </thead>
          <tbody>
            {queue.map((row) => (
              <tr key={row.attempt_id} data-testid="grading-queue-row">
                <td>
                  <strong>{row.student_name || row.student_user_id}</strong>
                </td>
                <td>
                  <span className={`status-pill status-${row.status.toLowerCase()}`}>
                    {row.status}
                  </span>
                </td>
                <td>
                  {row.submitted_at
                    ? new Date(row.submitted_at).toLocaleString('vi-VN')
                    : '—'}
                </td>
                <td>
                  {row.pending_items} / {row.total_non_mcq}
                </td>
                <td>
                  <Link
                    to={`/app/grading/${row.attempt_id}`}
                    className="primary"
                    data-testid="grading-queue-grade-link"
                  >
                    Chấm
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
