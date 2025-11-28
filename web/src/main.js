import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import vuetify from './plugins/vuetify'
import { registerSW } from 'virtual:pwa-register'

import './assets/main.css'

// Register Service Worker for PWA with silent auto-update
// When a new version is detected, immediately reload the page
const updateSW = registerSW({
  onNeedRefresh() {
    // Silent auto-reload: immediately activate new service worker and reload
    console.log('New version available, reloading...')
    updateSW(true)
  },
  onOfflineReady() {
    console.log('App ready to work offline')
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

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(vuetify)

app.mount('#app')
