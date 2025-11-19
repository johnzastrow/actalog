import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vuetify from 'vite-plugin-vuetify'
import { VitePWA } from 'vite-plugin-pwa'
import { fileURLToPath, URL } from 'node:url'
import fs from 'fs'

// Optional local HTTPS support: if `web/certs/<host>.pem` and
// `web/certs/<host>-key.pem` exist the Vite config will use them for
// an HTTPS dev/preview server. To opt-in, generate certs using mkcert
// and place them in `web/certs` (the `scripts/start-frontend.sh` helper
// can generate these for you).
const CERT_DIR = fileURLToPath(new URL('./certs', import.meta.url))
const DEFAULT_HOST = process.env.VITE_DEV_HOST || 'subdomain.example.com'
const KEY_PATH = `${CERT_DIR}/${DEFAULT_HOST}-key.pem`
const CERT_PATH = `${CERT_DIR}/${DEFAULT_HOST}.pem`
let httpsOptions = undefined
if (fs.existsSync(KEY_PATH) && fs.existsSync(CERT_PATH)) {
  httpsOptions = {
    key: fs.readFileSync(KEY_PATH),
    cert: fs.readFileSync(CERT_PATH)
  }
}

// https://vitejs.dev/config/
export default defineConfig({
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
        // for your deployment. During local development you may want to
        // replace this with a placeholder (e.g. `https://subdomain.example.com/`)
        // but remember that Service Workers and some PWA features require a
        // secure origin (HTTPS) and exact origin matching. If you test locally
        // using a mapped hosts entry (see README) set these values to the
        // production/staging origin you plan to use.
        scope: 'https://subdomain.example.com/',
        start_url: 'https://subdomain.example.com/',
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
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
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
        navigateFallback: 'index.html'
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
    // Default here is the example hostname to use if you map it in hosts file.
    // If you prefer not to map hosts, start Vite with `--host 0.0.0.0` instead.
    host: 'subdomain.example.com',
    port: 3000,
    https: httpsOptions ? httpsOptions : false,
    hmr: {
      // HMR client will try to connect to this host; set to the same host
      // you use to access the dev server. If running with `--host 0.0.0.0`,
      // consider leaving this unset or overriding via CLI.
      host: 'subdomain.example.com'
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
    // Serve built assets from the production hostname
    base: 'https://subdomain.example.com/',
    outDir: 'dist',
    sourcemap: true,
  }
})
