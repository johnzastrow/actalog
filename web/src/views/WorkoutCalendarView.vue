<template>
  <div class="mobile-view-wrapper">
    <!-- Header -->
    <div style="background: #2c3e50; padding: 16px; position: fixed; top: 56px; left: 0; right: 0; z-index: 5">
      <h2 style="color: white; margin: 0; font-size: 1.25rem">Workout Calendar</h2>
    </div>

    <!-- Content (with padding for fixed header) -->
    <div style="padding: 72px 16px 16px 16px; overflow-y: auto">
      <!-- Calendar Component -->
      <WorkoutCalendar :workoutDates="workoutDates" @daySelected="handleDaySelected" />

      <!-- Selected Day Workouts -->
      <div v-if="selectedDate" class="mt-4">
        <h3 class="text-h6 mb-3" style="color: #1a1a1a">
          {{ formatSelectedDate(selectedDate) }}
        </h3>

        <div v-if="loadingDayWorkouts" class="text-center py-8">
          <v-progress-circular indeterminate color="#00bcd4"></v-progress-circular>
        </div>

        <div v-else-if="selectedDayWorkouts.length === 0" class="text-center py-8">
          <v-icon size="48" color="#ccc">mdi-calendar-blank</v-icon>
          <p class="mt-2" style="color: #666">No workouts on this day</p>
        </div>

        <v-card
          v-else
          v-for="workout in selectedDayWorkouts"
          :key="workout.id"
          elevation="0"
          rounded="lg"
          class="mb-3"
          style="background: white"
          @click="viewWorkout(workout.id)"
        >
          <v-card-text>
            <div class="d-flex justify-space-between align-center mb-2">
              <div class="text-subtitle-1 font-weight-medium" style="color: #1a1a1a">
                {{ workout.workout_name || workout.template_name || 'Workout' }}
              </div>
              <v-chip v-if="workout.workout_type" size="small" color="#00bcd4" text-color="white">
                {{ workout.workout_type }}
              </v-chip>
            </div>

            <!-- Movements -->
            <div v-if="workout.movements && workout.movements.length > 0" class="mb-2">
              <div class="text-caption font-weight-medium mb-1" style="color: #666">
                Movements
              </div>
              <div v-for="movement in workout.movements" :key="movement.id" class="d-flex align-center mb-1">
                <v-icon size="small" :color="movement.is_pr ? '#ffc107' : '#00bcd4'" class="mr-1">
                  {{ movement.is_pr ? 'mdi-trophy' : 'mdi-dumbbell' }}
                </v-icon>
                <span class="text-body-2" style="color: #1a1a1a">
                  {{ movement.movement_name }}
                  <span v-if="movement.sets" style="color: #666">
                    - {{ movement.sets }}x{{ movement.reps }}
                    <span v-if="movement.weight">@ {{ movement.weight }}lbs</span>
                  </span>
                </span>
              </div>
            </div>

            <!-- WODs -->
            <div v-if="workout.wods && workout.wods.length > 0">
              <div class="text-caption font-weight-medium mb-1" style="color: #666">
                WODs
              </div>
              <div v-for="wod in workout.wods" :key="wod.id" class="d-flex align-center mb-1">
                <v-icon size="small" :color="wod.is_pr ? '#ffc107' : '#00bcd4'" class="mr-1">
                  {{ wod.is_pr ? 'mdi-trophy' : 'mdi-run-fast' }}
                </v-icon>
                <span class="text-body-2" style="color: #1a1a1a">
                  {{ wod.wod_name }}
                  <span v-if="wod.score_value" style="color: #666">
                    - {{ wod.score_value }}
                  </span>
                </span>
              </div>
            </div>

            <!-- Notes -->
            <div v-if="workout.notes" class="mt-2">
              <div class="text-caption" style="color: #999">{{ workout.notes }}</div>
            </div>
          </v-card-text>
        </v-card>
      </div>
    </div>

    <!-- Bottom Navigation -->
    <v-bottom-navigation style="position: fixed; bottom: 0; left: 0; right: 0" elevation="8">
      <v-btn to="/dashboard">
        <v-icon>mdi-view-dashboard</v-icon>
        <span>Dashboard</span>
      </v-btn>
      <v-btn to="/log-workout">
        <v-icon>mdi-plus-circle</v-icon>
        <span>Log</span>
      </v-btn>
      <v-btn to="/workouts">
        <v-icon>mdi-history</v-icon>
        <span>History</span>
      </v-btn>
      <v-btn to="/profile">
        <v-icon>mdi-account</v-icon>
        <span>Profile</span>
      </v-btn>
    </v-bottom-navigation>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'
import WorkoutCalendar from '@/components/WorkoutCalendar.vue'

const router = useRouter()
const workouts = ref([])
const workoutDates = ref([])
const selectedDate = ref(null)
const selectedDayWorkouts = ref([])
const loadingDayWorkouts = ref(false)

async function fetchWorkouts() {
  try {
    const response = await axios.get('/api/workouts')
    const workoutsData = Array.isArray(response.data) ? response.data : (response.data.workouts || [])
    workouts.value = workoutsData
    workoutDates.value = workoutsData.map(w => w.workout_date)
  } catch (error) {
    console.error('Failed to fetch workouts:', error)
  }
}

function handleDaySelected(date) {
  selectedDate.value = date
  loadingDayWorkouts.value = true

  // Create date string in YYYY-MM-DD format from the selected date (local timezone)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const dateStr = `${year}-${month}-${day}`

  selectedDayWorkouts.value = workouts.value.filter(w => {
    // Extract YYYY-MM-DD from workout_date (handles both ISO strings and date objects)
    const workoutDateStr = w.workout_date.split('T')[0]
    return workoutDateStr === dateStr
  })

  loadingDayWorkouts.value = false
}

function formatSelectedDate(date) {
  return date.toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
    year: 'numeric'
  })
}

function viewWorkout(id) {
  router.push(`/workouts/${id}`)
}

onMounted(() => {
  fetchWorkouts()
})
</script>
