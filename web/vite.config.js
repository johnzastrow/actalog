import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vuetify from 'vite-plugin-vuetify'
import { VitePWA } from 'vite-plugin-pwa'
import { fileURLToPath, URL } from 'node:url'
import fs from 'fs'

// https://vitejs.dev/config/

// Optional local HTTPS support: if `web/certs/<host>.pem` and
// `web/certs/<host>-key.pem` exist the Vite config will use them for
// an HTTPS dev/preview server. To opt-in, generate certs using mkcert
// and place them in `web/certs` (the `scripts/start-frontend.sh` helper
// can generate these for you).
const CERT_DIR = fileURLToPath(new URL('./certs', import.meta.url))

// Read configuration from environment variables (set by start-frontend.sh)
const DEFAULT_HOST = process.env.VITE_DEV_HOST || 'localhost'
const DEFAULT_PORT = process.env.VITE_DEV_PORT || '3000'
const USE_HTTPS = (process.env.VITE_USE_HTTPS || 'false') === 'true'
const DEPLOYMENT_URL = process.env.VITE_DEPLOYMENT_URL || `http://localhost:${DEFAULT_PORT}`

// Ensure DEPLOYMENT_URL has protocol
const FULL_DEPLOYMENT_URL = DEPLOYMENT_URL.match(/^https?:\/\//)
  ? DEPLOYMENT_URL
  : (USE_HTTPS ? `https://${DEPLOYMENT_URL}` : `http://${DEPLOYMENT_URL}`)

const KEY_PATH = `${CERT_DIR}/${DEFAULT_HOST}-key.pem`
const CERT_PATH = `${CERT_DIR}/${DEFAULT_HOST}.pem`
let httpsOptions = undefined
if (USE_HTTPS && fs.existsSync(KEY_PATH) && fs.existsSync(CERT_PATH)) {
  httpsOptions = {
    key: fs.readFileSync(KEY_PATH),
    cert: fs.readFileSync(CERT_PATH)
  }
}

// https://vitejs.dev/config/
export default defineConfig({
  optimizeDeps: {
    include: [
      'vuetify',
      'vuetify/components',
      'vuetify/directives',
      'workbox-window'
    ],
    exclude: []
  },
  plugins: [
    vue(),
    vuetify({ autoImport: true }),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'robots.txt', 'apple-touch-icon.png'],
      manifest: {
        name: 'ActaLog - CrossFit Workout Tracker',
        short_name: 'ActaLog',
        description: 'Track your CrossFit workouts, log weights and reps, and monitor your fitness progress',
        theme_color: '#00bcd4',
        background_color: '#ffffff',
        display: 'standalone',
        orientation: 'portrait',
        // NOTE: set `scope` and `start_url` to the canonical origin
        // for your deployment. These are now read from VITE_DEPLOYMENT_URL
        // environment variable (set by scripts/start-frontend.sh)
        scope: `${FULL_DEPLOYMENT_URL}/`,
        start_url: `${FULL_DEPLOYMENT_URL}/`,
        icons: [
          {
            src: '/icons/icon-72x72.png',
            sizes: '72x72',
            type: 'image/png',
            purpose: 'any'
          },
          {
            src: '/icons/icon-96x96.png',
            sizes: '96x96',
            type: 'image/png',
            purpose: 'any'
          },
          {
            src: '/icons/icon-128x128.png',
            sizes: '128x128',
            type: 'image/png',
            purpose: 'any'
          },
          {
            src: '/icons/icon-144x144.png',
            sizes: '144x144',
            type: 'image/png',
            purpose: 'any'
          },
          {
            src: '/icons/icon-152x152.png',
            sizes: '152x152',
            type: 'image/png',
            purpose: 'any'
          },
          {
            src: '/icons/icon-192x192.png',
            sizes: '192x192',
            type: 'image/png',
            purpose: 'any'
          },
          {
            src: '/icons/icon-384x384.png',
            sizes: '384x384',
            type: 'image/png',
            purpose: 'any'
          },
          {
            src: '/icons/icon-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any'
          },
          {
            src: '/icons/icon-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'maskable'
          }
        ]
      },
      workbox: {
        globDirectory: 'dist',
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
        navigateFallback: 'index.html',
        navigateFallbackDenylist: [/^\/api/],
        cleanupOutdatedCaches: true,
        skipWaiting: true,
        clientsClaim: true,
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/fonts\.googleapis\.com\/.*/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'google-fonts-cache',
              expiration: {
                maxEntries: 10,
                maxAgeSeconds: 60 * 60 * 24 * 365 // 1 year
              },
              cacheableResponse: {
                statuses: [0, 200]
              }
            }
          },
          {
            urlPattern: /^https:\/\/fonts\.gstatic\.com\/.*/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'gstatic-fonts-cache',
              expiration: {
                maxEntries: 10,
                maxAgeSeconds: 60 * 60 * 24 * 365 // 1 year
              },
              cacheableResponse: {
                statuses: [0, 200]
              }
            }
          },
          {
            urlPattern: /^https:\/\/cdn\.jsdelivr\.net\/.*/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'cdn-cache',
              expiration: {
                maxEntries: 10,
                maxAgeSeconds: 60 * 60 * 24 * 365 // 1 year
              },
              cacheableResponse: {
                statuses: [0, 200]
              }
            }
          },
          {
            urlPattern: /\/api\/.*\/*.json/,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-cache',
              expiration: {
                maxEntries: 50,
                maxAgeSeconds: 60 * 5 // 5 minutes
              },
              cacheableResponse: {
                statuses: [0, 200]
              },
              networkTimeoutSeconds: 10
            }
          }
        ]
      },
      devOptions: {
        enabled: true,
        type: 'module',
        navigateFallback: 'index.html',
        suppressWarnings: true
      }
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  // SERVER CONFIGURATION
  // --------------------
  // The `server` block controls Vite's dev server behavior. Typical workflows:
  // - Local development (no hosts file): bind to 0.0.0.0 or localhost and use
  //   `http://localhost:3000`.
  // - Local development with a mapped hostname (recommended for PWA/cookie
  //   testing): map a name (example: subdomain.example.com) to your loopback
  //   interface in your OS hosts file, then set `host` to that name so HMR and
  //   Service Worker expectations match the origin.
  // - Named server (staging/production): do NOT use the dev server. Build and
  //   deploy the `dist/` output to your web host, ensuring the `build.base` and
  //   PWA manifest values match the deployed origin.
  //
  // To run the dev server without editing your hosts file, start Vite with
  // `--host 0.0.0.0` (or export HOST=0.0.0.0). Example:
  //   cd web && npm run dev -- --host 0.0.0.0
  // This will bind Vite to all interfaces and allow testing via
  // `http://localhost:3000` or `http://<your-ip>:3000`.
  server: {
    // Host and port are read from environment variables (set by start-frontend.sh)
    // or can be overridden via CLI: npm run dev -- --host <host> --port <port>
    host: DEFAULT_HOST,
    port: parseInt(DEFAULT_PORT),
    https: httpsOptions ? httpsOptions : false,
    hmr: {
      // HMR client will try to connect to this host; set to the same host
      // you use to access the dev server. CLI args override this.
      host: DEFAULT_HOST
    },
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  },
  build: {
    // Serve built assets from the deployment URL (read from environment)
    base: `${FULL_DEPLOYMENT_URL}/`,
    outDir: 'dist',
    sourcemap: true,
  }
})
