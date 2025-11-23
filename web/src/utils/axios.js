import axios from 'axios'
import {
  saveWorkoutOffline,
  addToPendingSync,
  syncWithServer,
  getPendingSync
} from '@/utils/offlineStorage'

// Create axios instance with default config
const instance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '', // Use relative URLs to leverage Vite proxy
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Function to check if request can be handled offline
function canHandleOffline(config) {
  // Only handle POST requests to /api/workouts offline
  return config.method === 'post' && config.url.includes('/api/workouts')
}

// Function to save request for later sync
async function saveForOfflineSync(config, error) {
  try {
    const requestData = {
      method: config.method,
      url: config.url,
      data: config.data,
      headers: config.headers,
      timestamp: Date.now()
    }

    await addToPendingSync({
      operation: 'API_REQUEST',
      data: requestData,
      timestamp: Date.now()
    })

    console.log('Request saved for offline sync:', config.url)
    return {
      data: {
        success: true,
        offline: true,
        message: 'Saved offline. Will sync when connection is restored.'
      },
      status: 202, // Accepted
      statusText: 'Accepted (Offline)',
      config
    }
  } catch (saveError) {
    console.error('Failed to save request offline:', saveError)
    throw error // Throw original error if we can't save offline
  }
}

// Request interceptor
instance.interceptors.request.use(
  (config) => {
    // Add token to request if it exists
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Track if we're currently refreshing the token
let isRefreshing = false
let failedQueue = []

const processQueue = (error, token = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve(token)
    }
  })
  failedQueue = []
}

// Response interceptor
instance.interceptors.response.use(
  (response) => {
    return response
  },
  async (error) => {
    const originalRequest = error.config

    // Handle network errors (offline)
    if (!error.response && error.message === 'Network Error') {
      console.log('Network error detected, checking if can handle offline...')

      // Check if this request can be handled offline
      if (canHandleOffline(originalRequest)) {
        console.log('Request can be handled offline, saving for sync...')
        return saveForOfflineSync(originalRequest, error)
      }

      // For GET requests, we might have cached data (handled by service worker)
      // Let the error propagate for the UI to handle
      console.log('Request cannot be handled offline')
    }

    // Handle 401 Unauthorized
    if (error.response?.status === 401 && !originalRequest._retry) {
      // Check if this is the refresh endpoint failing
      if (originalRequest.url === '/api/auth/refresh') {
        // Refresh token is invalid/expired, clear everything and redirect
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        localStorage.removeItem('refreshToken')
        window.location.href = '/login'
        return Promise.reject(error)
      }

      // If we're already refreshing, queue this request
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        }).then(token => {
          originalRequest.headers.Authorization = `Bearer ${token}`
          return instance(originalRequest)
        }).catch(err => {
          return Promise.reject(err)
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      const refreshToken = localStorage.getItem('refreshToken')

      if (!refreshToken) {
        // No refresh token available, redirect to login
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        window.location.href = '/login'
        return Promise.reject(error)
      }

      try {
        // Attempt to refresh the access token
        const response = await instance.post('/api/auth/refresh', {
          refresh_token: refreshToken
        })

        const newToken = response.data.token
        const newUser = response.data.user

        // Update localStorage
        localStorage.setItem('token', newToken)
        localStorage.setItem('user', JSON.stringify(newUser))

        // Update the authorization header
        instance.defaults.headers.common['Authorization'] = `Bearer ${newToken}`
        originalRequest.headers.Authorization = `Bearer ${newToken}`

        // Process queued requests
        processQueue(null, newToken)

        // Retry the original request
        return instance(originalRequest)
      } catch (refreshError) {
        // Refresh failed, clear everything and redirect
        processQueue(refreshError, null)
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        localStorage.removeItem('refreshToken')
        window.location.href = '/login'
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error)
  }
)

// Export sync function for external use
export async function triggerSync() {
  try {
    console.log('Triggering background sync...')
    await syncWithServer(instance)
    console.log('Background sync completed')
    return true
  } catch (error) {
    console.error('Background sync failed:', error)
    return false
  }
}

// Export pending sync count checker
export async function getPendingSyncCount() {
  try {
    const pending = await getPendingSync()
    return pending.length
  } catch (error) {
    console.error('Failed to get pending sync count:', error)
    return 0
  }
}

export default instance
