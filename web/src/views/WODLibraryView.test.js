import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import WODLibraryView from './WODLibraryView.vue'

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

// Mock wods store
const mockWods = [
  {
    id: 1,
    name: 'Fran',
    description: '21-15-9 Thrusters and Pull-ups',
    type: 'Girl',
    regime: 'For Time',
    score_type: 'Time (HH:MM:SS)',
    is_standard: true
  },
  {
    id: 2,
    name: 'Murph',
    description: '1 mile run, 100 pull-ups, 200 push-ups, 300 squats, 1 mile run',
    type: 'Hero',
    regime: 'For Time',
    score_type: 'Time (HH:MM:SS)',
    is_standard: true
  },
  {
    id: 3,
    name: 'Fight Gone Bad',
    description: '3 rounds of: Wall balls, SDHP, Box jumps, Push press, Row',
    type: 'Benchmark',
    regime: 'AMRAP',
    score_type: 'Rounds+Reps',
    is_standard: true
  },
  {
    id: 4,
    name: 'Ringer 1',
    description: 'CrossFit Games event',
    type: 'Games',
    regime: 'For Max Weight',
    score_type: 'Max Weight',
    is_standard: true
  },
  {
    id: 5,
    name: 'My Custom WOD',
    description: 'Custom workout',
    type: 'Girl',
    regime: 'For Time',
    score_type: 'Time (HH:MM:SS)',
    is_standard: false
  }
]

const mockWodsStore = {
  wods: [],
  loading: false,
  error: null,
  fetchWods: vi.fn(),
  filterByType: vi.fn((type) => mockWods.filter(w => w.type === type))
}

vi.mock('@/stores/wods', () => ({
  useWodsStore: () => mockWodsStore
}))

// Mock DetailViewDialog component
vi.mock('@/components/DetailViewDialog.vue', () => ({
  default: {
    template: '<div class="detail-view-dialog"></div>',
    props: ['modelValue', 'type', 'id', 'handleQuickLogExternally']
  }
}))

import axios from '@/utils/axios'
import { useRouter, useRoute } from 'vue-router'

describe('WODLibraryView', () => {
  let wrapper

  const setupDefaultStore = () => {
    mockWodsStore.wods = []
    mockWodsStore.loading = false
    mockWodsStore.error = null
    mockWodsStore.fetchWods.mockResolvedValue()
  }

  const setupStoreWithData = () => {
    mockWodsStore.wods = mockWods
    mockWodsStore.loading = false
    mockWodsStore.error = null
    mockWodsStore.fetchWods.mockResolvedValue()
  }

  const createWrapper = (routeQuery = {}) => {
    const pinia = createPinia()
    setActivePinia(pinia)

    // Setup route mock with query
    useRoute.mockReturnValue({
      query: routeQuery
    })

    wrapper = shallowMount(WODLibraryView, {
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
          'DetailViewDialog': true
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    mockPush.mockClear()
    setupDefaultStore()
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
    it('fetches WODs on mount', async () => {
      setupDefaultStore()
      createWrapper()
      await flushPromises()

      expect(mockWodsStore.fetchWods).toHaveBeenCalled()
    })

    it('initializes with default state', () => {
      setupDefaultStore()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.searchQuery).toBe('')
      expect(vm.selectedType).toBe('all')
    })

    it('shows loading state from store', async () => {
      mockWodsStore.loading = true
      createWrapper()

      const vm = wrapper.vm

      expect(vm.loading).toBe(true)
    })

    it('shows error from store', async () => {
      mockWodsStore.error = 'Failed to load WODs'
      createWrapper()

      expect(mockWodsStore.error).toBe('Failed to load WODs')
    })
  })

  // ==========================================================================
  // FILTERING TESTS
  // ==========================================================================

  describe('Filtering', () => {
    it('shows all WODs when type is all', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedType = 'all'

      expect(vm.filteredWODs).toHaveLength(5)
    })

    it('filters WODs by type - Girl', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedType = 'Girl'

      // filteredWODs should call filterByType
      expect(vm.filteredWODs).toHaveLength(2)
      expect(vm.filteredWODs.every(w => w.type === 'Girl')).toBe(true)
    })

    it('filters WODs by type - Hero', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedType = 'Hero'

      expect(vm.filteredWODs).toHaveLength(1)
      expect(vm.filteredWODs[0].name).toBe('Murph')
    })

    it('filters WODs by type - Benchmark', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedType = 'Benchmark'

      expect(vm.filteredWODs).toHaveLength(1)
      expect(vm.filteredWODs[0].name).toBe('Fight Gone Bad')
    })

    it('filters WODs by type - Games', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedType = 'Games'

      expect(vm.filteredWODs).toHaveLength(1)
      expect(vm.filteredWODs[0].name).toBe('Ringer 1')
    })

    it('filters WODs by search query - name match', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.searchQuery = 'fran'

      expect(vm.filteredWODs).toHaveLength(1)
      expect(vm.filteredWODs[0].name).toBe('Fran')
    })

    it('filters WODs by search query - description match', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.searchQuery = 'thrusters'

      expect(vm.filteredWODs).toHaveLength(1)
      expect(vm.filteredWODs[0].name).toBe('Fran')
    })

    it('filters WODs by search query - case insensitive', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.searchQuery = 'MURPH'

      expect(vm.filteredWODs).toHaveLength(1)
      expect(vm.filteredWODs[0].name).toBe('Murph')
    })

    it('combines type and search filters', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedType = 'Girl'
      vm.searchQuery = 'custom'

      expect(vm.filteredWODs).toHaveLength(1)
      expect(vm.filteredWODs[0].name).toBe('My Custom WOD')
    })

    it('returns empty array when no matches', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.searchQuery = 'nonexistent'

      expect(vm.filteredWODs).toHaveLength(0)
    })
  })

  // ==========================================================================
  // SELECTION MODE TESTS
  // ==========================================================================

  describe('Selection Mode', () => {
    it('detects selection mode from route query', async () => {
      setupDefaultStore()
      createWrapper({ select: 'true' })
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.selectionMode).toBe(true)
    })

    it('is not in selection mode by default', async () => {
      setupDefaultStore()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.selectionMode).toBe(false)
    })

    it('navigates with selected WOD in selection mode', async () => {
      setupStoreWithData()
      createWrapper({ select: 'true', returnPath: '/workouts/templates/create' })
      await flushPromises()

      const vm = wrapper.vm
      vm.handleWODClick(mockWods[0])

      expect(mockPush).toHaveBeenCalledWith({
        path: '/workouts/templates/create',
        query: { selectedWOD: 1 }
      })
    })

    it('uses default return path in selection mode', async () => {
      setupStoreWithData()
      createWrapper({ select: 'true' })
      await flushPromises()

      const vm = wrapper.vm
      vm.handleWODClick(mockWods[0])

      expect(mockPush).toHaveBeenCalledWith({
        path: '/workouts/templates/create',
        query: { selectedWOD: 1 }
      })
    })

    it('shows detail dialog when not in selection mode', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.handleWODClick(mockWods[0])

      expect(vm.showDetailDialog).toBe(true)
      expect(vm.selectedWODId).toBe(1)
    })
  })

  // ==========================================================================
  // NAVIGATION TESTS
  // ==========================================================================

  describe('Navigation', () => {
    it('navigates to create WOD page', async () => {
      setupDefaultStore()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.createNewWOD()

      expect(mockPush).toHaveBeenCalledWith('/wods/create')
    })

    it('navigates to create WOD with return path in selection mode', async () => {
      setupDefaultStore()
      createWrapper({ select: 'true', returnPath: '/some/path' })
      await flushPromises()

      const vm = wrapper.vm
      vm.createNewWOD()

      expect(mockPush).toHaveBeenCalledWith({
        path: '/wods/create',
        query: { returnPath: '/some/path', select: 'true' }
      })
    })

    it('navigates to edit WOD page', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.editWOD(5)

      expect(mockPush).toHaveBeenCalledWith('/wods/5/edit')
    })
  })

  // ==========================================================================
  // HELPER FUNCTION TESTS
  // ==========================================================================

  describe('Helper Functions', () => {
    describe('getTodayDate', () => {
      it('returns date in YYYY-MM-DD format', async () => {
        setupDefaultStore()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.getTodayDate()

        expect(result).toMatch(/^\d{4}-\d{2}-\d{2}$/)
      })
    })

    describe('formatQuickLogName', () => {
      it('returns day-based workout name', async () => {
        setupDefaultStore()
        createWrapper()

        const vm = wrapper.vm
        // Monday Jan 1, 2024
        const result = vm.formatQuickLogName('2024-01-01')

        expect(result).toBe('Monday Workout')
      })

      it('handles empty date', async () => {
        setupDefaultStore()
        createWrapper()

        const vm = wrapper.vm
        const result = vm.formatQuickLogName('')

        expect(result).toBe('Workout')
      })

      it('handles null date', async () => {
        setupDefaultStore()
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
    it('opens quick log dialog with WOD data', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockWods[0])

      expect(vm.quickLogDialog).toBe(true)
      expect(vm.selectedWOD).toEqual(mockWods[0])
      expect(vm.quickLogData.date).toBeTruthy()
      expect(vm.quickLogData.name).toContain('Workout')
      expect(vm.quickLogData.wod.scoreType).toBe('Time (HH:MM:SS)')
    })

    it('does not open quick log without WOD', async () => {
      setupDefaultStore()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(null)

      expect(vm.quickLogDialog).toBe(false)
    })

    it('closes quick log dialog', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockWods[0])
      expect(vm.quickLogDialog).toBe(true)

      vm.closeQuickLog()

      expect(vm.quickLogDialog).toBe(false)
      expect(vm.selectedWOD).toBe(null)
    })

    it('submits quick log with time-based score', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockWods[0]) // Fran - Time based
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.timeHours = 0
      vm.quickLogData.wod.timeMinutes = 3
      vm.quickLogData.wod.timeSecondsInput = 45

      await vm.submitQuickLog()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', {
        workout_date: '2024-01-15',
        workout_name: 'Test Workout',
        total_time: null,
        notes: null,
        movements: [],
        wods: [
          {
            wod_id: 1,
            notes: null,
            time_seconds: 225 // 3 * 60 + 45
          }
        ]
      })
    })

    it('submits quick log with rounds+reps score', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockWods[2]) // Fight Gone Bad - Rounds+Reps
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.rounds = 10
      vm.quickLogData.wod.reps = 15

      await vm.submitQuickLog()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', {
        workout_date: '2024-01-15',
        workout_name: 'Test Workout',
        total_time: null,
        notes: null,
        movements: [],
        wods: [
          {
            wod_id: 3,
            notes: null,
            rounds: 10,
            reps: 15
          }
        ]
      })
    })

    it('submits quick log with max weight score', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockWods[3]) // Ringer 1 - Max Weight
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.weight = 315

      await vm.submitQuickLog()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', {
        workout_date: '2024-01-15',
        workout_name: 'Test Workout',
        total_time: null,
        notes: null,
        movements: [],
        wods: [
          {
            wod_id: 4,
            notes: null,
            weight: 315
          }
        ]
      })
    })

    it('navigates to dashboard after successful quick log', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockWods[0])
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.timeMinutes = 5

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

      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockWods[0])
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.timeMinutes = 5

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

      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockWods[0])
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.timeMinutes = 5

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

      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockWods[0])
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.timeMinutes = 5

      const submitPromise = vm.submitQuickLog()

      expect(vm.quickLogSubmitting).toBe(true)

      resolvePost({ data: {} })
      await submitPromise
      await flushPromises()

      expect(vm.quickLogSubmitting).toBe(false)
    })

    it('includes notes in quick log submission', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockWods[0])
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.timeMinutes = 5
      vm.quickLogData.wod.notes = 'Felt good!'

      await vm.submitQuickLog()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', expect.objectContaining({
        wods: [expect.objectContaining({
          notes: 'Felt good!'
        })]
      }))
    })

    it('calculates time in seconds correctly', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog(mockWods[0])
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.timeHours = 1
      vm.quickLogData.wod.timeMinutes = 30
      vm.quickLogData.wod.timeSecondsInput = 15

      await vm.submitQuickLog()
      await flushPromises()

      // 1*3600 + 30*60 + 15 = 3600 + 1800 + 15 = 5415
      expect(axios.post).toHaveBeenCalledWith('/api/workouts', expect.objectContaining({
        wods: [expect.objectContaining({
          time_seconds: 5415
        })]
      }))
    })
  })

  // ==========================================================================
  // DETAIL DIALOG TESTS
  // ==========================================================================

  describe('Detail Dialog', () => {
    it('handles quick log from detail dialog', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.handleDetailQuickLog({ data: mockWods[0] })

      expect(vm.quickLogDialog).toBe(true)
      expect(vm.selectedWOD).toEqual(mockWods[0])
    })

    it('does not open quick log if no data from detail dialog', async () => {
      setupDefaultStore()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.handleDetailQuickLog({ data: null })

      expect(vm.quickLogDialog).toBe(false)
    })
  })

  // ==========================================================================
  // SCORE TYPE SPECIFIC TESTS
  // ==========================================================================

  describe('Score Type Handling', () => {
    it('sets score type from WOD when opening quick log', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      // Test Time-based WOD
      vm.openQuickLog(mockWods[0])
      expect(vm.quickLogData.wod.scoreType).toBe('Time (HH:MM:SS)')
      vm.closeQuickLog()

      // Test Rounds+Reps WOD
      vm.openQuickLog(mockWods[2])
      expect(vm.quickLogData.wod.scoreType).toBe('Rounds+Reps')
      vm.closeQuickLog()

      // Test Max Weight WOD
      vm.openQuickLog(mockWods[3])
      expect(vm.quickLogData.wod.scoreType).toBe('Max Weight')
    })

    it('resets quick log data between WODs', async () => {
      setupStoreWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      // Open first WOD and set some data
      vm.openQuickLog(mockWods[0])
      vm.quickLogData.wod.timeMinutes = 10
      vm.closeQuickLog()

      // Open another WOD
      vm.openQuickLog(mockWods[2])

      // Time should be reset
      expect(vm.quickLogData.wod.timeMinutes).toBe(0)
      expect(vm.quickLogData.wod.timeHours).toBe(0)
      expect(vm.quickLogData.wod.timeSecondsInput).toBe(0)
    })
  })
})
