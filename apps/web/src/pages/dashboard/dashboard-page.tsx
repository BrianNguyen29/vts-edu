import { Link } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';
import { DEMO_ATTEMPT_ID } from '@/shared/config/demo-attempt';

export function DashboardPage() {
  const auth = useAuth();

  return (
    <div className="dashboard-page">
      <h1>Trang làm việc</h1>
      <p>
        Xin chào, <strong>{auth.actor?.displayName ?? 'bạn'}</strong>.
      </p>
      <p>
        Đây là bản khung MVP. Các chức năng lớp học, bài kiểm tra và sổ điểm
        sẽ được bổ sung theo kế hoạch backend.
      </p>

      <div className="dashboard-cards">
        <article className="dashboard-card">
          <h2>Bài kiểm tra</h2>
          <p>Xem các bài kiểm tra được giao.</p>
          <Link to={`/exam/attempts/${DEMO_ATTEMPT_ID}`} className="card-link">
            Thi thử demo
          </Link>
        </article>
        <article className="dashboard-card">
          <h2>Kết quả</h2>
          <p>Xem kết quả và nhận xét.</p>
        </article>
        <article className="dashboard-card">
          <h2>Cài đặt</h2>
          <p>Quản lý thông tin cá nhân và bảo mật.</p>
        </article>
      </div>
    </div>
  );
}
