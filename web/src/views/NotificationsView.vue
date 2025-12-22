<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <!-- Header -->
      <div class="d-flex justify-space-between align-center mb-3">
        <h1 class="text-h6 font-weight-bold">Notifications</h1>
        <v-btn
          v-if="notifications.length > 0"
          color="teal"
          variant="text"
          size="small"
          @click="markAllAsRead"
          :loading="markingAllAsRead"
          style="text-transform: none"
        >
          Mark all as read
        </v-btn>
      </div>

      <!-- Filter Tabs -->
      <v-tabs
        v-model="currentTab"
        color="teal"
        class="mb-3"
      >
        <v-tab value="all">All</v-tab>
        <v-tab value="unread">Unread ({{ unreadCount }})</v-tab>
      </v-tabs>

      <!-- Success/Error Alerts -->
      <v-alert
        v-if="successMessage"
        type="success"
        closable
        @click:close="successMessage = ''"
        class="mb-3"
      >
        {{ successMessage }}
      </v-alert>

      <v-alert
        v-if="errorMessage"
        type="error"
        closable
        @click:close="errorMessage = ''"
        class="mb-3"
      >
        {{ errorMessage }}
      </v-alert>

      <!-- Loading State -->
      <div v-if="loading" class="text-center py-6">
        <v-progress-circular indeterminate color="teal" />
      </div>

      <!-- Empty State -->
      <div v-else-if="notifications.length === 0" class="text-center py-8">
        <v-icon size="64" color="surface-variant" class="mb-3">mdi-bell-outline</v-icon>
        <p class="text-body-1 text-grey">
          {{ currentTab === 'unread' ? 'No unread notifications' : 'No notifications yet' }}
        </p>
      </div>

      <!-- Notifications List -->
      <div v-else>
        <v-card
          v-for="notification in notifications"
          :key="notification.id"
          elevation="0"
          rounded="lg"
          class="mb-2 pa-3"
          :style="{
            background: notification.read_at ? '#f5f5f5' : 'white',
            border: notification.read_at ? 'none' : '1px solid #00bcd4'
          }"
        >
          <div class="d-flex justify-space-between align-start">
            <div class="flex-grow-1" @click="markAsRead(notification)">
              <!-- Notification Icon -->
              <v-icon
                :color="getNotificationColor(notification.type)"
                size="small"
                class="mr-2"
              >
                {{ getNotificationIcon(notification.type) }}
              </v-icon>

              <!-- Notification Title -->
              <span
                class="text-body-2"
                :class="notification.read_at ? 'text-grey' : 'font-weight-bold'"
              >
                {{ notification.title }}
              </span>

              <!-- Notification Message -->
              <div
                class="text-body-2 mt-1 mb-0"
                :class="notification.read_at ? 'text-grey-lighten-1' : 'text-grey-darken-1'"
              >
                <MarkdownRenderer :content="notification.message" />
              </div>

              <!-- Timestamp -->
              <p class="text-caption text-grey mt-1 mb-0">
                {{ formatTimestamp(notification.created_at) }}
              </p>

              <!-- Notification Likes -->
              <NotificationLikes
                :notification-id="notification.id"
                @liked="handleLiked(notification)"
                @unliked="handleUnliked(notification)"
              />
            </div>

            <!-- Action Menu -->
            <v-menu>
              <template v-slot:activator="{ props }">
                <v-btn
                  v-bind="props"
                  icon="mdi-dots-vertical"
                  variant="text"
                  size="small"
                  @click.stop
                />
              </template>
              <v-list>
                <v-list-item
                  v-if="!notification.read_at"
                  @click="markAsRead(notification)"
                >
                  <v-list-item-title>
                    <v-icon size="small" class="mr-2">mdi-check</v-icon>
                    Mark as read
                  </v-list-item-title>
                </v-list-item>
                <v-list-item @click="deleteNotification(notification.id)">
                  <v-list-item-title>
                    <v-icon size="small" class="mr-2">mdi-delete</v-icon>
                    Delete
                  </v-list-item-title>
                </v-list-item>
              </v-list>
            </v-menu>
          </div>
        </v-card>

        <!-- Load More Button -->
        <v-btn
          v-if="hasMore"
          block
          color="teal"
          variant="outlined"
          rounded="lg"
          class="mt-3"
          @click="loadMore"
          :loading="loadingMore"
          style="text-transform: none"
        >
          Load more
        </v-btn>
      </div>
    </v-container>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import NotificationLikes from '@/components/NotificationLikes.vue'

const router = useRouter()

// State
const currentTab = ref('all')
const notifications = ref([])
const unreadCount = ref(0)
const loading = ref(false)
const loadingMore = ref(false)
const markingAllAsRead = ref(false)
const successMessage = ref('')
const errorMessage = ref('')
const offset = ref(0)
const limit = 20
const hasMore = ref(true)

// Computed
const apiEndpoint = computed(() => {
  return currentTab.value === 'unread' ? '/api/notifications/unread' : '/api/notifications'
})

// Methods
const loadNotifications = async (reset = true) => {
  try {
    if (reset) {
      loading.value = true
      offset.value = 0
    } else {
      loadingMore.value = true
    }

    const response = await axios.get(apiEndpoint.value, {
      params: {
        limit,
        offset: reset ? 0 : offset.value
      }
    })

    if (reset) {
      notifications.value = response.data.notifications || []
    } else {
      notifications.value.push(...(response.data.notifications || []))
    }

    hasMore.value = (response.data.notifications || []).length === limit

    // Also fetch unread count
    const countResponse = await axios.get('/api/notifications/count')
    unreadCount.value = countResponse.data.count || 0

  } catch (error) {
    errorMessage.value = 'Failed to load notifications'
    console.error('Load notifications error:', error)
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

const loadMore = () => {
  offset.value += limit
  loadNotifications(false)
}

const markAsRead = async (notification) => {
  if (notification.read_at) return

  try {
    await axios.put(`/api/notifications/${notification.id}/read`)
    notification.read_at = new Date().toISOString()
    unreadCount.value = Math.max(0, unreadCount.value - 1)
    successMessage.value = 'Notification marked as read'
  } catch (error) {
    errorMessage.value = 'Failed to mark notification as read'
    console.error('Mark as read error:', error)
  }
}

const markAllAsRead = async () => {
  try {
    markingAllAsRead.value = true
    await axios.put('/api/notifications/read-all')

    // Update all notifications in the list
    notifications.value.forEach(n => {
      n.read_at = new Date().toISOString()
    })

    unreadCount.value = 0
    successMessage.value = 'All notifications marked as read'
  } catch (error) {
    errorMessage.value = 'Failed to mark all notifications as read'
    console.error('Mark all as read error:', error)
  } finally {
    markingAllAsRead.value = false
  }
}

const deleteNotification = async (id) => {
  try {
    await axios.delete(`/api/notifications/${id}`)

    // Remove from list
    const index = notifications.value.findIndex(n => n.id === id)
    if (index !== -1) {
      if (!notifications.value[index].read_at) {
        unreadCount.value = Math.max(0, unreadCount.value - 1)
      }
      notifications.value.splice(index, 1)
    }

    successMessage.value = 'Notification deleted'
  } catch (error) {
    errorMessage.value = 'Failed to delete notification'
    console.error('Delete notification error:', error)
  }
}

const getNotificationIcon = (type) => {
  switch (type) {
    case 'pr_achievement':
      return 'mdi-trophy'
    case 'weekly_streak':
      return 'mdi-fire'
    case 'wod_milestone':
      return 'mdi-star'
    default:
      return 'mdi-bell'
  }
}

const getNotificationColor = (type) => {
  switch (type) {
    case 'pr_achievement':
      return '#ffc107' // gold
    case 'weekly_streak':
      return '#ff5722' // orange
    case 'wod_milestone':
      return '#00bcd4' // cyan
    default:
      return '#666'
  }
}

const formatTimestamp = (timestamp) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now - date
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`

  return date.toLocaleDateString()
}

const handleLiked = (notification) => {
  // Optionally reload notifications to update unread status
  // (When someone likes a notification, it marks it unread for the recipient)
  console.log('Notification liked:', notification.id)
}

const handleUnliked = (notification) => {
  console.log('Notification unliked:', notification.id)
}

// Watchers
watch(currentTab, () => {
  loadNotifications()
})

// Lifecycle
onMounted(() => {
  loadNotifications()
})
</script>

<style scoped>
.mobile-view-wrapper {
  min-height: 100vh;
  background: #f5f7fa;
}
</style>
