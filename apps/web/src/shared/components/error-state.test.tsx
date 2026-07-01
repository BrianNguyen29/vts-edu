import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ErrorState } from './error-state';
import { ApiResponseError } from '@/shared/api/api-error';

describe('ErrorState', () => {
  const writeText = vi.fn().mockResolvedValue(undefined);

  beforeEach(() => {
    writeText.mockClear();
    Object.assign(navigator, { clipboard: { writeText } });
  });

  it('renders a safe friendly message and request id', () => {
    const error = new ApiResponseError(403, {
      error: {
        code: 'forbidden',
        message: 'Backend message.',
        request_id: 'req-state-123',
      },
    });

    render(<ErrorState error={error} />);

    expect(screen.getByTestId('error-message')).toHaveTextContent(
      'Không có quyền truy cập.'
    );
    expect(screen.getByTestId('error-request-id')).toHaveTextContent(
      'req-state-123'
    );
  });

  it('copies the request id when the copy button is clicked', async () => {
    const error = new ApiResponseError(500, {
      error: {
        code: 'internal',
        message: 'Lỗi.',
        request_id: 'req-copy-456',
      },
    });

    render(<ErrorState error={error} />);
    await userEvent.click(screen.getByTestId('error-copy-button'));

    expect(writeText).toHaveBeenCalledWith('req-copy-456');
  });

  it('does not render request id when unavailable', () => {
    const error = new ApiResponseError(404, {
      error: { code: 'not_found', message: 'Không thấy.' },
    });

    render(<ErrorState error={error} />);

    expect(screen.queryByTestId('error-request-id')).not.toBeInTheDocument();
  });

  it('uses the title prop', () => {
    const error = new ApiResponseError(429, {
      error: { code: 'rate_limit', message: 'Chậm.' },
    });

    render(<ErrorState error={error} title="Quá tải" />);

    expect(screen.getByRole('heading', { name: 'Quá tải' })).toBeInTheDocument();
  });

  it('uses the message override prop while still showing request id', () => {
    const error = new ApiResponseError(409, {
      error: {
        code: 'conflict',
        message: 'Conflict.',
        request_id: 'req-override',
      },
    });

    render(<ErrorState error={error} message="Bạn đã hết số lần làm bài." />);

    expect(screen.getByTestId('error-message')).toHaveTextContent(
      'Bạn đã hết số lần làm bài.'
    );
    expect(screen.getByTestId('error-request-id')).toHaveTextContent(
      'req-override'
    );
  });

  it('calls onRetry when the retry button is clicked', async () => {
    const onRetry = vi.fn();
    const error = new Error('network');

    render(<ErrorState error={error} onRetry={onRetry} />);
    await userEvent.click(screen.getByTestId('error-retry-button'));

    expect(onRetry).toHaveBeenCalledTimes(1);
  });
});
