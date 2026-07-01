import { ApiResponseError, createApiError } from './api-error';
import { getOpenAPIClient } from './openapi-client';
import { loadRuntimeConfig } from '../config/runtime-config';
import { getAccessToken } from '@/shared/auth/auth-session-store';
import {
  fetchCsrfToken,
  getCsrfHeaderName,
  getCsrfToken,
} from './csrf-middleware';
import { joinApiUrl } from './join-api-url';
import type { components } from './openapi-schema';

export type ResourceEnvelope = components['schemas']['Resource'];
export type Resource = components['schemas']['Resource']['data'];
export type ResourceList = components['schemas']['ResourceList'];
export type ResourceFileEnvelope = components['schemas']['ResourceFile'];
export type ResourceFile = components['schemas']['ResourceFile']['data'];
export type CreateResourceRequest = components['schemas']['CreateResourceRequest'];

export async function listResources(): Promise<ResourceList> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.GET('/resources');
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data;
}

export async function createResource(
  body: CreateResourceRequest
): Promise<Resource> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.POST('/resources', { body });
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data.data;
}

export async function publishResource(resourceId: string): Promise<Resource> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.POST(
    '/resources/{id}/publish',
    { params: { path: { id: resourceId } } }
  );
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data.data;
}

export async function archiveResource(resourceId: string): Promise<void> {
  const client = await getOpenAPIClient();
  const { error, response } = await client.DELETE('/resources/{id}', {
    params: { path: { id: resourceId } },
  });
  if (error) {
    throw createApiError(response.status, error);
  }
}

export async function uploadResourceFile(
  resourceId: string,
  file: File
): Promise<ResourceFile> {
  const config = await loadRuntimeConfig();
  const url = joinApiUrl(config.apiBaseUrl, `/resources/${resourceId}/files`);

  const form = new FormData();
  form.append('file', file, file.name);

  const headers = new Headers();
  const accessToken = getAccessToken();
  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`);
  }
  let csrfToken = getCsrfToken();
  if (!csrfToken) {
    csrfToken = await fetchCsrfToken(config.apiBaseUrl);
  }
  headers.set(getCsrfHeaderName(), csrfToken);

  const response = await fetch(url, {
    method: 'POST',
    body: form,
    credentials: 'include',
    headers,
  });
  if (!response.ok) {
    throw createApiError(response.status, {}, response);
  }
  const payload = (await response.json()) as { data: ResourceFile };
  return payload.data;
}

export function buildResourceDownloadUrl(resourceId: string): string {
  return `/resources/${resourceId}/download`;
}

export async function fetchResourceDownload(
  resourceId: string,
  filename: string
): Promise<void> {
  const config = await loadRuntimeConfig();
  const url = joinApiUrl(config.apiBaseUrl, `/resources/${resourceId}/download`);

  const headers = new Headers();
  const accessToken = getAccessToken();
  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`);
  }
  const response = await fetch(url, {
    method: 'GET',
    credentials: 'include',
    headers,
  });
  if (!response.ok) {
    throw createApiError(response.status, {}, response);
  }
  const blob = await response.blob();
  const objectUrl = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = objectUrl;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);
  URL.revokeObjectURL(objectUrl);
}

export { ApiResponseError };

