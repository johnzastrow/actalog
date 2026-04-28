import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminEmailLogsView from './AdminEmailLogsView.vue'

vi.mock('@/utils/axios', () => ({
  default: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))
vi.mock('@/components/AdminHeader.vue', () => ({
  default: { template: '<div />', props: ['title', 'subtitle', 'breadcrumbs'] },
}))
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return { ...actual, useRouter: vi.fn(() => ({ push: vi.fn() })) }
})

import axios from '@/utils/axios'

const mockLogs = [
  { id: 1, recipient: 'a@example.com', subject: 'Welcome', success: true, created_at: new Date().toISOString() },
  { id: 2, recipient: 'b@example.com', subject: 'Reset', success: false, created_at: new Date().toISOString() },
]

const createWrapper = () => {
  const pinia = createPinia()
  setActivePinia(pinia)
  return shallowMount(AdminEmailLogsView, {
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
        'v-checkbox': { template: '<input type="checkbox" />' },
        'v-list': { template: '<ul><slot /></ul>' },
        'v-list-item': { template: '<li><slot /></li>' },
      },
    },
  })
}

describe('AdminEmailLogsView', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
    // EmailLogsView reads response.data.count for totalCount, not response.data.total
    axios.get.mockResolvedValue({ data: { logs: mockLogs, count: 2 } })
    axios.post.mockResolvedValue({ data: { deleted: 10 } })
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
  })

  it('fetches email logs on mount', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith(expect.stringContaining('/api/admin/email-logs'))
  })

  it('stores fetched logs in state', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.logs).toHaveLength(2)
    // view reads response.data.count for totalCount
    expect(wrapper.vm.totalCount).toBe(2)
  })

  it('sets error on fetch failure', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    axios.get.mockRejectedValue(new Error('Server error'))

    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.error).toBeTruthy()
    consoleSpy.mockRestore()
  })

  it('calls cleanup API and reloads on confirmCleanup', async () => {
    wrapper = createWrapper()
    await flushPromises()

    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: { logs: [], total: 0 } })
    axios.post.mockResolvedValue({ data: { deleted: 10 } })

    await wrapper.vm.confirmCleanup()
    await flushPromises()

    expect(axios.post).toHaveBeenCalledWith('/api/admin/email-logs/cleanup')
    expect(wrapper.vm.successMessage).toBeTruthy()
    expect(wrapper.vm.cleanupDialog).toBe(false)
  })

  it('shows details dialog when a log is clicked', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.showLogDetails(mockLogs[0])

    expect(wrapper.vm.detailsDialog).toBe(true)
    expect(wrapper.vm.selectedLog).toEqual(mockLogs[0])
  })
})
