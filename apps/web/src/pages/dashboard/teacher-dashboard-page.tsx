import { useAuth } from '@/app/providers/auth-provider';

export function TeacherDashboardPage() {
  const auth = useAuth();

  return (
    <div className="dashboard-page">
      <h1>Trang giáo viên</h1>
      <p>
        Xin chào, <strong>{auth.actor?.displayName ?? 'bạn'}</strong>.
      </p>
      <p>
        Đây là khu vực dành cho giáo viên. Các chức năng quản lý lớp, đề thi
        và chấm điểm sẽ được bổ sung sau.
      </p>

      <div className="dashboard-cards">
        <article className="dashboard-card">
          <h2>Lớp học</h2>
          <p>Quản lý danh sách lớp và học sinh.</p>
        </article>
        <article className="dashboard-card">
          <h2>Đề thi</h2>
          <p>Tạo và phát hành bài kiểm tra.</p>
        </article>
        <article className="dashboard-card">
          <h2>Chấm điểm</h2>
          <p>Xem bài làm và cập nhật điểm.</p>
        </article>
      </div>
    </div>
  );
}
