import { test as base, expect } from '@playwright/test';

// Console errors that are known-benign and must not fail a test:
//  - the redundant jsdelivr @mdi/font <link> blocked by CSP (icons are bundled
//    locally, so this is cosmetic — tracked for cleanup, not a real failure).
const BENIGN_CONSOLE = [
  /cdn\.jsdelivr\.net/i,
  /Content Security Policy/i,
  /favicon/i,
];

// `test` extends the base with an automatic console-error collector. Any page
// error or console.error that is not in the benign list is attached to the test
// and can be asserted with `expectNoConsoleErrors`.
export const test = base.extend({
  consoleErrors: [
    async ({ page }, use) => {
      const errors = [];
      page.on('console', (msg) => {
        if (msg.type() !== 'error') return;
        const text = msg.text();
        if (BENIGN_CONSOLE.some((re) => re.test(text))) return;
        errors.push(text);
      });
      page.on('pageerror', (err) => errors.push(`pageerror: ${err.message}`));
      await use(errors);
    },
    { auto: true },
  ],
});

export { expect };

// Assert no unexpected console/page errors accumulated so far.
export function expectNoConsoleErrors(errors) {
  expect(errors, `unexpected console errors:\n${errors.join('\n')}`).toHaveLength(0);
}
