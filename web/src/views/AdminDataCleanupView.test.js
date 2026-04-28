import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminDataCleanupView from './AdminDataCleanupView.vue'

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

const mockMismatches = [
  {
    id: 1,
    wod_name: 'Fran',
    user_email: 'a@example.com',
    workout_date: '2026-01-01',
    expected_score_type: 'time',
    actual_score_type: 'reps',
  },
]

const createWrapper = () => {
  const pinia = createPinia()
  setActivePinia(pinia)
  return shallowMount(AdminDataCleanupView, {
    global: {
      plugins: [pinia],
      stubs: {
        'v-container': { template: '<div><slot /></div>' },
        'v-card': { template: '<div><slot /></div>' },
        'v-btn': { template: '<button @click="$attrs.onClick"><slot /></button>' },
        'v-icon': { template: '<i />' },
        'v-alert': { template: '<div><slot /></div>' },
        'v-chip': { template: '<span><slot /></span>' },
        'v-dialog': { template: '<div><slot /></div>' },
        'v-spacer': { template: '<div />' },
        'v-text-field': { template: '<input />' },
        'v-select': { template: '<select />' },
        'v-divider': { template: '<hr />' },
      },
    },
  })
}

describe('AdminDataCleanupView', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: { mismatches: mockMismatches, total: 1 } })
    axios.delete.mockResolvedValue({ data: { fixed: 1 } })
    axios.put.mockResolvedValue({ data: {} })
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

  it('scans for mismatches and sets scanned = true on success', async () => {
    wrapper = createWrapper()
    await flushPromises()

    await wrapper.vm.scanForMismatches()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith('/api/admin/data-cleanup/wod-mismatches')
    expect(wrapper.vm.scanned).toBe(true)
    expect(wrapper.vm.mismatches).toHaveLength(1)
  })

  it('sets scanning flag during scan', async () => {
    let resolve
    axios.get.mockReturnValue(new Promise(r => { resolve = r }))

    wrapper = createWrapper()
    const scanPromise = wrapper.vm.scanForMismatches()

    expect(wrapper.vm.scanning).toBe(true)

    resolve({ data: { mismatches: [], total: 0 } })
    await scanPromise
    await flushPromises()

    expect(wrapper.vm.scanning).toBe(false)
  })

  it('sets scanning to false and logs error when scan fails', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    axios.get.mockRejectedValue(new Error('Server error'))

    wrapper = createWrapper()
    await wrapper.vm.scanForMismatches()
    await flushPromises()

    // No error ref in this view; scanning resets to false on failure
    expect(wrapper.vm.scanning).toBe(false)
    expect(consoleSpy).toHaveBeenCalled()
    consoleSpy.mockRestore()
  })

  it('fixes all mismatches via delete endpoint', async () => {
    wrapper = createWrapper()
    await flushPromises()

    // First scan to enable the fix button
    axios.get.mockResolvedValue({ data: { mismatches: mockMismatches, total: 1 } })
    await wrapper.vm.scanForMismatches()
    await flushPromises()

    // actual function is fixMismatches, not fixAllMismatches
    await wrapper.vm.fixMismatches()
    await flushPromises()

    expect(axios.delete).toHaveBeenCalledWith('/api/admin/data-cleanup/wod-mismatches')
  })

  it('opens edit dialog for a mismatch', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.openEditDialog(mockMismatches[0])

    expect(wrapper.vm.editDialog).toBe(true)
    expect(wrapper.vm.editingMismatch).toEqual(mockMismatches[0])
  })
})
