import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import LogWorkoutView from './LogWorkoutView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
  }
}))

// Mock vue-router
const mockPush = vi.fn()
let mockRouteQuery = {}
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: mockPush,
    })),
    useRoute: vi.fn(() => ({
      query: mockRouteQuery
    }))
  }
})

// Mock RPE options
vi.mock('@/utils/rpe', () => ({
  RPE_OPTIONS: [
    { value: 6, title: 'RPE 6', description: 'Light', color: 'green' },
    { value: 8, title: 'RPE 8', description: 'Moderate', color: 'orange' },
    { value: 10, title: 'RPE 10', description: 'Max', color: 'red' }
  ]
}))

import axios from '@/utils/axios'

describe('LogWorkoutView', () => {
  let wrapper

  const mockTemplates = [
    {
      id: 1,
      name: 'Standard Workout',
      notes: 'Basic strength workout',
      created_by: null,
      movements: [
        { movement_id: 1, movement: { name: 'Back Squat' } }
      ],
      wods: []
    },
    {
      id: 2,
      name: 'Custom Workout',
      notes: 'My personal workout',
      created_by: 1,
      movements: [],
      wods: [
        { wod_id: 1, wod: { name: 'Fran', score_type: 'Time (HH:MM:SS)' }, wod_score_type: 'Time (HH:MM:SS)' }
      ]
    }
  ]

  const mockTemplateDetail = {
    id: 1,
    name: 'Standard Workout',
    notes: 'Basic strength workout',
    movements: [
      { movement_id: 1, movement: { name: 'Back Squat' } },
      { movement_id: 2, movement: { name: 'Deadlift' } }
    ],
    wods: []
  }

  const mockWorkoutForEdit = {
    id: 100,
    workout_name: 'My Workout',
    workout_id: 1,
    workout_date: '2024-01-15',
    total_time: 3600,
    notes: 'Great workout',
    performance_movements: [
      {
        movement_id: 1,
        movement: { name: 'Back Squat' },
        sets: 5,
        reps: 5,
        weight: 225,
        time_seconds: null,
        distance: null,
        notes: 'Felt strong',
        rpe: 8
      }
    ],
    performance_wods: []
  }

  const setupDefaultMocks = () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/workouts/standard') {
        return Promise.resolve({ data: { workouts: [mockTemplates[0]] } })
      }
      if (url === '/api/workouts/my-templates') {
        return Promise.resolve({ data: { workouts: [mockTemplates[1]] } })
      }
      if (url.startsWith('/api/templates/')) {
        return Promise.resolve({ data: { template: mockTemplateDetail } })
      }
      if (url.startsWith('/api/workouts/')) {
        return Promise.resolve({ data: mockWorkoutForEdit })
      }
      return Promise.resolve({ data: {} })
    })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(LogWorkoutView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-form': { template: '<form @submit.prevent><slot /></form>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-text-field': { template: '<input />' },
          'v-textarea': { template: '<textarea></textarea>' },
          'v-autocomplete': { template: '<select></select>' },
          'v-select': { template: '<select></select>' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-icon': { template: '<i></i>' },
          'v-chip': { template: '<span><slot /></span>' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-list-item': { template: '<div><slot /></div>' },
          'v-list-item-title': { template: '<div><slot /></div>' },
          'v-list-item-subtitle': { template: '<div><slot /></div>' },
          'v-progress-circular': { template: '<div></div>' },
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
    it('fetches templates on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/workouts/standard')
      expect(axios.get).toHaveBeenCalledWith('/api/workouts/my-templates')
    })

    it('combines standard and custom templates', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.workoutTemplates).toHaveLength(2)
    })

    it('initializes in create mode by default', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.isEditMode).toBe(false)
      expect(vm.editWorkoutId).toBe(null)
    })

    it('sets workout date to today', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const today = new Date().toISOString().split('T')[0]

      expect(vm.workoutDate).toBe(today)
    })
  })

  // ==========================================================================
  // EDIT MODE TESTS
  // ==========================================================================

  describe('Edit Mode', () => {
    it('enters edit mode when edit query param present', async () => {
      mockRouteQuery = { edit: '100' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.isEditMode).toBe(true)
      expect(axios.get).toHaveBeenCalledWith('/api/workouts/100')
    })

    it('loads workout data in edit mode', async () => {
      mockRouteQuery = { edit: '100' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.workoutName).toBe('My Workout')
      expect(vm.workoutDate).toBe('2024-01-15')
      expect(vm.totalTimeMinutes).toBe(60) // 3600 seconds = 60 minutes
      expect(vm.notes).toBe('Great workout')
    })

    it('loads movement performance in edit mode', async () => {
      mockRouteQuery = { edit: '100' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.movementPerformance).toHaveLength(1)
      expect(vm.movementPerformance[0].sets).toBe(5)
      expect(vm.movementPerformance[0].reps).toBe(5)
      expect(vm.movementPerformance[0].weight).toBe(225)
    })

    it('sets editDataLoaded flag after loading', async () => {
      mockRouteQuery = { edit: '100' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.editDataLoaded).toBe(true)
    })
  })

  // ==========================================================================
  // TEMPLATE SELECTION TESTS
  // ==========================================================================

  describe('Template Selection', () => {
    it('selects template from query param', async () => {
      mockRouteQuery = { template: '1' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.selectedTemplateId).toBe(1)
    })

    it('fetches template details on selection', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.onTemplateSelected(1)
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/templates/1')
    })

    it('initializes performance arrays when template selected', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedTemplateId = 1
      await vm.onTemplateSelected(1)
      await flushPromises()

      // Template has 2 movements
      expect(vm.movementPerformance.length).toBeGreaterThan(0)
    })
  })

  // ==========================================================================
  // COMPUTED PROPERTIES TESTS
  // ==========================================================================

  describe('Computed Properties', () => {
    describe('selectedTemplate', () => {
      it('returns null when no template selected', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.selectedTemplate).toBe(null)
      })

      it('returns template when selected', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.selectedTemplateId = 1

        expect(vm.selectedTemplate).not.toBe(null)
        expect(vm.selectedTemplate.name).toBe('Standard Workout')
      })
    })

    describe('shouldShowMovements', () => {
      it('returns falsy when no template with movements', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.shouldShowMovements).toBeFalsy()
      })

      it('returns true when template has movements', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.selectedTemplateId = 1

        expect(vm.shouldShowMovements).toBe(true)
      })

      it('returns true in edit mode with movements', async () => {
        mockRouteQuery = { edit: '100' }
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.shouldShowMovements).toBe(true)
      })
    })

    describe('shouldShowWODs', () => {
      it('returns falsy when no template with WODs', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.shouldShowWODs).toBeFalsy()
      })

      it('returns true when template has WODs', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.selectedTemplateId = 2

        expect(vm.shouldShowWODs).toBe(true)
      })
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

    describe('truncateText', () => {
      it('returns short text unchanged', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.truncateText('Short', 10)).toBe('Short')
      })

      it('truncates long text', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        const result = vm.truncateText('This is a very long text', 10)

        expect(result).toBe('This is a ...')
      })

      it('handles null text', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm

        expect(vm.truncateText(null, 10)).toBe(null)
      })
    })

    describe('getMovementName', () => {
      it('returns movement name from template', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.selectedTemplateId = 1
        await vm.onTemplateSelected(1)
        await flushPromises()

        const result = vm.getMovementName(1)

        expect(result).toBe('Back Squat')
      })

      it('returns fallback in edit mode', async () => {
        mockRouteQuery = { edit: '100' }
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        const result = vm.getMovementName(1)

        expect(result).toBe('Back Squat')
      })
    })

    describe('getWODName', () => {
      it('returns WOD name from template', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.selectedTemplateId = 2

        const result = vm.getWODName(1)

        expect(result).toBe('Fran')
      })
    })
  })

  // ==========================================================================
  // BUILD PAYLOAD TESTS
  // ==========================================================================

  describe('Build Payload Functions', () => {
    describe('buildMovementsPayload', () => {
      it('filters out empty movement entries', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.movementPerformance = [
          { movement_id: 1, sets: null, reps: null, weight: null, time: null, distance: null, notes: '', rpe: null, order_index: 0 },
          { movement_id: 2, sets: 3, reps: 10, weight: 135, time: null, distance: null, notes: 'Good', rpe: 7, order_index: 1 }
        ]

        const result = vm.buildMovementsPayload()

        expect(result).toHaveLength(1)
        expect(result[0].movement_id).toBe(2)
      })

      it('includes movement with any data', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.movementPerformance = [
          { movement_id: 1, sets: null, reps: null, weight: 100, time: null, distance: null, notes: '', rpe: null, order_index: 0 }
        ]

        const result = vm.buildMovementsPayload()

        expect(result).toHaveLength(1)
      })
    })

    describe('buildWODsPayload', () => {
      it('calculates time_seconds from hours, minutes, seconds', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.wodPerformance = [
          {
            wod_id: 1,
            score_type: 'Time (HH:MM:SS)',
            time_hours: 0,
            time_minutes: 5,
            time_seconds: 30,
            rounds: null,
            reps: null,
            weight: null,
            notes: '',
            rpe: null,
            order_index: 0
          }
        ]

        const result = vm.buildWODsPayload()

        expect(result).toHaveLength(1)
        expect(result[0].time_seconds).toBe(330) // 5*60 + 30
        expect(result[0].score_value).toBe('00:05:30')
      })

      it('builds rounds+reps score_value', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.wodPerformance = [
          {
            wod_id: 1,
            score_type: 'Rounds+Reps',
            time_hours: null,
            time_minutes: null,
            time_seconds: null,
            rounds: 10,
            reps: 15,
            weight: null,
            notes: '',
            rpe: null,
            order_index: 0
          }
        ]

        const result = vm.buildWODsPayload()

        expect(result).toHaveLength(1)
        expect(result[0].rounds).toBe(10)
        expect(result[0].reps).toBe(15)
        expect(result[0].score_value).toBe('10+15')
      })

      it('builds max weight score_value', async () => {
        setupDefaultMocks()
        createWrapper()
        await flushPromises()

        const vm = wrapper.vm
        vm.wodPerformance = [
          {
            wod_id: 1,
            score_type: 'Max Weight',
            time_hours: null,
            time_minutes: null,
            time_seconds: null,
            rounds: null,
            reps: null,
            weight: 225,
            notes: '',
            rpe: null,
            order_index: 0
          }
        ]

        const result = vm.buildWODsPayload()

        expect(result).toHaveLength(1)
        expect(result[0].weight).toBe(225)
        expect(result[0].score_value).toBe('225')
      })
    })
  })

  // ==========================================================================
  // SUBMIT TESTS
  // ==========================================================================

  describe('Log Workout', () => {
    it('validates template and date in create mode', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedTemplateId = null
      vm.workoutDate = '2024-01-15'

      await vm.logWorkout()

      expect(vm.error).toContain('select a template')
      expect(axios.post).not.toHaveBeenCalled()
    })

    it('validates date in edit mode', async () => {
      mockRouteQuery = { edit: '100' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.workoutDate = ''

      await vm.logWorkout()

      expect(vm.error).toContain('select a date')
      expect(axios.put).not.toHaveBeenCalled()
    })

    it('creates new workout in create mode', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedTemplateId = 1
      vm.workoutDate = '2024-01-15'
      vm.totalTimeMinutes = 45
      vm.notes = 'Test notes'

      await vm.logWorkout()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', expect.objectContaining({
        workout_id: 1,
        workout_date: '2024-01-15',
        total_time: 2700, // 45 * 60
        notes: 'Test notes'
      }))
      expect(vm.success).toContain('logged successfully')
    })

    it('updates workout in edit mode', async () => {
      mockRouteQuery = { edit: '100' }
      axios.put.mockResolvedValue({ data: {} })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.notes = 'Updated notes'

      await vm.logWorkout()
      await flushPromises()

      expect(axios.put).toHaveBeenCalledWith('/api/workouts/100', expect.objectContaining({
        notes: 'Updated notes'
      }))
      expect(vm.success).toContain('updated successfully')
    })

    it('includes movement performance in payload', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedTemplateId = 1
      vm.workoutDate = '2024-01-15'
      vm.movementPerformance = [
        { movement_id: 1, sets: 5, reps: 5, weight: 225, time: null, distance: null, notes: '', rpe: null, order_index: 0 }
      ]

      await vm.logWorkout()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', expect.objectContaining({
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

    it('handles submit error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue({
        response: { data: { message: 'Failed to save' } }
      })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedTemplateId = 1
      vm.workoutDate = '2024-01-15'

      await vm.logWorkout()
      await flushPromises()

      expect(vm.error).toBe('Failed to save')

      consoleSpy.mockRestore()
    })

    it('sets submitting state during save', async () => {
      let resolvePost
      axios.post.mockReturnValue(new Promise(resolve => { resolvePost = resolve }))
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedTemplateId = 1
      vm.workoutDate = '2024-01-15'

      const submitPromise = vm.logWorkout()

      expect(vm.submitting).toBe(true)

      resolvePost({ data: {} })
      await submitPromise
      await flushPromises()

      expect(vm.submitting).toBe(false)
    })

    it('redirects to dashboard after create', async () => {
      vi.useFakeTimers()
      axios.post.mockResolvedValue({ data: {} })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedTemplateId = 1
      vm.workoutDate = '2024-01-15'

      await vm.logWorkout()
      await flushPromises()

      vi.advanceTimersByTime(1500)
      expect(mockPush).toHaveBeenCalledWith('/dashboard')

      vi.useRealTimers()
    })

    it('redirects to workout detail after edit', async () => {
      vi.useFakeTimers()
      mockRouteQuery = { edit: '100' }
      axios.put.mockResolvedValue({ data: {} })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      await vm.logWorkout()
      await flushPromises()

      vi.advanceTimersByTime(1500)
      expect(mockPush).toHaveBeenCalledWith('/workouts/100')

      vi.useRealTimers()
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('handles template fetch error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockImplementation((url) => {
        if (url === '/api/workouts/standard') {
          return Promise.reject(new Error('Network error'))
        }
        return Promise.resolve({ data: { workouts: [] } })
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBeTruthy()

      consoleSpy.mockRestore()
    })

    it('handles edit load error', async () => {
      mockRouteQuery = { edit: '999' }
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockImplementation((url) => {
        if (url === '/api/workouts/999') {
          return Promise.reject({ response: { data: { message: 'Workout not found' } } })
        }
        if (url === '/api/workouts/standard' || url === '/api/workouts/my-templates') {
          return Promise.resolve({ data: { workouts: [] } })
        }
        return Promise.resolve({ data: {} })
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Workout not found')

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // PERFORMANCE ARRAYS INITIALIZATION TESTS
  // ==========================================================================

  describe('Performance Arrays Initialization', () => {
    it('initializes movement performance from template', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedTemplateId = 1
      await vm.onTemplateSelected(1)
      await flushPromises()

      expect(vm.movementPerformance.length).toBeGreaterThan(0)
      expect(vm.movementPerformance[0].movement_id).toBe(1)
      expect(vm.movementPerformance[0].sets).toBe(null)
      expect(vm.movementPerformance[0].reps).toBe(null)
      expect(vm.movementPerformance[0].weight).toBe(null)
    })

    it('initializes WOD performance with score type from template', async () => {
      axios.get.mockImplementation((url) => {
        if (url === '/api/workouts/standard') {
          return Promise.resolve({ data: { workouts: [] } })
        }
        if (url === '/api/workouts/my-templates') {
          return Promise.resolve({ data: { workouts: [mockTemplates[1]] } })
        }
        if (url === '/api/templates/2') {
          return Promise.resolve({
            data: {
              template: {
                id: 2,
                name: 'WOD Template',
                movements: [],
                wods: [{ wod_id: 1, wod: { name: 'Fran', score_type: 'Time (HH:MM:SS)' }, wod_score_type: 'Time (HH:MM:SS)' }]
              }
            }
          })
        }
        return Promise.resolve({ data: {} })
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectedTemplateId = 2
      await vm.onTemplateSelected(2)
      await flushPromises()

      expect(vm.wodPerformance.length).toBeGreaterThan(0)
      expect(vm.wodPerformance[0].score_type).toBe('Time (HH:MM:SS)')
    })
  })
})
