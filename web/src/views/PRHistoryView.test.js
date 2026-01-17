import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PRHistoryView from './PRHistoryView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
  }
}))

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
    return 'Fri, Jan 15, 2024'
  })
}))

import axios from '@/utils/axios'

describe('PRHistoryView', () => {
  let wrapper

  const mockMovementPRs = [
    {
      id: 1,
      movement_name: 'Back Squat',
      workout_date: '2024-01-15',
      sets: 3,
      reps: 5,
      weight: 225
    },
    {
      id: 2,
      movement_name: 'Back Squat',
      workout_date: '2024-01-10',
      sets: 3,
      reps: 3,
      weight: 200
    },
    {
      id: 3,
      movement_name: 'Deadlift',
      workout_date: '2024-01-12',
      sets: 1,
      reps: 1,
      weight: 315
    }
  ]

  const mockWODPRs = [
    {
      id: 1,
      wod_name: 'Fran',
      workout_date: '2024-01-15',
      time_seconds: 180
    },
    {
      id: 2,
      wod_name: 'Murph',
      workout_date: '2024-01-10',
      time_seconds: 3600,
      rounds: null
    }
  ]

  const setupDefaultMocks = () => {
    axios.get.mockResolvedValue({
      data: { pr_movements: [], pr_wods: [] }
    })
  }

  const setupMocksWithData = () => {
    axios.get.mockResolvedValue({
      data: { pr_movements: mockMovementPRs, pr_wods: mockWODPRs }
    })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(PRHistoryView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-progress-circular': { template: '<div class="loading"></div>' },
          'v-icon': { template: '<i></i>' },
          'v-text-field': { template: '<input />' }
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
    it('fetches personal records on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/workouts/personal-records')
    })

    it('shows loading state initially', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.loading).toBe(true)
    })

    it('initializes with empty search query', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.searchQuery).toBe('')
    })

    it('stores PR data after fetch', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.movementPRs).toHaveLength(3)
      expect(vm.wodPRs).toHaveLength(2)
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
    it('handles fetch error gracefully', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue(new Error('Network error'))

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)
      expect(vm.movementPRs).toEqual([])

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // FILTERING TESTS
  // ==========================================================================

  describe('Filtering', () => {
    it('filters movement PRs by search query', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.searchQuery = 'squat'

      expect(vm.filteredMovementPRs).toHaveLength(2)
      expect(vm.filteredMovementPRs.every(pr => pr.movement_name.toLowerCase().includes('squat'))).toBe(true)
    })

    it('filters WOD PRs by search query', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.searchQuery = 'fran'

      expect(vm.filteredWODPRs).toHaveLength(1)
      expect(vm.filteredWODPRs[0].wod_name).toBe('Fran')
    })

    it('returns all PRs when search query is empty', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.searchQuery = ''

      expect(vm.filteredMovementPRs).toHaveLength(3)
      expect(vm.filteredWODPRs).toHaveLength(2)
    })

    it('filters case-insensitively', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.searchQuery = 'SQUAT'

      expect(vm.filteredMovementPRs).toHaveLength(2)
    })
  })

  // ==========================================================================
  // GROUPING TESTS
  // ==========================================================================

  describe('Grouping', () => {
    it('groups movement PRs by name', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const groups = vm.groupedMovementPRs

      expect(Object.keys(groups)).toContain('Back Squat')
      expect(Object.keys(groups)).toContain('Deadlift')
      expect(groups['Back Squat']).toHaveLength(2)
      expect(groups['Deadlift']).toHaveLength(1)
    })

    it('groups WOD PRs by name', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const groups = vm.groupedWODPRs

      expect(Object.keys(groups)).toContain('Fran')
      expect(Object.keys(groups)).toContain('Murph')
    })

    it('uses Unknown for PRs without names', async () => {
      axios.get.mockResolvedValue({
        data: {
          pr_movements: [{ id: 1, movement_name: null, weight: 100 }],
          pr_wods: []
        }
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const groups = vm.groupedMovementPRs

      expect(Object.keys(groups)).toContain('Unknown')
    })
  })

  // ==========================================================================
  // FORMAT FUNCTIONS TESTS
  // ==========================================================================

  describe('Format Functions', () => {
    describe('formatPerformance', () => {
      it('formats sets and reps', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatPerformance({ sets: 3, reps: 10 })

        expect(result).toContain('3 x')
        expect(result).toContain('10')
      })

      it('formats weight', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatPerformance({ weight: 225 })

        expect(result).toContain('225 lbs')
      })

      it('formats time in seconds', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatPerformance({ time_seconds: 125 })

        expect(result).toContain('2:05')
      })

      it('returns Completed for empty performance', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatPerformance({})

        expect(result).toBe('Completed')
      })
    })

    describe('formatWODPerformance', () => {
      it('formats time in seconds', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatWODPerformance({ time_seconds: 180 })

        expect(result).toBe('3:00')
      })

      it('formats rounds only', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatWODPerformance({ rounds: 10 })

        expect(result).toContain('10 rounds')
      })

      it('formats rounds and reps', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatWODPerformance({ rounds: 10, reps: 15 })

        expect(result).toContain('10 rounds + 15 reps')
      })

      it('formats weight', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatWODPerformance({ weight: 225 })

        expect(result).toContain('225 lbs')
      })

      it('returns score_value or Completed for empty', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.formatWODPerformance({ score_value: '100 points' })).toBe('100 points')
        expect(vm.formatWODPerformance({})).toBe('Completed')
      })
    })

    describe('formatDate', () => {
      it('returns empty string for null date', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatDate(null)

        expect(result).toBe('')
      })

      it('formats valid date', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatDate('2024-01-15')

        expect(result).toBe('Fri, Jan 15, 2024')
      })
    })
  })
})
