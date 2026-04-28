import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

// ── Mocks (hoisted before imports) ───────────────────────────────────────────

vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn().mockResolvedValue({ data: { count: 0 } }),
    post: vi.fn().mockResolvedValue({}),
    defaults: { headers: { common: {} } },
  },
}))

vi.mock('@/utils/offlineStorage', () => ({
  addToPendingSync: vi.fn(),
  syncWithServer: vi.fn(),
  getPendingSync: vi.fn().mockResolvedValue([]),
}))

vi.mock('@/utils/timezone', () => ({
  formatDateInTimezone: vi.fn(() => 'Fri, Apr 3, 2026'),
}))

vi.mock('@/utils/url', () => ({
  getProfileImageUrl: vi.fn((p) => p),
}))

// vue-router — use a mutable plain object so computed properties in App.vue
// read the raw string value (not a ref) and react when we change it per-test.
const mockRoute = { name: 'dashboard', path: '/dashboard' }
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRoute: vi.fn(() => mockRoute),
    useRouter: vi.fn(() => ({ push: vi.fn() })),
    RouterLink: { template: '<a><slot /></a>' },
    RouterView: { template: '<div />' },
  }
})

// Vuetify useTheme
vi.mock('vuetify', async () => {
  const actual = await vi.importActual('vuetify')
  return {
    ...actual,
    useTheme: vi.fn(() => ({ global: { name: { value: 'light' } } })),
  }
})

// Auth store
const mockAuthStore = {
  isAuthenticated: false,
  user: null,
  logoPath: '/logo.svg',
}
vi.mock('@/stores/auth', () => ({ useAuthStore: () => mockAuthStore }))

// Network store
const mockNetworkStore = {
  isOnline: true,
  initNetworkListeners: vi.fn(),
  incrementPendingSync: vi.fn(),
}
vi.mock('@/stores/network', () => ({ useNetworkStore: () => mockNetworkStore }))

// Subscription store
const mockSubscriptionStore = {
  isExpired: false,
  fetchStatus: vi.fn().mockResolvedValue({}),
}
vi.mock('@/stores/subscription', () => ({ useSubscriptionStore: () => mockSubscriptionStore }))

// Settings store
const mockSettingsStore = { timezone: 'UTC' }
vi.mock('@/stores/settings', () => ({ useSettingsStore: () => mockSettingsStore }))

// Child components
vi.mock('@/components/InstallPrompt.vue', () => ({ default: { template: '<div />' } }))
vi.mock('@/components/UpdatePrompt.vue', () => ({ default: { template: '<div />' } }))
vi.mock('@/components/SubscriptionExpiredBanner.vue', () => ({ default: { template: '<div />' } }))
vi.mock('@/components/ThemeSwitcher.vue', () => ({ default: { template: '<div />' } }))

import App from './App.vue'
import axios from '@/utils/axios'

// ── Helpers ───────────────────────────────────────────────────────────────────

const createWrapper = () => {
  const pinia = createPinia()
  setActivePinia(pinia)
  return shallowMount(App, { global: { plugins: [pinia] } })
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe('App.vue', () => {
  let wrapper

  beforeEach(() => {
    vi.useFakeTimers()
    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: { count: 0 } })
    mockAuthStore.isAuthenticated = false
    mockAuthStore.user = null
    mockRoute.name = 'dashboard'
    mockNetworkStore.isOnline = true
  })

  afterEach(async () => {
    if (wrapper) {
      wrapper.unmount()
      wrapper = null
    }
    vi.useRealTimers()
    await flushPromises()
  })

  // =========================================================================
  // showAppBar computed
  // =========================================================================

  describe('showAppBar', () => {
    it('is false when not authenticated', async () => {
      mockAuthStore.isAuthenticated = false
      wrapper = createWrapper()
      await flushPromises()

      expect(wrapper.vm.showAppBar).toBe(false)
    })

    it('is false on the login route even when authenticated', async () => {
      mockAuthStore.isAuthenticated = true
      mockRoute.name = 'login'
      wrapper = createWrapper()
      await flushPromises()

      expect(wrapper.vm.showAppBar).toBe(false)
    })

    it('is false on the register route', async () => {
      mockAuthStore.isAuthenticated = true
      mockRoute.name = 'register'
      wrapper = createWrapper()
      await flushPromises()

      expect(wrapper.vm.showAppBar).toBe(false)
    })

    it('is true when authenticated and on a protected route', async () => {
      mockAuthStore.isAuthenticated = true
      mockRoute.name = 'dashboard'
      wrapper = createWrapper()
      await flushPromises()

      expect(wrapper.vm.showAppBar).toBe(true)
    })
  })

  // =========================================================================
  // showBottomNav computed
  // =========================================================================

  describe('showBottomNav', () => {
    it('mirrors showAppBar — false when not authenticated', async () => {
      mockAuthStore.isAuthenticated = false
      wrapper = createWrapper()
      await flushPromises()

      expect(wrapper.vm.showBottomNav).toBe(false)
    })

    it('mirrors showAppBar — true when authenticated on dashboard', async () => {
      mockAuthStore.isAuthenticated = true
      mockRoute.name = 'dashboard'
      wrapper = createWrapper()
      await flushPromises()

      expect(wrapper.vm.showBottomNav).toBe(true)
    })
  })

  // =========================================================================
  // mainStyle computed
  // =========================================================================

  describe('mainStyle', () => {
    it('has paddingTop 0 when not showing app bar', async () => {
      mockAuthStore.isAuthenticated = false
      wrapper = createWrapper()
      await flushPromises()

      expect(wrapper.vm.mainStyle.paddingTop).toBe('0')
    })

    it('uses CSS variable for paddingTop when showing app bar', async () => {
      mockAuthStore.isAuthenticated = true
      mockRoute.name = 'dashboard'
      wrapper = createWrapper()
      await flushPromises()

      expect(wrapper.vm.mainStyle.paddingTop).toBe('var(--v-layout-top, 56px)')
    })

    it('has paddingBottom 0 when not showing bottom nav', async () => {
      mockAuthStore.isAuthenticated = false
      wrapper = createWrapper()
      await flushPromises()

      expect(wrapper.vm.mainStyle.paddingBottom).toBe('0')
    })

    it('uses CSS variable with safe-area for paddingBottom when showing bottom nav', async () => {
      mockAuthStore.isAuthenticated = true
      mockRoute.name = 'dashboard'
      wrapper = createWrapper()
      await flushPromises()

      expect(wrapper.vm.mainStyle.paddingBottom).toContain('var(--v-layout-bottom, 56px)')
      expect(wrapper.vm.mainStyle.paddingBottom).toContain('safe-area-inset-bottom')
    })

    it('always sets minHeight and overflowX', async () => {
      wrapper = createWrapper()
      await flushPromises()

      expect(wrapper.vm.mainStyle.minHeight).toBe('100%')
      expect(wrapper.vm.mainStyle.maxWidth).toBe('100vw')
      expect(wrapper.vm.mainStyle.overflowX).toBe('hidden')
    })
  })

  // =========================================================================
  // Notification badge count
  // =========================================================================

  describe('unreadNotificationCount', () => {
    it('starts at 0', async () => {
      wrapper = createWrapper()
      await flushPromises()

      expect(wrapper.vm.unreadNotificationCount).toBe(0)
    })

    it('fetches count from API when authenticated on mount', async () => {
      mockAuthStore.isAuthenticated = true
      axios.get.mockResolvedValue({ data: { count: 5 } })

      wrapper = createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/notifications/count')
      expect(wrapper.vm.unreadNotificationCount).toBe(5)
    })

    it('does not fetch when not authenticated', async () => {
      mockAuthStore.isAuthenticated = false
      wrapper = createWrapper()
      await flushPromises()

      expect(axios.get).not.toHaveBeenCalledWith('/api/notifications/count')
      expect(wrapper.vm.unreadNotificationCount).toBe(0)
    })
  })

  // =========================================================================
  // provide — refreshNotificationCount
  // =========================================================================

  describe('provide refreshNotificationCount', () => {
    it('re-fetches count when fetchUnreadNotificationCount is called directly', async () => {
      mockAuthStore.isAuthenticated = true
      axios.get
        .mockResolvedValueOnce({ data: { count: 3 } }) // initial fetch on mount
        .mockResolvedValueOnce({ data: { count: 0 } }) // after explicit refresh

      wrapper = createWrapper()
      await flushPromises()
      expect(wrapper.vm.unreadNotificationCount).toBe(3)

      // Call the internal function directly (same one exposed via provide)
      await wrapper.vm.fetchUnreadNotificationCount?.()
      await flushPromises()

      // If fetchUnreadNotificationCount is exposed, count updates;
      // either way the API must have been called twice
      expect(axios.get).toHaveBeenCalledTimes(2)
    })
  })
})
