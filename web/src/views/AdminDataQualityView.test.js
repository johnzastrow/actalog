import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminDataQualityView from './AdminDataQualityView.vue'

vi.mock('@/utils/axios', () => ({
  default: { get: vi.fn(), post: vi.fn() },
}))
vi.mock('@/components/AdminHeader.vue', () => ({
  default: { template: '<div />', props: ['title', 'subtitle', 'breadcrumbs'] },
}))
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return { ...actual, useRouter: vi.fn(() => ({ push: vi.fn() })) }
})

import axios from '@/utils/axios'

// runFullScan expects: response.data.{ duplicates, issues, summary.total_duplicate_records, summary.total_data_quality_issues }
const mockScanResponse = {
  duplicates: { movements: { count: 3 }, wods: { count: 0 } },
  issues: { orphaned_workouts: 0, missing_wod_refs: 1 },
  summary: { total_duplicate_records: 3, total_data_quality_issues: 1 },
}

const createWrapper = () => {
  const pinia = createPinia()
  setActivePinia(pinia)
  return shallowMount(AdminDataQualityView, {
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
        'v-btn': { template: '<button @click="$attrs.onClick"><slot /></button>' },
        'v-icon': { template: '<i />' },
        'v-alert': { template: '<div><slot /></div>' },
        'v-chip': { template: '<span><slot /></span>' },
        'v-chip-group': { template: '<div><slot /></div>' },
        'v-radio-group': { template: '<div><slot /></div>' },
        'v-radio': { template: '<input type="radio" />' },
        'v-dialog': { template: '<div><slot /></div>' },
        'v-divider': { template: '<hr />' },
        'v-row': { template: '<div><slot /></div>' },
        'v-col': { template: '<div><slot /></div>' },
        'v-progress-linear': { template: '<div />' },
        'v-list': { template: '<ul><slot /></ul>' },
        'v-list-item': { template: '<li><slot /></li>' },
      },
    },
  })
}

describe('AdminDataQualityView', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: {} })
    axios.post.mockResolvedValue({ data: {} })
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
  })

  it('initializes with scanned = false', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.scanned).toBe(false)
  })

  it('starts on the overview tab', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.activeTab).toBe('overview')
  })

  it('fetches full scan summary and sets scanned = true', async () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/admin/data-quality/full-scan')
        return Promise.resolve({ data: mockScanResponse })
      return Promise.resolve({ data: {} })
    })

    wrapper = createWrapper()
    await flushPromises()

    await wrapper.vm.runFullScan()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith('/api/admin/data-quality/full-scan')
    expect(wrapper.vm.scanned).toBe(true)
  })

  it('sets scanning flag during scan', async () => {
    let resolve
    axios.get.mockReturnValue(new Promise(r => { resolve = r }))

    wrapper = createWrapper()
    const scanPromise = wrapper.vm.runFullScan()

    expect(wrapper.vm.scanning).toBe(true)

    resolve({ data: mockScanResponse })
    await scanPromise
    await flushPromises()

    expect(wrapper.vm.scanning).toBe(false)
  })

  it('resets scanning to false and does not set scanned on failure', async () => {
    // The view calls alert() on error, not error.value
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
    axios.get.mockRejectedValue({ response: { data: { error: 'Server error' } } })

    wrapper = createWrapper()
    await wrapper.vm.runFullScan()
    await flushPromises()

    expect(wrapper.vm.scanned).toBe(false)
    expect(wrapper.vm.scanning).toBe(false)
    alertSpy.mockRestore()
  })

  it('fetches entity-level duplicates for selected entity type', async () => {
    axios.get.mockResolvedValue({ data: { groups: [] } })

    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.selectedEntityType = 'movements'
    // actual function name is scanSelectedEntity
    await wrapper.vm.scanSelectedEntity()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith('/api/admin/data-quality/duplicates/movements')
  })
})
