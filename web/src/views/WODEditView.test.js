import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import WODEditView from './WODEditView.vue'

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

describe('WODEditView', () => {
  let wrapper

  const mockMovements = [
    { id: 1, name: 'Back Squat' },
    { id: 2, name: 'Deadlift' },
    { id: 3, name: 'Pull-ups' }
  ]

  const mockWOD = {
    id: 1,
    name: 'Test WOD',
    wod_type: 'for_time',
    scoring_type: 'time',
    time_cap: 20,
    rounds: 3,
    is_benchmark: true,
    description: 'Test description',
    movements: [
      {
        movement_id: 1,
        reps: 21,
        rx_weight: 95,
        scaled_weight: 65,
        distance_meters: null,
        notes: 'Thrusters',
        order_index: 1
      }
    ]
  }

  const setupDefaultMocks = () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/movements') {
        return Promise.resolve({ data: { movements: mockMovements } })
      }
      if (url.startsWith('/api/wods/')) {
        return Promise.resolve({ data: { wod: mockWOD } })
      }
      return Promise.resolve({ data: {} })
    })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(WODEditView, {
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
          'v-autocomplete': { template: '<input />' },
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
    it('fetches movements on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/movements')
    })

    it('initializes with empty form in create mode', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.wod.name).toBe('')
      expect(vm.wod.movements).toEqual([])
      expect(vm.isEditMode).toBe(false)
    })

    it('loads existing WOD in edit mode', async () => {
      mockRouteParams = { id: '1' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(axios.get).toHaveBeenCalledWith('/api/wods/1')
      expect(vm.wod.name).toBe('Test WOD')
      expect(vm.isEditMode).toBe(true)
    })

    it('stores fetched movements', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.availableMovements).toEqual(mockMovements)
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('handles fetch movements error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue(new Error('Network error'))

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Failed to load movements')

      consoleSpy.mockRestore()
    })

    it('handles load WOD error', async () => {
      mockRouteParams = { id: '1' }
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockImplementation((url) => {
        if (url === '/api/movements') {
          return Promise.resolve({ data: { movements: mockMovements } })
        }
        return Promise.reject(new Error('WOD not found'))
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Failed to load WOD')

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // MOVEMENT MANAGEMENT TESTS
  // ==========================================================================

  describe('Movement Management', () => {
    it('adds new movement', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.wod.movements.length).toBe(0)

      vm.addMovement()

      expect(vm.wod.movements.length).toBe(1)
      expect(vm.wod.movements[0].movement_id).toBe(null)
      expect(vm.wod.movements[0].order_index).toBe(1)
    })

    it('removes movement', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.addMovement()
      vm.addMovement()

      expect(vm.wod.movements.length).toBe(2)

      vm.removeMovement(0)

      expect(vm.wod.movements.length).toBe(1)
    })

    it('updates order indices after removal', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.addMovement()
      vm.addMovement()
      vm.addMovement()

      vm.wod.movements[0].movement_id = 1
      vm.wod.movements[1].movement_id = 2
      vm.wod.movements[2].movement_id = 3

      vm.removeMovement(1) // Remove middle movement

      expect(vm.wod.movements.length).toBe(2)
      expect(vm.wod.movements[0].order_index).toBe(1)
      expect(vm.wod.movements[1].order_index).toBe(2)
    })
  })

  // ==========================================================================
  // VALIDATION TESTS
  // ==========================================================================

  describe('Validation', () => {
    it('validates name is required', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.wod.name = ''
      vm.addMovement()
      vm.wod.movements[0].movement_id = 1

      const isValid = vm.validateWOD()

      expect(isValid).toBe(false)
      expect(vm.validationErrors.name).toBe('WOD name is required')
    })

    it('validates at least one movement required', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.wod.name = 'Test WOD'

      const isValid = vm.validateWOD()

      expect(isValid).toBe(false)
      expect(vm.error).toContain('at least one movement')
    })

    it('validates movement selection required', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.wod.name = 'Test WOD'
      vm.addMovement()
      // Don't select a movement

      const isValid = vm.validateWOD()

      expect(isValid).toBe(false)
      expect(vm.error).toContain('Please select a movement')
    })

    it('passes validation with valid data', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.wod.name = 'Test WOD'
      vm.addMovement()
      vm.wod.movements[0].movement_id = 1

      const isValid = vm.validateWOD()

      expect(isValid).toBe(true)
    })
  })

  // ==========================================================================
  // SAVE TESTS
  // ==========================================================================

  describe('Save WOD', () => {
    it('creates new WOD in create mode', async () => {
      vi.useFakeTimers()
      axios.post.mockResolvedValue({ data: { wod: { id: 100 } } })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.wod.name = 'New WOD'
      vm.wod.wod_type = 'amrap'
      vm.wod.scoring_type = 'rounds_reps'
      vm.wod.description = 'Test description'
      vm.addMovement()
      vm.wod.movements[0].movement_id = 1
      vm.wod.movements[0].reps = 10

      await vm.saveWOD()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/wods', expect.objectContaining({
        name: 'New WOD',
        wod_type: 'amrap',
        scoring_type: 'rounds_reps',
        description: 'Test description',
        movements: [expect.objectContaining({
          movement_id: 1,
          reps: 10,
          order_index: 1
        })]
      }))
      expect(vm.successMessage).toContain('created successfully')

      // Advance timer to trigger redirect
      vi.advanceTimersByTime(1500)
      expect(mockPush).toHaveBeenCalledWith('/wods/100/edit')

      vi.useRealTimers()
    })

    it('updates existing WOD in edit mode', async () => {
      mockRouteParams = { id: '1' }
      axios.put.mockResolvedValue({})
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.wod.name = 'Updated WOD'

      await vm.saveWOD()
      await flushPromises()

      expect(axios.put).toHaveBeenCalledWith('/api/wods/1', expect.objectContaining({
        name: 'Updated WOD'
      }))
      expect(vm.successMessage).toContain('updated successfully')
    })

    it('handles save error with message', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue({
        response: { data: { message: 'Duplicate WOD name' } }
      })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.wod.name = 'Test WOD'
      vm.addMovement()
      vm.wod.movements[0].movement_id = 1

      await vm.saveWOD()
      await flushPromises()

      expect(vm.error).toBe('Duplicate WOD name')

      consoleSpy.mockRestore()
    })

    it('handles generic save error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue(new Error('Network error'))
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.wod.name = 'Test WOD'
      vm.addMovement()
      vm.wod.movements[0].movement_id = 1

      await vm.saveWOD()
      await flushPromises()

      expect(vm.error).toContain('Failed to save WOD')

      consoleSpy.mockRestore()
    })

    it('sets saving state during save', async () => {
      let resolvePost
      axios.post.mockReturnValue(new Promise(resolve => { resolvePost = resolve }))
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.wod.name = 'Test WOD'
      vm.addMovement()
      vm.wod.movements[0].movement_id = 1

      const savePromise = vm.saveWOD()

      expect(vm.saving).toBe(true)

      resolvePost({ data: { wod: { id: 100 } } })
      await savePromise
      await flushPromises()

      expect(vm.saving).toBe(false)
    })
  })

  // ==========================================================================
  // DELETE TESTS
  // ==========================================================================

  describe('Delete WOD', () => {
    it('opens delete confirmation dialog', async () => {
      mockRouteParams = { id: '1' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.confirmDelete()

      expect(vm.deleteDialog).toBe(true)
    })

    it('deletes WOD successfully', async () => {
      vi.useFakeTimers()
      mockRouteParams = { id: '1' }
      axios.delete.mockResolvedValue({})
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      await vm.deleteWOD()
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/wods/1')
      expect(vm.successMessage).toContain('deleted successfully')

      vi.advanceTimersByTime(1000)
      expect(mockPush).toHaveBeenCalledWith('/wods')

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

      await vm.deleteWOD()
      await flushPromises()

      expect(vm.error).toContain('Failed to delete WOD')
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
      const deletePromise = vm.deleteWOD()

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
    it('has WOD type options', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.wodTypes.length).toBeGreaterThan(0)
      expect(vm.wodTypes.some(t => t.value === 'for_time')).toBe(true)
      expect(vm.wodTypes.some(t => t.value === 'amrap')).toBe(true)
      expect(vm.wodTypes.some(t => t.value === 'emom')).toBe(true)
    })

    it('has scoring type options', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.scoringTypes.length).toBeGreaterThan(0)
      expect(vm.scoringTypes.some(s => s.value === 'time')).toBe(true)
      expect(vm.scoringTypes.some(s => s.value === 'rounds_reps')).toBe(true)
    })
  })
})
