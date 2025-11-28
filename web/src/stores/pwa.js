import { defineStore } from 'pinia'
import { ref } from 'vue'

/**
 * PWA Update Store
 *
 * Manages PWA service worker update state for user-controlled updates
 * instead of disruptive auto-reloads
 */
export const usePwaStore = defineStore('pwa', () => {
  // State
  const needsRefresh = ref(false)
  const offlineReady = ref(false)
  const showUpdatePrompt = ref(false)
  const updateSW = ref(null) // Store the update function from registerSW

  // Actions
  function setNeedsRefresh(value) {
    needsRefresh.value = value
    if (value) {
      showUpdatePrompt.value = true
    }
  }

  function setOfflineReady(value) {
    offlineReady.value = value
  }

  function setUpdateFunction(fn) {
    updateSW.value = fn
  }

  async function applyUpdate() {
    if (updateSW.value) {
      console.log('Applying PWA update...')
      await updateSW.value(true)
    }
  }

  function dismissUpdatePrompt() {
    showUpdatePrompt.value = false
    // Don't clear needsRefresh - user can still update via settings later
  }

  return {
    // State
    needsRefresh,
    offlineReady,
    showUpdatePrompt,

    // Actions
    setNeedsRefresh,
    setOfflineReady,
    setUpdateFunction,
    applyUpdate,
    dismissUpdatePrompt
  }
})
