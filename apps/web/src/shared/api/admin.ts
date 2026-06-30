import { getOpenAPIClient } from './openapi-client';
import {
  unwrapData,
  unwrapPaged,
  unwrapVoid,
  type ListOptions,
  type PagedList,
} from './attempts';
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

function cleanListQuery(opts: ListOptions) {
  return {
    q: opts.q,
    limit: opts.limit,
    offset: opts.offset,
    cursor: opts.cursor,
    count: opts.count,
  };
}

function cleanAuditLogQuery(opts: AuditLogListOptions) {
  return {
    action: opts.action,
    actor_user_id: opts.actor_user_id,
    from: opts.from,
    to: opts.to,
    limit: opts.limit,
    offset: opts.offset,
    cursor: opts.cursor,
    count: opts.count,
  };
}

export async function getOrganization(): Promise<Organization> {
  const client = await getOpenAPIClient();
  return unwrapData<Organization>(
    await client.GET('/organizations/current')
  );
}

export async function updateOrganization(
  req: UpdateOrganizationRequest
): Promise<Organization> {
  const client = await getOpenAPIClient();
  return unwrapData<Organization>(
    await client.PATCH('/organizations/current', { body: req })
  );
}

export async function listUsers(
  opts: ListOptions = {}
): Promise<PagedList<User>> {
  const client = await getOpenAPIClient();
  return unwrapPaged<User>(
    await client.GET('/users', { params: { query: cleanListQuery(opts) } })
  );
}

export async function createUser(req: CreateUserRequest): Promise<User> {
  const client = await getOpenAPIClient();
  return unwrapData<User>(await client.POST('/users', { body: req }));
}

export async function updateUserRoles(
  userId: string,
  req: UpdateRolesRequest
): Promise<void> {
  const client = await getOpenAPIClient();
  unwrapVoid(
    await client.PUT('/users/{user_id}/roles', {
      params: { path: { user_id: userId } },
      body: req,
    })
  );
}

export async function resetUserPassword(
  userId: string,
  req: ResetPasswordRequest
): Promise<void> {
  const client = await getOpenAPIClient();
  unwrapVoid(
    await client.POST('/users/{user_id}/reset-password', {
      params: { path: { user_id: userId } },
      body: req,
    })
  );
}

export async function listAuditLogs(
  opts: AuditLogListOptions = {}
): Promise<PagedList<AuditLog>> {
  const client = await getOpenAPIClient();
  return unwrapPaged<AuditLog>(
    await client.GET('/audit-logs', {
      params: { query: cleanAuditLogQuery(opts) },
    })
  );
}
