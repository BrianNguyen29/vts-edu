import { getAccessToken } from '@/shared/auth/auth-session-store';
import { loadRuntimeConfig } from '../config/runtime-config';
import {
  fetchCsrfToken,
  getCsrfHeaderName,
  getCsrfToken,
} from './csrf-middleware';
import { joinApiUrl } from './join-api-url';

const UNSAFE_METHODS = new Set(['POST', 'PUT', 'PATCH', 'DELETE']);

export interface ApiClientOptions extends RequestInit {
  skipCsrf?: boolean;
}

export async function apiClient(
  path: string,
  options: ApiClientOptions = {}
): Promise<Response> {
  const config = await loadRuntimeConfig();
  const url = joinApiUrl(config.apiBaseUrl, path);

  const headers = new Headers(options.headers);
  if (!headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const accessToken = getAccessToken();
  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`);
  }

  const init: RequestInit = {
    ...options,
    credentials: 'include',
    headers,
  };

  const method = (options.method || 'GET').toUpperCase();
  if (UNSAFE_METHODS.has(method) && !options.skipCsrf) {
    let token = getCsrfToken();
    if (!token) {
      token = await fetchCsrfToken(config.apiBaseUrl);
    }
    init.headers = new Headers(init.headers);
    (init.headers as Headers).set(getCsrfHeaderName(), token);
  }

  return fetch(url, init);
}

export { fetchCsrfToken, getCsrfToken } from './csrf-middleware';
