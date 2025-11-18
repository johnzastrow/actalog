# PWA Health Check Script

## Overview

The `pwa-health-check.sh` script is a comprehensive tool for validating whether a website meets all the requirements to function as a Progressive Web App (PWA). It performs automated checks across multiple categories and provides a detailed report with actionable recommendations.

## Features

### Comprehensive PWA Validation

The script checks **8 major categories**:

1. **HTTPS Connection** - Verifies secure connection and SSL certificate validity
2. **Web App Manifest** - Validates manifest.json structure and required properties
3. **Manifest Properties** - Checks name, icons, display mode, theme colors, etc.
4. **Service Worker** - Detects service worker registration and offline capabilities
5. **HTML Meta Tags** - Validates viewport, theme-color, and mobile-specific tags
6. **Lighthouse Audit** - Optional automated PWA scoring (if Lighthouse is installed)
7. **Performance** - Basic performance checks (load time, HTTP/2, compression)
8. **Mobile Optimization** - Apple-specific tags and responsive design

### Visual Reporting

- **Color-coded output** - Green (pass), Red (fail), Yellow (warning), Blue (info)
- **Unicode symbols** - ✓ (pass), ✗ (fail), ⚠ (warning), ℹ (info)
- **Progress tracking** - Real-time test execution feedback
- **Summary report** - Overall PWA status and recommendations

### Smart Detection

- Automatically normalizes URLs (adds https:// if missing)
- Detects manifest.json from HTML link tags
- Follows relative and absolute paths for resources
- Checks multiple common service worker filenames
- Parses JSON manifests for required fields
- Validates icon sizes and formats

## Installation

### Prerequisites

**Required dependencies:**
```bash
sudo apt-get install curl jq openssl
```

**Optional (for detailed Lighthouse scoring):**
```bash
npm install -g lighthouse
```

### Setup

```bash
# Make script executable
chmod +x pwa-health-check.sh
```

## Usage

### Interactive Mode

Run without arguments to be prompted for URL:

```bash
./pwa-health-check.sh
```

You'll be asked to enter the URL:
```
Enter the URL to check:
> https://example.com
```

### Command-Line Mode

Pass URL as argument:

```bash
./pwa-health-check.sh https://example.com
```

The script automatically:
- Adds `https://` if no protocol is specified
- Removes trailing slashes
- Normalizes the URL format

### Examples

```bash
# Check a production PWA
./pwa-health-check.sh https://app.example.com

# Check localhost (will warn about HTTPS)
./pwa-health-check.sh http://localhost:3000

# Check without protocol (automatically adds https://)
./pwa-health-check.sh app.example.com

# Check subdomain
./pwa-health-check.sh https://pwa.mysite.org
```

## What Gets Checked

### 1. HTTPS Connection ✓

**Critical for PWA** - Service workers require HTTPS (except localhost)

- ✓ Verifies site is served over HTTPS
- ✓ Validates SSL certificate
- ✓ Checks certificate expiration
- ✗ HTTP-only sites fail (except localhost)

**Example output:**
```
━━━ 1. HTTPS Connection ━━━

✓ Site is served over HTTPS
✓ SSL certificate is valid
```

### 2. Web App Manifest ✓

**Critical for PWA** - Required for installation

- ✓ Detects manifest link in HTML (`<link rel="manifest">`)
- ✓ Fetches manifest.json file
- ✓ Validates JSON syntax
- ✓ Constructs full URL for relative paths

**Example output:**
```
━━━ 3. Web App Manifest ━━━

✓ Manifest link found in HTML: /manifest.json
   Manifest URL: https://example.com/manifest.json
✓ Manifest fetched and is valid JSON
```

### 3. Manifest Properties ✓

**Critical fields** - Required for installability

Checks for:
- ✓ **name** or **short_name** (required)
- ✓ **icons** array with 192x192 and 512x512 sizes (required)
- ✓ **start_url** (recommended)
- ✓ **display** mode (standalone, fullscreen, minimal-ui recommended)
- ⚠ **theme_color** (recommended)
- ⚠ **background_color** (recommended for splash screen)
- ⚠ **description** (recommended)
- ℹ **scope** (optional)
- ℹ **orientation** (optional)

**Example output:**
```
━━━ 4. Manifest Properties ━━━

✓ Name: "ActaLog - CrossFit Tracker"
✓ Icons defined: 3 icon(s)
✓ Has 192x192 icon (required for installability)
✓ Has 512x512 icon (required for splash screen)
   Icon details:
   - 192x192 image/png /icons/icon-192x192.png
   - 512x512 image/png /icons/icon-512x512.png
   - 180x180 image/png /icons/apple-touch-icon.png
✓ Start URL: "/"
✓ Display mode: "standalone"
   Good! Display mode is PWA-friendly
✓ Theme color: #00bcd4
✓ Background color: #ffffff
✓ Description: "Track your CrossFit workouts and PRs"
```

### 4. Service Worker ✓

**Critical for PWA** - Required for offline functionality

- ✓ Detects service worker registration code
- ✓ Checks common service worker filenames (sw.js, service-worker.js, etc.)
- ✓ Validates fetch event handler (offline support)
- ✓ Checks install event handler
- ✓ Checks activate event handler

**Example output:**
```
━━━ 5. Service Worker ━━━

✓ Service worker registration code found
✓ Service worker file found: sw.js (HTTP 200)
✓ Service worker has fetch event handler (offline support)
✓ Service worker has install event handler
✓ Service worker has activate event handler
```

### 5. HTML Meta Tags ✓

**Important for mobile experience**

- ✓ **viewport** meta tag (required for responsive design)
- ⚠ **theme-color** meta tag (recommended)
- ⚠ **description** meta tag (recommended for SEO)
- ℹ **apple-mobile-web-app-capable** (optional for iOS)
- ℹ **apple-touch-icon** (optional for iOS home screen)

**Example output:**
```
━━━ 6. HTML Meta Tags ━━━

✓ Viewport meta tag present
   Content: width=device-width, initial-scale=1
✓ Theme color meta tag: #00bcd4
✓ Description meta tag present
   Track your CrossFit workouts, log PRs, and analyze your fitness prog...
✓ Apple mobile web app capable tag present
✓ Apple touch icon present (good for iOS)
```

### 6. Lighthouse PWA Audit (Optional) ✓

**Requires Lighthouse** - Install with `npm install -g lighthouse`

- Runs full Lighthouse PWA audit
- Scores on 0-100 scale
- Identifies specific failed audits
- Provides actionable recommendations

**Example output:**
```
━━━ 7. Lighthouse PWA Audit (Optional) ━━━

ℹ Running Lighthouse PWA audit (this may take 30-60 seconds)...
✓ Lighthouse PWA Score: 92% (Excellent!)
   Key findings:
   ✗ Does not provide a valid apple-touch-icon
```

**Score interpretation:**
- **90-100%** - Excellent! PWA is production-ready
- **70-89%** - Good, but can be improved
- **0-69%** - Needs significant improvement

### 7. Performance Basics ✓

**Important for user experience**

- ✓ Measures page load time
- ✓ Checks HTTP/2 support
- ✓ Detects compression (gzip/brotli)

**Example output:**
```
━━━ 8. Basic Performance Check ━━━

✓ Page load time: 324ms (Excellent!)
✓ HTTP/2 supported (better performance)
✓ Compression enabled: gzip
```

**Load time benchmarks:**
- **< 1000ms** - Excellent
- **1000-3000ms** - Acceptable
- **> 3000ms** - Too slow

### 8. Summary Report

Final report with overall status and recommendations:

```
═══════════════════════════════════════════════════════════════
                  PWA Health Check Summary
═══════════════════════════════════════════════════════════════

Target URL: https://example.com

Results:
  Passed:   24
  Failed:   0
  Warnings: 2
  Total:    26

✓ PWA Status: EXCELLENT
Your site meets all critical PWA requirements and is ready for installation!

ℹ Next Steps:
1. Fix any failed checks (marked with ✗)
2. Address warnings for better user experience (marked with ⚠)
3. Test installation on mobile devices
4. Test offline functionality
5. Consider running full Lighthouse audit for detailed analysis
```

## PWA Status Levels

The script categorizes PWA readiness into four levels:

### ✓ EXCELLENT (Green)
- **0 failures**
- **90%+ tests passed**
- Site is production-ready and fully installable
- May have minor warnings for optional features

### ✓ GOOD (Green)
- **0 failures**
- **< 90% tests passed**
- Site meets PWA requirements
- Consider addressing warnings

### ⚠ NEEDS IMPROVEMENT (Yellow)
- **1-2 failures**
- Some PWA features present
- Requires fixes to be fully installable

### ✗ NOT A PWA (Red)
- **3+ failures**
- Does not meet PWA requirements
- Must address critical failed checks

## Common Issues and Fixes

### Issue: "Site is NOT served over HTTPS"

**Fix:**
```bash
# Development: Use localhost (allowed without HTTPS)
npm run dev  # Usually runs on http://localhost:3000

# Production: Get SSL certificate
# Option 1: Let's Encrypt (free)
sudo certbot --nginx -d yourdomain.com

# Option 2: Use Caddy (automatic HTTPS)
caddy run
```

### Issue: "No manifest.json link found"

**Fix:**
```html
<!-- Add to <head> section of index.html -->
<link rel="manifest" href="/manifest.json">
```

### Issue: "Missing 192x192 icon"

**Fix in manifest.json:**
```json
{
  "icons": [
    {
      "src": "/icons/icon-192x192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any maskable"
    },
    {
      "src": "/icons/icon-512x512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "any maskable"
    }
  ]
}
```

### Issue: "No service worker registration found"

**Fix in main JavaScript:**
```javascript
// Register service worker
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/sw.js')
    .then(registration => {
      console.log('Service Worker registered:', registration);
    })
    .catch(error => {
      console.error('Service Worker registration failed:', error);
    });
}
```

### Issue: "Missing viewport meta tag"

**Fix in HTML:**
```html
<!-- Add to <head> section -->
<meta name="viewport" content="width=device-width, initial-scale=1">
```

### Issue: "Display mode may not provide app-like experience"

**Fix in manifest.json:**
```json
{
  "display": "standalone"
}
```

Options: `standalone`, `fullscreen`, `minimal-ui`

## Advanced Usage

### Check ActaLog PWA

```bash
# Local development
./pwa-health-check.sh http://localhost:3000

# Production
./pwa-health-check.sh https://actalog.yourdomain.com
```

### Automated Testing (CI/CD)

```bash
# Run in CI pipeline
./pwa-health-check.sh https://staging.example.com

# Check exit code
if [ $? -eq 0 ]; then
  echo "PWA checks passed"
else
  echo "PWA checks failed"
fi
```

### Generate Report for Team

```bash
# Redirect output to file
./pwa-health-check.sh https://example.com > pwa-report.txt

# Share with team
cat pwa-report.txt
```

### Test Multiple Environments

```bash
# Create test script
for env in dev staging prod; do
  echo "Testing $env..."
  ./pwa-health-check.sh https://$env.example.com
  echo ""
done
```

## Troubleshooting

### Error: "Missing required dependencies"

Install missing tools:
```bash
sudo apt-get update
sudo apt-get install curl jq openssl
```

### Error: "Failed to fetch page content"

Possible causes:
- Site is down or unreachable
- Firewall blocking curl
- Invalid URL
- Timeout (default 10 seconds)

**Solution:** Increase timeout by editing script:
```bash
TIMEOUT=30  # Change from 10 to 30 seconds
```

### Lighthouse Not Running

Install Lighthouse:
```bash
# Global installation
npm install -g lighthouse

# Or use npx (no installation)
npx lighthouse https://example.com --view
```

### False Negatives for Service Worker

The script checks common filenames. If your service worker has a custom name:

1. Check the HTML source manually
2. Look for `navigator.serviceWorker.register('/custom-sw.js')`
3. The script will detect the registration code even if it can't find the file

## Performance Tips

### For Faster Checks

```bash
# Skip Lighthouse (fastest)
# Comment out line 535 in script:
# run_lighthouse
```

### For More Thorough Checks

```bash
# Install Lighthouse for detailed audit
npm install -g lighthouse

# Run full Lighthouse report separately
lighthouse https://example.com --view
```

## Output Examples

### Perfect PWA
```
Results:
  Passed:   26
  Failed:   0
  Warnings: 0
  Total:    26

✓ PWA Status: EXCELLENT
```

### Good PWA with Warnings
```
Results:
  Passed:   22
  Failed:   0
  Warnings: 4
  Total:    26

✓ PWA Status: GOOD
```

### Needs Improvement
```
Results:
  Passed:   18
  Failed:   2
  Warnings: 6
  Total:    26

⚠ PWA Status: NEEDS IMPROVEMENT
```

### Not a PWA
```
Results:
  Passed:   8
  Failed:   12
  Warnings: 6
  Total:    26

✗ PWA Status: NOT A PWA
```

## PWA Checklist

Use this script to validate your PWA checklist:

- [ ] Served over HTTPS
- [ ] Has valid manifest.json
- [ ] Manifest has name/short_name
- [ ] Manifest has 192x192 icon
- [ ] Manifest has 512x512 icon
- [ ] Manifest has start_url
- [ ] Manifest has display: standalone
- [ ] Has service worker
- [ ] Service worker handles fetch events
- [ ] Has viewport meta tag
- [ ] Has theme-color
- [ ] Responsive design
- [ ] Fast load time (< 3s)
- [ ] Works offline

## Resources

### PWA Documentation
- [MDN PWA Guide](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps)
- [web.dev PWA](https://web.dev/progressive-web-apps/)
- [Google PWA Checklist](https://web.dev/pwa-checklist/)

### Tools
- [Lighthouse](https://github.com/GoogleChrome/lighthouse)
- [PWA Builder](https://www.pwabuilder.com/)
- [Manifest Generator](https://www.simicart.com/manifest-generator.html/)

### Testing
- [Chrome DevTools Application Panel](https://developer.chrome.com/docs/devtools/progressive-web-apps/)
- [PWA Testing Tool](https://www.pwabuilder.com/)

## Support

For issues with the PWA health check script:
1. Ensure all dependencies are installed (`curl`, `jq`, `openssl`)
2. Check that the URL is accessible
3. Review the detailed output for specific failures
4. Consult the "Common Issues and Fixes" section above
5. Open an issue on GitHub with the full output

## License

MIT License - Free to use and modify
