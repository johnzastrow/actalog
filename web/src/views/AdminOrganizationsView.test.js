import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminOrganizationsView from './AdminOrganizationsView.vue'

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

const mockOrgs = [
  { id: 1, name: 'CrossFit Alpha', description: 'First gym', member_count: 10 },
  { id: 2, name: 'CrossFit Beta', description: 'Second gym', member_count: 5 },
]

const createWrapper = () => {
  const pinia = createPinia()
  setActivePinia(pinia)
  return shallowMount(AdminOrganizationsView, {
    global: {
      plugins: [pinia],
      stubs: {
        'v-container': { template: '<div><slot /></div>' },
        'v-card': { template: '<div><slot /></div>' },
        'v-card-title': { template: '<div><slot /></div>' },
        'v-card-text': { template: '<div><slot /></div>' },
        'v-card-actions': { template: '<div><slot /></div>' },
        'v-btn': { template: '<button @click="$attrs.onClick"><slot /></button>' },
        'v-icon': { template: '<i />' },
        'v-data-table': { template: '<div />' },
        'v-text-field': { template: '<input />' },
        'v-textarea': { template: '<textarea />' },
        'v-row': { template: '<div><slot /></div>' },
        'v-col': { template: '<div><slot /></div>' },
        'v-alert': { template: '<div><slot /></div>' },
        'v-progress-linear': { template: '<div />' },
        'v-dialog': { template: '<div><slot /></div>' },
        'v-divider': { template: '<hr />' },
        'v-chip': { template: '<span><slot /></span>' },
        'v-spacer': { template: '<div />' },
        'v-list': { template: '<ul><slot /></ul>' },
        'v-list-item': { template: '<li><slot /></li>' },
        'v-list-item-title': { template: '<span><slot /></span>' },
        'v-list-item-subtitle': { template: '<span><slot /></span>' },
      },
    },
  })
}

describe('AdminOrganizationsView', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: { organizations: mockOrgs, total: 2 } })
    axios.post.mockResolvedValue({ data: { id: 3, name: 'New Gym' } })
    axios.delete.mockResolvedValue({})
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
  })

  it('fetches organizations on mount', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(axios.get).toHaveBeenCalledWith(expect.stringContaining('/api/admin/organizations'))
  })

  it('stores fetched organizations in state', async () => {
    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.organizations).toHaveLength(2)
    expect(wrapper.vm.total).toBe(2)
  })

  it('sets error state on fetch failure', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    axios.get.mockRejectedValue(new Error('Server error'))

    wrapper = createWrapper()
    await flushPromises()

    expect(wrapper.vm.error).toBeTruthy()
    consoleSpy.mockRestore()
  })

  it('opens dialog in create mode when openCreateDialog is called', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.openCreateDialog()

    expect(wrapper.vm.dialog).toBe(true)
    expect(wrapper.vm.formData.name).toBe('')
  })

  it('creates a new organization via POST', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.openCreateDialog()
    wrapper.vm.formData.name = 'New Gym'
    wrapper.vm.formData.description = 'A new gym'

    axios.get.mockResolvedValue({ data: { organizations: [...mockOrgs, { id: 3, name: 'New Gym' }], total: 3 } })
    await wrapper.vm.saveOrganization()
    await flushPromises()

    expect(axios.post).toHaveBeenCalledWith('/api/admin/organizations', expect.objectContaining({ name: 'New Gym' }))
    expect(wrapper.vm.dialog).toBe(false)
  })

  it('opens delete confirmation dialog via deleteOrganization', async () => {
    wrapper = createWrapper()
    await flushPromises()

    // deleteOrganization(org) sets organizationToDelete and opens dialog
    wrapper.vm.deleteOrganization(mockOrgs[0])

    expect(wrapper.vm.deleteDialog).toBe(true)
    expect(wrapper.vm.organizationToDelete).toEqual(mockOrgs[0])
  })

  it('deletes organization via confirmDelete and reloads', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.organizationToDelete = mockOrgs[0]
    axios.get.mockResolvedValue({ data: { organizations: [mockOrgs[1]], total: 1 } })

    // confirmDelete() performs the actual API call
    await wrapper.vm.confirmDelete()
    await flushPromises()

    expect(axios.delete).toHaveBeenCalledWith('/api/admin/organizations/1')
    expect(wrapper.vm.deleteDialog).toBe(false)
  })
})
