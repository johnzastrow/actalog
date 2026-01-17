import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import WorkoutTimelineView from './WorkoutTimelineView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
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
    }))
  }
})

import axios from '@/utils/axios'

describe('WorkoutTimelineView', () => {
  let wrapper

  const mockWorkouts = [
    {
      id: 1,
      workout_date: '2024-01-15T00:00:00Z',
      workout_name: 'Morning Workout',
      workout_type: 'strength',
      total_time: 3600,
      notes: 'Felt great today',
      movements: [
        { id: 1, movement_id: 1, movement_name: 'Back Squat', is_pr: true }
      ],
      wods: []
    },
    {
      id: 2,
      workout_date: '2024-01-15T00:00:00Z',
      workout_name: 'Evening WOD',
      workout_type: 'cardio',
      total_time: 1800,
      notes: null,
      movements: [],
      wods: [
        { id: 1, wod_name: 'Fran', is_pr: false }
      ]
    },
    {
      id: 3,
      workout_date: '2024-01-14T00:00:00Z',
      workout_name: 'Yesterday Workout',
      workout_type: 'strength',
      total_time: null,
      notes: null,
      movements: [
        { id: 2, movement_id: 2, movement_name: 'Deadlift', is_pr: false },
        { id: 3, movement_id: 3, movement_name: 'Pull-ups', is_pr: false },
        { id: 4, movement_id: 4, movement_name: 'Push-ups', is_pr: false },
        { id: 5, movement_id: 5, movement_name: 'Squats', is_pr: false }
      ],
      wods: []
    }
  ]

  const mockMovements = [
    { id: 1, name: 'Back Squat' },
    { id: 2, name: 'Deadlift' },
    { id: 3, name: 'Pull-ups' }
  ]

  const setupDefaultMocks = () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/workouts') {
        return Promise.resolve({ data: { workouts: mockWorkouts } })
      }
      if (url === '/api/movements') {
        return Promise.resolve({ data: { movements: mockMovements } })
      }
      return Promise.resolve({ data: {} })
    })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(WorkoutTimelineView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-card': { template: '<div><slot /></div>' },
          'v-card-text': { template: '<div><slot /></div>' },
          'v-icon': { template: '<i></i>' },
          'v-chip': { template: '<span><slot /></span>' },
          'v-select': { template: '<select></select>' },
          'v-divider': { template: '<hr />' },
          'v-progress-circular': { template: '<div></div>' },
          'v-bottom-navigation': { template: '<div><slot /></div>' },
          'v-btn': { template: '<button><slot /></button>' }
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
    it('fetches workouts and movements on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/workouts')
      expect(axios.get).toHaveBeenCalledWith('/api/movements')
    })

    it('stores fetched data', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.workouts).toHaveLength(3)
      expect(vm.movements).toHaveLength(3)
    })

    it('sets loading to false after fetch', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)
    })

    it('initializes with null filters', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.filterType).toBe(null)
      expect(vm.filterMovement).toBe(null)
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('handles fetch error gracefully', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue(new Error('Network error'))

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // COMPUTED PROPERTIES TESTS
  // ==========================================================================

  describe('Computed Properties', () => {
    describe('typeOptions', () => {
      it('extracts unique workout types', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.typeOptions).toContain('strength')
        expect(vm.typeOptions).toContain('cardio')
        expect(vm.typeOptions).toHaveLength(2)
      })
    })

    describe('movementOptions', () => {
      it('maps movements to options format', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.movementOptions).toHaveLength(3)
        expect(vm.movementOptions[0]).toEqual({ title: 'Back Squat', value: 1 })
      })
    })

    describe('filteredWorkouts', () => {
      it('returns all workouts when no filters', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.filteredWorkouts).toHaveLength(3)
      })

      it('filters by workout type', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.filterType = 'strength'

        expect(vm.filteredWorkouts).toHaveLength(2)
        expect(vm.filteredWorkouts.every(w => w.workout_type === 'strength')).toBe(true)
      })

      it('filters by movement', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.filterMovement = 1

        expect(vm.filteredWorkouts).toHaveLength(1)
        expect(vm.filteredWorkouts[0].id).toBe(1)
      })

      it('sorts by date descending', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.filteredWorkouts[0].workout_date).toBe('2024-01-15T00:00:00Z')
      })
    })

    describe('groupedWorkouts', () => {
      it('groups workouts by date', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        const groups = vm.groupedWorkouts

        expect(Object.keys(groups)).toContain('2024-01-15')
        expect(Object.keys(groups)).toContain('2024-01-14')
        expect(groups['2024-01-15']).toHaveLength(2)
        expect(groups['2024-01-14']).toHaveLength(1)
      })
    })
  })

  // ==========================================================================
  // FORMAT FUNCTIONS TESTS
  // ==========================================================================

  describe('Format Functions', () => {
    describe('formatDate', () => {
      it('formats date correctly', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        const result = vm.formatDate('2024-01-10')

        expect(result).toBeTruthy()
      })
    })

    describe('formatTime', () => {
      it('formats seconds only', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.formatTime(45)).toBe('45s')
      })

      it('formats minutes and seconds', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.formatTime(125)).toBe('2m 5s')
      })

      it('formats hours and minutes', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.formatTime(3660)).toBe('1h 1m')
      })
    })

    describe('truncateNotes', () => {
      it('returns short notes unchanged', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.truncateNotes('Short note')).toBe('Short note')
      })

      it('truncates long notes', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const longNote = 'A'.repeat(150)
        const result = vm.truncateNotes(longNote)

        expect(result.length).toBe(103) // 100 + '...'
        expect(result.endsWith('...')).toBe(true)
      })
    })
  })

  // ==========================================================================
  // NAVIGATION TESTS
  // ==========================================================================

  describe('Navigation', () => {
    it('navigates to workout detail', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.viewWorkout(1)

      expect(mockPush).toHaveBeenCalledWith('/workouts/1')
    })
  })
})
