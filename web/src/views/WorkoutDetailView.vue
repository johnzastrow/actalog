<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <!-- Back Button -->
      <v-btn
        variant="text"
        color="primary"
        class="mb-2"
        @click="router.back()"
      >
        <v-icon start>mdi-arrow-left</v-icon>
        Back
      </v-btn>

      <!-- Loading State -->
      <div v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="primary" size="64" />
        <p class="mt-4 text-body-2 text-medium-emphasis">Loading workout...</p>
      </div>

      <!-- Error State -->
      <v-alert v-else-if="error" type="error" class="mb-4">
        {{ error }}
      </v-alert>

      <!-- Workout Details -->
      <div v-else-if="workout">
        <!-- Workout Header Card -->
        <v-card elevation="0" rounded="lg" class="mb-2 pa-2" bg-color="surface">
          <div class="d-flex align-center mb-2">
            <v-icon color="primary" size="32" class="mr-2">mdi-dumbbell</v-icon>
            <div class="flex-grow-1">
              <h2 class="text-h5 font-weight-bold" >
                {{ workout.workout_name || 'Custom Workout' }}
              </h2>
              <div class="text-caption text-medium-emphasis">
                {{ formatDate(workout.workout_date) }}
              </div>
            </div>
            <v-chip
              v-if="hasPR"
              color="primary"
              size="small"
            >
              <v-icon size="small" class="mr-1">mdi-trophy</v-icon>
              PR
            </v-chip>
          </div>

          <!-- Total Time -->
          <div v-if="workout.total_time" class="mt-3">
            <v-chip color="primary"  size="small">
              <v-icon size="small" class="mr-1">mdi-clock-outline</v-icon>
              {{ formatTime(workout.total_time) }}
            </v-chip>
          </div>
        </v-card>

        <!-- Notes Section -->
        <v-card v-if="workout.notes" elevation="0" rounded="lg" class="mb-2 pa-2" bg-color="surface">
          <div class="d-flex align-center mb-2">
            <v-icon color="primary" size="small" class="mr-2">mdi-note-text</v-icon>
            <h3 class="text-body-1 font-weight-bold" >Notes</h3>
          </div>
          <div class="text-body-2 text-medium-emphasis">
            <markdown-renderer :content="workout.notes" />
          </div>
        </v-card>

        <!-- Movements Section -->
        <v-card
          v-if="workout.performance_movements && workout.performance_movements.length > 0"
          elevation="0"
          rounded="lg"
          class="mb-2 pa-2"
          bg-color="surface"
        >
          <div class="d-flex align-center mb-3">
            <v-icon color="primary" size="small" class="mr-2">mdi-weight-lifter</v-icon>
            <h3 class="text-body-1 font-weight-bold" >
              Movements ({{ workout.performance_movements.length }})
            </h3>
          </div>

          <div class="movements-list">
            <v-card
              v-for="(movement, index) in workout.performance_movements"
              :key="index"
              elevation="0"
              class="mb-2 pa-2"
              style="background-color: rgb(var(--v-theme-background))"
              rounded="lg"
            >
              <div class="d-flex align-center">
                <div class="flex-grow-1">
                  <div class="d-flex align-center mb-1">
                    <span class="text-body-2 font-weight-bold" >
                      {{ movement.movement?.name || 'Unknown Movement' }}
                    </span>
                    <v-chip
                      v-if="movement.is_pr"
                      color="primary"
                      size="x-small"
                      class="ml-2"
                      style="height: 18px"
                    >
                      <v-icon size="x-small" class="mr-1">mdi-trophy</v-icon>
                      PR
                    </v-chip>
                  </div>

                  <!-- Movement Details -->
                  <div class="text-caption text-medium-emphasis">
                    <span v-if="movement.sets">{{ movement.sets }} sets</span>
                    <span v-if="movement.sets && movement.reps"> × </span>
                    <span v-if="movement.reps">{{ movement.reps }} reps</span>
                    <span v-if="movement.weight"> @ {{ movement.weight }}lb</span>
                  </div>

                  <div v-if="movement.time_seconds" class="text-caption mt-1" style="color: rgb(var(--v-theme-primary))">
                    <v-icon size="x-small" color="primary">mdi-clock-outline</v-icon>
                    {{ formatTime(movement.time_seconds) }}
                  </div>

                  <div v-if="movement.distance" class="text-caption mt-1" style="color: rgb(var(--v-theme-primary))">
                    <v-icon size="x-small" color="primary">mdi-map-marker-distance</v-icon>
                    {{ movement.distance }}{{ movement.distance_unit || 'm' }}
                  </div>

                  <div v-if="movement.notes" class="text-caption mt-1" style="color: #999; font-style: italic">
                    <markdown-renderer :content="movement.notes" />
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
          class="mb-2 pa-2"
          bg-color="surface"
        >
          <div class="d-flex align-center mb-3">
            <v-icon color="primary" size="small" class="mr-2">mdi-fire</v-icon>
            <h3 class="text-body-1 font-weight-bold" >
              WODs ({{ workout.performance_wods.length }})
            </h3>
          </div>

          <div>
            <v-card
              v-for="wod in workout.performance_wods"
              :key="wod.id"
              elevation="0"
              class="mb-2 pa-2"
              style="background: #fff8e1; border: 1px solid #ffc107"
              rounded="lg"
            >
              <div class="font-weight-bold text-body-2 mb-1" >
                {{ wod.wod?.name || 'Custom WOD' }}
              </div>
              <div v-if="wod.score_value" class="text-caption text-medium-emphasis">
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
                <markdown-renderer :content="wod.notes" />
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
          bg-color="surface"
        >
          <v-icon size="64" color="surface-variant">mdi-clipboard-text-outline</v-icon>
          <p class="text-body-1 mt-2 text-medium-emphasis">
            No movements or WODs logged for this workout
          </p>
        </v-card>

        <!-- Action Buttons -->
        <v-row dense class="mt-3">
          <v-col cols="6">
            <v-btn
              block
              color="primary"
              variant="flat"
              size="large"
              rounded="lg"
              prepend-icon="mdi-pencil"
              style="text-transform: none; font-weight: 600"
              @click="editWorkout"
            >
              Edit Workout
            </v-btn>
          </v-col>
          <v-col cols="6">
            <v-btn
              block
              color="error"
              
              size="large"
              rounded="lg"
              prepend-icon="mdi-delete"
              style="text-transform: none; font-weight: 600"
              @click="confirmDelete"
            >
              Delete
            </v-btn>
          </v-col>
        </v-row>
      </div>
    </v-container>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="text-h6">Delete Workout?</v-card-title>
        <v-card-text>
          Are you sure you want to delete this workout? This action cannot be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn style="text-transform: none" @click="deleteDialog = false">
            Cancel
          </v-btn>
          <v-btn
            color="error"
            :loading="deleting"
            style="text-transform: none"
            @click="deleteWorkout"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/axios'
import { useSettingsStore } from '@/stores/settings'
import { formatDateInTimezone, getTodayInTimezone, getYesterdayInTimezone } from '@/utils/timezone'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

const router = useRouter()
const route = useRoute()
const settingsStore = useSettingsStore()
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

// Format date for display using user's timezone
function formatDate(dateString) {
  const tz = settingsStore.timezone
  const datePart = dateString.split('T')[0]

  // Get today and yesterday in user's timezone
  const todayStr = getTodayInTimezone(tz)
  const yesterdayStr = getYesterdayInTimezone(tz)

  const formattedDate = formatDateInTimezone(dateString, tz, 'MMM d, yyyy')

  if (datePart === todayStr) {
    return 'Today, ' + formattedDate
  } else if (datePart === yesterdayStr) {
    return 'Yesterday, ' + formattedDate
  } else {
    return formatDateInTimezone(dateString, tz, 'EEE, MMM d, yyyy')
  }
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

// Edit workout - navigates to log view with edit mode
function editWorkout() {
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
