import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminMetricsDashboardView from './AdminMetricsDashboardView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
  }
}))

// Mock components
vi.mock('@/components/admin/StatCard.vue', () => ({
  default: {
    template: '<div class="stat-card"></div>',
    props: ['icon', 'color', 'value', 'label', 'subtitle', 'format']
  }
}))

vi.mock('@/components/admin/WorkoutTrendsChart.vue', () => ({
  default: {
    template: '<div class="trends-chart"></div>',
    props: ['trends', 'loading']
  }
}))

import axios from '@/utils/axios'

describe('AdminMetricsDashboardView', () => {
  let wrapper

  const mockMetrics = {
    user_stats: {
      total_users: 100,
      active_this_month: 50,
      new_this_month: 10,
      disabled_users: 5
    },
    workout_stats: {
      total_workouts: 500,
      workouts_this_month: 100,
      avg_workouts_per_user: 5.5
    },
    content_stats: {
      total_movements: 200,
      user_created_movements: 50,
      total_wods: 100,
      user_created_wods: 25,
      total_workout_templates: 75,
      user_created_templates: 20
    },
    subscription_stats: {
      active_subscriptions: 80,
      expiring_soon: 5,
      expired_subscriptions: 10
    },
    system_health: {
      recent_audit_events_24h: 150,
      email_success_rate: 98.5,
      total_emails_sent: 500,
      failed_emails: 8
    },
    workout_trends: [
      { date: '2024-01-01', count: 10 },
      { date: '2024-01-02', count: 15 }
    ]
  }

  const setupDefaultMocks = () => {
    axios.get.mockResolvedValue({ data: mockMetrics })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(AdminMetricsDashboardView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-btn': { template: '<button @click="$attrs.onClick"><slot /></button>' },
          'v-icon': { template: '<i></i>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-progress-circular': { template: '<div></div>' },
          'v-bottom-navigation': { template: '<div><slot /></div>' },
          'stat-card': true,
          'workout-trends-chart': true
        }
      }
    })

    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
      wrapper = null
    }
  })

  describe('Initialization', () => {
    it('fetches metrics on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/admin/metrics')
    })

    it('stores fetched metrics', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.metrics).not.toBe(null)
      expect(vm.metrics.user_stats.total_users).toBe(100)
    })

    it('sets loading to false after fetch', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)
    })

    it('sets lastUpdated after fetch', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.lastUpdated).not.toBe('')
    })
  })

  describe('Metrics Data', () => {
    it('has user statistics', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.metrics.user_stats.total_users).toBe(100)
      expect(vm.metrics.user_stats.active_this_month).toBe(50)
      expect(vm.metrics.user_stats.new_this_month).toBe(10)
      expect(vm.metrics.user_stats.disabled_users).toBe(5)
    })

    it('has workout statistics', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.metrics.workout_stats.total_workouts).toBe(500)
      expect(vm.metrics.workout_stats.workouts_this_month).toBe(100)
      expect(vm.metrics.workout_stats.avg_workouts_per_user).toBe(5.5)
    })

    it('has content statistics', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.metrics.content_stats.total_movements).toBe(200)
      expect(vm.metrics.content_stats.total_wods).toBe(100)
    })

    it('has subscription statistics', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.metrics.subscription_stats.active_subscriptions).toBe(80)
      expect(vm.metrics.subscription_stats.expiring_soon).toBe(5)
    })

    it('has system health stats', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.metrics.system_health.email_success_rate).toBe(98.5)
      expect(vm.metrics.system_health.failed_emails).toBe(8)
    })
  })

  describe('Refresh', () => {
    it('can refresh metrics', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      axios.get.mockClear()

      await vm.fetchMetrics()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/admin/metrics')
    })
  })

  describe('Error Handling', () => {
    it('handles fetch error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue({
        response: { data: { error: 'Failed to load metrics' } }
      })
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toBe('Failed to load metrics')
      expect(vm.loading).toBe(false)

      consoleSpy.mockRestore()
    })

    it('handles generic fetch error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      axios.get.mockRejectedValue(new Error('Network error'))
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.error).toContain('Failed to load metrics')

      consoleSpy.mockRestore()
    })
  })
})
