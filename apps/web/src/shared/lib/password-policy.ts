export interface PasswordValidationResult {
  valid: boolean;
  minLength: boolean;
  hasUppercase: boolean;
  hasLowercase: boolean;
  hasDigit: boolean;
  notBlocked: boolean;
}

const BLOCKED_PASSWORDS = new Set([
  'password',
  'password123',
  '12345678',
  'qwertyuiop',
  '11111111',
  'password123!',
]);

export function validatePassword(password: string): PasswordValidationResult {
  const lower = password.toLowerCase();
  const minLength = password.length >= 8;
  let hasUppercase = false;
  let hasLowercase = false;
  let hasDigit = false;

  for (const char of password) {
    const code = char.charCodeAt(0);
    if (code >= 65 && code <= 90) hasUppercase = true;
    if (code >= 97 && code <= 122) hasLowercase = true;
    if (code >= 48 && code <= 57) hasDigit = true;
  }

  const notBlocked = !BLOCKED_PASSWORDS.has(lower);
  const valid =
    minLength && hasUppercase && hasLowercase && hasDigit && notBlocked;

  return { valid, minLength, hasUppercase, hasLowercase, hasDigit, notBlocked };
}
