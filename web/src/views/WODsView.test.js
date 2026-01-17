import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import WODsView from './WODsView.vue'

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
const mockPush = vi.fn()
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: mockPush,
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

describe('WODsView', () => {
  let wrapper

  const mockStandardWODs = [
    {
      id: 1,
      name: 'Fran',
      type: 'girl',
      regime: 'for_time',
      score_type: 'time',
      description: '21-15-9 Thrusters and Pull-ups',
      is_standard: true
    },
    {
      id: 2,
      name: 'Murph',
      type: 'hero',
      regime: 'for_time',
      score_type: 'time',
      description: '1 mile run, 100 pull-ups, 200 push-ups, 300 squats, 1 mile run',
      is_standard: true
    }
  ]

  const mockCustomWODs = [
    {
      id: 10,
      name: 'My Custom WOD',
      type: 'self-created',
      regime: 'amrap',
      score_type: 'rounds_reps',
      description: 'Custom workout',
      is_standard: false
    }
  ]

  const setupDefaultMocks = () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/wods/standard') {
        return Promise.resolve({ data: { wods: [] } })
      }
      if (url === '/api/wods/my-wods') {
        return Promise.resolve({ data: { wods: [] } })
      }
      return Promise.resolve({ data: {} })
    })
  }

  const setupMocksWithData = () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/wods/standard') {
        // Clone to prevent mutation by component
        return Promise.resolve({ data: { wods: mockStandardWODs.map(w => ({...w})) } })
      }
      if (url === '/api/wods/my-wods') {
        // Clone to prevent mutation by component
        return Promise.resolve({ data: { wods: mockCustomWODs.map(w => ({...w})) } })
      }
      return Promise.resolve({ data: {} })
    })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(WODsView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-card-title': { template: '<div><slot /></div>' },
          'v-card-text': { template: '<div><slot /></div>' },
          'v-card-actions': { template: '<div><slot /></div>' },
          'v-tabs': { template: '<div><slot /></div>' },
          'v-tab': { template: '<button><slot /></button>' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-icon': { template: '<i></i>' },
          'v-chip': { template: '<span><slot /></span>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-progress-circular': { template: '<div></div>' },
          'v-dialog': { template: '<div><slot /></div>' },
          'v-text-field': { template: '<input />' },
          'v-textarea': { template: '<textarea></textarea>' },
          'v-select': { template: '<select></select>' },
          'v-spacer': { template: '<div></div>' },
          'v-avatar': { template: '<div><slot /></div>' },
          'v-bottom-navigation': { template: '<div><slot /></div>' },
          'markdown-renderer': true
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    mockPush.mockClear()
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
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/wods/standard')
      expect(axios.get).toHaveBeenCalledWith('/api/wods/my-wods')
    })

    it('initializes with standard tab active', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.activeWODTab).toBe('standard')
    })

    it('stores fetched WODs', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.standardWODs).toHaveLength(2)
      expect(vm.customWODs).toHaveLength(1)
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
    it('handles fetch error with response', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue({
        response: { status: 500, data: { message: 'Server error' } }
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Server error')

      consoleSpy.mockRestore()
    })

    it('handles fetch error without response', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue({ request: {} })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toContain('No response from server')

      consoleSpy.mockRestore()
    })

    it('handles generic fetch error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue(new Error('Network error'))

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Failed to fetch WODs')

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // DISPLAYED WODS TESTS
  // ==========================================================================

  describe('Displayed WODs', () => {
    it('shows standard WODs when standard tab is active', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.activeWODTab = 'standard'

      expect(vm.displayedWODs).toEqual(mockStandardWODs)
    })

    it('shows custom WODs when custom tab is active', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.activeWODTab = 'custom'

      expect(vm.displayedWODs).toEqual(mockCustomWODs)
    })
  })

  // ==========================================================================
  // WOD FORM TESTS
  // ==========================================================================

  describe('WOD Form', () => {
    it('initializes with empty form', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.wodForm.name).toBe('')
      expect(vm.wodForm.description).toBe('')
      expect(vm.wodForm.type).toBe(null)
    })

    it('validates name is required', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm
      vm.wodForm.name = ''
      vm.wodForm.description = 'Test description'

      await vm.saveWOD()

      expect(vm.error).toContain('name and description are required')
      expect(axios.post).not.toHaveBeenCalled()
    })

    it('validates description is required', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm
      vm.wodForm.name = 'Test WOD'
      vm.wodForm.description = ''

      await vm.saveWOD()

      expect(vm.error).toContain('name and description are required')
    })

    it('creates WOD successfully', async () => {
      axios.post.mockResolvedValue({
        data: { wod: { id: 100, name: 'New WOD', is_standard: false } }
      })
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.wodForm.name = 'New WOD'
      vm.wodForm.description = 'Test description'

      await vm.saveWOD()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/wods', expect.objectContaining({
        name: 'New WOD',
        description: 'Test description'
      }))
      expect(vm.success).toContain('created successfully')
      expect(vm.activeWODTab).toBe('custom')
    })

    it('resets form on cancel', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm
      vm.wodForm.name = 'Test'
      vm.wodForm.description = 'Test desc'
      vm.createWODDialog = true

      vm.cancelWODForm()

      expect(vm.wodForm.name).toBe('')
      expect(vm.wodForm.description).toBe('')
      expect(vm.createWODDialog).toBe(false)
    })
  })

  // ==========================================================================
  // WOD ACTIONS TESTS
  // ==========================================================================

  describe('WOD Actions', () => {
    it('selects WOD for detail view', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.selectWOD(mockStandardWODs[0])

      expect(vm.selectedWOD).toEqual(mockStandardWODs[0])
      expect(vm.detailDialog).toBe(true)
    })

    it('navigates to edit WOD page', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.editWOD(mockCustomWODs[0])

      expect(mockPush).toHaveBeenCalledWith('/wods/10/edit')
    })

    it('deletes WOD after confirmation', async () => {
      const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
      axios.delete.mockResolvedValue({})
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      await vm.deleteWOD(10)
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/wods/10')
      expect(vm.customWODs.find(w => w.id === 10)).toBeUndefined()
      expect(vm.success).toContain('deleted successfully')

      confirmSpy.mockRestore()
    })

    it('does not delete WOD if not confirmed', async () => {
      const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      await vm.deleteWOD(10)

      expect(axios.delete).not.toHaveBeenCalled()

      confirmSpy.mockRestore()
    })

    it('handles delete error', async () => {
      const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.delete.mockRejectedValue({
        response: { data: { message: 'Delete failed' } }
      })
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      await vm.deleteWOD(10)
      await flushPromises()

      expect(vm.error).toBe('Delete failed')

      confirmSpy.mockRestore()
      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // FORMAT FUNCTIONS TESTS
  // ==========================================================================

  describe('Format Functions', () => {
    it('formatWODType capitalizes words', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.formatWODType('self_created')).toBe('Self Created')
      expect(vm.formatWODType('hero')).toBe('Hero')
      expect(vm.formatWODType(null)).toBe('')
    })

    it('formatRegime handles acronyms', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.formatRegime('emom')).toBe('EMOM')
      expect(vm.formatRegime('amrap')).toBe('AMRAP')
      expect(vm.formatRegime('for_time')).toBe('For Time')
      expect(vm.formatRegime(null)).toBe('')
    })

    it('formatScoreType formats score types', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.formatScoreType('time')).toBe('Time')
      expect(vm.formatScoreType('rounds_reps')).toBe('Rounds + Reps')
      expect(vm.formatScoreType('max_weight')).toBe('Max Weight')
      expect(vm.formatScoreType(null)).toBe('')
    })

    it('truncateText truncates long text', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm
      const longText = 'This is a very long text that should be truncated'

      expect(vm.truncateText(longText, 20)).toBe('This is a very long ...')
      expect(vm.truncateText('Short', 20)).toBe('Short')
      expect(vm.truncateText(null, 20)).toBe(null)
    })
  })

  // ==========================================================================
  // OPTIONS DATA TESTS
  // ==========================================================================

  describe('Options Data', () => {
    it('has WOD type options', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.wodTypes.length).toBeGreaterThan(0)
      expect(vm.wodTypes.some(t => t.value === 'benchmark')).toBe(true)
      expect(vm.wodTypes.some(t => t.value === 'hero')).toBe(true)
    })

    it('has regime options', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.regimes.length).toBeGreaterThan(0)
      expect(vm.regimes.some(r => r.value === 'emom')).toBe(true)
      expect(vm.regimes.some(r => r.value === 'amrap')).toBe(true)
    })

    it('has score type options', () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.scoreTypes.length).toBeGreaterThan(0)
      expect(vm.scoreTypes.some(s => s.value === 'time')).toBe(true)
      expect(vm.scoreTypes.some(s => s.value === 'rounds_reps')).toBe(true)
    })
  })
})
