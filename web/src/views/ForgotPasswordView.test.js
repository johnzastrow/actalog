import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ForgotPasswordView from './ForgotPasswordView.vue'

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

import axios from '@/utils/axios'

describe('ForgotPasswordView', () => {
  let wrapper

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = mount(ForgotPasswordView, {
      global: {
        plugins: [pinia],
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>'
          }
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
  })

  // ==========================================================================
  // RENDER TESTS
  // ==========================================================================

  describe('Rendering', () => {
    it('renders the forgot password form', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Reset Password')
      expect(wrapper.text()).toContain('Enter your email to receive a reset link')
    })

    it('renders email input field', () => {
      createWrapper()

      const emailInput = wrapper.find('input[type="email"]')
      expect(emailInput.exists()).toBe(true)
    })

    it('renders submit button', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Send Reset Link')
    })

    it('renders back to sign in link', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Back to Sign In')
    })
  })

  // ==========================================================================
  // FORM SUBMISSION TESTS
  // ==========================================================================

  describe('Form Submission', () => {
    it('sends reset request with email', async () => {
      axios.post.mockResolvedValue({
        data: {
          message: 'Password reset email sent'
        }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.handleSubmit()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/auth/forgot-password', {
        email: 'test@example.com'
      })
    })

    it('shows success message after submission', async () => {
      axios.post.mockResolvedValue({
        data: {
          message: 'If your email is registered, you will receive a password reset link shortly'
        }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.successMessage).toContain('email is registered')
      expect(vm.submitted).toBe(true)
    })

    it('disables form after successful submission', async () => {
      axios.post.mockResolvedValue({
        data: {
          message: 'Password reset email sent'
        }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.submitted).toBe(true)
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('shows error message on API failure', async () => {
      axios.post.mockRejectedValue({
        response: {
          data: {
            message: 'Invalid email format'
          }
        }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.email = 'invalid-email'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.errorMessage).toBe('Invalid email format')
    })

    it('shows generic error on network failure', async () => {
      axios.post.mockRejectedValue(new Error('Network error'))

      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.errorMessage).toBe('Failed to send reset link. Please try again.')
    })

    it('clears previous messages on new submission', async () => {
      axios.post.mockResolvedValue({
        data: { message: 'Success' }
      })

      createWrapper()

      const vm = wrapper.vm

      // Set previous messages
      vm.errorMessage = 'Previous error'
      vm.successMessage = 'Previous success'

      vm.email = 'test@example.com'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.errorMessage).toBe('')
    })
  })

  // ==========================================================================
  // STATE MANAGEMENT TESTS
  // ==========================================================================

  describe('State Management', () => {
    it('sets loading state during submission', async () => {
      let resolvePost
      axios.post.mockReturnValue(
        new Promise((resolve) => {
          resolvePost = resolve
        })
      )

      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test@example.com'

      const submitPromise = vm.handleSubmit()

      expect(vm.loading).toBe(true)

      resolvePost({ data: { message: 'Success' } })
      await submitPromise
      await flushPromises()

      expect(vm.loading).toBe(false)
    })

    it('initializes with empty state', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.email).toBe('')
      expect(vm.loading).toBe(false)
      expect(vm.submitted).toBe(false)
      expect(vm.successMessage).toBe('')
      expect(vm.errorMessage).toBe('')
    })
  })

  // ==========================================================================
  // SPECIAL CHARACTERS TESTS
  // ==========================================================================

  describe('Special Characters', () => {
    it('handles email with plus sign', async () => {
      axios.post.mockResolvedValue({
        data: { message: 'Success' }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.email = 'test+alias@example.com'

      await vm.handleSubmit()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/auth/forgot-password', {
        email: 'test+alias@example.com'
      })
    })

    it('handles email with subdomain', async () => {
      axios.post.mockResolvedValue({
        data: { message: 'Success' }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.email = 'user@mail.subdomain.example.com'

      await vm.handleSubmit()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/auth/forgot-password', {
        email: 'user@mail.subdomain.example.com'
      })
    })
  })
})
