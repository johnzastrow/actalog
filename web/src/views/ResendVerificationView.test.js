import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ResendVerificationView from './ResendVerificationView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
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
    }))
  }
})

import axios from '@/utils/axios'

describe('ResendVerificationView', () => {
  let wrapper

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(ResendVerificationView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-card-title': { template: '<div><slot /></div>' },
          'v-card-text': { template: '<div><slot /></div>' },
          'v-form': { template: '<form @submit.prevent><slot /></form>' },
          'v-text-field': { template: '<input />' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-icon': { template: '<i></i>' },
          'router-link': { template: '<a><slot /></a>' }
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
    it('initializes with empty email', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.email).toBe('')
    })

    it('initializes with loading false', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)
    })

    it('initializes with emailSent false', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.emailSent).toBe(false)
    })

    it('initializes with empty errors', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.errors).toEqual({})
    })
  })

  // ==========================================================================
  // FORM VALIDATION TESTS
  // ==========================================================================

  describe('Form Validation', () => {
    it('shows error when email is empty', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.email = ''

      await vm.handleResend()

      expect(vm.errors.email).toBe('Email is required')
      expect(axios.post).not.toHaveBeenCalled()
    })

    it('clears errors before validation', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.errors = { email: 'Previous error', general: 'Another error' }
      vm.email = ''

      await vm.handleResend()

      // Should clear old general error but set new email error
      expect(vm.errors.general).toBeUndefined()
    })
  })

  // ==========================================================================
  // RESEND API TESTS
  // ==========================================================================

  describe('Resend API', () => {
    it('calls API with email', async () => {
      axios.post.mockResolvedValue({ status: 200 })
      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.handleResend()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/auth/resend-verification', {
        email: 'test@example.com'
      })
    })

    it('sets emailSent true on success', async () => {
      axios.post.mockResolvedValue({ status: 200 })
      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.handleResend()
      await flushPromises()

      expect(vm.emailSent).toBe(true)
    })

    it('sets loading state during API call', async () => {
      let resolvePost
      axios.post.mockReturnValue(new Promise(resolve => { resolvePost = resolve }))

      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      const resendPromise = vm.handleResend()

      expect(vm.loading).toBe(true)

      resolvePost({ status: 200 })
      await resendPromise
      await flushPromises()

      expect(vm.loading).toBe(false)
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('handles 404 - account not found', async () => {
      axios.post.mockRejectedValue({
        response: { status: 404 }
      })
      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.handleResend()
      await flushPromises()

      expect(vm.errors.general).toBe('No account found with this email address')
    })

    it('handles 400 - already verified or failed', async () => {
      axios.post.mockRejectedValue({
        response: {
          status: 400,
          data: { error: 'Email already verified' }
        }
      })
      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.handleResend()
      await flushPromises()

      expect(vm.errors.general).toBe('Email already verified')
    })

    it('handles 400 without error message', async () => {
      axios.post.mockRejectedValue({
        response: { status: 400, data: {} }
      })
      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.handleResend()
      await flushPromises()

      expect(vm.errors.general).toContain('already verified or request failed')
    })

    it('handles generic error', async () => {
      axios.post.mockRejectedValue(new Error('Network error'))
      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.handleResend()
      await flushPromises()

      expect(vm.errors.general).toBe('Failed to send verification email. Please try again.')
    })
  })

  // ==========================================================================
  // SUCCESS STATE TESTS
  // ==========================================================================

  describe('Success State', () => {
    it('shows success message when emailSent is true', async () => {
      axios.post.mockResolvedValue({ status: 200 })
      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.handleResend()
      await flushPromises()

      expect(vm.emailSent).toBe(true)
      expect(wrapper.text()).toContain('Verification Email Sent')
    })
  })
})
