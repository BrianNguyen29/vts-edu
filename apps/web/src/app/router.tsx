import { createBrowserRouter, Navigate, useLocation } from 'react-router-dom';
import { AuthLayout } from '@/app/layouts/auth-layout';
import { AppShellLayout } from '@/app/layouts/app-shell-layout';
import { ExamLayout } from '@/app/layouts/exam-layout';
import { LoginPage } from '@/pages/login/login-page';
import { DiagnosticsPage } from '@/pages/diagnostics/diagnostics-page';
import { DashboardPage } from '@/pages/dashboard/dashboard-page';
import { TeacherDashboardPage } from '@/pages/dashboard/teacher-dashboard-page';
import { AdminDashboardPage } from '@/pages/dashboard/admin-dashboard-page';
import { AssessmentBuilderPage } from '@/pages/assessment-builder/assessment-builder-page';
import { GradebookPage } from '@/pages/gradebook/gradebook-page';
import { ChangePasswordPage } from '@/pages/change-password/change-password-page';
import { ExamPage } from '@/pages/exam/exam-page';
import { AttemptReviewPage } from '@/pages/attempt-review/attempt-review-page';
import { ResourcesPage } from '@/pages/resources/resources-page';
import { QuestionBanksPage } from '@/pages/question-banks/question-banks-page';
import { GradingQueuePage } from '@/pages/grading/grading-queue-page';
import { GradingDetailPage } from '@/pages/grading/grading-detail-page';
import { NotFoundPage } from '@/pages/not-found/not-found-page';
import { ErrorPage } from '@/pages/error/error-page';
import { useAuth } from '@/app/providers/auth-provider';
import type { ReactNode } from 'react';

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
        element: <LoginPage />,
      },
    ],
  },
  {
    path: '/diagnostics',
    element: <DiagnosticsPage />,
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
        element: <DashboardPage />,
      },
      {
        path: 'teacher',
        element: <TeacherDashboardPage />,
      },
      {
        path: 'teacher/assessments/:assessmentId',
        element: <AssessmentBuilderPage />,
      },
      {
        path: 'teacher/gradebook',
        element: <GradebookPage />,
      },
      {
        path: 'resources',
        element: <ResourcesPage />,
      },
      {
        path: 'question-banks',
        element: <QuestionBanksPage />,
      },
      {
        path: 'grading',
        element: <GradingQueuePage />,
      },
      {
        path: 'grading/:attemptId',
        element: <GradingDetailPage />,
      },
      {
        path: 'admin',
        element: <AdminDashboardPage />,
      },
      {
        path: 'change-password',
        element: <ChangePasswordPage />,
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
        element: <ExamPage />,
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
        element: <AttemptReviewPage />,
      },
    ],
  },
  {
    path: '/error/:status?',
    element: <ErrorPage />,
  },
  {
    path: '*',
    element: <NotFoundPage />,
  },
]);
