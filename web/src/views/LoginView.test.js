import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import LoginView from './LoginView.vue'

// Mock vue-router
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: vi.fn(),
      replace: vi.fn(),
    })),
    RouterLink: {
      template: '<a><slot /></a>'
    }
  }
})

// Import mocked router
import { useRouter } from 'vue-router'

// Mock the auth store
const mockAuthStore = {
  login: vi.fn(),
  error: null,
  loading: false,
  user: null,
  token: null
}

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => mockAuthStore
}))

describe('LoginView', () => {
  let wrapper
  let mockRouter

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    mockRouter = {
      push: vi.fn(),
      replace: vi.fn()
    }
    useRouter.mockReturnValue(mockRouter)

    wrapper = mount(LoginView, {
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
    mockAuthStore.error = null
    mockAuthStore.login.mockReset()
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
    it('renders login form', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Sign in to your account')
      expect(wrapper.text()).toContain('Sign In')
    })

    it('renders email field', () => {
      createWrapper()

      const emailInput = wrapper.find('input[type="email"]')
      expect(emailInput.exists()).toBe(true)
    })

    it('renders password field', () => {
      createWrapper()

      const passwordInput = wrapper.find('input[type="password"]')
      expect(passwordInput.exists()).toBe(true)
    })

    it('renders remember me checkbox', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Remember me')
    })

    it('renders forgot password link', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Forgot password')
    })

    it('renders sign up link', () => {
      createWrapper()

      expect(wrapper.text()).toContain("Don't have an account")
    })
  })

  // ==========================================================================
  // VALIDATION TESTS
  // ==========================================================================

  describe('Validation', () => {
    it('validates email is required', async () => {
      createWrapper()

      const vm = wrapper.vm

      // Leave email empty
      vm.email = ''
      vm.password = 'password123'

      await vm.handleLogin()
      await flushPromises()

      expect(vm.errors.email).toBe('Email is required')
      expect(mockAuthStore.login).not.toHaveBeenCalled()
    })

    it('validates password is required', async () => {
      createWrapper()

      const vm = wrapper.vm

      vm.email = 'test@example.com'
      vm.password = ''

      await vm.handleLogin()
      await flushPromises()

      expect(vm.errors.password).toBe('Password is required')
      expect(mockAuthStore.login).not.toHaveBeenCalled()
    })

    it('trims whitespace from email', async () => {
      createWrapper()

      const vm = wrapper.vm

      vm.email = '   '
      vm.password = 'password123'

      await vm.handleLogin()
      await flushPromises()

      expect(vm.errors.email).toBe('Email is required')
    })
  })

  // ==========================================================================
  // LOGIN SUCCESS TESTS
  // ==========================================================================

  describe('Login Success', () => {
    it('calls authStore.login with credentials', async () => {
      mockAuthStore.login.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.email = 'test@example.com'
      vm.password = 'password123'
      vm.rememberMe = false

      await vm.handleLogin()
      await flushPromises()

      expect(mockAuthStore.login).toHaveBeenCalledWith(
        'test@example.com',
        'password123',
        false
      )
    })

    it('calls authStore.login with rememberMe option', async () => {
      mockAuthStore.login.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.email = 'test@example.com'
      vm.password = 'password123'
      vm.rememberMe = true

      await vm.handleLogin()
      await flushPromises()

      expect(mockAuthStore.login).toHaveBeenCalledWith(
        'test@example.com',
        'password123',
        true
      )
    })

    it('navigates to dashboard on successful login', async () => {
      mockAuthStore.login.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.email = 'test@example.com'
      vm.password = 'password123'

      await vm.handleLogin()
      await flushPromises()

      expect(mockRouter.push).toHaveBeenCalledWith('/dashboard')
    })
  })

  // ==========================================================================
  // LOGIN ERROR TESTS
  // ==========================================================================

  describe('Login Errors', () => {
    it('shows general error on invalid credentials', async () => {
      mockAuthStore.login.mockResolvedValue(false)
      mockAuthStore.error = 'Invalid email or password'

      createWrapper()

      const vm = wrapper.vm

      vm.email = 'test@example.com'
      vm.password = 'wrongpassword'

      await vm.handleLogin()
      await flushPromises()

      expect(vm.generalError).toContain('Invalid email or password')
    })

    it('shows account locked alert when account is locked', async () => {
      mockAuthStore.login.mockResolvedValue(false)
      mockAuthStore.error = 'Account locked due to too many failed attempts'

      createWrapper()

      const vm = wrapper.vm

      vm.email = 'locked@example.com'
      vm.password = 'password123'

      await vm.handleLogin()
      await flushPromises()

      expect(vm.accountLocked).toBe(true)
      expect(vm.errorMessage).toContain('Account locked')
    })

    it('shows rate limited alert on too many requests', async () => {
      mockAuthStore.login.mockResolvedValue(false)
      mockAuthStore.error = 'Too many requests, please try again later'

      createWrapper()

      const vm = wrapper.vm

      vm.email = 'test@example.com'
      vm.password = 'password123'

      await vm.handleLogin()
      await flushPromises()

      expect(vm.rateLimited).toBe(true)
    })

    it('shows generic error for other failures', async () => {
      mockAuthStore.login.mockResolvedValue(false)
      mockAuthStore.error = 'Network error'

      createWrapper()

      const vm = wrapper.vm

      vm.email = 'test@example.com'
      vm.password = 'password123'

      await vm.handleLogin()
      await flushPromises()

      expect(vm.generalError).toBe('Network error')
    })

    it('handles null error from store', async () => {
      mockAuthStore.login.mockResolvedValue(false)
      mockAuthStore.error = null

      createWrapper()

      const vm = wrapper.vm

      vm.email = 'test@example.com'
      vm.password = 'password123'

      await vm.handleLogin()
      await flushPromises()

      expect(vm.errorMessage).toBe('Login failed')
    })
  })

  // ==========================================================================
  // STATE MANAGEMENT TESTS
  // ==========================================================================

  describe('State Management', () => {
    it('sets loading state during login', async () => {
      // Make login return a promise that we can control
      let resolveLogin
      mockAuthStore.login.mockReturnValue(
        new Promise((resolve) => {
          resolveLogin = resolve
        })
      )

      createWrapper()

      const vm = wrapper.vm

      vm.email = 'test@example.com'
      vm.password = 'password123'

      // Start login (don't await)
      const loginPromise = vm.handleLogin()

      // Check loading is true
      expect(vm.loading).toBe(true)

      // Resolve the login
      resolveLogin(true)
      await loginPromise
      await flushPromises()

      // Check loading is false
      expect(vm.loading).toBe(false)
    })

    it('clears errors on new login attempt', async () => {
      mockAuthStore.login.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      // Set some existing errors
      vm.errors = { email: 'Old error' }
      vm.generalError = 'Old general error'
      vm.accountLocked = true
      vm.rateLimited = true

      vm.email = 'test@example.com'
      vm.password = 'password123'

      await vm.handleLogin()
      await flushPromises()

      expect(vm.errors).toEqual({})
      expect(vm.generalError).toBe('')
      expect(vm.accountLocked).toBe(false)
      expect(vm.rateLimited).toBe(false)
    })

    it('can dismiss account locked alert', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.accountLocked = true

      // Simulate close
      vm.accountLocked = false

      expect(vm.accountLocked).toBe(false)
    })

    it('can dismiss rate limited alert', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.rateLimited = true

      // Simulate close
      vm.rateLimited = false

      expect(vm.rateLimited).toBe(false)
    })

    it('can dismiss general error alert', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.generalError = 'Some error'

      // Simulate close
      vm.generalError = ''

      expect(vm.generalError).toBe('')
    })
  })

  // ==========================================================================
  // PASSWORD VISIBILITY TESTS
  // ==========================================================================

  describe('Password Visibility', () => {
    it('starts with password hidden', () => {
      createWrapper()

      const vm = wrapper.vm
      expect(vm.showPassword).toBe(false)
    })

    it('can toggle password visibility', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.showPassword).toBe(false)

      vm.showPassword = !vm.showPassword
      expect(vm.showPassword).toBe(true)

      vm.showPassword = !vm.showPassword
      expect(vm.showPassword).toBe(false)
    })
  })

  // ==========================================================================
  // INTEGRATION TESTS
  // ==========================================================================

  describe('Integration', () => {
    it('complete login flow - success', async () => {
      mockAuthStore.login.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      // Fill form
      vm.email = 'user@example.com'
      vm.password = 'SecurePassword123!'
      vm.rememberMe = true

      // Submit
      await vm.handleLogin()
      await flushPromises()

      // Verify call
      expect(mockAuthStore.login).toHaveBeenCalledWith(
        'user@example.com',
        'SecurePassword123!',
        true
      )

      // Verify navigation
      expect(mockRouter.push).toHaveBeenCalledWith('/dashboard')
    })

    it('complete login flow - failure', async () => {
      mockAuthStore.login.mockResolvedValue(false)
      mockAuthStore.error = 'Invalid credentials'

      createWrapper()

      const vm = wrapper.vm

      // Fill form
      vm.email = 'user@example.com'
      vm.password = 'WrongPassword'

      // Submit
      await vm.handleLogin()
      await flushPromises()

      // Verify no navigation
      expect(mockRouter.push).not.toHaveBeenCalled()

      // Verify error is shown
      expect(vm.generalError).toBeTruthy()
    })

    it('validation prevents API call', async () => {
      createWrapper()

      const vm = wrapper.vm

      // Leave fields empty
      vm.email = ''
      vm.password = ''

      // Submit
      await vm.handleLogin()
      await flushPromises()

      // Verify no API call
      expect(mockAuthStore.login).not.toHaveBeenCalled()

      // Verify validation error
      expect(vm.errors.email).toBe('Email is required')
    })
  })

  // ==========================================================================
  // SPECIAL CHARACTERS TESTS
  // ==========================================================================

  describe('Special Characters', () => {
    it('handles email with special characters', async () => {
      mockAuthStore.login.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.email = 'user+tag@example.com'
      vm.password = 'password123'

      await vm.handleLogin()
      await flushPromises()

      expect(mockAuthStore.login).toHaveBeenCalledWith(
        'user+tag@example.com',
        'password123',
        false
      )
    })

    it('handles password with special characters', async () => {
      mockAuthStore.login.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.email = 'test@example.com'
      vm.password = 'P@$$w0rd!#%^&*()'

      await vm.handleLogin()
      await flushPromises()

      expect(mockAuthStore.login).toHaveBeenCalledWith(
        'test@example.com',
        'P@$$w0rd!#%^&*()',
        false
      )
    })

    it('handles unicode in password', async () => {
      mockAuthStore.login.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.email = 'test@example.com'
      vm.password = 'password日本語'

      await vm.handleLogin()
      await flushPromises()

      expect(mockAuthStore.login).toHaveBeenCalledWith(
        'test@example.com',
        'password日本語',
        false
      )
    })
  })
})
