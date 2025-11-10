#!/usr/bin/env node
/**
 * PWA Icon Generator
 *
 * Generates all required PWA icon sizes from the source SVG logo.
 * Requires: sharp package (npm install --save-dev sharp)
 *
 * Usage: node scripts/generate-icons.js
 */

import sharp from 'sharp';
import { promises as fs } from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Configuration
const SOURCE_SVG = path.join(__dirname, '../../design/logo.svg');
const OUTPUT_DIR = path.join(__dirname, '../public/icons');

// Required icon sizes for PWA
const ICON_SIZES = [
  72,   // Android Chrome
  96,   // Android Chrome
  128,  // Android Chrome
  144,  // Microsoft
  152,  // iOS Safari
  192,  // Android Chrome (minimum required)
  384,  // Android Chrome
  512   // Android Chrome (recommended)
];

// Special icons
const APPLE_TOUCH_ICON_SIZE = 180;  // iOS home screen icon
const FAVICON_SIZES = [16, 32, 48]; // Browser favicon

async function ensureDirectory(dir) {
  try {
    await fs.access(dir);
  } catch {
    await fs.mkdir(dir, { recursive: true });
    console.log(`✓ Created directory: ${dir}`);
  }
}

async function generateIcon(size, filename) {
  try {
    const outputPath = path.join(OUTPUT_DIR, filename);

    await sharp(SOURCE_SVG)
      .resize(size, size, {
        fit: 'contain',
        background: { r: 0, g: 188, b: 212, alpha: 1 } // #00bcd4
      })
      .png()
      .toFile(outputPath);

    console.log(`✓ Generated ${filename} (${size}x${size})`);
  } catch (error) {
    console.error(`✗ Failed to generate ${filename}:`, error.message);
  }
}

async function main() {
  console.log('🎨 ActaLog PWA Icon Generator\n');
  console.log(`Source: ${SOURCE_SVG}`);
  console.log(`Output: ${OUTPUT_DIR}\n`);

  // Ensure output directory exists
  await ensureDirectory(OUTPUT_DIR);

  // Generate PWA icons
  console.log('Generating PWA icons...');
  for (const size of ICON_SIZES) {
    await generateIcon(size, `icon-${size}x${size}.png`);
  }

  // Generate Apple Touch Icon (special location)
  console.log('\nGenerating Apple Touch Icon...');
  const appleIconPath = path.join(__dirname, '../public/apple-touch-icon.png');
  await sharp(SOURCE_SVG)
    .resize(APPLE_TOUCH_ICON_SIZE, APPLE_TOUCH_ICON_SIZE, {
      fit: 'contain',
      background: { r: 0, g: 188, b: 212, alpha: 1 }
    })
    .png()
    .toFile(appleIconPath);
  console.log(`✓ Generated apple-touch-icon.png (${APPLE_TOUCH_ICON_SIZE}x${APPLE_TOUCH_ICON_SIZE})`);

  // Generate favicons
  console.log('\nGenerating favicons...');
  for (const size of FAVICON_SIZES) {
    await generateIcon(size, `favicon-${size}x${size}.png`);
  }

  // Generate main favicon.ico (using 32x32 as base)
  const faviconPath = path.join(__dirname, '../public/favicon.ico');
  await sharp(SOURCE_SVG)
    .resize(32, 32, {
      fit: 'contain',
      background: { r: 0, g: 188, b: 212, alpha: 1 }
    })
    .toFormat('png')
    .toFile(faviconPath);
  console.log(`✓ Generated favicon.ico (32x32)`);

  console.log('\n✅ Icon generation complete!');
  console.log('\nNext steps:');
  console.log('1. Update vite.config.js manifest icons array');
  console.log('2. Test PWA installation');
  console.log('3. Run Lighthouse PWA audit');
}

main().catch(error => {
  console.error('❌ Error:', error);
  process.exit(1);
});
