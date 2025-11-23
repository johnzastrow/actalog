import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { setupNetworkListeners, getPendingSync } from '@/utils/offlineStorage'
import { triggerSync, getPendingSyncCount } from '@/utils/axios'

/**
 * Network Status Store
 *
 * Manages online/offline state and sync status for PWA offline support
 */
export const useNetworkStore = defineStore('network', () => {
  // State
  const isOnline = ref(navigator.onLine)
  const isSyncing = ref(false)
  const lastSyncTime = ref(null)
  const showOfflineNotification = ref(false)
  const showOnlineNotification = ref(false)
  const showSyncNotification = ref(false)
  const pendingSyncCount = ref(0)

  // Computed
  const networkStatus = computed(() => {
    if (!isOnline.value) return 'offline'
    if (isSyncing.value) return 'syncing'
    return 'online'
  })

  const hasPendingSync = computed(() => pendingSyncCount.value > 0)

  // Actions
  async function setOnline() {
    const wasOffline = !isOnline.value
    isOnline.value = true
    showOfflineNotification.value = false

    if (wasOffline) {
      // Update pending count
      await updatePendingCount()

      // Show online notification
      showOnlineNotification.value = true
      setTimeout(() => {
        showOnlineNotification.value = false
      }, 3000)

      // Trigger sync if there are pending operations
      if (hasPendingSync.value) {
        await performSync()
      }
    }
  }

  function setOffline() {
    isOnline.value = false
    showOnlineNotification.value = false
    showSyncNotification.value = false

    // Show offline notification
    showOfflineNotification.value = true
  }

  function startSync() {
    isSyncing.value = true
  }

  function endSync(success = true) {
    isSyncing.value = false

    if (success) {
      lastSyncTime.value = new Date()
      pendingSyncCount.value = 0

      // Show sync success notification
      showSyncNotification.value = true
      setTimeout(() => {
        showSyncNotification.value = false
      }, 3000)
    }
  }

  function incrementPendingSync() {
    pendingSyncCount.value++
  }

  function decrementPendingSync() {
    if (pendingSyncCount.value > 0) {
      pendingSyncCount.value--
    }
  }

  function setPendingCount(count) {
    pendingSyncCount.value = count
  }

  async function performSync() {
    if (!isOnline.value) {
      console.log('Cannot sync while offline')
      return
    }

    if (pendingSyncCount.value === 0) {
      console.log('No pending operations to sync')
      return
    }

    console.log('Performing sync for', pendingSyncCount.value, 'pending operations')
    startSync()

    try {
      const success = await triggerSync()
      if (success) {
        await updatePendingCount()
        endSync(true)
      } else {
        endSync(false)
      }
    } catch (error) {
      console.error('Sync failed:', error)
      endSync(false)
    }
  }

  async function updatePendingCount() {
    try {
      const count = await getPendingSyncCount()
      pendingSyncCount.value = count
      console.log('Updated pending sync count:', count)
    } catch (error) {
      console.error('Failed to update pending count:', error)
    }
  }

  function dismissOfflineNotification() {
    showOfflineNotification.value = false
  }

  function dismissOnlineNotification() {
    showOnlineNotification.value = false
  }

  function dismissSyncNotification() {
    showSyncNotification.value = false
  }

  // Initialize network listeners
  function initNetworkListeners() {
    setupNetworkListeners(
      () => setOnline(),
      () => setOffline()
    )
  }

  return {
    // State
    isOnline,
    isSyncing,
    lastSyncTime,
    showOfflineNotification,
    showOnlineNotification,
    showSyncNotification,
    pendingSyncCount,

    // Computed
    networkStatus,
    hasPendingSync,

    // Actions
    setOnline,
    setOffline,
    startSync,
    endSync,
    incrementPendingSync,
    decrementPendingSync,
    setPendingCount,
    performSync,
    updatePendingCount,
    dismissOfflineNotification,
    dismissOnlineNotification,
    dismissSyncNotification,
    initNetworkListeners
  }
})
