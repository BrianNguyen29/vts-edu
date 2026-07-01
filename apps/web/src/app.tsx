import { RouterProvider } from 'react-router-dom';
import { AuthProvider } from '@/app/providers/auth-provider';
import { QueryProvider } from '@/app/providers/query-provider';
import { router } from '@/app/router';

export function App() {
  return (
    <QueryProvider>
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </QueryProvider>
  );
}
