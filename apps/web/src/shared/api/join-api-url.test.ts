import { describe, it, expect } from 'vitest';
import { joinApiUrl } from './join-api-url';

describe('joinApiUrl', () => {
  it('joins a same-origin base with a relative path', () => {
    expect(joinApiUrl('/api/v1', '/healthz')).toBe('/api/v1/healthz');
  });

  it('joins an absolute production base with a relative path', () => {
    expect(joinApiUrl('https://api.example.com/api/v1', '/me')).toBe(
      'https://api.example.com/api/v1/me'
    );
  });

  it('deduplicates /api/v1 when the path already includes it', () => {
    expect(joinApiUrl('https://api.example.com/api/v1', '/api/v1/assessments')).toBe(
      'https://api.example.com/api/v1/assessments'
    );
  });

  it('handles a base with trailing slashes', () => {
    expect(joinApiUrl('https://api.example.com/api/v1/', '/attempts')).toBe(
      'https://api.example.com/api/v1/attempts'
    );
  });

  it('handles a path without a leading slash', () => {
    expect(joinApiUrl('/api/v1', 'auth/login')).toBe('/api/v1/auth/login');
  });
});
