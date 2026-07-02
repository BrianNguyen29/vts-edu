import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  listNotifications,
  markRead,
  unreadCount,
  type Notification,
} from './notifications';
import { notificationKeys } from './query-keys';

const POLL_INTERVAL_MS = 30_000;

export function useNotificationsQuery(opts: { limit?: number; enabled?: boolean } = {}) {
  return useQuery<Notification[], Error>({
    queryKey: notificationKeys.list({ limit: opts.limit }),
    queryFn: () => listNotifications({ limit: opts.limit }),
    refetchInterval: POLL_INTERVAL_MS,
    refetchIntervalInBackground: true,
    enabled: opts.enabled ?? true,
  });
}

export function useUnreadCountQuery(opts: { enabled?: boolean } = {}) {
  return useQuery<number, Error>({
    queryKey: notificationKeys.unreadCount(),
    queryFn: () => unreadCount(),
    refetchInterval: POLL_INTERVAL_MS,
    refetchIntervalInBackground: true,
    enabled: opts.enabled ?? true,
  });
}

export function useMarkReadMutation() {
  const queryClient = useQueryClient();
  return useMutation<Notification, Error, string>({
    mutationFn: (id) => markRead(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: notificationKeys.all });
    },
  });
}
