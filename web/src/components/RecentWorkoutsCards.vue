<template>
  <v-card elevation="0" rounded="lg" class="pa-3" style="background: white">
    <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">
      Workouts in last 7 Days
    </h2>

    <!-- Loading State -->
    <div v-if="loading" class="text-center pa-4">
      <v-progress-circular indeterminate color="#00bcd4" size="48" />
    </div>

    <!-- Empty State -->
    <div v-else-if="groupedWorkouts.length === 0" class="text-center pa-4">
      <v-icon size="48" color="#ccc">mdi-calendar-blank</v-icon>
      <p class="text-body-2 mt-2" style="color: #666">No workouts in the last 7 days</p>
      <v-btn
        color="#00bcd4"
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
        <div class="text-caption font-weight-bold mb-2" style="color: #666">
          {{ formatDateHeader(group.date) }}
        </div>

        <!-- Workout Cards for this date -->
        <v-card
          v-for="workout in group.workouts"
          :key="workout.id"
          elevation="1"
          rounded="lg"
          class="pa-3 mb-2 workout-card"
          @click="viewWorkout(workout)"
          style="cursor: pointer; border: 1px solid #e0e0e0"
        >
          <div class="d-flex align-center">
            <div class="flex-grow-1">
              <!-- Workout Name/Type -->
              <div class="d-flex align-center mb-1">
                <span class="text-body-2 font-weight-bold" style="color: #1a1a1a">
                  {{ workout.workout_name || 'Custom Workout' }}
                </span>
                <v-icon
                  v-if="workout.is_pr"
                  color="#ffc107"
                  size="small"
                  class="ml-2"
                >
                  mdi-star
                </v-icon>
              </div>

              <!-- Movement Details -->
              <div v-if="workout.movements && workout.movements.length > 0" class="text-caption" style="color: #666">
                {{ formatMovements(workout.movements) }}
              </div>

              <!-- Time if available -->
              <div v-if="workout.total_time" class="text-caption mt-1" style="color: #00bcd4">
                <v-icon size="x-small" color="#00bcd4">mdi-clock-outline</v-icon>
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
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric'
  })
}

function formatRelativeDate(dateString) {
  const date = new Date(dateString)
  const today = new Date()
  const diffTime = today - date
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return 'Today'
  if (diffDays === 1) return 'Yesterday'
  if (diffDays <= 7) return `${diffDays} days ago`
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

function formatMovements(movements) {
  if (!movements || movements.length === 0) return ''

  // Show first 2 movements
  const first = movements.slice(0, 2).map(m => m.name).join(', ')
  if (movements.length > 2) {
    return `${first}, +${movements.length - 2} more`
  }
  return first
}

function formatTime(seconds) {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

function viewWorkout(workout) {
  router.push(`/workouts?id=${workout.id}`)
}
</script>

<style scoped>
.workout-card {
  transition: all 0.2s ease;
}

.workout-card:hover {
  border-color: #00bcd4 !important;
  box-shadow: 0 2px 8px rgba(0, 188, 212, 0.2) !important;
}
</style>
