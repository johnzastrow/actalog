<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-toolbar-title class="text-white font-weight-bold">Dashboard</v-toolbar-title>
    </v-app-bar>

    <v-container class="pa-3" style="margin-top: 56px; margin-bottom: 70px">
      <!-- Quick Actions -->
      <v-card elevation="0" rounded="lg" class="mb-3 pa-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-2" style="color: #1a1a1a">Quick Actions</h2>
        <v-btn
          color="#00bcd4"
          size="large"
          prepend-icon="mdi-plus"
          to="/workouts/log"
          block
          rounded="lg"
          class="text-none font-weight-bold"
          style="background: #00bcd4; color: white"
        >
          Log Today's Workout
        </v-btn>
      </v-card>

      <!-- Stats Grid -->
      <v-row dense class="mb-3">
        <v-col cols="6">
          <v-card elevation="0" rounded="lg" class="pa-3 text-center" style="background: white">
            <v-progress-circular
              v-if="loading"
              indeterminate
              color="#00bcd4"
              size="32"
            />
            <div v-else class="text-h4 font-weight-bold" style="color: #00bcd4">
              {{ totalWorkouts }}
            </div>
            <div class="text-caption" style="color: #666">Total Workouts</div>
          </v-card>
        </v-col>
        <v-col cols="6">
          <v-card elevation="0" rounded="lg" class="pa-3 text-center" style="background: white">
            <v-progress-circular
              v-if="loading"
              indeterminate
              color="#00bcd4"
              size="32"
            />
            <div v-else class="text-h4 font-weight-bold" style="color: #00bcd4">
              {{ monthWorkouts }}
            </div>
            <div class="text-caption" style="color: #666">This Month</div>
          </v-card>
        </v-col>
      </v-row>

      <!-- Recent Workouts -->
      <v-card elevation="0" rounded="lg" class="pa-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-2" style="color: #1a1a1a">Recent Workouts</h2>

        <!-- Loading State -->
        <div v-if="loading" class="text-center pa-4">
          <v-progress-circular indeterminate color="#00bcd4" size="48" />
        </div>

        <!-- Empty State -->
        <div v-else-if="workouts.length === 0" class="text-center pa-4">
          <v-icon size="48" color="#ccc">mdi-dumbbell</v-icon>
          <p class="text-body-2 mt-2" style="color: #666">No workouts logged yet</p>
          <v-btn
            color="#00bcd4"
            size="small"
            to="/workouts/log"
            rounded="lg"
            class="mt-2 text-none font-weight-600"
            style="color: white"
          >
            Log Your First Workout
          </v-btn>
        </div>

        <!-- Recent Workouts List -->
        <div v-else>
          <v-list bg-color="transparent" density="compact">
            <v-list-item
              v-for="workout in workouts.slice(0, 5)"
              :key="workout.id"
              :to="`/workouts`"
              rounded="lg"
              class="mb-1"
            >
              <template #prepend>
                <v-icon color="#00bcd4" size="small">mdi-dumbbell</v-icon>
              </template>
              <v-list-item-title class="text-body-2 font-weight-medium" style="color: #1a1a1a">
                {{ formatDate(workout.workout_date) }}
                <span v-if="workout.workout_name" class="ml-1" style="color: #00bcd4">
                  - {{ workout.workout_name }}
                </span>
              </v-list-item-title>
              <v-list-item-subtitle class="text-caption" style="color: #666">
                <span v-if="workout.movements && workout.movements.length > 0">
                  {{ workout.movements.length }} movement(s)
                </span>
              </v-list-item-subtitle>
            </v-list-item>
          </v-list>

          <v-btn
            v-if="workouts.length > 5"
            variant="text"
            color="#00bcd4"
            size="small"
            to="/workouts"
            block
            class="mt-2 text-none"
          >
            View All Workouts
          </v-btn>
        </div>
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
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'

const activeTab = ref('dashboard')
const loading = ref(false)
const workouts = ref([])
const totalWorkouts = ref(0)
const monthWorkouts = ref(0)

// Fetch dashboard data
async function fetchDashboardData() {
  loading.value = true
  try {
    const response = await axios.get('/api/workouts')
    workouts.value = response.data.workouts || []
    totalWorkouts.value = workouts.value.length

    // Count workouts this month
    const now = new Date()
    const currentMonth = now.getMonth()
    const currentYear = now.getFullYear()

    monthWorkouts.value = workouts.value.filter((workout) => {
      const workoutDate = new Date(workout.workout_date)
      return (
        workoutDate.getMonth() === currentMonth &&
        workoutDate.getFullYear() === currentYear
      )
    }).length
  } catch (err) {
    console.error('Failed to fetch dashboard data:', err)
  } finally {
    loading.value = false
  }
}

// Format date
function formatDate(dateString) {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  })
}

// Load data on mount
onMounted(() => {
  fetchDashboardData()
})
</script>
