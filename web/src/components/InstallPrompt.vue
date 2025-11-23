<template>
  <v-snackbar
    v-model="showPrompt"
    :timeout="-1"
    location="bottom"
    color="#2c3657"
    elevation="8"
    class="install-prompt"
  >
    <div class="d-flex align-center">
      <v-icon start size="large" color="#00bcd4">mdi-cellphone-arrow-down</v-icon>
      <div>
        <strong style="color: white">Install ActaLog</strong>
        <div class="text-caption" style="color: rgba(255, 255, 255, 0.8)">
          Get quick access and work offline
        </div>
      </div>
    </div>
    <template v-slot:actions>
      <v-btn
        variant="text"
        size="small"
        @click="dismissPrompt"
        style="color: rgba(255, 255, 255, 0.7)"
      >
        Not now
      </v-btn>
      <v-btn
        variant="flat"
        size="small"
        color="#00bcd4"
        @click="installApp"
      >
        Install
      </v-btn>
    </template>
  </v-snackbar>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'

const showPrompt = ref(false)
const deferredPrompt = ref(null)
let promptTimeout = null

// Check if user has already dismissed or installed
const DISMISSED_KEY = 'pwa-install-dismissed'
const INSTALL_PROMPT_DELAY = 60000 // 1 minute

onMounted(() => {
  // Check if already dismissed
  const dismissed = localStorage.getItem(DISMISSED_KEY)
  if (dismissed) {
    return
  }

  // Check if already installed
  if (window.matchMedia('(display-mode: standalone)').matches) {
    return
  }

  // Listen for beforeinstallprompt event
  window.addEventListener('beforeinstallprompt', handleBeforeInstall)

  // Listen for app installed event
  window.addEventListener('appinstalled', handleAppInstalled)
})

onBeforeUnmount(() => {
  // Clean up event listeners and timeout
  window.removeEventListener('beforeinstallprompt', handleBeforeInstall)
  window.removeEventListener('appinstalled', handleAppInstalled)
  if (promptTimeout) {
    clearTimeout(promptTimeout)
  }
})

function handleBeforeInstall(e) {
  // Prevent the default prompt
  e.preventDefault()

  // Save the event for later
  deferredPrompt.value = e

  // Show our custom prompt after a delay
  promptTimeout = setTimeout(() => {
    showPrompt.value = true
  }, INSTALL_PROMPT_DELAY)
}

function handleAppInstalled() {
  console.log('PWA was installed')
  showPrompt.value = false
  deferredPrompt.value = null
}

async function installApp() {
  if (!deferredPrompt.value) {
    console.log('No install prompt available')
    return
  }

  // Show the install prompt
  deferredPrompt.value.prompt()

  // Wait for the user's response
  const { outcome } = await deferredPrompt.value.userChoice
  console.log(`User response to install prompt: ${outcome}`)

  // Clear the deferred prompt
  deferredPrompt.value = null
  showPrompt.value = false
}

function dismissPrompt() {
  showPrompt.value = false
  // Remember the dismissal for 7 days
  const dismissedUntil = Date.now() + (7 * 24 * 60 * 60 * 1000)
  localStorage.setItem(DISMISSED_KEY, dismissedUntil.toString())
}
</script>

<style scoped>
.install-prompt {
  margin-bottom: 70px; /* Account for bottom navigation */
}
</style>
