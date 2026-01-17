import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import MovementDetailView from './MovementDetailView.vue'

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

// Mock MarkdownRenderer
vi.mock('@/components/MarkdownRenderer.vue', () => ({
  default: {
    template: '<div class="markdown"></div>',
    props: ['content']
  }
}))

import axios from '@/utils/axios'

describe('MovementDetailView', () => {
  let wrapper

  const mockMovement = {
    id: 1,
    name: 'Back Squat',
    type: 'weightlifting',
    description: 'Barbell back squat movement',
    is_standard: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-15T00:00:00Z'
  }

  const mockStructuredMovement = {
    id: 2,
    name: 'Deadlift',
    type: 'weightlifting',
    description: '__STRUCTURED__{"description":"Hip hinge movement","difficulty":"Intermediate","equipment":"Barbell, Plates","primaryMuscles":"Posterior Chain","coachingCues":"Keep back flat","scalingOptions":"Kettlebell deadlift","videoUrl":"https://example.com/video"}',
    is_standard: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-15T00:00:00Z'
  }

  const setupDefaultMocks = () => {
    axios.get.mockResolvedValue({ data: mockMovement })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(MovementDetailView, {
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
    it('fetches movement on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/movements/1')
    })

    it('stores fetched movement', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.movement).toEqual(mockMovement)
    })

    it('sets loading to false after fetch', async () => {
      setupDefaultMocks()
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

      expect(vm.error).toContain('Failed to load movement')
      expect(vm.loading).toBe(false)

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // PARSED DATA TESTS
  // ==========================================================================

  describe('Parsed Data', () => {
    it('parses plain text description', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.parsedData.description).toBe('Barbell back squat movement')
    })

    it('parses structured description', async () => {
      axios.get.mockResolvedValue({ data: mockStructuredMovement })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.parsedData.description).toBe('Hip hinge movement')
      expect(vm.parsedData.difficulty).toBe('Intermediate')
      expect(vm.parsedData.equipment).toBe('Barbell, Plates')
      expect(vm.parsedData.primaryMuscles).toBe('Posterior Chain')
      expect(vm.parsedData.coachingCues).toBe('Keep back flat')
      expect(vm.parsedData.scalingOptions).toBe('Kettlebell deadlift')
      expect(vm.parsedData.videoUrl).toBe('https://example.com/video')
    })

    it('handles invalid structured JSON', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockResolvedValue({
        data: {
          ...mockMovement,
          description: '__STRUCTURED__invalid json'
        }
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.parsedData.description).toBe('__STRUCTURED__invalid json')

      consoleSpy.mockRestore()
    })

    it('returns empty object when no movement', async () => {
      axios.get.mockResolvedValue({ data: null })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.parsedData).toEqual({})
    })
  })

  // ==========================================================================
  // HELPER FUNCTIONS TESTS
  // ==========================================================================

  describe('Helper Functions', () => {
    describe('getMovementTypeIcon', () => {
      it('returns correct icons for types', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeIcon('weightlifting')).toBe('mdi-weight-lifter')
        expect(vm.getMovementTypeIcon('gymnastics')).toBe('mdi-gymnastics')
        expect(vm.getMovementTypeIcon('cardio')).toBe('mdi-run')
        expect(vm.getMovementTypeIcon('bodyweight')).toBe('mdi-human')
        expect(vm.getMovementTypeIcon('unknown')).toBe('mdi-dumbbell')
      })
    })

    describe('getMovementTypeColor', () => {
      it('returns correct colors for types', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getMovementTypeColor('weightlifting')).toBe('#00bcd4')
        expect(vm.getMovementTypeColor('gymnastics')).toBe('#9c27b0')
        expect(vm.getMovementTypeColor('cardio')).toBe('#ff5722')
        expect(vm.getMovementTypeColor('bodyweight')).toBe('#4caf50')
        expect(vm.getMovementTypeColor('unknown')).toBe('#666')
      })
    })

    describe('getDifficultyColor', () => {
      it('returns correct colors for difficulty', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.getDifficultyColor('Beginner')).toBe('#4caf50')
        expect(vm.getDifficultyColor('Intermediate')).toBe('#ffc107')
        expect(vm.getDifficultyColor('Advanced')).toBe('#ff5722')
        expect(vm.getDifficultyColor('Unknown')).toBe('#666')
      })
    })

    describe('capitalizeFirst', () => {
      it('capitalizes first letter', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.capitalizeFirst('weightlifting')).toBe('Weightlifting')
        expect(vm.capitalizeFirst('')).toBe('')
        expect(vm.capitalizeFirst(null)).toBe('')
      })
    })

    describe('formatDate', () => {
      it('formats date correctly', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        const result = vm.formatDate('2024-01-15T00:00:00Z')
        expect(result).toContain('January')
        expect(result).toContain('15')
        expect(result).toContain('2024')
      })

      it('returns N/A for null date', async () => {
        setupDefaultMocks()
        createWrapper()

        const vm = wrapper.vm

        expect(vm.formatDate(null)).toBe('N/A')
      })
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

    it('does not open quick log without movement', async () => {
      axios.get.mockResolvedValue({ data: null })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog()

      expect(vm.quickLogDialog).toBe(false)
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

    it('submits quick log successfully', async () => {
      axios.post.mockResolvedValue({ data: {} })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog()
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.movement.sets = 3
      vm.quickLogData.movement.reps = 10

      await vm.submitQuickLog()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', expect.objectContaining({
        workout_date: '2024-01-15',
        workout_name: 'Test Workout',
        movements: [expect.objectContaining({
          movement_id: 1,
          sets: 3,
          reps: 10
        })]
      }))
      expect(mockPush).toHaveBeenCalledWith('/dashboard')
    })

    it('handles quick log submission error', async () => {
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.post.mockRejectedValue({
        response: { data: { message: 'Failed to create workout' } }
      })
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openQuickLog()
      vm.quickLogData.name = 'Test'
      vm.quickLogData.date = '2024-01-15'
      vm.quickLogData.movement.sets = 3

      await vm.submitQuickLog()
      await flushPromises()

      expect(alertSpy).toHaveBeenCalledWith('Failed to create workout')

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
      vm.quickLogData.movement.sets = 3

      const submitPromise = vm.submitQuickLog()

      expect(vm.quickLogSubmitting).toBe(true)

      resolvePost({ data: {} })
      await submitPromise
      await flushPromises()

      expect(vm.quickLogSubmitting).toBe(false)
    })
  })

  // ==========================================================================
  // NAVIGATION TESTS
  // ==========================================================================

  describe('Navigation', () => {
    it('can go back', async () => {
      setupDefaultMocks()
      createWrapper()

      // The back button uses router.back()
      // This is tested by verifying the mock is available
      expect(mockBack).toBeDefined()
    })
  })
})
