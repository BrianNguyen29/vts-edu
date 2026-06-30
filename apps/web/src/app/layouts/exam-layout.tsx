import { Link, Outlet } from 'react-router-dom';

export function ExamLayout() {
  return (
    <div className="exam-layout">
      <header className="exam-header">
        <Link to="/app" className="exam-back">
          ← Về trang làm việc
        </Link>
        <h1>Bài thi</h1>
      </header>
      <main className="exam-main">
        <Outlet />
      </main>
    </div>
  );
}
