# Vite Configuration Integration with start-frontend.sh

## Overview

The `start-frontend.sh` script now automatically configures `vite.config.js` via environment variables, ensuring that domain names, ports, and HTTPS settings are correctly applied to:
- Vite dev/preview server configuration
- PWA manifest (scope and start_url)
- Build configuration (base path)

## How It Works

### 1. Environment Variables Set by start-frontend.sh

Before running npm commands, the script exports these environment variables:

```bash
export VITE_DEV_HOST="$HOSTNAME"      # Domain or localhost
export VITE_DEV_PORT="$PORT"          # Port number (3000 or alternative)
export VITE_USE_HTTPS="true|false"    # Whether HTTPS is enabled
export VITE_DEPLOYMENT_URL="..."      # Full deployment URL
```

### 2. vite.config.js Reads Environment Variables

The Vite configuration reads these variables on startup:

```javascript
const DEFAULT_HOST = process.env.VITE_DEV_HOST || 'localhost'
const DEFAULT_PORT = process.env.VITE_DEV_PORT || '3000'
const USE_HTTPS = process.env.VITE_USE_HTTPS === 'true'
const DEPLOYMENT_URL = process.env.VITE_DEPLOYMENT_URL || `http://localhost:${DEFAULT_PORT}`
```

### 3. Configuration Applied

**Server Configuration:**
```javascript
server: {
  host: DEFAULT_HOST,
  port: parseInt(DEFAULT_PORT),
  https: httpsOptions ? httpsOptions : false,
}
```

**PWA Manifest:**
```javascript
manifest: {
  scope: `${FULL_DEPLOYMENT_URL}/`,
  start_url: `${FULL_DEPLOYMENT_URL}/`,
  // ... other manifest properties
}
```

**Build Configuration:**
```javascript
build: {
  base: `${FULL_DEPLOYMENT_URL}/`,
}
```

## Deployment URL Logic

The script constructs `VITE_DEPLOYMENT_URL` based on your configuration:

### Localhost Mode

**Without HTTPS:**
```
http://localhost:3000
```

**With HTTPS:**
```
https://localhost:3000
```

### Domain Mode with Reverse Proxy

When using a reverse proxy (like Caddy), the public URL doesn't include the port:
```
https://al.fluidgrid.site
```

The reverse proxy listens on port 80/443 and forwards requests to `localhost:3000` (or your configured port).

### Domain Mode without Reverse Proxy

When accessing the dev server directly, the URL includes the port:
```
https://al.fluidgrid.site:3000
```

## Why This Matters

### PWA Manifest Scope and Start URL

Progressive Web Apps require accurate `scope` and `start_url` values in the manifest:
- These must match the domain where the PWA is deployed
- Incorrect values prevent PWA installation or cause navigation issues
- These values are baked into the build at compile time

**Before (hardcoded placeholders):**
```javascript
scope: 'https://subdomain.example.com/',
start_url: 'https://subdomain.example.com/',
```

**After (dynamic based on deployment):**
```javascript
scope: 'https://al.fluidgrid.site/',
start_url: 'https://al.fluidgrid.site/',
```

### Build Base Path

The `build.base` setting determines the base path for all static assets:
- CSS files: `<link rel="stylesheet" href="${base}assets/index.css">`
- JS files: `<script src="${base}assets/index.js">`
- Images: `<img src="${base}logo.png">`

**Before (relative):**
```javascript
base: '/'
```

**After (absolute deployment URL):**
```javascript
base: 'https://al.fluidgrid.site/'
```

## Usage Examples

### Example 1: Localhost Development

```bash
$ ./scripts/start-frontend.sh
Run in (d)ev or (p)review mode? [d/p]: d
Will you access on localhost or domain? [l/D]: l
Expose to LAN? [y/N]: n
Use HTTPS locally? [y/N]: n

# Environment variables set:
# VITE_DEV_HOST=localhost
# VITE_DEV_PORT=3000
# VITE_USE_HTTPS=false
# VITE_DEPLOYMENT_URL=http://localhost:3000
```

**Result:**
- Dev server starts on `http://localhost:3000`
- PWA manifest scope: `http://localhost:3000/`
- Build base: `http://localhost:3000/`

### Example 2: Domain with Reverse Proxy

```bash
$ ./scripts/start-frontend.sh
Run in (d)ev or (p)review mode? [d/p]: d
Will you access on localhost or domain? [l/D]: D
Enter the domain: al.fluidgrid.site
Will you use a reverse proxy like Caddy? [Y/n]: Y

# Environment variables set:
# VITE_DEV_HOST=al.fluidgrid.site
# VITE_DEV_PORT=3000
# VITE_USE_HTTPS=false
# VITE_DEPLOYMENT_URL=https://al.fluidgrid.site
```

**Result:**
- Dev server starts on `localhost:3000`
- Caddy proxies `https://al.fluidgrid.site` → `localhost:3000`
- PWA manifest scope: `https://al.fluidgrid.site/`
- Build base: `https://al.fluidgrid.site/`

### Example 3: Domain Direct Access with Port 3001

```bash
$ ./scripts/start-frontend.sh
Run in (d)ev or (p)review mode? [d/p]: d
Will you access on localhost or domain? [l/D]: D
Enter the domain: al.fluidgrid.site
Will you use a reverse proxy like Caddy? [Y/n]: n
Will you use HTTPS directly on the frontend? [Y/n]: Y

# Port 3000 is in use, chose option 2 (use different port)
# Found available port: 3001

# Environment variables set:
# VITE_DEV_HOST=al.fluidgrid.site
# VITE_DEV_PORT=3001
# VITE_USE_HTTPS=true
# VITE_DEPLOYMENT_URL=https://al.fluidgrid.site:3001
```

**Result:**
- Dev server starts on `https://al.fluidgrid.site:3001`
- PWA manifest scope: `https://al.fluidgrid.site:3001/`
- Build base: `https://al.fluidgrid.site:3001/`

## Troubleshooting

### Issue: PWA Not Installing

**Symptoms:**
- Browser doesn't show "Install App" prompt
- PWA manifest errors in DevTools console

**Common Causes:**
1. Manifest `scope` doesn't match the current origin
2. Manifest `start_url` points to different domain

**Solution:**
Re-run `start-frontend.sh` with correct domain settings. The script will automatically update the manifest.

### Issue: Assets Not Loading (404 errors)

**Symptoms:**
- CSS not loading
- JavaScript files not found
- Images show broken

**Common Causes:**
- `build.base` doesn't match deployment URL
- Built with one domain, deployed to another

**Solution:**
For production builds, ensure you run:
```bash
./scripts/start-frontend.sh
# Choose (p)review mode
# Provide correct production domain
```

This builds with the correct base path.

### Issue: HMR (Hot Module Reload) Not Working

**Symptoms:**
- Changes to code don't auto-refresh
- WebSocket connection errors in console

**Common Causes:**
- `VITE_DEV_HOST` doesn't match how you're accessing the app
- Firewall blocking WebSocket connections

**Solution:**
If accessing via domain but dev server is on localhost, ensure:
1. Reverse proxy configured correctly
2. WebSocket proxying enabled in Caddy/nginx
3. `host` setting matches your access method

### Issue: Different Port Than Expected

**Symptoms:**
- Script says port 3001 but expected 3000
- "Port already in use" message

**Solution:**
The script detected port 3000 was in use and automatically selected an available port. Either:
1. Stop the process using port 3000
2. Update reverse proxy config to use the new port

## Verifying Configuration

### Check Environment Variables

Run the script with debug output:
```bash
# Add after line 329 in start-frontend.sh (temporarily):
echo "Debug: VITE_DEV_HOST=$VITE_DEV_HOST"
echo "Debug: VITE_DEV_PORT=$VITE_DEV_PORT"
echo "Debug: VITE_USE_HTTPS=$VITE_USE_HTTPS"
echo "Debug: VITE_DEPLOYMENT_URL=$VITE_DEPLOYMENT_URL"
```

### Check Built Manifest

After building, inspect the manifest:
```bash
cd web
npm run build
cat dist/manifest.webmanifest | jq '.scope, .start_url'
```

Should show your deployment URL, not placeholders.

### Check Vite Dev Server Output

When starting dev server, Vite shows the configuration:
```
VITE v5.x.x  ready in xxx ms

➜  Local:   http://localhost:3000/
➜  Network: http://192.168.1.100:3000/
➜  press h + enter to show help
```

Verify the displayed URLs match your expectations.

## Files Modified

- ✅ `scripts/start-frontend.sh` (lines 331-364) - Added environment variable exports
- ✅ `web/vite.config.js` (lines 15-24, 55-56, 209-210, 230) - Added environment variable reading

## Related Documentation

- `PORT_CONFLICT_HANDLING.md` - How port conflict detection works
- `CADDY_CONFIG_FIX.md` - Caddy reverse proxy configuration
- `CADDY_DEBUGGING.md` - Debugging Caddy issues
