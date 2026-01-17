import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ExportView from './ExportView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
  }
}))

import axios from '@/utils/axios'

describe('ExportView', () => {
  let wrapper

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(ExportView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-card-title': { template: '<div><slot /></div>' },
          'v-card-text': { template: '<div><slot /></div>' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-icon': { template: '<i></i>' },
          'v-checkbox': { template: '<input type="checkbox" />' },
          'v-text-field': { template: '<input />' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-bottom-navigation': { template: '<div><slot /></div>' },
          'v-divider': { template: '<hr />' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' }
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    // Mock URL methods globally
    global.URL.createObjectURL = vi.fn(() => 'blob:test-url')
    global.URL.revokeObjectURL = vi.fn()
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
    it('initializes with default export options', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.exportOptions.wods.includeStandard).toBe(true)
      expect(vm.exportOptions.wods.includeCustom).toBe(true)
      expect(vm.exportOptions.movements.includeStandard).toBe(true)
      expect(vm.exportOptions.movements.includeCustom).toBe(true)
      expect(vm.exportOptions.userWorkouts.startDate).toBe('')
      expect(vm.exportOptions.userWorkouts.endDate).toBe('')
    })

    it('initializes with no loading states', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.exportingWods).toBe(null)
      expect(vm.exportingMovements).toBe(null)
      expect(vm.exportingUserWorkouts).toBe(null)
    })

    it('initializes with no messages', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.successMessage).toBe(null)
      expect(vm.error).toBe(null)
    })
  })

  // ==========================================================================
  // DATE RANGE VALIDATION TESTS
  // ==========================================================================

  describe('Date Range Validation', () => {
    it('returns no error when both dates are empty', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.exportOptions.userWorkouts.startDate = ''
      vm.exportOptions.userWorkouts.endDate = ''

      expect(vm.dateRangeError).toBe(null)
    })

    it('returns error when only start date is provided', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.exportOptions.userWorkouts.startDate = '2024-01-01'
      vm.exportOptions.userWorkouts.endDate = ''

      expect(vm.dateRangeError).toContain('Both start and end dates')
    })

    it('returns error when only end date is provided', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.exportOptions.userWorkouts.startDate = ''
      vm.exportOptions.userWorkouts.endDate = '2024-12-31'

      expect(vm.dateRangeError).toContain('Both start and end dates')
    })

    it('returns error when start date is after end date', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.exportOptions.userWorkouts.startDate = '2024-12-31'
      vm.exportOptions.userWorkouts.endDate = '2024-01-01'

      expect(vm.dateRangeError).toContain('Start date must be before')
    })

    it('returns no error when both dates are valid', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.exportOptions.userWorkouts.startDate = '2024-01-01'
      vm.exportOptions.userWorkouts.endDate = '2024-12-31'

      expect(vm.dateRangeError).toBe(null)
    })

    it('returns no error when start date equals end date', () => {
      createWrapper()

      const vm = wrapper.vm
      vm.exportOptions.userWorkouts.startDate = '2024-06-15'
      vm.exportOptions.userWorkouts.endDate = '2024-06-15'

      expect(vm.dateRangeError).toBe(null)
    })
  })

  // ==========================================================================
  // EXPORT OPTIONS TESTS
  // ==========================================================================

  describe('Export Options', () => {
    it('can toggle WOD export options', () => {
      createWrapper()

      const vm = wrapper.vm

      vm.exportOptions.wods.includeStandard = false
      expect(vm.exportOptions.wods.includeStandard).toBe(false)

      vm.exportOptions.wods.includeCustom = false
      expect(vm.exportOptions.wods.includeCustom).toBe(false)
    })

    it('can toggle movement export options', () => {
      createWrapper()

      const vm = wrapper.vm

      vm.exportOptions.movements.includeStandard = false
      expect(vm.exportOptions.movements.includeStandard).toBe(false)

      vm.exportOptions.movements.includeCustom = false
      expect(vm.exportOptions.movements.includeCustom).toBe(false)
    })

    it('can set date range for user workouts export', () => {
      createWrapper()

      const vm = wrapper.vm

      vm.exportOptions.userWorkouts.startDate = '2024-01-01'
      vm.exportOptions.userWorkouts.endDate = '2024-12-31'

      expect(vm.exportOptions.userWorkouts.startDate).toBe('2024-01-01')
      expect(vm.exportOptions.userWorkouts.endDate).toBe('2024-12-31')
    })
  })

  // ==========================================================================
  // EXPORT API CALLS TESTS
  // ==========================================================================

  describe('Export API Calls', () => {
    it('calls WODs export API with correct parameters', async () => {
      axios.get.mockResolvedValue({
        data: new Blob(['test'], { type: 'text/csv' })
      })

      createWrapper()

      const vm = wrapper.vm
      vm.exportOptions.wods.includeStandard = true
      vm.exportOptions.wods.includeCustom = false

      await vm.exportWODs('csv')
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/export/wods', {
        params: {
          include_standard: true,
          include_custom: false,
          format: 'csv'
        },
        responseType: 'blob'
      })
    })

    it('calls movements export API with correct parameters', async () => {
      axios.get.mockResolvedValue({
        data: new Blob(['test'], { type: 'text/csv' })
      })

      createWrapper()

      const vm = wrapper.vm
      vm.exportOptions.movements.includeStandard = false
      vm.exportOptions.movements.includeCustom = true

      await vm.exportMovements('json')
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/export/movements', {
        params: {
          include_standard: false,
          include_custom: true,
          format: 'json'
        },
        responseType: 'blob'
      })
    })

    it('calls user workouts export API without date range', async () => {
      axios.get.mockResolvedValue({
        data: new Blob(['test'], { type: 'text/csv' })
      })

      createWrapper()

      const vm = wrapper.vm

      await vm.exportUserWorkouts('csv')
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/export/user-workouts', {
        params: {
          format: 'csv'
        },
        responseType: 'blob'
      })
    })

    it('calls user workouts export API with date range', async () => {
      axios.get.mockResolvedValue({
        data: new Blob(['test'], { type: 'text/csv' })
      })

      createWrapper()

      const vm = wrapper.vm
      vm.exportOptions.userWorkouts.startDate = '2024-01-01'
      vm.exportOptions.userWorkouts.endDate = '2024-06-30'

      await vm.exportUserWorkouts('json')
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/export/user-workouts', {
        params: {
          format: 'json',
          start_date: '2024-01-01',
          end_date: '2024-06-30'
        },
        responseType: 'blob'
      })
    })
  })

  // ==========================================================================
  // LOADING STATE TESTS
  // ==========================================================================

  describe('Loading States', () => {
    it('sets loading state during WODs export', async () => {
      let resolveExport
      axios.get.mockReturnValue(
        new Promise((resolve) => {
          resolveExport = resolve
        })
      )

      createWrapper()

      const vm = wrapper.vm
      const exportPromise = vm.exportWODs('csv')

      expect(vm.exportingWods).toBe('csv')

      resolveExport({ data: new Blob(['test']) })
      await exportPromise
      await flushPromises()

      expect(vm.exportingWods).toBe(null)
    })

    it('sets loading state during movements export', async () => {
      let resolveExport
      axios.get.mockReturnValue(
        new Promise((resolve) => {
          resolveExport = resolve
        })
      )

      createWrapper()

      const vm = wrapper.vm
      const exportPromise = vm.exportMovements('json')

      expect(vm.exportingMovements).toBe('json')

      resolveExport({ data: new Blob(['test']) })
      await exportPromise
      await flushPromises()

      expect(vm.exportingMovements).toBe(null)
    })

    it('sets loading state during user workouts export', async () => {
      let resolveExport
      axios.get.mockReturnValue(
        new Promise((resolve) => {
          resolveExport = resolve
        })
      )

      createWrapper()

      const vm = wrapper.vm
      const exportPromise = vm.exportUserWorkouts('csv')

      expect(vm.exportingUserWorkouts).toBe('csv')

      resolveExport({ data: new Blob(['test']) })
      await exportPromise
      await flushPromises()

      expect(vm.exportingUserWorkouts).toBe(null)
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('handles WODs export error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      axios.get.mockRejectedValue({
        response: { data: { error: 'Export failed' } }
      })

      createWrapper()

      const vm = wrapper.vm
      await vm.exportWODs('csv')
      await flushPromises()

      expect(vm.error).toContain('Export failed')
      expect(vm.exportingWods).toBe(null)

      consoleSpy.mockRestore()
    })

    it('handles movements export error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      axios.get.mockRejectedValue({
        response: { data: { error: 'Movement export failed' } }
      })

      createWrapper()

      const vm = wrapper.vm
      await vm.exportMovements('csv')
      await flushPromises()

      expect(vm.error).toContain('Movement export failed')

      consoleSpy.mockRestore()
    })

    it('handles user workouts export error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      axios.get.mockRejectedValue({
        response: { data: { error: 'Workout export failed' } }
      })

      createWrapper()

      const vm = wrapper.vm
      await vm.exportUserWorkouts('csv')
      await flushPromises()

      expect(vm.error).toContain('Workout export failed')

      consoleSpy.mockRestore()
    })

    it('handles generic error without response data', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      axios.get.mockRejectedValue(new Error('Network error'))

      createWrapper()

      const vm = wrapper.vm
      await vm.exportUserWorkouts('csv')
      await flushPromises()

      expect(vm.error).toContain('Failed to export user workouts')

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // SUCCESS MESSAGE TESTS
  // ==========================================================================

  describe('Success Messages', () => {
    it('sets success message after WODs export', async () => {
      axios.get.mockResolvedValue({
        data: new Blob(['test'], { type: 'text/csv' })
      })

      createWrapper()

      const vm = wrapper.vm
      await vm.exportWODs('csv')
      await flushPromises()

      expect(vm.successMessage).toContain('WODs exported successfully')
      expect(vm.successMessage).toContain('CSV')
    })

    it('sets success message after movements export', async () => {
      axios.get.mockResolvedValue({
        data: new Blob(['test'], { type: 'application/json' })
      })

      createWrapper()

      const vm = wrapper.vm
      await vm.exportMovements('json')
      await flushPromises()

      expect(vm.successMessage).toContain('Movements exported successfully')
      expect(vm.successMessage).toContain('JSON')
    })

    it('sets success message after user workouts export', async () => {
      axios.get.mockResolvedValue({
        data: new Blob(['test'], { type: 'text/csv' })
      })

      createWrapper()

      const vm = wrapper.vm
      await vm.exportUserWorkouts('csv')
      await flushPromises()

      expect(vm.successMessage).toContain('User workouts exported successfully')
    })

    it('clears previous error on new export', async () => {
      axios.get.mockResolvedValue({
        data: new Blob(['test'], { type: 'text/csv' })
      })

      createWrapper()

      const vm = wrapper.vm
      vm.error = 'Previous error'

      await vm.exportWODs('csv')
      await flushPromises()

      expect(vm.error).toBe(null)
    })
  })
})
