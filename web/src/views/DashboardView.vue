<template>
  <v-container fluid class="pa-3" style="background-color: #f5f7fa; min-height: calc(100vh - 64px)">
    <!-- Calendar Section -->
    <workout-calendar
      :workout-dates="workoutDates"
      @day-selected="onDaySelected"
      class="mb-3"
    />

    <!-- Recent Workouts Cards -->
    <recent-workouts-cards
      :workouts="last7DaysWorkouts"
      :loading="loading"
    />
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from '@/utils/axios'
import WorkoutCalendar from '@/components/WorkoutCalendar.vue'
import RecentWorkoutsCards from '@/components/RecentWorkoutsCards.vue'

const loading = ref(false)
const workouts = ref([])

// Get all workout dates for calendar
const workoutDates = computed(() => {
  return workouts.value.map(w => w.workout_date)
})

// Get workouts from last 7 days
const last7DaysWorkouts = computed(() => {
  const now = new Date()
  const sevenDaysAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)

  return workouts.value
    .filter(workout => {
      const workoutDate = new Date(workout.workout_date)
      return workoutDate >= sevenDaysAgo && workoutDate <= now
    })
    .sort((a, b) => new Date(b.workout_date) - new Date(a.workout_date))
})

// Fetch all workouts
async function fetchWorkouts() {
  loading.value = true
  try {
    const response = await axios.get('/api/workouts')
    workouts.value = response.data.workouts || []
  } catch (err) {
    console.error('Failed to fetch workouts:', err)
    // Fallback to empty array
    workouts.value = []
  } finally {
    loading.value = false
  }
}

// Handle day selected from calendar
function onDaySelected(date) {
  console.log('Selected date:', date)
  // Could navigate to workouts view filtered by this date
  // or show a dialog with workouts for this day
}

// Load data on mount
onMounted(() => {
  fetchWorkouts()
})
</script>

<style scoped>
/* Dashboard specific styles */
</style>
