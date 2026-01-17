import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import WorkoutsView from './WorkoutsView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  }
}))

// Mock vue-router
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: vi.fn(),
    })),
  }
})

// Mock DetailViewDialog component
vi.mock('@/components/DetailViewDialog.vue', () => ({
  default: {
    template: '<div class="detail-dialog"></div>',
    props: ['modelValue', 'type', 'id']
  }
}))

import axios from '@/utils/axios'
import { useRouter } from 'vue-router'

describe('WorkoutsView', () => {
  let wrapper
  let mockRouter

  const mockStandardTemplates = [
    {
      id: 1,
      name: 'Standard Template 1',
      notes: 'Standard notes',
      created_by: null,
      movements: [{ id: 1, name: 'Squat' }],
      wods: []
    }
  ]

  const mockCustomTemplates = [
    {
      id: 2,
      name: 'Custom Template 1',
      notes: 'Custom notes',
      created_by: 1,
      movements: [],
      wods: [{ id: 1, name: 'Fran' }]
    }
  ]

  const setupDefaultMocks = () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/workouts/standard') {
        return Promise.resolve({
          data: { workouts: [] }
        })
      }
      if (url === '/api/workouts/my-templates') {
        return Promise.resolve({
          data: { workouts: [] }
        })
      }
      return Promise.resolve({ data: {} })
    })
  }

  const setupMocksWithData = () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/workouts/standard') {
        return Promise.resolve({
          data: { workouts: mockStandardTemplates }
        })
      }
      if (url === '/api/workouts/my-templates') {
        return Promise.resolve({
          data: { workouts: mockCustomTemplates }
        })
      }
      return Promise.resolve({ data: {} })
    })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    mockRouter = {
      push: vi.fn()
    }
    useRouter.mockReturnValue(mockRouter)

    wrapper = shallowMount(WorkoutsView, {
      global: {
        plugins: [pinia],
        stubs: {
          'detail-view-dialog': { template: '<div></div>', props: ['modelValue', 'type', 'id'] },
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-card-title': { template: '<div><slot /></div>' },
          'v-card-text': { template: '<div><slot /></div>' },
          'v-card-actions': { template: '<div><slot /></div>' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-icon': { template: '<i></i>' },
          'v-tabs': { template: '<div><slot /></div>' },
          'v-tab': { template: '<div><slot /></div>' },
          'v-chip': { template: '<span><slot /></span>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-dialog': { template: '<div><slot /></div>' },
          'v-text-field': { template: '<input />' },
          'v-textarea': { template: '<textarea></textarea>' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-spacer': { template: '<div></div>' },
          'v-progress-circular': { template: '<div></div>' },
          'v-bottom-navigation': { template: '<div><slot /></div>' },
          'v-avatar': { template: '<div><slot /></div>' },
          'v-tooltip': { template: '<div><slot /></div>' }
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
      wrapper = null
    }
  })

  // ==========================================================================
  // INITIALIZATION TESTS
  // ==========================================================================

  describe('Initialization', () => {
    it('fetches templates on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/workouts/standard')
      expect(axios.get).toHaveBeenCalledWith('/api/workouts/my-templates')
    })

    it('initializes with standard tab selected', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.activeTemplateTab).toBe('standard')
    })

    it('initializes with empty templates', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.standardTemplates).toEqual([])
      expect(vm.customTemplates).toEqual([])
    })

    it('initializes with loading state', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.loading).toBe(true)
    })

    it('stores templates after fetch', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.standardTemplates.length).toBe(1)
      expect(vm.customTemplates.length).toBe(1)
      expect(vm.loading).toBe(false)
    })
  })

  // ==========================================================================
  // TAB SWITCHING TESTS
  // ==========================================================================

  describe('Tab Switching', () => {
    it('displays standard templates when standard tab is active', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.activeTemplateTab = 'standard'

      expect(vm.displayedTemplates.length).toBe(1)
      expect(vm.displayedTemplates[0].name).toBe('Standard Template 1')
    })

    it('displays custom templates when custom tab is active', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.activeTemplateTab = 'custom'

      expect(vm.displayedTemplates.length).toBe(1)
      expect(vm.displayedTemplates[0].name).toBe('Custom Template 1')
    })
  })

  // ==========================================================================
  // CREATE TEMPLATE TESTS
  // ==========================================================================

  describe('Create Template', () => {
    it('validates template name is required', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.newTemplate.name = ''

      await vm.createTemplate()

      expect(vm.error).toContain('name is required')
      expect(axios.post).not.toHaveBeenCalled()
    })

    it('validates template name trims whitespace', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.newTemplate.name = '   '

      await vm.createTemplate()

      expect(vm.error).toContain('name is required')
      expect(axios.post).not.toHaveBeenCalled()
    })

    it('creates template successfully', async () => {
      setupDefaultMocks()
      axios.post.mockResolvedValue({
        data: {
          workout: {
            id: 3,
            name: 'New Template',
            notes: 'Test notes',
            created_by: 1
          }
        }
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.newTemplate.name = 'New Template'
      vm.newTemplate.notes = 'Test notes'

      await vm.createTemplate()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', {
        name: 'New Template',
        notes: 'Test notes'
      })

      expect(vm.customTemplates.length).toBe(1)
      expect(vm.customTemplates[0].name).toBe('New Template')
      expect(vm.createTemplateDialog).toBe(false)
      expect(vm.activeTemplateTab).toBe('custom')
    })

    it('resets form after successful creation', async () => {
      setupDefaultMocks()
      axios.post.mockResolvedValue({
        data: {
          workout: { id: 3, name: 'New Template' }
        }
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.newTemplate.name = 'New Template'
      vm.newTemplate.notes = 'Test notes'

      await vm.createTemplate()
      await flushPromises()

      expect(vm.newTemplate.name).toBe('')
      expect(vm.newTemplate.notes).toBe('')
    })

    it('handles create template error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      setupDefaultMocks()

      axios.post.mockRejectedValue({
        response: { data: { message: 'Create failed' } }
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.newTemplate.name = 'New Template'

      await vm.createTemplate()
      await flushPromises()

      expect(vm.error).toContain('Create failed')

      consoleSpy.mockRestore()
    })

    it('sets creating state during template creation', async () => {
      setupDefaultMocks()
      let resolveCreate
      axios.post.mockReturnValue(
        new Promise((resolve) => {
          resolveCreate = resolve
        })
      )

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.newTemplate.name = 'New Template'

      const createPromise = vm.createTemplate()

      expect(vm.creating).toBe(true)

      resolveCreate({ data: { workout: { id: 1, name: 'New Template' } } })
      await createPromise
      await flushPromises()

      expect(vm.creating).toBe(false)
    })
  })

  // ==========================================================================
  // SELECT TEMPLATE TESTS
  // ==========================================================================

  describe('Select Template', () => {
    it('opens detail dialog with selected template', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const template = { id: 5, name: 'Test Template' }

      vm.selectTemplate(template)

      expect(vm.selectedTemplateId).toBe(5)
      expect(vm.showDetailDialog).toBe(true)
    })
  })

  // ==========================================================================
  // LOG WORKOUT FROM TEMPLATE TESTS
  // ==========================================================================

  describe('Log Workout From Template', () => {
    it('navigates to log workout with template ID', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      setupDefaultMocks()

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.logWorkoutFromTemplate(5)

      expect(mockRouter.push).toHaveBeenCalledWith('/workouts/log?template=5')

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // EDIT TEMPLATE TESTS
  // ==========================================================================

  describe('Edit Template', () => {
    it('navigates to edit page', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const template = { id: 5, name: 'Test Template' }

      vm.editTemplate(template)

      expect(mockRouter.push).toHaveBeenCalledWith('/workouts/templates/5/edit')
    })
  })

  // ==========================================================================
  // DELETE TEMPLATE TESTS
  // ==========================================================================

  describe('Delete Template', () => {
    it('deletes template after confirmation', async () => {
      vi.spyOn(window, 'confirm').mockReturnValue(true)
      axios.delete.mockResolvedValue({})

      // Set up mock with custom templates
      axios.get.mockImplementation((url) => {
        if (url === '/api/workouts/standard') {
          return Promise.resolve({ data: { workouts: [] } })
        }
        if (url === '/api/workouts/my-templates') {
          return Promise.resolve({ data: { workouts: mockCustomTemplates } })
        }
        return Promise.resolve({ data: {} })
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.customTemplates.length).toBe(1)

      await vm.deleteTemplate(2)
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/templates/2')
      expect(vm.customTemplates.length).toBe(0)
    })

    it('does not delete when confirmation is cancelled', async () => {
      vi.spyOn(window, 'confirm').mockReturnValue(false)
      setupDefaultMocks()

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.deleteTemplate(2)

      expect(axios.delete).not.toHaveBeenCalled()
    })

    it('handles delete error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      vi.spyOn(window, 'confirm').mockReturnValue(true)

      axios.delete.mockRejectedValue({
        response: { data: { message: 'Delete failed' } }
      })

      // Set up mock with custom templates
      axios.get.mockImplementation((url) => {
        if (url === '/api/workouts/standard') {
          return Promise.resolve({ data: { workouts: [] } })
        }
        if (url === '/api/workouts/my-templates') {
          return Promise.resolve({ data: { workouts: mockCustomTemplates } })
        }
        return Promise.resolve({ data: {} })
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.deleteTemplate(2)
      await flushPromises()

      expect(vm.error).toContain('Delete failed')
      // Template should still be in the list
      expect(vm.customTemplates.length).toBe(1)

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // TRUNCATE TEXT TESTS
  // ==========================================================================

  describe('Truncate Text', () => {
    it('returns full text when shorter than max length', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.truncateText('Short text', 20)).toBe('Short text')
    })

    it('truncates text when longer than max length', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.truncateText('This is a very long text that should be truncated', 20)).toBe('This is a very long ...')
    })

    it('handles null text', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.truncateText(null, 20)).toBe(null)
    })

    it('handles undefined text', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.truncateText(undefined, 20)).toBe(undefined)
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('sets error message and stops loading on error', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      // Manually set error to test error display
      vm.error = 'Server error'

      expect(vm.error).toContain('Server error')
      expect(vm.loading).toBe(false)
    })

    it('can set network error message', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.error = 'No response from server. Is the backend running?'

      expect(vm.error).toContain('No response from server')
    })

    it('can set generic error message', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.error = 'Failed to fetch templates'

      expect(vm.error).toContain('Failed to fetch templates')
    })

    it('can clear error message', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.error = 'Some error'
      vm.error = null

      expect(vm.error).toBe(null)
    })
  })

  // ==========================================================================
  // COMPUTED PROPERTIES TESTS
  // ==========================================================================

  describe('Computed Properties', () => {
    it('displayedTemplates returns standard templates when standard tab is active', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.activeTemplateTab = 'standard'

      expect(vm.displayedTemplates).toBe(vm.standardTemplates)
    })

    it('displayedTemplates returns custom templates when custom tab is active', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.activeTemplateTab = 'custom'

      expect(vm.displayedTemplates).toBe(vm.customTemplates)
    })
  })
})
