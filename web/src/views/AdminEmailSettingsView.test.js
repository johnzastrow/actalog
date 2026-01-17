import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminEmailSettingsView from './AdminEmailSettingsView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  }
}))

import axios from '@/utils/axios'

describe('AdminEmailSettingsView', () => {
  let wrapper

  const mockConfig = {
    smtp_host: 'mail.example.com',
    smtp_port: 587,
    smtp_user: 'user@example.com',
    from_address: 'noreply@example.com',
    from_name: 'Test App',
    tls_mode: 'STARTTLS',
    enabled: true
  }

  const mockDebugInfo = {
    connection_host: 'mail.example.com',
    connection_port: 587,
    connection_tls: 'STARTTLS',
    connection_time: '120ms',
    connection_error: null,
    auth_method: 'PLAIN',
    auth_time: '50ms',
    auth_error: null,
    send_time: '200ms',
    send_error: null,
    total_duration: '370ms',
    final_error: null,
    smtp_responses: ['250 OK']
  }

  const setupDefaultMocks = () => {
    axios.get.mockResolvedValue({ data: mockConfig })
    axios.post.mockResolvedValue({
      data: {
        success: true,
        message: 'Test email sent',
        debug_info: mockDebugInfo
      }
    })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(AdminEmailSettingsView, {
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
          'v-alert': { template: '<div><slot /></div>' },
          'v-form': { template: '<form @submit.prevent><slot /></form>' },
          'v-text-field': { template: '<input />' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-divider': { template: '<hr />' },
          'v-spacer': { template: '<div></div>' },
          'v-progress-linear': { template: '<div></div>' },
          'v-sheet': { template: '<div><slot /></div>' }
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
    it('fetches config on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/admin/email/config')
    })

    it('stores fetched config', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.config.smtp_host).toBe('mail.example.com')
      expect(vm.config.smtp_port).toBe(587)
      expect(vm.config.enabled).toBe(true)
    })

    it('sets loading to false after fetch', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)
    })

    it('initializes testEmail as empty', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.testEmail).toBe('')
    })
  })

  describe('Send Test Email', () => {
    it('sends test email successfully', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.testEmail = 'test@example.com'

      await vm.sendTestEmail()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/admin/email/test', {
        recipient_email: 'test@example.com'
      })
      expect(vm.successMessage).toContain('sent')
    })

    it('displays debug info after send', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.testEmail = 'test@example.com'

      await vm.sendTestEmail()
      await flushPromises()

      expect(vm.debugInfo).not.toBe(null)
      expect(vm.debugInfo.connection_host).toBe('mail.example.com')
    })

    it('does not send if no email', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.testEmail = ''

      await vm.sendTestEmail()

      expect(axios.post).not.toHaveBeenCalled()
    })

    it('handles send failure response', async () => {
      axios.get.mockResolvedValue({ data: mockConfig })
      axios.post.mockResolvedValue({
        data: {
          success: false,
          message: 'Connection failed',
          debug_info: { ...mockDebugInfo, final_error: 'Connection timeout' }
        }
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.testEmail = 'test@example.com'

      await vm.sendTestEmail()
      await flushPromises()

      expect(vm.error).toContain('Connection failed')
    })

    it('handles send error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockResolvedValue({ data: mockConfig })
      axios.post.mockRejectedValue({
        response: { data: { error: 'Network error' } }
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.testEmail = 'test@example.com'

      await vm.sendTestEmail()
      await flushPromises()

      expect(vm.error).toBe('Network error')

      consoleSpy.mockRestore()
    })

    it('sets sending state during send', async () => {
      let resolvePost
      axios.get.mockResolvedValue({ data: mockConfig })
      axios.post.mockReturnValue(new Promise(resolve => { resolvePost = resolve }))
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.testEmail = 'test@example.com'

      const sendPromise = vm.sendTestEmail()

      expect(vm.sending).toBe(true)

      resolvePost({ data: { success: true, debug_info: mockDebugInfo } })
      await sendPromise
      await flushPromises()

      expect(vm.sending).toBe(false)
    })
  })

  describe('Error Handling', () => {
    it('handles config fetch error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue({
        response: { data: { error: 'Failed to load config' } }
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Failed to load config')
      expect(vm.loading).toBe(false)

      consoleSpy.mockRestore()
    })
  })
})
