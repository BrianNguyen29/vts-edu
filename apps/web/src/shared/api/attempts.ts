import { getOpenAPIClient } from './openapi-client';
import type { components } from './openapi-schema';

export type AttemptSnapshot = components['schemas']['AttemptSnapshot']['data'];
export type AttemptItem = components['schemas']['AttemptItem'];
export type QuestionPrompt = components['schemas']['AttemptItem']['prompt'];
export type QuestionChoice = components['schemas']['AttemptItem']['choices'][number];
export type AnswerSnapshot = NonNullable<components['schemas']['AttemptItem']['answer']>;
export type AnswerSaved = components['schemas']['SaveAnswerResponse']['data'];
export type AttemptSubmitted = components['schemas']['AttemptSubmitted']['data'];
export type AssignedAssessment = components['schemas']['AssignedAssessment'];
export type StudentAttempt = components['schemas']['StudentAttempt'];
export type AttemptResult = components['schemas']['AttemptResult']['data'];
export type AttemptResultItem = components['schemas']['AttemptResultItem'];

import {
  ApiResponseError,
  createApiError,
  formatFriendlyError,
  getApiErrorDetails,
  isApiError,
  unwrapData,
  unwrapPaged,
  unwrapVoid,
  type ApiError,
  type ApiErrorDetails,
  type PagedList,
} from './api-error';

export {
  ApiResponseError,
  createApiError,
  formatFriendlyError,
  getApiErrorDetails,
  isApiError,
  unwrapData,
  unwrapPaged,
  unwrapVoid,
  type ApiError,
  type ApiErrorDetails,
  type PagedList,
};

export interface ListOptions {
  q?: string;
  limit?: number;
  offset?: number;
  cursor?: string;
  count?: boolean;
}

export type PageInfo = components['schemas']['PageInfo'];

export async function listAssignedAssessments(): Promise<AssignedAssessment[]> {
  const client = await getOpenAPIClient();
  return unwrapData<AssignedAssessment[]>(
    await client.GET('/me/assessments')
  );
}

export async function startAttempt(assessmentId: string): Promise<AttemptSnapshot> {
  const client = await getOpenAPIClient();
  return unwrapData<AttemptSnapshot>(
    await client.POST('/assessments/{assessment_id}/attempts', {
      params: { path: { assessment_id: assessmentId } },
    })
  );
}

export interface AttemptHistoryOptions {
  limit?: number;
  cursor?: string;
}

export async function listAttemptHistory(
  opts: AttemptHistoryOptions = {}
): Promise<PagedList<StudentAttempt>> {
  const client = await getOpenAPIClient();
  const query: { limit?: number; cursor?: string } = {};
  if (opts.limit) query.limit = opts.limit;
  if (opts.cursor) query.cursor = opts.cursor;
  return unwrapPaged<StudentAttempt>(
    await client.GET('/me/attempts', { params: { query } })
  );
}

export async function getAttemptResult(attemptId: string): Promise<AttemptResult> {
  const client = await getOpenAPIClient();
  return unwrapData<AttemptResult>(
    await client.GET('/attempts/{attempt_id}/result', {
      params: { path: { attempt_id: attemptId } },
    })
  );
}

export async function getAttempt(attemptId: string): Promise<AttemptSnapshot> {
  const client = await getOpenAPIClient();
  return unwrapData<AttemptSnapshot>(
    await client.GET('/attempts/{attempt_id}', {
      params: { path: { attempt_id: attemptId } },
    })
  );
}

export async function saveAnswer(
  attemptId: string,
  itemId: string,
  answerPayload: unknown
): Promise<AnswerSaved> {
  const client = await getOpenAPIClient();
  return unwrapData<AnswerSaved>(
    await client.PUT('/attempts/{attempt_id}/answers/{attempt_item_id}', {
      params: { path: { attempt_id: attemptId, attempt_item_id: itemId } },
      body: { answer_payload: answerPayload as { selected_option?: string } },
    })
  );
}

export async function submitAttempt(
  attemptId: string
): Promise<AttemptSubmitted> {
  const client = await getOpenAPIClient();
  return unwrapData<AttemptSubmitted>(
    await client.POST('/attempts/{attempt_id}/submit', {
      params: { path: { attempt_id: attemptId } },
    })
  );
}
