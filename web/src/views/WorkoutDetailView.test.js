import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import WorkoutDetailView from './WorkoutDetailView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    delete: vi.fn(),
  }
}))

// Mock vue-router
const mockPush = vi.fn()
const mockBack = vi.fn()
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: mockPush,
      back: mockBack,
    })),
    useRoute: vi.fn(() => ({
      params: { id: '1' }
    }))
  }
})

// Mock settings store
const mockSettingsStore = {
  timezone: 'America/New_York'
}
vi.mock('@/stores/settings', () => ({
  useSettingsStore: () => mockSettingsStore
}))

// Mock timezone utility
vi.mock('@/utils/timezone', () => ({
  formatDateInTimezone: vi.fn((date, tz, format) => {
    if (!date) return ''
    return 'Mon, Jan 15, 2024'
  }),
  getTodayInTimezone: vi.fn(() => '2024-01-16'),
  getYesterdayInTimezone: vi.fn(() => '2024-01-15')
}))

// Mock MarkdownRenderer
vi.mock('@/components/MarkdownRenderer.vue', () => ({
  default: {
    template: '<div class="markdown"></div>',
    props: ['content']
  }
}))

import axios from '@/utils/axios'

describe('WorkoutDetailView', () => {
  let wrapper

  const mockWorkout = {
    id: 1,
    workout_name: 'Morning Workout',
    workout_date: '2024-01-15T00:00:00Z',
    total_time: 3600,
    notes: 'Felt great today',
    performance_movements: [
      {
        id: 1,
        movement: { name: 'Back Squat' },
        movement_type: 'weightlifting',
        sets: 3,
        reps: 10,
        weight: 225,
        is_pr: true,
        time_seconds: null,
        distance: null,
        notes: null
      },
      {
        id: 2,
        movement: { name: 'Pull-ups' },
        movement_type: 'gymnastics',
        sets: 3,
        reps: 12,
        weight: null,
        is_pr: false,
        time_seconds: null,
        distance: null,
        notes: 'Used bands'
      }
    ],
    performance_wods: [
      {
        id: 1,
        wod: { name: 'Fran' },
        score_value: null,
        rounds: null,
        time_seconds: 180,
        notes: 'As prescribed'
      }
    ]
  }

  const mockWorkoutWithoutPR = {
    ...mockWorkout,
    performance_movements: mockWorkout.performance_movements.map(m => ({
      ...m,
      is_pr: false
    }))
  }

  const setupDefaultMocks = () => {
    axios.get.mockResolvedValue({ data: mockWorkout })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(WorkoutDetailView, {
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
          'v-alert': { template: '<div><slot /></div>' },
          'v-progress-circular': { template: '<div></div>' },
          'v-dialog': { template: '<div><slot /></div>' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-spacer': { template: '<div></div>' },
          'v-avatar': { template: '<div><slot /></div>' },
          'v-bottom-navigation': { template: '<div><slot /></div>' },
          'markdown-renderer': true
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    mockPush.mockClear()
    mockBack.mockClear()
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
    it('fetches workout on mount', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/workouts/1')
      consoleSpy.mockRestore()
    })

    it('stores fetched workout', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.workout).toEqual(mockWorkout)
      consoleSpy.mockRestore()
    })

    it('sets loading to false after fetch', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)
      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('handles 404 error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue({ response: { status: 404 } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Workout not found')

      consoleSpy.mockRestore()
    })

    it('handles 403 error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue({ response: { status: 403 } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toContain('permission')

      consoleSpy.mockRestore()
    })

    it('handles generic error with message', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue({
        response: { status: 500, data: { message: 'Server error' } }
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Server error')

      consoleSpy.mockRestore()
    })

    it('handles generic error without message', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue(new Error('Network error'))

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Failed to load workout')

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // COMPUTED PROPERTIES TESTS
  // ==========================================================================

  describe('Computed Properties', () => {
    it('hasPR returns true when workout has PR movement', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.hasPR).toBe(true)
      consoleSpy.mockRestore()
    })

    it('hasPR returns false when workout has no PR movement', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      axios.get.mockResolvedValue({ data: mockWorkoutWithoutPR })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.hasPR).toBe(false)
      consoleSpy.mockRestore()
    })

    it('hasPR returns false when no workout', async () => {
      axios.get.mockResolvedValue({ data: null })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.hasPR).toBe(false)
    })

    it('workoutId returns route param', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.workoutId).toBe('1')
    })
  })

  // ==========================================================================
  // FORMAT TIME TESTS
  // ==========================================================================

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

    it('formats hours, minutes and seconds', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.formatTime(3665)).toBe('1h 1m 5s')
    })

    it('returns empty string for null', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.formatTime(null)).toBe('')
    })

    it('returns empty string for 0', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.formatTime(0)).toBe('')
    })
  })

  // ==========================================================================
  // HELPER FUNCTIONS TESTS
  // ==========================================================================

  describe('Helper Functions', () => {
    describe('getMovementTypeColor', () => {
      it('returns correct colors for types', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeColor('weightlifting')).toBe('#00bcd4')
        expect(vm.getMovementTypeColor('gymnastics')).toBe('#9c27b0')
        expect(vm.getMovementTypeColor('cardio')).toBe('#f44336')
        expect(vm.getMovementTypeColor('bodyweight')).toBe('#4caf50')
        expect(vm.getMovementTypeColor('unknown')).toBe('#666')
        expect(vm.getMovementTypeColor(null)).toBe('#666')
      })
    })

    describe('getMovementTypeIcon', () => {
      it('returns correct icons for types', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeIcon('weightlifting')).toBe('mdi-dumbbell')
        expect(vm.getMovementTypeIcon('gymnastics')).toBe('mdi-gymnastics')
        expect(vm.getMovementTypeIcon('cardio')).toBe('mdi-run')
        expect(vm.getMovementTypeIcon('bodyweight')).toBe('mdi-arm-flex')
        expect(vm.getMovementTypeIcon('unknown')).toBe('mdi-weight-lifter')
        expect(vm.getMovementTypeIcon(null)).toBe('mdi-weight-lifter')
      })
    })
  })

  // ==========================================================================
  // NAVIGATION TESTS
  // ==========================================================================

  describe('Navigation', () => {
    it('navigates to edit workout page', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.editWorkout()

      expect(mockPush).toHaveBeenCalledWith('/workouts/log?edit=1')
      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // DELETE TESTS
  // ==========================================================================

  describe('Delete Workout', () => {
    it('opens delete confirmation dialog', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.confirmDelete()

      expect(vm.deleteDialog).toBe(true)
      consoleSpy.mockRestore()
    })

    it('deletes workout successfully', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      axios.delete.mockResolvedValue({})
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.deleteDialog = true

      await vm.deleteWorkout()
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/workouts/1')
      expect(mockPush).toHaveBeenCalledWith('/dashboard')
      consoleSpy.mockRestore()
    })

    it('handles delete error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      axios.delete.mockRejectedValue({
        response: { data: { message: 'Cannot delete workout' } }
      })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.deleteDialog = true

      await vm.deleteWorkout()
      await flushPromises()

      expect(vm.error).toBe('Cannot delete workout')
      expect(vm.deleteDialog).toBe(false)

      consoleSpy.mockRestore()
      logSpy.mockRestore()
    })

    it('sets deleting state during deletion', async () => {
      const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      let resolveDelete
      axios.delete.mockReturnValue(new Promise(resolve => { resolveDelete = resolve }))
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const deletePromise = vm.deleteWorkout()

      expect(vm.deleting).toBe(true)

      resolveDelete({})
      await deletePromise
      await flushPromises()

      expect(vm.deleting).toBe(false)
      logSpy.mockRestore()
    })
  })

  // ==========================================================================
  // FORMAT DATE TESTS
  // ==========================================================================

  describe('formatDate', () => {
    it('formats date using timezone utility', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm
      const result = vm.formatDate('2024-01-15T00:00:00Z')

      // The mock returns 'Mon, Jan 15, 2024'
      expect(result).toBeTruthy()
    })
  })
})
