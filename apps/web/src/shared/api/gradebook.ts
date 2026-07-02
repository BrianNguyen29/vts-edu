import { getOpenAPIClient } from './openapi-client';
import { unwrapData, createApiError } from './attempts';
import type { components } from './openapi-schema';

export type AssessmentAttempt = components['schemas']['AssessmentAttempt'];
export type AssessmentResult =
  components['schemas']['AssessmentResult']['data'];
export type ClassGradebookEntry = components['schemas']['ClassGradebookEntry'];

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

async function downloadCsv(
  response: Response,
  filename: string
): Promise<void> {
  if (!response.ok) {
    let raw: unknown;
    try {
      raw = await response.json();
    } catch {
      raw = null;
    }
    throw createApiError(response.status, raw, response);
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
  const client = await getOpenAPIClient();
  const { response } = await client.GET(
    '/assessments/{id}/attempts/export',
    {
      params: { path: { id: assessmentId } },
    }
  );
  return downloadCsv(response, `assessment-${assessmentId}-attempts.csv`);
}

export async function exportClassGradebookCSV(classId: string): Promise<void> {
  const client = await getOpenAPIClient();
  const { response } = await client.GET('/classes/{class_id}/gradebook/export', {
    params: { path: { class_id: classId } },
  });
  return downloadCsv(response, `class-${classId}-gradebook.csv`);
}
