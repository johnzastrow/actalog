import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import MovementEditView from './MovementEditView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  }
}))

// Mock vue-router
const mockPush = vi.fn()
let mockRouteParams = { id: undefined }
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: mockPush,
    })),
    useRoute: vi.fn(() => ({
      params: mockRouteParams
    }))
  }
})

import axios from '@/utils/axios'

describe('MovementEditView', () => {
  let wrapper

  const mockMovement = {
    id: 1,
    name: 'Test Movement',
    type: 'weightlifting',
    primary_muscle_group: 'upper_body',
    description: 'Test description',
    is_weighted: true,
    is_timed: false,
    is_custom: true,
    equipment: ['Barbell', 'Dumbbells'],
    video_url: 'https://example.com/video',
    demo_image_url: 'https://example.com/image.jpg',
    rx_weight_male: 135,
    rx_weight_female: 95,
    scaled_weight_male: 95,
    scaled_weight_female: 65
  }

  const mockStandardMovement = {
    ...mockMovement,
    id: 2,
    name: 'Back Squat',
    is_custom: false
  }

  const setupDefaultMocks = () => {
    axios.get.mockResolvedValue({ data: { movement: mockMovement } })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(MovementEditView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-card-title': { template: '<div><slot /></div>' },
          'v-card-text': { template: '<div><slot /></div>' },
          'v-card-actions': { template: '<div><slot /></div>' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-icon': { template: '<i></i>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-text-field': { template: '<input />' },
          'v-textarea': { template: '<textarea></textarea>' },
          'v-select': { template: '<select></select>' },
          'v-combobox': { template: '<input />' },
          'v-switch': { template: '<input type="checkbox" />' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-spacer': { template: '<div></div>' },
          'v-dialog': { template: '<div><slot /></div>' }
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    mockPush.mockClear()
    mockRouteParams = { id: undefined }
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
    it('initializes with empty form in create mode', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.movement.name).toBe('')
      expect(vm.movement.movement_type).toBe('weightlifting')
      expect(vm.movement.is_custom).toBe(true)
      expect(vm.isEditMode).toBe(false)
    })

    it('loads existing movement in edit mode', async () => {
      mockRouteParams = { id: '1' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(axios.get).toHaveBeenCalledWith('/api/movements/1')
      expect(vm.movement.name).toBe('Test Movement')
      expect(vm.isEditMode).toBe(true)
    })

    it('maps movement type from data', async () => {
      mockRouteParams = { id: '1' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.movement.movement_type).toBe('weightlifting')
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('handles load movement error', async () => {
      mockRouteParams = { id: '1' }
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue(new Error('Not found'))

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Failed to load movement')

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // VALIDATION TESTS
  // ==========================================================================

  describe('Validation', () => {
    it('validates name is required', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.movement.name = ''

      const isValid = vm.validateMovement()

      expect(isValid).toBe(false)
      expect(vm.validationErrors.name).toBe('Movement name is required')
    })

    it('validates whitespace-only name as invalid', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.movement.name = '   '

      const isValid = vm.validateMovement()

      expect(isValid).toBe(false)
      expect(vm.validationErrors.name).toBe('Movement name is required')
    })

    it('passes validation with valid name', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.movement.name = 'Test Movement'

      const isValid = vm.validateMovement()

      expect(isValid).toBe(true)
      expect(Object.keys(vm.validationErrors).length).toBe(0)
    })
  })

  // ==========================================================================
  // SAVE TESTS
  // ==========================================================================

  describe('Save Movement', () => {
    it('creates new movement in create mode', async () => {
      vi.useFakeTimers()
      axios.post.mockResolvedValue({ data: { movement: { id: 100 } } })
      createWrapper()

      const vm = wrapper.vm
      vm.movement.name = 'New Movement'
      vm.movement.movement_type = 'gymnastics'
      vm.movement.primary_muscle_group = 'upper_body'
      vm.movement.description = 'Test description'
      vm.movement.equipment = ['Pull-up Bar']

      await vm.saveMovement()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/movements', expect.objectContaining({
        name: 'New Movement',
        type: 'gymnastics',
        primary_muscle_group: 'upper_body',
        description: 'Test description',
        equipment: ['Pull-up Bar']
      }))
      expect(vm.successMessage).toContain('created successfully')

      vi.advanceTimersByTime(1500)
      expect(mockPush).toHaveBeenCalledWith('/movements/100/edit')

      vi.useRealTimers()
    })

    it('updates existing movement in edit mode', async () => {
      mockRouteParams = { id: '1' }
      axios.put.mockResolvedValue({})
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.movement.name = 'Updated Movement'

      await vm.saveMovement()
      await flushPromises()

      expect(axios.put).toHaveBeenCalledWith('/api/movements/1', expect.objectContaining({
        name: 'Updated Movement'
      }))
      expect(vm.successMessage).toContain('updated successfully')
    })

    it('does not save if validation fails', async () => {
      createWrapper()

      const vm = wrapper.vm
      vm.movement.name = ''

      await vm.saveMovement()
      await flushPromises()

      expect(axios.post).not.toHaveBeenCalled()
    })

    it('handles save error with message', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue({
        response: { data: { message: 'Duplicate movement name' } }
      })
      createWrapper()

      const vm = wrapper.vm
      vm.movement.name = 'Test Movement'

      await vm.saveMovement()
      await flushPromises()

      expect(vm.error).toBe('Duplicate movement name')

      consoleSpy.mockRestore()
    })

    it('handles generic save error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue(new Error('Network error'))
      createWrapper()

      const vm = wrapper.vm
      vm.movement.name = 'Test Movement'

      await vm.saveMovement()
      await flushPromises()

      expect(vm.error).toContain('Failed to save movement')

      consoleSpy.mockRestore()
    })

    it('sets saving state during save', async () => {
      let resolvePost
      axios.post.mockReturnValue(new Promise(resolve => { resolvePost = resolve }))
      createWrapper()

      const vm = wrapper.vm
      vm.movement.name = 'Test Movement'

      const savePromise = vm.saveMovement()

      expect(vm.saving).toBe(true)

      resolvePost({ data: { movement: { id: 100 } } })
      await savePromise
      await flushPromises()

      expect(vm.saving).toBe(false)
    })

    it('trims whitespace from string fields', async () => {
      axios.post.mockResolvedValue({ data: { movement: { id: 100 } } })
      createWrapper()

      const vm = wrapper.vm
      vm.movement.name = '  Test Movement  '
      vm.movement.description = '  Test description  '
      vm.movement.video_url = '  https://example.com  '

      await vm.saveMovement()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/movements', expect.objectContaining({
        name: 'Test Movement',
        description: 'Test description',
        video_url: 'https://example.com'
      }))
    })
  })

  // ==========================================================================
  // DELETE TESTS
  // ==========================================================================

  describe('Delete Movement', () => {
    it('opens delete confirmation for custom movement', async () => {
      mockRouteParams = { id: '1' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.confirmDelete()

      expect(vm.deleteDialog).toBe(true)
    })

    it('does not allow deleting standard movement', async () => {
      mockRouteParams = { id: '2' }
      axios.get.mockResolvedValue({ data: { movement: mockStandardMovement } })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.confirmDelete()

      expect(vm.deleteDialog).toBe(false)
      expect(vm.error).toContain('Standard movements cannot be deleted')
    })

    it('deletes movement successfully', async () => {
      vi.useFakeTimers()
      mockRouteParams = { id: '1' }
      axios.delete.mockResolvedValue({})
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      await vm.deleteMovement()
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/movements/1')
      expect(vm.successMessage).toContain('deleted successfully')

      vi.advanceTimersByTime(1000)
      expect(mockPush).toHaveBeenCalledWith('/workouts')

      vi.useRealTimers()
    })

    it('handles delete error', async () => {
      mockRouteParams = { id: '1' }
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.delete.mockRejectedValue(new Error('Cannot delete'))
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.deleteDialog = true

      await vm.deleteMovement()
      await flushPromises()

      expect(vm.error).toContain('Failed to delete movement')
      expect(vm.deleteDialog).toBe(false)

      consoleSpy.mockRestore()
    })

    it('sets deleting state during deletion', async () => {
      mockRouteParams = { id: '1' }
      let resolveDelete
      axios.delete.mockReturnValue(new Promise(resolve => { resolveDelete = resolve }))
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const deletePromise = vm.deleteMovement()

      expect(vm.deleting).toBe(true)

      resolveDelete({})
      await deletePromise
      await flushPromises()

      expect(vm.deleting).toBe(false)
    })
  })

  // ==========================================================================
  // OPTIONS DATA TESTS
  // ==========================================================================

  describe('Options Data', () => {
    it('has movement type options', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.movementTypes.length).toBeGreaterThan(0)
      expect(vm.movementTypes.some(t => t.value === 'weightlifting')).toBe(true)
      expect(vm.movementTypes.some(t => t.value === 'gymnastics')).toBe(true)
      expect(vm.movementTypes.some(t => t.value === 'cardio')).toBe(true)
    })

    it('has muscle group options', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.muscleGroups.length).toBeGreaterThan(0)
      expect(vm.muscleGroups.some(g => g.value === 'full_body')).toBe(true)
      expect(vm.muscleGroups.some(g => g.value === 'upper_body')).toBe(true)
      expect(vm.muscleGroups.some(g => g.value === 'lower_body')).toBe(true)
    })

    it('has common equipment options', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.commonEquipment.length).toBeGreaterThan(0)
      expect(vm.commonEquipment).toContain('Barbell')
      expect(vm.commonEquipment).toContain('Dumbbells')
      expect(vm.commonEquipment).toContain('Kettlebell')
    })
  })
})
