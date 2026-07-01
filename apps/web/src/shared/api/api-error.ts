import type { components } from './openapi-schema';

export interface ApiError {
  error: {
    code: string;
    message: string;
    request_id?: string;
  };
}

export function isApiError(value: unknown): value is ApiError {
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

export class ApiResponseError extends Error {
  readonly status: number;
  readonly code: string;
  readonly body: ApiError;
  readonly requestId?: string;

  constructor(status: number, body: ApiError) {
    super(body.error.message);
    this.name = 'ApiResponseError';
    this.status = status;
    this.code = body.error.code;
    this.body = body;
    this.requestId = body.error.request_id;
  }
}

export function createApiError(
  status: number,
  raw: unknown,
  response?: Response
): ApiResponseError {
  const body = isApiError(raw)
    ? raw
    : { error: { code: 'unknown', message: 'Yêu cầu thất bại.' } };

  const requestId =
    body.error.request_id ?? response?.headers.get('x-request-id') ?? undefined;

  return new ApiResponseError(status, {
    error: { ...body.error, request_id: requestId },
  });
}

export interface PagedList<T> {
  data: T[];
  page?: components['schemas']['PageInfo'];
}

export function unwrapData<T>(res: {
  data?: unknown;
  error?: unknown;
  response: Response;
}): T {
  if (res.error) {
    throw createApiError(res.response.status, res.error, res.response);
  }
  return (res.data as { data: T }).data;
}

export function unwrapPaged<T>(res: {
  data?: unknown;
  error?: unknown;
  response: Response;
}): PagedList<T> {
  if (res.error) {
    throw createApiError(res.response.status, res.error, res.response);
  }
  return res.data as PagedList<T>;
}

export function unwrapVoid(res: {
  error?: unknown;
  response: Response;
}): void {
  if (res.error) {
    throw createApiError(res.response.status, res.error, res.response);
  }
}

export interface ApiErrorDetails {
  status: number | null;
  code: string;
  message: string;
  requestId?: string;
  isNetwork: boolean;
}

export function getApiErrorDetails(err: unknown): ApiErrorDetails {
  if (err instanceof ApiResponseError) {
    return {
      status: err.status,
      code: err.code,
      message: err.body.error.message,
      requestId: err.requestId,
      isNetwork: false,
    };
  }

  if (err instanceof Error && err.message === 'network') {
    return {
      status: null,
      code: 'network',
      message: 'Không thể kết nối đến máy chủ. Vui lòng thử lại.',
      isNetwork: true,
    };
  }

  return {
    status: null,
    code: 'unknown',
    message: 'Đã xảy ra lỗi không mong muốn.',
    isNetwork: false,
  };
}

const DEFAULT_STATUS_MESSAGES: Record<number, string> = {
  400: 'Yêu cầu không hợp lệ.',
  401: 'Phiên làm việc đã hết hạn. Vui lòng đăng nhập lại.',
  403: 'Không có quyền truy cập.',
  404: 'Không tìm thấy tài nguyên.',
  409: 'Yêu cầu xung đột. Vui lòng thử lại.',
  429: 'Quá nhiều yêu cầu. Vui lòng đợi một chút và thử lại.',
  500: 'Máy chủ gặp sự cố. Vui lòng thử lại sau.',
  502: 'Máy chủ gặp sự cố. Vui lòng thử lại sau.',
  503: 'Máy chủ gặp sự cố. Vui lòng thử lại sau.',
  504: 'Máy chủ gặp sự cố. Vui lòng thử lại sau.',
};

export function formatFriendlyError(
  err: unknown,
  overrides?: Record<number, string>
): string {
  const details = getApiErrorDetails(err);

  if (details.isNetwork) {
    return details.message;
  }

  if (details.status !== null) {
    const override = overrides?.[details.status] ?? overrides?.[Math.floor(details.status / 100) * 100];
    if (override) return override;

    const fallback = DEFAULT_STATUS_MESSAGES[details.status];
    if (fallback) return fallback;

    // For other API errors, the backend message is considered safe to surface.
    if (details.message) return details.message;
  }

  return 'Đã xảy ra lỗi không mong muốn.';
}
