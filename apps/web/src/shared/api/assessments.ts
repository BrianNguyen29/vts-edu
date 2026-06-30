import { apiClient } from './api-client';
import { ApiResponseError, createApiError } from './attempts';

export interface AssessmentListItem {
  id: string;
  title: string;
  status: string;
  duration_minutes: number;
}

export async function listAssessments(): Promise<AssessmentListItem[]> {
  const res = await apiClient('/assessments');
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: AssessmentListItem[] }).data;
}

export { ApiResponseError };
