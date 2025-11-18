# PWA Health Check - Quick Reference

## Installation

```bash
# Install dependencies
sudo apt-get install curl jq openssl

# Optional: Install Lighthouse for detailed scoring
npm install -g lighthouse

# Make script executable
chmod +x pwa-health-check.sh
```

## Usage

```bash
# Interactive mode (prompts for URL)
./pwa-health-check.sh

# Command-line mode
./pwa-health-check.sh https://example.com

# Local development
./pwa-health-check.sh http://localhost:3000
```

## Quick Checklist

| Requirement | Priority | Fixed By |
|-------------|----------|----------|
| HTTPS | ✓ Critical | SSL certificate, Caddy, or Let's Encrypt |
| manifest.json | ✓ Critical | `<link rel="manifest" href="/manifest.json">` |
| App name | ✓ Critical | `"name": "Your App"` in manifest |
| 192x192 icon | ✓ Critical | Add icon to manifest |
| 512x512 icon | ⚠ Important | Add icon to manifest |
| Service worker | ✓ Critical | Register SW in JavaScript |
| Fetch handler | ✓ Critical | Add fetch event in SW |
| Viewport tag | ✓ Critical | `<meta name="viewport" content="...">` |
| Display mode | ⚠ Important | `"display": "standalone"` in manifest |
| Theme color | ⚠ Recommended | `"theme_color": "#00bcd4"` in manifest |
| Start URL | ⚠ Recommended | `"start_url": "/"` in manifest |

## Status Levels

| Status | Criteria | Meaning |
|--------|----------|---------|
| ✓ EXCELLENT | 0 failures, 90%+ pass | Production ready |
| ✓ GOOD | 0 failures, <90% pass | Installable, has warnings |
| ⚠ NEEDS IMPROVEMENT | 1-2 failures | Some features missing |
| ✗ NOT A PWA | 3+ failures | Not installable |

## Common Fixes

### Missing Manifest

```html
<!-- index.html -->
<link rel="manifest" href="/manifest.json">
```

```json
// manifest.json
{
  "name": "Your App Name",
  "short_name": "App",
  "start_url": "/",
  "display": "standalone",
  "theme_color": "#00bcd4",
  "background_color": "#ffffff",
  "icons": [
    {
      "src": "/icons/icon-192x192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/icons/icon-512x512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
```

### Missing Service Worker

```javascript
// main.js or index.html
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/sw.js')
    .then(reg => console.log('SW registered'))
    .catch(err => console.error('SW registration failed', err));
}
```

```javascript
// sw.js
self.addEventListener('install', event => {
  event.waitUntil(
    caches.open('v1').then(cache => {
      return cache.addAll(['/']);
    })
  );
});

self.addEventListener('fetch', event => {
  event.respondWith(
    caches.match(event.request).then(response => {
      return response || fetch(event.request);
    })
  );
});
```

### Missing Viewport

```html
<meta name="viewport" content="width=device-width, initial-scale=1">
```

### Missing HTTPS (Production)

```bash
# Option 1: Let's Encrypt
sudo certbot --nginx -d yourdomain.com

# Option 2: Caddy (automatic HTTPS)
caddy reverse-proxy --from yourdomain.com --to localhost:8080
```

## Test Output Key

| Symbol | Meaning | Action |
|--------|---------|--------|
| ✓ | Pass | Great! Keep it |
| ✗ | Fail | Must fix for PWA |
| ⚠ | Warning | Recommended to fix |
| ℹ | Info | Optional enhancement |

## ActaLog Quick Test

```bash
# Development
cd web && npm run dev
# In new terminal:
./pwa-health-check.sh http://localhost:3000

# Production build test
cd web && npm run build && npm run preview
# In new terminal:
./pwa-health-check.sh http://localhost:4173

# Live production
./pwa-health-check.sh https://actalog.yourdomain.com
```

## CI/CD Integration

```yaml
# .github/workflows/pwa-check.yml
name: PWA Health Check
on: [push, pull_request]
jobs:
  pwa-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install dependencies
        run: sudo apt-get install -y curl jq openssl
      - name: Build
        run: cd web && npm install && npm run build
      - name: Preview
        run: cd web && npm run preview &
      - name: Wait for server
        run: sleep 5
      - name: PWA Check
        run: ./pwa-health-check.sh http://localhost:4173
```

## Mobile Testing

### Android (Chrome)

1. Visit your URL in Chrome
2. Menu (⋮) → Install app
3. Test offline: Airplane mode → Open app

### iOS (Safari)

1. Visit your URL in Safari
2. Share button → Add to Home Screen
3. Test offline: Airplane mode → Open app

## Performance Benchmarks

| Metric | Excellent | Good | Poor |
|--------|-----------|------|------|
| Load time | <1s | 1-3s | >3s |
| Lighthouse PWA | 90-100 | 70-89 | <70 |
| Icon sizes | 192+512 | 192 only | None |

## Dependencies

**Required:**
- curl (HTTP requests)
- jq (JSON parsing)
- openssl (SSL checks)

**Optional:**
- lighthouse (detailed scoring)

## Resources

- [MDN PWA Guide](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps)
- [web.dev PWA](https://web.dev/progressive-web-apps/)
- [PWA Builder](https://www.pwabuilder.com/)
- [Lighthouse](https://github.com/GoogleChrome/lighthouse)

## Getting Help

```bash
# Show full documentation
cat PWA_HEALTH_CHECK_README.md

# Show examples
cat PWA_TESTING_EXAMPLES.md

# Debug with verbose curl
curl -v https://yourdomain.com

# Validate JSON manually
cat manifest.json | jq .
```
