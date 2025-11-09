# PWA Icons

This directory contains app icons for the Progressive Web App (PWA) installation.

## Required Icon Sizes

The following icon sizes are required for optimal PWA support across all devices:

- 72x72 (Android)
- 96x96 (Android)
- 128x128 (Android, Chrome Web Store)
- 144x144 (Windows)
- 152x152 (iOS)
- 192x192 (Android, standard)
- 384x384 (Android)
- 512x512 (Android, splash screens)

## Generating Icons

### Option 1: Using Online Tools (Easiest)

1. **PWA Asset Generator**: https://www.pwabuilder.com/imageGenerator
   - Upload your source logo (design/logo.png or logo.svg)
   - Download generated icons
   - Extract to this directory

2. **RealFaviconGenerator**: https://realfavicongenerator.net/
   - Upload source image
   - Configure iOS, Android, and Windows settings
   - Generate and download icons

### Option 2: Using ImageMagick (Command Line)

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
```

### Option 3: Using Sharp (Node.js)

Install sharp:
```bash
npm install -g sharp-cli
```

Generate icons:
```bash
sharp -i design/logo.png -o web/public/icons/icon-72x72.png resize 72 72
sharp -i design/logo.png -o web/public/icons/icon-96x96.png resize 96 96
sharp -i design/logo.png -o web/public/icons/icon-128x128.png resize 128 128
sharp -i design/logo.png -o web/public/icons/icon-144x144.png resize 144 144
sharp -i design/logo.png -o web/public/icons/icon-152x152.png resize 152 152
sharp -i design/logo.png -o web/public/icons/icon-192x192.png resize 192 192
sharp -i design/logo.png -o web/public/icons/icon-384x384.png resize 384 384
sharp -i design/logo.png -o web/public/icons/icon-512x512.png resize 512 512
```

## iOS-Specific Icons

For better iOS support, also create:

```bash
# Apple touch icon (recommended 180x180)
convert design/logo.svg -resize 180x180 web/public/apple-touch-icon.png
```

## Favicon

Create a standard favicon:
```bash
convert design/logo.svg -resize 32x32 web/public/favicon.ico
```

## Design Guidelines

- **Background**: Icons should have a background color (use ActaLog theme: #2c3657 or #00bcd4)
- **Padding**: Add 10-15% padding around the logo for better appearance
- **Transparency**: PNG format supports transparency for non-square logos
- **Maskable**: Icons should work with Android's maskable icon feature (safe zone in center 80%)

## Verification

After generating icons, verify they appear correctly:

1. **Development**: Run `npm run dev` and check browser DevTools → Application → Manifest
2. **Production**: Build with `npm run build` and test the PWA install prompt
3. **Lighthouse**: Run Lighthouse audit to verify all icons are present

## Current Status

⚠️ **Icons need to be generated** - This directory should contain 8 icon files before deployment.
