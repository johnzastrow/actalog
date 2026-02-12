import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import CoachDashboardView from './CoachDashboardView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn()
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
    RouterLink: {
      template: '<a><slot /></a>'
    }
  }
})

// Import mocked modules
import axios from '@/utils/axios'

describe('CoachDashboardView', () => {
  let wrapper

  const mockSessions = [
    {
      id: 1,
      name: 'Morning WOD',
      start_time: '2026-01-20T09:00:00Z',
      end_time: '2026-01-20T10:00:00Z',
      status: 'scheduled',
      capacity: 20,
      reservation_count: 5,
      location_name: 'Main Room'
    },
    {
      id: 2,
      name: 'Afternoon Strength',
      start_time: '2026-01-20T17:00:00Z',
      end_time: '2026-01-20T18:00:00Z',
      status: 'in_progress',
      capacity: 15,
      reservation_count: 10,
      location_name: 'Weight Room'
    }
  ]

  const mockRoster = [
    {
      id: 1,
      user_id: 101,
      user_name: 'John Doe',
      user_email: 'john@example.com',
      status: 'reserved'
    },
    {
      id: 2,
      user_id: 102,
      user_name: 'Jane Smith',
      user_email: 'jane@example.com',
      status: 'checked_in'
    }
  ]

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = mount(CoachDashboardView, {
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
    axios.get.mockResolvedValue({ data: { sessions: [] } })
    axios.post.mockResolvedValue({ data: {} })
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
    it('renders coach dashboard view', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Coach Dashboard')
      expect(wrapper.text()).toContain('Manage your assigned class sessions')
    })

    it('shows loading indicator during fetch', async () => {
      let resolvePromise
      axios.get.mockReturnValueOnce(
        new Promise((resolve) => {
          resolvePromise = resolve
        })
      )

      createWrapper()

      const vm = wrapper.vm
      expect(vm.loading).toBe(true)

      resolvePromise({ data: { sessions: [] } })
      await flushPromises()

      expect(vm.loading).toBe(false)
    })

    it('shows empty state when no sessions', async () => {
      axios.get.mockResolvedValueOnce({ data: { sessions: [] } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('No Upcoming Sessions')
      expect(wrapper.text()).toContain("You don't have any sessions assigned to you yet")
    })
  })

  // ==========================================================================
  // DATA LOADING TESTS
  // ==========================================================================

  describe('Data Loading', () => {
    it('fetches coach sessions on mount', async () => {
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/coaches/me/sessions')
    })

    it('displays sessions when loaded', async () => {
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Morning WOD')
      expect(wrapper.text()).toContain('Afternoon Strength')
    })

    it('shows session count in header', async () => {
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('My Sessions (2)')
    })

    it('handles fetch error gracefully', async () => {
      axios.get.mockRejectedValueOnce({
        response: { data: { error: 'Failed to fetch sessions' } }
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.error).toBe('Failed to fetch sessions')
    })
  })

  // ==========================================================================
  // SESSION DETAILS TESTS
  // ==========================================================================

  describe('Session Details', () => {
    it('displays session name', async () => {
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Morning WOD')
    })

    it('displays location name', async () => {
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Main Room')
    })

    it('displays reservation count', async () => {
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('5 / 20 enrolled')
    })

    it('displays session status', async () => {
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Scheduled')
      expect(wrapper.text()).toContain('In Progress')
    })
  })

  // ==========================================================================
  // ROSTER DIALOG TESTS
  // ==========================================================================

  describe('Roster Dialog', () => {
    it('opens roster dialog when View Roster clicked', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { roster: mockRoster } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.openRoster(mockSessions[0])
      await flushPromises()

      expect(vm.rosterDialog).toBe(true)
      expect(axios.get).toHaveBeenCalledWith('/api/coaches/sessions/1/roster')
    })

    it('displays roster members', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { roster: mockRoster } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.openRoster(mockSessions[0])
      await flushPromises()

      expect(vm.roster.length).toBe(2)
      expect(vm.roster[0].user_name).toBe('John Doe')
    })
  })

  // ==========================================================================
  // CHECK IN TESTS
  // ==========================================================================

  describe('Check In', () => {
    it('calls checkInReservation with reservation', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { roster: mockRoster } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.openRoster(mockSessions[0])
      await flushPromises()

      axios.post.mockResolvedValueOnce({ data: {} })

      await vm.checkInReservation(mockRoster[0])
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/coaches/sessions/1/check-in/1')
    })

    it('shows success message after check-in', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { roster: mockRoster } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.openRoster(mockSessions[0])
      await flushPromises()

      axios.post.mockResolvedValueOnce({ data: {} })

      await vm.checkInReservation(mockRoster[0])
      await flushPromises()

      expect(vm.successMessage).toContain('Checked in')
    })

    it('handles check-in error', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { roster: mockRoster } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.openRoster(mockSessions[0])
      await flushPromises()

      axios.post.mockRejectedValueOnce({
        response: { data: { error: 'Check-in failed' } }
      })

      await vm.checkInReservation(mockRoster[0])
      await flushPromises()

      expect(vm.error).toBe('Check-in failed')
    })
  })

  // ==========================================================================
  // NO SHOW TESTS
  // ==========================================================================

  describe('No Show', () => {
    it('calls markNoShow with reservation', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { roster: mockRoster } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.openRoster(mockSessions[0])
      await flushPromises()

      axios.post.mockResolvedValueOnce({ data: {} })

      await vm.markNoShow(mockRoster[0])
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/coaches/sessions/1/no-show/1')
    })

    it('shows success message after marking no-show', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { roster: mockRoster } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.openRoster(mockSessions[0])
      await flushPromises()

      axios.post.mockResolvedValueOnce({ data: {} })

      await vm.markNoShow(mockRoster[0])
      await flushPromises()

      expect(vm.successMessage).toContain('no-show')
    })

    it('handles no-show error', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { roster: mockRoster } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.openRoster(mockSessions[0])
      await flushPromises()

      axios.post.mockRejectedValueOnce({
        response: { data: { error: 'Failed to mark no-show' } }
      })

      await vm.markNoShow(mockRoster[0])
      await flushPromises()

      expect(vm.error).toBe('Failed to mark no-show')
    })
  })

  // ==========================================================================
  // COMPLETE SESSION TESTS
  // ==========================================================================

  describe('Complete Session', () => {
    it('opens complete dialog when confirm complete called', async () => {
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.confirmCompleteSession(mockSessions[1])

      expect(vm.completeDialog).toBe(true)
      expect(vm.sessionToComplete).toEqual(mockSessions[1])
    })

    it('calls completeSession API', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { sessions: [] } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.sessionToComplete = mockSessions[1]

      axios.post.mockResolvedValueOnce({ data: {} })

      await vm.completeSession()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/coaches/sessions/2/complete')
    })

    it('shows success message after completing session', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { sessions: [] } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.sessionToComplete = mockSessions[1]

      axios.post.mockResolvedValueOnce({ data: {} })

      await vm.completeSession()
      await flushPromises()

      expect(vm.successMessage).toContain('completed')
    })

    it('handles complete session error', async () => {
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.sessionToComplete = mockSessions[1]

      axios.post.mockRejectedValueOnce({
        response: { data: { error: 'Cannot complete session' } }
      })

      await vm.completeSession()
      await flushPromises()

      expect(vm.error).toBe('Cannot complete session')
    })
  })

  // ==========================================================================
  // HELPER FUNCTION TESTS
  // ==========================================================================

  describe('Helper Functions', () => {
    it('formats date time correctly', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const result = vm.formatDateTime('2026-01-20T09:00:00Z')
      expect(result).toBeTruthy()
      expect(typeof result).toBe('string')
    })

    it('formats time correctly', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const result = vm.formatTime('2026-01-20T09:00:00Z')
      expect(result).toBeTruthy()
      expect(typeof result).toBe('string')
    })

    it('returns correct status color for scheduled', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getStatusColor('scheduled')).toBe('primary')
    })

    it('returns correct status color for in_progress', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getStatusColor('in_progress')).toBe('warning')
    })

    it('returns correct status color for completed', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getStatusColor('completed')).toBe('success')
    })

    it('formats status correctly', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.formatStatus('scheduled')).toBe('Scheduled')
      expect(vm.formatStatus('in_progress')).toBe('In Progress')
      expect(vm.formatStatus('completed')).toBe('Completed')
    })

    it('returns correct session card class for cancelled', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getSessionCardClass({ status: 'cancelled' })).toBe('bg-grey-lighten-3')
    })

    it('returns correct session card class for in_progress', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getSessionCardClass({ status: 'in_progress' })).toBe('border-warning')
    })

    it('returns correct reservation status color', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getReservationStatusColor('reserved')).toBe('primary')
      expect(vm.getReservationStatusColor('checked_in')).toBe('success')
      expect(vm.getReservationStatusColor('no_show')).toBe('error')
    })

    it('formats reservation status correctly', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.formatReservationStatus('reserved')).toBe('Reserved')
      expect(vm.formatReservationStatus('checked_in')).toBe('Checked In')
      expect(vm.formatReservationStatus('no_show')).toBe('No Show')
    })

    it('gets initials correctly', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getInitials('John Doe')).toBe('JD')
      expect(vm.getInitials('Jane')).toBe('JA')
      expect(vm.getInitials(null)).toBe('?')
    })
  })

  // ==========================================================================
  // ALERT DISMISSAL TESTS
  // ==========================================================================

  describe('Alert Dismissal', () => {
    it('can dismiss error alert', async () => {
      axios.get.mockRejectedValueOnce({
        response: { data: { error: 'Test error' } }
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.error).toBe('Test error')

      vm.error = null
      expect(vm.error).toBe(null)
    })

    it('can dismiss success alert', async () => {
      axios.get.mockResolvedValueOnce({ data: { sessions: mockSessions } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.successMessage = 'Test success'
      expect(vm.successMessage).toBe('Test success')

      vm.successMessage = null
      expect(vm.successMessage).toBe(null)
    })
  })
})
