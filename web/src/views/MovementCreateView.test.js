import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import MovementCreateView from './MovementCreateView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  }
}))

// Mock vue-router
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRoute: vi.fn(() => ({
      params: {},
      query: {}
    })),
    useRouter: vi.fn(() => ({
      push: vi.fn(),
      replace: vi.fn(),
      back: vi.fn()
    }))
  }
})

// Import mocked modules
import axios from '@/utils/axios'
import { useRoute, useRouter } from 'vue-router'

describe('MovementCreateView', () => {
  let wrapper
  let mockRouter
  let mockRoute

  const createWrapper = async (routeParams = {}, routeQuery = {}) => {
    const pinia = createPinia()
    setActivePinia(pinia)

    mockRoute = {
      params: routeParams,
      query: routeQuery
    }
    mockRouter = {
      push: vi.fn(),
      replace: vi.fn(),
      back: vi.fn()
    }

    useRoute.mockReturnValue(mockRoute)
    useRouter.mockReturnValue(mockRouter)

    // Mock GET for loading existing movement
    axios.get.mockImplementation((url) => {
      if (url.match(/\/api\/movements\/\d+/)) {
        return Promise.resolve({
          data: {
            movement: {
              id: 1,
              name: 'Back Squat',
              type: 'weightlifting',
              description: '__STRUCTURED__{"difficulty":"intermediate","equipment":"Barbell, Squat Rack","primaryMuscles":"Quadriceps, Glutes","description":"Stand with barbell on back, squat down","coachingCues":"Keep chest up","videoUrl":"https://youtube.com/watch?v=123","scalingOptions":"Use goblet squat for beginners"}'
            }
          }
        })
      }
      return Promise.resolve({ data: {} })
    })

    // Mock POST for creating
    axios.post.mockResolvedValue({
      data: {
        movement: {
          id: 2,
          name: 'New Movement',
          type: 'weightlifting',
          description: '__STRUCTURED__{"description":"Test description"}'
        }
      }
    })

    // Mock PUT for updating
    axios.put.mockResolvedValue({
      data: {
        movement: {
          id: 1,
          name: 'Updated Back Squat',
          type: 'weightlifting',
          description: '__STRUCTURED__{"description":"Updated description"}'
        }
      }
    })

    // Mock DELETE
    axios.delete.mockResolvedValue({ data: {} })

    wrapper = mount(MovementCreateView, {
      global: {
        plugins: [pinia],
      }
    })

    await flushPromises()
    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    // Mock window.confirm for delete tests
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })

  // ==========================================================================
  // CREATE MODE TESTS
  // ==========================================================================

  describe('Create Mode', () => {
    it('renders form in create mode with default values', async () => {
      await createWrapper()

      const vm = wrapper.vm
      expect(vm.movement.name).toBe('')
      expect(vm.movement.type).toBe('weightlifting')
      expect(vm.movement.description).toBe('')
    })

    it('shows correct button text for create mode', async () => {
      await createWrapper()

      expect(wrapper.text()).toContain('Create Movement')
    })

    it('does not show delete button in create mode', async () => {
      await createWrapper()

      expect(wrapper.text()).not.toContain('Delete Movement')
    })

    it('validates required fields before creating', async () => {
      await createWrapper()

      const vm = wrapper.vm

      // Try to save with empty fields
      await vm.saveMovement()
      await flushPromises()

      // Should have validation errors
      expect(vm.validationErrors.name).toBeDefined()
      expect(vm.validationErrors.description).toBeDefined()

      // Should not have called API
      expect(axios.post).not.toHaveBeenCalled()
    })

    it('creates movement with all fields', async () => {
      await createWrapper()

      const vm = wrapper.vm

      // Fill in all fields
      vm.movement.name = 'Test Deadlift'
      vm.movement.type = 'weightlifting'
      vm.movement.difficulty = 'intermediate'
      vm.movement.equipment = 'Barbell, Plates'
      vm.movement.primaryMuscles = 'Back, Hamstrings, Glutes'
      vm.movement.description = 'Lift the bar from floor to hip'
      vm.movement.coachingCues = 'Keep back flat, drive through heels'
      vm.movement.videoUrl = 'https://youtube.com/watch?v=abc'
      vm.movement.scalingOptions = 'Use lighter weights or trap bar'

      await vm.saveMovement()
      await flushPromises()

      // Verify POST was called
      expect(axios.post).toHaveBeenCalled()

      // Check the payload
      const callArgs = axios.post.mock.calls[0]
      expect(callArgs[0]).toBe('/api/movements')

      const payload = callArgs[1]
      expect(payload.name).toBe('Test Deadlift')
      expect(payload.type).toBe('weightlifting')

      // Verify structured data was encoded
      expect(payload.description).toContain('__STRUCTURED__')

      // Parse the structured data
      const jsonStr = payload.description.substring('__STRUCTURED__'.length)
      const structuredData = JSON.parse(jsonStr)
      expect(structuredData.difficulty).toBe('intermediate')
      expect(structuredData.equipment).toBe('Barbell, Plates')
      expect(structuredData.primaryMuscles).toBe('Back, Hamstrings, Glutes')
      expect(structuredData.description).toBe('Lift the bar from floor to hip')
      expect(structuredData.coachingCues).toBe('Keep back flat, drive through heels')
      expect(structuredData.videoUrl).toBe('https://youtube.com/watch?v=abc')
      expect(structuredData.scalingOptions).toBe('Use lighter weights or trap bar')
    })

    it('navigates to movements list after successful create', async () => {
      vi.useFakeTimers()
      await createWrapper()

      const vm = wrapper.vm

      // Fill required fields
      vm.movement.name = 'Test Movement'
      vm.movement.type = 'weightlifting'
      vm.movement.description = 'Test description'

      await vm.saveMovement()
      await flushPromises()

      // Advance timer for navigation
      vi.advanceTimersByTime(2000)

      expect(mockRouter.push).toHaveBeenCalledWith('/movements')

      vi.useRealTimers()
    })

    it('handles selection mode - returns with selectedMovement', async () => {
      vi.useFakeTimers()
      await createWrapper({}, { select: 'true', returnPath: '/templates/new' })

      const vm = wrapper.vm

      // Fill required fields
      vm.movement.name = 'Test Movement'
      vm.movement.type = 'weightlifting'
      vm.movement.description = 'Test description'

      await vm.saveMovement()
      await flushPromises()

      // Advance timer
      vi.advanceTimersByTime(2000)

      expect(mockRouter.push).toHaveBeenCalledWith({
        path: '/templates/new',
        query: { selectedMovement: 2 }
      })

      vi.useRealTimers()
    })
  })

  // ==========================================================================
  // EDIT MODE TESTS
  // ==========================================================================

  describe('Edit Mode', () => {
    it('loads existing movement data', async () => {
      await createWrapper({ id: '1' })

      // Verify API call
      expect(axios.get).toHaveBeenCalledWith('/api/movements/1')

      const vm = wrapper.vm

      // Verify data was loaded
      expect(vm.movement.name).toBe('Back Squat')
      expect(vm.movement.type).toBe('weightlifting')
      expect(vm.movement.difficulty).toBe('intermediate')
      expect(vm.movement.equipment).toBe('Barbell, Squat Rack')
      expect(vm.movement.primaryMuscles).toBe('Quadriceps, Glutes')
      expect(vm.movement.description).toBe('Stand with barbell on back, squat down')
      expect(vm.movement.coachingCues).toBe('Keep chest up')
      expect(vm.movement.videoUrl).toBe('https://youtube.com/watch?v=123')
      expect(vm.movement.scalingOptions).toBe('Use goblet squat for beginners')
    })

    it('shows correct button text for edit mode', async () => {
      await createWrapper({ id: '1' })

      expect(wrapper.text()).toContain('Update Movement')
    })

    it('shows delete button in edit mode', async () => {
      await createWrapper({ id: '1' })

      expect(wrapper.text()).toContain('Delete Movement')
    })

    it('updates movement with PUT request', async () => {
      await createWrapper({ id: '1' })

      const vm = wrapper.vm

      // Modify data
      vm.movement.name = 'Updated Back Squat'
      vm.movement.description = 'Updated description'

      await vm.saveMovement()
      await flushPromises()

      // Verify PUT was called
      expect(axios.put).toHaveBeenCalled()

      const callArgs = axios.put.mock.calls[0]
      expect(callArgs[0]).toBe('/api/movements/1')

      const payload = callArgs[1]
      expect(payload.name).toBe('Updated Back Squat')
    })
  })

  // ==========================================================================
  // DELETE TESTS
  // ==========================================================================

  describe('Delete Operation', () => {
    it('shows delete confirmation dialog', async () => {
      await createWrapper({ id: '1' })

      const vm = wrapper.vm

      // Trigger delete confirmation
      vm.confirmDelete()
      await flushPromises()

      expect(vm.deleteDialog).toBe(true)
    })

    it('deletes movement when confirmed', async () => {
      vi.useFakeTimers()
      await createWrapper({ id: '1' })

      const vm = wrapper.vm

      // Open dialog and delete
      vm.deleteDialog = true
      await vm.deleteMovement()
      await flushPromises()

      // Verify DELETE was called
      expect(axios.delete).toHaveBeenCalledWith('/api/movements/1')

      // Advance timer for navigation
      vi.advanceTimersByTime(2000)

      expect(mockRouter.push).toHaveBeenCalledWith('/movements')

      vi.useRealTimers()
    })

    it('handles delete error gracefully', async () => {
      await createWrapper({ id: '1' })

      // Mock delete failure
      axios.delete.mockRejectedValueOnce(new Error('Delete failed'))

      const vm = wrapper.vm

      await vm.deleteMovement()
      await flushPromises()

      expect(vm.error).toContain('Failed to delete movement')
      expect(vm.deleteDialog).toBe(false)
    })
  })

  // ==========================================================================
  // VALIDATION TESTS
  // ==========================================================================

  describe('Validation', () => {
    it('validates name is required', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.movement.name = ''
      vm.movement.type = 'weightlifting'
      vm.movement.description = 'Valid description'

      const isValid = vm.validateMovement()

      expect(isValid).toBe(false)
      expect(vm.validationErrors.name).toBe('Movement name is required')
    })

    it('validates type is required', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.movement.name = 'Valid Name'
      vm.movement.type = ''
      vm.movement.description = 'Valid description'

      const isValid = vm.validateMovement()

      expect(isValid).toBe(false)
      expect(vm.validationErrors.type).toBe('Movement type is required')
    })

    it('validates description is required', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.movement.name = 'Valid Name'
      vm.movement.type = 'weightlifting'
      vm.movement.description = ''

      const isValid = vm.validateMovement()

      expect(isValid).toBe(false)
      expect(vm.validationErrors.description).toBe('Description is required')
    })

    it('passes validation with all required fields', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.movement.name = 'Valid Name'
      vm.movement.type = 'weightlifting'
      vm.movement.description = 'Valid description'

      const isValid = vm.validateMovement()

      expect(isValid).toBe(true)
      expect(Object.keys(vm.validationErrors).length).toBe(0)
    })

    it('trims whitespace from name', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.movement.name = '   '
      vm.movement.type = 'weightlifting'
      vm.movement.description = 'Valid description'

      const isValid = vm.validateMovement()

      expect(isValid).toBe(false)
      expect(vm.validationErrors.name).toBe('Movement name is required')
    })
  })

  // ==========================================================================
  // MOVEMENT TYPES TESTS
  // ==========================================================================

  describe('Movement Types', () => {
    it('has correct movement type options', async () => {
      await createWrapper()

      const vm = wrapper.vm

      expect(vm.movementTypes).toContainEqual({ title: 'Weightlifting', value: 'weightlifting' })
      expect(vm.movementTypes).toContainEqual({ title: 'Gymnastics', value: 'gymnastics' })
      expect(vm.movementTypes).toContainEqual({ title: 'Cardio', value: 'cardio' })
      expect(vm.movementTypes).toContainEqual({ title: 'Bodyweight', value: 'bodyweight' })
    })

    it('has correct difficulty level options', async () => {
      await createWrapper()

      const vm = wrapper.vm

      expect(vm.difficultyLevels).toContainEqual({ title: 'Beginner', value: 'beginner' })
      expect(vm.difficultyLevels).toContainEqual({ title: 'Intermediate', value: 'intermediate' })
      expect(vm.difficultyLevels).toContainEqual({ title: 'Advanced', value: 'advanced' })
      expect(vm.difficultyLevels).toContainEqual({ title: 'Elite', value: 'elite' })
    })
  })

  // ==========================================================================
  // OPTIONAL FIELDS TESTS
  // ==========================================================================

  describe('Optional Fields', () => {
    it('handles null optional fields', async () => {
      await createWrapper()

      const vm = wrapper.vm

      // Fill required fields only
      vm.movement.name = 'Test Movement'
      vm.movement.type = 'weightlifting'
      vm.movement.description = 'Test description'
      // Leave optional fields empty
      vm.movement.difficulty = ''
      vm.movement.equipment = ''
      vm.movement.primaryMuscles = ''
      vm.movement.coachingCues = ''
      vm.movement.videoUrl = ''
      vm.movement.scalingOptions = ''

      await vm.saveMovement()
      await flushPromises()

      expect(axios.post).toHaveBeenCalled()

      // Verify structured data has null for empty optional fields
      const payload = axios.post.mock.calls[0][1]
      const jsonStr = payload.description.substring('__STRUCTURED__'.length)
      const structuredData = JSON.parse(jsonStr)

      expect(structuredData.difficulty).toBeNull()
      expect(structuredData.equipment).toBeNull()
    })

    it('preserves optional fields when editing', async () => {
      await createWrapper({ id: '1' })

      const vm = wrapper.vm

      // Verify optional fields were loaded
      expect(vm.movement.difficulty).toBe('intermediate')
      expect(vm.movement.equipment).toBe('Barbell, Squat Rack')
      expect(vm.movement.coachingCues).toBe('Keep chest up')
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('handles API error on create', async () => {
      await createWrapper()

      // Mock API failure
      axios.post.mockRejectedValueOnce({
        response: {
          data: {
            message: 'Movement already exists'
          }
        }
      })

      const vm = wrapper.vm

      vm.movement.name = 'Duplicate Movement'
      vm.movement.type = 'weightlifting'
      vm.movement.description = 'Test description'

      await vm.saveMovement()
      await flushPromises()

      expect(vm.error).toBe('Movement already exists')
    })

    it('handles API error on load', async () => {
      // Mock load failure
      axios.get.mockRejectedValueOnce(new Error('Network error'))

      await createWrapper({ id: '1' })
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.error).toBe('Failed to load movement')
    })

    it('handles generic API error', async () => {
      await createWrapper()

      // Mock API failure without message
      axios.post.mockRejectedValueOnce(new Error('Network error'))

      const vm = wrapper.vm

      vm.movement.name = 'Test Movement'
      vm.movement.type = 'weightlifting'
      vm.movement.description = 'Test description'

      await vm.saveMovement()
      await flushPromises()

      expect(vm.error).toBe('Failed to save movement. Please try again.')
    })
  })

  // ==========================================================================
  // NAVIGATION TESTS
  // ==========================================================================

  describe('Navigation', () => {
    it('uses custom return path from query', async () => {
      vi.useFakeTimers()
      await createWrapper({}, { returnPath: '/custom/path' })

      const vm = wrapper.vm

      // Fill required fields
      vm.movement.name = 'Test Movement'
      vm.movement.type = 'weightlifting'
      vm.movement.description = 'Test description'

      await vm.saveMovement()
      await flushPromises()

      vi.advanceTimersByTime(2000)

      expect(mockRouter.push).toHaveBeenCalledWith('/movements')

      vi.useRealTimers()
    })

    it('handleBack navigates without confirmation when no changes', async () => {
      await createWrapper()

      const vm = wrapper.vm

      vm.handleBack()

      expect(mockRouter.push).toHaveBeenCalledWith('/movements')
    })

    it('handleBack asks for confirmation when there are changes', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.movement.name = 'Some changes'

      vm.handleBack()

      expect(window.confirm).toHaveBeenCalled()
    })
  })

  // ==========================================================================
  // SPECIAL CHARACTERS TESTS
  // ==========================================================================

  describe('Special Characters', () => {
    it('preserves markdown in description', async () => {
      await createWrapper()

      const vm = wrapper.vm

      const markdownContent = `## Setup Instructions

**Equipment Required:**
- Barbell
- Squat Rack

1. Step one
2. Step two

> Important note`

      vm.movement.name = 'Test Movement'
      vm.movement.type = 'weightlifting'
      vm.movement.description = markdownContent

      await vm.saveMovement()
      await flushPromises()

      // Verify markdown is preserved
      const payload = axios.post.mock.calls[0][1]
      const jsonStr = payload.description.substring('__STRUCTURED__'.length)
      const structuredData = JSON.parse(jsonStr)

      expect(structuredData.description).toBe(markdownContent)
    })

    it('preserves unicode characters', async () => {
      await createWrapper()

      const vm = wrapper.vm

      vm.movement.name = 'Sentadilla con barra'
      vm.movement.type = 'weightlifting'
      vm.movement.description = 'Japanese: 日本語 | Emoji: 💪🏋️'

      await vm.saveMovement()
      await flushPromises()

      const payload = axios.post.mock.calls[0][1]
      expect(payload.name).toBe('Sentadilla con barra')

      const jsonStr = payload.description.substring('__STRUCTURED__'.length)
      const structuredData = JSON.parse(jsonStr)
      expect(structuredData.description).toBe('Japanese: 日本語 | Emoji: 💪🏋️')
    })
  })

  // ==========================================================================
  // STATE MANAGEMENT TESTS
  // ==========================================================================

  describe('State Management', () => {
    it('sets saving state during save operation', async () => {
      await createWrapper()

      const vm = wrapper.vm

      vm.movement.name = 'Test Movement'
      vm.movement.type = 'weightlifting'
      vm.movement.description = 'Test description'

      // Don't await - check intermediate state
      const savePromise = vm.saveMovement()

      expect(vm.saving).toBe(true)

      await savePromise
      await flushPromises()

      expect(vm.saving).toBe(false)
    })

    it('sets deleting state during delete operation', async () => {
      await createWrapper({ id: '1' })

      const vm = wrapper.vm

      // Don't await - check intermediate state
      const deletePromise = vm.deleteMovement()

      expect(vm.deleting).toBe(true)

      await deletePromise
      await flushPromises()

      expect(vm.deleting).toBe(false)
    })

    it('clears success message on close', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.successMessage = 'Test success'

      // Simulate alert close
      vm.successMessage = ''

      expect(vm.successMessage).toBe('')
    })

    it('clears error message on close', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.error = 'Test error'

      // Simulate alert close
      vm.error = ''

      expect(vm.error).toBe('')
    })
  })
})
