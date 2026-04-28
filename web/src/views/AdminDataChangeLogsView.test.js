import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminDataChangeLogsView from './AdminDataChangeLogsView.vue'

vi.mock('@/utils/axios', () => ({
  default: { get: vi.fn() },
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
  { id: 1, entity_type: 'workout', operation: 'INSERT', user_email: 'a@example.com', created_at: new Date().toISOString() },
  { id: 2, entity_type: 'movement', operation: 'UPDATE', user_email: 'b@example.com', created_at: new Date().toISOString() },
]

const createWrapper = () => {
  const pinia = createPinia()
  setActivePinia(pinia)
  return shallowMount(AdminDataChangeLogsView, {
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

describe('AdminDataChangeLogsView', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: { logs: mockLogs, total: 2 } })
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
  })

  it('fetches data change logs on mount', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith(expect.stringContaining('/api/admin/data-change-logs'))
  })

  it('stores fetched logs', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.logs).toHaveLength(2)
    expect(wrapper.vm.total).toBe(2)
  })

  it('sets error on fetch failure', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    axios.get.mockRejectedValue(new Error('Server error'))

    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.error).toBeTruthy()
    consoleSpy.mockRestore()
  })

  it('paginates forward and backward', async () => {
    // total must exceed limit (50) for nextPage to advance
    axios.get.mockResolvedValue({ data: { logs: mockLogs, total: 200 } })

    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.nextPage()
    await flushPromises()
    expect(wrapper.vm.offset).toBeGreaterThan(0)

    wrapper.vm.previousPage()
    await flushPromises()
    expect(wrapper.vm.offset).toBe(0)
  })

  it('opens details dialog with selected log', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.showDetails(mockLogs[0])

    expect(wrapper.vm.detailsDialog).toBe(true)
    expect(wrapper.vm.selectedLog).toEqual(mockLogs[0])
  })

  it('re-fetches when a filter changes', async () => {
    wrapper = createWrapper()
    await flushPromises()

    const callsBefore = axios.get.mock.calls.length
    wrapper.vm.loadLogs()
    await flushPromises()

    expect(axios.get.mock.calls.length).toBeGreaterThan(callsBefore)
  })
})
