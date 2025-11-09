<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-toolbar-title class="text-white font-weight-bold">Profile</v-toolbar-title>
    </v-app-bar>

    <v-container class="pa-3" style="margin-top: 56px; margin-bottom: 70px">
      <!-- Profile Card -->
      <v-card elevation="0" rounded="lg" class="pa-4 mb-3" style="background: white">
        <div class="text-center mb-3">
          <v-avatar size="80" color="#00bcd4">
            <v-icon size="40" color="white">mdi-account</v-icon>
          </v-avatar>
          <h2 class="text-h6 mt-3 font-weight-bold" style="color: #1a1a1a">
            {{ user?.name || 'User' }}
          </h2>
          <p class="text-body-2" style="color: #666">{{ user?.email || 'email@example.com' }}</p>
        </div>
      </v-card>

      <!-- Menu Options -->
      <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
        <v-list bg-color="transparent" density="compact">
          <v-list-item prepend-icon="mdi-cog" to="/settings" rounded="lg">
            <v-list-item-title class="font-weight-medium" style="color: #1a1a1a">
              Settings
            </v-list-item-title>
          </v-list-item>
          <v-list-item prepend-icon="mdi-logout" @click="handleLogout" rounded="lg">
            <v-list-item-title class="font-weight-medium" style="color: #1a1a1a">
              Logout
            </v-list-item-title>
          </v-list-item>
        </v-list>
      </v-card>
    </v-container>

    <!-- Bottom Navigation -->
    <v-bottom-navigation
      v-model="activeTab"
      grow
      style="position: fixed; bottom: 0; background: white"
      elevation="8"
    >
      <v-btn value="dashboard" to="/dashboard">
        <v-icon>mdi-view-dashboard</v-icon>
        <span style="font-size: 10px">Dashboard</span>
      </v-btn>
      <v-btn value="performance" to="/performance">
        <v-icon>mdi-chart-line</v-icon>
        <span style="font-size: 10px">Performance</span>
      </v-btn>
      <v-btn value="log" to="/workouts/log" style="position: relative; bottom: 20px">
        <v-avatar color="#ffc107" size="56" style="box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2)">
          <v-icon color="white" size="32">mdi-plus</v-icon>
        </v-avatar>
      </v-btn>
      <v-btn value="workouts" to="/workouts">
        <v-icon>mdi-dumbbell</v-icon>
        <span style="font-size: 10px">Workouts</span>
      </v-btn>
      <v-btn value="profile" to="/profile">
        <v-icon>mdi-account</v-icon>
        <span style="font-size: 10px">Profile</span>
      </v-btn>
    </v-bottom-navigation>
  </v-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const activeTab = ref('profile')

const user = computed(() => authStore.user)

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}
</script>
