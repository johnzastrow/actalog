import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { flushPromises } from '@vue/test-utils'

// Mock axios before importing the store
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    defaults: { headers: { common: {} } },
  },
}))

// Mock url utilities (used for profile_image URL resolution)
vi.mock('@/utils/url', () => ({
  getProfileImageUrl: vi.fn((path) => `http://localhost${path}`),
}))

import axios from '@/utils/axios'
import { useAuthStore } from './auth'

const mockUser = { id: 1, name: 'Test User', email: 'test@example.com', role: 'athlete' }
const mockToken = 'eyJhbGciOiJIUzI1NiJ9.test'
const mockRefreshToken = 'refresh-token-value'

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    delete axios.defaults.headers.common['Authorization']
    vi.clearAllMocks()

    // Suppress init() and fetchSchedulingEnabled() side effects that run on store creation
    axios.get.mockResolvedValue({ data: { scheduling_enabled: false, logo_variant: 'logo' } })
    axios.post.mockResolvedValue({})
  })

  afterEach(() => {
    localStorage.clear()
  })

  // =========================================================================
  // isAuthenticated computed
  // =========================================================================

  describe('isAuthenticated', () => {
    it('is false when no token or user is stored', async () => {
      const store = useAuthStore()
      await flushPromises()

      expect(store.isAuthenticated).toBe(false)
    })

    it('is true when both token and user are set', async () => {
      localStorage.setItem('token', mockToken)
      localStorage.setItem('user', JSON.stringify(mockUser))

      // fetchProfile is called by init()
      axios.get.mockImplementation((url) => {
        if (url === '/api/users/profile')
          return Promise.resolve({ data: { user: mockUser, is_coach: false } })
        return Promise.resolve({ data: { scheduling_enabled: false, logo_variant: 'logo' } })
      })

      const store = useAuthStore()
      await flushPromises()

      expect(store.isAuthenticated).toBe(true)
    })
  })

  // =========================================================================
  // isAdmin / isCoachRole
  // =========================================================================

  describe('role computed properties', () => {
    it('isAdmin is true when user role is admin', async () => {
      const store = useAuthStore()
      await flushPromises()

      store.user = { ...mockUser, role: 'admin' }
      expect(store.isAdmin).toBe(true)
    })

    it('isCoachRole is true when user role is coach', async () => {
      const store = useAuthStore()
      await flushPromises()

      store.user = { ...mockUser, role: 'coach' }
      expect(store.isCoachRole).toBe(true)
    })

    it('isCoachOrAdmin is true for both coach and admin', async () => {
      const store = useAuthStore()
      await flushPromises()

      store.user = { ...mockUser, role: 'admin' }
      expect(store.isCoachOrAdmin).toBe(true)

      store.user = { ...mockUser, role: 'coach' }
      expect(store.isCoachOrAdmin).toBe(true)

      store.user = { ...mockUser, role: 'athlete' }
      expect(store.isCoachOrAdmin).toBe(false)
    })
  })

  // =========================================================================
  // login()
  // =========================================================================

  describe('login()', () => {
    it('sets token and user on success, returns true', async () => {
      axios.get.mockResolvedValue({ data: { scheduling_enabled: false, logo_variant: 'logo' } })
      axios.post.mockImplementation((url) => {
        if (url === '/api/auth/login')
          return Promise.resolve({ data: { token: mockToken, user: mockUser } })
        return Promise.resolve({})
      })

      const store = useAuthStore()
      await flushPromises()

      // Reset after store init side effects
      axios.get.mockResolvedValue({ data: { user: mockUser, is_coach: false } })

      const result = await store.login('test@example.com', 'password')

      expect(result).toBe(true)
      expect(store.token).toBe(mockToken)
      expect(store.user).toEqual(mockUser)
    })

    it('stores token and user in localStorage on success', async () => {
      axios.post.mockImplementation((url) => {
        if (url === '/api/auth/login')
          return Promise.resolve({ data: { token: mockToken, user: mockUser } })
        return Promise.resolve({})
      })
      axios.get.mockResolvedValue({ data: { user: mockUser, is_coach: false } })

      const store = useAuthStore()
      await flushPromises()

      await store.login('test@example.com', 'password')

      expect(localStorage.getItem('token')).toBe(mockToken)
      expect(JSON.parse(localStorage.getItem('user'))).toEqual(mockUser)
    })

    it('stores refresh token in localStorage when provided', async () => {
      axios.post.mockImplementation((url) => {
        if (url === '/api/auth/login')
          return Promise.resolve({
            data: { token: mockToken, user: mockUser, refresh_token: mockRefreshToken }
          })
        return Promise.resolve({})
      })
      axios.get.mockResolvedValue({ data: { user: mockUser, is_coach: false } })

      const store = useAuthStore()
      await flushPromises()

      await store.login('test@example.com', 'password', true)

      expect(localStorage.getItem('refreshToken')).toBe(mockRefreshToken)
    })

    it('sets error message and returns false on 401', async () => {
      axios.post.mockImplementation((url) => {
        if (url === '/api/auth/login')
          return Promise.reject({
            response: { status: 401, data: { message: 'Invalid credentials' } }
          })
        return Promise.resolve({})
      })

      const store = useAuthStore()
      await flushPromises()

      const result = await store.login('bad@example.com', 'wrong')

      expect(result).toBe(false)
      expect(store.error).toBe('Invalid credentials')
    })

    it('sets network error message when server is unreachable', async () => {
      axios.post.mockImplementation((url) => {
        if (url === '/api/auth/login')
          return Promise.reject({ request: {}, message: 'Network Error' })
        return Promise.resolve({})
      })

      const store = useAuthStore()
      await flushPromises()

      const result = await store.login('test@example.com', 'password')

      expect(result).toBe(false)
      expect(store.error).toContain('Cannot connect')
    })

    it('sets loading false after login completes', async () => {
      axios.post.mockImplementation((url) => {
        if (url === '/api/auth/login')
          return Promise.resolve({ data: { token: mockToken, user: mockUser } })
        return Promise.resolve({})
      })
      axios.get.mockResolvedValue({ data: { user: mockUser, is_coach: false } })

      const store = useAuthStore()
      await flushPromises()

      const loginPromise = store.login('test@example.com', 'password')
      expect(store.loading).toBe(true)
      await loginPromise
      await flushPromises()

      expect(store.loading).toBe(false)
    })
  })

  // =========================================================================
  // logout()
  // =========================================================================

  describe('logout()', () => {
    it('clears token, user and refreshToken state', async () => {
      const store = useAuthStore()
      await flushPromises()

      store.token = mockToken
      store.user = mockUser
      store.refreshToken = null // no refresh token so we skip revocation

      await store.logout()

      expect(store.token).toBeNull()
      expect(store.user).toBeNull()
    })

    it('removes auth items from localStorage', async () => {
      localStorage.setItem('token', mockToken)
      localStorage.setItem('user', JSON.stringify(mockUser))
      localStorage.setItem('refreshToken', mockRefreshToken)

      const store = useAuthStore()
      await flushPromises()

      store.token = mockToken
      store.refreshToken = null

      await store.logout()

      expect(localStorage.getItem('token')).toBeNull()
      expect(localStorage.getItem('user')).toBeNull()
    })

    it('removes Authorization header from axios defaults', async () => {
      axios.defaults.headers.common['Authorization'] = `Bearer ${mockToken}`

      const store = useAuthStore()
      await flushPromises()

      store.token = mockToken
      store.refreshToken = null

      await store.logout()

      expect(axios.defaults.headers.common['Authorization']).toBeUndefined()
    })

    it('calls revoke endpoint when refresh token exists', async () => {
      const store = useAuthStore()
      await flushPromises()

      store.token = mockToken
      store.refreshToken = mockRefreshToken
      axios.post.mockResolvedValue({})

      await store.logout()

      expect(axios.post).toHaveBeenCalledWith('/api/auth/revoke', {
        refresh_token: mockRefreshToken,
      })
    })
  })

  // =========================================================================
  // refreshAccessToken()
  // =========================================================================

  describe('refreshAccessToken()', () => {
    it('returns false when no refresh token', async () => {
      const store = useAuthStore()
      await flushPromises()

      store.refreshToken = null

      const result = await store.refreshAccessToken()

      expect(result).toBe(false)
    })

    it('updates token and returns true on success', async () => {
      const newToken = 'new-access-token'
      axios.post.mockImplementation((url) => {
        if (url === '/api/auth/refresh')
          return Promise.resolve({ data: { token: newToken, user: mockUser } })
        return Promise.resolve({})
      })
      axios.get.mockResolvedValue({ data: { scheduling_enabled: false, logo_variant: 'logo' } })

      const store = useAuthStore()
      await flushPromises()

      store.refreshToken = mockRefreshToken

      const result = await store.refreshAccessToken()

      expect(result).toBe(true)
      expect(store.token).toBe(newToken)
      expect(localStorage.getItem('token')).toBe(newToken)
    })

    it('calls logout and returns false when refresh fails', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockImplementation((url) => {
        if (url === '/api/auth/refresh')
          return Promise.reject(new Error('Refresh failed'))
        return Promise.resolve({})
      })
      axios.get.mockResolvedValue({ data: { scheduling_enabled: false, logo_variant: 'logo' } })

      const store = useAuthStore()
      await flushPromises()

      store.token = mockToken
      store.refreshToken = mockRefreshToken
      store.user = mockUser

      const result = await store.refreshAccessToken()

      expect(result).toBe(false)
      expect(store.token).toBeNull() // logout was called
      consoleSpy.mockRestore()
    })
  })

  // =========================================================================
  // logoPath computed
  // =========================================================================

  describe('logoPath', () => {
    it('returns /logo.svg for default variant', async () => {
      const store = useAuthStore()
      await flushPromises()

      store.logoVariant = 'logo'
      expect(store.logoPath).toBe('/logo.svg')
    })

    it('returns /beta/logo.svg for betalogo variant', async () => {
      const store = useAuthStore()
      await flushPromises()

      store.logoVariant = 'betalogo'
      expect(store.logoPath).toBe('/beta/logo.svg')
    })
  })
})
