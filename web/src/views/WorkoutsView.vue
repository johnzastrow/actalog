<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4 mb-6">Workouts</h1>
      </v-col>

      <!-- Error Alert -->
      <v-col v-if="error" cols="12">
        <v-alert type="error" closable @click:close="error = null">
          {{ error }}
        </v-alert>
      </v-col>

      <v-col cols="12">
        <v-card elevation="2" rounded="lg">
          <v-card-title class="d-flex justify-space-between align-center">
            <span>Workout History</span>
            <v-btn
              color="primary"
              prepend-icon="mdi-plus"
              to="/workouts/log"
              variant="elevated"
            >
              Log Workout
            </v-btn>
          </v-card-title>
          <v-card-text>
            <!-- Loading State -->
            <div v-if="loading" class="text-center py-8">
              <v-progress-circular indeterminate color="primary" />
              <p class="mt-4 text-medium-emphasis">Loading workouts...</p>
            </div>

            <!-- Empty State -->
            <div
              v-else-if="!loading && workouts.length === 0"
              class="text-center py-8"
            >
              <v-icon size="64" color="grey">mdi-dumbbell</v-icon>
              <p class="text-h6 mt-4">No workouts logged yet</p>
              <p class="text-medium-emphasis">
                Start logging your workouts to see them here
              </p>
              <v-btn
                color="primary"
                class="mt-4"
                prepend-icon="mdi-plus"
                to="/workouts/log"
              >
                Log Your First Workout
              </v-btn>
            </div>

            <!-- Workouts List -->
            <v-list v-else>
              <v-list-item
                v-for="workout in workouts"
                :key="workout.id"
                class="mb-2"
                border
                rounded
              >
                <template v-slot:prepend>
                  <v-icon
                    :color="workout.workout_type === 'named_wod' ? 'primary' : 'secondary'"
                  >
                    mdi-dumbbell
                  </v-icon>
                </template>

                <v-list-item-title class="font-weight-bold">
                  {{ formatDate(workout.workout_date) }}
                  <v-chip
                    v-if="workout.workout_name"
                    size="small"
                    color="primary"
                    class="ml-2"
                  >
                    {{ workout.workout_name }}
                  </v-chip>
                </v-list-item-title>

                <v-list-item-subtitle>
                  <!-- Display movements -->
                  <div v-if="workout.movements && workout.movements.length > 0">
                    <div
                      v-for="(movement, idx) in workout.movements"
                      :key="movement.id"
                      class="mt-1"
                    >
                      <v-icon size="small">mdi-chevron-right</v-icon>
                      <strong>{{ movement.movement?.name || 'Unknown Movement' }}</strong>
                      <span v-if="movement.weight"> - {{ movement.weight }} lbs</span>
                      <span v-if="movement.sets"> × {{ movement.sets }} sets</span>
                      <span v-if="movement.reps"> × {{ movement.reps }} reps</span>
                      <v-chip
                        v-if="movement.is_rx"
                        size="x-small"
                        color="success"
                        class="ml-1"
                      >
                        Rx
                      </v-chip>
                    </div>
                  </div>
                  <div v-if="workout.notes" class="mt-2 text-caption">
                    📝 {{ workout.notes }}
                  </div>
                </v-list-item-subtitle>

                <template v-slot:append>
                  <v-btn
                    icon="mdi-chevron-right"
                    variant="text"
                    size="small"
                    @click="viewWorkout(workout.id)"
                  />
                </template>
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
