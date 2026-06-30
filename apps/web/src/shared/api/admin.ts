import { apiClient } from './api-client';
import { createApiError, type ListOptions, type PagedList } from './attempts';
import type { components } from './openapi-schema';

export type Organization = components['schemas']['Organization']['data'];
export type User = components['schemas']['User'];
export type CreateUserRequest = components['schemas']['CreateUserRequest'];
export type UpdateRolesRequest = components['schemas']['UpdateRolesRequest'];
export type ResetPasswordRequest = components['schemas']['ResetPasswordRequest'];
export type UpdateOrganizationRequest = components['schemas']['UpdateOrganizationRequest'];
export type AuditLog = components['schemas']['AuditLog'];

export interface AuditLogListOptions {
  action?: string;
  actor_user_id?: string;
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
  cursor?: string;
  count?: boolean;
}

function buildQueryString(opts: ListOptions): string {
  const params = new URLSearchParams();
  if (opts.q) params.set('q', opts.q);
  if (opts.limit !== undefined) params.set('limit', String(opts.limit));
  if (opts.offset !== undefined) params.set('offset', String(opts.offset));
  if (opts.cursor) params.set('cursor', opts.cursor);
  if (opts.count) params.set('count', 'true');
  const query = params.toString();
  return query ? `?${query}` : '';
}

function buildAuditLogQueryString(opts: AuditLogListOptions): string {
  const params = new URLSearchParams();
  if (opts.action) params.set('action', opts.action);
  if (opts.actor_user_id) params.set('actor_user_id', opts.actor_user_id);
  if (opts.from) params.set('from', opts.from);
  if (opts.to) params.set('to', opts.to);
  if (opts.limit !== undefined) params.set('limit', String(opts.limit));
  if (opts.offset !== undefined) params.set('offset', String(opts.offset));
  if (opts.cursor) params.set('cursor', opts.cursor);
  if (opts.count) params.set('count', 'true');
  const query = params.toString();
  return query ? `?${query}` : '';
}

export async function getOrganization(): Promise<Organization> {
  const res = await apiClient('/organizations/current');
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: Organization }).data;
}

export async function updateOrganization(
  req: UpdateOrganizationRequest
): Promise<Organization> {
  const res = await apiClient('/organizations/current', {
    method: 'PATCH',
    body: JSON.stringify(req),
  });
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: Organization }).data;
}

export async function listUsers(
  opts: ListOptions = {}
): Promise<PagedList<User>> {
  const res = await apiClient(`/users${buildQueryString(opts)}`);
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return json as PagedList<User>;
}

export async function createUser(req: CreateUserRequest): Promise<User> {
  const res = await apiClient('/users', {
    method: 'POST',
    body: JSON.stringify(req),
  });
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return (json as { data: User }).data;
}

export async function updateUserRoles(
  userId: string,
  req: UpdateRolesRequest
): Promise<void> {
  const res = await apiClient(`/users/${userId}/roles`, {
    method: 'PUT',
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const json = (await res.json()) as unknown;
    throw createApiError(res.status, json);
  }
}

export async function resetUserPassword(
  userId: string,
  req: ResetPasswordRequest
): Promise<void> {
  const res = await apiClient(`/users/${userId}/reset-password`, {
    method: 'POST',
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const json = (await res.json()) as unknown;
    throw createApiError(res.status, json);
  }
}

export async function listAuditLogs(
  opts: AuditLogListOptions = {}
): Promise<PagedList<AuditLog>> {
  const res = await apiClient(`/audit-logs${buildAuditLogQueryString(opts)}`);
  const json = (await res.json()) as unknown;
  if (!res.ok) {
    throw createApiError(res.status, json);
  }
  return json as PagedList<AuditLog>;
}
