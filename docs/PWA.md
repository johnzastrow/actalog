# PWA Local HTTPS & Testing (ActaLog)

This page describes how to test ActaLog as a Progressive Web App (PWA) locally, including creating local HTTPS certificates (recommended), mapping hostnames, running the Vite dev server with HTTPS, and running the repository's PWA health-check script.

Files referenced from `scripts/`:

- `scripts/pwa-health-check.sh` — automated PWA validator (requires `curl`, `jq`, `openssl`).
- `scripts/PWA_HEALTH_CHECK_README.md` — detailed README for the health-check script.
- `scripts/PWA_QUICK_REFERENCE.md` / `scripts/PWA_TESTING_EXAMPLES.md` — quick reference and examples.

## Goals

- Enable local HTTPS for named-host testing (service worker + installability require HTTPS in production).
- Provide compact, repeatable commands for Windows/macOS/Linux developers.
- Show how to run the repo's PWA health-check against dev/preview/prod.

## Prerequisites

- Node.js 18+ and `npm` (frontend)
- `mkcert` (recommended) or an HTTPS-capable proxy (Caddy / nginx)
- `curl`, `jq`, `openssl` (for `pwa-health-check.sh`)

### Install mkcert

mkcert is simple and platform-supported. Follow the instructions at https://github.com/FiloSottile/mkcert.

Windows (recommended using PowerShell as Admin):

```powershell
# Install Chocolatey if not present, then mkcert
choco install mkcert -y
mkcert -install
```

macOS (Homebrew):

```bash
brew install mkcert
mkcert -install
```

Linux (varies by distro):

```bash
# Example (Debian/Ubuntu):
sudo apt install libnss3-tools
# install mkcert binary per instructions and then:
mkcert -install
```

## Create cert for a local hostname

Pick a test hostname you want to use for the PWA (example: `pwa.local.example` or `subdomain.example.com`). You will need to map this hostname to `127.0.0.1` in your hosts file.

Generate a certificate and key using mkcert:

```bash
# from project root (or any folder)
mkcert -cert-file cert.pem -key-file key.pem "subdomain.example.com" "localhost" "127.0.0.1"
```

This creates `cert.pem` and `key.pem` in the current directory. Keep them private.

### Hosts file

Map the hostname to localhost so your browser resolves it to your machine.

Windows hosts file (run editor as Administrator):

```
# Add to C:\Windows\System32\drivers\etc\hosts
127.0.0.1 subdomain.example.com
```

Linux / macOS `/etc/hosts` (requires sudo):

```
sudo -- sh -c 'echo "127.0.0.1 subdomain.example.com" >> /etc/hosts'
```

Note: Use a name under your control or a dev-specific subdomain to avoid conflicting with real domains.

## Run Vite dev server with HTTPS (local named-host)

Vite supports `--host` and `--https` options. Use the certificates you generated with mkcert.

From `web/`:

```bash
# Example: using the generated cert/key and a named host
cd web
npm ci
npm run dev -- --host subdomain.example.com --https --cert ../cert.pem --key ../key.pem
```

If you prefer to listen on all interfaces (quick test), run:

```bash
npm run dev -- --host 0.0.0.0
```

This avoids editing the hosts file, but it doesn't give you the named-host URL needed for some PWA scope/start_url tests.

## Production-preview test (build + preview)

1. Build the production assets

```bash
cd web
npm ci
npm run build
```

2. Start the preview server (Vite's `preview`) — it listens on `localhost:4173` by default. You can also run it with HTTPS using a small proxy or Caddy.

```bash
npm run preview
```

## Using the pwa-health-check script

The repo includes `scripts/pwa-health-check.sh` for automated validation. Examples:

```bash
# Check a dev server running on localhost
./scripts/pwa-health-check.sh http://localhost:3000

# Check a local HTTPS named host (after mkcert + hosts file)
./scripts/pwa-health-check.sh https://subdomain.example.com

# Check the preview server
./scripts/pwa-health-check.sh http://localhost:4173
```

See `scripts/PWA_HEALTH_CHECK_README.md` for a full explanation of the checks and `scripts/PWA_QUICK_REFERENCE.md` for quick commands.

## Mobile testing

- Android (Chrome): open the named-host HTTPS URL in Chrome and use the Install option from the menu. Test offline (airplane mode) after install to verify SW caching.
- iOS (Safari): use Share → Add to Home Screen. Note: iOS has different PWA limitations (no background sync, limited SW features). Test offline behaviour.

## Alternatives to mkcert

- Caddy: automatically provisions HTTPS via Let's Encrypt for publicly reachable domains; can be used as a reverse-proxy to your local backend during end-to-end testing.
- Self-signed certs: possible but harder to trust in browsers without importing CA.

Example Caddyfile (reverse-proxy to local port 3000):

```
subdomain.example.com {
  reverse_proxy localhost:3000
}
```

Run Caddy from your machine; it will handle HTTPS for that domain if publicly resolvable (or you can use a local domain and trust the cert via mkcert).

## Troubleshooting

- ERR_NAME_NOT_RESOLVED / ENOTFOUND: check hosts file entry and flush DNS.
- Browser still warns about cert: ensure you created and installed mkcert root CA (`mkcert -install`).
- Service worker not registering: ensure the app is served over HTTPS (or `localhost`) and the SW script is reachable at the registered path (typically `/sw.js` or as generated by VitePWA plugin).
- PWA health-check missing manifest or icons: check that `dist/manifest.webmanifest` exists after build and that `index.html` links to it.

## Quick checklist

- [ ] Create mkcert certs and `mkcert -install`
- [ ] Add host entry for `subdomain.example.com` → `127.0.0.1`
- [ ] Run `npm run dev -- --host subdomain.example.com --https --cert ../cert.pem --key ../key.pem`
- [ ] Visit `https://subdomain.example.com:3000` (or default port) and open DevTools → Application to inspect manifest and service worker
- [ ] Run `./scripts/pwa-health-check.sh https://subdomain.example.com` and fix any failures/warnings

## References

- `scripts/pwa-health-check.sh`
- `scripts/PWA_HEALTH_CHECK_README.md`
- `scripts/PWA_QUICK_REFERENCE.md`
- `scripts/PWA_TESTING_EXAMPLES.md`
- Vite dev server docs: https://vitejs.dev/guide/cli.html
- mkcert: https://github.com/FiloSottile/mkcert

---

If you'd like, I can also:

- Add an npm `dev:https` script in `web/package.json` that bundles the `--host`/`--https` flags using a configurable path to the certs, or
- Add a small PowerShell helper `scripts/mkcert-windows.ps1` (if you want platform-automated cert generation + hosts-file editing prompts).

Which of those would you prefer next? 
