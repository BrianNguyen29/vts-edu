import createClient, { type Middleware } from 'openapi-fetch';
import type { paths } from './openapi-schema';
import { loadRuntimeConfig } from '../config/runtime-config';
import { getAccessToken } from '@/shared/auth/auth-session-store';
import {
  fetchCsrfToken,
  getCsrfHeaderName,
  getCsrfToken,
} from './csrf-middleware';

const UNSAFE_METHODS = new Set(['POST', 'PUT', 'PATCH', 'DELETE']);

let clientPromise: ReturnType<typeof createOpenAPIClient> | null = null;

async function createOpenAPIClient() {
  const config = await loadRuntimeConfig();
  const client = createClient<paths>({
    baseUrl: config.apiBaseUrl,
  });

  const middleware: Middleware = {
    async onRequest({ request }) {
      const headers = new Headers(request.headers);

      const accessToken = getAccessToken();
      if (accessToken) {
        headers.set('Authorization', `Bearer ${accessToken}`);
      }

      const method = request.method.toUpperCase();
      if (UNSAFE_METHODS.has(method)) {
        let csrfToken = getCsrfToken();
        if (!csrfToken) {
          csrfToken = await fetchCsrfToken(config.apiBaseUrl);
        }
        headers.set(getCsrfHeaderName(), csrfToken);
      }

      return new Request(request, {
        headers,
        credentials: 'include',
      });
    },
  };

  client.use(middleware);

  return client;
}

export async function getOpenAPIClient() {
  if (!clientPromise) {
    clientPromise = createOpenAPIClient();
  }
  return clientPromise;
}
