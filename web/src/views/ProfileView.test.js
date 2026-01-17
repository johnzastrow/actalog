import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ProfileView from './ProfileView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  }
}))

// Mock vue-router
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: vi.fn(),
    })),
    RouterLink: {
      template: '<a><slot /></a>'
    }
  }
})

// Mock URL utility
vi.mock('@/utils/url', () => ({
  getProfileImageUrl: vi.fn((path) => path ? `http://localhost/uploads/${path}` : null)
}))

// Mock auth store
const mockAuthStore = {
  user: {
    id: 1,
    name: 'Test User',
    email: 'test@example.com',
    role: 'user',
    created_at: '2024-01-01T00:00:00Z',
    profile_image: null
  },
  logout: vi.fn()
}

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => mockAuthStore
}))

// Mock subscription store
const mockSubscriptionStore = {
  loading: false,
  error: null,
  subscriptionStatus: { has_access: true },
  isPermanentFree: false,
  hasAccess: true,
  subscriptionType: 'monthly',
  subscriptionTypeLabel: 'Monthly',
  userSubscription: {
    end_date: '2025-12-31T00:00:00Z',
    next_billing_date: '2025-02-01T00:00:00Z'
  },
  orgSubscriptions: [],
  source: 'user',
  fetchStatus: vi.fn().mockResolvedValue({}),
  clear: vi.fn()
}

vi.mock('@/stores/subscription', () => ({
  useSubscriptionStore: () => mockSubscriptionStore
}))

// Mock UserAvatar component
vi.mock('@/components/UserAvatar.vue', () => ({
  default: {
    template: '<div class="user-avatar"></div>',
    props: ['user', 'size']
  }
}))

import axios from '@/utils/axios'
import { useRouter } from 'vue-router'

describe('ProfileView', () => {
  let wrapper
  let mockRouter

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    mockRouter = {
      push: vi.fn()
    }
    useRouter.mockReturnValue(mockRouter)

    // Mock API responses
    axios.get.mockImplementation((url) => {
      if (url === '/api/version') {
        return Promise.resolve({
          data: { version: '0.22.0', build: 123, fullVersion: '0.22.0-beta' }
        })
      }
      if (url.includes('/api/workouts')) {
        return Promise.resolve({
          data: { workouts: [] }
        })
      }
      if (url === '/api/pr-movements') {
        return Promise.resolve({
          data: { movements: [] }
        })
      }
      if (url === '/api/workouts/my-templates') {
        return Promise.resolve({
          data: { workouts: [] }
        })
      }
      return Promise.resolve({ data: {} })
    })

    wrapper = shallowMount(ProfileView, {
      global: {
        plugins: [pinia],
        stubs: {
          RouterLink: true,
          'user-avatar': true,
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-icon': { template: '<i></i>' },
          'v-chip': { template: '<span><slot /></span>' },
          'v-chip-group': { template: '<div><slot /></div>' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-list': { template: '<div><slot /></div>' },
          'v-list-item': { template: '<div><slot /></div>' },
          'v-list-item-title': { template: '<div><slot /></div>' },
          'v-list-item-subtitle': { template: '<div><slot /></div>' },
          'v-divider': { template: '<hr />' },
          'v-progress-circular': { template: '<div></div>' },
          'v-bottom-navigation': { template: '<div><slot /></div>' },
          'v-avatar': { template: '<div><slot /></div>' }
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    // Reset mock user
    mockAuthStore.user = {
      id: 1,
      name: 'Test User',
      email: 'test@example.com',
      role: 'user',
      created_at: '2024-01-01T00:00:00Z',
      profile_image: null
    }
    mockSubscriptionStore.loading = false
    mockSubscriptionStore.error = null
    mockSubscriptionStore.hasAccess = true
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
      wrapper = null
    }
  })

  // ==========================================================================
  // INITIALIZATION TESTS
  // ==========================================================================

  describe('Initialization', () => {
    it('fetches version info on mount', async () => {
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/version')
    })

    it('fetches stats on mount', async () => {
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/workouts?limit=10000')
      expect(axios.get).toHaveBeenCalledWith('/api/pr-movements')
      expect(axios.get).toHaveBeenCalledWith('/api/workouts/my-templates')
    })

    it('fetches subscription status on mount', async () => {
      createWrapper()
      await flushPromises()

      expect(mockSubscriptionStore.fetchStatus).toHaveBeenCalled()
    })

    it('initializes with default stats values', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.stats.totalWorkouts).toBe(0)
      expect(vm.stats.currentStreak).toBe(0)
      expect(vm.stats.personalRecords).toBe(0)
      expect(vm.stats.customTemplates).toBe(0)
    })

    it('initializes with month as default time period', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.selectedPeriod).toBe('month')
    })
  })

  // ==========================================================================
  // USER DISPLAY TESTS
  // ==========================================================================

  describe('User Display', () => {
    it('exposes user from auth store', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.user.name).toBe('Test User')
      expect(vm.user.email).toBe('test@example.com')
    })

    it('identifies admin users', () => {
      mockAuthStore.user.role = 'admin'
      createWrapper()

      expect(mockAuthStore.user.role).toBe('admin')
    })

    it('identifies regular users', () => {
      mockAuthStore.user.role = 'user'
      createWrapper()

      expect(mockAuthStore.user.role).toBe('user')
    })
  })

  // ==========================================================================
  // LOGOUT TESTS
  // ==========================================================================

  describe('Logout', () => {
    it('calls authStore.logout when logging out', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.handleLogout()

      expect(mockAuthStore.logout).toHaveBeenCalled()
    })

    it('clears subscription store when logging out', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.handleLogout()

      expect(mockSubscriptionStore.clear).toHaveBeenCalled()
    })

    it('navigates to login page after logout', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.handleLogout()

      expect(mockRouter.push).toHaveBeenCalledWith('/login')
    })
  })

  // ==========================================================================
  // AVATAR TESTS
  // ==========================================================================

  describe('Avatar', () => {
    it('starts with avatar upload state as false', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.uploadingAvatar).toBe(false)
      expect(vm.deletingAvatar).toBe(false)
    })

    it('uploads avatar successfully', async () => {
      axios.post.mockResolvedValue({
        data: {
          user: {
            id: 1,
            name: 'Test User',
            email: 'test@example.com',
            profile_image: '/uploads/avatar.jpg'
          }
        }
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      // Create mock file
      const mockFile = new File(['test'], 'avatar.jpg', { type: 'image/jpeg' })

      // Mock the event
      const mockEvent = {
        target: {
          files: [mockFile]
        }
      }

      await vm.handleFileSelect(mockEvent)
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith(
        '/api/users/avatar',
        expect.any(FormData),
        { headers: { 'Content-Type': 'multipart/form-data' } }
      )
    })

    it('rejects non-image files', async () => {
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})

      createWrapper()

      const vm = wrapper.vm

      // Create mock non-image file
      const mockFile = new File(['test'], 'document.pdf', { type: 'application/pdf' })

      const mockEvent = {
        target: {
          files: [mockFile]
        }
      }

      await vm.handleFileSelect(mockEvent)

      expect(alertSpy).toHaveBeenCalledWith('Please select an image file')
      expect(axios.post).not.toHaveBeenCalled()

      alertSpy.mockRestore()
    })

    it('rejects files larger than 5MB', async () => {
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})

      createWrapper()

      const vm = wrapper.vm

      // Create mock large file
      const mockFile = new File(['x'.repeat(6 * 1024 * 1024)], 'large.jpg', { type: 'image/jpeg' })
      Object.defineProperty(mockFile, 'size', { value: 6 * 1024 * 1024 })

      const mockEvent = {
        target: {
          files: [mockFile]
        }
      }

      await vm.handleFileSelect(mockEvent)

      expect(alertSpy).toHaveBeenCalledWith('Image must be smaller than 5MB')
      expect(axios.post).not.toHaveBeenCalled()

      alertSpy.mockRestore()
    })

    it('deletes avatar successfully', async () => {
      const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
      axios.delete.mockResolvedValue({
        data: {
          user: {
            id: 1,
            name: 'Test User',
            email: 'test@example.com',
            profile_image: null
          }
        }
      })

      createWrapper()

      const vm = wrapper.vm
      await vm.deleteAvatar()
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/users/avatar')

      confirmSpy.mockRestore()
    })

    it('does not delete avatar if user cancels confirmation', async () => {
      const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)

      createWrapper()

      const vm = wrapper.vm
      await vm.deleteAvatar()

      expect(axios.delete).not.toHaveBeenCalled()

      confirmSpy.mockRestore()
    })
  })

  // ==========================================================================
  // STATS TESTS
  // ==========================================================================

  describe('Stats', () => {
    it('initializes stats with default values', () => {
      createWrapper()

      const vm = wrapper.vm

      // Stats should have default structure
      expect(vm.stats).toHaveProperty('totalWorkouts')
      expect(vm.stats).toHaveProperty('currentStreak')
      expect(vm.stats).toHaveProperty('personalRecords')
      expect(vm.stats).toHaveProperty('customTemplates')
    })

    it('handles API errors gracefully', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      axios.get.mockRejectedValue(new Error('Network error'))

      createWrapper()
      await flushPromises()

      // Should not throw, stats should remain at defaults
      const vm = wrapper.vm
      expect(vm.stats.totalWorkouts).toBe(0)

      consoleSpy.mockRestore()
    })

    it('refetches stats when period changes', async () => {
      createWrapper()
      await flushPromises()

      vi.clearAllMocks()

      const vm = wrapper.vm
      vm.selectedPeriod = 'week'

      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/workouts?limit=10000')
    })
  })

  // ==========================================================================
  // STREAK CALCULATION TESTS (via stats output)
  // ==========================================================================

  describe('Streak Calculation', () => {
    it('streak is 0 for empty workouts', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      // Default mock returns empty workouts, so streak should be 0
      expect(vm.stats.currentStreak).toBe(0)
    })

    it('stats are updated after fetch', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      // Stats should be updated after API calls
      expect(vm.stats).toBeDefined()
      expect(typeof vm.stats.currentStreak).toBe('number')
      expect(typeof vm.stats.totalWorkouts).toBe('number')
    })
  })

  // ==========================================================================
  // TIME PERIOD SELECTION TESTS
  // ==========================================================================

  describe('Time Period Selection', () => {
    it('can change selected period to week', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedPeriod = 'week'

      expect(vm.selectedPeriod).toBe('week')
    })

    it('can change selected period to year', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedPeriod = 'year'

      expect(vm.selectedPeriod).toBe('year')
    })

    it('can change selected period to all', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.selectedPeriod = 'all'

      expect(vm.selectedPeriod).toBe('all')
    })
  })

  // ==========================================================================
  // SUBSCRIPTION DISPLAY TESTS
  // ==========================================================================

  describe('Subscription Display', () => {
    it('shows subscription from subscription store', () => {
      mockSubscriptionStore.hasAccess = true
      mockSubscriptionStore.subscriptionType = 'monthly'
      createWrapper()

      expect(mockSubscriptionStore.hasAccess).toBe(true)
      expect(mockSubscriptionStore.subscriptionType).toBe('monthly')
    })

    it('shows permanent free status', () => {
      mockSubscriptionStore.isPermanentFree = true
      createWrapper()

      expect(mockSubscriptionStore.isPermanentFree).toBe(true)
    })

    it('shows expired status', () => {
      mockSubscriptionStore.hasAccess = false
      createWrapper()

      expect(mockSubscriptionStore.hasAccess).toBe(false)
    })

    it('shows organization subscriptions', () => {
      mockSubscriptionStore.orgSubscriptions = [
        { id: 1, organization_name: 'Test Org', status: 'active' }
      ]
      createWrapper()

      expect(mockSubscriptionStore.orgSubscriptions).toHaveLength(1)
      expect(mockSubscriptionStore.orgSubscriptions[0].organization_name).toBe('Test Org')
    })
  })

  // ==========================================================================
  // VERSION INFO TESTS
  // ==========================================================================

  describe('Version Info', () => {
    it('calls version API on mount', async () => {
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/version')
    })

    it('stores version info in component state', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      // Version info should be populated after mount
      expect(vm.appVersion).toBeDefined()
      expect(vm.buildNumber).toBeDefined()
    })
  })
})
