import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// Mock offlineStorage before importing axios (vi.mock is hoisted)
vi.mock('@/utils/offlineStorage', () => ({
  addToPendingSync: vi.fn().mockResolvedValue(undefined),
  syncWithServer: vi.fn().mockResolvedValue(undefined),
  getPendingSync: vi.fn().mockResolvedValue([]),
}))

// Mock subscription store — the 402 interceptor calls useSubscriptionStore()
// dynamically; without this Pinia throws because no app.use(pinia) exists in tests
vi.mock('@/stores/subscription', () => ({
  useSubscriptionStore: vi.fn(() => ({
    setExpired: vi.fn(),
  })),
}))

import instance, { getPendingSyncCount, triggerSync } from './axios'
import { addToPendingSync, syncWithServer, getPendingSync } from '@/utils/offlineStorage'

// Helper: build a mock axios adapter that injects specific responses in sequence.
// Each call to the adapter consumes the next entry; the last entry repeats.
const makeSequentialAdapter = (...responses) => {
  let index = 0
  return async (config) => {
    const r = responses[Math.min(index++, responses.length - 1)]
    if (r.status >= 400) {
      const err = Object.assign(new Error(`Request failed with status code ${r.status}`), {
        response: { status: r.status, data: r.data ?? {}, config },
        config,
      })
      throw err
    }
    return { data: r.data ?? {}, status: r.status ?? 200, headers: {}, config }
  }
}

describe('axios utility', () => {
  let originalAdapter

  beforeEach(() => {
    originalAdapter = instance.defaults.adapter
    localStorage.clear()
    delete instance.defaults.headers.common['Authorization']
    vi.clearAllMocks()
  })

  afterEach(() => {
    instance.defaults.adapter = originalAdapter
  })

  // =========================================================================
  // REQUEST INTERCEPTOR — JWT injection
  // =========================================================================

  describe('Request interceptor', () => {
    it('adds Authorization header when token is in localStorage', async () => {
      localStorage.setItem('token', 'my-jwt-token')

      let capturedConfig
      instance.defaults.adapter = async (config) => {
        capturedConfig = config
        return { data: {}, status: 200, headers: {}, config }
      }

      await instance.get('/api/test')

      expect(capturedConfig.headers.Authorization).toBe('Bearer my-jwt-token')
    })

    it('does not add Authorization header when no token in localStorage', async () => {
      let capturedConfig
      instance.defaults.adapter = async (config) => {
        capturedConfig = config
        return { data: {}, status: 200, headers: {}, config }
      }

      await instance.get('/api/test')

      expect(capturedConfig.headers.Authorization).toBeUndefined()
    })

    it('uses latest token from localStorage on each request', async () => {
      localStorage.setItem('token', 'token-v1')

      const configs = []
      instance.defaults.adapter = async (config) => {
        configs.push(config)
        return { data: {}, status: 200, headers: {}, config }
      }

      await instance.get('/api/first')
      localStorage.setItem('token', 'token-v2')
      await instance.get('/api/second')

      expect(configs[0].headers.Authorization).toBe('Bearer token-v1')
      expect(configs[1].headers.Authorization).toBe('Bearer token-v2')
    })
  })

  // =========================================================================
  // RESPONSE INTERCEPTOR — 402 subscription expired
  // =========================================================================

  describe('Response interceptor — 402', () => {
    it('dispatches subscription-expired custom event on 402', async () => {
      instance.defaults.adapter = makeSequentialAdapter({ status: 402, data: { message: 'Subscription required' } })

      const dispatchSpy = vi.spyOn(window, 'dispatchEvent')

      await instance.get('/api/protected').catch(() => {})

      const dispatched = dispatchSpy.mock.calls.find(
        ([event]) => event instanceof CustomEvent && event.type === 'subscription-expired'
      )
      expect(dispatched).toBeDefined()
      expect(dispatched[0].detail.message).toBe('Subscription required')

      dispatchSpy.mockRestore()
    })

    it('still rejects the promise on 402 after dispatching event', async () => {
      instance.defaults.adapter = makeSequentialAdapter({ status: 402 })

      await expect(instance.get('/api/protected')).rejects.toMatchObject({
        response: { status: 402 },
      })
    })
  })

  // =========================================================================
  // RESPONSE INTERCEPTOR — network errors / offline
  // =========================================================================

  describe('Response interceptor — offline handling', () => {
    it('saves POST to workout endpoint offline when network error occurs', async () => {
      instance.defaults.adapter = async (config) => {
        const err = Object.assign(new Error('Network Error'), { code: 'ERR_NETWORK', config })
        throw err
      }

      const response = await instance.post('/api/workouts', { name: 'Test Workout' })

      expect(addToPendingSync).toHaveBeenCalledOnce()
      expect(response.status).toBe(202)
      expect(response.data.offline).toBe(true)
    })

    it('dispatches offline-save event when workout saved offline', async () => {
      instance.defaults.adapter = async (config) => {
        const err = Object.assign(new Error('Network Error'), { code: 'ERR_NETWORK', config })
        throw err
      }

      const dispatchSpy = vi.spyOn(window, 'dispatchEvent')

      await instance.post('/api/workouts', { name: 'Test Workout' })

      const dispatched = dispatchSpy.mock.calls.find(
        ([event]) => event instanceof CustomEvent && event.type === 'offline-save'
      )
      expect(dispatched).toBeDefined()

      dispatchSpy.mockRestore()
    })

    it('propagates network errors for non-workout endpoints', async () => {
      instance.defaults.adapter = async (config) => {
        const err = Object.assign(new Error('Network Error'), { code: 'ERR_NETWORK', config })
        throw err
      }

      await expect(instance.get('/api/movements')).rejects.toMatchObject({ message: 'Network Error' })
    })
  })

  // =========================================================================
  // UTILITY EXPORTS
  // =========================================================================

  describe('getPendingSyncCount', () => {
    it('returns length of pending sync queue', async () => {
      getPendingSync.mockResolvedValue([{ id: 1 }, { id: 2 }])

      const count = await getPendingSyncCount()

      expect(count).toBe(2)
    })

    it('returns 0 when queue is empty', async () => {
      getPendingSync.mockResolvedValue([])

      const count = await getPendingSyncCount()

      expect(count).toBe(0)
    })

    it('returns 0 on error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      getPendingSync.mockRejectedValue(new Error('Storage error'))

      const count = await getPendingSyncCount()

      expect(count).toBe(0)
      consoleSpy.mockRestore()
    })
  })

  describe('triggerSync', () => {
    it('calls syncWithServer and returns true on success', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      syncWithServer.mockResolvedValue(undefined)

      const result = await triggerSync()

      expect(syncWithServer).toHaveBeenCalledOnce()
      expect(result).toBe(true)
      consoleSpy.mockRestore()
    })

    it('returns false on sync failure', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      syncWithServer.mockRejectedValue(new Error('Sync failed'))

      const result = await triggerSync()

      expect(result).toBe(false)
      consoleSpy.mockRestore()
    })
  })
})
