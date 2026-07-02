import {
  createBrowserRouter,
  Navigate,
  useLocation,
} from 'react-router-dom';
import { lazy, Suspense, type ReactNode } from 'react';
import { AuthLayout } from '@/app/layouts/auth-layout';
import { AppShellLayout } from '@/app/layouts/app-shell-layout';
import { ExamLayout } from '@/app/layouts/exam-layout';
import { useAuth } from '@/app/providers/auth-provider';

// Route pages are split into their own chunks via React.lazy. They are
// loaded on demand by react-router's data router so the initial JS only
// includes the layout + auth shells.
const LoginPage = lazy(() =>
  import('@/pages/login/login-page').then((m) => ({ default: m.LoginPage }))
);
const DiagnosticsPage = lazy(() =>
  import('@/pages/diagnostics/diagnostics-page').then((m) => ({
    default: m.DiagnosticsPage,
  }))
);
const DashboardPage = lazy(() =>
  import('@/pages/dashboard/dashboard-page').then((m) => ({
    default: m.DashboardPage,
  }))
);
const TeacherDashboardPage = lazy(() =>
  import('@/pages/dashboard/teacher-dashboard-page').then((m) => ({
    default: m.TeacherDashboardPage,
  }))
);
const AdminDashboardPage = lazy(() =>
  import('@/pages/dashboard/admin-dashboard-page').then((m) => ({
    default: m.AdminDashboardPage,
  }))
);
const AssessmentBuilderPage = lazy(() =>
  import('@/pages/assessment-builder/assessment-builder-page').then((m) => ({
    default: m.AssessmentBuilderPage,
  }))
);
const GradebookPage = lazy(() =>
  import('@/pages/gradebook/gradebook-page').then((m) => ({
    default: m.GradebookPage,
  }))
);
const ChangePasswordPage = lazy(() =>
  import('@/pages/change-password/change-password-page').then((m) => ({
    default: m.ChangePasswordPage,
  }))
);
const ExamPage = lazy(() =>
  import('@/pages/exam/exam-page').then((m) => ({ default: m.ExamPage }))
);
const AttemptReviewPage = lazy(() =>
  import('@/pages/attempt-review/attempt-review-page').then((m) => ({
    default: m.AttemptReviewPage,
  }))
);
const ResourcesPage = lazy(() =>
  import('@/pages/resources/resources-page').then((m) => ({
    default: m.ResourcesPage,
  }))
);
const QuestionBanksPage = lazy(() =>
  import('@/pages/question-banks/question-banks-page').then((m) => ({
    default: m.QuestionBanksPage,
  }))
);
const GradingQueuePage = lazy(() =>
  import('@/pages/grading/grading-queue-page').then((m) => ({
    default: m.GradingQueuePage,
  }))
);
const GradingDetailPage = lazy(() =>
  import('@/pages/grading/grading-detail-page').then((m) => ({
    default: m.GradingDetailPage,
  }))
);
const NotFoundPage = lazy(() =>
  import('@/pages/not-found/not-found-page').then((m) => ({
    default: m.NotFoundPage,
  }))
);
const ErrorPage = lazy(() =>
  import('@/pages/error/error-page').then((m) => ({ default: m.ErrorPage }))
);

function PageLoading() {
  return (
    <div
      className="loading-fallback"
      role="status"
      aria-live="polite"
      data-testid="page-loading"
    >
      Đang tải…
    </div>
  );
}

function SuspenseRoute({ children }: { children: ReactNode }) {
  return <Suspense fallback={<PageLoading />}>{children}</Suspense>;
}

function ProtectedRoute({ children }: { children: ReactNode }) {
  const auth = useAuth();
  const location = useLocation();

  if (auth.status === 'bootstrapping') {
    return <div className="loading-full">Đang khởi tạo phiên làm việc…</div>;
  }

  if (auth.status === 'anonymous' || auth.status === 'degraded') {
    return (
      <Navigate
        to={`/login?returnTo=${encodeURIComponent(location.pathname + location.search)}`}
        replace
      />
    );
  }

  if (
    auth.status === 'restricted' &&
    location.pathname !== '/app/change-password'
  ) {
    return <Navigate to="/app/change-password" replace />;
  }

  return children;
}

function GuestOnly({ children }: { children: ReactNode }) {
  const auth = useAuth();

  if (auth.status === 'bootstrapping') {
    return <div className="loading-full">Đang khởi tạo phiên làm việc…</div>;
  }

  if (auth.status === 'authenticated' || auth.status === 'restricted') {
    return <Navigate to="/app" replace />;
  }

  return children;
}

function LandingRedirect() {
  const auth = useAuth();

  if (auth.status === 'bootstrapping') {
    return <div className="loading-full">Đang khởi tạo phiên làm việc…</div>;
  }

  if (auth.status === 'authenticated') {
    return <Navigate to="/app" replace />;
  }

  return <Navigate to="/login" replace />;
}

function getRoleHomePath(roles: string[]): string {
  if (roles.includes('admin')) return '/app/admin';
  if (roles.includes('teacher')) return '/app/teacher';
  return '/app/student';
}

function RoleRedirect() {
  const auth = useAuth();

  if (auth.status === 'bootstrapping') {
    return <div className="loading-full">Đang khởi tạo phiên làm việc…</div>;
  }

  if (auth.status === 'restricted' && auth.actor) {
    return <Navigate to="/app/change-password" replace />;
  }

  if (auth.status !== 'authenticated' || !auth.actor) {
    return <Navigate to="/login" replace />;
  }

  return <Navigate to={getRoleHomePath(auth.actor.roles)} replace />;
}

export const router = createBrowserRouter([
  {
    path: '/',
    element: <LandingRedirect />,
  },
  {
    path: '/login',
    element: (
      <GuestOnly>
        <AuthLayout />
      </GuestOnly>
    ),
    children: [
      {
        index: true,
        element: (
          <SuspenseRoute>
            <LoginPage />
          </SuspenseRoute>
        ),
      },
    ],
  },
  {
    path: '/diagnostics',
    element: (
      <SuspenseRoute>
        <DiagnosticsPage />
      </SuspenseRoute>
    ),
  },
  {
    path: '/app',
    element: (
      <ProtectedRoute>
        <AppShellLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        index: true,
        element: <RoleRedirect />,
      },
      {
        path: 'student',
        element: (
          <SuspenseRoute>
            <DashboardPage />
          </SuspenseRoute>
        ),
      },
      {
        path: 'teacher',
        element: (
          <SuspenseRoute>
            <TeacherDashboardPage />
          </SuspenseRoute>
        ),
      },
      {
        path: 'teacher/assessments/:assessmentId',
        element: (
          <SuspenseRoute>
            <AssessmentBuilderPage />
          </SuspenseRoute>
        ),
      },
      {
        path: 'teacher/gradebook',
        element: (
          <SuspenseRoute>
            <GradebookPage />
          </SuspenseRoute>
        ),
      },
      {
        path: 'resources',
        element: (
          <SuspenseRoute>
            <ResourcesPage />
          </SuspenseRoute>
        ),
      },
      {
        path: 'question-banks',
        element: (
          <SuspenseRoute>
            <QuestionBanksPage />
          </SuspenseRoute>
        ),
      },
      {
        path: 'grading',
        element: (
          <SuspenseRoute>
            <GradingQueuePage />
          </SuspenseRoute>
        ),
      },
      {
        path: 'grading/:attemptId',
        element: (
          <SuspenseRoute>
            <GradingDetailPage />
          </SuspenseRoute>
        ),
      },
      {
        path: 'admin',
        element: (
          <SuspenseRoute>
            <AdminDashboardPage />
          </SuspenseRoute>
        ),
      },
      {
        path: 'change-password',
        element: (
          <SuspenseRoute>
            <ChangePasswordPage />
          </SuspenseRoute>
        ),
      },
    ],
  },
  {
    path: '/exam/attempts/:attemptId',
    element: (
      <ProtectedRoute>
        <ExamLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        index: true,
        element: (
          <SuspenseRoute>
            <ExamPage />
          </SuspenseRoute>
        ),
      },
    ],
  },
  {
    path: '/attempts/:attemptId/result',
    element: (
      <ProtectedRoute>
        <ExamLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        index: true,
        element: (
          <SuspenseRoute>
            <AttemptReviewPage />
          </SuspenseRoute>
        ),
      },
    ],
  },
  {
    path: '/error/:status?',
    element: (
      <SuspenseRoute>
        <ErrorPage />
      </SuspenseRoute>
    ),
  },
  {
    path: '*',
    element: (
      <SuspenseRoute>
        <NotFoundPage />
      </SuspenseRoute>
    ),
  },
]);
