import { getOpenAPIClient } from './openapi-client';
import { unwrapData } from './attempts';
import type { components } from './openapi-schema';

export type Notification = components['schemas']['Notification'];

export interface ListNotificationsOptions {
  limit?: number;
  before?: string;
}

export async function listNotifications(
  opts: ListNotificationsOptions = {}
): Promise<Notification[]> {
  const client = await getOpenAPIClient();
  return unwrapData<Notification[]>(
    await client.GET('/me/notifications', {
      params: {
        query: {
          limit: opts.limit,
          before: opts.before,
        },
      },
    })
  );
}

export async function unreadCount(): Promise<number> {
  const client = await getOpenAPIClient();
  const res = await client.GET('/me/notifications/unread-count', {});
  const data = unwrapData<{ count: number }>(res);
  return data?.count ?? 0;
}

export async function markRead(id: string): Promise<Notification> {
  const client = await getOpenAPIClient();
  return unwrapData<Notification>(
    await client.POST('/me/notifications/{id}/read', {
      params: { path: { id } },
    })
  );
}
