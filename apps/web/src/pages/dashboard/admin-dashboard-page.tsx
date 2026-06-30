import { useAuth } from '@/app/providers/auth-provider';

export function AdminDashboardPage() {
  const auth = useAuth();

  return (
    <div className="dashboard-page">
      <h1>Trang quản trị</h1>
      <p>
        Xin chào, <strong>{auth.actor?.displayName ?? 'bạn'}</strong>.
      </p>
      <p>
        Đây là khu vực dành cho quản trị viên. Các chức năng quản lý tổ chức,
        người dùng và cấu hình hệ thống sẽ được bổ sung sau.
      </p>

      <div className="dashboard-cards">
        <article className="dashboard-card">
          <h2>Tổ chức</h2>
          <p>Quản lý thông tin tổ chức.</p>
        </article>
        <article className="dashboard-card">
          <h2>Người dùng</h2>
          <p>Quản lý tài khoản và phân quyền.</p>
        </article>
        <article className="dashboard-card">
          <h2>Hệ thống</h2>
          <p>Cấu hình và giám sát hoạt động.</p>
        </article>
      </div>
    </div>
  );
}
