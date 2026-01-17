import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import MovementsLibraryView from './MovementsLibraryView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  }
}))

// Mock vue-router
const mockPush = vi.fn()
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: mockPush,
    })),
    useRoute: vi.fn(() => ({
      query: {}
    }))
  }
})

// Mock DetailViewDialog component
vi.mock('@/components/DetailViewDialog.vue', () => ({
  default: {
    template: '<div class="detail-view-dialog"></div>',
    props: ['modelValue', 'type', 'id', 'handleQuickLogExternally']
  }
}))

import axios from '@/utils/axios'
import { useRouter, useRoute } from 'vue-router'

describe('MovementsLibraryView', () => {
  let wrapper

  const mockMovements = [
    {
      id: 1,
      name: 'Back Squat',
      description: 'Barbell back squat',
      type: 'weightlifting',
      is_standard: true
    },
    {
      id: 2,
      name: 'Pull-up',
      description: 'Strict pull-up',
      type: 'gymnastics',
      is_standard: true
    },
    {
      id: 3,
      name: 'Running',
      description: '400m run',
      type: 'cardio',
      is_standard: true
    },
    {
      id: 4,
      name: 'Push-up',
      description: 'Standard push-up',
      type: 'bodyweight',
      is_standard: true
    },
    {
      id: 5,
      name: 'Custom Lift',
      description: 'My custom movement',
      type: 'weightlifting',
      is_standard: false
    }
  ]

  const setupDefaultMocks = () => {
    axios.get.mockResolvedValue({
      data: { movements: [] }
    })
  }

  const setupMocksWithData = () => {
    axios.get.mockResolvedValue({
      data: { movements: mockMovements }
    })
  }

  const createWrapper = (routeQuery = {}) => {
    const pinia = createPinia()
    setActivePinia(pinia)

    // Setup route mock with query
    useRoute.mockReturnValue({
      query: routeQuery
    })

    wrapper = shallowMount(MovementsLibraryView, {
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
          'v-chip': { template: '<span><slot /></span>' },
          'v-chip-group': { template: '<div><slot /></div>' },
          'v-text-field': { template: '<input />' },
          'v-textarea': { template: '<textarea></textarea>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-progress-circular': { template: '<div></div>' },
          'v-bottom-navigation': { template: '<div><slot /></div>' },
          'v-dialog': { template: '<div><slot /></div>' },
          'v-form': { template: '<form><slot /></form>' },
          'v-spacer': { template: '<div></div>' },
          'v-tooltip': { template: '<div><slot /></div>' },
          'v-menu': { template: '<div><slot /></div>' },
          'DetailViewDialog': true
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
  // INITIALIZATION TESTS
  // ==========================================================================

  describe('Initialization', () => {
    it('fetches movements on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/movements')
    })

    it('initializes with default state', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.movements).toEqual([])
      expect(vm.loading).toBe(true) // Still loading on initial render
      expect(vm.error).toBe('')
      expect(vm.searchQuery).toBe('')
      expect(vm.selectedType).toBe('all')
    })

    it('stores fetched movements', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.movements).toHaveLength(5)
      expect(vm.movements[0].name).toBe('Back Squat')
    })

    it('sets loading to false after fetch', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('handles fetch error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue(new Error('Network error'))

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Failed to load movements. Please try again.')
      expect(vm.loading).toBe(false)

      consoleSpy.mockRestore()
    })

    it('clears error on successful fetch', async () => {
      setupMocksWithData()
      createWrapper()

      const vm = wrapper.vm
      vm.error = 'Previous error'

      await vm.fetchMovements()
      await flushPromises()

      expect(vm.error).toBe('')
    })
  })

  // ==========================================================================
  // FILTERING TESTS
  // ==========================================================================

  describe('Filtering', () => {
    it('filters movements by type - weightlifting', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedType = 'weightlifting'

      expect(vm.filteredMovements).toHaveLength(2)
      expect(vm.filteredMovements.every(m => m.type === 'weightlifting')).toBe(true)
    })

    it('filters movements by type - gymnastics', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedType = 'gymnastics'

      expect(vm.filteredMovements).toHaveLength(1)
      expect(vm.filteredMovements[0].name).toBe('Pull-up')
    })

    it('filters movements by type - cardio', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedType = 'cardio'

      expect(vm.filteredMovements).toHaveLength(1)
      expect(vm.filteredMovements[0].name).toBe('Running')
    })

    it('filters movements by type - bodyweight', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedType = 'bodyweight'

      expect(vm.filteredMovements).toHaveLength(1)
      expect(vm.filteredMovements[0].name).toBe('Push-up')
    })

    it('shows all movements when type is all', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedType = 'all'

      expect(vm.filteredMovements).toHaveLength(5)
    })

    it('filters movements by search query - name match', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.searchQuery = 'squat'

      expect(vm.filteredMovements).toHaveLength(1)
      expect(vm.filteredMovements[0].name).toBe('Back Squat')
    })

    it('filters movements by search query - description match', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.searchQuery = 'barbell'

      expect(vm.filteredMovements).toHaveLength(1)
      expect(vm.filteredMovements[0].name).toBe('Back Squat')
    })

    it('filters movements by search query - case insensitive', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.searchQuery = 'PULL'

      expect(vm.filteredMovements).toHaveLength(1)
      expect(vm.filteredMovements[0].name).toBe('Pull-up')
    })

    it('combines type and search filters', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedType = 'weightlifting'
      vm.searchQuery = 'custom'

      expect(vm.filteredMovements).toHaveLength(1)
      expect(vm.filteredMovements[0].name).toBe('Custom Lift')
    })

    it('returns empty array when no matches', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.searchQuery = 'nonexistent'

      expect(vm.filteredMovements).toHaveLength(0)
    })
  })

  // ==========================================================================
  // SELECTION MODE TESTS
  // ==========================================================================

  describe('Selection Mode', () => {
    it('detects selection mode from route query', async () => {
      setupDefaultMocks()
      createWrapper({ select: 'true' })
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.selectionMode).toBe(true)
    })

    it('is not in selection mode by default', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.selectionMode).toBe(false)
    })

    it('navigates with selected movement in selection mode', async () => {
      setupMocksWithData()
      createWrapper({ select: 'true', returnPath: '/workouts/templates/create' })
      await flushPromises()

      const vm = wrapper.vm
      vm.handleMovementClick(mockMovements[0])

      expect(mockPush).toHaveBeenCalledWith({
        path: '/workouts/templates/create',
        query: { selectedMovement: 1 }
      })
    })

    it('uses default return path in selection mode', async () => {
      setupMocksWithData()
      createWrapper({ select: 'true' })
      await flushPromises()

      const vm = wrapper.vm
      vm.handleMovementClick(mockMovements[0])

      expect(mockPush).toHaveBeenCalledWith({
        path: '/workouts/templates/create',
        query: { selectedMovement: 1 }
      })
    })

    it('shows detail dialog when not in selection mode', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.handleMovementClick(mockMovements[0])

      expect(vm.showDetailDialog).toBe(true)
      expect(vm.selectedMovementId).toBe(1)
    })
  })

  // ==========================================================================
  // NAVIGATION TESTS
  // ==========================================================================

  describe('Navigation', () => {
    it('navigates to create movement page', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.createNewMovement()

      expect(mockPush).toHaveBeenCalledWith('/movements/create')
    })

    it('navigates to create movement with return path in selection mode', async () => {
      setupDefaultMocks()
      createWrapper({ select: 'true', returnPath: '/some/path' })
      await flushPromises()

      const vm = wrapper.vm
      vm.createNewMovement()

      expect(mockPush).toHaveBeenCalledWith({
        path: '/movements/create',
        query: { returnPath: '/some/path', select: 'true' }
      })
    })

    it('navigates to edit movement page', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.editMovement(5)

      expect(mockPush).toHaveBeenCalledWith('/movements/5/edit')
    })
  })

  // ==========================================================================
  // HELPER FUNCTION TESTS
  // ==========================================================================

  describe('Helper Functions', () => {
    describe('getMovementTypeIcon', () => {
      it('returns correct icon for weightlifting', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeIcon('weightlifting')).toBe('mdi-weight-lifter')
      })

      it('returns correct icon for gymnastics', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeIcon('gymnastics')).toBe('mdi-gymnastics')
      })

      it('returns correct icon for cardio', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeIcon('cardio')).toBe('mdi-run')
      })

      it('returns correct icon for bodyweight', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeIcon('bodyweight')).toBe('mdi-human')
      })

      it('returns default icon for unknown type', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeIcon('unknown')).toBe('mdi-dumbbell')
      })
    })

    describe('getMovementTypeColor', () => {
      it('returns correct color for weightlifting', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeColor('weightlifting')).toBe('#00bcd4')
      })

      it('returns correct color for gymnastics', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeColor('gymnastics')).toBe('#9c27b0')
      })

      it('returns correct color for cardio', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeColor('cardio')).toBe('#ff5722')
      })

      it('returns correct color for bodyweight', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeColor('bodyweight')).toBe('#4caf50')
      })

      it('returns default color for unknown type', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeColor('unknown')).toBe('#666')
      })
    })

    describe('capitalizeFirst', () => {
      it('capitalizes first letter', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.capitalizeFirst('weightlifting')).toBe('Weightlifting')
      })

      it('handles empty string', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.capitalizeFirst('')).toBe('')
      })

      it('handles null/undefined', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.capitalizeFirst(null)).toBe('')
        expect(vm.capitalizeFirst(undefined)).toBe('')
      })
    })

    describe('getTodayDate', () => {
      it('returns date in YYYY-MM-DD format', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.getTodayDate()

        expect(result).toMatch(/^\d{4}-\d{2}-\d{2}$/)
      })
    })

    describe('formatQuickLogName', () => {
      it('returns day-based workout name', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        // Monday Jan 1, 2024
        const result = vm.formatQuickLogName('2024-01-01')

        expect(result).toBe('Monday Workout')
      })

      it('handles empty date', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatQuickLogName('')

        expect(result).toBe('Workout')
      })

      it('handles null date', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatQuickLogName(null)

        expect(result).toBe('Workout')
      })
    })
  })

  // ==========================================================================
  // QUICK LOG TESTS
  // ==========================================================================

  describe('Quick Log', () => {
    it('opens quick log dialog with movement data', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockMovements[0])

      expect(vm.quickLogDialog).toBe(true)
      expect(vm.selectedMovement).toEqual(mockMovements[0])
      expect(vm.quickLogData.date).toBeTruthy()
      expect(vm.quickLogData.name).toContain('Workout')
    })

    it('does not open quick log without movement', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(null)

      expect(vm.quickLogDialog).toBe(false)
    })

    it('closes quick log dialog', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockMovements[0])
      expect(vm.quickLogDialog).toBe(true)

      vm.closeQuickLog()

      expect(vm.quickLogDialog).toBe(false)
      expect(vm.selectedMovement).toBe(null)
    })

    it('submits quick log with movement data', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockMovements[0])
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.movement.sets = 3
      vm.quickLogData.movement.reps = 10
      vm.quickLogData.movement.weight = 135

      await vm.submitQuickLog()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', {
        workout_date: '2024-01-15',
        workout_name: 'Test Workout',
        total_time: null,
        notes: null,
        movements: [
          {
            movement_id: 1,
            sets: 3,
            reps: 10,
            weight: 135,
            time_seconds: null,
            distance: null,
            notes: null
          }
        ],
        wods: []
      })
    })

    it('navigates to dashboard after successful quick log', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockMovements[0])
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.movement.sets = 3

      await vm.submitQuickLog()
      await flushPromises()

      expect(mockPush).toHaveBeenCalledWith('/dashboard')
      expect(vm.quickLogDialog).toBe(false)
    })

    it('handles quick log submission error', async () => {
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue({
        response: { data: { message: 'Failed to create workout' } }
      })

      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockMovements[0])
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.movement.sets = 3

      await vm.submitQuickLog()
      await flushPromises()

      expect(alertSpy).toHaveBeenCalledWith('Failed to create workout')
      expect(vm.quickLogSubmitting).toBe(false)

      alertSpy.mockRestore()
      consoleSpy.mockRestore()
    })

    it('handles generic quick log error', async () => {
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue(new Error('Network error'))

      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockMovements[0])
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.movement.sets = 3

      await vm.submitQuickLog()
      await flushPromises()

      expect(alertSpy).toHaveBeenCalledWith('Failed to log workout')

      alertSpy.mockRestore()
      consoleSpy.mockRestore()
    })

    it('sets loading state during quick log submission', async () => {
      let resolvePost
      axios.post.mockReturnValue(
        new Promise((resolve) => {
          resolvePost = resolve
        })
      )

      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockMovements[0])
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.movement.sets = 3

      const submitPromise = vm.submitQuickLog()

      expect(vm.quickLogSubmitting).toBe(true)

      resolvePost({ data: {} })
      await submitPromise
      await flushPromises()

      expect(vm.quickLogSubmitting).toBe(false)
    })

    it('does not include movement if no performance data', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockMovements[0])
      vm.quickLogData.name = 'Empty Workout'
      vm.quickLogData.date = '2024-01-15'
      // Don't set any movement data

      await vm.submitQuickLog()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', expect.objectContaining({
        movements: []
      }))
    })
  })

  // ==========================================================================
  // DETAIL DIALOG TESTS
  // ==========================================================================

  describe('Detail Dialog', () => {
    it('handles quick log from detail dialog', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.handleDetailQuickLog({ data: mockMovements[0] })

      expect(vm.quickLogDialog).toBe(true)
      expect(vm.selectedMovement).toEqual(mockMovements[0])
    })

    it('does not open quick log if no data from detail dialog', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.handleDetailQuickLog({ data: null })

      expect(vm.quickLogDialog).toBe(false)
    })
  })
})
