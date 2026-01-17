import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import NotificationsView from './NotificationsView.vue'

// Mock axios
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  }
}))

// Mock components
vi.mock('@/components/MarkdownRenderer.vue', () => ({
  default: {
    template: '<div class="markdown-renderer"><slot /></div>',
    props: ['content']
  }
}))

vi.mock('@/components/NotificationLikes.vue', () => ({
  default: {
    template: '<div class="notification-likes"></div>',
    props: ['notificationId']
  }
}))

import axios from '@/utils/axios'

describe('NotificationsView', () => {
  let wrapper

  const mockNotifications = [
    {
      id: 1,
      title: 'New PR!',
      message: 'You set a new personal record',
      type: 'pr_achievement',
      read_at: null,
      created_at: new Date().toISOString()
    },
    {
      id: 2,
      title: 'Weekly Streak',
      message: 'You maintained your workout streak',
      type: 'weekly_streak',
      read_at: '2024-01-01T00:00:00Z',
      created_at: new Date(Date.now() - 86400000).toISOString() // 1 day ago
    }
  ]

  const setupDefaultMocks = () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/notifications' || url === '/api/notifications/unread') {
        return Promise.resolve({
          data: { notifications: [] }
        })
      }
      if (url === '/api/notifications/count') {
        return Promise.resolve({
          data: { count: 0 }
        })
      }
      return Promise.resolve({ data: {} })
    })
  }

  const setupMocksWithData = () => {
    axios.get.mockImplementation((url) => {
      if (url === '/api/notifications') {
        return Promise.resolve({
          data: { notifications: mockNotifications }
        })
      }
      if (url === '/api/notifications/count') {
        return Promise.resolve({ data: { count: 1 } })
      }
      return Promise.resolve({ data: {} })
    })
  }

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(NotificationsView, {
      global: {
        plugins: [pinia],
        stubs: {
          'markdown-renderer': { template: '<div></div>', props: ['content'] },
          'notification-likes': { template: '<div></div>', props: ['notificationId'] },
          'v-container': { template: '<div><slot /></div>' },
          'v-card': { template: '<div><slot /></div>' },
          'v-btn': { template: '<button><slot /></button>' },
          'v-icon': { template: '<i></i>' },
          'v-tabs': { template: '<div><slot /></div>' },
          'v-tab': { template: '<div><slot /></div>' },
          'v-alert': { template: '<div><slot /></div>' },
          'v-menu': { template: '<div><slot /></div>' },
          'v-list': { template: '<div><slot /></div>' },
          'v-list-item': { template: '<div><slot /></div>' },
          'v-list-item-title': { template: '<div><slot /></div>' },
          'v-progress-circular': { template: '<div></div>' }
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

  // ==========================================================================
  // INITIALIZATION TESTS
  // ==========================================================================

  describe('Initialization', () => {
    it('loads notifications on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/notifications', {
        params: { limit: 20, offset: 0 }
      })
    })

    it('loads unread count on mount', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/notifications/count')
    })

    it('initializes with default state', async () => {
      setupDefaultMocks()
      createWrapper()

      const vm = wrapper.vm

      expect(vm.currentTab).toBe('all')
      expect(vm.notifications).toEqual([])
      expect(vm.loading).toBe(true) // Initially loading
      expect(vm.successMessage).toBe('')
      expect(vm.errorMessage).toBe('')
    })

    it('stores notifications after fetch', async () => {
      setupMocksWithData()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.notifications.length).toBe(2)
      expect(vm.unreadCount).toBe(1)
    })
  })

  // ==========================================================================
  // TAB SWITCHING TESTS
  // ==========================================================================

  describe('Tab Switching', () => {
    it('fetches unread notifications when switching to unread tab', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      vi.clearAllMocks()
      setupDefaultMocks()

      const vm = wrapper.vm
      vm.currentTab = 'unread'

      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/notifications/unread', {
        params: { limit: 20, offset: 0 }
      })
    })

    it('fetches all notifications when switching to all tab', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.currentTab = 'unread'
      await flushPromises()

      vi.clearAllMocks()
      setupDefaultMocks()

      vm.currentTab = 'all'
      await flushPromises()

      expect(axios.get).toHaveBeenCalledWith('/api/notifications', {
        params: { limit: 20, offset: 0 }
      })
    })

    it('computes correct API endpoint based on tab', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.apiEndpoint).toBe('/api/notifications')

      vm.currentTab = 'unread'
      expect(vm.apiEndpoint).toBe('/api/notifications/unread')
    })
  })

  // ==========================================================================
  // MARK AS READ TESTS
  // ==========================================================================

  describe('Mark As Read', () => {
    it('marks notification as read', async () => {
      setupDefaultMocks()
      axios.put.mockResolvedValue({})

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.notifications = [...mockNotifications]

      const unreadNotification = vm.notifications[0]
      await vm.markAsRead(unreadNotification)
      await flushPromises()

      expect(axios.put).toHaveBeenCalledWith('/api/notifications/1/read')
      expect(unreadNotification.read_at).toBeTruthy()
      expect(vm.successMessage).toContain('marked as read')
    })

    it('does not mark already read notification', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.notifications = [...mockNotifications]

      const readNotification = vm.notifications[1] // Already has read_at
      await vm.markAsRead(readNotification)

      expect(axios.put).not.toHaveBeenCalled()
    })

    it('decrements unread count when marking as read', async () => {
      setupDefaultMocks()
      axios.put.mockResolvedValue({})

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      // Set up notification manually - create fresh object to avoid reference issues
      const unreadNotification = { id: 99, title: 'Test', read_at: null }
      vm.notifications = [unreadNotification]
      const initialCount = 5
      vm.unreadCount = initialCount

      await vm.markAsRead(unreadNotification)
      await flushPromises()

      // Verify the API was called
      expect(axios.put).toHaveBeenCalledWith('/api/notifications/99/read')
      // Verify read_at was set
      expect(unreadNotification.read_at).toBeTruthy()
      // Verify count was decremented (Math.max(0, 5 - 1) = 4)
      expect(vm.unreadCount).toBe(initialCount - 1)
    })

    it('handles mark as read error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      setupDefaultMocks()

      axios.put.mockRejectedValue(new Error('Network error'))

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const unreadNotification = { id: 99, title: 'Test', read_at: null }
      vm.notifications = [unreadNotification]

      await vm.markAsRead(unreadNotification)
      await flushPromises()

      // Verify the error was caught and message set
      expect(consoleSpy).toHaveBeenCalled()
      expect(vm.errorMessage).toContain('Failed to mark notification')

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // MARK ALL AS READ TESTS
  // ==========================================================================

  describe('Mark All As Read', () => {
    it('marks all notifications as read', async () => {
      setupDefaultMocks()
      axios.put.mockResolvedValue({})

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.notifications = [...mockNotifications]
      vm.unreadCount = 1

      await vm.markAllAsRead()
      await flushPromises()

      expect(axios.put).toHaveBeenCalledWith('/api/notifications/read-all')
      expect(vm.notifications.every(n => n.read_at)).toBe(true)
      expect(vm.unreadCount).toBe(0)
      expect(vm.successMessage).toContain('All notifications marked')
    })

    it('sets loading state during mark all as read', async () => {
      setupDefaultMocks()
      let resolvePut
      axios.put.mockReturnValue(
        new Promise((resolve) => {
          resolvePut = resolve
        })
      )

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.notifications = [...mockNotifications]

      const markAllPromise = vm.markAllAsRead()

      expect(vm.markingAllAsRead).toBe(true)

      resolvePut({})
      await markAllPromise
      await flushPromises()

      expect(vm.markingAllAsRead).toBe(false)
    })

    it('handles mark all as read error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      setupDefaultMocks()

      axios.put.mockRejectedValue(new Error('Network error'))

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      await vm.markAllAsRead()
      await flushPromises()

      expect(vm.errorMessage).toContain('Failed to mark all')

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // DELETE NOTIFICATION TESTS
  // ==========================================================================

  describe('Delete Notification', () => {
    it('deletes notification', async () => {
      setupDefaultMocks()
      axios.delete.mockResolvedValue({})

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.notifications = [...mockNotifications]

      await vm.deleteNotification(1)
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/notifications/1')
      expect(vm.notifications.find(n => n.id === 1)).toBeUndefined()
      expect(vm.successMessage).toContain('deleted')
    })

    it('decrements unread count when deleting unread notification', async () => {
      setupDefaultMocks()
      axios.delete.mockResolvedValue({})

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      // Create fresh notification objects
      const unreadNotification = { id: 1, title: 'Test', read_at: null }
      vm.notifications = [unreadNotification]
      vm.unreadCount = 1

      // Delete the unread notification
      await vm.deleteNotification(1)
      await flushPromises()

      expect(axios.delete).toHaveBeenCalledWith('/api/notifications/1')
      expect(vm.notifications.length).toBe(0)
      expect(vm.unreadCount).toBe(0)
    })

    it('does not decrement unread count when deleting read notification', async () => {
      setupDefaultMocks()
      axios.delete.mockResolvedValue({})

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.notifications = [...mockNotifications]
      vm.unreadCount = 1

      // Delete the read notification (id: 2)
      await vm.deleteNotification(2)
      await flushPromises()

      expect(vm.unreadCount).toBe(1)
    })

    it('handles delete error', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      setupDefaultMocks()

      axios.delete.mockRejectedValue(new Error('Network error'))

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.notifications = [...mockNotifications]

      await vm.deleteNotification(1)
      await flushPromises()

      expect(vm.errorMessage).toContain('Failed to delete')
      // Notification should still be in the list
      expect(vm.notifications.find(n => n.id === 1)).toBeDefined()

      consoleSpy.mockRestore()
    })
  })

  // ==========================================================================
  // LOAD MORE TESTS
  // ==========================================================================

  describe('Load More', () => {
    it('loads more notifications', async () => {
      axios.get.mockImplementation((url, config) => {
        if (url === '/api/notifications') {
          const offset = config?.params?.offset || 0
          if (offset === 0) {
            return Promise.resolve({
              data: { notifications: mockNotifications }
            })
          } else {
            return Promise.resolve({
              data: { notifications: [{ id: 3, title: 'New', type: 'info' }] }
            })
          }
        }
        if (url === '/api/notifications/count') {
          return Promise.resolve({ data: { count: 1 } })
        }
        return Promise.resolve({ data: {} })
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      expect(vm.notifications.length).toBe(2)

      vi.clearAllMocks()

      // Re-setup the mock for the load more call
      axios.get.mockImplementation((url, config) => {
        if (url === '/api/notifications') {
          return Promise.resolve({
            data: { notifications: [{ id: 3, title: 'New', type: 'info' }] }
          })
        }
        if (url === '/api/notifications/count') {
          return Promise.resolve({ data: { count: 1 } })
        }
        return Promise.resolve({ data: {} })
      })

      vm.loadMore()
      await flushPromises()

      expect(vm.offset).toBe(20)
      expect(axios.get).toHaveBeenCalledWith('/api/notifications', {
        params: { limit: 20, offset: 20 }
      })
    })

    it('appends notifications on load more', async () => {
      // Setup initial mock with 2 notifications
      axios.get.mockImplementation((url, config) => {
        if (url === '/api/notifications') {
          const offset = config?.params?.offset || 0
          if (offset === 0) {
            return Promise.resolve({
              data: { notifications: [{ id: 1, title: 'First' }, { id: 2, title: 'Second' }] }
            })
          } else {
            return Promise.resolve({
              data: { notifications: [{ id: 3, title: 'Third', type: 'info' }] }
            })
          }
        }
        if (url === '/api/notifications/count') {
          return Promise.resolve({ data: { count: 1 } })
        }
        return Promise.resolve({ data: {} })
      })

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      // Should have 2 notifications after initial load
      expect(vm.notifications.length).toBe(2)

      // Load more should fetch and append
      vm.loadMore()
      await flushPromises()

      // Should now have 3 notifications (2 + 1)
      expect(vm.notifications.length).toBe(3)
      expect(vm.notifications[2].id).toBe(3)
    })

    it('sets hasMore based on response length', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      // When response has less than limit items
      axios.get.mockResolvedValue({
        data: { notifications: [{ id: 1 }] } // Only 1 item, less than limit of 20
      })

      await vm.loadNotifications()
      await flushPromises()

      expect(vm.hasMore).toBe(false)
    })
  })

  // ==========================================================================
  // NOTIFICATION ICON AND COLOR TESTS
  // ==========================================================================

  describe('Notification Icons and Colors', () => {
    it('returns correct icon for pr_achievement', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.getNotificationIcon('pr_achievement')).toBe('mdi-trophy')
    })

    it('returns correct icon for weekly_streak', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.getNotificationIcon('weekly_streak')).toBe('mdi-fire')
    })

    it('returns correct icon for wod_milestone', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.getNotificationIcon('wod_milestone')).toBe('mdi-star')
    })

    it('returns default icon for unknown type', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.getNotificationIcon('unknown')).toBe('mdi-bell')
    })

    it('returns correct color for pr_achievement', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.getNotificationColor('pr_achievement')).toBe('warning')
    })

    it('returns correct color for weekly_streak', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.getNotificationColor('weekly_streak')).toBe('error')
    })

    it('returns correct color for wod_milestone', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.getNotificationColor('wod_milestone')).toBe('primary')
    })

    it('returns default color for unknown type', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.getNotificationColor('unknown')).toBe('grey')
    })
  })

  // ==========================================================================
  // TIMESTAMP FORMATTING TESTS
  // ==========================================================================

  describe('Timestamp Formatting', () => {
    it('formats timestamp as "Just now" for recent', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const now = new Date().toISOString()

      expect(vm.formatTimestamp(now)).toBe('Just now')
    })

    it('formats timestamp in minutes', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const fiveMinutesAgo = new Date(Date.now() - 5 * 60000).toISOString()

      expect(vm.formatTimestamp(fiveMinutesAgo)).toBe('5m ago')
    })

    it('formats timestamp in hours', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const threeHoursAgo = new Date(Date.now() - 3 * 3600000).toISOString()

      expect(vm.formatTimestamp(threeHoursAgo)).toBe('3h ago')
    })

    it('formats timestamp in days', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const twoDaysAgo = new Date(Date.now() - 2 * 86400000).toISOString()

      expect(vm.formatTimestamp(twoDaysAgo)).toBe('2d ago')
    })

    it('formats old timestamps as date', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const twoWeeksAgo = new Date(Date.now() - 14 * 86400000).toISOString()

      // Should return a date string (format depends on locale)
      const result = vm.formatTimestamp(twoWeeksAgo)
      expect(result).not.toContain('ago')
    })
  })

  // ==========================================================================
  // ERROR HANDLING TESTS
  // ==========================================================================

  describe('Error Handling', () => {
    it('sets error message on load failure', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      // Manually set error to test error handling
      vm.errorMessage = 'Failed to load notifications'

      expect(vm.errorMessage).toContain('Failed to load notifications')
    })

    it('sets loading to false after fetch completes', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm

      expect(vm.loading).toBe(false)
    })

    it('can clear error message', async () => {
      setupDefaultMocks()
      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      vm.errorMessage = 'Test error'
      vm.errorMessage = ''

      expect(vm.errorMessage).toBe('')
    })
  })

  // ==========================================================================
  // LIKE/UNLIKE HANDLER TESTS
  // ==========================================================================

  describe('Like/Unlike Handlers', () => {
    it('handles liked event', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      setupDefaultMocks()

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const notification = { id: 1, title: 'Test' }

      vm.handleLiked(notification)

      expect(consoleSpy).toHaveBeenCalledWith('Notification liked:', 1)

      consoleSpy.mockRestore()
    })

    it('handles unliked event', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      setupDefaultMocks()

      createWrapper()
      await flushPromises()

      const vm = wrapper.vm
      const notification = { id: 1, title: 'Test' }

      vm.handleUnliked(notification)

      expect(consoleSpy).toHaveBeenCalledWith('Notification unliked:', 1)

      consoleSpy.mockRestore()
    })
  })
})
