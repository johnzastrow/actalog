<template>
  <v-app>
    <!-- Top App Bar (only show when authenticated) -->
    <v-app-bar
      v-if="showAppBar"
      color="#2c3657"
      elevation="0"
      density="comfortable"
      app
    >
      <v-app-bar-title class="d-flex align-center">
        <!-- Logo -->
        <router-link to="/dashboard" class="d-flex align-center text-decoration-none">
          <v-avatar size="32" color="#00bcd4" class="mr-2">
            <span class="text-white font-weight-bold">A</span>
          </v-avatar>
          <span class="text-white font-weight-bold">ActaLog</span>
        </router-link>
      </v-app-bar-title>

      <template v-slot:append>
        <!-- Current Date -->
        <div class="text-white text-caption mr-2" style="opacity: 0.9">
          {{ currentDate }}
        </div>
        <!-- Notifications icon -->
        <v-btn icon="mdi-bell-outline" color="white" variant="text" size="small"></v-btn>
      </template>
    </v-app-bar>

    <v-main :style="{ paddingBottom: showBottomNav ? '70px' : '0' }">
      <router-view />
    </v-main>

    <!-- Bottom Navigation (only show when authenticated) -->
    <v-bottom-navigation
      v-if="showBottomNav"
      v-model="activeTab"
      grow
      app
      elevation="8"
      bg-color="white"
      height="70"
    >
      <v-btn value="dashboard" to="/dashboard" size="small">
        <v-icon size="24">mdi-view-dashboard</v-icon>
        <span class="text-caption">Dashboard</span>
      </v-btn>

      <v-btn value="performance" to="/performance" size="small">
        <v-icon size="24">mdi-chart-line</v-icon>
        <span class="text-caption">Performance</span>
      </v-btn>

      <!-- Center FAB Button -->
      <v-btn
        value="log"
        to="/workouts/log"
        size="small"
        class="fab-button"
      >
        <v-avatar color="#ffc107" size="48">
          <v-icon color="white" size="28">mdi-plus</v-icon>
        </v-avatar>
      </v-btn>

      <v-btn value="workouts" to="/workouts" size="small">
        <v-icon size="24">mdi-dumbbell</v-icon>
        <span class="text-caption">Workouts</span>
      </v-btn>

      <v-btn value="profile" to="/profile" size="small">
        <!-- Show avatar if user has one, otherwise show default icon -->
        <v-avatar v-if="userAvatar" size="24">
          <v-img :src="userAvatar" alt="Profile" />
        </v-avatar>
        <v-icon v-else size="24" color="#597a6a">mdi-account-circle</v-icon>
        <span class="text-caption">Profile</span>
      </v-btn>
    </v-bottom-navigation>
  </v-app>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useTheme } from 'vuetify'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const theme = useTheme()
const authStore = useAuthStore()

const activeTab = ref('dashboard')
const currentDate = ref('')
const userAvatar = ref(null)

// Show app bar and bottom nav only when authenticated and not on login/register
const showAppBar = computed(() => {
  const publicRoutes = ['login', 'register', 'not-found']
  return authStore.isAuthenticated && !publicRoutes.includes(route.name)
})

const showBottomNav = computed(() => {
  const publicRoutes = ['login', 'register', 'not-found']
  return authStore.isAuthenticated && !publicRoutes.includes(route.name)
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

// Check for user's preferred theme
onMounted(() => {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme) {
    theme.global.name.value = savedTheme
  }

  updateCurrentDate()

  // Update date every minute
  setInterval(updateCurrentDate, 60000)

  // Load user avatar if available
  const user = authStore.user
  if (user && user.profile_image) {
    userAvatar.value = user.profile_image
  }
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

/* Typography from requirements */
h1, h2, h3, h4, h5, h6 {
  font-weight: 500;
}

p, span, div {
  font-weight: 400;
  font-size: 16px;
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
</style>
