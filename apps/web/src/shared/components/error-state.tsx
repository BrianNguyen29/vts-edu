import { useState } from 'react';
import {
  formatFriendlyError,
  getApiErrorDetails,
} from '@/shared/api/api-error';

export interface ErrorStateProps {
  error: unknown;
  title?: string;
  /** Override the computed friendly message while still showing request_id. */
  message?: string;
  onRetry?: () => void;
  overrides?: Record<number, string>;
  className?: string;
}

export function ErrorState({
  error,
  title,
  message: messageOverride,
  onRetry,
  overrides,
  className = 'error-banner',
}: ErrorStateProps) {
  const details = getApiErrorDetails(error);
  const message = messageOverride ?? formatFriendlyError(error, overrides);
  const [copied, setCopied] = useState(false);

  async function copyRequestId() {
    if (!details.requestId) return;
    try {
      await navigator.clipboard.writeText(details.requestId);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // Ignore clipboard failures (e.g., denied permission).
    }
  }

  return (
    <div className={className} role="alert" data-testid="error-state">
      {title && <h2 className="error-state-title">{title}</h2>}
      <p data-testid="error-message">{message}</p>
      {details.requestId && (
        <div className="error-request-id" data-testid="error-request-id">
          <span className="error-request-id-label">Mã yêu cầu:</span>
          <code className="error-request-id-code">{details.requestId}</code>
          <button
            type="button"
            onClick={copyRequestId}
            className="secondary"
            data-testid="error-copy-button"
          >
            {copied ? 'Đã sao chép' : 'Sao chép'}
          </button>
        </div>
      )}
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="secondary"
          data-testid="error-retry-button"
        >
          Thử lại
        </button>
      )}
    </div>
  );
}
