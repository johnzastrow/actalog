<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-toolbar-title class="text-white font-weight-bold">Dashboard</v-toolbar-title>
      <v-spacer />
      <v-btn icon="mdi-bell-outline" color="white" />
    </v-app-bar>

    <v-container class="pa-3" style="margin-top: 56px; margin-bottom: 70px">
      <!-- Email Verification Alert -->
      <v-alert
        v-if="authStore.user && !authStore.user.email_verified"
        type="warning"
        prominent
        closable
        class="mb-3"
      >
        <v-row align="center">
          <v-col class="grow">
            <div class="text-h6">Email Verification Required</div>
            <div class="text-body-2">
              Please verify your email address to access all features. Check your inbox for the verification link.
            </div>
          </v-col>
          <v-col class="shrink">
            <v-btn
              color="warning"
              variant="elevated"
              @click="$router.push('/resend-verification')"
            >
              Resend Email
            </v-btn>
          </v-col>
        </v-row>
      </v-alert>

      <!-- Stats Cards -->
      <v-row dense class="mb-3">
        <v-col cols="6">
          <v-card elevation="0" rounded="lg" class="pa-3" style="background: white">
            <div class="d-flex align-center">
              <v-icon color="#00bcd4" size="32" class="mr-2">mdi-dumbbell</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" style="color: #1a1a1a">
                  {{ totalWorkouts }}
                </div>
                <div class="text-caption" style="color: #666">Total Workouts</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="6">
          <v-card elevation="0" rounded="lg" class="pa-3" style="background: white">
            <div class="d-flex align-center">
              <v-icon color="#ffc107" size="32" class="mr-2">mdi-calendar-month</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" style="color: #1a1a1a">
                  {{ monthWorkouts }}
                </div>
                <div class="text-caption" style="color: #666">This Month</div>
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>

      <v-row dense class="mb-3">
        <v-col cols="6">
          <v-card elevation="0" rounded="lg" class="pa-3" style="background: white">
            <div class="d-flex align-center">
              <v-icon color="#4caf50" size="32" class="mr-2">mdi-fire</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" style="color: #1a1a1a">
                  {{ currentStreak }}
                </div>
                <div class="text-caption" style="color: #666">Day Streak</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="6">
          <v-card elevation="0" rounded="lg" class="pa-3" style="background: white">
            <div class="d-flex align-center">
              <v-icon color="#e91e63" size="32" class="mr-2">mdi-clock-outline</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" style="color: #1a1a1a">
                  {{ avgTimePerWorkout }}m
                </div>
                <div class="text-caption" style="color: #666">Avg Time</div>
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>

      <!-- Quick Action -->
      <v-card
        elevation="0"
        rounded="lg"
        class="mb-3 pa-4 text-center"
        style="background: linear-gradient(135deg, #00bcd4 0%, #00acc1 100%); cursor: pointer"
        @click="$router.push('/workouts/log')"
      >
        <v-icon size="48" color="white" class="mb-2">mdi-plus-circle</v-icon>
        <div class="text-h6 font-weight-bold text-white">Log Today's Workout</div>
        <div class="text-caption text-white" style="opacity: 0.9">
          Track your progress and stay consistent
        </div>
      </v-card>

      <!-- Recent Workouts -->
      <div class="mb-3">
        <div class="d-flex align-center justify-space-between mb-2">
          <h3 class="text-h6 font-weight-bold" style="color: #1a1a1a">Recent Workouts</h3>
          <v-btn
            size="small"
            variant="text"
            color="#00bcd4"
            style="text-transform: none"
            @click="$router.push('/workouts')"
          >
            View All
          </v-btn>
        </div>

        <!-- Loading State -->
        <div v-if="loading" class="text-center py-8">
          <v-progress-circular indeterminate color="#00bcd4" size="48" />
          <p class="mt-2 text-caption" style="color: #666">Loading workouts...</p>
        </div>

        <!-- Empty State -->
        <v-card
          v-else-if="!loading && recentWorkouts.length === 0"
          elevation="0"
          rounded="lg"
          class="pa-6 text-center"
          style="background: white"
        >
          <v-icon size="64" color="#ccc">mdi-clipboard-text-outline</v-icon>
          <p class="text-h6 mt-2" style="color: #2c3e50">No workouts logged yet</p>
          <p class="text-body-2 mb-3" style="color: #666">
            Start tracking your fitness journey today!
          </p>
          <v-btn
            color="#00bcd4"
            rounded="lg"
            style="text-transform: none; font-weight: 600"
            @click="$router.push('/workouts/log')"
          >
            <v-icon start>mdi-plus</v-icon>
            Log Your First Workout
          </v-btn>
        </v-card>

        <!-- Recent Workouts List -->
        <div v-else>
          <v-card
            v-for="workout in recentWorkouts"
            :key="workout.id"
            elevation="0"
            rounded="lg"
            class="mb-2 pa-3"
            style="background: white; cursor: pointer"
            @click="viewWorkout(workout.id)"
          >
            <div class="d-flex align-center mb-1">
              <v-icon color="#00bcd4" class="mr-2" size="small">mdi-dumbbell</v-icon>
              <div class="flex-grow-1">
                <div class="font-weight-bold text-body-1" style="color: #1a1a1a">
                  {{ workout.workout_name || 'Workout' }}
                </div>
                <div class="text-caption" style="color: #666">
                  {{ formatDate(workout.workout_date) }}
                  <span v-if="workout.workout_type"> • {{ formatWorkoutType(workout.workout_type) }}</span>
                  <span v-if="workout.total_time"> • {{ formatTime(workout.total_time) }}</span>
                </div>
              </div>
              <v-icon color="#ccc" size="small">mdi-chevron-right</v-icon>
            </div>

            <!-- Display movements count -->
            <div v-if="workout.movements && workout.movements.length > 0" class="ml-7">
              <v-chip size="x-small" color="#e0e0e0" class="mr-1">
                <v-icon start size="x-small">mdi-weight-lifter</v-icon>
                {{ workout.movements.length }} movement(s)
              </v-chip>
            </div>

            <!-- Display WODs count -->
            <div v-if="workout.wods && workout.wods.length > 0" class="ml-7 mt-1">
              <v-chip size="x-small" color="#ffc107" class="mr-1">
                <v-icon start size="x-small">mdi-fire</v-icon>
                {{ workout.wods.length }} WOD(s)
              </v-chip>
            </div>

            <!-- Notes -->
            <div v-if="workout.notes" class="ml-7 mt-2 text-caption" style="color: #666">
              📝 {{ truncateText(workout.notes, 80) }}
            </div>
          </v-card>
        </div>
      </div>
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
import axios from '@/utils/axios'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const activeTab = ref('dashboard')

const loading = ref(false)
const userWorkouts = ref([])

// Computed stats
const totalWorkouts = computed(() => userWorkouts.value.length)

const monthWorkouts = computed(() => {
  const now = new Date()
  const firstDayOfMonth = new Date(now.getFullYear(), now.getMonth(), 1)

  return userWorkouts.value.filter(w => {
    const workoutDate = new Date(w.workout_date)
    return workoutDate >= firstDayOfMonth
  }).length
})

const currentStreak = computed(() => {
  if (userWorkouts.value.length === 0) return 0

  // Sort workouts by date (newest first)
  const sortedWorkouts = [...userWorkouts.value].sort((a, b) =>
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
})

const avgTimePerWorkout = computed(() => {
  const workoutsWithTime = userWorkouts.value.filter(w => w.total_time)
  if (workoutsWithTime.length === 0) return 0

  const totalMinutes = workoutsWithTime.reduce((sum, w) => sum + (w.total_time / 60), 0)
  return Math.round(totalMinutes / workoutsWithTime.length)
})

const recentWorkouts = computed(() => {
  return [...userWorkouts.value]
    .sort((a, b) => new Date(b.workout_date) - new Date(a.workout_date))
    .slice(0, 5) // Show 5 most recent
})

// Fetch user's logged workouts
async function fetchUserWorkouts() {
  loading.value = true
  try {
    const response = await axios.get('/api/user-workouts')
    userWorkouts.value = response.data.user_workouts || []
    console.log('Fetched user workouts:', userWorkouts.value.length)
  } catch (err) {
    console.error('Failed to fetch user workouts:', err)
    userWorkouts.value = []
  } finally {
    loading.value = false
  }
}

// Format date for display
function formatDate(dateString) {
  const date = new Date(dateString)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  // Reset time parts for comparison
  const dateOnly = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  const todayOnly = new Date(today.getFullYear(), today.getMonth(), today.getDate())
  const yesterdayOnly = new Date(yesterday.getFullYear(), yesterday.getMonth(), yesterday.getDate())

  if (dateOnly.getTime() === todayOnly.getTime()) {
    return 'Today'
  } else if (dateOnly.getTime() === yesterdayOnly.getTime()) {
    return 'Yesterday'
  } else {
    const options = { weekday: 'short', month: 'short', day: 'numeric' }
    return date.toLocaleDateString('en-US', options)
  }
}

// Format workout type
function formatWorkoutType(type) {
  if (!type) return ''
  return type.charAt(0).toUpperCase() + type.slice(1)
}

// Format time (seconds to readable format)
function formatTime(seconds) {
  if (!seconds) return ''

  if (seconds < 60) {
    return `${seconds}s`
  } else if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60)
    return `${minutes}min`
  } else {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    return `${hours}h ${minutes}m`
  }
}

// Truncate text
function truncateText(text, maxLength) {
  if (!text || text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

// View workout details
function viewWorkout(workoutId) {
  console.log('View workout details:', workoutId)
  // TODO: Navigate to workout detail page
  // router.push(`/workouts/${workoutId}`)
}

// Load data on mount
onMounted(() => {
  fetchUserWorkouts()
})
</script>

<style scoped>
/* Dashboard specific styles */
</style>
