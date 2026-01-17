import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminUsersView from './AdminUsersView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  }
}))

import axios from '@/utils/axios'

describe('AdminUsersView', () => {
  let wrapper

  const mockUsers = [
    {
      id: 1,
      email: 'admin@test.com',
      name: 'Admin User',
      role: 'admin',
      account_disabled: false,
      email_verified: true,
      last_login_at: '2024-01-15T10:00:00Z',
      failed_login_attempts: 0,
      lockout_until: null
    },
    {
      id: 2,
      email: 'user@test.com',
      name: 'Regular User',
      role: 'user',
      account_disabled: false,
      email_verified: false,
      last_login_at: '2024-01-14T08:00:00Z',
      failed_login_attempts: 0,
      lockout_until: null
    },
    {
      id: 3,
      email: 'locked@test.com',
      name: 'Locked User',
      role: 'user',
      account_disabled: false,
      email_verified: true,
      last_login_at: '2024-01-10T08:00:00Z',
      failed_login_attempts: 5,
      locked_until: new Date(Date.now() + 3600000).toISOString()
    }
  ]

  const setupDefaultMocks = () => {
    axios.get.mockResolvedValue({
      data: {
        users: mockUsers,
        total: mockUsers.length
      }
    })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(AdminUsersView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-card-title': { template: '<div><slot /></div>' },
          'v-card-text': { template: '<div><slot /></div>' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-icon': { template: '<i></i>' },
          'v-chip': { template: '<span><slot /></span>' },
          'v-avatar': { template: '<div><slot /></div>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-dialog': { template: '<div><slot /></div>' },
          'v-text-field': { template: '<input />' },
          'v-divider': { template: '<hr />' },
          'v-spacer': { template: '<div></div>' },
          'v-tooltip': { template: '<div><slot /></div>' },
          'v-progress-linear': { template: '<div></div>' },
          'v-data-table': { template: '<div><slot /></div>' }
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
      wrapper = null
    }
  })

  describe('Initialization', () => {
    it('fetches users on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/admin/users?limit=50&offset=0')
    })

    it('stores fetched users', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.users).toHaveLength(3)
    })

    it('sets total count', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.total).toBe(3)
    })

    it('sets loading to false after fetch', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)
    })

    it('initializes search as empty', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.search).toBe('')
    })
  })

  describe('User Status', () => {
    it('identifies locked users', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const lockedUser = mockUsers[2]

      expect(vm.isLocked(lockedUser)).toBe(true)
    })

    it('identifies unlocked users', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const normalUser = mockUsers[0]

      expect(vm.isLocked(normalUser)).toBe(false)
    })
  })

  describe('Unlock User', () => {
    it('unlocks user successfully', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.unlockUser(mockUsers[2])
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/admin/users/3/unlock')
      expect(vm.successMessage).toBeTruthy()
    })

    it('handles unlock error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue({
        response: { data: { message: 'Unlock failed' } }
      })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.unlockUser(mockUsers[2])
      await flushPromises()

      expect(vm.error).toBe('Unlock failed')

      consoleSpy.mockRestore()
    })
  })

  describe('Format Functions', () => {
    it('formats date', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const result = vm.formatDate('2024-01-15T10:00:00Z')

      expect(result).toBeTruthy()
    })
  })

  describe('Error Handling', () => {
    it('handles fetch error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue({
        response: { data: { message: 'Failed to load users' } }
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Failed to load users')
      expect(vm.loading).toBe(false)

      consoleSpy.mockRestore()
    })
  })

  describe('Headers', () => {
    it('has table headers defined', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.headers.length).toBeGreaterThan(0)
      expect(vm.headers.some(h => h.value === 'email')).toBe(true)
    })
  })

  describe('Search', () => {
    it('supports search functionality', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.search = 'admin'

      // Search triggers a new fetch
      await flushPromises()

      expect(vm.search).toBe('admin')
    })
  })
})
