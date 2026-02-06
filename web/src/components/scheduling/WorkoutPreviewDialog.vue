<template>
  <v-dialog v-model="dialogVisible" max-width="600" scrollable>
    <v-card>
      <!-- Header -->
      <v-card-title
        class="d-flex align-center py-3 px-4"
        style="background-color: rgb(var(--v-theme-primary)); color: white"
      >
        <v-icon color="white" class="mr-2">mdi-dumbbell</v-icon>
        <span class="text-h6 font-weight-bold">{{ workout?.name || 'Workout Details' }}</span>
        <v-spacer />
        <v-btn icon variant="text" color="white" size="small" @click="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <!-- Loading State -->
      <v-card-text v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="primary" size="48" />
        <p class="text-body-2 mt-3 text-medium-emphasis">Loading workout details...</p>
      </v-card-text>

      <!-- Error State -->
      <v-card-text v-else-if="error" class="text-center py-8">
        <v-icon size="64" color="error">mdi-alert-circle</v-icon>
        <p class="text-body-1 mt-3 text-error">{{ error }}</p>
      </v-card-text>

      <!-- Workout Details -->
      <v-card-text v-else-if="workout" class="pa-4">
        <!-- Intro/Warmup -->
        <div v-if="workout.intro_warmup" class="mb-4">
          <div class="d-flex align-center mb-2">
            <v-icon color="primary" size="small" class="mr-2">mdi-run</v-icon>
            <span class="text-body-1 font-weight-bold">Warmup / Intro</span>
          </div>
          <div class="text-body-2 bg-grey-lighten-4 pa-3 rounded-lg">
            {{ workout.intro_warmup }}
          </div>
        </div>

        <!-- Movements Section -->
        <div v-if="workout.movements && workout.movements.length > 0" class="mb-4">
          <div class="d-flex align-center mb-2">
            <v-icon color="primary" size="small" class="mr-2">mdi-weight-lifter</v-icon>
            <span class="text-body-1 font-weight-bold">Movements ({{ workout.movements.length }})</span>
          </div>
          <div class="movements-list">
            <v-card
              v-for="(movement, index) in workout.movements"
              :key="index"
              elevation="0"
              class="mb-2 pa-3"
              color="grey-lighten-4"
              rounded="lg"
            >
              <div class="d-flex align-center">
                <div class="flex-grow-1">
                  <span class="text-body-2 font-weight-medium">
                    {{ movement.movement?.name || movement.name || 'Unknown Movement' }}
                  </span>
                  <div v-if="movement.notes" class="text-caption text-medium-emphasis mt-1">
                    {{ movement.notes }}
                  </div>
                </div>
                <v-chip
                  v-if="movement.movement?.type"
                  :color="getMovementTypeColor(movement.movement.type)"
                  size="x-small"
                  variant="tonal"
                >
                  {{ movement.movement.type }}
                </v-chip>
              </div>
            </v-card>
          </div>
        </div>

        <!-- WODs Section -->
        <div v-if="workout.wods && workout.wods.length > 0" class="mb-4">
          <div class="d-flex align-center mb-2">
            <v-icon color="warning" size="small" class="mr-2">mdi-fire</v-icon>
            <span class="text-body-1 font-weight-bold">WODs ({{ workout.wods.length }})</span>
          </div>
          <div class="wods-list">
            <v-card
              v-for="wod in workout.wods"
              :key="wod.id"
              elevation="0"
              class="mb-2 pa-3"
              style="background-color: #fff8e1; border: 1px solid #ffc107"
              rounded="lg"
            >
              <div class="font-weight-medium text-body-2 mb-1">
                {{ wod.wod?.name || wod.name || 'Custom WOD' }}
              </div>
              <div v-if="wod.wod?.wod_type" class="text-caption text-medium-emphasis">
                Type: {{ wod.wod.wod_type }}
              </div>
              <div v-if="wod.wod?.description" class="text-caption mt-1">
                {{ wod.wod.description }}
              </div>
            </v-card>
          </div>
        </div>

        <!-- Notes -->
        <div v-if="workout.notes" class="mb-4">
          <div class="d-flex align-center mb-2">
            <v-icon color="primary" size="small" class="mr-2">mdi-note-text</v-icon>
            <span class="text-body-1 font-weight-bold">Notes</span>
          </div>
          <div class="text-body-2 bg-grey-lighten-4 pa-3 rounded-lg">
            {{ workout.notes }}
          </div>
        </div>

        <!-- Empty State -->
        <div
          v-if="!workout.intro_warmup && (!workout.movements || workout.movements.length === 0) && (!workout.wods || workout.wods.length === 0) && !workout.notes"
          class="text-center py-4"
        >
          <v-icon size="48" color="grey-lighten-1">mdi-clipboard-text-outline</v-icon>
          <p class="text-body-2 mt-2 text-medium-emphasis">No details available for this workout</p>
        </div>
      </v-card-text>

      <!-- Actions -->
      <v-divider />
      <v-card-actions class="pa-3">
        <v-spacer />
        <v-btn variant="flat" color="primary" @click="close">
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import axios from '@/utils/axios'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  workoutId: {
    type: [Number, String],
    default: null
  },
  // Optional: pass workout directly to avoid re-fetching
  preloadedWorkout: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'close'])

// State
const workout = ref(null)
const loading = ref(false)
const error = ref(null)

// Computed
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// Methods
function getMovementTypeColor(type) {
  const colors = {
    weightlifting: 'cyan',
    gymnastics: 'purple',
    cardio: 'red',
    bodyweight: 'green'
  }
  return colors[type?.toLowerCase()] || 'grey'
}

async function fetchWorkout() {
  if (!props.workoutId) return

  // If preloaded workout is provided, use it
  if (props.preloadedWorkout) {
    workout.value = props.preloadedWorkout
    return
  }

  // Otherwise fetch from API
  loading.value = true
  error.value = null
  try {
    const response = await axios.get(`/api/templates/${props.workoutId}`)
    workout.value = response.data
  } catch (err) {
    console.error('Failed to fetch workout:', err)
    error.value = err.response?.data?.message || 'Failed to load workout details'
    workout.value = null
  } finally {
    loading.value = false
  }
}

function close() {
  dialogVisible.value = false
  emit('close')
}

// Watch for dialog open and workoutId change
watch(() => props.modelValue, (newVal) => {
  if (newVal && props.workoutId) {
    fetchWorkout()
  }
})

watch(() => props.workoutId, () => {
  if (props.modelValue && props.workoutId) {
    fetchWorkout()
  }
})
</script>

<style scoped>
.movements-list,
.wods-list {
  max-height: 300px;
  overflow-y: auto;
}
</style>
