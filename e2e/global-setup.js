import { chromium } from '@playwright/test';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { mkdirSync, existsSync } from 'node:fs';

// Creates (or reuses) an authenticated session and saves it to .auth/admin.json
// for the authenticated projects. Idempotent: logs in if the account already
// exists, otherwise registers it first (registration auto-authenticates and the
// first user becomes admin).
//
// Env:
//   BASE_URL       target under test (default http://localhost:8095)
//   E2E_EMAIL      test account email    (default e2e-admin@actalog.local)
//   E2E_PASSWORD   test account password (default TestPass123!)
//   E2E_REGISTER   set to "false" to never register (use for prod: supply real creds)
export default async function globalSetup() {
  // Skip all auth setup — for running the anonymous render specs against a real
  // deployment (beta/prod) without creating accounts or spending rate-limit budget.
  if (process.env.E2E_SKIP_AUTH === '1') return;

  const BASE = process.env.BASE_URL || 'http://localhost:8095';
  const EMAIL = process.env.E2E_EMAIL || 'e2e-admin@actalog.local';
  const PASS = process.env.E2E_PASSWORD || 'TestPass123!';
  const allowRegister = process.env.E2E_REGISTER !== 'false';

  const dir = dirname(fileURLToPath(import.meta.url));
  const authDir = join(dir, '.auth');
  const statePath = join(authDir, 'admin.json');
  mkdirSync(authDir, { recursive: true });

  const browser = await chromium.launch();

  // The auth endpoints are rate-limited (5 requests / 15 min / IP). Reuse a
  // still-valid saved session so re-runs cost zero auth calls. Set E2E_FRESH_AUTH=1
  // to force a new login.
  if (existsSync(statePath) && process.env.E2E_FRESH_AUTH !== '1') {
    const ctx = await browser.newContext({ storageState: statePath });
    const page = await ctx.newPage();
    await page.goto(`${BASE}/dashboard`, { waitUntil: 'networkidle' }).catch(() => {});
    if (page.url().includes('/dashboard')) {
      await browser.close();
      return; // existing session still valid
    }
    await ctx.close();
  }

  const ctx = await browser.newContext();
  const page = await ctx.newPage();

  async function login() {
    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.getByLabel('Email').fill(EMAIL);
    await page.getByLabel('Password', { exact: true }).fill(PASS);
    await page.getByRole('button', { name: /sign in/i }).click();
    try {
      await page.waitForURL('**/dashboard', { timeout: 10_000 });
      return true;
    } catch {
      return false;
    }
  }

  let ok = await login();
  if (!ok && allowRegister) {
    await page.goto(`${BASE}/register`, { waitUntil: 'networkidle' });
    await page.getByLabel('Name').fill('E2E Admin');
    await page.getByLabel('Email').fill(EMAIL);
    await page.getByLabel('Password', { exact: true }).fill(PASS);
    await page.getByLabel('Confirm Password', { exact: true }).fill(PASS);
    await page.getByRole('button', { name: /sign up|register|create account/i }).click();
    // Register auto-authenticates AND shows an email hint; reset client state so
    // we get a clean, deterministic logged-in session via the login form.
    await page.waitForTimeout(1500);
    await ctx.clearCookies();
    await page.evaluate(() => { localStorage.clear(); sessionStorage.clear(); }).catch(() => {});
    ok = await login();
  }

  if (!ok) {
    await browser.close();
    throw new Error(
      `Could not authenticate against ${BASE} as ${EMAIL}. ` +
        `Set E2E_EMAIL/E2E_PASSWORD to a valid account (and E2E_REGISTER=false for prod).`
    );
  }

  await ctx.storageState({ path: join(authDir, 'admin.json') });
  await browser.close();
}
