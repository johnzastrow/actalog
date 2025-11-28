<template>
  <v-app>
    <!-- Top App Bar (only show when authenticated) -->
    <v-app-bar
      v-if="showAppBar"
      color="#2c3e50"
      elevation="0"
      density="comfortable"
      app
    >
      <v-app-bar-title class="d-flex align-center">
        <!-- Logo -->
        <router-link to="/dashboard" class="d-flex align-center text-decoration-none">
          <img src="/logo.svg" alt="ActaLog Logo" style="height: 32px; width: 32px; margin-right: 8px;" />
          <span class="text-white font-weight-bold">ActaLog</span>
        </router-link>
      </v-app-bar-title>

      <template v-slot:append>
        <!-- Network Status Indicator -->
        <v-chip
          v-if="!networkStore.isOnline"
          size="small"
          color="warning"
          variant="flat"
          class="mr-2"
        >
          <v-icon start size="small">mdi-cloud-off-outline</v-icon>
          Offline
        </v-chip>
        <v-chip
          v-else-if="networkStore.isSyncing"
          size="small"
          color="info"
          variant="flat"
          class="mr-2"
        >
          <v-icon start size="small">mdi-sync</v-icon>
          Syncing...
        </v-chip>

        <!-- Current Date -->
        <div class="text-white text-caption mr-2" style="opacity: 0.9">
          {{ currentDate }}
        </div>
        <!-- Notifications icon -->
        <v-btn icon="mdi-bell-outline" color="white" variant="text" size="small"></v-btn>
      </template>
    </v-app-bar>

    <!-- Network Status Notifications -->
    <v-snackbar
      v-model="networkStore.showOfflineNotification"
      :timeout="-1"
      location="top"
      color="warning"
      elevation="8"
    >
      <div class="d-flex align-center">
        <v-icon start>mdi-cloud-off-outline</v-icon>
        <div>
          <strong>You're offline</strong>
          <div class="text-caption">Changes will be saved locally and synced when you're back online</div>
        </div>
      </div>
      <template v-slot:actions>
        <v-btn
          variant="text"
          size="small"
          @click="networkStore.dismissOfflineNotification()"
        >
          Dismiss
        </v-btn>
      </template>
    </v-snackbar>

    <v-snackbar
      v-model="networkStore.showOnlineNotification"
      :timeout="3000"
      location="top"
      color="success"
      elevation="8"
    >
      <div class="d-flex align-center">
        <v-icon start>mdi-cloud-check-outline</v-icon>
        <div>
          <strong>You're back online</strong>
          <div class="text-caption" v-if="networkStore.hasPendingSync">Syncing your changes...</div>
        </div>
      </div>
    </v-snackbar>

    <v-snackbar
      v-model="networkStore.showSyncNotification"
      :timeout="3000"
      location="top"
      color="success"
      elevation="8"
    >
      <div class="d-flex align-center">
        <v-icon start>mdi-check-circle-outline</v-icon>
        <div>
          <strong>All changes synced</strong>
          <div class="text-caption">Your data is up to date</div>
        </div>
      </div>
    </v-snackbar>

    <!-- Offline Save Notification -->
    <v-snackbar
      v-model="showOfflineSaveNotification"
      :timeout="4000"
      location="top"
      color="info"
      elevation="8"
    >
      <div class="d-flex align-center">
        <v-icon start>mdi-content-save-outline</v-icon>
        <div>
          <strong>Saved Offline</strong>
          <div class="text-caption">{{ offlineSaveMessage }}</div>
        </div>
      </div>
      <template v-slot:actions>
        <v-btn
          variant="text"
          size="small"
          @click="showOfflineSaveNotification = false"
        >
          OK
        </v-btn>
      </template>
    </v-snackbar>

    <v-main :style="mainStyle">
      <router-view />
    </v-main>

    <!-- Install Prompt -->
    <InstallPrompt v-if="authStore.isAuthenticated" />

    <!-- PWA Update Prompt -->
    <UpdatePrompt />

    <!-- Bottom Navigation (only show when authenticated) -->
    <v-bottom-navigation
      v-if="showBottomNav"
      v-model="activeTab"
      grow
      app
      elevation="8"
      bg-color="white"
      height="56"
      class="bottom-nav-compact"
    >
      <v-btn value="dashboard" to="/dashboard" size="x-small">
        <v-icon size="20">mdi-view-dashboard</v-icon>
        <span class="nav-label">Home</span>
      </v-btn>

      <v-btn value="performance" to="/performance" size="x-small">
        <v-icon size="20">mdi-chart-line</v-icon>
        <span class="nav-label">Stats</span>
      </v-btn>

      <!-- Center FAB Button -->
      <v-btn
        value="log"
        to="/dashboard?open=quick-log"
        size="x-small"
        class="fab-button"
      >
        <v-avatar color="teal" size="40">
          <v-icon color="white" size="24">mdi-plus</v-icon>
        </v-avatar>
      </v-btn>

      <v-btn value="workouts" to="/workouts" size="x-small">
        <v-icon size="20">mdi-dumbbell</v-icon>
        <span class="nav-label">Log</span>
      </v-btn>

      <v-btn value="profile" to="/profile" size="x-small">
        <!-- Show avatar if user has one, otherwise show default icon -->
        <v-avatar v-if="userAvatar" size="20">
          <v-img :src="userAvatar" alt="Profile" />
        </v-avatar>
        <v-icon v-else size="20" color="#597a6a">mdi-account-circle</v-icon>
        <span class="nav-label">Me</span>
      </v-btn>
    </v-bottom-navigation>
  </v-app>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { useTheme } from 'vuetify'
import { useAuthStore } from '@/stores/auth'
import { useNetworkStore } from '@/stores/network'
import InstallPrompt from '@/components/InstallPrompt.vue'
import UpdatePrompt from '@/components/UpdatePrompt.vue'

const route = useRoute()
const theme = useTheme()
const authStore = useAuthStore()
const networkStore = useNetworkStore()

const activeTab = ref('dashboard')
const currentDate = ref('')

// Offline save notification state
const showOfflineSaveNotification = ref(false)
const offlineSaveMessage = ref('')

// Reactive user avatar - watches auth store for changes
const userAvatar = computed(() => {
  const user = authStore.user
  if (user && user.profile_image) {
    return user.profile_image
  }
  return null
})

// Show app bar and bottom nav only when authenticated and not on login/register
const showAppBar = computed(() => {
  const publicRoutes = ['login', 'register', 'not-found']
  return authStore.isAuthenticated && !publicRoutes.includes(route.name)
})

const showBottomNav = computed(() => {
  const publicRoutes = ['login', 'register', 'not-found']
  return authStore.isAuthenticated && !publicRoutes.includes(route.name)
})

// Computed style for v-main to handle safe areas and bottom nav
const mainStyle = computed(() => {
  // Bottom nav height (56px) + safe area for devices with home indicator
  const bottomPadding = showBottomNav.value
    ? 'calc(56px + env(safe-area-inset-bottom, 0px))'
    : '0'
  return {
    paddingBottom: bottomPadding,
    minHeight: '100%',
    maxWidth: '100vw',
    overflowX: 'hidden'
  }
})

// Update current date
function updateCurrentDate() {
  const now = new Date()
  currentDate.value = now.toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
    year: 'numeric'
  })
}

// Handle offline save events
function handleOfflineSave(event) {
  offlineSaveMessage.value = event.detail?.message || 'Saved offline. Will sync when back online.'
  showOfflineSaveNotification.value = true
  // Update pending sync count
  networkStore.incrementPendingSync()
}

// Check for user's preferred theme
onMounted(() => {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme) {
    theme.global.name.value = savedTheme
  }

  updateCurrentDate()

  // Update date every minute
  setInterval(updateCurrentDate, 60000)

  // Initialize network status listeners
  networkStore.initNetworkListeners()

  // Listen for offline save events
  window.addEventListener('offline-save', handleOfflineSave)
})

onBeforeUnmount(() => {
  window.removeEventListener('offline-save', handleOfflineSave)
})
</script>

<style>
/* Global styles based on requirements */
* {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
    'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

body {
  margin: 0;
  padding: 0;
  background-color: #f5f7fa;
}

/* Typography - optimized for mobile */
h1 {
  font-size: 24px;
  font-weight: 600;
}

h2 {
  font-size: 20px;
  font-weight: 600;
}

h3 {
  font-size: 18px;
  font-weight: 500;
}

h4, h5, h6 {
  font-size: 16px;
  font-weight: 500;
}

p, span, div {
  font-weight: 400;
  font-size: 14px;
}

/* Smaller text variants */
.text-caption {
  font-size: 12px !important;
}

.text-body-2 {
  font-size: 13px !important;
}

.text-body-1 {
  font-size: 14px !important;
}

/* Spacing from requirements - 20px outer padding, 16px gutter */
.v-container {
  padding: 20px;
}

.v-row {
  margin: -8px;
}

.v-col {
  padding: 8px;
}

/* FAB button styling */
.fab-button .v-avatar {
  position: relative;
  top: -8px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

/* Bottom nav button styling */
.v-bottom-navigation .v-btn {
  flex-direction: column;
  height: 100% !important;
}

.v-bottom-navigation .v-btn .v-icon {
  margin-bottom: 2px;
}

/* Compact bottom navigation for mobile */
.bottom-nav-compact .v-btn {
  min-width: 50px !important;
  padding: 4px 2px !important;
}

.bottom-nav-compact .nav-label {
  font-size: 9px !important;
  line-height: 1.1;
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.bottom-nav-compact .fab-button .v-avatar {
  top: -4px;
}

/* Reduce overall app font sizes for mobile */
@media (max-width: 600px) {
  .v-card-title {
    font-size: 16px !important;
  }

  .v-card-subtitle {
    font-size: 12px !important;
  }

  .v-list-item-title {
    font-size: 14px !important;
  }

  .v-list-item-subtitle {
    font-size: 11px !important;
  }
}
</style>
