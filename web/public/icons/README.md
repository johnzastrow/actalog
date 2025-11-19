# PWA Icons

This directory contains app icons for the Progressive Web App (PWA) installation.

## Required Icon Sizes

The following icon sizes are required for optimal PWA support across all devices:

- 72x72 (Android Chrome)
- 96x96 (Android Chrome)
- 128x128 (Android Chrome, Chrome Web Store)
- 144x144 (Microsoft)
- 152x152 (iOS Safari)
- 192x192 (Android Chrome - minimum required)
- 384x384 (Android Chrome)
- 512x512 (Android Chrome - recommended for splash screens)

Plus special icons:
- **apple-touch-icon.png** (180x180) - iOS home screen
- **favicon.ico** (32x32) - Browser tab icon

## Generating Icons

### Automated Script (Recommended)

We've created an automated icon generation script that creates all required sizes from the source SVG:

# PWA Icons

This directory contains app icons used by the ActaLog Progressive Web App (PWA).

## Required Icon Sizes

The following icon sizes are recommended for broad device support:

- 72x72
- 96x96
- 128x128
- 144x144
- 152x152
- 192x192 (minimum for Android)
- 384x384
- 512x512 (recommended for splash screens)

Plus special icons:

- `apple-touch-icon.png` (180x180) — iOS home screen
- `favicon.ico` (32x32) — Browser tab icon

## Generating Icons

### Automated (recommended)

From the `web/` directory the repo includes a small script to generate PNG icons from the SVG source:

```bash
# from repo root
cd web
node scripts/generate-icons.js
```

The script creates the PNG sizes above, `apple-touch-icon.png`, and `favicon.ico` using `design/logo.svg` as the source. The script depends on the `sharp` package (in devDependencies).

### Manual / Online Tools

 PWA Asset Generator: [PWA Asset Generator](https://www.pwabuilder.com/imageGenerator)
 RealFaviconGenerator: [RealFaviconGenerator](https://realfavicongenerator.net/)

Use the tools above to generate icon packs and copy the produced files into `web/public/icons/`.

### ImageMagick example

```bash
# Convert SVG to PNG at various sizes (requires ImageMagick)
convert design/logo.svg -resize 72x72 web/public/icons/icon-72x72.png
convert design/logo.svg -resize 96x96 web/public/icons/icon-96x96.png
convert design/logo.svg -resize 128x128 web/public/icons/icon-128x128.png
convert design/logo.svg -resize 144x144 web/public/icons/icon-144x144.png
convert design/logo.svg -resize 152x152 web/public/icons/icon-152x152.png
convert design/logo.svg -resize 192x192 web/public/icons/icon-192x192.png
convert design/logo.svg -resize 384x384 web/public/icons/icon-384x384.png
convert design/logo.svg -resize 512x512 web/public/icons/icon-512x512.png

# Apple touch icon
convert design/logo.svg -resize 180x180 web/public/apple-touch-icon.png

# Favicon
convert design/logo.svg -resize 32x32 web/public/favicon.ico
```

## Manifest / VitePWA

Icons are referenced in the VitePWA manifest configuration (see `web/vite.config.js`). After building (`npm run build`) verify `dist/manifest.webmanifest` contains the expected icon entries and that the files are present in `dist/`.

## Verification

- Dev: `cd web && npm run dev` then open DevTools → Application → Manifest
- Preview: `cd web && npm run build && npm run preview` then run the PWA health-check script: `./scripts/pwa-health-check.sh http://localhost:4173`
- Lighthouse: run a Lighthouse PWA audit in DevTools

## Troubleshooting

- Icons missing in production: check `dist/manifest.webmanifest` and confirm files were copied to `dist/`.
- Browser caching: clear cache and reload (or use an incognito window).
- `sharp` install failures: try `npm cache clean --force` then `npm install` inside `web/`.

## Status

Generated icons (if present):

- icon-72x72.png
- icon-96x96.png
- icon-128x128.png
- icon-144x144.png
- icon-152x152.png
- icon-192x192.png
- icon-384x384.png
- icon-512x512.png
- apple-touch-icon.png (180x180)
- favicon.ico (32x32)

---

If you'd like, I can convert the remaining bare URLs to markdown link format across the repo, or add a short script entry `web/package.json` like `dev:https` to make running HTTPS dev servers easier. Which would you prefer?
- Create additional favicon sizes
- Use the ActaLog logo from `design/logo.svg`
- Apply the correct theme color background (`#00bcd4`)

**Prerequisites**: The `sharp` package must be installed (automatically included in devDependencies).

### Manual Methods (Alternative)

#### Option 1: Using Online Tools

1. **PWA Asset Generator**: [PWA Asset Generator](https://www.pwabuilder.com/imageGenerator)
   - Upload your source logo (`design/logo.png` or `logo.svg`)
   - Download generated icons
   - Extract to this directory

2. **RealFaviconGenerator**: [RealFaviconGenerator](https://realfavicongenerator.net/)
   - Upload source image
   - Configure iOS, Android, and Windows settings
   - Generate and download icons

#### Option 2: Using ImageMagick (Command Line)

```bash
# Navigate to project root
cd /path/to/actalog

# Convert SVG to PNG at various sizes
convert design/logo.svg -resize 72x72 web/public/icons/icon-72x72.png
convert design/logo.svg -resize 96x96 web/public/icons/icon-96x96.png
convert design/logo.svg -resize 128x128 web/public/icons/icon-128x128.png
convert design/logo.svg -resize 144x144 web/public/icons/icon-144x144.png
convert design/logo.svg -resize 152x152 web/public/icons/icon-152x152.png
convert design/logo.svg -resize 192x192 web/public/icons/icon-192x192.png
convert design/logo.svg -resize 384x384 web/public/icons/icon-384x384.png
convert design/logo.svg -resize 512x512 web/public/icons/icon-512x512.png

# Apple touch icon
convert design/logo.svg -resize 180x180 web/public/apple-touch-icon.png

# Favicon
convert design/logo.svg -resize 32x32 web/public/favicon.ico
```

## Design Guidelines

- **Background**: Icons use ActaLog's theme color (`#00bcd4`)
- **Safe Zone**: Keep important content within center 80% for maskable icons
- **Format**: PNG format with proper transparency
- **Source**: Always generate from `design/logo.svg` for consistency

## Manifest Configuration

Icons are configured in `vite.config.js` under the VitePWA plugin's manifest section. The configuration includes:

- Multiple icon sizes for different devices
- Proper purpose declarations (`any`/`maskable`)
- Correct MIME types

## Verification

After generating icons, verify they appear correctly:

1. **Development**:

   ```bash
   npm run dev
   ```

   Check: DevTools → Application → Manifest

2. **Production Build**:

   ```bash
   npm run build
   ```

   # PWA Icons

   This directory contains app icons for the Progressive Web App (PWA) installation.

   ## Required Icon Sizes

   The following icon sizes are required for optimal PWA support across all devices:

   - 72x72 (Android Chrome)
   - 96x96 (Android Chrome)
   - 128x128 (Android Chrome, Chrome Web Store)
   - 144x144 (Microsoft)
   - 152x152 (iOS Safari)
   - 192x192 (Android Chrome - minimum required)
   - 384x384 (Android Chrome)
   - 512x512 (Android Chrome - recommended for splash screens)

   Plus special icons:

   - **apple-touch-icon.png** (180x180) - iOS home screen
   - **favicon.ico** (32x32) - Browser tab icon

   ## Generating Icons

   ### Automated Script (Recommended)

   We've created an automated icon generation script that creates all required sizes from the source SVG.

   ```bash
   # From the web directory
   cd web

   # Run the icon generator
   node scripts/generate-icons.js
   ```

   This script will:

   - Generate all 8 PNG icon sizes (72px - 512px)
   - Create `apple-touch-icon.png` for iOS
   - Create `favicon.ico` for browsers
   - Create additional favicon sizes
   - Use the ActaLog logo from `design/logo.svg`
   - Apply the correct theme color background (`#00bcd4`)

   **Prerequisites**: The `sharp` package must be installed (automatically included in devDependencies).

   ### Manual Methods (Alternative)

   #### Option 1: Using Online Tools

   1. **PWA Asset Generator**: [PWA Asset Generator](https://www.pwabuilder.com/imageGenerator)
      - Upload your source logo (`design/logo.png` or `logo.svg`)
      - Download generated icons
      - Extract to this directory

   2. **RealFaviconGenerator**: [RealFaviconGenerator](https://realfavicongenerator.net/)
      - Upload source image
      - Configure iOS, Android, and Windows settings
      - Generate and download icons

   #### Option 2: Using ImageMagick (Command Line)

   ```bash
   # Navigate to project root
   cd /path/to/actalog

   # Convert SVG to PNG at various sizes
   convert design/logo.svg -resize 72x72 web/public/icons/icon-72x72.png
   convert design/logo.svg -resize 96x96 web/public/icons/icon-96x96.png
   convert design/logo.svg -resize 128x128 web/public/icons/icon-128x128.png
   convert design/logo.svg -resize 144x144 web/public/icons/icon-144x144.png
   convert design/logo.svg -resize 152x152 web/public/icons/icon-152x152.png
   convert design/logo.svg -resize 192x192 web/public/icons/icon-192x192.png
   convert design/logo.svg -resize 384x384 web/public/icons/icon-384x384.png
   convert design/logo.svg -resize 512x512 web/public/icons/icon-512x512.png

   # Apple touch icon
   convert design/logo.svg -resize 180x180 web/public/apple-touch-icon.png

   # Favicon
   convert design/logo.svg -resize 32x32 web/public/favicon.ico
   ```

   ## Design Guidelines

   - **Background**: Icons use ActaLog's theme color (`#00bcd4`)
   - **Safe Zone**: Keep important content within center 80% for maskable icons
   - **Format**: PNG format with proper transparency
   - **Source**: Always generate from `design/logo.svg` for consistency

   ## Manifest Configuration

   Icons are configured in `vite.config.js` under the VitePWA plugin's manifest section. The configuration includes:

   - Multiple icon sizes for different devices
   - Proper purpose declarations (`any`/`maskable`)
   - Correct MIME types

   ## Verification

   After generating icons, verify they appear correctly:

   1. **Development**:

      ```bash
      npm run dev
      ```

      Check: DevTools → Application → Manifest

   2. **Production Build**:

      ```bash
      npm run build
      ```

      # PWA Icons

      This directory contains app icons for the Progressive Web App (PWA) installation.

      ## Required Icon Sizes

      The following icon sizes are required for optimal PWA support across all devices:

      - 72x72 (Android Chrome)
      - 96x96 (Android Chrome)
      - 128x128 (Android Chrome, Chrome Web Store)
      - 144x144 (Microsoft)
      - 152x152 (iOS Safari)
      - 192x192 (Android Chrome - minimum required)
      - 384x384 (Android Chrome)
      - 512x512 (Android Chrome - recommended for splash screens)

      Plus special icons:

      - **apple-touch-icon.png** (180x180) - iOS home screen
      - **favicon.ico** (32x32) - Browser tab icon

      ## Generating Icons

      ### Automated Script (Recommended)

      We've created an automated icon generation script that creates all required sizes from the source SVG.

      ```bash
      # From the web directory
      cd web

      # Run the icon generator
      node scripts/generate-icons.js
      ```

      This script will:

      - Generate all 8 PNG icon sizes (72px - 512px)
      - Create `apple-touch-icon.png` for iOS
      - Create `favicon.ico` for browsers
      - Create additional favicon sizes
      - Use the ActaLog logo from `design/logo.svg`
      - Apply the correct theme color background (`#00bcd4`)

      **Prerequisites**: The `sharp` package must be installed (automatically included in devDependencies).

      ### Manual Methods (Alternative)

      #### Option 1: Using Online Tools

      1. **PWA Asset Generator**: [PWA Asset Generator](https://www.pwabuilder.com/imageGenerator)
         - Upload your source logo (`design/logo.png` or `logo.svg`)
         - Download generated icons
         - Extract to this directory

      2. **RealFaviconGenerator**: [RealFaviconGenerator](https://realfavicongenerator.net/)
         - Upload source image
         - Configure iOS, Android, and Windows settings
         - Generate and download icons

      #### Option 2: Using ImageMagick (Command Line)

      ```bash
      # Navigate to project root
      cd /path/to/actalog

      # Convert SVG to PNG at various sizes
      convert design/logo.svg -resize 72x72 web/public/icons/icon-72x72.png
      convert design/logo.svg -resize 96x96 web/public/icons/icon-96x96.png
      convert design/logo.svg -resize 128x128 web/public/icons/icon-128x128.png
      convert design/logo.svg -resize 144x144 web/public/icons/icon-144x144.png
      convert design/logo.svg -resize 152x152 web/public/icons/icon-152x152.png
      convert design/logo.svg -resize 192x192 web/public/icons/icon-192x192.png
      convert design/logo.svg -resize 384x384 web/public/icons/icon-384x384.png
      convert design/logo.svg -resize 512x512 web/public/icons/icon-512x512.png

      # Apple touch icon
      convert design/logo.svg -resize 180x180 web/public/apple-touch-icon.png

      # Favicon
      convert design/logo.svg -resize 32x32 web/public/favicon.ico
      ```

      ## Design Guidelines

      - **Background**: Icons use ActaLog's theme color (`#00bcd4`)
      - **Safe Zone**: Keep important content within center 80% for maskable icons
      - **Format**: PNG format with proper transparency
      - **Source**: Always generate from `design/logo.svg` for consistency

      ## Manifest Configuration

      Icons are configured in `vite.config.js` under the VitePWA plugin's manifest section. The configuration includes:

      - Multiple icon sizes for different devices
      - Proper purpose declarations (`any`/`maskable`)
      - Correct MIME types

      ## Verification

      After generating icons, verify they appear correctly:

      1. **Development**:

         ```bash
         npm run dev
         ```

         Check: DevTools → Application → Manifest

      2. **Production Build**:

         ```bash
         npm run build
         ```

         Verify all icons are copied to `dist/` directory

      3. **Lighthouse Audit**:

         - Open DevTools → Lighthouse
         - Run PWA audit
         - Target score: 90+

      4. **Visual Check**:

         ```bash
         ls -lh web/public/icons/
         ls -lh web/public/apple-touch-icon.png web/public/favicon.ico
         ```

      ## Current Status

      ✅ **Icons Generated** - All required icons have been created and are ready for deployment.

      Generated icons:

      - icon-72x72.png
      - icon-96x96.png
      - icon-128x128.png
      - icon-144x144.png
      - icon-152x152.png
      - icon-192x192.png
      - icon-384x384.png
      - icon-512x512.png
      - apple-touch-icon.png (180x180)
      - favicon.ico (32x32)
      - Additional favicon sizes (16x16, 32x32, 48x48)

      ## Troubleshooting

      **If icons don't appear after building:**

      1. Clear browser cache
      2. Rebuild the project: `npm run build`
      3. Verify `manifest.webmanifest` in `dist/` includes all icon paths
      4. Check browser console for 404 errors

      **If sharp installation fails:**

      ```bash
      cd web
      npm cache clean --force
      npm install --save-dev sharp
      ```

      **Regenerate icons after logo changes:**

      ```bash
      cd web
      node scripts/generate-icons.js
      npm run build
      ```
