<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-toolbar-title class="text-white font-weight-bold">Workouts</v-toolbar-title>
      <v-spacer />
      <v-btn
        icon="mdi-plus"
        color="#00bcd4"
        variant="flat"
        to="/workouts/log"
        style="background: #00bcd4"
      />
    </v-app-bar>

    <v-container class="pa-3" style="margin-top: 56px; margin-bottom: 70px">
      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable @click:close="error = null" class="mb-4">
        {{ error }}
      </v-alert>

      <div>
        <!-- Loading State -->
        <div v-if="loading" class="text-center py-8">
          <v-progress-circular indeterminate color="#00bcd4" size="64" />
          <p class="mt-4 text-medium-emphasis">Loading workouts...</p>
        </div>

        <!-- Empty State -->
        <v-card
          v-else-if="!loading && workouts.length === 0"
          elevation="0"
          rounded="lg"
          class="pa-8 text-center"
          style="background: white"
        >
          <v-icon size="64" color="#ccc">mdi-dumbbell</v-icon>
          <p class="text-h6 mt-4" style="color: #2c3e50">No workouts logged yet</p>
          <p class="text-body-2" style="color: #666">
            Start logging your workouts to see them here
          </p>
          <v-btn
            color="#00bcd4"
            class="mt-4"
            prepend-icon="mdi-plus"
            to="/workouts/log"
            rounded="lg"
            style="text-transform: none; font-weight: 600"
          >
            Log Your First Workout
          </v-btn>
        </v-card>

        <!-- Workouts List -->
        <div v-else>
          <v-card
            v-for="workout in workouts"
            :key="workout.id"
            elevation="0"
            rounded="lg"
            class="mb-2 pa-3"
            style="background: white"
            @click="viewWorkout(workout.id)"
          >
            <div class="d-flex align-center mb-1">
              <v-icon
                :color="workout.workout_type === 'named_wod' ? '#00bcd4' : '#666'"
                class="mr-2"
                size="small"
              >
                mdi-dumbbell
              </v-icon>
              <div class="flex-grow-1">
                <div class="font-weight-bold text-body-2" style="color: #1a1a1a">
                  {{ formatDate(workout.workout_date) }}
                  <v-chip
                    v-if="workout.workout_name"
                    size="x-small"
                    color="#00bcd4"
                    class="ml-1"
                    style="color: white"
                  >
                    {{ workout.workout_name }}
                  </v-chip>
                </div>
              </div>
              <v-icon color="#ccc" size="small">mdi-chevron-right</v-icon>
            </div>

            <!-- Display movements -->
            <div v-if="workout.movements && workout.movements.length > 0" class="ml-7">
              <div
                v-for="(movement, idx) in workout.movements"
                :key="movement.id"
                class="mt-1 text-caption d-flex align-center"
                style="color: #333"
              >
                <v-icon size="x-small" color="#00bcd4">mdi-chevron-right</v-icon>
                <strong style="color: #1a1a1a">
                  {{ movement.movement?.name || 'Unknown Movement' }}
                </strong>
                <span v-if="movement.weight"> - {{ movement.weight }} lbs</span>
                <span v-if="movement.sets"> × {{ movement.sets }} sets</span>
                <span v-if="movement.reps"> × {{ movement.reps }} reps</span>
                <v-chip
                  v-if="movement.is_rx"
                  size="x-small"
                  color="#00bcd4"
                  class="ml-1"
                  style="color: white; height: 16px"
                >
                  Rx
                </v-chip>
                <v-chip
                  v-if="movement.is_pr"
                  size="x-small"
                  color="#ffc107"
                  class="ml-1"
                  style="color: white; height: 16px"
                >
                  <v-icon size="x-small" class="mr-1">mdi-trophy</v-icon>
                  PR
                </v-chip>
              </div>
            </div>

            <div v-if="workout.notes" class="mt-2 text-caption" style="color: #666">
              📝 {{ workout.notes }}
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
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'

const router = useRouter()
const activeTab = ref('workouts')

// State
const workouts = ref([])
const loading = ref(false)
const error = ref(null)

// Fetch workouts from API
async function fetchWorkouts() {
  loading.value = true
  error.value = null

  try {
    const response = await axios.get('/api/workouts')
    workouts.value = response.data.workouts || []
    console.log('Fetched workouts:', workouts.value)
  } catch (err) {
    console.error('Failed to fetch workouts:', err)
    if (err.response) {
      error.value = err.response.data?.message || `Error ${err.response.status}`
    } else if (err.request) {
      error.value = 'No response from server. Is the backend running?'
    } else {
      error.value = 'Failed to fetch workouts'
    }
  } finally {
    loading.value = false
  }
}

// Format date for display
function formatDate(dateString) {
  const date = new Date(dateString)
  const options = {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }
  return date.toLocaleDateString('en-US', options)
}

// View workout details
function viewWorkout(workoutId) {
  console.log('View workout:', workoutId)
  // TODO: Navigate to workout detail page
  // router.push(`/workouts/${workoutId}`)
}

// Load workouts on mount
onMounted(() => {
  fetchWorkouts()
})
</script>
