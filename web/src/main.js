import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import vuetify from './plugins/vuetify'
import { registerSW } from 'virtual:pwa-register'
import { usePwaStore } from '@/stores/pwa'

// MDI icons - must be imported for icon font to work
import '@mdi/font/css/materialdesignicons.css'
import './assets/main.css'
import './assets/fonts.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(vuetify)

// Initialize PWA store (must be after pinia is installed)
const pwaStore = usePwaStore()

// Register Service Worker for PWA with user-controlled updates (prompt mode)
// vite-plugin-pwa 1.2.0: Uses 'prompt' registerType for better update control
// Flow: detect update -> notify user -> user clicks "Update Now" -> reload
const updateSW = registerSW({
  onNeedRefresh() {
    console.log('New version available - prompting user')
    pwaStore.setNeedsRefresh(true)
  },
  onOfflineReady() {
    console.log('App ready to work offline')
    pwaStore.setOfflineReady(true)
  },
  // Store registration for manual update checks and periodic auto-checks
  onRegisteredSW(swUrl, registration) {
    if (registration) {
      // Store registration for manual "Check for Updates" feature
      pwaStore.setRegistration(registration)

      // Check for updates every 60 seconds when the app is open
      // This ensures users are promptly notified of new versions
      setInterval(() => {
        registration.update()
      }, 60 * 1000)

      console.log('Service Worker registered:', swUrl)
    }
  },
  onRegisterError(error) {
    console.error('Service Worker registration error:', error)
  }
})

// Store the update function so the PWA store can trigger updates
pwaStore.setUpdateFunction(updateSW)

app.mount('#app')
