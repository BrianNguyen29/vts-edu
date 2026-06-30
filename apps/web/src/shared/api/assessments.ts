import { apiClient } from './api-client';
import {
  ApiResponseError,
  createApiError,
  type ListOptions,
} from './attempts';
import type { components, paths } from './openapi-schema';

export type AssessmentListItem = components['schemas']['AssessmentListItem'];
export type PagedAssessments =
  paths['/assessments']['get']['responses'][200]['content']['application/json'];

function buildQueryString(opts: ListOptions): string {
  const params = new URLSearchParams();
  if (opts.q) params.set('q', opts.q);
  if (opts.limit !== undefined) params.set('limit', String(opts.limit));
  if (opts.offset !== undefined) params.set('offset', String(opts.offset));
  const query = params.toString();
  return query ? `?${query}` : '';
}

export async function listAssessments(
  opts: ListOptions = {}
): Promise<PagedAssessments> {
  const res = await apiClient(`/assessments${buildQueryString(opts)}`);
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return json as PagedAssessments;
}

export { ApiResponseError };
