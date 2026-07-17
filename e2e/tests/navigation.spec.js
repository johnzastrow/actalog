import { test, expect, expectNoConsoleErrors } from '../fixtures.js';

// Main authenticated routes. Screenshotting each catches broad visual/layout
// regressions (e.g. a UI-framework upgrade dropping component styles) across the
// app, not just the login page.
const ROUTES = [
  ['dashboard', '/dashboard'],
  ['workouts', '/workouts'],
  ['calendar', '/workouts/calendar'],
  ['movements', '/movements'],
  ['wods', '/wods'],
  ['performance', '/performance'],
  ['prs', '/prs'],
  ['leaderboard', '/leaderboard'],
  ['notifications', '/notifications'],
  ['profile', '/profile'],
  ['settings', '/settings'],
  ['admin', '/admin'],
];

test.describe('authenticated navigation', () => {
  for (const [name, path] of ROUTES) {
    test(`renders ${name}`, async ({ page, consoleErrors }, testInfo) => {
      const resp = await page.goto(path, { waitUntil: 'networkidle' });

      // Auth must hold — a bounce back to /login means the session broke.
      expect(page.url(), `expected to stay on ${path}, not redirect to login`).not.toContain('/login');
      // The Vuetify app root should be present and the page should not be a hard error.
      await expect(page.locator('.v-application')).toBeVisible();
      if (resp) expect(resp.status(), `HTTP status for ${path}`).toBeLessThan(400);

      const shot = `screenshots/${testInfo.project.name}/${name}.png`;
      await page.screenshot({ path: shot, fullPage: true });

      expectNoConsoleErrors(consoleErrors);
    });
  }
});
