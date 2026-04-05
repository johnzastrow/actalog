import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import WorkoutCalendarView from './WorkoutCalendarView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
  }
}))

// Mock WorkoutCalendar component
vi.mock('@/components/WorkoutCalendar.vue', () => ({
  default: {
    template: '<div class="workout-calendar"></div>',
    props: ['workoutDates']
  }
}))

// Mock DateWorkoutDetailDialog component
vi.mock('@/components/DateWorkoutDetailDialog.vue', () => ({
  default: {
    template: '<div class="date-dialog"></div>',
    props: ['modelValue', 'date', 'preloadedWorkouts']
  }
}))

import axios from '@/utils/axios'

describe('WorkoutCalendarView', () => {
  let wrapper

  const mockWorkouts = [
    { id: 1, workout_date: '2024-01-15', workout_name: 'Morning Workout' },
    { id: 2, workout_date: '2024-01-16', workout_name: 'Evening Workout' },
    { id: 3, workout_date: '2024-01-15', workout_name: 'Second Workout' }
  ]

  const setupDefaultMocks = () => {
    axios.get.mockResolvedValue({ data: { workouts: mockWorkouts } })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(WorkoutCalendarView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-card': { template: '<div><slot /></div>' },
          'v-icon': { template: '<i></i>' },
          'v-bottom-navigation': { template: '<div><slot /></div>' },
          'v-btn': { template: '<button><slot /></button>' },
          'workout-calendar': true,
          'DateWorkoutDetailDialog': true
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
    it('fetches workouts on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/workouts?limit=10000')
    })

    it('stores fetched workouts', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.workouts).toEqual(mockWorkouts)
    })

    it('extracts workout dates', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.workoutDates).toEqual(['2024-01-15', '2024-01-16', '2024-01-15'])
    })

    it('handles array response format', async () => {
      axios.get.mockResolvedValue({ data: mockWorkouts })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.workouts).toEqual(mockWorkouts)
    })

    it('initializes with dialog closed', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.showDateDialog).toBe(false)
      expect(vm.selectedDate).toBe(null)
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

      expect(vm.workouts).toEqual([])

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // DAY SELECTION TESTS
  // ==========================================================================

  describe('Day Selection', () => {
    it('opens dialog when day is selected', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.handleDaySelected('2024-01-15')

      expect(vm.selectedDate).toBe('2024-01-15')
      expect(vm.showDateDialog).toBe(true)
    })
  })
})
