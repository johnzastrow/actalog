import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminView from './AdminView.vue'

// Mock vue-router
const mockPush = vi.fn()
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: mockPush,
    }))
  }
})

describe('AdminView', () => {
  let wrapper

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(AdminView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-card': { template: '<div @click="$attrs.onClick"><slot /></div>' },
          'v-icon': { template: '<i></i>' },
          'v-chip': { template: '<span><slot /></span>' },
          'v-divider': { template: '<hr />' },
          'v-bottom-navigation': { template: '<div><slot /></div>' },
          'v-btn': { template: '<button><slot /></button>' },
          AdminHeader: { template: '<div><h1>{{ title }}</h1><p>{{ subtitle }}</p></div>', props: ['title', 'subtitle', 'breadcrumbs'] }
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    mockPush.mockClear()
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
      wrapper = null
    }
  })

  // ==========================================================================
  // RENDER TESTS
  // ==========================================================================

  describe('Rendering', () => {
    it('renders admin page title', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Administration')
    })

    it('renders admin tool cards', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Data Cleanup')
      expect(wrapper.text()).toContain('Data Quality Checks')
      expect(wrapper.text()).toContain('User Management')
      expect(wrapper.text()).toContain('Gyms')
      expect(wrapper.text()).toContain('User Content')
      expect(wrapper.text()).toContain('Database Backups')
      expect(wrapper.text()).toContain('Audit Logs')
      expect(wrapper.text()).toContain('Data Change Logs')
      expect(wrapper.text()).toContain('Announcements')
      expect(wrapper.text()).toContain('Subscriptions')
      expect(wrapper.text()).toContain('Email Settings')
      expect(wrapper.text()).toContain('System Metrics')
      expect(wrapper.text()).toContain('User Import/Export')
    })
  })

  // ==========================================================================
  // NAVIGATION TESTS
  // ==========================================================================

  describe('Navigation', () => {
    it('navigates to data cleanup', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.navigateTo('/admin/data-cleanup')

      expect(mockPush).toHaveBeenCalledWith('/admin/data-cleanup')
    })

    it('navigates to various admin routes', () => {
      createWrapper()

      const vm = wrapper.vm

      vm.navigateTo('/admin/users')
      expect(mockPush).toHaveBeenCalledWith('/admin/users')

      vm.navigateTo('/admin/backups')
      expect(mockPush).toHaveBeenCalledWith('/admin/backups')

      vm.navigateTo('/admin/audit-logs')
      expect(mockPush).toHaveBeenCalledWith('/admin/audit-logs')
    })
  })
})
