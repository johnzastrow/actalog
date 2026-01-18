import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import MyReservationsView from './MyReservationsView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
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
    RouterLink: {
      template: '<a><slot /></a>'
    }
  }
})

// Import mocked modules
import axios from '@/utils/axios'
import { useRouter } from 'vue-router'

describe('MyReservationsView', () => {
  let wrapper
  let mockRouter

  const mockReservations = [
    {
      id: 1,
      session_id: 101,
      session_name: 'Morning WOD',
      session_start_time: '2026-01-20T09:00:00Z',
      organization_name: 'CrossFit Gym',
      location_name: 'Main Room',
      status: 'reserved'
    },
    {
      id: 2,
      session_id: 102,
      session_name: 'Evening Strength',
      session_start_time: '2026-01-21T18:00:00Z',
      organization_name: 'CrossFit Gym',
      location_name: 'Weight Room',
      status: 'checked_in'
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

    wrapper = mount(MyReservationsView, {
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
    axios.get.mockResolvedValue({ data: { reservations: [] } })
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
    it('renders my reservations view', () => {
      createWrapper()

      expect(wrapper.text()).toContain('My Reservations')
      expect(wrapper.text()).toContain('Upcoming class reservations')
    })

    it('renders upcoming classes section', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Upcoming Classes')
    })

    it('shows empty state when no reservations', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: [] } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain("You don't have any upcoming class reservations")
    })

    it('shows browse classes button in empty state', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: [] } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Browse Classes')
    })
  })

  // ==========================================================================
  // DATA LOADING TESTS
  // ==========================================================================

  describe('Data Loading', () => {
    it('fetches reservations on mount', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: mockReservations } })

      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/users/me/reservations/upcoming')
    })

    it('displays reservations when loaded', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: mockReservations } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Morning WOD')
      expect(wrapper.text()).toContain('Evening Strength')
    })

    it('shows reservation count in header', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: mockReservations } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Upcoming Classes (2)')
    })

    it('handles fetch error gracefully', async () => {
      axios.get.mockRejectedValueOnce({
        response: { data: { error: 'Failed to fetch reservations' } }
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.error).toBe('Failed to fetch reservations')
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

      resolvePromise({ data: { reservations: [] } })
      await flushPromises()

      expect(vm.loading).toBe(false)
    })
  })

  // ==========================================================================
  // RESERVATION DETAILS TESTS
  // ==========================================================================

  describe('Reservation Details', () => {
    it('displays session name', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: mockReservations } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Morning WOD')
    })

    it('displays organization name', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: mockReservations } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('CrossFit Gym')
    })

    it('displays location name', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: mockReservations } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Main Room')
    })

    it('displays reservation status', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: mockReservations } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Reserved')
      expect(wrapper.text()).toContain('Checked In')
    })
  })

  // ==========================================================================
  // CANCEL RESERVATION TESTS
  // ==========================================================================

  describe('Cancel Reservation', () => {
    it('calls cancelReservation with reservation', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: mockReservations } })

      createWrapper()
      await flushPromises()

      axios.delete.mockResolvedValueOnce({ data: {} })
      axios.get.mockResolvedValueOnce({ data: { reservations: [] } })

      const vm = wrapper.vm
      await vm.cancelReservation(mockReservations[0])
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/sessions/101/reserve')
    })

    it('shows success message after cancellation', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: mockReservations } })

      createWrapper()
      await flushPromises()

      axios.delete.mockResolvedValueOnce({ data: {} })
      axios.get.mockResolvedValueOnce({ data: { reservations: [] } })

      const vm = wrapper.vm
      await vm.cancelReservation(mockReservations[0])
      await flushPromises()

      expect(vm.successMessage).toContain('Reservation cancelled')
      expect(vm.successMessage).toContain('Morning WOD')
    })

    it('handles cancel error', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: mockReservations } })

      createWrapper()
      await flushPromises()

      axios.delete.mockRejectedValueOnce({
        response: { data: { error: 'Cannot cancel within 2 hours of start' } }
      })

      const vm = wrapper.vm
      await vm.cancelReservation(mockReservations[0])
      await flushPromises()

      expect(vm.error).toBe('Cannot cancel within 2 hours of start')
    })

    it('refreshes reservations after successful cancellation', async () => {
      axios.get.mockResolvedValueOnce({ data: { reservations: mockReservations } })

      createWrapper()
      await flushPromises()

      axios.delete.mockResolvedValueOnce({ data: {} })
      axios.get.mockResolvedValueOnce({ data: { reservations: [] } })

      const vm = wrapper.vm
      await vm.cancelReservation(mockReservations[0])
      await flushPromises()

      // Should have fetched reservations twice (initial + after cancel)
      expect(axios.get).toHaveBeenCalledTimes(2)
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

    it('returns correct status color for reserved', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getStatusColor('reserved')).toBe('primary')
    })

    it('returns correct status color for checked_in', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getStatusColor('checked_in')).toBe('success')
    })

    it('returns correct status color for attended', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getStatusColor('attended')).toBe('success')
    })

    it('returns correct status color for cancelled', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getStatusColor('cancelled')).toBe('grey')
    })

    it('returns correct status color for no_show', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getStatusColor('no_show')).toBe('error')
    })

    it('returns correct status icon for reserved', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getStatusIcon('reserved')).toBe('mdi-calendar-clock')
    })

    it('returns correct status icon for checked_in', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getStatusIcon('checked_in')).toBe('mdi-check-circle')
    })

    it('formats status correctly', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.formatStatus('reserved')).toBe('Reserved')
      expect(vm.formatStatus('checked_in')).toBe('Checked In')
      expect(vm.formatStatus('attended')).toBe('Attended')
      expect(vm.formatStatus('cancelled')).toBe('Cancelled')
      expect(vm.formatStatus('no_show')).toBe('No Show')
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
      axios.get.mockResolvedValueOnce({ data: { reservations: mockReservations } })

      createWrapper()
      await flushPromises()

      axios.delete.mockResolvedValueOnce({ data: {} })
      axios.get.mockResolvedValueOnce({ data: { reservations: [] } })

      const vm = wrapper.vm
      await vm.cancelReservation(mockReservations[0])
      await flushPromises()

      expect(vm.successMessage).toBeTruthy()

      vm.successMessage = null
      expect(vm.successMessage).toBe(null)
    })
  })
})
