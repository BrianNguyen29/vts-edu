import { Link } from 'react-router-dom';
import { useDocumentTitle } from '@/shared/lib/use-document-title';

export function NotFoundPage() {
  useDocumentTitle('Không tìm thấy trang');

  return (
    <main className="not-found-page" data-testid="not-found-page">
      <h1>404</h1>
      <p>Trang bạn tìm không tồn tại.</p>
      <Link to="/">Về trang chính</Link>
    </main>
  );
}
