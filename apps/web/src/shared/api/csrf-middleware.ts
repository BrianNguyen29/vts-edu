import { joinApiUrl } from './join-api-url';

const CSRF_COOKIE_NAME = 'vts_csrf';
const CSRF_HEADER_NAME = 'X-CSRF-Token';

let cachedToken: string | null = null;

function readCsrfCookie(): string | null {
  const match = document.cookie
    .split('; ')
    .find((row) => row.startsWith(`${CSRF_COOKIE_NAME}=`));
  return match ? decodeURIComponent(match.split('=')[1]) : null;
}

export function getCsrfToken(): string | null {
  if (cachedToken) return cachedToken;
  cachedToken = readCsrfCookie();
  return cachedToken;
}

export function setCsrfToken(token: string): void {
  cachedToken = token;
}

export function clearCsrfToken(): void {
  cachedToken = null;
}

export function getCsrfHeaderName(): string {
  return CSRF_HEADER_NAME;
}

export async function fetchCsrfToken(apiBaseUrl: string): Promise<string> {
  const res = await fetch(joinApiUrl(apiBaseUrl, '/auth/csrf-token'), {
    method: 'GET',
    credentials: 'include',
  });
  if (!res.ok) {
    throw new Error(`csrf token fetch failed: ${res.status}`);
  }
  const data = (await res.json()) as { csrf_token: string };
  setCsrfToken(data.csrf_token);
  return data.csrf_token;
}
