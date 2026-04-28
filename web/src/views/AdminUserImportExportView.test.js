import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminUserImportExportView from './AdminUserImportExportView.vue'

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

const mockUsers = [
  { id: 1, name: 'Alice', email: 'alice@example.com', role: 'athlete' },
  { id: 2, name: 'Bob', email: 'bob@example.com', role: 'athlete' },
]

const createWrapper = () => {
  const pinia = createPinia()
  setActivePinia(pinia)
  return shallowMount(AdminUserImportExportView, {
    global: {
      plugins: [pinia],
      stubs: {
        'v-container': { template: '<div><slot /></div>' },
        'v-tabs': { template: '<div><slot /></div>' },
        'v-tab': { template: '<div><slot /></div>' },
        'v-tabs-window': { template: '<div><slot /></div>' },
        'v-tabs-window-item': { template: '<div><slot /></div>' },
        'v-card': { template: '<div><slot /></div>' },
        'v-card-title': { template: '<div><slot /></div>' },
        'v-card-text': { template: '<div><slot /></div>' },
        'v-card-actions': { template: '<div><slot /></div>' },
        'v-btn': { template: '<button @click="$attrs.onClick"><slot /></button>' },
        'v-icon': { template: '<i />' },
        'v-data-table': { template: '<div />' },
        'v-text-field': { template: '<input />' },
        'v-checkbox': { template: '<input type="checkbox" />' },
        'v-row': { template: '<div><slot /></div>' },
        'v-col': { template: '<div><slot /></div>' },
        'v-alert': { template: '<div><slot /></div>' },
        'v-progress-linear': { template: '<div />' },
        'v-dialog': { template: '<div><slot /></div>' },
        'v-divider': { template: '<hr />' },
        'v-chip': { template: '<span><slot /></span>' },
        'v-spacer': { template: '<div />' },
        'v-select': { template: '<select />' },
      },
    },
  })
}

describe('AdminUserImportExportView', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: { users: mockUsers, total: 2 } })
    axios.post.mockResolvedValue({
      data: { preview: { users_to_create: 2, users_to_update: 0, errors: [] } },
    })
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
  })

  it('defaults to import tab', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.activeTab).toBe('import')
  })

  it('initializes without a selected file', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.selectedFile).toBeNull()
    expect(wrapper.vm.previewResult).toBeNull()
  })

  it('sets selectedFile when handleFileSelect is called', async () => {
    wrapper = createWrapper()
    await flushPromises()

    const mockFile = new File(['csv content'], 'users.csv', { type: 'text/csv' })
    wrapper.vm.handleFileSelect({ target: { files: [mockFile] } })

    expect(wrapper.vm.selectedFile).toEqual(mockFile)
  })

  it('previews import via POST', async () => {
    const mockFile = new File(['csv content'], 'users.csv', { type: 'text/csv' })
    axios.post.mockResolvedValue({
      data: { preview: { users_to_create: 2, users_to_update: 0, errors: [] } },
    })

    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.selectedFile = mockFile
    await wrapper.vm.previewImport()
    await flushPromises()

    expect(axios.post).toHaveBeenCalledWith(
      '/api/admin/user-management/import/preview',
      expect.any(FormData),
      expect.objectContaining({ headers: { 'Content-Type': 'multipart/form-data' } })
    )
    expect(wrapper.vm.previewResult).toBeDefined()
  })

  it('sets error when preview fails', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    axios.post.mockRejectedValue(new Error('Server error'))

    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.selectedFile = new File([''], 'users.csv')
    await wrapper.vm.previewImport()
    await flushPromises()

    expect(wrapper.vm.error).toBeTruthy()
    consoleSpy.mockRestore()
  })

  it('exports users via GET', async () => {
    axios.get.mockResolvedValue({
      data: 'name,email\nAlice,alice@example.com',
      headers: { 'content-type': 'text/csv' },
    })

    wrapper = createWrapper()
    await flushPromises()

    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: 'csv data', headers: { 'content-type': 'text/csv' } })

    await wrapper.vm.exportUsers()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith(
      '/api/admin/user-management/export',
      expect.objectContaining({ responseType: 'blob' })
    )
  })

  it('fetches filtered users for the filter tab', async () => {
    wrapper = createWrapper()
    await flushPromises()

    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: { users: mockUsers, total: 2 } })

    await wrapper.vm.loadUsers()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith(expect.stringContaining('/api/admin/user-management/filter'))
  })
})
