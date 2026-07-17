# ActaLog end-to-end tests (Playwright)

Reusable browser tests that drive a running ActaLog deployment: they verify the
login/register pages, sign in, walk the main authenticated screens, and capture a
screenshot of every screen (desktop + mobile) for visual review.

They are **self-contained** — this folder has its own `package.json`, so it does
not touch `web/`'s dependencies or lockfile.

## Why these exist

A dependency sweep once upgraded Vuetify across a major version (3 -> 4) without
migrating the app, which silently broke component styling (outlined inputs,
cards, spacing) app-wide. The build still passed CI. `auth.spec.js` now has an
explicit guard (the login field must render a real outlined border), and
`navigation.spec.js` screenshots every screen so this class of regression is
visible before shipping.

## Setup

```bash
cd e2e
npm install
npm run install-browsers   # one-time: installs the Chromium Playwright build
```

## Run

Target is chosen with `BASE_URL` (defaults to `http://localhost:8095`, a local
container).

```bash
# Against a local container (see below)
npm test

# Against beta
BASE_URL=https://albeta.fluidgrid.site npm test

# Against production — never auto-register; supply a real, disposable account
BASE_URL=https://al.fluidgrid.site \
  E2E_REGISTER=false E2E_EMAIL='you@example.com' E2E_PASSWORD='...' \
  npm test

npm run report   # open the HTML report
```

Screenshots land in `screenshots/` (git-ignored): `login.png`, `register.png`,
and `<project>/<screen>.png` for each authenticated screen.

### Spin up a local container to test a specific image

```bash
docker run -d --name actalog-local-test -p 8095:8080 \
  -e APP_ENV=development -e JWT_SECRET=local-test-secret \
  ghcr.io/johnzastrow/actalog:v1.3.0
```

## Environment variables

| Var | Default | Purpose |
|-----|---------|---------|
| `BASE_URL` | `http://localhost:8095` | Deployment under test |
| `E2E_EMAIL` | `e2e-admin@actalog.local` | Test account email |
| `E2E_PASSWORD` | `TestPass123!` | Test account password |
| `E2E_REGISTER` | `true` | If `false`, never register — log in only (use for prod) |

## Layout

- `playwright.config.js` — projects: `anon-desktop`, `authed-desktop`, `authed-mobile`
- `global-setup.js` — registers-or-logs-in once, saves `.auth/admin.json`
- `fixtures.js` — auto console-error collector (benign CSP/font noise filtered)
- `tests/auth.spec.js` — login/register rendering + outlined-input regression guard
- `tests/navigation.spec.js` — authenticated screens + full-page screenshots
