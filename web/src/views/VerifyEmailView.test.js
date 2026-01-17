import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import VerifyEmailView from './VerifyEmailView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  }
}))

// Mock vue-router
const mockPush = vi.fn()
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: mockPush,
    })),
    useRoute: vi.fn(() => ({
      query: { token: 'valid-token' }
    }))
  }
})

import axios from '@/utils/axios'
import { useRoute } from 'vue-router'

describe('VerifyEmailView', () => {
  let wrapper

  const createWrapper = (routeQuery = { token: 'valid-token' }) => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useRoute.mockReturnValue({ query: routeQuery })

    wrapper = shallowMount(VerifyEmailView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-progress-circular': { template: '<div class="loading"></div>' },
          'v-icon': { template: '<i></i>' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-text-field': { template: '<input />' }
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    mockPush.mockClear()
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
    it('shows loading state initially', () => {
      axios.get.mockReturnValue(new Promise(() => {})) // Never resolves
      createWrapper()

      const vm = wrapper.vm

      expect(vm.loading).toBe(true)
    })

    it('calls verify API with token on mount', async () => {
      axios.get.mockResolvedValue({ status: 200 })
      createWrapper({ token: 'test-token' })
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/auth/verify-email', {
        params: { token: 'test-token' }
      })
    })
  })

  // ==========================================================================
  // SUCCESS STATE TESTS
  // ==========================================================================

  describe('Success State', () => {
    it('shows success state when verification succeeds', async () => {
      axios.get.mockResolvedValue({ status: 200 })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.success).toBe(true)
      expect(vm.loading).toBe(false)
      expect(vm.error).toBe(false)
    })

    it('displays success message', async () => {
      axios.get.mockResolvedValue({ status: 200 })
      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Email Verified')
    })
  })

  // ==========================================================================
  // ERROR STATE TESTS
  // ==========================================================================

  describe('Error State', () => {
    it('shows error when no token provided', async () => {
      createWrapper({ token: null })
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe(true)
      expect(vm.errorMessage).toContain('Invalid verification link')
    })

    it('shows error when verification fails', async () => {
      axios.get.mockRejectedValue({
        response: { data: { message: 'Token expired' } }
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe(true)
      expect(vm.errorMessage).toBe('Token expired')
    })

    it('shows resend option when token expired', async () => {
      axios.get.mockRejectedValue({
        response: { data: { message: 'Token has expired' } }
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.showResend).toBe(true)
    })

    it('shows generic error for unknown errors', async () => {
      axios.get.mockRejectedValue(new Error('Network error'))
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe(true)
      expect(vm.errorMessage).toContain('error occurred')
    })
  })

  // ==========================================================================
  // RESEND VERIFICATION TESTS
  // ==========================================================================

  describe('Resend Verification', () => {
    it('does not resend without email', async () => {
      axios.get.mockRejectedValue({
        response: { data: { message: 'Token has expired' } }
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.email = ''

      await vm.resendVerification()

      expect(axios.post).not.toHaveBeenCalled()
    })

    it('calls resend API with email', async () => {
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
      axios.get.mockRejectedValue({
        response: { data: { message: 'Token has expired' } }
      })
      axios.post.mockResolvedValue({})

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.resendVerification()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/auth/resend-verification', {
        email: 'test@example.com'
      })

      alertSpy.mockRestore()
    })

    it('shows success alert on resend', async () => {
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
      axios.get.mockRejectedValue({
        response: { data: { message: 'Token has expired' } }
      })
      axios.post.mockResolvedValue({})

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.resendVerification()
      await flushPromises()

      expect(alertSpy).toHaveBeenCalledWith(expect.stringContaining('Verification email sent'))
      expect(vm.showResend).toBe(false)

      alertSpy.mockRestore()
    })

    it('handles resend error', async () => {
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
      axios.get.mockRejectedValue({
        response: { data: { message: 'Token has expired' } }
      })
      axios.post.mockRejectedValue({
        response: { data: { message: 'Email not found' } }
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.resendVerification()
      await flushPromises()

      expect(alertSpy).toHaveBeenCalledWith('Email not found')

      alertSpy.mockRestore()
    })

    it('sets resending state during API call', async () => {
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
      let resolvePost
      axios.get.mockRejectedValue({
        response: { data: { message: 'Token has expired' } }
      })
      axios.post.mockReturnValue(new Promise(resolve => { resolvePost = resolve }))

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      const resendPromise = vm.resendVerification()

      expect(vm.resending).toBe(true)

      resolvePost({})
      await resendPromise
      await flushPromises()

      expect(vm.resending).toBe(false)

      alertSpy.mockRestore()
    })
  })

  // ==========================================================================
  // NAVIGATION TESTS
  // ==========================================================================

  describe('Navigation', () => {
    it('navigates to login on goToLogin', async () => {
      axios.get.mockResolvedValue({ status: 200 })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.goToLogin()

      expect(mockPush).toHaveBeenCalledWith('/login')
    })
  })
})
