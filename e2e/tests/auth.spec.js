import { test, expect, expectNoConsoleErrors } from '../fixtures.js';

const EMAIL = process.env.E2E_EMAIL || 'e2e-admin@actalog.local';
const PASS = process.env.E2E_PASSWORD || 'TestPass123!';

test.describe('unauthenticated pages', () => {
  test('login page renders with outlined (styled) inputs', async ({ page, consoleErrors }) => {
    await page.goto('/login', { waitUntil: 'networkidle' });

    await expect(page.getByText('ActaLog')).toBeVisible();
    await expect(page.getByText(/sign in to your account/i)).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Password', { exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: /sign in/i })).toBeVisible();

    // Regression guard for the Vuetify 3->4 breakage. The inputs use the "solo"
    // variant, which renders a filled, elevated box (solid background + shadow).
    // Under the broken v4 build these fields collapsed to flat, transparent,
    // full-width rows. Assert the visual properties rather than a class name so
    // the guard is resilient across Vuetify versions.
    await expect(page.getByLabel('Email')).toBeVisible();
    const style = await page.getByLabel('Email').evaluate((input) => {
      const field = input.closest('.v-field');
      if (!field) return null;
      const cs = getComputedStyle(field);
      return { cls: field.className, bg: cs.backgroundColor, shadow: cs.boxShadow, height: field.getBoundingClientRect().height };
    });
    expect(style, 'email input must be wrapped in a .v-field').not.toBeNull();
    expect(style.height, 'input field height').toBeGreaterThan(20);
    // The solo variant renders an elevation shadow; the broken v4 render lost it.
    expect(style.shadow, 'input field must have an elevation shadow').not.toBe('none');
    expect(style.shadow.length, 'input field must have an elevation shadow').toBeGreaterThan(10);

    await page.screenshot({ path: 'screenshots/login.png' });
    expectNoConsoleErrors(consoleErrors);
  });

  test('login succeeds and lands on the dashboard', async ({ page }) => {
    await page.goto('/login', { waitUntil: 'networkidle' });
    await page.getByLabel('Email').fill(EMAIL);
    await page.getByLabel('Password', { exact: true }).fill(PASS);
    await page.getByRole('button', { name: /sign in/i }).click();
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 15_000 });
  });

  test('register page renders all fields', async ({ page, consoleErrors }) => {
    await page.goto('/register', { waitUntil: 'networkidle' });
    await expect(page.getByLabel('Name')).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Password', { exact: true })).toBeVisible();
    await expect(page.getByLabel('Confirm Password', { exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: /sign up|register|create account/i })).toBeVisible();
    await page.screenshot({ path: 'screenshots/register.png' });
    expectNoConsoleErrors(consoleErrors);
  });
});
