import { defineConfig, devices } from '@playwright/test';

// Target is controlled by BASE_URL so the same suite runs against a local
// container, the beta host, or production. Defaults to the local test container.
const BASE_URL = process.env.BASE_URL || 'http://localhost:8095';

export default defineConfig({
  testDir: './tests',
  // Auth state is created once (register-or-login) and reused by authed specs.
  globalSetup: './global-setup.js',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: [['list'], ['html', { open: 'never', outputFolder: 'playwright-report' }]],
  timeout: 45_000,
  expect: { timeout: 10_000 },
  outputDir: 'test-results',
  use: {
    baseURL: BASE_URL,
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
    video: 'off',
  },
  projects: [
    // Anonymous (logged-out) flows: login/register rendering + the Vuetify regression guard.
    {
      name: 'anon-desktop',
      testMatch: /auth\.spec\.js/,
      use: { ...devices['Desktop Chrome'], viewport: { width: 1280, height: 900 } },
    },
    // Authenticated navigation + visual screenshots, desktop and mobile (mobile-first PWA).
    {
      name: 'authed-desktop',
      testMatch: /navigation\.spec\.js/,
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1280, height: 900 },
        storageState: '.auth/admin.json',
      },
    },
    {
      name: 'authed-mobile',
      testMatch: /navigation\.spec\.js/,
      use: { ...devices['Pixel 5'], storageState: '.auth/admin.json' },
    },
  ],
});
