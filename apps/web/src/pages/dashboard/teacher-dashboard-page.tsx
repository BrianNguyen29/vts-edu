import { useEffect, useState } from 'react';
import { useAuth } from '@/app/providers/auth-provider';
import {
  ApiResponseError,
  listAssessments,
  type AssessmentListItem,
} from '@/shared/api/assessments';

function formatFriendlyError(err: unknown): string {
  if (err instanceof ApiResponseError) {
    switch (err.status) {
      case 401:
        return 'Phiên làm việc đã hết hạn. Vui lòng đăng nhập lại.';
      case 403:
        return 'Bạn không có quyền xem danh sách đề thi.';
      default:
        return err.body.error.message || 'Không thể tải danh sách đề thi.';
    }
  }
  if (err instanceof Error && err.message === 'network') {
    return 'Không thể kết nối đến máy chủ. Vui lòng thử lại.';
  }
  return 'Đã xảy ra lỗi không mong muốn.';
}

export function TeacherDashboardPage() {
  const auth = useAuth();

  const [assessments, setAssessments] = useState<AssessmentListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const data = await listAssessments();
        if (cancelled) return;
        setAssessments(data);
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
  }, []);

  return (
    <div className="dashboard-page">
      <h1>Trang giáo viên</h1>
      <p>
        Xin chào, <strong>{auth.actor?.displayName ?? 'bạn'}</strong>.
      </p>

      <section className="dashboard-section">
        <h2>Đề thi</h2>

        {loading && <p className="dashboard-status">Đang tải danh sách đề thi…</p>}

        {error && (
          <div className="error-banner" role="alert">
            {error}
          </div>
        )}

        {!loading && !error && assessments.length === 0 && (
          <p className="dashboard-status">Chưa có đề thi nào.</p>
        )}

        {!loading && !error && assessments.length > 0 && (
          <ul className="assessment-list">
            {assessments.map((assessment) => (
              <li key={assessment.id} className="assessment-item">
                <div className="assessment-title">{assessment.title}</div>
                <div className="assessment-meta">
                  <span className="assessment-status">{assessment.status}</span>
                  <span className="assessment-duration">
                    {assessment.duration_minutes} phút
                  </span>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>

      <div className="dashboard-cards">
        <article className="dashboard-card">
          <h2>Lớp học</h2>
          <p>Quản lý danh sách lớp và học sinh.</p>
        </article>
        <article className="dashboard-card">
          <h2>Chấm điểm</h2>
          <p>Xem bài làm và cập nhật điểm.</p>
        </article>
      </div>
    </div>
  );
}
