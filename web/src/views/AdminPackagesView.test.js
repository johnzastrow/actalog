import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminPackagesView from './AdminPackagesView.vue'

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

const mockOrgs = [{ id: 1, name: 'CrossFit Alpha' }, { id: 2, name: 'CrossFit Beta' }]
const mockPackages = [
  { id: 10, name: 'Monthly', credits: 20, price: 99.0, is_active: true },
  { id: 11, name: 'Drop-in', credits: 1, price: 15.0, is_active: true },
]
const mockDocuments = [{ id: 20, title: 'Membership Agreement', is_active: true }]

const createWrapper = () => {
  const pinia = createPinia()
  setActivePinia(pinia)
  return shallowMount(AdminPackagesView, {
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
        'v-textarea': { template: '<textarea />' },
        'v-select': { template: '<select />' },
        'v-autocomplete': { template: '<input />' },
        'v-checkbox': { template: '<input type="checkbox" />' },
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

describe('AdminPackagesView', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
    axios.get.mockImplementation((url) => {
      if (url === '/api/admin/organizations')
        return Promise.resolve({ data: { organizations: mockOrgs, total: 2 } })
      if (url.includes('/packages'))
        return Promise.resolve({ data: { packages: mockPackages } })
      if (url.includes('/documents'))
        return Promise.resolve({ data: mockDocuments })
      return Promise.resolve({ data: {} })
    })
    axios.post.mockResolvedValue({ data: { id: 99, name: 'New Package' } })
    axios.put.mockResolvedValue({ data: {} })
    axios.delete.mockResolvedValue({})
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
  })

  it('fetches organizations on mount', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith('/api/admin/organizations')
    expect(wrapper.vm.organizations).toHaveLength(2)
  })

  it('starts on packages tab', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.activeTab).toBe('packages')
  })

  it('fetches packages when an org is selected', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.selectedOrgId = 1
    await wrapper.vm.fetchData()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith('/api/gyms/1/packages?include_inactive=true')
    expect(wrapper.vm.packages).toHaveLength(2)
  })

  it('opens create package dialog', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.openPackageDialog()

    expect(wrapper.vm.packageDialog).toBe(true)
    expect(wrapper.vm.editingPackage).toBeNull()
  })

  it('opens edit package dialog with pre-filled data', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.openPackageDialog(mockPackages[0])

    expect(wrapper.vm.packageDialog).toBe(true)
    expect(wrapper.vm.editingPackage).toEqual(mockPackages[0])
    expect(wrapper.vm.packageForm.name).toBe('Monthly')
  })

  it('creates a new package via POST when no editingPackage', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.selectedOrgId = 1
    wrapper.vm.editingPackage = null
    wrapper.vm.packageForm = { name: 'New', description: '', credits: 10, price: 50, validity_days: 30, is_active: true }

    axios.get.mockResolvedValue({ data: { packages: mockPackages } })
    await wrapper.vm.savePackage()
    await flushPromises()

    expect(axios.post).toHaveBeenCalledWith(
      '/api/admin/gyms/1/packages',
      expect.objectContaining({ name: 'New' })
    )
  })

  it('updates a package via PUT when editingPackage is set', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.selectedOrgId = 1
    wrapper.vm.editingPackage = mockPackages[0]
    wrapper.vm.packageForm = { ...mockPackages[0], name: 'Updated Monthly' }

    axios.get.mockResolvedValue({ data: { packages: mockPackages } })
    await wrapper.vm.savePackage()
    await flushPromises()

    expect(axios.put).toHaveBeenCalledWith(
      '/api/admin/gyms/1/packages/10',
      expect.objectContaining({ name: 'Updated Monthly' })
    )
  })

  it('deletes a package via DELETE', async () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.selectedOrgId = 1
    axios.get.mockResolvedValue({ data: { packages: [mockPackages[1]] } })

    await wrapper.vm.deletePackage(mockPackages[0])
    await flushPromises()

    expect(axios.delete).toHaveBeenCalledWith('/api/admin/gyms/1/packages/10')
    confirmSpy.mockRestore()
  })
})
