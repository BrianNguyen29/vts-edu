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

export interface ResourceListFilter {
  contextType?: 'organization' | 'class';
  contextID?: string;
}

export type FileUploadStatus = 'pending' | 'uploading' | 'success' | 'error';

export interface FileUploadProgress {
  file: File;
  fileName: string;
  size: number;
  loaded: number;
  progress: number; // 0..100
  status: FileUploadStatus;
  error?: string;
  resultFile?: ResourceFile;
}

export async function listResources(filter: ResourceListFilter = {}): Promise<ResourceList> {
  const params: Record<string, string> = {};
  if (filter.contextType) params['context_type'] = filter.contextType;
  if (filter.contextID) params['context_id'] = filter.contextID;
  const url = new URL(joinApiUrl((await loadRuntimeConfig()).apiBaseUrl, '/resources'));
  for (const [k, v] of Object.entries(params)) {
    url.searchParams.set(k, v);
  }
  const accessToken = getAccessToken();
  const headers: HeadersInit = {};
  if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;
  const r = await fetch(url.toString(), { headers, credentials: 'include' });
  if (!r.ok) {
    throw createApiError(r.status, {}, r);
  }
  return (await r.json()) as ResourceList;
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

export async function listResourceFiles(resourceId: string): Promise<ResourceFile[]> {
  const config = await loadRuntimeConfig();
  const url = joinApiUrl(config.apiBaseUrl, `/resources/${resourceId}/files`);
  const headers = new Headers();
  const accessToken = getAccessToken();
  if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`);
  const r = await fetch(url, { headers, credentials: 'include' });
  if (!r.ok) {
    throw createApiError(r.status, {}, r);
  }
  const envelopes = (await r.json()) ?? [];
  return envelopes.map((e: { data: ResourceFile }) => e.data);
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

// uploadResourceFilesWithProgress uploads one or more files using XHR so
// per-file progress events are available. The returned list mirrors the
// per-file order in `files` and reports success / failure for each entry.
export function uploadResourceFilesWithProgress(
  resourceId: string,
  files: File[],
  onProgress: (progress: FileUploadProgress[]) => void,
  signal?: AbortSignal
): Promise<FileUploadProgress[]> {
  const configP = loadRuntimeConfig();
  const state: FileUploadProgress[] = files.map((file) => ({
    file,
    fileName: file.name,
    size: file.size,
    loaded: 0,
    progress: 0,
    status: 'pending' as FileUploadStatus,
  }));
  const notify = () => onProgress(state.map((s) => ({ ...s })));

  // Concurrency limit of 3 keeps the UI responsive without flooding the
  // server with parallel multipart bodies.
  const concurrency = 3;
  let cursor = 0;
  return configP.then(async (config) => {
    const accessToken = getAccessToken();
    let csrfToken = getCsrfToken();
    if (!csrfToken) {
      csrfToken = await fetchCsrfToken(config.apiBaseUrl);
    }
    const workers: Promise<void>[] = [];
    const next = async (): Promise<void> => {
      while (cursor < files.length) {
        if (signal?.aborted) {
          return;
        }
        const i = cursor++;
        const entry = state[i];
        entry.status = 'uploading';
        notify();
        try {
          const result = await new Promise<ResourceFile>((resolve, reject) => {
            const xhr = new XMLHttpRequest();
            const form = new FormData();
            form.append('files[]', entry.file, entry.fileName);
            xhr.open('POST', joinApiUrl(config.apiBaseUrl, `/resources/${resourceId}/files`));
            if (accessToken) {
              xhr.setRequestHeader('Authorization', `Bearer ${accessToken}`);
            }
            xhr.setRequestHeader(getCsrfHeaderName(), csrfToken);
            xhr.upload.onprogress = (evt) => {
              if (!evt.lengthComputable) return;
              entry.loaded = evt.loaded;
              entry.progress = Math.round((evt.loaded / evt.total) * 100);
              notify();
            };
            xhr.onload = () => {
              if (xhr.status >= 200 && xhr.status < 300) {
                try {
                  const payload = JSON.parse(xhr.responseText) as Array<{ data: ResourceFile }>;
                  const match = Array.isArray(payload)
                    ? payload.find((p) => p.data?.original_name === entry.fileName)
                    : undefined;
                  resolve(match ? match.data : (payload as unknown as ResourceFile));
                } catch (err) {
                  reject(err instanceof Error ? err : new Error(String(err)));
                }
              } else {
                reject(createApiError(xhr.status, {}, { status: xhr.status } as Response));
              }
            };
            xhr.onerror = () => reject(new Error('Network error during upload'));
            xhr.onabort = () => reject(new DOMException('Aborted', 'AbortError'));
            if (signal) {
              const onAbort = () => xhr.abort();
              signal.addEventListener('abort', onAbort, { once: true });
            }
            xhr.send(form);
          });
          entry.status = 'success';
          entry.progress = 100;
          entry.resultFile = result;
        } catch (err) {
          entry.status = 'error';
          entry.error = err instanceof Error ? err.message : String(err);
        }
        notify();
      }
    };
    for (let i = 0; i < Math.min(concurrency, files.length); i++) {
      workers.push(next());
    }
    await Promise.all(workers);
    return state.map((s) => ({ ...s }));
  });
}

export function buildResourceDownloadUrl(
  resourceId: string,
  options: { fileId?: string; disposition?: 'attachment' | 'inline' } = {}
): string {
  const params = new URLSearchParams();
  if (options.fileId) params.set('file_id', options.fileId);
  if (options.disposition) params.set('disposition', options.disposition);
  const q = params.toString();
  return `/resources/${resourceId}/download${q ? `?${q}` : ''}`;
}

export async function fetchResourceDownload(
  resourceId: string,
  filename: string,
  options: { fileId?: string } = {}
): Promise<void> {
  const config = await loadRuntimeConfig();
  const url = joinApiUrl(
    config.apiBaseUrl,
    buildResourceDownloadUrl(resourceId, options)
  );

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
