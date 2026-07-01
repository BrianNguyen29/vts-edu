import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E configuration for the VTS EDU web app.
 *
 * - Chromium only (no multi-browser matrix to keep CI/local cost low).
 * - Base URL points to the Vite dev server on 127.0.0.1:5173.
 * - webServer starts Vite automatically; the DB/API are assumed to be started
 *   by the caller (e.g. `pnpm e2e:browser` via `scripts/e2e_browser.sh`).
 * - Tests live in `apps/web/e2e`.
 */
export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL: 'http://127.0.0.1:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: 'pnpm dev',
    url: 'http://127.0.0.1:5173',
    reuseExistingServer: true,
    timeout: 30_000,
  },
  expect: {
    timeout: 10_000,
  },
  timeout: 60_000,
});
