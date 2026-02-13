import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PerformanceView from './PerformanceView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  }
}))

// Mock vue-router
const mockPush = vi.fn()
const mockReplace = vi.fn()
let mockRouteQuery = {}
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: mockPush,
      replace: mockReplace
    })),
    useRoute: vi.fn(() => ({
      query: mockRouteQuery
    }))
  }
})

// Mock vuetify theme
vi.mock('vuetify', async () => {
  const actual = await vi.importActual('vuetify')
  return {
    ...actual,
    useTheme: () => ({
      current: {
        value: {
          colors: {
            primary: '#00bcd4',
            warning: '#ffc107'
          }
        }
      }
    })
  }
})

// Mock chart.js
vi.mock('chart.js', () => {
  class MockChart {
    constructor() {
      this.destroy = vi.fn()
    }
    static register() {}
  }
  return {
    Chart: MockChart,
    registerables: []
  }
})

// Mock settings store
vi.mock('@/stores/settings', () => ({
  useSettingsStore: () => ({
    timezone: 'America/New_York'
  })
}))

// Mock timezone utils
vi.mock('@/utils/timezone', () => ({
  formatDateInTimezone: (date, tz, format) => {
    const d = new Date(date)
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  },
  getTodayInTimezone: () => new Date().toISOString().split('T')[0],
  getYesterdayInTimezone: () => {
    const d = new Date()
    d.setDate(d.getDate() - 1)
    return d.toISOString().split('T')[0]
  }
}))

// Mock RPE utils
vi.mock('@/utils/rpe', () => ({
  getRPEColor: (rpe) => rpe >= 9 ? 'red' : rpe >= 7 ? 'orange' : 'green',
  getRPEShortLabel: (rpe) => `RPE ${rpe}`
}))

import axios from '@/utils/axios'

describe('PerformanceView', () => {
  let wrapper

  const mockMovementPerformance = {
    performances: [
      {
        id: 1,
        movement_id: 1,
        user_workout_id: 100,
        workout_date: '2024-01-15T00:00:00Z',
        sets: 5,
        reps: 5,
        weight: 225,
        time_seconds: null,
        notes: 'Felt good',
        rpe: 8,
        is_pr: false,
        calculated_1rm: 253
      },
      {
        id: 2,
        movement_id: 1,
        user_workout_id: 101,
        workout_date: '2024-01-10T00:00:00Z',
        sets: 3,
        reps: 10,
        weight: 185,
        time_seconds: null,
        notes: null,
        rpe: 7,
        is_pr: false,
        calculated_1rm: 247
      },
      {
        id: 3,
        movement_id: 1,
        user_workout_id: 102,
        workout_date: '2024-01-08T00:00:00Z',
        sets: 1,
        reps: 1,
        weight: 275,
        time_seconds: null,
        notes: 'PR!',
        rpe: 10,
        is_pr: true,
        calculated_1rm: 275
      }
    ],
    best_1rm: 275,
    best_formula: 'Brzycki'
  }

  const mockWODPerformance = {
    performances: [
      {
        id: 1,
        wod_id: 1,
        user_workout_id: 200,
        workout_date: '2024-01-15T00:00:00Z',
        time_seconds: 330,
        rounds: null,
        reps: null,
        weight: null,
        score_value: '5:30',
        notes: 'Good pace',
        rpe: 9,
        is_pr: true,
        wod_score_type: 'Time (HH:MM:SS)'
      },
      {
        id: 2,
        wod_id: 1,
        user_workout_id: 201,
        workout_date: '2024-01-10T00:00:00Z',
        time_seconds: 420,
        rounds: null,
        reps: null,
        weight: null,
        score_value: '7:00',
        notes: null,
        rpe: 8,
        is_pr: false,
        wod_score_type: 'Time (HH:MM:SS)'
      }
    ]
  }

  const mockSearchResults = {
    results: [
      { type: 'movement', id: 1, name: 'Back Squat', data: { id: 1, name: 'Back Squat', type: 'weightlifting' } },
      { type: 'wod', id: 1, name: 'Fran', data: { id: 1, name: 'Fran', score_type: 'Time (HH:MM:SS)' } }
    ]
  }

  const mockMovements = [
    { id: 1, name: 'Back Squat' },
    { id: 2, name: 'Deadlift' }
  ]

  const mockWODs = [
    { id: 1, name: 'Fran', score_type: 'Time (HH:MM:SS)' },
    { id: 2, name: 'Cindy', score_type: 'Rounds+Reps' }
  ]

  const setupDefaultMocks = () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/performance/search') {
        return Promise.resolve({ data: mockSearchResults })
      }
      if (url.startsWith('/api/performance/movements/')) {
        return Promise.resolve({ data: mockMovementPerformance })
      }
      if (url.startsWith('/api/performance/wods/')) {
        return Promise.resolve({ data: mockWODPerformance })
      }
      if (url === '/api/movements') {
        return Promise.resolve({ data: { movements: mockMovements } })
      }
      if (url === '/api/wods/standard') {
        return Promise.resolve({ data: { wods: mockWODs } })
      }
      if (url === '/api/wods/my-wods') {
        return Promise.resolve({ data: { wods: [] } })
      }
      if (url === '/api/workouts/standard' || url === '/api/workouts/my-templates') {
        return Promise.resolve({ data: { workouts: [] } })
      }
      return Promise.resolve({ data: {} })
    })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(PerformanceView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-autocomplete': { template: '<select></select>' },
          'v-select': { template: '<select></select>' },
          'v-text-field': { template: '<input />' },
          'v-textarea': { template: '<textarea></textarea>' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-icon': { template: '<i></i>' },
          'v-chip': { template: '<span><slot /></span>' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-list-item': { template: '<div><slot /></div>' },
          'v-list-item-title': { template: '<div><slot /></div>' },
          'v-list-item-subtitle': { template: '<div><slot /></div>' },
          'v-progress-circular': { template: '<div></div>' },
          'v-expansion-panels': { template: '<div><slot /></div>' },
          'v-expansion-panel': { template: '<div><slot /></div>' },
          'v-expansion-panel-title': { template: '<div><slot /></div>' },
          'v-expansion-panel-text': { template: '<div><slot /></div>' },
          'v-table': { template: '<table><slot /></table>' },
          'v-dialog': { template: '<div><slot /></div>' },
          'v-form': { template: '<form><slot /></form>' },
          'v-spacer': { template: '<div></div>' },
          'v-bottom-navigation': { template: '<div><slot /></div>' },
          'v-avatar': { template: '<div><slot /></div>' }
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    mockPush.mockClear()
    mockReplace.mockClear()
    mockRouteQuery = {}
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
    it('initializes with no selection', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.selectedItem).toBe(null)
      expect(vm.performanceData).toEqual([])
    })

    it('restores selection from URL query params', async () => {
      mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.selectedItem).not.toBe(null)
      expect(vm.selectedItem.type).toBe('movement')
      expect(vm.selectedItem.name).toBe('Back Squat')
    })

    it('fetches performance data when selection restored', async () => {
      mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/performance/movements/1')
    })
  })

  // ==========================================================================
  // SEARCH TESTS
  // ==========================================================================

  describe('Search', () => {
    it('does not search with short query', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.handleSearch('a')
      await flushPromises()

      expect(axios.get).not.toHaveBeenCalledWith(expect.stringContaining('/api/performance/search'))
    })

    it('performs unified search with valid query', async () => {
      vi.useFakeTimers()
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.handleSearch('squat')

      vi.advanceTimersByTime(300)
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/performance/search', {
        params: { q: 'squat', limit: 20 }
      })

      vi.useRealTimers()
    })

    it('populates search results', async () => {
      vi.useFakeTimers()
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.handleSearch('squat')

      vi.advanceTimersByTime(300)
      await flushPromises()

      expect(vm.searchResults).toHaveLength(2)

      vi.useRealTimers()
    })
  })

  // ==========================================================================
  // SELECTION TESTS
  // ==========================================================================

  describe('Selection', () => {
    it('handles movement selection', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const item = { type: 'movement', id: 'movement-1', name: 'Back Squat', data: { id: 1 } }
      await vm.handleSelection(item)
      await flushPromises()

      expect(vm.selectedItem).toEqual(item)
      expect(axios.get).toHaveBeenCalledWith('/api/performance/movements/1')
    })

    it('handles WOD selection', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const item = { type: 'wod', id: 'wod-1', name: 'Fran', data: { id: 1 } }
      await vm.handleSelection(item)
      await flushPromises()

      expect(vm.selectedItem).toEqual(item)
      expect(axios.get).toHaveBeenCalledWith('/api/performance/wods/1')
    })

    it('updates URL query on selection', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const item = { type: 'movement', id: 'movement-1', name: 'Back Squat', data: { id: 1 } }
      await vm.handleSelection(item)
      await flushPromises()

      expect(mockReplace).toHaveBeenCalledWith({
        query: { type: 'movement', id: 1, name: 'Back Squat' }
      })
    })

    it('clears selection', async () => {
      mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.clearSelection()

      expect(vm.selectedItem).toBe(null)
      expect(vm.performanceData).toEqual([])
      expect(vm.best1RM).toBe(null)
    })
  })

  // ==========================================================================
  // COMPUTED PROPERTIES TESTS
  // ==========================================================================

  describe('Computed Properties', () => {
    describe('heaviestLifts', () => {
      it('returns empty for non-movement', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.selectedItem = { type: 'wod', data: { id: 1 } }

        expect(vm.heaviestLifts).toEqual([])
      })

      it('returns top 3 heaviest weights', async () => {
        mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.heaviestLifts).toHaveLength(3)
        expect(vm.heaviestLifts[0].weight).toBe(275)
      })
    })

    describe('heaviestWeight', () => {
      it('returns null when no lifts', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.heaviestWeight).toBe(null)
      })

      it('returns heaviest weight', async () => {
        mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.heaviestWeight).toBe(275)
      })
    })

    describe('showPercentOfBest', () => {
      it('returns falsy when no weight data', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.showPercentOfBest).toBeFalsy()
      })

      it('returns true when weight data exists', async () => {
        mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.showPercentOfBest).toBe(true)
      })
    })

    describe('percentages', () => {
      it('calculates percentage weights', async () => {
        mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        const percentages = vm.percentages

        expect(percentages.length).toBeGreaterThan(0)
        // 95% of 275 = 261.25 -> 261
        expect(percentages.find(p => p.percent === 95).weight).toBe(261)
        // 50% of 275 = 137.5 -> 138
        expect(percentages.find(p => p.percent === 50).weight).toBe(138)
      })
    })

    describe('bestWODPerformances', () => {
      it('returns empty for non-WOD', async () => {
        mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.bestWODPerformances).toEqual([])
      })

      it('returns top 3 for time-based WOD', async () => {
        mockRouteQuery = { type: 'wod', id: '1', name: 'Fran' }
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        // Should be sorted by time (lower is better)
        expect(vm.bestWODPerformances.length).toBeGreaterThan(0)
        expect(vm.bestWODPerformances[0].time_seconds).toBe(330)
      })
    })

    describe('filteredChartData', () => {
      it('returns all weight data when filter is All', async () => {
        mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.selectedRepScheme = 'All'

        expect(vm.filteredChartData).toHaveLength(3)
      })

      it('filters by rep scheme', async () => {
        mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.selectedRepScheme = '5 reps'

        expect(vm.filteredChartData).toHaveLength(1)
        expect(vm.filteredChartData[0].reps).toBe(5)
      })
    })

    describe('groupedHistory', () => {
      it('groups by year', async () => {
        mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        const groups = vm.groupedHistory

        expect(groups.length).toBeGreaterThan(0)
        expect(groups[0].year).toBe('2024')
        expect(groups[0].entries.length).toBe(3)
      })

      it('sorts entries by date descending', async () => {
        mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        const entries = vm.groupedHistory[0].entries

        // Most recent first
        expect(new Date(entries[0].workout_date) >= new Date(entries[1].workout_date)).toBe(true)
      })
    })
  })

  // ==========================================================================
  // FORMAT FUNCTIONS TESTS
  // ==========================================================================

  describe('Format Functions', () => {
    describe('formatWODScore', () => {
      it('formats time_seconds', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.formatWODScore({ time_seconds: 330 })).toBe('5:30')
      })

      it('formats rounds+reps', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.formatWODScore({ rounds: 10, reps: 15 })).toBe('10+15')
      })

      it('formats rounds only', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.formatWODScore({ rounds: 10, reps: null })).toBe('10 rounds')
      })

      it('formats weight', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        // Must include rounds: null and reps: null to skip the rounds+reps check
        expect(vm.formatWODScore({ weight: 225, rounds: null, reps: null })).toBe('225 lbs')
      })

      it('returns N/A for empty', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        // Must include null values to skip checks that use !== null
        expect(vm.formatWODScore({ rounds: null, reps: null, weight: null, score_value: null })).toBe('N/A')
      })
    })

    describe('formatTime', () => {
      it('formats seconds only', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.formatTime(45)).toBe('45s')
      })

      it('formats minutes and seconds', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.formatTime(330)).toBe('5:30')
      })

      it('formats hours and minutes', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.formatTime(3660)).toBe('1h 1m')
      })

      it('returns empty for null', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.formatTime(null)).toBe('')
      })
    })

    describe('formatMovementType', () => {
      it('capitalizes type', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.formatMovementType('weightlifting')).toBe('Weightlifting')
      })

      it('returns empty for null', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.formatMovementType(null)).toBe('')
      })
    })
  })

  // ==========================================================================
  // QUICK LOG TESTS
  // ==========================================================================

  describe('Quick Log', () => {
    it('opens quick log dialog', async () => {
      mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.quickLog()
      await flushPromises()

      expect(vm.quickLogDialog).toBe(true)
    })

    it('pre-populates with selected movement', async () => {
      mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.quickLog()
      await flushPromises()

      expect(vm.quickLogData.selectedItem).not.toBe(null)
      expect(vm.quickLogData.selectedItem.type).toBe('movement')
    })

    it('pre-populates with selected WOD', async () => {
      mockRouteQuery = { type: 'wod', id: '1', name: 'Fran' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.quickLog()
      await flushPromises()

      expect(vm.quickLogData.selectedItem).not.toBe(null)
      expect(vm.quickLogData.selectedItem.type).toBe('wod')
    })

    it('does not open without selection', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.quickLog()

      expect(vm.quickLogDialog).toBe(false)
    })

    it('closes quick log dialog', async () => {
      mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.quickLog()
      await flushPromises()
      expect(vm.quickLogDialog).toBe(true)

      vm.closeQuickLog()

      expect(vm.quickLogDialog).toBe(false)
    })

    it('submits quick log with movement', async () => {
      mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
      axios.post.mockResolvedValue({ data: {} })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.quickLog()
      await flushPromises()

      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.movement.sets = 5
      vm.quickLogData.movement.reps = 5
      vm.quickLogData.movement.weight = 225

      await vm.submitQuickLog()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', expect.objectContaining({
        workout_name: 'Test Workout',
        workout_date: '2024-01-15',
        movements: expect.arrayContaining([
          expect.objectContaining({
            movement_id: 1,
            sets: 5,
            reps: 5,
            weight: 225
          })
        ])
      }))
    })

    it('handles quick log submission error', async () => {
      mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue({
        response: { data: { message: 'Failed to log' } }
      })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.quickLog()
      await flushPromises()

      vm.quickLogData.name = 'Test'
      vm.quickLogData.date = '2024-01-15'

      await vm.submitQuickLog()
      await flushPromises()

      expect(alertSpy).toHaveBeenCalledWith('Failed to log')

      alertSpy.mockRestore()
      consoleSpy.mockRestore()
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
      vm.viewWorkout(100)

      expect(mockPush).toHaveBeenCalledWith('/workouts/100')
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('handles search error', async () => {
      vi.useFakeTimers()
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockImplementation((url) => {
        if (url === '/api/performance/search') {
          return Promise.reject(new Error('Search failed'))
        }
        return Promise.resolve({ data: {} })
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.handleSearch('squat')

      vi.advanceTimersByTime(300)
      await flushPromises()

      expect(vm.error).toContain('Failed to search')
      expect(vm.searchResults).toEqual([])

      vi.useRealTimers()
      consoleSpy.mockRestore()
    })

    it('handles performance fetch error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockImplementation((url) => {
        if (url.startsWith('/api/performance/movements/')) {
          return Promise.reject(new Error('Fetch failed'))
        }
        return Promise.resolve({ data: {} })
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedItem = { type: 'movement', data: { id: 1 } }
      await vm.fetchPerformanceData()
      await flushPromises()

      expect(vm.error).toContain('Failed to load')
      expect(vm.performanceData).toEqual([])

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // HELPER FUNCTIONS TESTS
  // ==========================================================================

  describe('Helper Functions', () => {
    describe('getTodayDate', () => {
      it('returns today in YYYY-MM-DD format', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        const result = vm.getTodayDate()

        expect(result).toMatch(/^\d{4}-\d{2}-\d{2}$/)
      })
    })

    describe('formatQuickLogName', () => {
      it('returns day-based name', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        const result = vm.formatQuickLogName('2024-01-15') // Monday

        expect(result).toContain('Workout')
      })

      it('returns Workout for null date', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.formatQuickLogName(null)).toBe('Workout')
      })
    })

    describe('updateQuickLogName', () => {
      it('updates name based on date', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.quickLogData.date = '2024-01-15'
        vm.updateQuickLogName()

        expect(vm.quickLogData.name).toContain('Workout')
      })
    })
  })

  // ==========================================================================
  // UNIFIED SEARCH ITEMS TESTS
  // ==========================================================================

  describe('Unified Search Items', () => {
    it('combines movements and WODs', async () => {
      mockRouteQuery = { type: 'movement', id: '1', name: 'Back Squat' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.quickLog()
      await flushPromises()

      expect(vm.unifiedSearchItems.length).toBeGreaterThan(0)
    })
  })
})
