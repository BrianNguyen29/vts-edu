import { describe, it, expect } from 'vitest';
import {
  ApiResponseError,
  createApiError,
  formatFriendlyError,
  getApiErrorDetails,
  isApiError,
  unwrapData,
} from './api-error';

describe('isApiError', () => {
  it('returns true for a valid API error shape', () => {
    expect(
      isApiError({ error: { code: 'not_found', message: 'Not found' } })
    ).toBe(true);
  });

  it('returns false for non-object values', () => {
    expect(isApiError(null)).toBe(false);
    expect(isApiError('error')).toBe(false);
  });

  it('returns false when code or message is missing', () => {
    expect(isApiError({ error: { message: 'only message' } })).toBe(false);
    expect(isApiError({ error: { code: 'only code' } })).toBe(false);
  });
});

describe('createApiError', () => {
  it('extracts request_id from the error body', () => {
    const err = createApiError(404, {
      error: { code: 'not_found', message: 'Không tìm thấy.', request_id: 'req-123' },
    });
    expect(err).toBeInstanceOf(ApiResponseError);
    expect(err.status).toBe(404);
    expect(err.code).toBe('not_found');
    expect(err.requestId).toBe('req-123');
  });

  it('falls back to the X-Request-ID response header', () => {
    const response = new Response(null, {
      status: 500,
      headers: { 'X-Request-ID': 'header-req-456' },
    });
    const err = createApiError(500, { error: { code: 'internal', message: 'Lỗi.' } }, response);
    expect(err.requestId).toBe('header-req-456');
  });

  it('prefers body request_id over header', () => {
    const response = new Response(null, {
      status: 403,
      headers: { 'X-Request-ID': 'header-id' },
    });
    const err = createApiError(
      403,
      { error: { code: 'forbidden', message: 'Cấm.', request_id: 'body-id' } },
      response
    );
    expect(err.requestId).toBe('body-id');
  });

  it('returns a generic error for unrecognised bodies', () => {
    const err = createApiError(500, { message: 'random' });
    expect(err.code).toBe('unknown');
    expect(err.body.error.message).toBe('Yêu cầu thất bại.');
  });
});

describe('unwrapData', () => {
  it('returns data when there is no error', () => {
    const response = new Response();
    expect(unwrapData({ data: { data: { id: '1' } }, response })).toEqual({ id: '1' });
  });

  it('throws ApiResponseError on error', () => {
    const response = new Response();
    expect(() =>
      unwrapData({
        error: { error: { code: 'bad_request', message: 'Sai.' } },
        response,
      })
    ).toThrow(ApiResponseError);
  });
});

describe('getApiErrorDetails', () => {
  it('returns details for an ApiResponseError', () => {
    const err = new ApiResponseError(403, {
      error: { code: 'forbidden', message: 'Cấm.', request_id: 'req-789' },
    });
    expect(getApiErrorDetails(err)).toEqual({
      status: 403,
      code: 'forbidden',
      message: 'Cấm.',
      requestId: 'req-789',
      isNetwork: false,
    });
  });

  it('detects network errors', () => {
    expect(getApiErrorDetails(new Error('network'))).toMatchObject({
      status: null,
      code: 'network',
      isNetwork: true,
    });
  });

  it('returns a safe fallback for unknown errors', () => {
    expect(getApiErrorDetails(new Error('unexpected internals'))).toMatchObject({
      status: null,
      code: 'unknown',
      message: 'Đã xảy ra lỗi không mong muốn.',
      isNetwork: false,
    });
  });
});

describe('formatFriendlyError', () => {
  it('uses per-status default messages', () => {
    expect(formatFriendlyError(new ApiResponseError(401, genericError()))).toContain(
      'đăng nhập lại'
    );
    expect(formatFriendlyError(new ApiResponseError(403, genericError()))).toBe(
      'Không có quyền truy cập.'
    );
    expect(formatFriendlyError(new ApiResponseError(500, genericError()))).toContain(
      'Máy chủ gặp sự cố'
    );
  });

  it('applies overrides by status', () => {
    expect(
      formatFriendlyError(new ApiResponseError(404, genericError()), {
        404: 'Không tìm thấy bài kiểm tra.',
      })
    ).toBe('Không tìm thấy bài kiểm tra.');
  });

  it('returns a network message for network errors', () => {
    expect(formatFriendlyError(new Error('network'))).toContain('Không thể kết nối');
  });

  it('does not leak unknown raw messages', () => {
    expect(formatFriendlyError(new Error('secret internal trace'))).toBe(
      'Đã xảy ra lỗi không mong muốn.'
    );
  });
});

function genericError() {
  return { error: { code: 'generic', message: 'Generic error.' } };
}
