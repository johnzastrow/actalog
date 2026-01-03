<template>
  <v-card elevation="0" rounded="lg" class="pa-3" bg-color="surface">
    <h2 class="text-body-1 font-weight-bold mb-3" style="color: rgb(var(--v-theme-on-surface))">
      Workouts in last 7 Days
    </h2>

    <!-- Loading State -->
    <div v-if="loading" class="text-center pa-4">
      <v-progress-circular indeterminate color="primary" size="48" />
    </div>

    <!-- Empty State -->
    <div v-else-if="groupedWorkouts.length === 0" class="text-center pa-4">
      <v-icon size="48" color="#ccc">mdi-calendar-blank</v-icon>
      <p class="text-body-2 mt-2 text-medium-emphasis">No workouts in the last 7 days</p>
      <v-btn
        color="primary"
        size="small"
        to="/workouts/log"
        rounded="lg"
        class="mt-2 text-none font-weight-600"
        style="color: white"
      >
        Log a Workout
      </v-btn>
    </div>

    <!-- Workouts by Date -->
    <div v-else>
      <div
        v-for="group in groupedWorkouts"
        :key="group.date"
        class="mb-3"
      >
        <!-- Date Header -->
        <div class="text-caption font-weight-bold mb-2 text-medium-emphasis">
          {{ formatDateHeader(group.date) }}
        </div>

        <!-- Workout Cards for this date -->
        <v-card
          v-for="workout in group.workouts"
          :key="workout.id"
          elevation="1"
          rounded="lg"
          class="pa-3 mb-2 workout-card"
          style="cursor: pointer; border: 1px solid #e0e0e0"
          @click="viewWorkout(workout)"
        >
          <div class="d-flex align-center">
            <div class="flex-grow-1">
              <!-- Workout Name/Type -->
              <div class="d-flex align-center mb-1">
                <span class="text-body-2 font-weight-bold" style="color: rgb(var(--v-theme-on-surface))">
                  {{ workout.workout_name || 'Custom Workout' }}
                </span>
                <v-chip
                  v-if="hasPR(workout)"
                  color="primary"
                  size="x-small"
                  class="ml-2"
                  style="height: 18px; font-size: 10px"
                >
                  <v-icon size="x-small" class="mr-1">mdi-trophy</v-icon>
                  PR
                </v-chip>
              </div>

              <!-- Movement Details with PR indicators -->
              <div v-if="workout.movements && workout.movements.length > 0" class="text-caption text-medium-emphasis">
                <div v-for="(movement, index) in workout.movements.slice(0, 3)" :key="index" class="d-flex align-center">
                  <v-icon v-if="movement.is_pr" color="primary" size="x-small" class="mr-1">mdi-trophy</v-icon>
                  <span>{{ movement.movement?.name || 'Movement' }}</span>
                  <span v-if="movement.weight" class="ml-1">- {{ movement.weight }}lb</span>
                  <span v-if="movement.reps" class="ml-1">x{{ movement.reps }}</span>
                </div>
                <div v-if="workout.movements.length > 3" class="text-caption mt-1" style="color: #999">
                  +{{ workout.movements.length - 3 }} more
                </div>
              </div>

              <!-- Time if available -->
              <div v-if="workout.total_time" class="text-caption mt-1" style="color: rgb(var(--v-theme-primary))">
                <v-icon size="x-small" color="primary">mdi-clock-outline</v-icon>
                {{ formatTime(workout.total_time) }}
              </div>
            </div>

            <!-- Date badge -->
            <div class="text-caption font-weight-medium" style="color: #999">
              {{ formatRelativeDate(workout.workout_date) }}
            </div>
          </div>
        </v-card>
      </div>
    </div>
  </v-card>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSettingsStore } from '@/stores/settings'
import { formatDateInTimezone, getTodayInTimezone } from '@/utils/timezone'

const settingsStore = useSettingsStore()

const props = defineProps({
  workouts: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const router = useRouter()

// Group workouts by date
const groupedWorkouts = computed(() => {
  const groups = {}

  props.workouts.forEach(workout => {
    const date = new Date(workout.workout_date).toDateString()
    if (!groups[date]) {
      groups[date] = {
        date: date,
        workouts: []
      }
    }
    groups[date].workouts.push(workout)
  })

  return Object.values(groups).sort((a, b) => {
    return new Date(b.date) - new Date(a.date)
  })
})

function formatDateHeader(dateString) {
  const tz = settingsStore.timezone
  return formatDateInTimezone(dateString, tz, 'EEE, MMM d')
}

function formatRelativeDate(dateString) {
  const tz = settingsStore.timezone
  const datePart = dateString.split('T')[0]
  const todayStr = getTodayInTimezone(tz)

  // Calculate yesterday
  const todayDate = new Date(todayStr)
  const yesterdayDate = new Date(todayDate)
  yesterdayDate.setDate(yesterdayDate.getDate() - 1)
  const yesterdayStr = yesterdayDate.toISOString().split('T')[0]

  if (datePart === todayStr) return 'Today'
  if (datePart === yesterdayStr) return 'Yesterday'

  // Calculate days ago
  const workoutDate = new Date(datePart)
  const today = new Date(todayStr)
  const diffTime = today - workoutDate
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))

  if (diffDays <= 7) return `${diffDays} days ago`
  return formatDateInTimezone(dateString, tz, 'MMM d')
}

function hasPR(workout) {
  return workout.movements?.some(m => m.is_pr) || false
}

function formatTime(seconds) {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

function viewWorkout(workout) {
  router.push(`/workouts/${workout.id}`)
}
</script>

<style scoped>
.workout-card {
  transition: all 0.2s ease;
}

.workout-card:hover {
  border-color: rgb(var(--v-theme-primary)) !important;
  box-shadow: 0 2px 8px rgba(0, 188, 212, 0.2) !important;
}
</style>
