import { RouterProvider } from 'react-router-dom';
import { AuthProvider } from '@/app/providers/auth-provider';
import { router } from '@/app/router';

export function App() {
  return (
    <AuthProvider>
      <RouterProvider router={router} />
    </AuthProvider>
  );
}
