import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';

function isActive(currentPath: string, targetPath: string, exact = false): boolean {
  if (exact) return currentPath === targetPath;
  return currentPath === targetPath || currentPath.startsWith(`${targetPath}/`);
}

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
  const isHomeActive = isActive(location.pathname, '/app');
  const isResourcesActive = isActive(location.pathname, '/app/resources');
  const isExamActive = isActive(location.pathname, '/exam/attempts');

  return (
    <div className="app-shell">
      <a href="#main-content" className="skip-link">
        Bỏ qua đến nội dung chính
      </a>
      <header className="app-shell-header" role="banner">
        <div className="app-shell-brand">
          <Link to="/app" className="app-shell-logo" aria-label="VTS EDU – trang chủ">
            VTS EDU
          </Link>
        </div>
        {!isRestricted && (
          <nav className="app-shell-nav" aria-label="Điều hướng chính">
            <Link
              to="/app"
              className={isHomeActive ? 'active' : ''}
              aria-current={isHomeActive ? 'page' : undefined}
            >
              Trang làm việc
            </Link>
            <Link
              to="/app/resources"
              className={isResourcesActive ? 'active' : ''}
              aria-current={isResourcesActive ? 'page' : undefined}
            >
              Tài liệu
            </Link>
            <Link
              to="/diagnostics"
              className={location.pathname === '/diagnostics' ? 'active' : ''}
              aria-current={location.pathname === '/diagnostics' ? 'page' : undefined}
            >
              Chẩn đoán
            </Link>
            <Link
              to="/exam/attempts/demo"
              className={isExamActive ? 'active' : ''}
              aria-current={isExamActive ? 'page' : undefined}
            >
              Thi thử
            </Link>
          </nav>
        )}
        <div className="app-shell-user">
          <span className="user-name" aria-label={`Người dùng ${displayName}`}>
            {displayName}
          </span>
          <button
            type="button"
            onClick={handleLogout}
            aria-label={`Đăng xuất khỏi tài khoản ${displayName}`}
          >
            Đăng xuất
          </button>
        </div>
      </header>
      <main id="main-content" className="app-shell-main" tabIndex={-1}>
        <Outlet />
      </main>
    </div>
  );
}
