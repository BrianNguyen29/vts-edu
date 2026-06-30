import { useState, type FormEvent } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';

export function LoginPage() {
  const auth = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const [organizationCode, setOrganizationCode] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const returnTo = searchParams.get('returnTo') ?? '/app';

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      await auth.login({ organizationCode, username, password });
      navigate(returnTo, { replace: true });
    } catch (err) {
      const message = err instanceof Error ? err.message : 'unknown';
      if (message === 'invalid') {
        setError('Thông tin đăng nhập không đúng.');
      } else if (message === 'rate-limit') {
        setError('Quá nhiều lần thử. Vui lòng đợi một lát.');
      } else {
        setError('Không thể kết nối đến máy chủ. Vui lòng thử lại.');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="login-form" noValidate>
      {error && (
        <div className="error-banner" role="alert">
          {error}
        </div>
      )}

      <div className="field">
        <label htmlFor="organizationCode">Mã tổ chức</label>
        <input
          id="organizationCode"
          name="organizationCode"
          type="text"
          autoComplete="organization"
          required
          value={organizationCode}
          onChange={(e) => setOrganizationCode(e.target.value)}
        />
      </div>

      <div className="field">
        <label htmlFor="username">Tên đăng nhập</label>
        <input
          id="username"
          name="username"
          type="text"
          autoComplete="username"
          required
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
      </div>

      <div className="field">
        <label htmlFor="password">Mật khẩu</label>
        <input
          id="password"
          name="password"
          type="password"
          autoComplete="current-password"
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </div>

      <button type="submit" disabled={isSubmitting} className="primary">
        {isSubmitting ? 'Đang đăng nhập…' : 'Đăng nhập'}
      </button>

      <p className="login-hint">
        Bản MVP hiện yêu cầu backend đang chạy. Nếu backend chưa sẵn sàng, bạn
        sẽ thấy lỗi kết nối.
      </p>
    </form>
  );
}
