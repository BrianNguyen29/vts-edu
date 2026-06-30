/**
 * In-memory access token storage.
 *
 * The refresh token lives in an HttpOnly cookie and is never readable here.
 */
let accessToken: string | null = null;

export function getAccessToken(): string | null {
  return accessToken;
}

export function setAccessToken(token: string | null): void {
  accessToken = token;
}

export function clearAccessToken(): void {
  accessToken = null;
}

export interface CurrentActor {
  id: string;
  organizationId: string;
  displayName: string;
  roles: string[];
  permissions: string[];
  mustChangePassword: boolean;
}

export type AuthStatus =
  | 'bootstrapping'
  | 'authenticated'
  | 'anonymous'
  | 'restricted'
  | 'degraded';

export interface AuthSession {
  status: AuthStatus;
  accessToken: string | null;
  actor: CurrentActor | null;
  error: string | null;
}

export interface LoginCredentials {
  organizationCode: string;
  username: string;
  password: string;
}
