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

export type AssessmentDetail = components['schemas']['AssessmentDetail']['data'];
export type Section = components['schemas']['Section']['data'];
export type Item = components['schemas']['Item']['data'];
export type Target = components['schemas']['Target']['data'];
export type CreateAssessmentRequest =
  components['schemas']['CreateAssessmentRequest'];
export type UpdateAssessmentRequest =
  components['schemas']['UpdateAssessmentRequest'];
export type CreateSectionRequest =
  components['schemas']['CreateSectionRequest'];
export type CreateItemRequest = components['schemas']['CreateItemRequest'];
export type CreateTargetRequest =
  components['schemas']['CreateTargetRequest'];
export type ValidationResult =
  components['schemas']['ValidationResult']['data'];
export type PublishResult = components['schemas']['PublishResult']['data'];

function buildQueryString(opts: ListOptions): string {
  const params = new URLSearchParams();
  if (opts.q) params.set('q', opts.q);
  if (opts.limit !== undefined) params.set('limit', String(opts.limit));
  if (opts.offset !== undefined) params.set('offset', String(opts.offset));
  if (opts.cursor) params.set('cursor', opts.cursor);
  if (opts.count) params.set('count', 'true');
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

export async function createAssessment(
  classSectionId: string,
  req: CreateAssessmentRequest
): Promise<AssessmentDetail> {
  const res = await apiClient(`/classes/${classSectionId}/assessments`, {
    method: 'POST',
    body: JSON.stringify(req),
  });
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: AssessmentDetail }).data;
}

export async function getAssessment(id: string): Promise<AssessmentDetail> {
  const res = await apiClient(`/assessments/${id}`);
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: AssessmentDetail }).data;
}

export async function updateAssessment(
  id: string,
  req: UpdateAssessmentRequest
): Promise<AssessmentDetail> {
  const res = await apiClient(`/assessments/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(req),
  });
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: AssessmentDetail }).data;
}

export async function createSection(
  assessmentId: string,
  req: CreateSectionRequest
): Promise<Section> {
  const res = await apiClient(`/assessments/${assessmentId}/sections`, {
    method: 'POST',
    body: JSON.stringify(req),
  });
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: Section }).data;
}

export async function createItem(
  sectionId: string,
  req: CreateItemRequest
): Promise<Item> {
  const res = await apiClient(`/assessment-sections/${sectionId}/items`, {
    method: 'POST',
    body: JSON.stringify(req),
  });
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: Item }).data;
}

export async function createTarget(
  assessmentId: string,
  req: CreateTargetRequest
): Promise<Target> {
  const res = await apiClient(`/assessments/${assessmentId}/targets`, {
    method: 'POST',
    body: JSON.stringify(req),
  });
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: Target }).data;
}

export async function validateAssessment(id: string): Promise<ValidationResult> {
  const res = await apiClient(`/assessments/${id}/validate`, {
    method: 'POST',
  });
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: ValidationResult }).data;
}

export async function publishAssessment(id: string): Promise<PublishResult> {
  const res = await apiClient(`/assessments/${id}/publish`, {
    method: 'POST',
  });
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: PublishResult }).data;
}

export { ApiResponseError };
