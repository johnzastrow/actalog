import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import WODDetailView from './WODDetailView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
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

// Mock auth store
const mockAuthStore = {
  user: { id: 1, role: 'athlete' }
}
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => mockAuthStore
}))

// Mock MarkdownRenderer
vi.mock('@/components/MarkdownRenderer.vue', () => ({
  default: {
    template: '<div class="markdown"></div>',
    props: ['content']
  }
}))

import axios from '@/utils/axios'

describe('WODDetailView', () => {
  let wrapper

  const mockWOD = {
    id: 1,
    name: 'Fran',
    type: 'girl',
    regime: 'for_time',
    score_type: 'Time (HH:MM:SS)',
    source: 'CrossFit',
    description: '21-15-9 Thrusters and Pull-ups',
    notes: 'Scale as needed',
    url: 'https://example.com/video',
    is_standard: true,
    created_by: 2,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-15T00:00:00Z'
  }

  const mockCustomWOD = {
    id: 2,
    name: 'My Custom WOD',
    type: 'benchmark',
    regime: 'amrap',
    score_type: 'Rounds+Reps',
    source: 'self-created',
    description: 'Custom workout',
    notes: null,
    url: null,
    is_standard: false,
    created_by: 1,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-15T00:00:00Z'
  }

  const setupDefaultMocks = () => {
    axios.get.mockResolvedValue({ data: { wod: mockWOD } })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(WODDetailView, {
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
          'v-form': { template: '<form><slot /></form>' },
          'v-text-field': { template: '<input />' },
          'v-textarea': { template: '<textarea></textarea>' },
          'v-spacer': { template: '<div></div>' },
          'v-tooltip': { template: '<div><slot /></div>' },
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
    mockAuthStore.user = { id: 1, role: 'athlete' }
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
    it('fetches WOD on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/wods/1')
    })

    it('stores fetched WOD', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.wod).toEqual(mockWOD)
    })

    it('sets loading to false after fetch', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)
    })

    it('handles WOD response without wrapper', async () => {
      axios.get.mockResolvedValue({ data: mockWOD })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.wod).toEqual(mockWOD)
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

      expect(vm.error).toContain('Failed to load WOD')
      expect(vm.loading).toBe(false)

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // EDIT PERMISSION TESTS
  // ==========================================================================

  describe('Edit Permissions', () => {
    it('allows admin to edit any WOD', async () => {
      mockAuthStore.user = { id: 99, role: 'admin' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.canEdit).toBe(true)
    })

    it('allows owner to edit custom WOD', async () => {
      mockAuthStore.user = { id: 1, role: 'athlete' }
      axios.get.mockResolvedValue({ data: { wod: mockCustomWOD } })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.canEdit).toBe(true)
    })

    it('does not allow non-owner to edit custom WOD', async () => {
      mockAuthStore.user = { id: 99, role: 'athlete' }
      axios.get.mockResolvedValue({ data: { wod: mockCustomWOD } })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.canEdit).toBe(false)
    })

    it('does not allow non-admin to edit standard WOD', async () => {
      mockAuthStore.user = { id: 1, role: 'athlete' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.canEdit).toBe(false)
    })

    it('returns false when no user', async () => {
      mockAuthStore.user = null
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.canEdit).toBe(false)
    })
  })

  // ==========================================================================
  // NAVIGATION TESTS
  // ==========================================================================

  describe('Navigation', () => {
    it('navigates to edit page', async () => {
      mockAuthStore.user = { id: 1, role: 'admin' }
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.editWOD()

      expect(mockPush).toHaveBeenCalledWith('/wods/1/edit')
    })
  })

  // ==========================================================================
  // FORMAT DATE TESTS
  // ==========================================================================

  describe('formatDate', () => {
    it('formats date correctly', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const result = vm.formatDate('2024-01-15T00:00:00Z')

      expect(result).toContain('January')
      expect(result).toContain('15')
      expect(result).toContain('2024')
    })

    it('returns N/A for null date', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.formatDate(null)).toBe('N/A')
    })
  })

  // ==========================================================================
  // QUICK LOG TESTS
  // ==========================================================================

  describe('Quick Log', () => {
    it('opens quick log dialog', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog()

      expect(vm.quickLogDialog).toBe(true)
      expect(vm.quickLogData.date).toBeTruthy()
      expect(vm.quickLogData.name).toContain('Workout')
    })

    it('does not open quick log without WOD', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue(new Error('Not found'))
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      // wod should be null after failed fetch
      expect(vm.wod).toBe(null)
      vm.openQuickLog()

      expect(vm.quickLogDialog).toBe(false)
      consoleSpy.mockRestore()
    })

    it('pre-populates score type from WOD', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog()

      expect(vm.quickLogData.wod.scoreType).toBe('Time (HH:MM:SS)')
    })

    it('closes quick log dialog', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog()
      expect(vm.quickLogDialog).toBe(true)

      vm.closeQuickLog()

      expect(vm.quickLogDialog).toBe(false)
    })

    it('submits quick log with time-based WOD', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog()
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.timeHours = 0
      vm.quickLogData.wod.timeMinutes = 5
      vm.quickLogData.wod.timeSecondsInput = 30

      await vm.submitQuickLog()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', expect.objectContaining({
        workout_date: '2024-01-15',
        workout_name: 'Test Workout',
        wods: [expect.objectContaining({
          wod_id: 1,
          time_seconds: 330
        })]
      }))
      expect(mockPush).toHaveBeenCalledWith('/dashboard')
    })

    it('submits quick log with rounds+reps WOD', async () => {
      axios.post.mockResolvedValue({ data: {} })
      axios.get.mockResolvedValue({ data: { wod: mockCustomWOD } })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog()
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.rounds = 10
      vm.quickLogData.wod.reps = 15

      await vm.submitQuickLog()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', expect.objectContaining({
        wods: [expect.objectContaining({
          wod_id: 2,
          rounds: 10,
          reps: 15
        })]
      }))
    })

    it('submits quick log with max weight WOD', async () => {
      const maxWeightWOD = { ...mockWOD, id: 3, score_type: 'Max Weight' }
      axios.post.mockResolvedValue({ data: {} })
      axios.get.mockResolvedValue({ data: { wod: maxWeightWOD } })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog()
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.weight = 225

      await vm.submitQuickLog()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', expect.objectContaining({
        wods: [expect.objectContaining({
          wod_id: 3,
          weight: 225
        })]
      }))
    })

    it('handles quick log submission error', async () => {
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue({
        response: { data: { message: 'Failed to log workout' } }
      })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog()
      vm.quickLogData.name = 'Test'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.timeMinutes = 5

      await vm.submitQuickLog()
      await flushPromises()

      expect(alertSpy).toHaveBeenCalledWith('Failed to log workout')

      alertSpy.mockRestore()
      consoleSpy.mockRestore()
    })

    it('sets submitting state during submission', async () => {
      let resolvePost
      axios.post.mockReturnValue(new Promise(resolve => { resolvePost = resolve }))
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog()
      vm.quickLogData.name = 'Test'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.wod.timeMinutes = 5

      const submitPromise = vm.submitQuickLog()

      expect(vm.quickLogSubmitting).toBe(true)

      resolvePost({ data: {} })
      await submitPromise
      await flushPromises()

      expect(vm.quickLogSubmitting).toBe(false)
    })
  })

  // ==========================================================================
  // HELPER FUNCTIONS TESTS
  // ==========================================================================

  describe('Helper Functions', () => {
    it('getTodayDate returns today in YYYY-MM-DD format', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const result = vm.getTodayDate()

      expect(result).toMatch(/^\d{4}-\d{2}-\d{2}$/)
    })

    it('formatQuickLogName returns day-based name', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const result = vm.formatQuickLogName('2024-01-15') // Monday

      expect(result).toContain('Workout')
    })

    it('formatQuickLogName returns Workout for null date', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.formatQuickLogName(null)).toBe('Workout')
    })
  })
})
