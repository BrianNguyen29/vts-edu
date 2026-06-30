import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react';
import { apiClient } from '@/shared/api/api-client';
import {
  clearAccessToken,
  getAccessToken,
  setAccessToken,
  type AuthSession,
  type CurrentActor,
  type LoginCredentials,
} from '@/shared/auth/auth-session-store';

interface AuthContextValue extends AuthSession {
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<boolean>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

let inFlightRefresh: Promise<boolean> | null = null;

async function performRefresh(): Promise<boolean> {
  try {
    const res = await apiClient('/auth/refresh', {
      method: 'POST',
      // The refresh endpoint itself must not trigger a recursive refresh.
      skipCsrf: false,
    });

    if (res.status === 200) {
      const json = (await res.json()) as {
        data: { access_token: string; expires_in?: number };
      };
      setAccessToken(json.data.access_token);
      return true;
    }

    if (res.status === 401) {
      clearAccessToken();
      return false;
    }

    throw new Error(`refresh failed: ${res.status}`);
  } catch (err) {
    if (err instanceof TypeError) {
      throw new Error('network');
    }
    throw err;
  }
}

/**
 * Serialize refresh attempts across tabs and within the same JS context.
 *
 * Uses Web Locks API when available; falls back to an in-memory single-flight
 * promise for browsers that do not support locks.
 */
async function serializedRefresh(): Promise<boolean> {
  if (inFlightRefresh) return inFlightRefresh;

  const run = async (): Promise<boolean> => {
    if ('locks' in navigator) {
      try {
        return await navigator.locks.request('vts-auth-refresh', async () =>
          performRefresh()
        );
      } catch {
        return performRefresh();
      }
    }
    return performRefresh();
  };

  inFlightRefresh = run().finally(() => {
    inFlightRefresh = null;
  });
  return inFlightRefresh;
}

async function fetchActor(): Promise<CurrentActor> {
  const res = await apiClient('/me');
  if (!res.ok) {
    throw new Error(`me failed: ${res.status}`);
  }
  const json = (await res.json()) as {
    data: {
      id: string;
      organization_id: string;
      display_name: string;
      roles: string[];
      permissions: string[];
    };
  };
  const data = json.data;
  return {
    id: data.id,
    organizationId: data.organization_id,
    displayName: data.display_name,
    roles: data.roles,
    permissions: data.permissions,
    mustChangePassword: false,
  };
}

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [session, setSession] = useState<AuthSession>({
    status: 'bootstrapping',
    accessToken: getAccessToken(),
    actor: null,
    error: null,
  });

  const applyActor = useCallback((actor: CurrentActor) => {
    if (actor.mustChangePassword) {
      setSession({
        status: 'restricted',
        accessToken: getAccessToken(),
        actor,
        error: null,
      });
    } else {
      setSession({
        status: 'authenticated',
        accessToken: getAccessToken(),
        actor,
        error: null,
      });
    }
  }, []);

  const bootstrap = useCallback(async () => {
    try {
      const ok = await serializedRefresh();
      if (!ok) {
        setSession({
          status: 'anonymous',
          accessToken: null,
          actor: null,
          error: null,
        });
        return;
      }
      const actor = await fetchActor();
      applyActor(actor);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'unknown';
      // While the backend is still being implemented, treat bootstrap failures
      // as anonymous so the login screen remains usable. A production build
      // would surface `degraded` with a retry option.
      setSession({
        status: 'anonymous',
        accessToken: null,
        actor: null,
        error: message,
      });
    }
  }, [applyActor]);

  useEffect(() => {
    void bootstrap();
  }, [bootstrap]);

  const login = useCallback(async (credentials: LoginCredentials) => {
    const res = await apiClient('/auth/login', {
      method: 'POST',
      body: JSON.stringify({
        organization_code: credentials.organizationCode,
        username: credentials.username,
        password: credentials.password,
      }),
    });

    if (!res.ok) {
      if (res.status === 401 || res.status === 403) {
        throw new Error('invalid');
      }
      if (res.status === 429) {
        throw new Error('rate-limit');
      }
      throw new Error(`login failed: ${res.status}`);
    }

    const json = (await res.json()) as {
      data: {
        access_token: string;
        expires_in: number;
        user: { id: string; display_name: string };
      };
    };
    setAccessToken(json.data.access_token);

    const actor = await fetchActor();
    if (actor.mustChangePassword) {
      setSession({
        status: 'restricted',
        accessToken: getAccessToken(),
        actor,
        error: null,
      });
    } else {
      setSession({
        status: 'authenticated',
        accessToken: getAccessToken(),
        actor,
        error: null,
      });
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await apiClient('/auth/logout', { method: 'POST' });
    } catch {
      // Ignore network failures; local session is cleared regardless.
    } finally {
      clearAccessToken();
      setSession({
        status: 'anonymous',
        accessToken: null,
        actor: null,
        error: null,
      });
    }
  }, []);

  const refresh = useCallback(async () => {
    try {
      const ok = await serializedRefresh();
      if (!ok) {
        setSession({
          status: 'anonymous',
          accessToken: null,
          actor: null,
          error: null,
        });
      }
      return ok;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'unknown';
      setSession((prev) => ({ ...prev, error: message }));
      return false;
    }
  }, []);

  return (
    <AuthContext.Provider value={{ ...session, login, logout, refresh }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used inside AuthProvider');
  }
  return ctx;
}
