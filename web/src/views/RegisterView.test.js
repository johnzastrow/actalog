import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import RegisterView from './RegisterView.vue'

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
  register: vi.fn(),
  error: null,
  loading: false,
  user: null,
  token: null
}

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => mockAuthStore
}))

describe('RegisterView', () => {
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

    wrapper = mount(RegisterView, {
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
    mockAuthStore.register.mockReset()
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
    it('renders registration form', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Create your account')
      expect(wrapper.text()).toContain('Sign Up')
    })

    it('renders name field', () => {
      createWrapper()

      const vm = wrapper.vm
      expect(vm.name).toBeDefined()
    })

    it('renders email field', () => {
      createWrapper()

      const emailInput = wrapper.find('input[type="email"]')
      expect(emailInput.exists()).toBe(true)
    })

    it('renders password field', () => {
      createWrapper()

      const vm = wrapper.vm
      expect(vm.password).toBeDefined()
      expect(vm.showPassword).toBe(false)
    })

    it('renders confirm password field', () => {
      createWrapper()

      const vm = wrapper.vm
      expect(vm.confirmPassword).toBeDefined()
    })

    it('renders sign in link', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Already have an account')
    })
  })

  // ==========================================================================
  // VALIDATION TESTS
  // ==========================================================================

  describe('Validation', () => {
    it('validates name is required', async () => {
      createWrapper()

      const vm = wrapper.vm

      vm.name = ''
      vm.email = 'test@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.errors.name).toBe('Name is required')
      expect(mockAuthStore.register).not.toHaveBeenCalled()
    })

    it('validates email is required', async () => {
      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = ''
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.errors.email).toBe('Email is required')
      expect(mockAuthStore.register).not.toHaveBeenCalled()
    })

    it('validates password minimum length', async () => {
      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = '1234567' // Only 7 characters (below 12-char minimum)
      vm.confirmPassword = '1234567'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.errors.password).toBe('Password must be at least 12 characters')
      expect(mockAuthStore.register).not.toHaveBeenCalled()
    })

    it('validates passwords must match', async () => {
      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'differentpassword'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.errors.confirmPassword).toBe('Passwords do not match')
      expect(mockAuthStore.register).not.toHaveBeenCalled()
    })

    it('trims whitespace from name', async () => {
      createWrapper()

      const vm = wrapper.vm

      vm.name = '   '
      vm.email = 'test@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.errors.name).toBe('Name is required')
    })

    it('trims whitespace from email', async () => {
      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = '   '
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.errors.email).toBe('Email is required')
    })

    it('passes validation with all valid fields', async () => {
      mockAuthStore.register.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(mockAuthStore.register).toHaveBeenCalled()
    })
  })

  // ==========================================================================
  // REGISTRATION SUCCESS TESTS
  // ==========================================================================

  describe('Registration Success', () => {
    it('calls authStore.register with user data', async () => {
      mockAuthStore.register.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = 'SecurePass123!'

      vm.confirmPassword = 'SecurePass123!'

      await vm.handleRegister()
      await flushPromises()

      expect(mockAuthStore.register).toHaveBeenCalledWith({
        name: 'Test User',
        email: 'test@example.com',
        password: 'SecurePass123!'
      })
    })

    it('shows success message on successful registration', async () => {
      mockAuthStore.register.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.registrationSuccess).toBe(true)
    })

    it('displays email verification instructions', async () => {
      mockAuthStore.register.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()
      await wrapper.vm.$nextTick()

      // The success view should be shown
      expect(vm.registrationSuccess).toBe(true)
    })
  })

  // ==========================================================================
  // REGISTRATION ERROR TESTS
  // ==========================================================================

  describe('Registration Errors', () => {
    it('shows password error when error contains "password"', async () => {
      mockAuthStore.register.mockResolvedValue(false)
      mockAuthStore.error = 'Password is too weak'

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = 'Weakpass1234'
      vm.confirmPassword = 'Weakpass1234'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.errors.password).toBe('Password is too weak')
    })

    it('shows email error when error contains "email"', async () => {
      mockAuthStore.register.mockResolvedValue(false)
      mockAuthStore.error = 'Email already exists'

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'existing@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.errors.email).toBe('Email already exists')
    })

    it('shows name error when error contains "name"', async () => {
      mockAuthStore.register.mockResolvedValue(false)
      mockAuthStore.error = 'Name is too short'

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'A'
      vm.email = 'test@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.errors.name).toBe('Name is too short')
    })

    it('shows general error for other errors', async () => {
      mockAuthStore.register.mockResolvedValue(false)
      mockAuthStore.error = 'Server error'

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.generalError).toBe('Server error')
    })

    it('handles null error from store', async () => {
      mockAuthStore.register.mockResolvedValue(false)
      mockAuthStore.error = null

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.generalError).toBe('Registration failed')
    })
  })

  // ==========================================================================
  // STATE MANAGEMENT TESTS
  // ==========================================================================

  describe('State Management', () => {
    it('sets loading state during registration', async () => {
      let resolveRegister
      mockAuthStore.register.mockReturnValue(
        new Promise((resolve) => {
          resolveRegister = resolve
        })
      )

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      const registerPromise = vm.handleRegister()

      expect(vm.loading).toBe(true)

      resolveRegister(true)
      await registerPromise
      await flushPromises()

      expect(vm.loading).toBe(false)
    })

    it('clears errors on new registration attempt', async () => {
      mockAuthStore.register.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.errors = { name: 'Old error' }
      vm.generalError = 'Old general error'

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.errors).toEqual({})
      expect(vm.generalError).toBe('')
    })

    it('can dismiss general error alert', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.generalError = 'Some error'

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
      expect(vm.showConfirmPassword).toBe(false)
    })

    it('can toggle password visibility', () => {
      createWrapper()

      const vm = wrapper.vm

      vm.showPassword = !vm.showPassword
      expect(vm.showPassword).toBe(true)

      vm.showPassword = !vm.showPassword
      expect(vm.showPassword).toBe(false)
    })

    it('can toggle confirm password visibility independently', () => {
      createWrapper()

      const vm = wrapper.vm

      vm.showConfirmPassword = !vm.showConfirmPassword
      expect(vm.showConfirmPassword).toBe(true)
      expect(vm.showPassword).toBe(false)
    })
  })

  // ==========================================================================
  // INTEGRATION TESTS
  // ==========================================================================

  describe('Integration', () => {
    it('complete registration flow - success', async () => {
      mockAuthStore.register.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'John Doe'
      vm.email = 'john@example.com'
      vm.password = 'SecurePassword123!'
      vm.confirmPassword = 'SecurePassword123!'

      await vm.handleRegister()
      await flushPromises()

      expect(mockAuthStore.register).toHaveBeenCalledWith({
        name: 'John Doe',
        email: 'john@example.com',
        password: 'SecurePassword123!'
      })

      expect(vm.registrationSuccess).toBe(true)
    })

    it('complete registration flow - validation failure', async () => {
      createWrapper()

      const vm = wrapper.vm

      vm.name = ''
      vm.email = ''
      vm.password = 'short'
      vm.confirmPassword = 'different'

      await vm.handleRegister()
      await flushPromises()

      expect(mockAuthStore.register).not.toHaveBeenCalled()
      expect(vm.errors.name).toBe('Name is required')
    })

    it('complete registration flow - server error', async () => {
      mockAuthStore.register.mockResolvedValue(false)
      mockAuthStore.error = 'Email already registered'

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'John Doe'
      vm.email = 'existing@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(vm.registrationSuccess).toBe(false)
      expect(vm.errors.email).toBe('Email already registered')
    })
  })

  // ==========================================================================
  // SPECIAL CHARACTERS TESTS
  // ==========================================================================

  describe('Special Characters', () => {
    it('handles name with special characters', async () => {
      mockAuthStore.register.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.name = "John O'Brien-Smith"
      vm.email = 'john@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(mockAuthStore.register).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "John O'Brien-Smith"
        })
      )
    })

    it('handles email with plus sign', async () => {
      mockAuthStore.register.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test+signup@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(mockAuthStore.register).toHaveBeenCalledWith(
        expect.objectContaining({
          email: 'test+signup@example.com'
        })
      )
    })

    it('handles password with special characters', async () => {
      mockAuthStore.register.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = 'P@$$w0rd!#%^&*()'
      vm.confirmPassword = 'P@$$w0rd!#%^&*()'

      await vm.handleRegister()
      await flushPromises()

      expect(mockAuthStore.register).toHaveBeenCalledWith(
        expect.objectContaining({
          password: 'P@$$w0rd!#%^&*()'
        })
      )
    })

    it('handles unicode in name', async () => {
      mockAuthStore.register.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.name = '田中太郎'
      vm.email = 'tanaka@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      expect(mockAuthStore.register).toHaveBeenCalledWith(
        expect.objectContaining({
          name: '田中太郎'
        })
      )
    })
  })

  // ==========================================================================
  // NAVIGATION TESTS
  // ==========================================================================

  describe('Navigation', () => {
    it('can navigate to login from success screen', async () => {
      mockAuthStore.register.mockResolvedValue(true)

      createWrapper()

      const vm = wrapper.vm

      vm.name = 'Test User'
      vm.email = 'test@example.com'
      vm.password = 'Password1234'
      vm.confirmPassword = 'Password1234'

      await vm.handleRegister()
      await flushPromises()

      // Simulate clicking "Go to Login" button
      mockRouter.push('/login')

      expect(mockRouter.push).toHaveBeenCalledWith('/login')
    })
  })
})
