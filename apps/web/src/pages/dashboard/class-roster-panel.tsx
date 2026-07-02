import { useEffect, useState } from 'react';
import {
  ApiResponseError,
  listEnrollments,
  type ClassSection,
  type Enrollment,
} from '@/shared/api/academics';

function formatFriendlyError(err: unknown): string {
  if (err instanceof ApiResponseError) {
    switch (err.status) {
      case 401:
        return 'Phiên làm việc đã hết hạn. Vui lòng đăng nhập lại.';
      case 403:
        return 'Bạn không có quyền xem danh sách học sinh.';
      case 404:
        return 'Không tìm thấy lớp học.';
      default:
        return err.body.error.message || 'Không thể tải danh sách học sinh.';
    }
  }
  if (err instanceof Error && err.message === 'network') {
    return 'Không thể kết nối đến máy chủ. Vui lòng thử lại.';
  }
  return 'Đã xảy ra lỗi không mong muốn.';
}

interface ClassRosterPanelProps {
  classSection: ClassSection;
  onClose: () => void;
}

export function ClassRosterPanel({
  classSection,
  onClose,
}: ClassRosterPanelProps) {
  const [enrollments, setEnrollments] = useState<Enrollment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const response = await listEnrollments(classSection.id);
        if (cancelled) return;
        setEnrollments(response.data);
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
  }, [classSection.id]);

  return (
    <div className="roster-panel">
      <div className="roster-header">
        <div>
          <h3>{classSection.name}</h3>
          <p className="roster-meta">
            {classSection.student_count} học sinh ·{' '}
            {classSection.teacher_count} giáo viên
          </p>
        </div>
        <button type="button" onClick={onClose} aria-label={`Đóng danh sách lớp ${classSection.name}`}>
          Đóng
        </button>
      </div>

      {loading && <p className="dashboard-status" role="status" aria-live="polite">Đang tải danh sách…</p>}

      {error && (
        <div className="error-banner" role="alert" aria-live="assertive">
          {error}
        </div>
      )}

      {!loading && !error && enrollments.length === 0 && (
        <p className="dashboard-status">Lớp chưa có học sinh.</p>
      )}

      {!loading && !error && enrollments.length > 0 && (
        <div className="roster-table-wrapper">
          <table className="roster-table">
            <caption className="visually-hidden">
              Danh sách học sinh của lớp {classSection.name}
            </caption>
            <thead>
              <tr>
                <th scope="col">Họ tên</th>
                <th scope="col">Trạng thái</th>
              </tr>
            </thead>
            <tbody>
              {enrollments.map((enrollment) => (
                <tr key={enrollment.id}>
                  <td>{enrollment.display_name}</td>
                  <td>{enrollment.status === 'ACTIVE' ? 'Đang học' : enrollment.status}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
