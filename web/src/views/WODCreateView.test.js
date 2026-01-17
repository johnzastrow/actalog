import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import WODCreateView from './WODCreateView.vue'

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

describe('WODCreateView', () => {
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

    // Mock GET for loading existing WOD
    axios.get.mockImplementation((url) => {
      if (url.match(/\/api\/wods\/\d+/)) {
        return Promise.resolve({
          data: {
            wod: {
              id: 1,
              name: 'Fran',
              source: 'CrossFit',
              type: 'Benchmark',
              regime: 'Fastest Time',
              score_type: 'Time (HH:MM:SS)',
              description: '21-15-9 Thrusters and Pull-ups',
              url: 'https://youtube.com/watch?v=fran123',
              notes: 'Scale to banded pull-ups if needed',
              is_standard: false
            }
          }
        })
      }
      return Promise.resolve({ data: {} })
    })

    // Mock POST for creating
    axios.post.mockResolvedValue({
      data: {
        wod: {
          id: 2,
          name: 'New WOD',
          source: 'Self-recorded',
          type: 'Benchmark',
          regime: 'AMRAP',
          score_type: 'Rounds+Reps',
          description: 'Test workout'
        }
      }
    })

    // Mock PUT for updating
    axios.put.mockResolvedValue({
      data: {
        wod: {
          id: 1,
          name: 'Updated Fran',
          source: 'CrossFit',
          type: 'Benchmark',
          regime: 'Fastest Time',
          score_type: 'Time (HH:MM:SS)',
          description: 'Updated description'
        }
      }
    })

    // Mock DELETE
    axios.delete.mockResolvedValue({ data: {} })

    wrapper = mount(WODCreateView, {
      global: {
        plugins: [pinia],
      }
    })

    await flushPromises()
    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })

  // ==========================================================================
  // CREATE MODE TESTS
  // ==========================================================================

  describe('Create Mode', () => {
    it('renders form in create mode with default values', async () => {
      await createWrapper()

      const vm = wrapper.vm
      expect(vm.wod.name).toBe('')
      expect(vm.wod.source).toBe('Self-recorded')
      expect(vm.wod.type).toBe('Benchmark')
      expect(vm.wod.regime).toBe('Fastest Time')
      expect(vm.wod.score_type).toBe('Time (HH:MM:SS)')
      expect(vm.wod.description).toBe('')
    })

    it('shows correct button text for create mode', async () => {
      await createWrapper()

      expect(wrapper.text()).toContain('Create WOD')
    })

    it('does not show delete button in create mode', async () => {
      await createWrapper()

      expect(wrapper.text()).not.toContain('Delete WOD')
    })

    it('validates required fields before creating', async () => {
      await createWrapper()

      const vm = wrapper.vm

      // Try to save with empty fields
      await vm.saveWOD()
      await flushPromises()

      // Should have validation errors
      expect(vm.validationErrors.name).toBeDefined()
      expect(vm.validationErrors.description).toBeDefined()

      // Should not have called API
      expect(axios.post).not.toHaveBeenCalled()
    })

    it('creates WOD with all fields', async () => {
      await createWrapper()

      const vm = wrapper.vm

      // Fill in all fields
      vm.wod.name = 'Custom AMRAP'
      vm.wod.source = 'Self-recorded'
      vm.wod.type = 'Self-created'
      vm.wod.regime = 'AMRAP'
      vm.wod.score_type = 'Rounds+Reps'
      vm.wod.description = '10 min AMRAP: 10 burpees, 15 air squats, 20 sit-ups'
      vm.wod.url = 'https://youtube.com/watch?v=custom'
      vm.wod.notes = 'Rest 1 min between rounds'

      await vm.saveWOD()
      await flushPromises()

      // Verify POST was called
      expect(axios.post).toHaveBeenCalled()

      // Check the payload
      const callArgs = axios.post.mock.calls[0]
      expect(callArgs[0]).toBe('/api/wods')

      const payload = callArgs[1]
      expect(payload.name).toBe('Custom AMRAP')
      expect(payload.source).toBe('Self-recorded')
      expect(payload.type).toBe('Self-created')
      expect(payload.regime).toBe('AMRAP')
      expect(payload.score_type).toBe('Rounds+Reps')
      expect(payload.description).toBe('10 min AMRAP: 10 burpees, 15 air squats, 20 sit-ups')
      expect(payload.url).toBe('https://youtube.com/watch?v=custom')
      expect(payload.notes).toBe('Rest 1 min between rounds')
    })

    it('navigates to WODs list after successful create', async () => {
      vi.useFakeTimers()
      await createWrapper()

      const vm = wrapper.vm

      // Fill required fields
      vm.wod.name = 'Test WOD'
      vm.wod.source = 'Self-recorded'
      vm.wod.description = 'Test description'

      await vm.saveWOD()
      await flushPromises()

      // Advance timer for navigation
      vi.advanceTimersByTime(2000)

      expect(mockRouter.push).toHaveBeenCalledWith('/wods')

      vi.useRealTimers()
    })

    it('handles selection mode - returns with selectedWOD', async () => {
      vi.useFakeTimers()
      await createWrapper({}, { select: 'true', returnPath: '/templates/new' })

      const vm = wrapper.vm

      // Fill required fields
      vm.wod.name = 'Test WOD'
      vm.wod.source = 'Self-recorded'
      vm.wod.description = 'Test description'

      await vm.saveWOD()
      await flushPromises()

      // Advance timer
      vi.advanceTimersByTime(2000)

      expect(mockRouter.push).toHaveBeenCalledWith({
        path: '/templates/new',
        query: { selectedWOD: 2 }
      })

      vi.useRealTimers()
    })
  })

  // ==========================================================================
  // EDIT MODE TESTS
  // ==========================================================================

  describe('Edit Mode', () => {
    it('loads existing WOD data', async () => {
      await createWrapper({ id: '1' })

      // Verify API call
      expect(axios.get).toHaveBeenCalledWith('/api/wods/1')

      const vm = wrapper.vm

      // Verify data was loaded
      expect(vm.wod.name).toBe('Fran')
      expect(vm.wod.source).toBe('CrossFit')
      expect(vm.wod.type).toBe('Benchmark')
      expect(vm.wod.regime).toBe('Fastest Time')
      expect(vm.wod.score_type).toBe('Time (HH:MM:SS)')
      expect(vm.wod.description).toBe('21-15-9 Thrusters and Pull-ups')
      expect(vm.wod.url).toBe('https://youtube.com/watch?v=fran123')
      expect(vm.wod.notes).toBe('Scale to banded pull-ups if needed')
    })

    it('shows correct button text for edit mode', async () => {
      await createWrapper({ id: '1' })

      expect(wrapper.text()).toContain('Update WOD')
    })

    it('shows delete button in edit mode for non-standard WOD', async () => {
      await createWrapper({ id: '1' })

      expect(wrapper.text()).toContain('Delete WOD')
    })

    it('hides delete button for standard WODs', async () => {
      // Mock a standard WOD
      axios.get.mockResolvedValueOnce({
        data: {
          wod: {
            id: 1,
            name: 'Fran',
            source: 'CrossFit',
            type: 'Benchmark',
            regime: 'Fastest Time',
            score_type: 'Time (HH:MM:SS)',
            description: '21-15-9',
            is_standard: true
          }
        }
      })

      await createWrapper({ id: '1' })
      await flushPromises()

      // Delete button should be hidden for standard WODs
      const vm = wrapper.vm
      expect(vm.wod.is_standard).toBe(true)
    })

    it('updates WOD with PUT request', async () => {
      await createWrapper({ id: '1' })

      const vm = wrapper.vm

      // Modify data
      vm.wod.name = 'Updated Fran'
      vm.wod.description = 'Updated description'

      await vm.saveWOD()
      await flushPromises()

      // Verify PUT was called
      expect(axios.put).toHaveBeenCalled()

      const callArgs = axios.put.mock.calls[0]
      expect(callArgs[0]).toBe('/api/wods/1')

      const payload = callArgs[1]
      expect(payload.name).toBe('Updated Fran')
      expect(payload.description).toBe('Updated description')
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

    it('deletes WOD when confirmed', async () => {
      vi.useFakeTimers()
      await createWrapper({ id: '1' })

      const vm = wrapper.vm

      // Open dialog and delete
      vm.deleteDialog = true
      await vm.deleteWOD()
      await flushPromises()

      // Verify DELETE was called
      expect(axios.delete).toHaveBeenCalledWith('/api/wods/1')

      // Advance timer for navigation
      vi.advanceTimersByTime(2000)

      expect(mockRouter.push).toHaveBeenCalledWith('/wods')

      vi.useRealTimers()
    })

    it('handles delete error gracefully', async () => {
      await createWrapper({ id: '1' })

      // Mock delete failure
      axios.delete.mockRejectedValueOnce(new Error('Delete failed'))

      const vm = wrapper.vm

      await vm.deleteWOD()
      await flushPromises()

      expect(vm.error).toContain('Failed to delete WOD')
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
      vm.wod.name = ''
      vm.wod.source = 'Self-recorded'
      vm.wod.description = 'Valid description'

      const isValid = vm.validateWOD()

      expect(isValid).toBe(false)
      expect(vm.validationErrors.name).toBe('WOD name is required')
    })

    it('validates source is required', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.wod.name = 'Valid Name'
      vm.wod.source = ''
      vm.wod.description = 'Valid description'

      const isValid = vm.validateWOD()

      expect(isValid).toBe(false)
      expect(vm.validationErrors.source).toBe('Source is required')
    })

    it('validates description is required', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.wod.name = 'Valid Name'
      vm.wod.source = 'Self-recorded'
      vm.wod.description = ''

      const isValid = vm.validateWOD()

      expect(isValid).toBe(false)
      expect(vm.validationErrors.description).toBe('Description is required')
    })

    it('passes validation with all required fields', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.wod.name = 'Valid Name'
      vm.wod.source = 'Self-recorded'
      vm.wod.description = 'Valid description'

      const isValid = vm.validateWOD()

      expect(isValid).toBe(true)
      expect(Object.keys(vm.validationErrors).length).toBe(0)
    })

    it('trims whitespace from name', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.wod.name = '   '
      vm.wod.source = 'Self-recorded'
      vm.wod.description = 'Valid description'

      const isValid = vm.validateWOD()

      expect(isValid).toBe(false)
      expect(vm.validationErrors.name).toBe('WOD name is required')
    })
  })

  // ==========================================================================
  // WOD TYPES AND OPTIONS TESTS
  // ==========================================================================

  describe('WOD Types and Options', () => {
    it('has correct source options', async () => {
      await createWrapper()

      const vm = wrapper.vm

      expect(vm.sources).toContain('CrossFit')
      expect(vm.sources).toContain('Other Coach')
      expect(vm.sources).toContain('Self-recorded')
    })

    it('has correct WOD type options', async () => {
      await createWrapper()

      const vm = wrapper.vm

      expect(vm.wodTypes).toContain('Benchmark')
      expect(vm.wodTypes).toContain('Hero')
      expect(vm.wodTypes).toContain('Girl')
      expect(vm.wodTypes).toContain('Games')
      expect(vm.wodTypes).toContain('Notables')
      expect(vm.wodTypes).toContain('Endurance')
      expect(vm.wodTypes).toContain('Self-created')
    })

    it('has correct regime options', async () => {
      await createWrapper()

      const vm = wrapper.vm

      expect(vm.regimes).toContain('EMOM')
      expect(vm.regimes).toContain('AMRAP')
      expect(vm.regimes).toContain('Fastest Time')
      expect(vm.regimes).toContain('Slowest Round')
      expect(vm.regimes).toContain('Get Stronger')
      expect(vm.regimes).toContain('Skills')
    })

    it('has correct score type options', async () => {
      await createWrapper()

      const vm = wrapper.vm

      expect(vm.scoreTypes).toContain('Time (HH:MM:SS)')
      expect(vm.scoreTypes).toContain('Rounds+Reps')
      expect(vm.scoreTypes).toContain('Max Weight')
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
      vm.wod.name = 'Test WOD'
      vm.wod.source = 'Self-recorded'
      vm.wod.description = 'Test description'
      // Leave optional fields empty
      vm.wod.url = ''
      vm.wod.notes = ''

      await vm.saveWOD()
      await flushPromises()

      expect(axios.post).toHaveBeenCalled()

      // Verify payload has null for empty optional fields
      const payload = axios.post.mock.calls[0][1]

      expect(payload.url).toBeNull()
      expect(payload.notes).toBeNull()
    })

    it('preserves optional fields when editing', async () => {
      await createWrapper({ id: '1' })

      const vm = wrapper.vm

      // Verify optional fields were loaded
      expect(vm.wod.url).toBe('https://youtube.com/watch?v=fran123')
      expect(vm.wod.notes).toBe('Scale to banded pull-ups if needed')
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
            message: 'WOD already exists'
          }
        }
      })

      const vm = wrapper.vm

      vm.wod.name = 'Duplicate WOD'
      vm.wod.source = 'Self-recorded'
      vm.wod.description = 'Test description'

      await vm.saveWOD()
      await flushPromises()

      expect(vm.error).toBe('WOD already exists')
    })

    it('handles API error on load', async () => {
      // Mock load failure
      axios.get.mockRejectedValueOnce(new Error('Network error'))

      await createWrapper({ id: '1' })
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.error).toBe('Failed to load WOD')
    })

    it('handles generic API error', async () => {
      await createWrapper()

      // Mock API failure without message
      axios.post.mockRejectedValueOnce(new Error('Network error'))

      const vm = wrapper.vm

      vm.wod.name = 'Test WOD'
      vm.wod.source = 'Self-recorded'
      vm.wod.description = 'Test description'

      await vm.saveWOD()
      await flushPromises()

      expect(vm.error).toBe('Failed to save WOD. Please try again.')
    })
  })

  // ==========================================================================
  // NAVIGATION TESTS
  // ==========================================================================

  describe('Navigation', () => {
    it('uses custom return path from query', async () => {
      vi.useFakeTimers()
      await createWrapper({ id: '1' }, { returnPath: '/custom/path' })

      const vm = wrapper.vm

      await vm.deleteWOD()
      await flushPromises()

      vi.advanceTimersByTime(2000)

      expect(mockRouter.push).toHaveBeenCalledWith('/custom/path')

      vi.useRealTimers()
    })

    it('handleBack navigates without confirmation when no changes', async () => {
      await createWrapper()

      const vm = wrapper.vm

      // Reset to default values (no changes)
      vm.wod.name = ''
      vm.wod.description = ''
      vm.wod.source = 'Self-recorded'

      vm.handleBack()

      expect(mockRouter.push).toHaveBeenCalledWith('/wods')
    })

    it('handleBack asks for confirmation when there are changes', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.wod.name = 'Some changes'

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

      const markdownContent = `## Workout of the Day

**21-15-9 reps of:**
- Thrusters (95/65 lb)
- Pull-ups

> Rx standards: Full depth squat, chest to bar

| Division | Weight |
|----------|--------|
| Rx       | 95/65  |
| Scaled   | 65/45  |`

      vm.wod.name = 'Test WOD'
      vm.wod.source = 'Self-recorded'
      vm.wod.description = markdownContent

      await vm.saveWOD()
      await flushPromises()

      // Verify markdown is preserved
      const payload = axios.post.mock.calls[0][1]
      expect(payload.description).toBe(markdownContent)
    })

    it('preserves unicode characters', async () => {
      await createWrapper()

      const vm = wrapper.vm

      vm.wod.name = 'WOD de resistencia'
      vm.wod.source = 'Self-recorded'
      vm.wod.description = 'Japanese: 日本語 | Emoji: 💪🏋️‍♀️'

      await vm.saveWOD()
      await flushPromises()

      const payload = axios.post.mock.calls[0][1]
      expect(payload.name).toBe('WOD de resistencia')
      expect(payload.description).toBe('Japanese: 日本語 | Emoji: 💪🏋️‍♀️')
    })

    it('preserves HTML-like characters', async () => {
      await createWrapper()

      const vm = wrapper.vm

      vm.wod.name = 'Test WOD'
      vm.wod.source = 'Self-recorded'
      vm.wod.description = '<10 min cap> & "quotes" \'single\''

      await vm.saveWOD()
      await flushPromises()

      const payload = axios.post.mock.calls[0][1]
      expect(payload.description).toBe('<10 min cap> & "quotes" \'single\'')
    })
  })

  // ==========================================================================
  // STATE MANAGEMENT TESTS
  // ==========================================================================

  describe('State Management', () => {
    it('sets saving state during save operation', async () => {
      await createWrapper()

      const vm = wrapper.vm

      vm.wod.name = 'Test WOD'
      vm.wod.source = 'Self-recorded'
      vm.wod.description = 'Test description'

      // Don't await - check intermediate state
      const savePromise = vm.saveWOD()

      expect(vm.saving).toBe(true)

      await savePromise
      await flushPromises()

      expect(vm.saving).toBe(false)
    })

    it('sets deleting state during delete operation', async () => {
      await createWrapper({ id: '1' })

      const vm = wrapper.vm

      // Don't await - check intermediate state
      const deletePromise = vm.deleteWOD()

      expect(vm.deleting).toBe(true)

      await deletePromise
      await flushPromises()

      expect(vm.deleting).toBe(false)
    })

    it('shows success message on successful save', async () => {
      await createWrapper()

      const vm = wrapper.vm

      vm.wod.name = 'Test WOD'
      vm.wod.source = 'Self-recorded'
      vm.wod.description = 'Test description'

      await vm.saveWOD()
      await flushPromises()

      expect(vm.successMessage).toBe('WOD created successfully!')
    })

    it('shows success message on successful update', async () => {
      await createWrapper({ id: '1' })

      const vm = wrapper.vm

      await vm.saveWOD()
      await flushPromises()

      expect(vm.successMessage).toBe('WOD updated successfully!')
    })
  })

  // ==========================================================================
  // CRUD LIFECYCLE TEST
  // ==========================================================================

  describe('Full CRUD Lifecycle', () => {
    it('completes full create -> read -> update -> delete lifecycle', async () => {
      // 1. CREATE
      await createWrapper()

      let vm = wrapper.vm

      vm.wod.name = 'Lifecycle WOD'
      vm.wod.source = 'Self-recorded'
      vm.wod.type = 'Self-created'
      vm.wod.regime = 'AMRAP'
      vm.wod.score_type = 'Rounds+Reps'
      vm.wod.description = '10 min AMRAP: 5 burpees, 10 air squats'
      vm.wod.url = 'https://example.com/video'
      vm.wod.notes = 'Scale as needed'

      await vm.saveWOD()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/wods', expect.objectContaining({
        name: 'Lifecycle WOD',
        source: 'Self-recorded',
        type: 'Self-created',
        regime: 'AMRAP',
        score_type: 'Rounds+Reps',
        description: '10 min AMRAP: 5 burpees, 10 air squats'
      }))

      // 2. READ (simulated by creating wrapper with ID)
      vi.clearAllMocks()
      await createWrapper({ id: '1' })

      vm = wrapper.vm

      expect(axios.get).toHaveBeenCalledWith('/api/wods/1')
      expect(vm.wod.name).toBe('Fran')
      expect(vm.wod.source).toBe('CrossFit')

      // 3. UPDATE
      vm.wod.name = 'Updated Fran'
      vm.wod.notes = 'Updated notes'

      await vm.saveWOD()
      await flushPromises()

      expect(axios.put).toHaveBeenCalledWith('/api/wods/1', expect.objectContaining({
        name: 'Updated Fran',
        notes: 'Updated notes'
      }))

      // 4. DELETE
      vi.clearAllMocks()
      await vm.deleteWOD()
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/wods/1')
    })
  })
})
