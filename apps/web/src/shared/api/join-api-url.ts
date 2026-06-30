const API_VERSION_PREFIX = '/api/v1';

/**
 * Join a base API URL with a request path without duplicating `/api/v1`.
 *
 * Supports both same-origin dev bases (`/api/v1`) and absolute production
 * bases (`https://api.example.com/api/v1`). Callers should pass paths
 * relative to the versioned base (e.g. `/healthz`); legacy paths that still
 * include `/api/v1` are deduplicated as a safety net.
 */
export function joinApiUrl(baseUrl: string, path: string): string {
  const base = baseUrl.replace(/\/+$/, '');
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;

  if (
    base.endsWith(API_VERSION_PREFIX) &&
    normalizedPath.startsWith(API_VERSION_PREFIX)
  ) {
    return base + normalizedPath.slice(API_VERSION_PREFIX.length);
  }

  return base + normalizedPath;
}
