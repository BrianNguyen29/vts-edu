import { Link } from 'react-router-dom';

export function NotFoundPage() {
  return (
    <main className="not-found-page">
      <h1>404</h1>
      <p>Trang bạn tìm không tồn tại.</p>
      <Link to="/">Về trang chính</Link>
    </main>
  );
}
