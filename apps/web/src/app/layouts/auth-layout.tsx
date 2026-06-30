import { Outlet } from 'react-router-dom';

export function AuthLayout() {
  return (
    <div className="auth-layout">
      <div className="auth-card">
        <header className="auth-header">
          <h1>VTS EDU</h1>
          <p>Hệ thống quản lý học tập và đánh giá</p>
        </header>
        <Outlet />
      </div>
    </div>
  );
}

