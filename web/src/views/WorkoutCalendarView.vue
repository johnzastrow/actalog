<template>
  <div class="mobile-view-wrapper">
    <!-- Header -->
    <div style="background: #2c3e50; padding: 16px; position: fixed; top: 56px; left: 0; right: 0; z-index: 5">
      <h2 style="color: white; margin: 0; font-size: 1.25rem">Workout Calendar</h2>
    </div>

    <!-- Content (with padding for fixed header) -->
    <div style="padding: 72px 16px 16px 16px; overflow-y: auto">
      <!-- Calendar Component -->
      <workout-calendar :workout-dates="workoutDates" @day-selected="handleDaySelected" />

      <!-- Instructions -->
      <v-card elevation="0" rounded="lg" class="mt-4 pa-4 text-center" bg-color="surface">
        <v-icon size="48" color="primary" class="mb-2">mdi-gesture-tap</v-icon>
        <p class="text-body-2 text-medium-emphasis mb-0">
          Tap any date to view workouts for that day
        </p>
      </v-card>
    </div>

    <!-- Date Workout Detail Dialog -->
    <DateWorkoutDetailDialog
      v-model="showDateDialog"
      :date="selectedDate"
      :preloaded-workouts="workouts"
    />

  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'
import WorkoutCalendar from '@/components/WorkoutCalendar.vue'
import DateWorkoutDetailDialog from '@/components/DateWorkoutDetailDialog.vue'

const workouts = ref([])
const workoutDates = ref([])
const selectedDate = ref(null)
const showDateDialog = ref(false)

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
  showDateDialog.value = true
}

onMounted(() => {
  fetchWorkouts()
})
</script>
