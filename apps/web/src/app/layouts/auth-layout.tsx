import { Outlet } from 'react-router-dom';

export function AuthLayout() {
  return (
    <div className="auth-layout">
      <a href="#main-content" className="skip-link">
        Bỏ qua đến nội dung chính
      </a>
      <main id="main-content" className="auth-card" tabIndex={-1}>
        <header className="auth-header">
          <h1>VTS EDU</h1>
          <p>Hệ thống quản lý học tập và đánh giá</p>
        </header>
        <Outlet />
      </main>
    </div>
  );
}

