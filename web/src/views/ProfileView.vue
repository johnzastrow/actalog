<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-toolbar-title class="text-white font-weight-bold">Profile</v-toolbar-title>
      <v-spacer />
      <v-btn icon="mdi-cog" color="white" @click="$router.push('/settings')" />
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
          <v-chip
            v-if="user?.role === 'admin'"
            size="small"
            color="#e91e63"
            class="mt-2"
          >
            <v-icon start size="x-small">mdi-shield-crown</v-icon>
            Admin
          </v-chip>
        </div>

        <!-- Member Since -->
        <div v-if="user?.created_at" class="text-center text-caption" style="color: #999">
          Member since {{ formatMemberSince(user.created_at) }}
        </div>
      </v-card>

      <!-- Stats Summary -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Workout Summary</h2>

        <!-- Loading State -->
        <div v-if="loadingStats" class="text-center py-4">
          <v-progress-circular indeterminate color="#00bcd4" size="32" />
        </div>

        <!-- Stats Grid -->
        <v-row v-else dense>
          <v-col cols="6">
            <v-card elevation="0" rounded="lg" class="pa-2 text-center" style="background: #f5f7fa">
              <div class="text-h5 font-weight-bold" style="color: #00bcd4">
                {{ stats.totalWorkouts }}
              </div>
              <div class="text-caption" style="color: #666">Total Workouts</div>
            </v-card>
          </v-col>
          <v-col cols="6">
            <v-card elevation="0" rounded="lg" class="pa-2 text-center" style="background: #f5f7fa">
              <div class="text-h5 font-weight-bold" style="color: #4caf50">
                {{ stats.currentStreak }}
              </div>
              <div class="text-caption" style="color: #666">Day Streak</div>
            </v-card>
          </v-col>
          <v-col cols="6">
            <v-card elevation="0" rounded="lg" class="pa-2 text-center" style="background: #f5f7fa">
              <div class="text-h5 font-weight-bold" style="color: #ffc107">
                {{ stats.personalRecords }}
              </div>
              <div class="text-caption" style="color: #666">Personal Records</div>
            </v-card>
          </v-col>
          <v-col cols="6">
            <v-card elevation="0" rounded="lg" class="pa-2 text-center" style="background: #f5f7fa">
              <div class="text-h5 font-weight-bold" style="color: #e91e63">
                {{ stats.customTemplates }}
              </div>
              <div class="text-caption" style="color: #666">Custom Templates</div>
            </v-card>
          </v-col>
        </v-row>
      </v-card>

      <!-- Quick Actions -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-2" style="color: #1a1a1a">Quick Actions</h2>
        <v-list bg-color="transparent" density="compact">
          <v-list-item
            prepend-icon="mdi-dumbbell"
            @click="$router.push('/workouts')"
            rounded="lg"
            style="cursor: pointer"
          >
            <v-list-item-title class="font-weight-medium" style="color: #1a1a1a">
              My Templates
            </v-list-item-title>
            <template #append>
              <v-icon color="#ccc" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-fire"
            @click="$router.push('/wods')"
            rounded="lg"
            style="cursor: pointer"
          >
            <v-list-item-title class="font-weight-medium" style="color: #1a1a1a">
              Benchmark WODs
            </v-list-item-title>
            <template #append>
              <v-icon color="#ccc" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-trophy"
            @click="$router.push('/prs')"
            rounded="lg"
            style="cursor: pointer"
          >
            <v-list-item-title class="font-weight-medium" style="color: #1a1a1a">
              Personal Records
            </v-list-item-title>
            <template #append>
              <v-icon color="#ccc" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>
        </v-list>
      </v-card>

      <!-- Account Actions -->
      <v-card elevation="0" rounded="lg" class="pa-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-2" style="color: #1a1a1a">Account</h2>
        <v-list bg-color="transparent" density="compact">
          <v-list-item
            prepend-icon="mdi-cog"
            @click="$router.push('/settings')"
            rounded="lg"
            style="cursor: pointer"
          >
            <v-list-item-title class="font-weight-medium" style="color: #1a1a1a">
              Settings
            </v-list-item-title>
            <template #append>
              <v-icon color="#ccc" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-logout"
            @click="handleLogout"
            rounded="lg"
            style="cursor: pointer"
          >
            <v-list-item-title class="font-weight-medium" style="color: #e91e63">
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
        <span style="font-size: 10px">Templates</span>
      </v-btn>
      <v-btn value="profile" to="/profile">
        <v-icon>mdi-account</v-icon>
        <span style="font-size: 10px">Profile</span>
      </v-btn>
    </v-bottom-navigation>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const router = useRouter()
const authStore = useAuthStore()
const activeTab = ref('profile')

const user = computed(() => authStore.user)
const loadingStats = ref(false)

// Stats
const stats = ref({
  totalWorkouts: 0,
  currentStreak: 0,
  personalRecords: 0,
  customTemplates: 0
})

// Fetch user statistics
async function fetchStats() {
  loadingStats.value = true
  try {
    // Fetch workouts for stats
    const [workoutsRes, prsRes, templatesRes] = await Promise.all([
      axios.get('/api/user-workouts').catch(() => ({ data: { user_workouts: [] } })),
      axios.get('/api/movements/personal-records').catch(() => ({ data: { personal_records: [] } })),
      axios.get('/api/workouts/my-templates').catch(() => ({ data: { workouts: [] } }))
    ])

    const userWorkouts = workoutsRes.data.user_workouts || []

    // Calculate stats
    stats.value = {
      totalWorkouts: userWorkouts.length,
      currentStreak: calculateStreak(userWorkouts),
      personalRecords: (prsRes.data.personal_records || []).length,
      customTemplates: (templatesRes.data.workouts || []).length
    }
  } catch (err) {
    console.error('Failed to fetch stats:', err)
  } finally {
    loadingStats.value = false
  }
}

// Calculate current streak
function calculateStreak(workouts) {
  if (workouts.length === 0) return 0

  const sortedWorkouts = [...workouts].sort((a, b) =>
    new Date(b.workout_date) - new Date(a.workout_date)
  )

  let streak = 0
  let currentDate = new Date()
  currentDate.setHours(0, 0, 0, 0)

  for (const workout of sortedWorkouts) {
    const workoutDate = new Date(workout.workout_date)
    workoutDate.setHours(0, 0, 0, 0)

    const diffDays = Math.floor((currentDate - workoutDate) / (1000 * 60 * 60 * 24))

    if (diffDays === streak) {
      streak++
      currentDate.setDate(currentDate.getDate() - 1)
    } else if (diffDays > streak) {
      break
    }
  }

  return streak
}

// Format member since date
function formatMemberSince(dateString) {
  const date = new Date(dateString)
  const options = { month: 'long', year: 'numeric' }
  return date.toLocaleDateString('en-US', options)
}

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

onMounted(() => {
  fetchStats()
})
</script>
