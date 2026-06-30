import { apiClient } from './api-client';

export interface AttemptSnapshot {
  id: string;
  organization_id: string;
  assessment_id: string;
  publication_id?: string;
  status: 'CREATED' | 'IN_PROGRESS' | 'SUBMITTED' | 'EXPIRED';
  started_at?: string;
  expires_at?: string;
  submitted_at?: string;
  items: AttemptItem[];
}

export interface AttemptItem {
  id: string;
  question_version_id: string;
  position: number;
  points: string;
  prompt?: QuestionPrompt;
  choices?: QuestionChoice[];
  answer?: AnswerSnapshot;
}

export interface QuestionPrompt {
  text?: string;
}

export interface QuestionChoice {
  id: string;
  text?: string;
}

export interface AnswerSnapshot {
  answer_payload: unknown;
  revision: number;
  answered_at: string;
}

export interface AnswerSaved {
  attempt_item_id: string;
  revision: number;
  answer_payload: unknown;
  answered_at: string;
}

export interface AttemptSubmitted {
  id: string;
  status: string;
  submitted_at: string;
  score?: string;
  max_score?: string;
  grading_status?: string;
}

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
}

export interface PageInfo {
  limit: number;
  offset: number;
}

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
