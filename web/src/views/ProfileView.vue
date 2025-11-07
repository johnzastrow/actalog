<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4 mb-6">Profile</h1>
      </v-col>

      <v-col cols="12">
        <v-card elevation="2" rounded="lg">
          <v-card-text class="pa-6">
            <div class="text-center mb-6">
              <v-avatar size="100" color="primary">
                <v-icon size="50">mdi-account</v-icon>
              </v-avatar>
              <h2 class="text-h6 mt-4">{{ user?.name || 'User' }}</h2>
              <p class="text-body-2 text-medium-emphasis">{{ user?.email || 'email@example.com' }}</p>
            </div>

            <v-divider class="my-6" />

            <v-list>
              <v-list-item prepend-icon="mdi-cog" to="/settings">
                <v-list-item-title>Settings</v-list-item-title>
              </v-list-item>
              <v-list-item prepend-icon="mdi-logout" @click="handleLogout">
                <v-list-item-title>Logout</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-bottom-navigation v-model="activeTab" grow>
      <v-btn value="dashboard" to="/dashboard">
        <v-icon>mdi-view-dashboard</v-icon>
        <span>Dashboard</span>
      </v-btn>
      <v-btn value="performance" to="/performance">
        <v-icon>mdi-chart-line</v-icon>
        <span>Performance</span>
      </v-btn>
      <v-btn value="log" to="/workouts/log" color="gold">
        <v-icon size="large">mdi-plus</v-icon>
      </v-btn>
      <v-btn value="workouts" to="/workouts">
        <v-icon>mdi-dumbbell</v-icon>
        <span>Workouts</span>
      </v-btn>
      <v-btn value="profile" to="/profile">
        <v-icon>mdi-account</v-icon>
        <span>Profile</span>
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
