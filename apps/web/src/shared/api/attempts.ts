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

export interface ApiError {
  error: {
    code: string;
    message: string;
  };
}

export interface ListOptions {
  q?: string;
  limit?: number;
  offset?: number;
  cursor?: string;
  count?: boolean;
}

export type PageInfo = components['schemas']['PageInfo'];

export interface PagedList<T> {
  data: T[];
  page?: PageInfo;
}

export class ApiResponseError extends Error {
  constructor(
    public readonly status: number,
    public readonly body: ApiError
  ) {
    super(body.error.message);
  }
}

function isApiError(value: unknown): value is ApiError {
  return (
    typeof value === 'object' &&
    value !== null &&
    'error' in value &&
    typeof (value as ApiError).error === 'object' &&
    (value as ApiError).error !== null &&
    typeof (value as ApiError).error.code === 'string' &&
    typeof (value as ApiError).error.message === 'string'
  );
}

export function createApiError(status: number, raw: unknown): ApiResponseError {
  if (isApiError(raw)) {
    return new ApiResponseError(status, raw);
  }
  return new ApiResponseError(status, {
    error: { code: 'unknown', message: 'Yêu cầu thất bại.' },
  });
}

export function unwrapData<T>(res: {
  data?: unknown;
  error?: unknown;
  response: Response;
}): T {
  if (res.error) {
    throw createApiError(res.response.status, res.error);
  }
  return (res.data as { data: T }).data;
}

export function unwrapPaged<T>(res: {
  data?: unknown;
  error?: unknown;
  response: Response;
}): PagedList<T> {
  if (res.error) {
    throw createApiError(res.response.status, res.error);
  }
  return res.data as PagedList<T>;
}

export function unwrapVoid(res: {
  error?: unknown;
  response: Response;
}): void {
  if (res.error) {
    throw createApiError(res.response.status, res.error);
  }
}

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

export async function listAttemptHistory(): Promise<StudentAttempt[]> {
  const client = await getOpenAPIClient();
  return unwrapData<StudentAttempt[]>(
    await client.GET('/me/attempts')
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
