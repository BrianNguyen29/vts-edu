import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';
import { PasswordPolicyHints } from '@/shared/components/password-policy-hints';
import { validatePassword } from '@/shared/lib/password-policy';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

function getRoleHomePath(roles: string[]): string {
  if (roles.includes('admin')) return '/app/admin';
  if (roles.includes('teacher')) return '/app/teacher';
  return '/app/student';
}

export function ChangePasswordPage() {
  const auth = useAuth();
  const navigate = useNavigate();

  useDocumentTitle('Đổi mật khẩu');

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setSuccess(false);

    const passwordCheck = validatePassword(newPassword);
    if (!passwordCheck.valid) {
      setError('Mật khẩu mới chưa đáp ứng yêu cầu bảo mật.');
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('Mật khẩu xác nhận không khớp.');
      return;
    }

    setIsSubmitting(true);

    try {
      await auth.changePassword({ currentPassword, newPassword });
      setSuccess(true);
      // The backend revokes all refresh sessions on password change, so the
      // user must log in again with the new password.
      setTimeout(() => {
        void auth.logout().then(() => navigate('/login'));
      }, 1500);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'unknown';
      if (message === 'invalid') {
        setError('Mật khẩu hiện tại không đúng.');
      } else {
        setError('Không thể đổi mật khẩu. Vui lòng thử lại.');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const isRequired = auth.status === 'restricted';
  const homePath = auth.actor ? getRoleHomePath(auth.actor.roles) : '/app';

  return (
    <div className="change-password-page">
      <h1>Đổi mật khẩu</h1>

      {isRequired ? (
        <p>
          Tài khoản của bạn yêu cầu đổi mật khẩu trước khi tiếp tục sử dụng hệ
          thống.
        </p>
      ) : (
        <p>Bạn có thể đổi mật khẩu tại đây.</p>
      )}

      <div
        className="error-banner"
        role="alert"
        aria-live="assertive"
        data-testid="change-password-error"
        data-error-visible={error ? 'true' : 'false'}
        style={error ? undefined : { display: 'none' }}
      >
        {error}
      </div>

      {success && (
        <div className="success-banner" role="status" aria-live="polite">
          Đổi mật khẩu thành công. Vui lòng đăng nhập lại.
        </div>
      )}

      {!isRequired && (
        <p>
          <Link to={homePath}>← Quay lại trang làm việc</Link>
        </p>
      )}

      <form
        onSubmit={handleSubmit}
        className="change-password-form"
        noValidate
        aria-describedby={error ? 'change-password-error' : undefined}
        data-testid="change-password-form"
      >
        <div className="field">
          <label htmlFor="currentPassword">Mật khẩu hiện tại</label>
          <input
            id="currentPassword"
            name="currentPassword"
            type="password"
            autoComplete="current-password"
            required
            value={currentPassword}
            onChange={(e) => setCurrentPassword(e.target.value)}
            data-testid="current-password-input"
          />
        </div>

        <div className="field">
          <label htmlFor="newPassword">Mật khẩu mới</label>
          <input
            id="newPassword"
            name="newPassword"
            type="password"
            autoComplete="new-password"
            required
            minLength={8}
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            aria-describedby="password-policy"
            data-testid="new-password-input"
          />
          <PasswordPolicyHints password={newPassword} id="password-policy" />
        </div>

        <div className="field">
          <label htmlFor="confirmPassword">Xác nhận mật khẩu mới</label>
          <input
            id="confirmPassword"
            name="confirmPassword"
            type="password"
            autoComplete="new-password"
            required
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            data-testid="confirm-password-input"
          />
        </div>

        <button
          type="submit"
          className="primary"
          disabled={isSubmitting || success}
          aria-busy={isSubmitting}
          data-testid="change-password-submit"
        >
          {isSubmitting ? 'Đang cập nhật…' : 'Đổi mật khẩu'}
        </button>
      </form>
    </div>
  );
}
