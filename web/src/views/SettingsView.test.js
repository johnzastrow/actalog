import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import SettingsView from './SettingsView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
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

// Mock timezone utilities
vi.mock('@/utils/timezone', () => ({
  getTimezoneOptions: vi.fn(() => [
    { label: 'America/New_York (UTC-05:00)', value: 'America/New_York' },
    { label: 'America/Los_Angeles (UTC-08:00)', value: 'America/Los_Angeles' },
    { label: 'Europe/London (UTC+00:00)', value: 'Europe/London' },
  ]),
  getBrowserTimezone: vi.fn(() => 'America/New_York')
}))

// Mock auth store
const mockAuthStore = {
  user: {
    id: 1,
    name: 'Test User',
    email: 'test@example.com',
    birthday: '1990-05-15T00:00:00Z',
    role: 'user'
  },
  logout: vi.fn()
}

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => mockAuthStore
}))

// Mock subscription store
const mockSubscriptionStore = {
  fetchStatus: vi.fn().mockResolvedValue({}),
  clear: vi.fn()
}

vi.mock('@/stores/subscription', () => ({
  useSubscriptionStore: () => mockSubscriptionStore
}))

// Mock theme store
const mockThemeStore = {
  currentTheme: 'light',
  availableThemes: [
    { id: 'light', name: 'Light', icon: 'mdi-white-balance-sunny' },
    { id: 'dark', name: 'Dark', icon: 'mdi-weather-night' }
  ],
  useSystemPreference: false,
  setTheme: vi.fn(),
  setUseSystemPreference: vi.fn()
}

vi.mock('@/stores/theme', () => ({
  useThemeStore: () => mockThemeStore
}))

// Mock font store
const mockFontStore = {
  currentFont: 'inter',
  availableFonts: [
    { id: 'inter', name: 'Inter', icon: 'mdi-format-font' },
    { id: 'roboto', name: 'Roboto', icon: 'mdi-format-font' }
  ],
  accessibilityFonts: []
}

vi.mock('@/stores/font', () => ({
  useFontStore: () => mockFontStore
}))

// Mock settings store
const mockSettingsStore = {
  timezone: 'America/New_York',
  weightUnit: 'lbs',
  adminUserEventNotifications: true,
  fetchSettings: vi.fn().mockResolvedValue({}),
  updateTimezone: vi.fn().mockResolvedValue({}),
  updateWeightUnit: vi.fn().mockResolvedValue({}),
  updateFontFamily: vi.fn().mockResolvedValue({}),
  updateAdminUserEventNotifications: vi.fn().mockResolvedValue({})
}

vi.mock('@/stores/settings', () => ({
  useSettingsStore: () => mockSettingsStore
}))

// Mock SubscriptionStatusBadge component
vi.mock('@/components/SubscriptionStatusBadge.vue', () => ({
  default: {
    template: '<div class="subscription-badge">Subscription Status</div>'
  }
}))

import axios from '@/utils/axios'
import { useRouter } from 'vue-router'

describe('SettingsView', () => {
  let wrapper
  let mockRouter

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    mockRouter = {
      push: vi.fn()
    }
    useRouter.mockReturnValue(mockRouter)

    wrapper = shallowMount(SettingsView, {
      global: {
        plugins: [pinia],
        stubs: {
          RouterLink: true,
          'subscription-status-badge': true,
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-card-title': { template: '<div><slot /></div>' },
          'v-card-text': { template: '<div><slot /></div>' },
          'v-card-actions': { template: '<div><slot /></div>' },
          'v-form': { template: '<form><slot /></form>' },
          'v-text-field': { template: '<input />' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-select': { template: '<select></select>' },
          'v-switch': { template: '<input type="checkbox" />' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-dialog': { template: '<div><slot /></div>' },
          'v-snackbar': { template: '<div><slot /></div>' },
          'v-bottom-navigation': { template: '<div><slot /></div>' },
          'v-icon': { template: '<i></i>' },
          'v-avatar': { template: '<div><slot /></div>' },
          'v-divider': { template: '<hr />' },
          'v-spacer': { template: '<div></div>' },
          'v-file-input': { template: '<input type="file" />' },
          'v-expansion-panels': { template: '<div><slot /></div>' },
          'v-expansion-panel': { template: '<div><slot /></div>' },
          'v-expansion-panel-title': { template: '<div><slot /></div>' },
          'v-expansion-panel-text': { template: '<div><slot /></div>' },
          'v-list': { template: '<div><slot /></div>' },
          'v-list-item': { template: '<div><slot /></div>' },
          'v-chip': { template: '<div><slot /></div>' },
          'v-chip-group': { template: '<div><slot /></div>' }
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
      birthday: '1990-05-15T00:00:00Z',
      role: 'user'
    }
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
    it('loads user data into form on mount', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.profileForm.name).toBe('Test User')
      expect(vm.profileForm.email).toBe('test@example.com')
    })

    it('fetches settings on mount', async () => {
      createWrapper()
      await flushPromises()

      expect(mockSettingsStore.fetchSettings).toHaveBeenCalled()
    })

    it('fetches subscription status on mount', async () => {
      createWrapper()
      await flushPromises()

      expect(mockSubscriptionStore.fetchStatus).toHaveBeenCalled()
    })

    it('initializes with correct default state', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)
      expect(vm.passwordLoading).toBe(false)
      expect(vm.deleteDialog).toBe(false)
      expect(vm.importDialog).toBe(false)
    })
  })

  // ==========================================================================
  // PROFILE UPDATE TESTS
  // ==========================================================================

  describe('Profile Update', () => {
    it('validates name is required', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.profileForm.name = ''
      vm.profileForm.email = 'test@example.com'

      await vm.updateProfile()
      await flushPromises()

      expect(vm.errors.name).toBe('Name is required')
      expect(axios.put).not.toHaveBeenCalled()
    })

    it('validates email is required', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.profileForm.name = 'Test User'
      vm.profileForm.email = ''

      await vm.updateProfile()
      await flushPromises()

      expect(vm.errors.email).toBe('Email is required')
      expect(axios.put).not.toHaveBeenCalled()
    })

    it('sends profile update request', async () => {
      axios.put.mockResolvedValue({
        status: 200,
        data: {
          user: {
            id: 1,
            name: 'Updated Name',
            email: 'test@example.com'
          }
        }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.profileForm.name = 'Updated Name'
      vm.profileForm.email = 'test@example.com'
      vm.profileForm.birthday = '1990-05-15'

      await vm.updateProfile()
      await flushPromises()

      expect(axios.put).toHaveBeenCalledWith('/api/users/profile', {
        name: 'Updated Name',
        email: 'test@example.com',
        birthday: '1990-05-15'
      })
    })

    it('shows success message on profile update', async () => {
      axios.put.mockResolvedValue({
        status: 200,
        data: {
          user: { name: 'Updated' }
        }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.profileForm.name = 'Updated Name'
      vm.profileForm.email = 'test@example.com'

      await vm.updateProfile()
      await flushPromises()

      expect(vm.successMessage).toContain('Profile updated')
    })

    it('handles email conflict error', async () => {
      axios.put.mockRejectedValue({
        response: {
          status: 409,
          data: { message: 'Email in use' }
        }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.profileForm.name = 'Test'
      vm.profileForm.email = 'existing@example.com'

      await vm.updateProfile()
      await flushPromises()

      expect(vm.errors.email).toBe('Email already in use by another account')
    })

    it('sets loading state during profile update', async () => {
      let resolveUpdate
      axios.put.mockReturnValue(
        new Promise((resolve) => {
          resolveUpdate = resolve
        })
      )

      createWrapper()

      const vm = wrapper.vm
      vm.profileForm.name = 'Test'
      vm.profileForm.email = 'test@example.com'

      const updatePromise = vm.updateProfile()

      expect(vm.loading).toBe(true)

      resolveUpdate({ status: 200, data: { user: {} } })
      await updatePromise
      await flushPromises()

      expect(vm.loading).toBe(false)
    })
  })

  // ==========================================================================
  // PASSWORD CHANGE TESTS
  // ==========================================================================

  describe('Password Change', () => {
    it('validates current password is required', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.passwordForm.currentPassword = ''
      vm.passwordForm.newPassword = 'newpassword123'
      vm.passwordForm.confirmPassword = 'newpassword123'

      await vm.changePassword()
      await flushPromises()

      expect(vm.passwordErrors.currentPassword).toBe('Current password is required')
      expect(axios.put).not.toHaveBeenCalled()
    })

    it('validates new password is required', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.passwordForm.currentPassword = 'oldpassword'
      vm.passwordForm.newPassword = ''
      vm.passwordForm.confirmPassword = ''

      await vm.changePassword()
      await flushPromises()

      expect(vm.passwordErrors.newPassword).toBe('New password is required')
    })

    it('validates password minimum length', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.passwordForm.currentPassword = 'oldpassword'
      vm.passwordForm.newPassword = 'short'
      vm.passwordForm.confirmPassword = 'short'

      await vm.changePassword()
      await flushPromises()

      expect(vm.passwordErrors.newPassword).toBe('Password must be at least 8 characters')
    })

    it('validates passwords must match', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.passwordForm.currentPassword = 'oldpassword'
      vm.passwordForm.newPassword = 'newpassword123'
      vm.passwordForm.confirmPassword = 'different123'

      await vm.changePassword()
      await flushPromises()

      expect(vm.passwordErrors.confirmPassword).toBe('Passwords do not match')
    })

    it('sends password change request', async () => {
      axios.put.mockResolvedValue({
        status: 200,
        data: {}
      })

      createWrapper()

      const vm = wrapper.vm
      vm.passwordForm.currentPassword = 'oldpassword'
      vm.passwordForm.newPassword = 'newpassword123'
      vm.passwordForm.confirmPassword = 'newpassword123'

      await vm.changePassword()
      await flushPromises()

      expect(axios.put).toHaveBeenCalledWith('/api/users/password', {
        old_password: 'oldpassword',
        new_password: 'newpassword123'
      })
    })

    it('clears form after successful password change', async () => {
      axios.put.mockResolvedValue({
        status: 200,
        data: {}
      })

      createWrapper()

      const vm = wrapper.vm
      vm.passwordForm.currentPassword = 'oldpassword'
      vm.passwordForm.newPassword = 'newpassword123'
      vm.passwordForm.confirmPassword = 'newpassword123'

      await vm.changePassword()
      await flushPromises()

      expect(vm.passwordForm.currentPassword).toBe('')
      expect(vm.passwordForm.newPassword).toBe('')
      expect(vm.passwordForm.confirmPassword).toBe('')
    })

    it('handles incorrect current password error', async () => {
      axios.put.mockRejectedValue({
        response: {
          status: 401,
          data: { message: 'Incorrect password' }
        }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.passwordForm.currentPassword = 'wrongpassword'
      vm.passwordForm.newPassword = 'newpassword123'
      vm.passwordForm.confirmPassword = 'newpassword123'

      await vm.changePassword()
      await flushPromises()

      expect(vm.passwordErrors.currentPassword).toBe('Current password is incorrect')
    })

    it('sets password loading state during password change', async () => {
      let resolveUpdate
      axios.put.mockReturnValue(
        new Promise((resolve) => {
          resolveUpdate = resolve
        })
      )

      createWrapper()

      const vm = wrapper.vm
      vm.passwordForm.currentPassword = 'old'
      vm.passwordForm.newPassword = 'newpassword123'
      vm.passwordForm.confirmPassword = 'newpassword123'

      const changePromise = vm.changePassword()

      expect(vm.passwordLoading).toBe(true)

      resolveUpdate({ status: 200, data: {} })
      await changePromise
      await flushPromises()

      expect(vm.passwordLoading).toBe(false)
    })
  })

  // ==========================================================================
  // PREFERENCES TESTS
  // ==========================================================================

  describe('Preferences', () => {
    it('saves notification preference to localStorage', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.notifications = false
      vm.saveNotifications()

      expect(localStorage.getItem('notifications')).toBe('false')
    })

    it('saves weight unit preference', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.weightUnit = 'kg'

      await vm.saveWeightUnit()

      expect(localStorage.getItem('weightUnit')).toBe('kg')
      expect(mockSettingsStore.updateWeightUnit).toHaveBeenCalledWith('kg')
    })

    it('saves timezone preference', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.timezone = 'Europe/London'

      await vm.saveTimezone()
      await flushPromises()

      expect(mockSettingsStore.updateTimezone).toHaveBeenCalledWith('Europe/London')
    })

    it('detects browser timezone', async () => {
      createWrapper()

      const vm = wrapper.vm

      await vm.detectTimezone()

      expect(vm.timezone).toBe('America/New_York')
    })
  })
})
