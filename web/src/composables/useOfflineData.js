import { ref } from 'vue'
import axios from '@/utils/axios'
import {
  saveMovementsOffline,
  getMovementsOffline
} from '@/utils/offlineStorage'
import { useNetworkStore } from '@/stores/network'

/**
 * Composable for handling offline-capable data fetching
 *
 * Provides network-aware data loading with automatic caching and fallback
 */
export function useOfflineData() {
  const networkStore = useNetworkStore()

  /**
   * Fetch movements with offline support
   * - When online: Fetches from API and caches to IndexedDB
   * - When offline: Returns cached data from IndexedDB
   */
  async function fetchMovements() {
    try {
      if (networkStore.isOnline) {
        // Online: Fetch from API
        const response = await axios.get('/api/movements')
        const movements = response.data

        // Cache for offline use
        await saveMovementsOffline(movements)

        return { data: movements, source: 'api' }
      } else {
        // Offline: Return cached data
        const cachedMovements = await getMovementsOffline()

        if (cachedMovements.length === 0) {
          throw new Error('No cached movements available offline')
        }

        return { data: cachedMovements, source: 'cache' }
      }
    } catch (error) {
      // If API fails but we're online, try cache as fallback
      if (networkStore.isOnline && error.message !== 'No cached movements available offline') {
        console.log('API failed, falling back to cached data')
        const cachedMovements = await getMovementsOffline()

        if (cachedMovements.length > 0) {
          return { data: cachedMovements, source: 'cache-fallback' }
        }
      }

      throw error
    }
  }

  /**
   * Generic offline-capable fetch with caching
   */
  async function fetchWithCache({
    url,
    cacheKey,
    saveToCache,
    getFromCache
  }) {
    try {
      if (networkStore.isOnline) {
        // Online: Fetch from API
        const response = await axios.get(url)
        const data = response.data

        // Cache if functions provided
        if (saveToCache) {
          await saveToCache(data)
        }

        return { data, source: 'api' }
      } else {
        // Offline: Return cached data
        if (!getFromCache) {
          throw new Error('No cache handler provided for offline mode')
        }

        const cachedData = await getFromCache()

        if (!cachedData || (Array.isArray(cachedData) && cachedData.length === 0)) {
          throw new Error(`No cached data available for ${cacheKey}`)
        }

        return { data: cachedData, source: 'cache' }
      }
    } catch (error) {
      // Fallback to cache if API fails
      if (networkStore.isOnline && getFromCache) {
        console.log(`API failed for ${url}, falling back to cache`)
        const cachedData = await getFromCache()

        if (cachedData && (!Array.isArray(cachedData) || cachedData.length > 0)) {
          return { data: cachedData, source: 'cache-fallback' }
        }
      }

      throw error
    }
  }

  return {
    fetchMovements,
    fetchWithCache
  }
}
