import { describe, it, expect } from 'vitest';
import { validatePassword } from './password-policy';

describe('validatePassword', () => {
  it('approves a strong password', () => {
    expect(validatePassword('StrongPass1')).toEqual({
      valid: true,
      minLength: true,
      hasUppercase: true,
      hasLowercase: true,
      hasDigit: true,
      notBlocked: true,
    });
  });

  it('rejects a password that is too short', () => {
    const result = validatePassword('Short1');
    expect(result.valid).toBe(false);
    expect(result.minLength).toBe(false);
  });

  it('rejects a password without an uppercase letter', () => {
    const result = validatePassword('lowercase1');
    expect(result.valid).toBe(false);
    expect(result.hasUppercase).toBe(false);
  });

  it('rejects a password without a lowercase letter', () => {
    const result = validatePassword('UPPERCASE1');
    expect(result.valid).toBe(false);
    expect(result.hasLowercase).toBe(false);
  });

  it('rejects a password without a digit', () => {
    const result = validatePassword('NoDigitsHere');
    expect(result.valid).toBe(false);
    expect(result.hasDigit).toBe(false);
  });

  it('rejects a common password', () => {
    const result = validatePassword('Password123');
    expect(result.valid).toBe(false);
    expect(result.notBlocked).toBe(false);
  });

  it('rejects a common password regardless of case', () => {
    const result = validatePassword('PASSWORD123');
    expect(result.notBlocked).toBe(false);
  });
});
