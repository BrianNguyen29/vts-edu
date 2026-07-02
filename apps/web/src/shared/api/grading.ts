import { getOpenAPIClient } from './openapi-client';
import { unwrapData } from './api-error';
import type { components } from './openapi-schema';

export type ReviewQueueEntry = components['schemas']['ReviewQueueEntry'];
export type AttemptGradingContext =
  components['schemas']['AttemptGradingContext']['data'];
export type GradingItemDetail = components['schemas']['GradingItemDetail'];
export type GradingItemGrade = components['schemas']['GradingItemGrade'];
export type GradingStudentAnswer = components['schemas']['GradingStudentAnswer'];
export type GradeItemRequest = components['schemas']['GradeItemRequest'];
export type GradeItemResponse =
  components['schemas']['GradeItemResponse']['data'];

export async function listReviewQueue(
  assessmentId: string
): Promise<ReviewQueueEntry[]> {
  const client = await getOpenAPIClient();
  return unwrapData<ReviewQueueEntry[]>(
    await client.GET('/assessments/{id}/review-queue', {
      params: { path: { id: assessmentId } },
    })
  );
}

export async function getAttemptForReview(
  attemptId: string
): Promise<AttemptGradingContext> {
  const client = await getOpenAPIClient();
  return unwrapData<AttemptGradingContext>(
    await client.GET('/attempts/{attempt_id}/review', {
      params: { path: { attempt_id: attemptId } },
    })
  );
}

export async function gradeAttemptItem(
  attemptId: string,
  itemId: string,
  payload: GradeItemRequest
): Promise<GradeItemResponse> {
  const client = await getOpenAPIClient();
  return unwrapData<GradeItemResponse>(
    await client.PUT(
      '/attempts/{attempt_id}/items/{item_id}/grade',
      {
        params: { path: { attempt_id: attemptId, item_id: itemId } },
        body: payload,
      }
    )
  );
}
