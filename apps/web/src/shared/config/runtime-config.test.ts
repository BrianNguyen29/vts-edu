import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';

describe('loadRuntimeConfig', () => {
  beforeEach(() => {
    vi.resetModules();
    vi.stubGlobal('fetch', vi.fn());
    vi.stubEnv('VITE_API_BASE_URL', 'https://default.example.com/api/v1');
  });

  it('returns fetched config when available', async () => {
    const customConfig = {
      apiBaseUrl: 'https://api.example.com/api/v1',
      environment: 'production',
      release: 'v1.0.0',
      features: { pwaInstall: true, analyticsDashboard: false },
    };
    (globalThis.fetch as Mock).mockResolvedValue({
      ok: true,
      json: async () => customConfig,
    });

    const { loadRuntimeConfig } = await import('./runtime-config');
    const config = await loadRuntimeConfig();
    expect(config).toEqual(customConfig);
  });

  it('falls back to default config when fetch fails', async () => {
    (globalThis.fetch as Mock).mockRejectedValue(new Error('network'));

    const { loadRuntimeConfig } = await import('./runtime-config');
    const config = await loadRuntimeConfig();
    expect(config.apiBaseUrl).toBe('https://default.example.com/api/v1');
    expect(config.environment).toBe('development');
    expect(config.features).toEqual({
      pwaInstall: false,
      analyticsDashboard: false,
    });
  });

  it('falls back when response is not ok', async () => {
    (globalThis.fetch as Mock).mockResolvedValue({
      ok: false,
      status: 500,
    });

    const { loadRuntimeConfig } = await import('./runtime-config');
    const config = await loadRuntimeConfig();
    expect(config.apiBaseUrl).toBe('https://default.example.com/api/v1');
  });
});
