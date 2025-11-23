/**
 * Offline Storage Utility
 *
 * Provides IndexedDB-based offline storage for workout data.
 * Allows the app to function offline and sync when connection is restored.
 */

const DB_NAME = 'ActaLogDB'
const DB_VERSION = 1
const STORES = {
  WORKOUTS: 'workouts',
  MOVEMENTS: 'movements',
  PENDING_SYNC: 'pendingSync'
}

/**
 * Initialize IndexedDB
 */
export function initDB() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION)

    request.onerror = () => reject(request.error)
    request.onsuccess = () => resolve(request.result)

    request.onupgradeneeded = (event) => {
      const db = event.target.result

      // Workouts store
      if (!db.objectStoreNames.contains(STORES.WORKOUTS)) {
        const workoutStore = db.createObjectStore(STORES.WORKOUTS, { keyPath: 'id', autoIncrement: true })
        workoutStore.createIndex('workout_date', 'workout_date', { unique: false })
        workoutStore.createIndex('user_id', 'user_id', { unique: false })
        workoutStore.createIndex('synced', 'synced', { unique: false })
      }

      // Movements store
      if (!db.objectStoreNames.contains(STORES.MOVEMENTS)) {
        const movementStore = db.createObjectStore(STORES.MOVEMENTS, { keyPath: 'id', autoIncrement: true })
        movementStore.createIndex('name', 'name', { unique: false })
        movementStore.createIndex('type', 'type', { unique: false })
      }

      // Pending sync operations store
      if (!db.objectStoreNames.contains(STORES.PENDING_SYNC)) {
        const syncStore = db.createObjectStore(STORES.PENDING_SYNC, { keyPath: 'id', autoIncrement: true })
        syncStore.createIndex('timestamp', 'timestamp', { unique: false })
        syncStore.createIndex('operation', 'operation', { unique: false })
      }
    }
  })
}

/**
 * Save workout to offline storage
 */
export async function saveWorkoutOffline(workout) {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([STORES.WORKOUTS], 'readwrite')
    const store = transaction.objectStore(STORES.WORKOUTS)

    const workoutData = {
      ...workout,
      synced: false,
      offline_created: true,
      created_at: new Date().toISOString()
    }

    const request = store.add(workoutData)

    request.onsuccess = () => {
      // Add to pending sync queue
      addToPendingSync({
        operation: 'CREATE_WORKOUT',
        data: workoutData,
        timestamp: Date.now()
      })
      resolve(request.result)
    }
    request.onerror = () => reject(request.error)
  })
}

/**
 * Get all workouts from offline storage
 */
export async function getWorkoutsOffline(userId) {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([STORES.WORKOUTS], 'readonly')
    const store = transaction.objectStore(STORES.WORKOUTS)
    const index = store.index('user_id')
    const request = index.getAll(userId)

    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

/**
 * Save movements to offline storage (for caching)
 */
export async function saveMovementsOffline(movements) {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([STORES.MOVEMENTS], 'readwrite')
    const store = transaction.objectStore(STORES.MOVEMENTS)

    // Clear existing movements
    store.clear()

    // Add all movements
    movements.forEach(movement => {
      store.add(movement)
    })

    transaction.oncomplete = () => resolve()
    transaction.onerror = () => reject(transaction.error)
  })
}

/**
 * Get all movements from offline storage
 */
export async function getMovementsOffline() {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([STORES.MOVEMENTS], 'readonly')
    const store = transaction.objectStore(STORES.MOVEMENTS)
    const request = store.getAll()

    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

/**
 * Add operation to pending sync queue
 */
export async function addToPendingSync(syncOperation) {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([STORES.PENDING_SYNC], 'readwrite')
    const store = transaction.objectStore(STORES.PENDING_SYNC)
    const request = store.add(syncOperation)

    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

/**
 * Get all pending sync operations
 */
export async function getPendingSync() {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([STORES.PENDING_SYNC], 'readonly')
    const store = transaction.objectStore(STORES.PENDING_SYNC)
    const request = store.getAll()

    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

/**
 * Remove synced operation from pending queue
 */
export async function removePendingSync(id) {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([STORES.PENDING_SYNC], 'readwrite')
    const store = transaction.objectStore(STORES.PENDING_SYNC)
    const request = store.delete(id)

    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

/**
 * Mark workout as synced
 */
export async function markWorkoutSynced(workoutId) {
  const db = await initDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([STORES.WORKOUTS], 'readwrite')
    const store = transaction.objectStore(STORES.WORKOUTS)
    const request = store.get(workoutId)

    request.onsuccess = () => {
      const workout = request.result
      if (workout) {
        workout.synced = true
        const updateRequest = store.put(workout)
        updateRequest.onsuccess = () => resolve()
        updateRequest.onerror = () => reject(updateRequest.error)
      } else {
        resolve()
      }
    }
    request.onerror = () => reject(request.error)
  })
}

/**
 * Sync pending operations with server
 */
export async function syncWithServer(apiClient) {
  const pendingOps = await getPendingSync()

  if (pendingOps.length === 0) {
    console.log('No pending operations to sync')
    return { success: true, synced: 0, failed: 0 }
  }

  console.log(`Syncing ${pendingOps.length} pending operations...`)
  let syncedCount = 0
  let failedCount = 0

  for (const op of pendingOps) {
    try {
      if (op.operation === 'CREATE_WORKOUT') {
        // Send workout to server
        await apiClient.post('/api/workouts', op.data)
        // Mark as synced and remove from queue
        await removePendingSync(op.id)
        syncedCount++
      } else if (op.operation === 'API_REQUEST') {
        // Replay the original API request
        const requestData = op.data
        await apiClient({
          method: requestData.method,
          url: requestData.url,
          data: requestData.data,
          headers: requestData.headers
        })
        // Remove from queue
        await removePendingSync(op.id)
        syncedCount++
      }
      // Add more operations as needed (UPDATE, DELETE, etc.)
    } catch (error) {
      console.error('Sync failed for operation:', op, error)
      failedCount++
      // Keep in queue for retry
    }
  }

  console.log(`Sync completed: ${syncedCount} synced, ${failedCount} failed`)
  return { success: failedCount === 0, synced: syncedCount, failed: failedCount }
}

/**
 * Check if app is online
 */
export function isOnline() {
  return navigator.onLine
}

/**
 * Setup online/offline event listeners
 */
export function setupNetworkListeners(onOnline, onOffline) {
  window.addEventListener('online', onOnline)
  window.addEventListener('offline', onOffline)

  // Return cleanup function
  return () => {
    window.removeEventListener('online', onOnline)
    window.removeEventListener('offline', onOffline)
  }
}
