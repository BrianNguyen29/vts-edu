import { apiClient } from './api-client';
import { createApiError, type ListOptions, type PagedList } from './attempts';

export interface Organization {
  id: string;
  code: string;
  name: string;
}

export interface User {
  id: string;
  display_name: string;
  email: string;
  login_name: string;
  roles: string[];
  must_change_password: boolean;
}

export interface CreateUserRequest {
  login_name: string;
  display_name: string;
  email: string;
  temporary_password: string;
  roles: string[];
}

export interface UpdateRolesRequest {
  roles: string[];
}

export interface ResetPasswordRequest {
  temporary_password: string;
}

export interface UpdateOrganizationRequest {
  name: string;
}

function buildQueryString(opts: ListOptions): string {
  const params = new URLSearchParams();
  if (opts.q) params.set('q', opts.q);
  if (opts.limit !== undefined) params.set('limit', String(opts.limit));
  if (opts.offset !== undefined) params.set('offset', String(opts.offset));
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

export async function listUsers(opts: ListOptions = {}): Promise<PagedList<User>> {
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
