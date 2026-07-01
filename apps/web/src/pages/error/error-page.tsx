import { Link, useParams, useSearchParams } from 'react-router-dom';
import { ErrorState } from '@/shared/components/error-state';
import { ApiResponseError } from '@/shared/api/api-error';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

const STATUS_TITLES: Record<number, string> = {
  403: 'Truy cập bị từ chối',
  429: 'Quá nhiều yêu cầu',
  500: 'Lỗi máy chủ',
  502: 'Lỗi máy chủ',
  503: 'Lỗi máy chủ',
  504: 'Lỗi máy chủ',
};

function buildSyntheticError(
  status: number | undefined,
  requestId: string | undefined
): unknown {
  if (!status) {
    return new Error('unknown');
  }
  return new ApiResponseError(status, {
    error: {
      code: 'http_error',
      message: '',
      request_id: requestId,
    },
  });
}

export function ErrorPage() {
  const { status } = useParams<{ status?: string }>();
  const [searchParams] = useSearchParams();
  const requestId = searchParams.get('requestId') ?? undefined;

  useDocumentTitle('Lỗi');

  const statusCode = status ? parseInt(status, 10) : undefined;
  const error = buildSyntheticError(statusCode, requestId);
  const title = statusCode ? STATUS_TITLES[statusCode] : 'Đã xảy ra lỗi';

  return (
    <main className="error-page" data-testid="error-page">
      <h1 className="visually-hidden">Đã xảy ra lỗi</h1>
      <ErrorState error={error} title={title} />
      <Link to="/">Về trang chính</Link>
    </main>
  );
}
