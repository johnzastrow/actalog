import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminAnnouncementsView from './AdminAnnouncementsView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  }
}))

// Mock MarkdownRenderer
vi.mock('@/components/MarkdownRenderer.vue', () => ({
  default: {
    template: '<div class="markdown-renderer"></div>',
    props: ['content']
  }
}))

import axios from '@/utils/axios'

describe('AdminAnnouncementsView', () => {
  let wrapper

  const mockOrganizations = [
    { id: 1, name: 'Org 1' },
    { id: 2, name: 'Org 2' }
  ]

  const mockUsers = [
    { id: 1, email: 'user1@test.com' },
    { id: 2, email: 'user2@test.com' }
  ]

  const setupDefaultMocks = () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/admin/organizations') {
        return Promise.resolve({ data: { organizations: mockOrganizations } })
      }
      if (url === '/api/admin/users') {
        return Promise.resolve({ data: { users: mockUsers } })
      }
      return Promise.resolve({ data: {} })
    })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(AdminAnnouncementsView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-card-title': { template: '<div><slot /></div>' },
          'v-card-text': { template: '<div><slot /></div>' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-icon': { template: '<i></i>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-form': { template: '<form @submit.prevent><slot /></form>' },
          'v-text-field': { template: '<input />' },
          'v-textarea': { template: '<textarea></textarea>' },
          'v-select': { template: '<select></select>' },
          'v-autocomplete': { template: '<select></select>' },
          'v-divider': { template: '<hr />' },
          'v-tabs': { template: '<div><slot /></div>' },
          'v-tab': { template: '<div><slot /></div>' },
          'v-window': { template: '<div><slot /></div>' },
          'v-window-item': { template: '<div><slot /></div>' },
          'markdown-renderer': true
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
    it('loads organizations and users on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/admin/organizations')
      expect(axios.get).toHaveBeenCalledWith('/api/admin/users')
    })

    it('stores loaded data', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.organizations).toHaveLength(2)
      expect(vm.users).toHaveLength(2)
    })

    it('initializes with default announcement', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.announcement.title).toBe('')
      expect(vm.announcement.message).toBe('')
      expect(vm.announcement.targetType).toBe('all')
    })
  })

  describe('Target Options', () => {
    it('has all target type options', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.targetOptions).toHaveLength(3)
      expect(vm.targetOptions.find(t => t.value === 'all')).toBeTruthy()
      expect(vm.targetOptions.find(t => t.value === 'organization')).toBeTruthy()
      expect(vm.targetOptions.find(t => t.value === 'users')).toBeTruthy()
    })
  })

  describe('Send Announcement', () => {
    it('sends announcement successfully', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.announcement.title = 'Test Title'
      vm.announcement.message = 'Test Message'
      vm.announcement.targetType = 'all'
      // Mock form validation
      vm.form = { validate: vi.fn().mockResolvedValue({ valid: true }), reset: vi.fn() }

      await vm.sendAnnouncement()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/admin/notifications/announce', expect.objectContaining({
        title: 'Test Title',
        message: 'Test Message',
        target_type: 'all'
      }))
      expect(vm.successMessage).toContain('successfully')
    })

    it('resets form after successful send', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.announcement.title = 'Test Title'
      vm.announcement.message = 'Test Message'
      const mockReset = vi.fn()
      vm.form = { validate: vi.fn().mockResolvedValue({ valid: true }), reset: mockReset }

      await vm.sendAnnouncement()
      await flushPromises()

      expect(vm.announcement.title).toBe('')
      expect(mockReset).toHaveBeenCalled()
    })

    it('handles send error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue({
        response: { data: { error: 'Send failed' } }
      })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.announcement.title = 'Test'
      vm.announcement.message = 'Test'
      vm.form = { validate: vi.fn().mockResolvedValue({ valid: true }) }

      await vm.sendAnnouncement()
      await flushPromises()

      expect(vm.errorMessage).toBe('Send failed')

      consoleSpy.mockRestore()
    })

    it('does not send if validation fails', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.form = { validate: vi.fn().mockResolvedValue({ valid: false }) }

      await vm.sendAnnouncement()

      expect(axios.post).not.toHaveBeenCalled()
    })

    it('sets sending state during submission', async () => {
      let resolvePost
      axios.post.mockReturnValue(new Promise(resolve => { resolvePost = resolve }))
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.announcement.title = 'Test'
      vm.announcement.message = 'Test'
      vm.form = { validate: vi.fn().mockResolvedValue({ valid: true }), reset: vi.fn() }

      const sendPromise = vm.sendAnnouncement()
      // Allow validation promise to resolve
      await Promise.resolve()

      expect(vm.sending).toBe(true)

      resolvePost({ data: {} })
      await sendPromise
      await flushPromises()

      expect(vm.sending).toBe(false)
    })
  })

  describe('Error Handling', () => {
    it('handles organization load error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockImplementation((url) => {
        if (url === '/api/admin/organizations') {
          return Promise.reject(new Error('Failed'))
        }
        return Promise.resolve({ data: { users: [] } })
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.errorMessage).toContain('organizations')

      consoleSpy.mockRestore()
    })

    it('handles users load error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockImplementation((url) => {
        if (url === '/api/admin/users') {
          return Promise.reject(new Error('Failed'))
        }
        return Promise.resolve({ data: { organizations: [] } })
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.errorMessage).toContain('users')

      consoleSpy.mockRestore()
    })
  })
})
