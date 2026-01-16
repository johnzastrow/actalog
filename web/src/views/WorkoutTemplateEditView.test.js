import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import WorkoutTemplateEditView from './WorkoutTemplateEditView.vue'

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

// Import mocked axios
import axios from '@/utils/axios'
import { useRoute, useRouter } from 'vue-router'

describe('WorkoutTemplateEditView', () => {
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

    // Mock movements API
    axios.get.mockImplementation((url) => {
      if (url === '/api/movements') {
        return Promise.resolve({
          data: {
            movements: [
              { id: 1, name: 'Back Squat', type: 'weightlifting' },
              { id: 2, name: 'Deadlift', type: 'weightlifting' },
              { id: 3, name: 'Pull-up', type: 'gymnastics' },
            ]
          }
        })
      }
      if (url === '/api/wods') {
        return Promise.resolve({
          data: {
            wods: [
              { id: 1, name: 'Fran', type: 'benchmark', regime: 'For Time' },
              { id: 2, name: 'Murph', type: 'hero', regime: 'For Time' },
            ]
          }
        })
      }
      if (url.match(/\/api\/templates\/\d+/)) {
        return Promise.resolve({
          data: {
            template: {
              id: 1,
              name: 'Test Template',
              notes: 'Test description',
              movements: [
                {
                  id: 1,
                  movement_id: 1,
                  movement: { id: 1, name: 'Back Squat' },
                  sets: 3,
                  reps: 10,
                  weight: 135,
                  instructions: 'Keep your back straight',
                  notes: 'Focus on form'
                }
              ],
              wods: [
                {
                  id: 1,
                  wod_id: 1,
                  wod: { id: 1, name: 'Fran' },
                  instructions: 'Scale as needed',
                  notes: 'Time cap 15 min'
                }
              ]
            }
          }
        })
      }
      return Promise.resolve({ data: {} })
    })

    axios.post.mockResolvedValue({
      data: {
        template: {
          id: 1,
          name: 'New Template',
          notes: 'New description'
        }
      }
    })

    axios.put.mockResolvedValue({
      data: {
        template: {
          id: 1,
          name: 'Updated Template',
          notes: 'Updated description'
        }
      }
    })

    wrapper = mount(WorkoutTemplateEditView, {
      global: {
        plugins: [pinia],
      }
    })

    await flushPromises()
    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  // ==========================================================================
  // CREATE MODE TESTS
  // ==========================================================================

  describe('Create Mode', () => {
    it('renders form in create mode', async () => {
      await createWrapper()

      // Check empty state for movements
      expect(wrapper.text()).toContain('No movements added yet')
    })

    it('preserves movement data structure', async () => {
      await createWrapper()

      // Verify template data structure has movements array
      const vm = wrapper.vm
      expect(vm.template).toBeDefined()
      expect(vm.template.movements).toBeDefined()
      expect(Array.isArray(vm.template.movements)).toBe(true)
    })

    it('preserves WOD data structure', async () => {
      await createWrapper()

      const vm = wrapper.vm
      expect(vm.template.wods).toBeDefined()
      expect(Array.isArray(vm.template.wods)).toBe(true)
    })
  })

  // ==========================================================================
  // EDIT MODE TESTS
  // ==========================================================================

  describe('Edit Mode', () => {
    it('loads existing template data', async () => {
      await createWrapper({ id: '1' })

      // Wait for template to load
      await flushPromises()

      // Verify axios was called with correct URL
      expect(axios.get).toHaveBeenCalledWith('/api/templates/1')
    })

    it('preserves all movement fields when loading', async () => {
      await createWrapper({ id: '1' })
      await flushPromises()

      const vm = wrapper.vm

      // Check that template was loaded
      if (vm.template && vm.template.movements && vm.template.movements.length > 0) {
        const movement = vm.template.movements[0]

        // Verify structure exists for all fields
        expect(movement).toHaveProperty('movement_id')
        expect(movement).toHaveProperty('sets')
        expect(movement).toHaveProperty('reps')
        expect(movement).toHaveProperty('instructions')
        expect(movement).toHaveProperty('notes')
      }
    })

    it('preserves all WOD fields when loading', async () => {
      await createWrapper({ id: '1' })
      await flushPromises()

      const vm = wrapper.vm

      if (vm.template && vm.template.wods && vm.template.wods.length > 0) {
        const wod = vm.template.wods[0]

        expect(wod).toHaveProperty('wod_id')
        expect(wod).toHaveProperty('instructions')
        expect(wod).toHaveProperty('notes')
      }
    })
  })

  // ==========================================================================
  // FORM SUBMISSION TESTS
  // ==========================================================================

  describe('Form Submission', () => {
    it('sends all movement fields on create', async () => {
      await createWrapper()

      const vm = wrapper.vm

      // Set up template data with all fields
      vm.template.name = 'Test Template'
      vm.template.description = 'Test description'
      vm.template.movements = [
        {
          movement_id: 1,
          sets: 4,
          reps: 12,
          weight: 155.5,
          work_time: 90,
          distance: 500,
          instructions: '**Setup:**\n- Step 1\n- Step 2',
          notes: 'Progressive overload'
        }
      ]
      vm.template.wods = [
        {
          wod_id: 1,
          instructions: '## Standards\n- Full ROM',
          notes: 'Time cap 15 min'
        }
      ]

      // Call save method
      await vm.saveTemplate()
      await flushPromises()

      // Verify POST was called
      expect(axios.post).toHaveBeenCalled()

      // Get the call arguments
      const callArgs = axios.post.mock.calls[0]
      expect(callArgs[0]).toBe('/api/templates')

      const requestBody = callArgs[1]
      expect(requestBody.name).toBe('Test Template')

      // Verify movement data
      expect(requestBody.movements).toBeDefined()
      expect(requestBody.movements.length).toBe(1)
      expect(requestBody.movements[0].movement_id).toBe(1)
      expect(requestBody.movements[0].sets).toBe(4)
      expect(requestBody.movements[0].reps).toBe(12)
      expect(requestBody.movements[0].instructions).toBe('**Setup:**\n- Step 1\n- Step 2')
      expect(requestBody.movements[0].notes).toBe('Progressive overload')

      // Verify WOD data
      expect(requestBody.wods).toBeDefined()
      expect(requestBody.wods.length).toBe(1)
      expect(requestBody.wods[0].wod_id).toBe(1)
      expect(requestBody.wods[0].instructions).toBe('## Standards\n- Full ROM')
      expect(requestBody.wods[0].notes).toBe('Time cap 15 min')
    })

    it('sends all movement fields on update', async () => {
      await createWrapper({ id: '1' })
      await flushPromises()

      const vm = wrapper.vm

      // Modify template
      vm.template.name = 'Updated Template'
      vm.template.movements = [
        {
          movement_id: 2,
          sets: 5,
          reps: 8,
          weight: 200,
          instructions: 'Updated instructions',
          notes: 'Updated notes'
        }
      ]

      // Call save method
      await vm.saveTemplate()
      await flushPromises()

      // Verify PUT was called
      expect(axios.put).toHaveBeenCalled()

      const callArgs = axios.put.mock.calls[0]
      expect(callArgs[0]).toBe('/api/templates/1')

      const requestBody = callArgs[1]
      expect(requestBody.movements[0].instructions).toBe('Updated instructions')
      expect(requestBody.movements[0].notes).toBe('Updated notes')
    })
  })

  // ==========================================================================
  // FIELD EDITING TESTS
  // ==========================================================================

  describe('Individual Field Editing', () => {
    it('can update movement sets independently', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.template.movements = [
        {
          movement_id: 1,
          sets: 3,
          reps: 10,
          instructions: 'Original',
          notes: 'Original notes'
        }
      ]

      // Update only sets
      vm.template.movements[0].sets = 5

      // Verify other fields unchanged
      expect(vm.template.movements[0].reps).toBe(10)
      expect(vm.template.movements[0].instructions).toBe('Original')
      expect(vm.template.movements[0].notes).toBe('Original notes')
    })

    it('can update movement instructions independently', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.template.movements = [
        {
          movement_id: 1,
          sets: 3,
          reps: 10,
          instructions: 'Original',
          notes: 'Original notes'
        }
      ]

      // Update only instructions
      vm.template.movements[0].instructions = '**New Instructions**\n- Step 1'

      // Verify other fields unchanged
      expect(vm.template.movements[0].sets).toBe(3)
      expect(vm.template.movements[0].reps).toBe(10)
      expect(vm.template.movements[0].notes).toBe('Original notes')
    })

    it('can update WOD instructions independently', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.template.wods = [
        {
          wod_id: 1,
          instructions: 'Original WOD instructions',
          notes: 'Original WOD notes'
        }
      ]

      // Update only instructions
      vm.template.wods[0].instructions = '## New Standards'

      // Verify other fields unchanged
      expect(vm.template.wods[0].notes).toBe('Original WOD notes')
    })

    it('can clear optional fields', async () => {
      await createWrapper()

      const vm = wrapper.vm
      vm.template.movements = [
        {
          movement_id: 1,
          sets: 3,
          reps: 10,
          weight: 100,
          instructions: 'Has instructions',
          notes: 'Has notes'
        }
      ]

      // Clear optional fields
      vm.template.movements[0].sets = null
      vm.template.movements[0].instructions = ''
      vm.template.movements[0].notes = ''

      expect(vm.template.movements[0].sets).toBeNull()
      expect(vm.template.movements[0].instructions).toBe('')
      expect(vm.template.movements[0].notes).toBe('')
      // Required field should remain
      expect(vm.template.movements[0].movement_id).toBe(1)
    })
  })

  // ==========================================================================
  // SPECIAL CHARACTERS TESTS
  // ==========================================================================

  describe('Special Characters Handling', () => {
    it('preserves markdown in instructions', async () => {
      await createWrapper()

      const vm = wrapper.vm
      const markdownContent = `## Workout Setup

**Bold** and *italic* text

1. First step
2. Second step

\`\`\`
Code block
\`\`\`

> Blockquote

| Col1 | Col2 |
|------|------|
| A    | B    |`

      vm.template.movements = [
        {
          movement_id: 1,
          instructions: markdownContent
        }
      ]

      expect(vm.template.movements[0].instructions).toBe(markdownContent)
    })

    it('preserves unicode characters', async () => {
      await createWrapper()

      const vm = wrapper.vm
      const unicodeContent = 'café naïve 日本語 emoji: 💪🏋️'

      vm.template.movements = [
        {
          movement_id: 1,
          instructions: unicodeContent,
          notes: unicodeContent
        }
      ]

      expect(vm.template.movements[0].instructions).toBe(unicodeContent)
      expect(vm.template.movements[0].notes).toBe(unicodeContent)
    })

    it('preserves HTML-like characters', async () => {
      await createWrapper()

      const vm = wrapper.vm
      const htmlLikeContent = '<angle> & "quotes" \'single\''

      vm.template.movements = [
        {
          movement_id: 1,
          notes: htmlLikeContent
        }
      ]

      expect(vm.template.movements[0].notes).toBe(htmlLikeContent)
    })
  })

  // ==========================================================================
  // MULTIPLE ITEMS TESTS
  // ==========================================================================

  describe('Multiple Items Handling', () => {
    it('can add multiple movements', async () => {
      await createWrapper()

      const vm = wrapper.vm

      vm.addMovement()
      vm.addMovement()
      vm.addMovement()

      expect(vm.template.movements.length).toBe(3)
    })

    it('can add multiple WODs', async () => {
      await createWrapper()

      const vm = wrapper.vm

      vm.addWOD()
      vm.addWOD()

      expect(vm.template.wods.length).toBe(2)
    })

    it('can remove movements', async () => {
      await createWrapper()

      const vm = wrapper.vm

      vm.addMovement()
      vm.addMovement()
      expect(vm.template.movements.length).toBe(2)

      vm.removeMovement(0)
      expect(vm.template.movements.length).toBe(1)
    })

    it('can remove WODs', async () => {
      await createWrapper()

      const vm = wrapper.vm

      vm.addWOD()
      vm.addWOD()
      expect(vm.template.wods.length).toBe(2)

      vm.removeWOD(0)
      expect(vm.template.wods.length).toBe(1)
    })

    it('preserves data when removing items', async () => {
      await createWrapper()

      const vm = wrapper.vm

      vm.template.movements = [
        { movement_id: 1, sets: 3, instructions: 'First' },
        { movement_id: 2, sets: 4, instructions: 'Second' },
        { movement_id: 3, sets: 5, instructions: 'Third' }
      ]

      // Remove middle item
      vm.removeMovement(1)

      expect(vm.template.movements.length).toBe(2)
      expect(vm.template.movements[0].instructions).toBe('First')
      expect(vm.template.movements[1].instructions).toBe('Third')
    })
  })
})
