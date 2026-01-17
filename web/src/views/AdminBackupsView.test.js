import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminBackupsView from './AdminBackupsView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  }
}))

import axios from '@/utils/axios'

describe('AdminBackupsView', () => {
  let wrapper

  const mockBackups = [
    {
      id: 1,
      filename: 'backup_2024-01-15.zip',
      file_size: 1024000,
      created_at: '2024-01-15T10:00:00Z',
      created_by_email: 'admin@test.com',
      total_users: 50,
      total_workouts: 200,
      version: '1.0.0'
    },
    {
      id: 2,
      filename: 'backup_2024-01-10.zip',
      file_size: 512000,
      created_at: '2024-01-10T08:00:00Z',
      created_by_email: 'admin@test.com',
      total_users: 45,
      total_workouts: 180,
      version: '1.0.0'
    }
  ]

  const setupDefaultMocks = () => {
    axios.get.mockResolvedValue({ data: { backups: mockBackups } })
    axios.post.mockResolvedValue({ data: { message: 'Backup created' } })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(AdminBackupsView, {
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
    it('fetches backups on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/admin/backups')
    })

    it('stores fetched backups', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.backups).toHaveLength(2)
    })

    it('sets loading to false after fetch', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)
    })
  })

  describe('Create Backup', () => {
    it('creates backup successfully', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.createBackup()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/admin/backups')
      expect(vm.successMessage).toBeTruthy()
    })

    it('refreshes backup list after create', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      axios.get.mockClear()

      await vm.createBackup()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/admin/backups')
    })

    it('handles create error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockResolvedValue({ data: { backups: [] } })
      axios.post.mockRejectedValue({
        response: { data: { error: 'Backup failed' } }
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.createBackup()
      await flushPromises()

      expect(vm.error).toBe('Backup failed')

      consoleSpy.mockRestore()
    })

    it('sets creating state during creation', async () => {
      let resolvePost
      axios.get.mockResolvedValue({ data: { backups: [] } })
      axios.post.mockReturnValue(new Promise(resolve => { resolvePost = resolve }))
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const createPromise = vm.createBackup()

      expect(vm.creating).toBe(true)

      resolvePost({ data: { message: 'Created' } })
      await createPromise
      await flushPromises()

      expect(vm.creating).toBe(false)
    })
  })

  describe('Format Functions', () => {
    it('formats file size', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.formatFileSize(1024)).toBe('1 KB')
      expect(vm.formatFileSize(1048576)).toBe('1 MB')
    })

    it('formats date', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const result = vm.formatDate('2024-01-15T10:00:00Z')

      expect(result).toBeTruthy()
    })

    it('formats time', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const result = vm.formatTime('2024-01-15T10:00:00Z')

      expect(result).toBeTruthy()
    })
  })

  describe('Error Handling', () => {
    it('handles fetch error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue({
        response: { data: { error: 'Failed to load backups' } }
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Failed to load backups')
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
      expect(vm.headers.some(h => h.key === 'filename')).toBe(true)
    })
  })
})
