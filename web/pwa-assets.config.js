import { defineConfig, minimalPreset } from '@vite-pwa/assets-generator/config'

/**
 * PWA Assets Generator Configuration
 *
 * This configuration generates all PWA icons from a single source image.
 * Run with: npx pwa-assets-generator
 *
 * Source: public/logo.svg (432x432 SVG with cyan circle and white AL logo)
 * Output: public/icons/ and public/ directories
 */
export default defineConfig({
  // Use minimal preset as base, then customize
  preset: {
    ...minimalPreset,
    // Override with custom sizes for ActaLog
    transparent: {
      sizes: [72, 96, 128, 144, 152, 192, 384, 512, 1024],
      favicons: [[16, 'favicon-16x16.png'], [32, 'favicon-32x32.png'], [48, 'favicon-48x48.png']]
    },
    maskable: {
      sizes: [512],
      padding: 0.1 // 10% padding for safe area
    },
    apple: {
      sizes: [120, 152, 167, 180],
      padding: 0 // No padding for Apple icons
    }
  },
  images: ['public/logo.svg'],
  // Output configuration
  headLinkOptions: {
    // Generate link tags for HTML
    preset: 'minimal'
  }
})
