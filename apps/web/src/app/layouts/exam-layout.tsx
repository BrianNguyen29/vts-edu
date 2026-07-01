import { Link, Outlet } from 'react-router-dom';

export function ExamLayout() {
  return (
    <div className="exam-layout">
      <a href="#main-content" className="skip-link">
        Bỏ qua đến nội dung chính
      </a>
      <header className="exam-header" role="banner">
        <Link to="/app" className="exam-back" aria-label="Về trang làm việc">
          ← Về trang làm việc
        </Link>
        <h1>Bài thi</h1>
      </header>
      <main id="main-content" className="exam-main" tabIndex={-1}>
        <Outlet />
      </main>
    </div>
  );
}
