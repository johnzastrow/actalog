import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ResetPasswordView from './ResetPasswordView.vue'

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
    useRoute: vi.fn(() => ({
      params: { token: 'valid-reset-token' }
    })),
    useRouter: vi.fn(() => ({
      push: vi.fn(),
    })),
    RouterLink: {
      template: '<a><slot /></a>'
    }
  }
})

import axios from '@/utils/axios'
import { useRoute } from 'vue-router'

describe('ResetPasswordView', () => {
  let wrapper

  const createWrapper = (routeParams = { token: 'valid-reset-token' }) => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useRoute.mockReturnValue({
      params: routeParams
    })

    wrapper = mount(ResetPasswordView, {
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
    it('renders the reset password form', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Set New Password')
      expect(wrapper.text()).toContain('Enter your new password below')
    })

    it('renders password input fields', () => {
      createWrapper()

      const passwordInputs = wrapper.findAll('input[type="password"]')
      expect(passwordInputs.length).toBe(2)
    })

    it('renders submit button', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Reset Password')
    })

    it('renders back to sign in link', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Back to Sign In')
    })

    it('shows error when no token provided', async () => {
      createWrapper({ token: '' })
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.errorMessage).toContain('Invalid reset link')
    })
  })

  // ==========================================================================
  // VALIDATION TESTS
  // ==========================================================================

  describe('Validation', () => {
    it('validates password minimum length', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.newPassword = 'short'
      vm.confirmPassword = 'short'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.errors.password).toBe('Password must be at least 8 characters long')
      expect(axios.post).not.toHaveBeenCalled()
    })

    it('validates passwords must match', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.newPassword = 'password123'
      vm.confirmPassword = 'different123'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.errors.confirmPassword).toBe('Passwords do not match')
      expect(axios.post).not.toHaveBeenCalled()
    })

    it('passes validation with valid passwords', async () => {
      axios.post.mockResolvedValue({
        data: { message: 'Password reset successfully' }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.newPassword = 'validpassword123'
      vm.confirmPassword = 'validpassword123'

      await vm.handleSubmit()
      await flushPromises()

      expect(axios.post).toHaveBeenCalled()
    })
  })

  // ==========================================================================
  // FORM SUBMISSION TESTS
  // ==========================================================================

  describe('Form Submission', () => {
    it('sends reset request with token and new password', async () => {
      axios.post.mockResolvedValue({
        data: {
          message: 'Password has been reset successfully'
        }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.newPassword = 'newpassword123'
      vm.confirmPassword = 'newpassword123'

      await vm.handleSubmit()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/auth/reset-password', {
        token: 'valid-reset-token',
        new_password: 'newpassword123'
      })
    })

    it('shows success message after successful reset', async () => {
      axios.post.mockResolvedValue({
        data: {
          message: 'Password has been reset successfully. You can now sign in with your new password.'
        }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.newPassword = 'newpassword123'
      vm.confirmPassword = 'newpassword123'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.successMessage).toContain('Password has been reset')
      expect(vm.resetSuccess).toBe(true)
    })

    it('hides form after successful reset', async () => {
      axios.post.mockResolvedValue({
        data: { message: 'Success' }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.newPassword = 'newpassword123'
      vm.confirmPassword = 'newpassword123'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.resetSuccess).toBe(true)
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('shows error on expired token', async () => {
      axios.post.mockRejectedValue({
        response: {
          data: {
            message: 'Reset token has expired'
          }
        }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.newPassword = 'newpassword123'
      vm.confirmPassword = 'newpassword123'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.errorMessage).toBe('Reset token has expired')
    })

    it('shows error on invalid token', async () => {
      axios.post.mockRejectedValue({
        response: {
          data: {
            message: 'Invalid reset token'
          }
        }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.newPassword = 'newpassword123'
      vm.confirmPassword = 'newpassword123'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.errorMessage).toBe('Invalid reset token')
    })

    it('shows generic error on network failure', async () => {
      axios.post.mockRejectedValue(new Error('Network error'))

      createWrapper()

      const vm = wrapper.vm
      vm.newPassword = 'newpassword123'
      vm.confirmPassword = 'newpassword123'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.errorMessage).toContain('Failed to reset password')
    })

    it('clears previous errors on new submission', async () => {
      axios.post.mockResolvedValue({
        data: { message: 'Success' }
      })

      createWrapper()

      const vm = wrapper.vm

      // Set previous errors
      vm.errors = { password: 'Previous error' }
      vm.errorMessage = 'Previous error message'

      vm.newPassword = 'newpassword123'
      vm.confirmPassword = 'newpassword123'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.errors).toEqual({})
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
      vm.newPassword = 'newpassword123'
      vm.confirmPassword = 'newpassword123'

      const submitPromise = vm.handleSubmit()

      expect(vm.loading).toBe(true)

      resolvePost({ data: { message: 'Success' } })
      await submitPromise
      await flushPromises()

      expect(vm.loading).toBe(false)
    })

    it('extracts token from route params on mount', async () => {
      createWrapper({ token: 'my-special-token' })
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.token).toBe('my-special-token')
    })

    it('initializes with empty password fields', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.newPassword).toBe('')
      expect(vm.confirmPassword).toBe('')
      expect(vm.loading).toBe(false)
      expect(vm.resetSuccess).toBe(false)
    })
  })

  // ==========================================================================
  // SPECIAL CHARACTERS TESTS
  // ==========================================================================

  describe('Special Characters', () => {
    it('handles password with special characters', async () => {
      axios.post.mockResolvedValue({
        data: { message: 'Success' }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.newPassword = 'P@$$w0rd!#%^&*()'
      vm.confirmPassword = 'P@$$w0rd!#%^&*()'

      await vm.handleSubmit()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/auth/reset-password', {
        token: 'valid-reset-token',
        new_password: 'P@$$w0rd!#%^&*()'
      })
    })

    it('handles password with unicode characters', async () => {
      axios.post.mockResolvedValue({
        data: { message: 'Success' }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.newPassword = 'password日本語123'
      vm.confirmPassword = 'password日本語123'

      await vm.handleSubmit()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/auth/reset-password', {
        token: 'valid-reset-token',
        new_password: 'password日本語123'
      })
    })

    it('handles password with spaces', async () => {
      axios.post.mockResolvedValue({
        data: { message: 'Success' }
      })

      createWrapper()

      const vm = wrapper.vm
      vm.newPassword = 'pass word with spaces'
      vm.confirmPassword = 'pass word with spaces'

      await vm.handleSubmit()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/auth/reset-password', {
        token: 'valid-reset-token',
        new_password: 'pass word with spaces'
      })
    })
  })

  // ==========================================================================
  // INTEGRATION TESTS
  // ==========================================================================

  describe('Integration', () => {
    it('complete reset flow - success', async () => {
      axios.post.mockResolvedValue({
        data: {
          message: 'Password has been reset successfully'
        }
      })

      createWrapper({ token: 'valid-token-123' })

      const vm = wrapper.vm

      vm.newPassword = 'SecureNewPassword123!'
      vm.confirmPassword = 'SecureNewPassword123!'

      await vm.handleSubmit()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/auth/reset-password', {
        token: 'valid-token-123',
        new_password: 'SecureNewPassword123!'
      })
      expect(vm.resetSuccess).toBe(true)
      expect(vm.successMessage).toBeTruthy()
    })

    it('complete reset flow - validation failure', async () => {
      createWrapper()

      const vm = wrapper.vm

      // Short password
      vm.newPassword = 'short'
      vm.confirmPassword = 'short'

      await vm.handleSubmit()
      await flushPromises()

      expect(axios.post).not.toHaveBeenCalled()
      expect(vm.errors.password).toBeTruthy()
      expect(vm.resetSuccess).toBe(false)
    })

    it('complete reset flow - token expired', async () => {
      axios.post.mockRejectedValue({
        response: {
          data: {
            message: 'Reset token has expired. Please request a new password reset.'
          }
        }
      })

      createWrapper({ token: 'expired-token' })

      const vm = wrapper.vm

      vm.newPassword = 'newpassword123'
      vm.confirmPassword = 'newpassword123'

      await vm.handleSubmit()
      await flushPromises()

      expect(vm.errorMessage).toContain('expired')
      expect(vm.resetSuccess).toBe(false)
    })
  })
})
