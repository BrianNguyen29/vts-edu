export interface RuntimeConfig {
  apiBaseUrl: string;
  environment: string;
  release: string;
  features: {
    pwaInstall: boolean;
    analyticsDashboard: boolean;
  };
}

const defaultConfig: RuntimeConfig = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  environment: 'development',
  release: 'dev',
  features: {
    pwaInstall: false,
    analyticsDashboard: false,
  },
};

let cachedConfig: RuntimeConfig | null = null;

export async function loadRuntimeConfig(): Promise<RuntimeConfig> {
  if (cachedConfig) return cachedConfig;

  try {
    const res = await fetch('/app-config.json');
    if (!res.ok) throw new Error('failed to load runtime config');
    const data = (await res.json()) as RuntimeConfig;
    cachedConfig = data;
    return data;
  } catch {
    cachedConfig = defaultConfig;
    return defaultConfig;
  }
}
