<template>
  <v-dialog v-model="dialogVisible" max-width="600" scrollable>
    <v-card>
      <!-- Header -->
      <v-card-title
        class="d-flex align-center py-3 px-4"
        style="background-color: rgb(var(--v-theme-primary)); color: white"
      >
        <v-icon color="white" class="mr-2">mdi-calendar</v-icon>
        <span class="text-h6 font-weight-bold">{{ formattedDate }}</span>
        <v-spacer />
        <v-btn icon variant="text" color="white" size="small" @click="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <!-- Loading State -->
      <v-card-text v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="primary" size="48" />
        <p class="text-body-2 mt-3 text-medium-emphasis">Loading workouts...</p>
      </v-card-text>

      <!-- Empty State -->
      <v-card-text v-else-if="workouts.length === 0" class="text-center py-8">
        <v-icon size="64" color="surface-variant">mdi-calendar-blank</v-icon>
        <p class="text-h6 mt-3 text-medium-emphasis">No workouts on this day</p>
        <p class="text-body-2 text-disabled">Click QuickLog to add a workout for this date</p>
      </v-card-text>

      <!-- Workouts List -->
      <v-card-text v-else class="pa-0">
        <v-list>
          <v-list-item
            v-for="workout in workouts"
            :key="workout.id"
            class="px-4 py-3"
            :border="true"
          >
            <template #prepend>
              <v-icon color="primary" class="mr-3">mdi-clipboard-text</v-icon>
            </template>

            <v-list-item-title class="text-body-1 font-weight-bold mb-1">
              {{ workout.workout_name || workout.template_name || 'Workout' }}
            </v-list-item-title>

            <v-list-item-subtitle>
              <!-- Movements Summary -->
              <div v-if="workout.performance_movements?.length" class="mb-1">
                <v-icon size="x-small" color="primary" class="mr-1">mdi-dumbbell</v-icon>
                <span class="text-caption">
                  {{ workout.performance_movements.length }} movement{{ workout.performance_movements.length > 1 ? 's' : '' }}
                </span>
                <span v-if="hasMovementPR(workout)" class="ml-1">
                  <v-icon size="x-small" color="warning">mdi-trophy</v-icon>
                </span>
              </div>

              <!-- WODs Summary -->
              <div v-if="workout.performance_wods?.length" class="mb-1">
                <v-icon size="x-small" color="#ff5722" class="mr-1">mdi-fire</v-icon>
                <span class="text-caption">
                  {{ workout.performance_wods.length }} WOD{{ workout.performance_wods.length > 1 ? 's' : '' }}
                </span>
                <span v-if="hasWodPR(workout)" class="ml-1">
                  <v-icon size="x-small" color="warning">mdi-trophy</v-icon>
                </span>
              </div>

              <!-- Notes -->
              <div v-if="workout.notes" class="text-caption text-disabled mt-1">
                {{ truncateNotes(workout.notes) }}
              </div>
            </v-list-item-subtitle>

            <template #append>
              <div class="d-flex gap-1">
                <v-btn
                  icon
                  size="small"
                  variant="text"
                  color="primary"
                  @click.stop="viewWorkout(workout)"
                >
                  <v-icon>mdi-eye</v-icon>
                  <v-tooltip activator="parent" location="top">View Details</v-tooltip>
                </v-btn>
                <v-btn
                  icon
                  size="small"
                  variant="text"
                  color="primary"
                  @click.stop="editWorkout(workout)"
                >
                  <v-icon>mdi-pencil</v-icon>
                  <v-tooltip activator="parent" location="top">Edit</v-tooltip>
                </v-btn>
              </div>
            </template>
          </v-list-item>
        </v-list>
      </v-card-text>

      <!-- Actions -->
      <v-divider />
      <v-card-actions class="pa-3">
        <v-btn variant="text" @click="close">
          <v-icon start>mdi-arrow-left</v-icon>
          Back
        </v-btn>
        <v-spacer />
        <v-btn
          color="primary"
          variant="flat"
          @click="quickLogForDate"
        >
          <v-icon start>mdi-lightning-bolt</v-icon>
          QuickLog
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  date: {
    type: Date,
    default: null
  },
  // Optional: pass workouts directly to avoid re-fetching
  preloadedWorkouts: {
    type: Array,
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'close', 'quicklog', 'view', 'edit'])

const router = useRouter()

// State
const workouts = ref([])
const loading = ref(false)

// Computed
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const formattedDate = computed(() => {
  if (!props.date) return ''
  return props.date.toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
    year: 'numeric'
  })
})

const dateString = computed(() => {
  if (!props.date) return ''
  const year = props.date.getFullYear()
  const month = String(props.date.getMonth() + 1).padStart(2, '0')
  const day = String(props.date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
})

// Methods
function hasMovementPR(workout) {
  return workout.performance_movements?.some(m => m.is_pr)
}

function hasWodPR(workout) {
  return workout.performance_wods?.some(w => w.is_pr)
}

function truncateNotes(notes, maxLength = 60) {
  if (!notes || notes.length <= maxLength) return notes
  return notes.substring(0, maxLength) + '...'
}

async function fetchWorkouts() {
  if (!props.date) return

  // If preloaded workouts are provided, filter them
  if (props.preloadedWorkouts) {
    workouts.value = props.preloadedWorkouts.filter(w => {
      const workoutDateStr = w.workout_date?.split('T')[0]
      return workoutDateStr === dateString.value
    })
    return
  }

  // Otherwise fetch from API
  loading.value = true
  try {
    const response = await axios.get('/api/workouts', {
      params: { date: dateString.value }
    })
    const data = response.data
    workouts.value = Array.isArray(data) ? data : (data.workouts || [])
  } catch (err) {
    console.error('Failed to fetch workouts for date:', err)
    workouts.value = []
  } finally {
    loading.value = false
  }
}

function close() {
  dialogVisible.value = false
  emit('close')
}

function viewWorkout(workout) {
  emit('view', workout)
  router.push(`/workouts/${workout.id}`)
  close()
}

function editWorkout(workout) {
  emit('edit', workout)
  router.push(`/workouts/log?edit=${workout.id}`)
  close()
}

function quickLogForDate() {
  emit('quicklog', { date: props.date, dateString: dateString.value })
  router.push(`/workouts/log?date=${dateString.value}`)
  close()
}

// Watch for dialog open and date change
watch(() => props.modelValue, (newVal) => {
  if (newVal && props.date) {
    fetchWorkouts()
  }
})

watch(() => props.date, () => {
  if (props.modelValue && props.date) {
    fetchWorkouts()
  }
})
</script>
