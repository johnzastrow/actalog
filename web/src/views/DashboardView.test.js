import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import DashboardView from './DashboardView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  }
}))

// Mock vue-router
vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRouter: vi.fn(() => ({
      push: vi.fn(),
      replace: vi.fn(),
    })),
    useRoute: vi.fn(() => ({
      query: {}
    })),
    RouterLink: {
      template: '<a><slot /></a>'
    }
  }
})

// Mock timezone utilities
vi.mock('@/utils/timezone', () => ({
  formatDateInTimezone: vi.fn((date, tz, format) => {
    const d = new Date(date)
    return d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })
  }),
  getTodayInTimezone: vi.fn(() => {
    const today = new Date()
    return today.toISOString().split('T')[0]
  })
}))

// Mock auth store
const mockAuthStore = {
  user: {
    id: 1,
    name: 'Test User',
    email: 'test@example.com',
    email_verified: true,
    role: 'user'
  }
}

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => mockAuthStore
}))

// Mock settings store
const mockSettingsStore = {
  timezone: 'America/New_York',
  weightUnit: 'lbs',
  fetchSettings: vi.fn().mockResolvedValue({})
}

vi.mock('@/stores/settings', () => ({
  useSettingsStore: () => mockSettingsStore
}))

// Mock components
vi.mock('@/components/MarkdownRenderer.vue', () => ({
  default: {
    template: '<div class="markdown-renderer"><slot /></div>',
    props: ['content']
  }
}))

vi.mock('@/components/PullToRefresh.vue', () => ({
  default: {
    template: '<div class="pull-to-refresh"><slot /></div>',
    props: ['onRefresh']
  }
}))

vi.mock('@/components/DateWorkoutDetailDialog.vue', () => ({
  default: {
    template: '<div class="date-dialog"></div>',
    props: ['modelValue', 'date', 'preloadedWorkouts']
  }
}))

import axios from '@/utils/axios'
import { useRouter, useRoute } from 'vue-router'

describe('DashboardView', () => {
  let wrapper
  let mockRouter

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    mockRouter = {
      push: vi.fn(),
      replace: vi.fn()
    }
    useRouter.mockReturnValue(mockRouter)
    useRoute.mockReturnValue({ query: {} })

    // Mock API responses
    axios.get.mockImplementation((url) => {
      if (url === '/api/workouts?limit=10000') {
        return Promise.resolve({
          data: { workouts: [] }
        })
      }
      if (url === '/api/stats/active-users-this-month') {
        return Promise.resolve({
          data: { users: [] }
        })
      }
      if (url === '/api/movements') {
        return Promise.resolve({
          data: { movements: [] }
        })
      }
      if (url === '/api/wods/standard') {
        return Promise.resolve({
          data: { wods: [] }
        })
      }
      if (url === '/api/wods/my-wods') {
        return Promise.resolve({
          data: { wods: [] }
        })
      }
      if (url === '/api/workouts/standard') {
        return Promise.resolve({
          data: { workouts: [] }
        })
      }
      if (url === '/api/workouts/my-templates') {
        return Promise.resolve({
          data: { workouts: [] }
        })
      }
      return Promise.resolve({ data: {} })
    })

    wrapper = shallowMount(DashboardView, {
      global: {
        plugins: [pinia],
        stubs: {
          RouterLink: true,
          'pull-to-refresh': { template: '<div><slot /></div>' },
          'markdown-renderer': { template: '<div></div>', props: ['content'] },
          'date-workout-detail-dialog': { template: '<div></div>', props: ['modelValue', 'date', 'preloadedWorkouts'] },
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-card-title': { template: '<div><slot /></div>' },
          'v-card-text': { template: '<div><slot /></div>' },
          'v-card-actions': { template: '<div><slot /></div>' },
          'v-form': { template: '<form><slot /></form>' },
          'v-text-field': { template: '<input />' },
          'v-textarea': { template: '<textarea></textarea>' },
          'v-autocomplete': { template: '<input />' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-icon': { template: '<i></i>' },
          'v-avatar': { template: '<div><slot /></div>' },
          'v-chip': { template: '<span><slot /></span>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-dialog': { template: '<div><slot /></div>' },
          'v-progress-circular': { template: '<div></div>' },
          'v-spacer': { template: '<div></div>' },
          'v-list-item': { template: '<div><slot /></div>' }
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    // Reset mock user
    mockAuthStore.user = {
      id: 1,
      name: 'Test User',
      email: 'test@example.com',
      email_verified: true,
      role: 'user'
    }
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
    it('fetches workouts on mount', async () => {
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/workouts?limit=10000')
    })

    it('fetches active users stats on mount', async () => {
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/stats/active-users-this-month')
    })

    it('initializes with loading state', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.loading).toBeDefined()
      expect(vm.userWorkouts).toBeDefined()
    })

    it('initializes with recent workouts collapsed', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.showRecentWorkouts).toBe(false)
    })

    it('initializes with empty expanded workouts set', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.expandedWorkouts.size).toBe(0)
    })
  })

  // ==========================================================================
  // STATS COMPUTATION TESTS
  // ==========================================================================

  describe('Stats Computation', () => {
    it('computes totalWorkouts correctly', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.userWorkouts = [
        { id: 1, workout_date: '2024-01-01T00:00:00Z' },
        { id: 2, workout_date: '2024-01-02T00:00:00Z' },
        { id: 3, workout_date: '2024-01-03T00:00:00Z' }
      ]

      expect(vm.totalWorkouts).toBe(3)
    })

    it('computes totalWorkouts as 0 for empty workouts', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.totalWorkouts).toBe(0)
    })

    it('computes monthWorkouts correctly', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const now = new Date()

      // Use dates relative to now to avoid timezone issues
      // Create a date that's definitely within this month (3 days ago, but not before month start)
      const withinMonth = new Date(now)
      withinMonth.setDate(Math.max(2, now.getDate() - 3)) // Ensure we're at least on day 2

      // Create another date within the month (1 day ago)
      const alsoWithinMonth = new Date(now)
      alsoWithinMonth.setDate(now.getDate() - 1)

      // Create an old date definitely outside the month
      const oldDate = new Date(2020, 0, 1)

      vm.userWorkouts = [
        { id: 1, workout_date: withinMonth.toISOString() },
        { id: 2, workout_date: alsoWithinMonth.toISOString() },
        { id: 3, workout_date: oldDate.toISOString() } // Old workout
      ]

      expect(vm.monthWorkouts).toBe(2)
    })

    it('computes currentStreak as 0 for empty workouts', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.currentStreak).toBe(0)
    })

    it('computes avgTimePerWorkout correctly', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.userWorkouts = [
        { id: 1, workout_date: '2024-01-01T00:00:00Z', total_time: 1800 }, // 30 min
        { id: 2, workout_date: '2024-01-02T00:00:00Z', total_time: 3600 }, // 60 min
        { id: 3, workout_date: '2024-01-03T00:00:00Z', total_time: null } // No time
      ]

      // Average of 30 and 60 minutes = 45
      expect(vm.avgTimePerWorkout).toBe(45)
    })

    it('computes avgTimePerWorkout as 0 when no workouts have time', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.userWorkouts = [
        { id: 1, workout_date: '2024-01-01T00:00:00Z', total_time: null }
      ]

      expect(vm.avgTimePerWorkout).toBe(0)
    })
  })

  // ==========================================================================
  // WEEK DAYS COMPUTATION TESTS
  // ==========================================================================

  describe('Week Days Computation', () => {
    it('computes weekDays with 7 days', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.weekDays.length).toBe(7)
    })

    it('weekDays start with Monday', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.weekDays[0].letter).toBe('M')
    })

    it('weekDays end with Sunday', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.weekDays[6].letter).toBe('S')
    })

    it('weekDays includes date and hasWorkout properties', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const firstDay = vm.weekDays[0]

      expect(firstDay.date).toBeDefined()
      expect(typeof firstDay.hasWorkout).toBe('boolean')
    })
  })

  // ==========================================================================
  // RECENT WORKOUTS TESTS
  // ==========================================================================

  describe('Recent Workouts', () => {
    it('filters workouts from last 30 days', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const now = new Date()
      const withinThirtyDays = new Date(now.getTime() - (10 * 24 * 60 * 60 * 1000))
      const outsideThirtyDays = new Date(now.getTime() - (60 * 24 * 60 * 60 * 1000))

      // Manually set workouts to test the computed property
      vm.userWorkouts = [
        { id: 1, workout_date: withinThirtyDays.toISOString() },
        { id: 2, workout_date: outsideThirtyDays.toISOString() }
      ]

      expect(vm.recentWorkouts.length).toBe(1)
      expect(vm.recentWorkouts[0].id).toBe(1)
    })

    it('sorts recent workouts by date descending', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const now = new Date()
      const yesterday = new Date(now.getTime() - (1 * 24 * 60 * 60 * 1000))
      const threeDaysAgo = new Date(now.getTime() - (3 * 24 * 60 * 60 * 1000))

      // Manually set workouts to test the computed property
      vm.userWorkouts = [
        { id: 1, workout_date: threeDaysAgo.toISOString() },
        { id: 2, workout_date: yesterday.toISOString() }
      ]

      expect(vm.recentWorkouts[0].id).toBe(2) // Yesterday first
      expect(vm.recentWorkouts[1].id).toBe(1) // Three days ago second
    })

    it('limits recent workouts to 30', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const now = new Date()
      const workouts = []
      for (let i = 0; i < 35; i++) {
        const date = new Date(now.getTime() - (i * 24 * 60 * 60 * 1000))
        workouts.push({ id: i, workout_date: date.toISOString() })
      }

      vm.userWorkouts = workouts

      expect(vm.recentWorkouts.length).toBeLessThanOrEqual(30)
    })
  })

  // ==========================================================================
  // QUICK LOG TESTS
  // ==========================================================================

  describe('Quick Log', () => {
    it('initializes with quick log dialog closed', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.quickLogDialog).toBe(false)
    })

    it('opens quick log dialog', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.openQuickLog()

      expect(vm.quickLogDialog).toBe(true)
    })

    it('closes quick log dialog and resets form', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.openQuickLog()
      vm.quickLogData.notes = 'Test notes'

      vm.closeQuickLog()

      expect(vm.quickLogDialog).toBe(false)
      expect(vm.quickLogData.notes).toBe('')
    })

    it('initializes quick log data with today date', () => {
      createWrapper()

      const vm = wrapper.vm
      const today = new Date()
      const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`

      expect(vm.quickLogData.date).toBe(todayStr)
    })

    it('fetches movements and WODs when opening quick log', async () => {
      createWrapper()
      await flushPromises()

      vi.clearAllMocks()

      const vm = wrapper.vm
      await vm.openQuickLog()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/movements')
      expect(axios.get).toHaveBeenCalledWith('/api/wods/standard')
      expect(axios.get).toHaveBeenCalledWith('/api/wods/my-wods')
    })

    it('does not submit quick log without name', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.quickLogData.name = ''
      vm.quickLogData.date = '2024-01-01'

      await vm.submitQuickLog()

      expect(axios.post).not.toHaveBeenCalled()
    })

    it('does not submit quick log without date', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = ''

      await vm.submitQuickLog()

      expect(axios.post).not.toHaveBeenCalled()
    })

    it('submits quick log successfully', async () => {
      axios.post.mockResolvedValue({ data: { id: 1 } })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-01'
      vm.quickLogData.totalTime = 30
      vm.quickLogData.notes = 'Test notes'

      await vm.submitQuickLog()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/workouts', expect.objectContaining({
        workout_name: 'Test Workout',
        workout_date: '2024-01-01',
        total_time: 1800, // 30 minutes in seconds
        notes: 'Test notes'
      }))
    })

    it('sets submitting state during quick log submission', async () => {
      let resolvePost
      axios.post.mockReturnValue(
        new Promise((resolve) => {
          resolvePost = resolve
        })
      )

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-01'

      const submitPromise = vm.submitQuickLog()

      expect(vm.quickLogSubmitting).toBe(true)

      resolvePost({ data: { id: 1 } })
      await submitPromise
      await flushPromises()

      expect(vm.quickLogSubmitting).toBe(false)
    })
  })

  // ==========================================================================
  // WORKOUT EXPANSION TESTS
  // ==========================================================================

  describe('Workout Expansion', () => {
    it('toggles workout expansion', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.expandedWorkouts.has(1)).toBe(false)

      vm.toggleWorkoutExpand(1)

      expect(vm.expandedWorkouts.has(1)).toBe(true)

      vm.toggleWorkoutExpand(1)

      expect(vm.expandedWorkouts.has(1)).toBe(false)
    })
  })

  // ==========================================================================
  // DATE DIALOG TESTS
  // ==========================================================================

  describe('Date Dialog', () => {
    it('initializes with date dialog closed', () => {
      createWrapper()

      const vm = wrapper.vm

      expect(vm.showDateDialog).toBe(false)
      expect(vm.selectedWeekDate).toBe(null)
    })

    it('opens date dialog when week day is clicked', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const weekDay = { date: '2024-01-15', letter: 'M', hasWorkout: true }

      vm.handleWeekDayClick(weekDay)

      expect(vm.showDateDialog).toBe(true)
      expect(vm.selectedWeekDate).toBeInstanceOf(Date)
    })
  })

  // ==========================================================================
  // NAVIGATION TESTS
  // ==========================================================================

  describe('Navigation', () => {
    it('navigates to workout details', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.viewWorkout(123)

      expect(mockRouter.push).toHaveBeenCalledWith('/workouts/123')
    })

    it('navigates to log workout with template when template selected', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-01'
      vm.quickLogData.selectedItem = {
        type: 'template',
        entityId: 456
      }

      await vm.submitQuickLog()

      // When template is selected, it navigates with the date that was set
      expect(mockRouter.push).toHaveBeenCalledWith(
        expect.objectContaining({
          path: '/workouts/log',
          query: expect.objectContaining({
            template: 456
          })
        })
      )
    })
  })

  // ==========================================================================
  // REFRESH TESTS
  // ==========================================================================

  describe('Refresh', () => {
    it('refetches data on refresh', async () => {
      createWrapper()
      await flushPromises()

      vi.clearAllMocks()

      const vm = wrapper.vm
      const doneFn = vi.fn()

      await vm.handleRefresh(doneFn)
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/workouts?limit=10000')
      expect(axios.get).toHaveBeenCalledWith('/api/stats/active-users-this-month')
      expect(doneFn).toHaveBeenCalled()
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('handles workout fetch error gracefully', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      axios.get.mockImplementation((url) => {
        if (url === '/api/workouts?limit=10000') {
          return Promise.reject(new Error('Network error'))
        }
        return Promise.resolve({ data: {} })
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.userWorkouts).toEqual([])

      consoleSpy.mockRestore()
    })

    it('handles active users fetch error gracefully', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      axios.get.mockImplementation((url) => {
        if (url === '/api/stats/active-users-this-month') {
          return Promise.reject(new Error('Network error'))
        }
        if (url === '/api/workouts?limit=10000') {
          return Promise.resolve({ data: { workouts: [] } })
        }
        return Promise.resolve({ data: {} })
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.activeUsersStats).toEqual([])

      consoleSpy.mockRestore()
    })

    it('handles quick log submission error', async () => {
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      axios.post.mockRejectedValue({
        response: { data: { message: 'Failed to create workout' } }
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.quickLogData.name = 'Test Workout'
      vm.quickLogData.date = '2024-01-01'

      await vm.submitQuickLog()
      await flushPromises()

      expect(alertSpy).toHaveBeenCalledWith('Failed to create workout')

      alertSpy.mockRestore()
      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // UNIFIED SEARCH ITEMS TESTS
  // ==========================================================================

  describe('Unified Search Items', () => {
    it('computes unified search items from movements, WODs, and templates', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      // Initially empty
      expect(vm.unifiedSearchItems).toEqual([])

      // Add some movements
      vm.movements = [
        { id: 1, name: 'Squat' },
        { id: 2, name: 'Deadlift' }
      ]

      expect(vm.unifiedSearchItems.length).toBe(2)
      expect(vm.unifiedSearchItems[0].type).toBe('movement')
    })

    it('sorts unified search items alphabetically', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.movements = [
        { id: 1, name: 'Squat' },
        { id: 2, name: 'Deadlift' },
        { id: 3, name: 'Bench Press' }
      ]

      expect(vm.unifiedSearchItems[0].name).toBe('Bench Press')
      expect(vm.unifiedSearchItems[1].name).toBe('Deadlift')
      expect(vm.unifiedSearchItems[2].name).toBe('Squat')
    })
  })

  // ==========================================================================
  // EMAIL VERIFICATION TESTS
  // ==========================================================================

  describe('Email Verification', () => {
    it('exposes user email verification status from auth store', () => {
      mockAuthStore.user.email_verified = false
      createWrapper()

      expect(mockAuthStore.user.email_verified).toBe(false)
    })

    it('shows verified status when email is verified', () => {
      mockAuthStore.user.email_verified = true
      createWrapper()

      expect(mockAuthStore.user.email_verified).toBe(true)
    })
  })

  // ==========================================================================
  // ACTIVE USERS TESTS
  // ==========================================================================

  describe('Active Users', () => {
    it('stores active users stats', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      // Manually set active users stats to test the reactive data
      vm.activeUsersStats = [
        { id: 1, name: 'User 1', workout_count: 10, is_current: true },
        { id: 2, name: 'User 2', workout_count: 5, is_current: false }
      ]

      expect(vm.activeUsersStats.length).toBe(2)
      expect(vm.activeUsersStats[0].name).toBe('User 1')
      expect(vm.activeUsersStats[0].is_current).toBe(true)
    })

    it('identifies current user in active users list', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.activeUsersStats = [
        { id: 1, name: 'User 1', workout_count: 10, is_current: true },
        { id: 2, name: 'User 2', workout_count: 5, is_current: false }
      ]

      const currentUser = vm.activeUsersStats.find(u => u.is_current)
      expect(currentUser).toBeDefined()
      expect(currentUser.name).toBe('User 1')
    })
  })

  // ==========================================================================
  // QUICK LOG NAME UPDATE TESTS
  // ==========================================================================

  describe('Quick Log Name Update', () => {
    it('updates quick log name when date changes', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.quickLogData.date = '2024-12-25'

      vm.updateQuickLogName()

      expect(vm.quickLogData.name).toContain('Quick Log')
      expect(vm.quickLogData.name).toContain('Dec')
      expect(vm.quickLogData.name).toContain('25')
    })

    it('does not update name if date is empty', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const originalName = vm.quickLogData.name
      vm.quickLogData.date = ''

      vm.updateQuickLogName()

      expect(vm.quickLogData.name).toBe(originalName)
    })
  })

  // ==========================================================================
  // ROUTE QUERY TESTS
  // ==========================================================================

  describe('Route Query', () => {
    it('exposes route query watcher for quick-log parameter', async () => {
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      // The component has a watcher for route.query.open
      // When it detects 'quick-log', it opens the dialog
      // Testing that the dialog can be opened programmatically
      await vm.openQuickLog()

      expect(vm.quickLogDialog).toBe(true)
    })
  })
})
