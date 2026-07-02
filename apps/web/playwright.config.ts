import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E configuration for the VTS EDU web app.
 *
 * Default project is Chromium so `pnpm e2e:browser` stays a fast local
 * path. Set `PLAYWRIGHT_BROWSERS=1` to enable the full matrix
 * (Chromium + Firefox + WebKit) — typically from CI or a manual cross
 * browser run via `pnpm e2e:browser:all`.
 *
 * WebKit additionally needs system libs that are not always present
 * (libgtk-4, libgraphene-1.0, libxslt, libevent-2.1, libopus,
 * libgstallocators, …). The script that drives the matrix reports the
 * missing libraries and skips the project gracefully instead of
 * failing the whole run.
 */
const matrix = process.env.PLAYWRIGHT_BROWSERS === '1';

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
  projects: matrix
    ? [
        {
          name: 'chromium',
          use: { ...devices['Desktop Chrome'] },
        },
        {
          name: 'firefox',
          use: { ...devices['Desktop Firefox'] },
        },
        {
          name: 'webkit',
          use: { ...devices['Desktop Safari'] },
        },
      ]
    : [
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
