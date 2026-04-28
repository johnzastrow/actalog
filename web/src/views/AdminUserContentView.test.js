import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminUserContentView from './AdminUserContentView.vue'

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

const mockWods = [
  { id: 1, name: 'Fran', created_by_email: 'a@example.com', is_standard: false },
  { id: 2, name: 'Annie', created_by_email: 'b@example.com', is_standard: false },
]
const mockMovements = [{ id: 1, name: 'Deadlift', created_by_email: 'a@example.com' }]
const mockWorkouts = [{ id: 1, name: 'Monday WOD', created_by_email: 'a@example.com' }]

const createWrapper = () => {
  const pinia = createPinia()
  setActivePinia(pinia)
  return shallowMount(AdminUserContentView, {
    global: {
      plugins: [pinia],
      stubs: {
        'v-container': { template: '<div><slot /></div>' },
        'v-tabs': { template: '<div><slot /></div>' },
        'v-tab': { template: '<div><slot /></div>' },
        'v-window': { template: '<div><slot /></div>' },
        'v-window-item': { template: '<div><slot /></div>' },
        'v-card': { template: '<div><slot /></div>' },
        'v-card-title': { template: '<div><slot /></div>' },
        'v-card-text': { template: '<div><slot /></div>' },
        'v-card-actions': { template: '<div><slot /></div>' },
        'v-btn': { template: '<button @click="$attrs.onClick"><slot /></button>' },
        'v-icon': { template: '<i />' },
        'v-data-table': { template: '<div />' },
        'v-text-field': { template: '<input />' },
        'v-select': { template: '<select />' },
        'v-row': { template: '<div><slot /></div>' },
        'v-col': { template: '<div><slot /></div>' },
        'v-alert': { template: '<div><slot /></div>' },
        'v-progress-linear': { template: '<div />' },
        'v-dialog': { template: '<div><slot /></div>' },
        'v-divider': { template: '<hr />' },
        'v-chip': { template: '<span><slot /></span>' },
        'v-spacer': { template: '<div />' },
      },
    },
  })
}

describe('AdminUserContentView', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
    // View reads response.data.data (not .wods/.movements/.workouts) and response.data.count
    axios.get.mockImplementation((url) => {
      if (url.includes('wods')) return Promise.resolve({ data: { data: mockWods, count: 2 } })
      if (url.includes('movements')) return Promise.resolve({ data: { data: mockMovements, count: 1 } })
      if (url.includes('workouts')) return Promise.resolve({ data: { data: mockWorkouts, count: 1 } })
      return Promise.resolve({ data: {} })
    })
    axios.post.mockResolvedValue({ data: {} })
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
  })

  it('fetches user-created WODs on mount', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith(expect.stringContaining('/api/admin/user-created/wods'))
  })

  it('stores fetched WODs', async () => {
    wrapper = createWrapper()
    await flushPromises()

    // View uses wodCount (not wods.length) for display; wods array stores items
    expect(wrapper.vm.wods).toHaveLength(2)
    expect(wrapper.vm.wodCount).toBe(2)
  })

  it('sets error state on fetch failure', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    axios.get.mockRejectedValue(new Error('Server error'))

    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.error).toBeTruthy()
    consoleSpy.mockRestore()
  })

  it('fetches movements when loadMovements is called', async () => {
    wrapper = createWrapper()
    await flushPromises()

    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: { data: mockMovements, count: 1 } })

    await wrapper.vm.loadMovements()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith(expect.stringContaining('/api/admin/user-created/movements'))
    expect(wrapper.vm.movements).toHaveLength(1)
  })

  it('fetches workouts when loadWorkouts is called', async () => {
    wrapper = createWrapper()
    await flushPromises()

    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: { data: mockWorkouts, count: 1 } })

    await wrapper.vm.loadWorkouts()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith(expect.stringContaining('/api/admin/user-created/workouts'))
    expect(wrapper.vm.workouts).toHaveLength(1)
  })

  it('opens copy dialog for a WOD', async () => {
    wrapper = createWrapper()
    await flushPromises()

    // openCopyDialog(type, item) — type arg comes first
    wrapper.vm.openCopyDialog('wod', mockWods[0])

    expect(wrapper.vm.copyDialog).toBe(true)
    expect(wrapper.vm.selectedItem).toEqual(mockWods[0])
    expect(wrapper.vm.copyType).toBe('wod')
  })

  it('copies item to standard via confirmCopy', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.selectedItem = mockWods[0]
    wrapper.vm.copyType = 'wod'
    wrapper.vm.newName = 'Fran (Standard)'

    // actual function is confirmCopy, not copyToStandard
    await wrapper.vm.confirmCopy()
    await flushPromises()

    expect(axios.post).toHaveBeenCalledWith(
      '/api/admin/user-created/wods/1/copy-to-standard',
      expect.objectContaining({ new_name: 'Fran (Standard)' })
    )
  })
})
