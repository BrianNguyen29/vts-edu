import { validatePassword } from '@/shared/lib/password-policy';

interface PasswordPolicyHintsProps {
  password: string;
}

export function PasswordPolicyHints({ password }: PasswordPolicyHintsProps) {
  const result = validatePassword(password);

  return (
    <ul className="password-policy-hints" aria-label="Yêu cầu mật khẩu">
      <li className={result.minLength ? 'valid' : ''}>
        Ít nhất 8 ký tự
      </li>
      <li className={result.hasUppercase ? 'valid' : ''}>
        Có chữ hoa (A–Z)
      </li>
      <li className={result.hasLowercase ? 'valid' : ''}>
        Có chữ thường (a–z)
      </li>
      <li className={result.hasDigit ? 'valid' : ''}>
        Có chữ số (0–9)
      </li>
      <li className={result.notBlocked ? 'valid' : ''}>
        Không phải mật khẩu phổ biến
      </li>
    </ul>
  );
}
