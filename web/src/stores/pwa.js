import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

/**
 * PWA Update Store
 *
 * Manages PWA service worker update state for user-controlled updates.
 * Uses vite-plugin-pwa 1.2.0's prompt registerType for better control.
 *
 * Flow:
 * 1. Service worker detects new version -> onNeedRefresh() called
 * 2. Store sets needsRefresh=true, shows UpdatePrompt
 * 3. User clicks "Update Now" -> applyUpdate() called
 * 4. Service worker activates, page reloads with new version
 */
export const usePwaStore = defineStore('pwa', () => {
  // State
  const needsRefresh = ref(false)
  const offlineReady = ref(false)
  const showUpdatePrompt = ref(false)
  const isUpdating = ref(false)
  const updateSW = ref(null) // Store the update function from registerSW
  const registration = ref(null) // Store the SW registration for manual checks
  const lastUpdateCheck = ref(null)

  // Computed
  const hasUpdate = computed(() => needsRefresh.value)
  const canCheckForUpdates = computed(() => registration.value !== null)

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

  function setRegistration(reg) {
    registration.value = reg
  }

  async function checkForUpdates() {
    if (registration.value) {
      try {
        lastUpdateCheck.value = new Date()
        await registration.value.update()
        console.log('PWA update check completed')
      } catch (error) {
        console.warn('PWA update check failed:', error)
      }
    }
  }

  async function applyUpdate() {
    if (updateSW.value) {
      isUpdating.value = true
      console.log('Applying PWA update...')
      try {
        await updateSW.value(true)
        // Page will reload, so isUpdating won't be reset
      } catch (error) {
        console.error('Failed to apply PWA update:', error)
        isUpdating.value = false
      }
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
    isUpdating,
    lastUpdateCheck,

    // Computed
    hasUpdate,
    canCheckForUpdates,

    // Actions
    setNeedsRefresh,
    setOfflineReady,
    setUpdateFunction,
    setRegistration,
    checkForUpdates,
    applyUpdate,
    dismissUpdatePrompt
  }
})
