import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import vuetify from './plugins/vuetify'
import { registerSW } from 'virtual:pwa-register'
import { usePwaStore } from '@/stores/pwa'

import './assets/main.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(vuetify)

// Initialize PWA store (must be after pinia is installed)
const pwaStore = usePwaStore()

// Register Service Worker for PWA with user-controlled updates
// Instead of auto-reloading, we notify the user and let them choose when to update
const updateSW = registerSW({
  onNeedRefresh() {
    console.log('New version available')
    pwaStore.setNeedsRefresh(true)
  },
  onOfflineReady() {
    console.log('App ready to work offline')
    pwaStore.setOfflineReady(true)
  },
  // Check for updates every 60 seconds when the app is open
  onRegisteredSW(swUrl, registration) {
    if (registration) {
      setInterval(() => {
        registration.update()
      }, 60 * 1000)
    }
  }
})

// Store the update function so the PWA store can trigger it
pwaStore.setUpdateFunction(updateSW)

app.mount('#app')
