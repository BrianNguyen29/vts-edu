import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';

export function AppShellLayout() {
  const auth = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = async () => {
    await auth.logout();
    navigate('/login', { replace: true });
  };

  const displayName = auth.actor?.displayName ?? 'Người dùng';
  const isRestricted = auth.status === 'restricted';
  const isHomeActive =
    location.pathname === '/app' || location.pathname.startsWith('/app/');

  return (
    <div className="app-shell">
      <header className="app-shell-header">
        <div className="app-shell-brand">
          <Link to="/app" className="app-shell-logo">
            VTS EDU
          </Link>
        </div>
        {!isRestricted && (
          <nav className="app-shell-nav" aria-label="Main">
            <Link to="/app" className={isHomeActive ? 'active' : ''}>
              Trang làm việc
            </Link>
            <Link
              to="/diagnostics"
              className={location.pathname === '/diagnostics' ? 'active' : ''}
            >
              Chẩn đoán
            </Link>
            <Link
              to="/exam/attempts/demo"
              className={
                location.pathname.startsWith('/exam/attempts') ? 'active' : ''
              }
            >
              Thi thử
            </Link>
          </nav>
        )}
        <div className="app-shell-user">
          <span className="user-name">{displayName}</span>
          <button type="button" onClick={handleLogout}>
            Đăng xuất
          </button>
        </div>
      </header>
      <main className="app-shell-main">
        <Outlet />
      </main>
    </div>
  );
}
