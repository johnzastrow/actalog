import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ScheduleView from './ScheduleView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn()
  }
}))

// Mock vue-router
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: vi.fn(),
      back: vi.fn(),
    })),
    useRoute: vi.fn(() => ({
      params: { gym_id: null }
    })),
    RouterLink: {
      template: '<a><slot /></a>'
    }
  }
})

// Import mocked modules
import axios from '@/utils/axios'
import { useRouter } from 'vue-router'

describe('ScheduleView', () => {
  let wrapper
  let mockRouter

  const mockOrganizations = [
    { id: 1, name: 'CrossFit Gym 1' },
    { id: 2, name: 'CrossFit Gym 2' }
  ]

  const mockSessions = [
    {
      id: 1,
      name: 'Morning WOD',
      start_time: '2026-01-18T09:00:00Z',
      end_time: '2026-01-18T10:00:00Z',
      status: 'scheduled',
      capacity: 20,
      reservation_count: 5,
      available_spots: 15,
      current_user_status: null,
      location_name: 'Main Room'
    },
    {
      id: 2,
      name: 'Afternoon Strength',
      start_time: '2026-01-18T17:00:00Z',
      end_time: '2026-01-18T18:00:00Z',
      status: 'scheduled',
      capacity: 15,
      reservation_count: 15,
      available_spots: 0,
      current_user_status: null,
      location_name: 'Weight Room'
    }
  ]

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    mockRouter = {
      push: vi.fn(),
      back: vi.fn()
    }
    useRouter.mockReturnValue(mockRouter)

    wrapper = mount(ScheduleView, {
      global: {
        plugins: [pinia],
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>'
          }
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: { organizations: [], sessions: [] } })
    axios.post.mockResolvedValue({ data: {} })
    axios.delete.mockResolvedValue({ data: {} })
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
  })

  // ==========================================================================
  // RENDER TESTS
  // ==========================================================================

  describe('Rendering', () => {
    it('renders schedule view', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Class Schedule')
      expect(wrapper.text()).toContain('Book your classes')
    })

    it('renders organization selector', async () => {
      axios.get.mockResolvedValue({ data: { organizations: [{ id: 1, name: 'Gym A' }, { id: 2, name: 'Gym B' }], sessions: [] } })
      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Select Gym')
    })

    it('renders date navigation', () => {
      createWrapper()

      expect(wrapper.find('.text-subtitle-1').exists()).toBe(true)
    })

    it('shows info message when no organization selected', async () => {
      axios.get.mockResolvedValue({ data: { organizations: [{ id: 1, name: 'Gym A' }, { id: 2, name: 'Gym B' }], sessions: [] } })
      createWrapper()
      await flushPromises()

      wrapper.vm.selectedOrgId = null
      await wrapper.vm.$nextTick()

      expect(wrapper.text()).toContain('Please select a gym to view the class schedule')
    })
  })

  // ==========================================================================
  // DATA LOADING TESTS
  // ==========================================================================

  describe('Data Loading', () => {
    it('fetches organizations on mount', async () => {
      axios.get.mockResolvedValueOnce({ data: { organizations: mockOrganizations } })

      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/admin/organizations')
    })

    it('auto-selects first organization', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.selectedOrgId).toBe(1)
    })

    it('fetches sessions when organization selected', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith(
        expect.stringContaining('/api/gyms/1/sessions')
      )
    })

    it('handles fetch error gracefully', async () => {
      axios.get.mockRejectedValueOnce({
        response: { data: { error: 'Failed to fetch organizations' } }
      })

      createWrapper()
      await flushPromises()

      // Should not throw and component should still render
      expect(wrapper.exists()).toBe(true)
    })
  })

  // ==========================================================================
  // RESERVATION TESTS
  // ==========================================================================

  describe('Reservations', () => {
    it('calls makeReservation with session', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      axios.post.mockResolvedValueOnce({ data: {} })
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      const vm = wrapper.vm
      await vm.makeReservation(mockSessions[0])
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/sessions/1/reserve')
    })

    it('shows success message after reservation', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      axios.post.mockResolvedValueOnce({ data: {} })
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      const vm = wrapper.vm
      await vm.makeReservation(mockSessions[0])
      await flushPromises()

      expect(vm.successMessage).toContain('Successfully reserved')
    })

    it('handles reservation error', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      axios.post.mockRejectedValueOnce({
        response: { data: { error: 'Session is full' } }
      })

      const vm = wrapper.vm
      await vm.makeReservation(mockSessions[0])
      await flushPromises()

      expect(vm.error).toBe('Session is full')
    })

    it('calls cancelReservation with session', async () => {
      const sessionsWithReservation = [
        { ...mockSessions[0], current_user_status: 'reserved' }
      ]

      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { sessions: sessionsWithReservation } })

      createWrapper()
      await flushPromises()

      axios.delete.mockResolvedValueOnce({ data: {} })
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      const vm = wrapper.vm
      await vm.cancelReservation(sessionsWithReservation[0])
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/sessions/1/reserve')
    })
  })

  // ==========================================================================
  // DATE NAVIGATION TESTS
  // ==========================================================================

  describe('Date Navigation', () => {
    it('navigates to previous week', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const originalStart = new Date(vm.startDate)

      axios.get.mockResolvedValueOnce({ data: { sessions: [] } })

      vm.previousWeek()
      await flushPromises()

      const expectedDate = new Date(originalStart)
      expectedDate.setDate(expectedDate.getDate() - 7)

      expect(vm.startDate.toDateString()).toBe(expectedDate.toDateString())
    })

    it('navigates to next week', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const originalStart = new Date(vm.startDate)

      axios.get.mockResolvedValueOnce({ data: { sessions: [] } })

      vm.nextWeek()
      await flushPromises()

      const expectedDate = new Date(originalStart)
      expectedDate.setDate(expectedDate.getDate() + 7)

      expect(vm.startDate.toDateString()).toBe(expectedDate.toDateString())
    })
  })

  // ==========================================================================
  // HELPER FUNCTION TESTS
  // ==========================================================================

  describe('Helper Functions', () => {
    it('formats date range correctly', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const start = new Date('2026-01-18')
      const end = new Date('2026-01-24')

      const result = vm.formatDateRange(start, end)
      expect(result).toContain('Jan')
    })

    it('formats time correctly', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const result = vm.formatTime('2026-01-18T09:00:00Z')
      expect(result).toBeTruthy()
    })

    it('returns correct session card class for cancelled session', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const cancelledSession = { status: 'cancelled' }
      expect(vm.getSessionCardClass(cancelledSession)).toBe('bg-grey-lighten-3')
    })

    it('returns correct session card class for reserved session', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const reservedSession = { status: 'scheduled', current_user_status: 'reserved' }
      expect(vm.getSessionCardClass(reservedSession)).toBe('border-primary')
    })

    it('returns empty class for normal session', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const normalSession = { status: 'scheduled', current_user_status: null }
      expect(vm.getSessionCardClass(normalSession)).toBe('')
    })
  })

  // ==========================================================================
  // GROUPED SESSIONS TESTS
  // ==========================================================================

  describe('Grouped Sessions', () => {
    it('groups sessions by date', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.groupedSessions).toBeTruthy()
      expect(Array.isArray(vm.groupedSessions)).toBe(true)
    })

    it('creates entries for all days in range', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { sessions: [] } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      // Default range is 7 days
      expect(vm.groupedSessions.length).toBe(7)
    })
  })
})
