import { apiClient } from './api-client';
import type { components } from './openapi-schema';

export type AttemptSnapshot = components['schemas']['AttemptSnapshot']['data'];
export type AttemptItem = components['schemas']['AttemptItem'];
export type QuestionPrompt = components['schemas']['AttemptItem']['prompt'];
export type QuestionChoice = components['schemas']['AttemptItem']['choices'][number];
export type AnswerSnapshot = NonNullable<components['schemas']['AttemptItem']['answer']>;
export type AnswerSaved = components['schemas']['SaveAnswerResponse']['data'];
export type AttemptSubmitted = components['schemas']['AttemptSubmitted']['data'];

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

export async function getAttempt(attemptId: string): Promise<AttemptSnapshot> {
  const res = await apiClient(`/attempts/${attemptId}`);
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: AttemptSnapshot }).data;
}

export async function saveAnswer(
  attemptId: string,
  itemId: string,
  answerPayload: unknown
): Promise<AnswerSaved> {
  const res = await apiClient(`/attempts/${attemptId}/answers/${itemId}`, {
    method: 'PUT',
    body: JSON.stringify({ answer_payload: answerPayload }),
  });
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: AnswerSaved }).data;
}

export async function submitAttempt(
  attemptId: string
): Promise<AttemptSubmitted> {
  const res = await apiClient(`/attempts/${attemptId}/submit`, {
    method: 'POST',
  });
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: AttemptSubmitted }).data;
}
