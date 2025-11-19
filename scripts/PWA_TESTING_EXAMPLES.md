# PWA Health Check - Testing Examples

## Quick Start

### Test Your Local Development

```bash
# Make sure your app is running first
cd web
npm run dev  # Should start on http://localhost:3000

# In another terminal, run the PWA check
cd ..
./pwa-health-check.sh http://localhost:3000
```

**Expected Results for Localhost:**
- ⚠ Warning about HTTPS (expected for localhost)
- All other checks should pass if PWA is configured

### Test Production Site

```bash
./pwa-health-check.sh https://yourdomain.com
```

## Example Output Walkthrough

### 1. Perfect PWA (All Green)

```
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║              PWA Health Check Tool v1.0                       ║
║              Progressive Web App Validator                    ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝

Checking: https://actalog.example.com

━━━ 1. HTTPS Connection ━━━

✓ Site is served over HTTPS
✓ SSL certificate is valid

━━━ 2. Fetching Page Content ━━━

✓ Successfully fetched page content
   Content size: 45321 bytes

━━━ 3. Web App Manifest ━━━

✓ Manifest link found in HTML: /manifest.json
   Manifest URL: https://actalog.example.com/manifest.json
✓ Manifest fetched and is valid JSON

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
✓ Description: "Track your CrossFit workouts and personal records"

━━━ 5. Service Worker ━━━

✓ Service worker registration code found
✓ Service worker file found: sw.js (HTTP 200)
✓ Service worker has fetch event handler (offline support)
✓ Service worker has install event handler
✓ Service worker has activate event handler

━━━ 6. HTML Meta Tags ━━━

✓ Viewport meta tag present
   Content: width=device-width, initial-scale=1
✓ Theme color meta tag: #00bcd4
✓ Description meta tag present
   Track your CrossFit workouts, log personal records, and analyze your...
✓ Apple mobile web app capable tag present
✓ Apple touch icon present (good for iOS)

━━━ 7. Lighthouse PWA Audit (Optional) ━━━

ℹ Running Lighthouse PWA audit (this may take 30-60 seconds)...
✓ Lighthouse PWA Score: 100% (Excellent!)

━━━ 8. Basic Performance Check ━━━

✓ Page load time: 287ms (Excellent!)
✓ HTTP/2 supported (better performance)
✓ Compression enabled: gzip

═══════════════════════════════════════════════════════════════
                  PWA Health Check Summary
═══════════════════════════════════════════════════════════════

Target URL: https://actalog.example.com

Results:
  Passed:   26
  Failed:   0
  Warnings: 0
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

### 2. Localhost Development (Expected Warnings)

```bash
./pwa-health-check.sh http://localhost:3000
```

**Output:**
```
━━━ 1. HTTPS Connection ━━━

✗ Site is NOT served over HTTPS (PWA requires HTTPS)
   HTTP is only allowed for localhost during development

━━━ 2. Fetching Page Content ━━━

✓ Successfully fetched page content

━━━ 3. Web App Manifest ━━━

✓ Manifest link found in HTML: /manifest.json
✓ Manifest fetched and is valid JSON

[... rest of checks ...]

Results:
  Passed:   20
  Failed:   1
  Warnings: 2
  Total:    23

⚠ PWA Status: NEEDS IMPROVEMENT
Your site has some PWA features but needs fixes to be fully installable.
```

**Note:** The HTTPS failure is expected for localhost. This is acceptable during development.

### 3. Missing Manifest (Common Issue)

```
━━━ 3. Web App Manifest ━━━

✗ No manifest.json link found in HTML
   Add: <link rel="manifest" href="/manifest.json">

━━━ 4. Manifest Properties ━━━

(Skipped - no manifest found)

Results:
  Passed:   12
  Failed:   8
  Warnings: 3
  Total:    23

✗ PWA Status: NOT A PWA
Your site does not meet PWA requirements. Address the failed checks above.
```

**Fix:** Add manifest link to `index.html`:
```html
<link rel="manifest" href="/manifest.json">
```

### 4. Missing Service Worker

```
━━━ 5. Service Worker ━━━

✗ No service worker registration found
   PWA requires a service worker for offline functionality
⚠ Could not locate service worker file (checked: sw.js service-worker.js serviceworker.js)
   Service worker might use a different filename
```

**Fix:** Register service worker in your main JavaScript:
```javascript
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/sw.js');
}
```

### 5. Good PWA with Warnings

```
Results:
  Passed:   22
  Failed:   0
  Warnings: 4
  Total:    26

✓ PWA Status: GOOD
Your site meets PWA requirements. Consider addressing warnings for optimal experience.
```

**Common warnings:**
- ⚠ Missing 512x512 icon (recommended for splash screen)
- ⚠ Missing 'theme_color' (recommended)
- ⚠ Missing 'background_color' (recommended)
- ⚠ Missing apple-touch-icon (optional for iOS)

## ActaLog-Specific Testing

### Current ActaLog Setup (as of v0.4.5)

ActaLog should have:
- ✓ Vue.js 3 with Vite
- ✓ Vuetify 3 UI framework
- ✓ Service worker via Vite PWA plugin
- ✓ Manifest.json with proper configuration

### Testing ActaLog Locally

```bash
# 1. Start the development server
cd web
npm run dev

# 2. In another terminal, run PWA check
cd ..
./pwa-health-check.sh http://localhost:3000

# 3. Build for production and test
cd web
npm run build
npm run preview  # Starts preview server

# 4. Check the preview
cd ..
./pwa-health-check.sh http://localhost:4173
```

### Testing ActaLog Production

```bash
# Replace with your actual domain
./pwa-health-check.sh https://actalog.yourdomain.com
```

### Expected ActaLog Results

**Development (localhost:3000):**
```
Results:
  Passed:   20-22
  Failed:   1 (HTTPS - expected)
  Warnings: 2-4
  Total:    23-26

Status: NEEDS IMPROVEMENT (due to HTTP, acceptable for dev)
```

**Production (https://domain):**
```
Results:
  Passed:   24-26
  Failed:   0
  Warnings: 0-2
  Total:    26

Status: EXCELLENT or GOOD
```

## Common Testing Scenarios

### Before Deployment Checklist

```bash
# 1. Build production version
cd web
npm run build

# 2. Test production build locally
npm run preview

# 3. Run PWA check on preview
cd ..
./pwa-health-check.sh http://localhost:4173

# 4. Fix any failures
# 5. Deploy
# 6. Test production URL
./pwa-health-check.sh https://yourdomain.com
```

### After Deployment Verification

```bash
# Test main domain
./pwa-health-check.sh https://yourdomain.com

# Test with www
./pwa-health-check.sh https://www.yourdomain.com

# Test subdomain (if applicable)
./pwa-health-check.sh https://app.yourdomain.com
```

### Mobile Testing Workflow

1. **Run PWA check on desktop:**
   ```bash
   ./pwa-health-check.sh https://yourdomain.com
   ```

2. **If all checks pass, test on mobile:**
   - Open Chrome on Android
   - Visit your URL
   - Look for "Install" prompt in menu
   - Install and test offline functionality

3. **Test on iOS:**
   - Open Safari on iPhone
   - Visit your URL
   - Tap Share → Add to Home Screen
   - Test app experience

### CI/CD Integration

```bash
# In your CI/CD pipeline (e.g., GitHub Actions)

- name: Build PWA
  run: |
    cd web
    npm install
    npm run build

- name: Preview Build
  run: |
    cd web
    npm run preview &
    sleep 5

- name: PWA Health Check
  run: |
    ./pwa-health-check.sh http://localhost:4173

- name: Deploy
  if: success()
  run: |
    # Your deployment commands
```

## Interpreting Results

### Critical Failures (Must Fix)

These prevent PWA installation:
- ✗ Site NOT served over HTTPS (production)
- ✗ No manifest.json
- ✗ Missing name/short_name in manifest
- ✗ Missing 192x192 icon
- ✗ No service worker

### Important Warnings (Should Fix)

These affect user experience:
- ⚠ Missing 512x512 icon (splash screen won't look good)
- ⚠ Missing theme_color (no colored browser UI)
- ⚠ Missing viewport tag (not responsive)
- ⚠ Slow load time (poor UX)

### Optional Info (Nice to Have)

These enhance the experience:
- ℹ Apple touch icon (better iOS experience)
- ℹ HTTP/2 support (better performance)
- ℹ Compression (faster loading)

## Troubleshooting Common Issues

### Issue: "Connection timed out"

**Cause:** Site is slow or unreachable

**Fix:** Increase timeout in script:
```bash
# Edit pwa-health-check.sh
TIMEOUT=30  # Change from 10 to 30
```

### Issue: "Manifest not found (404)"

**Cause:** Manifest path is incorrect

**Fix:** Check manifest link in HTML:
```html
<!-- Make sure this matches your actual file location -->
<link rel="manifest" href="/manifest.json">
```

### Issue: "Invalid JSON in manifest"

**Cause:** Syntax error in manifest.json

**Fix:** Validate JSON:
```bash
# Use jq to validate
cat web/public/manifest.json | jq .

# Or use online validator
# https://jsonlint.com/
```

### Issue: "Service worker not detected"

**Cause:** Service worker might have custom filename

**Fix:** Check your service worker registration:
```javascript
// Look for this in your JavaScript
navigator.serviceWorker.register('/custom-sw.js')
```

## Next Steps After Green Results

Once you get all green checkmarks:

1. **Test installation on real devices:**
   - Android: Chrome menu → Install app
   - iOS: Safari → Share → Add to Home Screen

2. **Test offline functionality:**
   - Install the app
   - Turn off WiFi/data
   - Open app and verify it works

3. **Test app update flow:**
   - Make changes to your app
   - Deploy new version
   - Verify service worker updates properly

4. **Monitor with Lighthouse:**
   ```bash
   lighthouse https://yourdomain.com --view
   ```

5. **Add to app stores (optional):**
   - [Google Play](https://developers.google.com/web/android/trusted-web-activity)
   - [Microsoft Store](https://www.pwabuilder.com/)

## Additional Resources

### Testing Tools
- Chrome DevTools → Application panel
- Firefox → Application → Manifest
- [PWA Builder Testing](https://www.pwabuilder.com/)

### Validation
- [Lighthouse CI](https://github.com/GoogleChrome/lighthouse-ci)
- [PWA Feature Detector](https://www.pwafeatures.com/)

### Debugging
- Chrome: `chrome://serviceworker-internals/`
- Firefox: `about:debugging#/runtime/this-firefox`
