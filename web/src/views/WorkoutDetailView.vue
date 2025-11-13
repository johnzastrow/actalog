<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-btn icon="mdi-arrow-left" color="white" @click="$router.back()" />
      <v-toolbar-title class="text-white font-weight-bold">Workout Details</v-toolbar-title>
      <v-spacer />
      <v-btn
        v-if="workout"
        icon="mdi-pencil"
        color="white"
        @click="editWorkout"
      />
      <v-btn
        v-if="workout"
        icon="mdi-delete"
        color="white"
        @click="confirmDelete"
      />
    </v-app-bar>

    <v-container class="pa-3" style="margin-top: 56px; margin-bottom: 70px">
      <!-- Loading State -->
      <div v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="#00bcd4" size="64" />
        <p class="mt-4 text-body-2" style="color: #666">Loading workout...</p>
      </div>

      <!-- Error State -->
      <v-alert v-else-if="error" type="error" class="mb-4">
        {{ error }}
      </v-alert>

      <!-- Workout Details -->
      <div v-else-if="workout">
        <!-- Workout Header Card -->
        <v-card elevation="0" rounded="lg" class="mb-3 pa-4" style="background: white">
          <div class="d-flex align-center mb-2">
            <v-icon color="#00bcd4" size="32" class="mr-2">mdi-dumbbell</v-icon>
            <div class="flex-grow-1">
              <h2 class="text-h5 font-weight-bold" style="color: #1a1a1a">
                {{ workout.workout_name || 'Custom Workout' }}
              </h2>
              <div class="text-caption" style="color: #666">
                {{ formatDate(workout.workout_date) }}
                <span v-if="workout.workout_type"> • {{ formatWorkoutType(workout.workout_type) }}</span>
              </div>
            </div>
            <v-chip
              v-if="hasPR"
              color="#ffc107"
              size="small"
            >
              <v-icon size="small" class="mr-1">mdi-trophy</v-icon>
              PR
            </v-chip>
          </div>

          <!-- Total Time -->
          <div v-if="workout.total_time" class="mt-3">
            <v-chip color="#00bcd4" variant="outlined" size="small">
              <v-icon size="small" class="mr-1">mdi-clock-outline</v-icon>
              {{ formatTime(workout.total_time) }}
            </v-chip>
          </div>
        </v-card>

        <!-- Notes Section -->
        <v-card v-if="workout.notes" elevation="0" rounded="lg" class="mb-3 pa-4" style="background: white">
          <div class="d-flex align-center mb-2">
            <v-icon color="#00bcd4" size="small" class="mr-2">mdi-note-text</v-icon>
            <h3 class="text-body-1 font-weight-bold" style="color: #1a1a1a">Notes</h3>
          </div>
          <p class="text-body-2" style="color: #666">{{ workout.notes }}</p>
        </v-card>

        <!-- Movements Section -->
        <v-card
          v-if="workout.performance_movements && workout.performance_movements.length > 0"
          elevation="0"
          rounded="lg"
          class="mb-3 pa-4"
          style="background: white"
        >
          <div class="d-flex align-center mb-3">
            <v-icon color="#00bcd4" size="small" class="mr-2">mdi-weight-lifter</v-icon>
            <h3 class="text-body-1 font-weight-bold" style="color: #1a1a1a">
              Movements ({{ workout.performance_movements.length }})
            </h3>
          </div>

          <div class="movements-list">
            <v-card
              v-for="(movement, index) in workout.performance_movements"
              :key="index"
              elevation="0"
              class="mb-2 pa-3"
              style="background: #f5f7fa; border: 1px solid #e0e0e0"
              rounded="lg"
            >
              <div class="d-flex align-center">
                <div class="flex-grow-1">
                  <div class="d-flex align-center mb-1">
                    <span class="text-body-2 font-weight-bold" style="color: #1a1a1a">
                      {{ movement.movement_name || 'Unknown Movement' }}
                    </span>
                    <v-chip
                      v-if="movement.is_pr"
                      color="#ffc107"
                      size="x-small"
                      class="ml-2"
                      style="height: 18px"
                    >
                      <v-icon size="x-small" class="mr-1">mdi-trophy</v-icon>
                      PR
                    </v-chip>
                  </div>

                  <!-- Movement Details -->
                  <div class="text-caption" style="color: #666">
                    <span v-if="movement.sets">{{ movement.sets }} sets</span>
                    <span v-if="movement.sets && movement.reps"> × </span>
                    <span v-if="movement.reps">{{ movement.reps }} reps</span>
                    <span v-if="movement.weight"> @ {{ movement.weight }}lb</span>
                  </div>

                  <div v-if="movement.time_seconds" class="text-caption mt-1" style="color: #00bcd4">
                    <v-icon size="x-small" color="#00bcd4">mdi-clock-outline</v-icon>
                    {{ formatTime(movement.time_seconds) }}
                  </div>

                  <div v-if="movement.distance" class="text-caption mt-1" style="color: #00bcd4">
                    <v-icon size="x-small" color="#00bcd4">mdi-map-marker-distance</v-icon>
                    {{ movement.distance }}{{ movement.distance_unit || 'm' }}
                  </div>

                  <div v-if="movement.notes" class="text-caption mt-1" style="color: #999; font-style: italic">
                    {{ movement.notes }}
                  </div>
                </div>

                <!-- Movement Type Icon -->
                <v-icon
                  :color="getMovementTypeColor(movement.movement_type)"
                  size="large"
                >
                  {{ getMovementTypeIcon(movement.movement_type) }}
                </v-icon>
              </div>
            </v-card>
          </div>
        </v-card>

        <!-- WODs Section -->
        <v-card
          v-if="workout.performance_wods && workout.performance_wods.length > 0"
          elevation="0"
          rounded="lg"
          class="mb-3 pa-4"
          style="background: white"
        >
          <div class="d-flex align-center mb-3">
            <v-icon color="#ffc107" size="small" class="mr-2">mdi-fire</v-icon>
            <h3 class="text-body-1 font-weight-bold" style="color: #1a1a1a">
              WODs ({{ workout.performance_wods.length }})
            </h3>
          </div>

          <div>
            <v-card
              v-for="wod in workout.performance_wods"
              :key="wod.id"
              elevation="0"
              class="mb-2 pa-3"
              style="background: #fff8e1; border: 1px solid #ffc107"
              rounded="lg"
            >
              <div class="font-weight-bold text-body-2 mb-1" style="color: #1a1a1a">
                {{ wod.wod_name || 'Custom WOD' }}
              </div>
              <div v-if="wod.score_value" class="text-caption" style="color: #666">
                Score: {{ wod.score_value }}
              </div>
              <div v-if="wod.rounds" class="text-caption mt-1" style="color: #f57c00">
                <v-icon size="x-small" color="#f57c00">mdi-repeat</v-icon>
                {{ wod.rounds }} rounds
              </div>
              <div v-if="wod.time_seconds" class="text-caption mt-1" style="color: #f57c00">
                <v-icon size="x-small" color="#f57c00">mdi-clock-outline</v-icon>
                {{ formatTime(wod.time_seconds) }}
              </div>
              <div v-if="wod.notes" class="text-caption mt-1" style="color: #999; font-style: italic">
                {{ wod.notes }}
              </div>
            </v-card>
          </div>
        </v-card>

        <!-- Empty State -->
        <v-card
          v-if="(!workout.performance_movements || workout.performance_movements.length === 0) && (!workout.performance_wods || workout.performance_wods.length === 0)"
          elevation="0"
          rounded="lg"
          class="pa-6 text-center"
          style="background: white"
        >
          <v-icon size="64" color="#ccc">mdi-clipboard-text-outline</v-icon>
          <p class="text-body-1 mt-2" style="color: #666">
            No movements or WODs logged for this workout
          </p>
        </v-card>

        <!-- Action Buttons -->
        <v-row dense class="mt-3">
          <v-col cols="6">
            <v-btn
              block
              color="#00bcd4"
              variant="flat"
              size="large"
              rounded="lg"
              prepend-icon="mdi-pencil"
              @click="editWorkout"
              style="text-transform: none; font-weight: 600"
            >
              Edit Workout
            </v-btn>
          </v-col>
          <v-col cols="6">
            <v-btn
              block
              color="error"
              variant="outlined"
              size="large"
              rounded="lg"
              prepend-icon="mdi-delete"
              @click="confirmDelete"
              style="text-transform: none; font-weight: 600"
            >
              Delete
            </v-btn>
          </v-col>
        </v-row>
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

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="text-h6">Delete Workout?</v-card-title>
        <v-card-text>
          Are you sure you want to delete this workout? This action cannot be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="deleteDialog = false" style="text-transform: none">
            Cancel
          </v-btn>
          <v-btn
            color="error"
            :loading="deleting"
            @click="deleteWorkout"
            style="text-transform: none"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/axios'

const router = useRouter()
const route = useRoute()
const activeTab = ref('')

const workout = ref(null)
const loading = ref(false)
const error = ref(null)
const deleteDialog = ref(false)
const deleting = ref(false)

const workoutId = computed(() => route.params.id)

// Check if workout has any PR movements
const hasPR = computed(() => {
  return workout.value?.performance_movements?.some(m => m.is_pr) || false
})

// Fetch workout details
async function fetchWorkout() {
  loading.value = true
  error.value = null

  try {
    const response = await axios.get(`/api/workouts/${workoutId.value}`)
    workout.value = response.data
    console.log('Fetched workout:', workout.value)
  } catch (err) {
    console.error('Failed to fetch workout:', err)
    if (err.response?.status === 404) {
      error.value = 'Workout not found'
    } else if (err.response?.status === 403) {
      error.value = 'You do not have permission to view this workout'
    } else {
      error.value = err.response?.data?.message || 'Failed to load workout'
    }
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
    return 'Today, ' + date.toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })
  } else if (dateOnly.getTime() === yesterdayOnly.getTime()) {
    return 'Yesterday, ' + date.toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })
  } else {
    return date.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })
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

  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60

  if (hours > 0) {
    return `${hours}h ${minutes}m ${secs}s`
  } else if (minutes > 0) {
    return `${minutes}m ${secs}s`
  } else {
    return `${secs}s`
  }
}

// Get movement type color
function getMovementTypeColor(type) {
  const colors = {
    weightlifting: '#00bcd4',
    gymnastics: '#9c27b0',
    cardio: '#f44336',
    bodyweight: '#4caf50'
  }
  return colors[type?.toLowerCase()] || '#666'
}

// Get movement type icon
function getMovementTypeIcon(type) {
  const icons = {
    weightlifting: 'mdi-dumbbell',
    gymnastics: 'mdi-gymnastics',
    cardio: 'mdi-run',
    bodyweight: 'mdi-arm-flex'
  }
  return icons[type?.toLowerCase()] || 'mdi-weight-lifter'
}

// Edit workout
function editWorkout() {
  // TODO: Implement edit workout functionality
  // For now, navigate to log workout with pre-filled data
  console.log('Edit workout:', workoutId.value)
  router.push(`/workouts/log?edit=${workoutId.value}`)
}

// Confirm delete
function confirmDelete() {
  deleteDialog.value = true
}

// Delete workout
async function deleteWorkout() {
  deleting.value = true

  try {
    await axios.delete(`/api/workouts/${workoutId.value}`)
    // Navigate back to dashboard after successful deletion
    router.push('/dashboard')
  } catch (err) {
    console.error('Failed to delete workout:', err)
    error.value = err.response?.data?.message || 'Failed to delete workout'
    deleteDialog.value = false
  } finally {
    deleting.value = false
  }
}

// Load workout on mount
onMounted(() => {
  fetchWorkout()
})
</script>

<style scoped>
.movements-list {
  max-height: none;
}
</style>
