import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState, type ReactNode } from 'react';
import { ApiResponseError } from '@/shared/api/attempts';

function shouldRetry(error: unknown): boolean {
  if (error instanceof ApiResponseError && error.status >= 500) {
    return true;
  }
  if (error instanceof Error && error.message === 'network') {
    return true;
  }
  return false;
}

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 0,
        refetchOnWindowFocus: false,
        retry: (failureCount, error) =>
          failureCount < 1 && shouldRetry(error),
      },
      mutations: {
        retry: false,
      },
    },
  });
}

interface QueryProviderProps {
  children: ReactNode;
}

export function QueryProvider({ children }: QueryProviderProps) {
  const [queryClient] = useState(createQueryClient);
  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}
