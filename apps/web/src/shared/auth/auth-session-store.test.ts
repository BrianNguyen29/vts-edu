import { describe, it, expect } from 'vitest';
import {
  getAccessToken,
  setAccessToken,
  clearAccessToken,
} from './auth-session-store';

describe('auth session store', () => {
  it('starts with no token', () => {
    clearAccessToken();
    expect(getAccessToken()).toBeNull();
  });

  it('stores and returns an access token', () => {
    setAccessToken('access-token-123');
    expect(getAccessToken()).toBe('access-token-123');
  });

  it('clears the access token', () => {
    setAccessToken('access-token-123');
    clearAccessToken();
    expect(getAccessToken()).toBeNull();
  });

  it('accepts null without error', () => {
    setAccessToken(null);
    expect(getAccessToken()).toBeNull();
  });
});
