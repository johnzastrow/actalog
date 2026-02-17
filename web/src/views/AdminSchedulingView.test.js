import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminSchedulingView from './AdminSchedulingView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
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

describe('AdminSchedulingView', () => {
  let wrapper
  let mockRouter

  const mockOrganizations = [
    { id: 1, name: 'CrossFit Gym 1' },
    { id: 2, name: 'CrossFit Gym 2' }
  ]

  const mockLocations = [
    { id: 1, organization_id: 1, name: 'Main Room', capacity: 30, is_active: true },
    { id: 2, organization_id: 1, name: 'Weight Room', capacity: 15, is_active: true }
  ]

  const mockTemplates = [
    {
      id: 1,
      organization_id: 1,
      name: 'Morning WOD',
      description: 'Daily CrossFit workout',
      duration_minutes: 60,
      capacity: 20,
      is_active: true
    }
  ]

  const mockSessions = [
    {
      id: 1,
      name: 'Morning WOD',
      start_time: '2026-01-20T09:00:00Z',
      end_time: '2026-01-20T10:00:00Z',
      status: 'scheduled',
      capacity: 20,
      reservation_count: 5
    }
  ]

  const mockCoaches = [
    {
      id: 1,
      user_id: 101,
      organization_id: 1,
      user_name: 'John Coach',
      user_email: 'john@example.com',
      is_active: true
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

    wrapper = mount(AdminSchedulingView, {
      global: {
        plugins: [pinia],
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>'
          },
          AdminHeader: {
            template: '<div class="admin-header"><h1>{{ title }}</h1><p>{{ subtitle }}</p></div>',
            props: ['title', 'subtitle', 'breadcrumbs']
          }
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    axios.get.mockResolvedValue({ data: { organizations: [] } })
    axios.post.mockResolvedValue({ data: {} })
    axios.put.mockResolvedValue({ data: {} })
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
    it('renders admin scheduling view', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Class Scheduling')
      expect(wrapper.text()).toContain('Manage templates, sessions, and coaches')
    })

    it('renders organization selector', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Select Gym')
    })

    it('renders tabs for different sections', () => {
      createWrapper()

      expect(wrapper.text()).toContain('Locations')
      expect(wrapper.text()).toContain('Classes')
      expect(wrapper.text()).toContain('Sessions')
      expect(wrapper.text()).toContain('Coaches')
    })
  })

  // ==========================================================================
  // ORGANIZATION LOADING TESTS
  // ==========================================================================

  describe('Organization Loading', () => {
    it('fetches organizations on mount', async () => {
      axios.get.mockResolvedValueOnce({ data: { organizations: mockOrganizations } })

      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/admin/organizations')
    })

    it('auto-selects first organization', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: mockLocations } })
        .mockResolvedValueOnce({ data: { templates: mockTemplates } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { coaches: mockCoaches } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.selectedOrgId).toBe(1)
    })
  })

  // ==========================================================================
  // LOCATIONS TAB TESTS
  // ==========================================================================

  describe('Locations Tab', () => {
    it('fetches locations when organization selected', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: mockLocations } })
        .mockResolvedValueOnce({ data: { templates: mockTemplates } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { coaches: mockCoaches } })

      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/gyms/1/locations?include_inactive=true')
    })

    it('displays locations', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: mockLocations } })
        .mockResolvedValueOnce({ data: { templates: mockTemplates } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { coaches: mockCoaches } })

      createWrapper()
      await flushPromises()

      expect(wrapper.text()).toContain('Main Room')
    })

    it('can open add location dialog', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: [] } })
        .mockResolvedValueOnce({ data: { templates: [] } })
        .mockResolvedValueOnce({ data: { sessions: [] } })
        .mockResolvedValueOnce({ data: { coaches: [] } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openLocationDialog()

      expect(vm.locationDialog).toBe(true)
      expect(vm.editingLocation).toBe(null)
    })

    it('can open edit location dialog', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: mockLocations } })
        .mockResolvedValueOnce({ data: { templates: [] } })
        .mockResolvedValueOnce({ data: { sessions: [] } })
        .mockResolvedValueOnce({ data: { coaches: [] } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.editLocation(mockLocations[0])

      expect(vm.locationDialog).toBe(true)
      expect(vm.editingLocation).toEqual(mockLocations[0])
    })

    it('saves new location', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: [] } })
        .mockResolvedValueOnce({ data: { templates: [] } })
        .mockResolvedValueOnce({ data: { sessions: [] } })
        .mockResolvedValueOnce({ data: { coaches: [] } })

      createWrapper()
      await flushPromises()

      axios.post.mockResolvedValueOnce({ data: { location: { id: 3 } } })
      axios.get.mockResolvedValueOnce({ data: { locations: mockLocations } })

      const vm = wrapper.vm
      vm.locationForm = { name: 'New Room', capacity: 25 }
      await vm.saveLocation()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/admin/gyms/1/locations', {
        name: 'New Room',
        capacity: 25
      })
    })

    it('deletes location with confirm', async () => {
      // Mock window.confirm
      vi.spyOn(window, 'confirm').mockReturnValue(true)

      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: mockLocations } })
        .mockResolvedValueOnce({ data: { templates: [] } })
        .mockResolvedValueOnce({ data: { sessions: [] } })
        .mockResolvedValueOnce({ data: { coaches: [] } })

      createWrapper()
      await flushPromises()

      axios.delete.mockResolvedValueOnce({ data: {} })
      axios.get.mockResolvedValueOnce({ data: { locations: [] } })

      const vm = wrapper.vm
      await vm.deleteLocation(mockLocations[0])
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/admin/gyms/1/locations/1')
    })
  })

  // ==========================================================================
  // TEMPLATES TAB TESTS
  // ==========================================================================

  describe('Templates Tab', () => {
    it('fetches templates when organization selected', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: mockLocations } })
        .mockResolvedValueOnce({ data: { templates: mockTemplates } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { coaches: mockCoaches } })

      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/gyms/1/templates?include_inactive=true')
    })

    it('can open add template dialog', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: [] } })
        .mockResolvedValueOnce({ data: { templates: [] } })
        .mockResolvedValueOnce({ data: { sessions: [] } })
        .mockResolvedValueOnce({ data: { coaches: [] } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openTemplateDialog()

      expect(vm.templateDialog).toBe(true)
      expect(vm.editingTemplate).toBe(null)
    })

    it('opens template dialog for new template', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: [] } })
        .mockResolvedValueOnce({ data: { templates: [] } })
        .mockResolvedValueOnce({ data: { sessions: [] } })
        .mockResolvedValueOnce({ data: { coaches: [] } })

      createWrapper()
      await flushPromises()
      await flushPromises()

      const vm = wrapper.vm
      vm.openTemplateDialog()

      expect(vm.templateDialog).toBe(true)
      expect(vm.editingTemplate).toBe(null)
    })
  })

  // ==========================================================================
  // SESSIONS TAB TESTS
  // ==========================================================================

  describe('Sessions Tab', () => {
    it('fetches sessions when organization selected', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: mockLocations } })
        .mockResolvedValueOnce({ data: { templates: mockTemplates } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { coaches: mockCoaches } })

      createWrapper()
      await flushPromises()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith(
        expect.stringContaining('/api/gyms/1/sessions')
      )
    })

    it('can open add session dialog', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: [] } })
        .mockResolvedValueOnce({ data: { templates: [] } })
        .mockResolvedValueOnce({ data: { sessions: [] } })
        .mockResolvedValueOnce({ data: { coaches: [] } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openSessionDialog()

      expect(vm.sessionDialog).toBe(true)
      expect(vm.editingSession).toBe(null)
    })

    it('can cancel session with prompt', async () => {
      // Mock window.prompt
      vi.spyOn(window, 'prompt').mockReturnValue('Test cancellation reason')

      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: [] } })
        .mockResolvedValueOnce({ data: { templates: [] } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { coaches: [] } })

      createWrapper()
      await flushPromises()
      await flushPromises()

      axios.post.mockResolvedValueOnce({ data: {} })
      axios.get.mockResolvedValueOnce({ data: { sessions: [] } })

      const vm = wrapper.vm
      await vm.cancelSession(mockSessions[0])
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/admin/gyms/1/sessions/1/cancel', { reason: 'Test cancellation reason' })
    })
  })

  // ==========================================================================
  // COACHES TAB TESTS
  // ==========================================================================

  describe('Coaches Tab', () => {
    it('fetches coaches when organization selected', async () => {
      axios.get.mockResolvedValueOnce({ data: { organizations: mockOrganizations } })

      createWrapper()
      await flushPromises()

      // Set org directly and fetch coaches
      const vm = wrapper.vm
      vm.selectedOrgId = 1
      axios.get.mockResolvedValueOnce({ data: { coaches: mockCoaches } })
      await vm.fetchCoaches()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/admin/gyms/1/coaches?include_inactive=true')
    })

    it('can open add coach dialog', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: [] } })
        .mockResolvedValueOnce({ data: { templates: [] } })
        .mockResolvedValueOnce({ data: { sessions: [] } })
        .mockResolvedValueOnce({ data: { coaches: [] } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.openCoachDialog()

      expect(vm.coachDialog).toBe(true)
    })

    it('removes coach with confirm', async () => {
      // Mock window.confirm
      vi.spyOn(window, 'confirm').mockReturnValue(true)

      axios.get.mockResolvedValueOnce({ data: { organizations: mockOrganizations } })

      createWrapper()
      await flushPromises()

      // Set org and coaches directly
      const vm = wrapper.vm
      vm.selectedOrgId = 1

      axios.delete.mockResolvedValueOnce({ data: {} })
      axios.get.mockResolvedValueOnce({ data: { coaches: [] } })

      await vm.removeCoach(mockCoaches[0])
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/admin/gyms/1/coaches/1')
    })
  })

  // ==========================================================================
  // ROSTER TESTS
  // ==========================================================================

  describe('Roster Management', () => {
    it('can open roster dialog', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: [] } })
        .mockResolvedValueOnce({ data: { templates: [] } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { coaches: [] } })

      createWrapper()
      await flushPromises()

      const mockRoster = [
        { id: 1, user_id: 101, user_name: 'John Doe', status: 'reserved' }
      ]
      axios.get.mockResolvedValueOnce({ data: { roster: mockRoster } })

      const vm = wrapper.vm
      await vm.viewRoster(mockSessions[0])
      await flushPromises()

      expect(vm.rosterDialog).toBe(true)
      expect(axios.get).toHaveBeenCalledWith('/api/admin/sessions/1/roster')
    })

    it('can check in reservation', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: [] } })
        .mockResolvedValueOnce({ data: { templates: [] } })
        .mockResolvedValueOnce({ data: { sessions: mockSessions } })
        .mockResolvedValueOnce({ data: { coaches: [] } })

      createWrapper()
      await flushPromises()

      const mockRoster = [
        { id: 1, user_id: 101, user_name: 'John Doe', status: 'reserved' }
      ]
      axios.get.mockResolvedValueOnce({ data: { roster: mockRoster } })

      const vm = wrapper.vm
      await vm.viewRoster(mockSessions[0])
      await flushPromises()

      axios.post.mockResolvedValueOnce({ data: {} })
      axios.get.mockResolvedValueOnce({ data: { roster: [{ ...mockRoster[0], status: 'checked_in' }] } })

      await vm.checkInReservation(mockRoster[0])
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/admin/sessions/1/check-in/1')
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('handles fetch error gracefully', async () => {
      // Override the default mock entirely for this test
      axios.get.mockReset()
      axios.get.mockRejectedValue({
        response: { data: { error: 'Failed to fetch organizations' } }
      })

      createWrapper()
      for (let i = 0; i < 5; i++) await flushPromises()

      // Should not throw and component should still render
      expect(wrapper.exists()).toBe(true)
      // Error should be set
      const vm = wrapper.vm
      expect(vm.error).toBe('Failed to fetch organizations')
    })

    it('shows error when location save fails', async () => {
      axios.get
        .mockResolvedValueOnce({ data: { organizations: mockOrganizations } })
        .mockResolvedValueOnce({ data: { locations: [] } })
        .mockResolvedValueOnce({ data: { templates: [] } })
        .mockResolvedValueOnce({ data: { sessions: [] } })
        .mockResolvedValueOnce({ data: { coaches: [] } })

      createWrapper()
      for (let i = 0; i < 5; i++) await flushPromises()

      // Reset error to null after initial load
      const vm = wrapper.vm
      vm.error = null

      axios.post.mockRejectedValueOnce({
        response: { data: { error: 'Location name already exists' } }
      })

      vm.locationForm = { name: 'Duplicate', capacity: 20 }
      await vm.saveLocation()
      await flushPromises()

      expect(vm.error).toBe('Location name already exists')
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

    it('returns correct session status color', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getSessionStatusColor('scheduled')).toBe('primary')
      expect(vm.getSessionStatusColor('in_progress')).toBe('warning')
      expect(vm.getSessionStatusColor('completed')).toBe('success')
      expect(vm.getSessionStatusColor('cancelled')).toBe('error')
    })

    it('returns correct reservation status color', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.getReservationStatusColor('reserved')).toBe('primary')
      expect(vm.getReservationStatusColor('checked_in')).toBe('success')
      expect(vm.getReservationStatusColor('cancelled')).toBe('grey')
      expect(vm.getReservationStatusColor('no_show')).toBe('error')
    })
  })

  // ==========================================================================
  // TAB NAVIGATION TESTS
  // ==========================================================================

  describe('Tab Navigation', () => {
    it('starts on locations tab', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.activeTab).toBe('locations')
    })

    it('can switch tabs', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.activeTab = 'templates'
      expect(vm.activeTab).toBe('templates')

      vm.activeTab = 'sessions'
      expect(vm.activeTab).toBe('sessions')

      vm.activeTab = 'coaches'
      expect(vm.activeTab).toBe('coaches')
    })
  })
})
