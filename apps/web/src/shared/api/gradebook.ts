import { getOpenAPIClient } from './openapi-client';
import { apiClient } from './api-client';
import { unwrapData, type ApiError } from './attempts';
import type { components } from './openapi-schema';

export type AssessmentAttempt = components['schemas']['AssessmentAttempt'];
export type AssessmentResult =
  components['schemas']['AssessmentResult']['data'];
export type ClassGradebookEntry = components['schemas']['ClassGradebookEntry'];

function createApiError(raw: unknown): Error {
  if (
    typeof raw === 'object' &&
    raw !== null &&
    'error' in raw &&
    typeof (raw as ApiError).error === 'object' &&
    (raw as ApiError).error !== null &&
    typeof (raw as ApiError).error.code === 'string' &&
    typeof (raw as ApiError).error.message === 'string'
  ) {
    return new Error((raw as ApiError).error.message);
  }
  return new Error('Yêu cầu thất bại.');
}

export async function listAssessmentAttempts(
  assessmentId: string
): Promise<AssessmentAttempt[]> {
  const client = await getOpenAPIClient();
  return unwrapData<AssessmentAttempt[]>(
    await client.GET('/assessments/{id}/attempts', {
      params: { path: { id: assessmentId } },
    })
  );
}

export async function getAssessmentResults(
  assessmentId: string
): Promise<AssessmentResult> {
  const client = await getOpenAPIClient();
  return unwrapData<AssessmentResult>(
    await client.GET('/assessments/{id}/results', {
      params: { path: { id: assessmentId } },
    })
  );
}

export async function getClassGradebook(
  classId: string
): Promise<ClassGradebookEntry[]> {
  const client = await getOpenAPIClient();
  return unwrapData<ClassGradebookEntry[]>(
    await client.GET('/classes/{class_id}/gradebook', {
      params: { path: { class_id: classId } },
    })
  );
}

async function downloadCsv(path: string, filename: string): Promise<void> {
  const response = await apiClient(path, { method: 'GET' });
  if (!response.ok) {
    let raw: unknown;
    try {
      raw = await response.json();
    } catch {
      raw = null;
    }
    throw createApiError(raw);
  }

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
}

export async function exportAssessmentAttemptsCSV(
  assessmentId: string
): Promise<void> {
  return downloadCsv(
    `/assessments/${assessmentId}/attempts/export`,
    `assessment-${assessmentId}-attempts.csv`
  );
}

export async function exportClassGradebookCSV(classId: string): Promise<void> {
  return downloadCsv(
    `/classes/${classId}/gradebook/export`,
    `class-${classId}-gradebook.csv`
  );
}
