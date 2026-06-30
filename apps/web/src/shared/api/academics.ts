import { ApiResponseError, createApiError } from './attempts';
import { getOpenAPIClient } from './openapi-client';
import type { components } from './openapi-schema';

export type ClassSection = components['schemas']['ClassSection']['data'];
export type Enrollment = components['schemas']['Enrollment']['data'];
export type ClassSectionList = components['schemas']['ClassSectionList'];
export type EnrollmentList = components['schemas']['EnrollmentList'];

export async function listClasses(): Promise<ClassSectionList> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.GET('/classes');
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data;
}

export async function listEnrollments(
  classId: string
): Promise<EnrollmentList> {
  const client = await getOpenAPIClient();
  const { data, error, response } = await client.GET(
    '/classes/{class_id}/enrollments',
    {
      params: {
        path: {
          class_id: classId,
        },
      },
    }
  );
  if (!data || error) {
    throw createApiError(response.status, error ?? {});
  }
  return data;
}

export { ApiResponseError };
