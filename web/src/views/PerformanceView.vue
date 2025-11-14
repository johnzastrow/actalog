<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" density="compact" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-toolbar-title class="text-white font-weight-bold">Performance</v-toolbar-title>
      <v-spacer />
      <v-btn icon="mdi-refresh" color="white" size="small" @click="refreshData" />
    </v-app-bar>

    <v-container class="pa-2" style="margin-top: 36px; margin-bottom: 70px">
      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable @click:close="error = null" class="mb-2">
        {{ error }}
      </v-alert>

      <!-- Personal Records Summary -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <div class="d-flex align-center justify-space-between mb-2">
          <h2 class="text-h6 font-weight-bold" style="color: #1a1a1a">Personal Records</h2>
          <v-chip size="small" color="#4caf50">
            <v-icon start size="x-small">mdi-trophy</v-icon>
            {{ personalRecords.length }} PRs
          </v-chip>
        </div>

        <!-- Loading State -->
        <div v-if="loadingPRs" class="text-center py-4">
          <v-progress-circular indeterminate color="#00bcd4" size="32" />
          <p class="mt-2 text-caption" style="color: #666">Loading PRs...</p>
        </div>

        <!-- Empty State -->
        <div v-else-if="personalRecords.length === 0" class="text-center py-4">
          <v-icon size="48" color="#ccc">mdi-trophy-outline</v-icon>
          <p class="text-body-2 mt-2" style="color: #666">No personal records yet</p>
          <p class="text-caption" style="color: #999">
            Start logging workouts with weights to track your PRs
          </p>
        </div>

        <!-- PRs List -->
        <div v-else>
          <v-card
            v-for="pr in personalRecords.slice(0, 5)"
            :key="pr.movement_id"
            elevation="0"
            rounded="lg"
            class="mb-2 pa-2"
            style="background: #f5f7fa"
          >
            <div class="d-flex align-center">
              <v-icon color="#ffc107" class="mr-2">mdi-trophy</v-icon>
              <div class="flex-grow-1">
                <div class="font-weight-bold text-body-2" style="color: #1a1a1a">
                  {{ pr.movement_name }}
                </div>
                <div class="text-caption" style="color: #666">
                  {{ pr.max_weight }} {{ pr.weight_unit || 'lbs' }}
                  <span v-if="pr.date_achieved"> • {{ formatDate(pr.date_achieved) }}</span>
                </div>
              </div>
            </div>
          </v-card>

          <v-btn
            v-if="personalRecords.length > 5"
            variant="text"
            size="small"
            color="#00bcd4"
            block
            class="mt-2"
            style="text-transform: none"
            @click="$router.push('/prs')"
          >
            View All {{ personalRecords.length }} PRs
            <v-icon end size="small">mdi-arrow-right</v-icon>
          </v-btn>
        </div>
      </v-card>

      <!-- Movement Progress Tracking -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-2" style="color: #1a1a1a">Track Progress</h2>
        <v-autocomplete
          v-model="selectedMovement"
          :items="movements"
          item-title="title"
          item-value="value"
          :loading="loadingMovements"
          placeholder="Search for a movement..."
          variant="outlined"
          density="compact"
          rounded="lg"
          clearable
          auto-select-first
          hide-details
          style="color: #1a1a1a"
          @update:model-value="fetchMovementHistory"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-magnify</v-icon>
          </template>
          <template #item="{ props, item }">
            <v-list-item v-bind="props" density="compact">
              <template #prepend>
                <v-icon
                  :color="item.raw.type === 'weightlifting' ? '#00bcd4' : '#666'"
                  size="small"
                >
                  mdi-dumbbell
                </v-icon>
              </template>
              <v-list-item-title class="text-body-2">
                {{ item.raw.title }}
              </v-list-item-title>
              <v-list-item-subtitle class="text-caption">
                {{ formatMovementType(item.raw.type) }}
              </v-list-item-subtitle>
            </v-list-item>
          </template>
        </v-autocomplete>
      </v-card>

      <!-- Movement History -->
      <v-card
        v-if="selectedMovement && movementHistory.length > 0"
        elevation="0"
        rounded="lg"
        class="pa-3 mb-3"
        style="background: white"
      >
        <h2 class="text-body-1 font-weight-bold mb-2" style="color: #1a1a1a">
          {{ selectedMovementName }} History
        </h2>

        <!-- Loading State -->
        <div v-if="loadingHistory" class="text-center py-4">
          <v-progress-circular indeterminate color="#00bcd4" size="32" />
        </div>

        <!-- History List -->
        <div v-else>
          <v-card
            v-for="(entry, index) in movementHistory"
            :key="index"
            elevation="0"
            rounded="lg"
            class="mb-2 pa-2"
            style="background: #f5f7fa"
          >
            <div class="d-flex align-center">
              <div class="flex-grow-1">
                <div class="font-weight-bold text-body-2" style="color: #1a1a1a">
                  {{ entry.sets }} × {{ entry.reps }} @ {{ entry.weight }} {{ entry.weight_unit || 'lbs' }}
                </div>
                <div class="text-caption" style="color: #666">
                  {{ formatDate(entry.workout_date) }}
                  <span v-if="entry.workout_name"> • {{ entry.workout_name }}</span>
                </div>
              </div>
              <v-chip
                v-if="entry.is_pr"
                size="x-small"
                color="#ffc107"
                class="ml-2"
              >
                <v-icon start size="x-small">mdi-trophy</v-icon>
                PR
              </v-chip>
            </div>
          </v-card>
        </div>
      </v-card>

      <!-- Recent Workouts Summary -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <div class="d-flex align-center justify-space-between mb-2">
          <h2 class="text-h6 font-weight-bold" style="color: #1a1a1a">Last 30 Days</h2>
          <v-btn
            size="small"
            variant="text"
            color="#00bcd4"
            style="text-transform: none"
            @click="$router.push('/dashboard')"
          >
            View All
          </v-btn>
        </div>

        <!-- Loading State -->
        <div v-if="loadingWorkouts" class="text-center py-4">
          <v-progress-circular indeterminate color="#00bcd4" size="32" />
        </div>

        <!-- Empty State -->
        <div v-else-if="recentWorkouts.length === 0" class="text-center py-4">
          <v-icon size="48" color="#ccc">mdi-calendar-blank</v-icon>
          <p class="text-body-2 mt-2" style="color: #666">No recent workouts</p>
        </div>

        <!-- Workouts List -->
        <div v-else>
          <v-card
            v-for="workout in recentWorkouts.slice(0, 30)"
            :key="workout.id"
            elevation="0"
            rounded="lg"
            class="mb-2 pa-2"
            style="background: #f5f7fa; cursor: pointer"
            @click="viewWorkout(workout.id)"
          >
            <div class="d-flex align-center mb-1">
              <v-icon color="#00bcd4" class="mr-2" size="small">mdi-dumbbell</v-icon>
              <div class="flex-grow-1">
                <div class="font-weight-bold text-body-2" style="color: #1a1a1a">
                  {{ workout.workout_name || 'Workout' }}
                </div>
                <div class="text-caption" style="color: #666">
                  {{ formatDate(workout.workout_date) }}
                  <span v-if="workout.total_time"> • {{ formatTime(workout.total_time) }}</span>
                </div>
              </div>
              <v-icon color="#ccc" size="small">mdi-chevron-right</v-icon>
            </div>

            <!-- Display movement performance -->
            <div v-if="workout.performance_movements && workout.performance_movements.length > 0" class="ml-6 mt-1">
              <div v-for="(perf, index) in workout.performance_movements" :key="index" class="text-caption mb-1" style="color: #666">
                <v-icon size="x-small" color="#00bcd4">mdi-weight-lifter</v-icon>
                <strong>{{ perf.movement?.name || 'Movement' }}:</strong>
                <span v-if="perf.weight"> {{ perf.weight }}lbs</span>
                <span v-if="perf.sets"> {{ perf.sets }}x</span><span v-if="perf.reps">{{ perf.reps }}</span>
                <span v-if="perf.distance"> {{ perf.distance }}m</span>
                <span v-if="perf.time_seconds"> {{ formatTime(perf.time_seconds) }}</span>
              </div>
            </div>

            <!-- Display WOD performance -->
            <div v-if="workout.performance_wods && workout.performance_wods.length > 0" class="ml-6 mt-1">
              <div v-for="(perf, index) in workout.performance_wods" :key="index" class="text-caption mb-1" style="color: #666">
                <v-icon size="x-small" color="#ffc107">mdi-fire</v-icon>
                <strong>{{ perf.wod?.name || 'WOD' }}:</strong>
                <span v-if="perf.time_seconds"> {{ formatTime(perf.time_seconds) }}</span>
                <span v-if="perf.rounds && perf.reps"> {{ perf.rounds }}+{{ perf.reps }}</span>
                <span v-else-if="perf.rounds"> {{ perf.rounds }} rounds</span>
                <span v-else-if="perf.reps"> {{ perf.reps }} reps</span>
                <span v-if="perf.score_value"> ({{ perf.score_value }})</span>
              </div>
            </div>
          </v-card>
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

const router = useRouter()
const activeTab = ref('performance')

// State
const selectedMovement = ref(null)
const movements = ref([])
const personalRecords = ref([])
const movementHistory = ref([])
const recentWorkouts = ref([])

const loadingMovements = ref(false)
const loadingPRs = ref(false)
const loadingHistory = ref(false)
const loadingWorkouts = ref(false)
const error = ref(null)

// Computed
const selectedMovementName = computed(() => {
  if (!selectedMovement.value) return ''
  const movement = movements.value.find(m => m.value === selectedMovement.value)
  return movement ? movement.title : ''
})

// Fetch available movements
async function fetchMovements() {
  loadingMovements.value = true
  try {
    const response = await axios.get('/api/movements')
    movements.value = response.data.movements.map((m) => ({
      value: m.id,
      title: m.name,
      type: m.type,
    }))
  } catch (err) {
    console.error('Failed to fetch movements:', err)
    error.value = 'Failed to load movements'
  } finally {
    loadingMovements.value = false
  }
}

// Fetch personal records
async function fetchPersonalRecords() {
  loadingPRs.value = true
  try {
    const response = await axios.get('/api/movements/personal-records')
    personalRecords.value = response.data.personal_records || []
  } catch (err) {
    console.error('Failed to fetch personal records:', err)
    // Silently fail - not critical
    personalRecords.value = []
  } finally {
    loadingPRs.value = false
  }
}

// Fetch movement history for selected movement
async function fetchMovementHistory() {
  if (!selectedMovement.value) {
    movementHistory.value = []
    return
  }

  loadingHistory.value = true
  try {
    const response = await axios.get(`/api/movements/${selectedMovement.value}/history`)
    movementHistory.value = response.data.history || []
  } catch (err) {
    console.error('Failed to fetch movement history:', err)
    error.value = 'Failed to load movement history'
    movementHistory.value = []
  } finally {
    loadingHistory.value = false
  }
}

// Fetch recent workouts (last 30 days)
async function fetchRecentWorkouts() {
  loadingWorkouts.value = true
  try {
    const response = await axios.get('/api/workouts')
    const allWorkouts = response.data.workouts || []

    // Filter to last 30 days
    const now = new Date()
    const thirtyDaysAgo = new Date(now.getTime() - (30 * 24 * 60 * 60 * 1000))

    recentWorkouts.value = allWorkouts
      .filter(w => {
        const workoutDate = new Date(w.workout_date)
        return workoutDate >= thirtyDaysAgo
      })
      .sort((a, b) => new Date(b.workout_date) - new Date(a.workout_date))
  } catch (err) {
    console.error('Failed to fetch recent workouts:', err)
    recentWorkouts.value = []
  } finally {
    loadingWorkouts.value = false
  }
}

// Refresh all data
async function refreshData() {
  await Promise.all([
    fetchMovements(),
    fetchPersonalRecords(),
    fetchRecentWorkouts()
  ])
  if (selectedMovement.value) {
    await fetchMovementHistory()
  }
}

// Format date for display
function formatDate(dateString) {
  // Parse as local date to avoid timezone conversion issues
  // Extract YYYY-MM-DD from the date string
  const datePart = dateString.split('T')[0]
  const [year, month, day] = datePart.split('-').map(Number)
  const date = new Date(year, month - 1, day) // Month is 0-indexed

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
    const options = { month: 'short', day: 'numeric', year: 'numeric' }
    return date.toLocaleDateString('en-US', options)
  }
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

// Format movement type
function formatMovementType(type) {
  if (!type) return ''
  return type.charAt(0).toUpperCase() + type.slice(1)
}

// View workout details
function viewWorkout(workoutId) {
  console.log('View workout details:', workoutId)
  router.push(`/workouts/${workoutId}`)
}

// Load data on mount
onMounted(async () => {
  await refreshData()
})
</script>
