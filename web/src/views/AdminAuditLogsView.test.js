import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminAuditLogsView from './AdminAuditLogsView.vue'

vi.mock('@/utils/axios', () => ({
  default: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))
vi.mock('@/components/AdminHeader.vue', () => ({
  default: { template: '<div><slot /></div>', props: ['title', 'subtitle', 'breadcrumbs'] },
}))
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return { ...actual, useRouter: vi.fn(() => ({ push: vi.fn() })) }
})

import axios from '@/utils/axios'

const mockLogs = [
  { id: 1, event_type: 'login', user_email: 'a@example.com', ip_address: '127.0.0.1', created_at: new Date().toISOString() },
  { id: 2, event_type: 'logout', user_email: 'b@example.com', ip_address: '127.0.0.1', created_at: new Date().toISOString() },
]

const createWrapper = () => {
  const pinia = createPinia()
  setActivePinia(pinia)
  return shallowMount(AdminAuditLogsView, {
    global: {
      plugins: [pinia],
      stubs: {
        'v-container': { template: '<div><slot /></div>' },
        'v-card': { template: '<div><slot /></div>' },
        'v-card-title': { template: '<div><slot /></div>' },
        'v-card-text': { template: '<div><slot /></div>' },
        'v-btn': { template: '<button @click="$attrs.onClick"><slot /></button>' },
        'v-icon': { template: '<i />' },
        'v-data-table': { template: '<div />' },
        'v-select': { template: '<select />' },
        'v-text-field': { template: '<input />' },
        'v-row': { template: '<div><slot /></div>' },
        'v-col': { template: '<div><slot /></div>' },
        'v-alert': { template: '<div><slot /></div>' },
        'v-progress-linear': { template: '<div />' },
        'v-dialog': { template: '<div><slot /></div>' },
        'v-divider': { template: '<hr />' },
        'v-chip': { template: '<span><slot /></span>' },
      },
    },
  })
}

describe('AdminAuditLogsView', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: { logs: mockLogs, total: 2 } })
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
  })

  it('fetches audit logs on mount', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith(expect.stringContaining('/api/admin/audit-logs'))
  })

  it('stores fetched logs in state', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.logs).toHaveLength(2)
    expect(wrapper.vm.total).toBe(2)
  })

  it('starts in loading state, then clears it', async () => {
    let resolveGet
    axios.get.mockReturnValue(new Promise(r => { resolveGet = r }))

    wrapper = createWrapper()
    expect(wrapper.vm.loading).toBe(true)

    resolveGet({ data: { logs: [], total: 0 } })
    await flushPromises()

    expect(wrapper.vm.loading).toBe(false)
  })

  it('sets error state on fetch failure', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    axios.get.mockRejectedValue(new Error('Server error'))

    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.error).toBeTruthy()
    consoleSpy.mockRestore()
  })

  it('advances to next page when nextPage is called', async () => {
    // total must exceed limit (50) for nextPage to advance offset
    axios.get.mockResolvedValue({ data: { logs: mockLogs, total: 200 } })

    wrapper = createWrapper()
    await flushPromises()

    const initialOffset = wrapper.vm.offset
    wrapper.vm.nextPage()
    await flushPromises()

    expect(wrapper.vm.offset).toBeGreaterThan(initialOffset)
    expect(axios.get).toHaveBeenCalledTimes(2)
  })

  it('goes back to previous page when previousPage is called', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.offset = 50
    wrapper.vm.previousPage()
    await flushPromises()

    expect(wrapper.vm.offset).toBe(0)
  })

  it('opens details dialog when showDetails is called', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.showDetails(mockLogs[0])

    expect(wrapper.vm.detailsDialog).toBe(true)
    expect(wrapper.vm.selectedLog).toEqual(mockLogs[0])
  })

  it('formats event types for display', async () => {
    wrapper = createWrapper()
    await flushPromises()

    // login → Login, user_created → User Created
    expect(wrapper.vm.formatEventType('login')).toBe('Login')
    expect(wrapper.vm.formatEventType('user_created')).toContain('User')
  })
})
