<template>
  <v-dialog v-model="dialogVisible" max-width="600" scrollable>
    <v-card>
      <!-- Header -->
      <v-card-title
        class="d-flex align-center py-3 px-4"
        style="background-color: rgb(var(--v-theme-primary)); color: white;"
      >
        <v-icon color="white" class="mr-2">mdi-calendar-clock</v-icon>
        <span class="text-h6 font-weight-bold">{{ session?.name || 'Session Details' }}</span>
        <v-spacer />
        <v-btn icon variant="text" color="white" size="small" @click="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <!-- Loading State -->
      <v-card-text v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="primary" size="48" />
        <p class="text-body-2 mt-3 text-medium-emphasis">Loading session details...</p>
      </v-card-text>

      <!-- Error State -->
      <v-card-text v-else-if="error" class="py-4">
        <v-alert type="error" variant="tonal">{{ error }}</v-alert>
      </v-card-text>

      <!-- Content -->
      <v-card-text v-else-if="session" class="pa-4">
        <!-- Session Info Section -->
        <div class="mb-4">
          <!-- Time and Location -->
          <div class="d-flex align-center mb-2">
            <v-icon color="primary" size="small" class="mr-2">mdi-clock-outline</v-icon>
            <span class="text-body-2">
              {{ formatDateTime(session.start_time) }} - {{ formatTime(session.end_time) }}
            </span>
          </div>
          <div v-if="session.location_name" class="d-flex align-center mb-2">
            <v-icon color="primary" size="small" class="mr-2">mdi-map-marker</v-icon>
            <span class="text-body-2">{{ session.location_name }}</span>
          </div>

          <!-- Coaches -->
          <div v-if="session.coaches && session.coaches.length > 0" class="d-flex align-center mb-2">
            <v-icon color="primary" size="small" class="mr-2">mdi-account-tie</v-icon>
            <span class="text-body-2">
              {{ session.coaches.map(c => c.user_name || c.user_email).join(', ') }}
            </span>
          </div>

          <!-- Capacity -->
          <div class="d-flex align-center mb-2">
            <v-icon color="primary" size="small" class="mr-2">mdi-account-multiple</v-icon>
            <span class="text-body-2">
              {{ session.reservation_count }}/{{ session.capacity }} spots
              <span v-if="session.available_spots > 0" class="text-success ml-1">
                ({{ session.available_spots }} available)
              </span>
              <span v-else class="text-error ml-1">(Full)</span>
            </span>
          </div>

          <!-- Status chips -->
          <div class="d-flex flex-wrap gap-2 mt-3">
            <v-chip v-if="session.status === 'cancelled'" color="error" size="small">
              Cancelled
            </v-chip>
            <v-chip v-else-if="session.status === 'completed'" color="success" size="small">
              Completed
            </v-chip>
            <v-chip v-else-if="session.current_user_status === 'reserved'" color="primary" size="small">
              You're Reserved
            </v-chip>
            <v-chip v-else-if="session.current_user_status === 'checked_in'" color="success" size="small">
              Checked In
            </v-chip>
          </div>
        </div>

        <v-divider class="my-4" />

        <!-- Workout Details Section -->
        <template v-if="session.workout">
          <div class="d-flex align-center mb-3">
            <v-icon color="warning" size="small" class="mr-2">mdi-dumbbell</v-icon>
            <h3 class="text-subtitle-1 font-weight-bold">Workout: {{ session.workout.name }}</h3>
          </div>

          <!-- Intro & Warmup -->
          <div v-if="session.workout.intro_warmup" class="mb-4">
            <p class="text-caption font-weight-bold mb-1 text-medium-emphasis">
              <v-icon size="x-small" class="mr-1">mdi-flag-checkered</v-icon>
              Intro & Warmup
            </p>
            <div class="text-body-2 pa-2 rounded" style="background-color: rgba(var(--v-theme-on-surface), 0.05);">
              <markdown-renderer :content="session.workout.intro_warmup" />
            </div>
          </div>

          <!-- Movements -->
          <div v-if="session.workout.movements && session.workout.movements.length > 0" class="mb-4">
            <p class="text-caption font-weight-bold mb-2 text-medium-emphasis">
              <v-icon size="x-small" class="mr-1">mdi-weight-lifter</v-icon>
              Movements ({{ session.workout.movements.length }})
            </p>
            <v-list density="compact" bg-color="transparent" class="py-0">
              <v-list-item v-for="m in session.workout.movements" :key="m.id" class="px-0 mb-2">
                <template #prepend>
                  <v-icon size="small" color="primary" class="mt-1">mdi-weight-lifter</v-icon>
                </template>
                <div>
                  <v-list-item-title class="text-body-2 font-weight-medium">
                    {{ m.movement?.name || 'Movement' }}
                  </v-list-item-title>
                  <!-- Instructions -->
                  <div v-if="m.instructions" class="mt-1">
                    <div class="text-caption text-info">
                      <markdown-renderer :content="m.instructions" />
                    </div>
                  </div>
                  <!-- Movement specifics -->
                  <div class="d-flex flex-wrap gap-1 mt-1">
                    <v-chip v-if="m.sets" size="x-small" color="primary" variant="tonal">
                      {{ m.sets }} sets
                    </v-chip>
                    <v-chip v-if="m.reps" size="x-small" color="secondary" variant="tonal">
                      {{ m.reps }} reps
                    </v-chip>
                    <v-chip v-if="m.weight" size="x-small" color="warning" variant="tonal">
                      {{ m.weight }} {{ m.weight_unit || 'lbs' }}
                    </v-chip>
                    <v-chip v-if="m.time" size="x-small" color="info" variant="tonal">
                      {{ formatDuration(m.time) }}
                    </v-chip>
                    <v-chip v-if="m.distance" size="x-small" color="success" variant="tonal">
                      {{ m.distance }} {{ m.distance_unit || 'm' }}
                    </v-chip>
                  </div>
                  <!-- Notes -->
                  <p v-if="m.notes" class="text-caption text-medium-emphasis mt-1 mb-0">
                    {{ m.notes }}
                  </p>
                </div>
              </v-list-item>
            </v-list>
          </div>

          <!-- WODs -->
          <div v-if="session.workout.wods && session.workout.wods.length > 0" class="mb-4">
            <p class="text-caption font-weight-bold mb-2 text-medium-emphasis">
              <v-icon size="x-small" class="mr-1">mdi-fire</v-icon>
              WODs ({{ session.workout.wods.length }})
            </p>
            <v-list density="compact" bg-color="transparent" class="py-0">
              <v-list-item v-for="w in session.workout.wods" :key="w.id" class="px-0 mb-2">
                <template #prepend>
                  <v-icon size="small" color="#ff5722" class="mt-1">mdi-fire</v-icon>
                </template>
                <div>
                  <v-list-item-title class="text-body-2 font-weight-medium">
                    {{ w.wod_name || w.wod?.name || 'WOD' }}
                  </v-list-item-title>
                  <!-- Instructions -->
                  <div v-if="w.instructions" class="mt-1">
                    <div class="text-caption text-info">
                      <markdown-renderer :content="w.instructions" />
                    </div>
                  </div>
                  <!-- WOD specifics -->
                  <div class="d-flex flex-wrap gap-1 mt-1">
                    <v-chip v-if="w.wod_type" size="x-small" color="secondary" variant="tonal">
                      {{ w.wod_type }}
                    </v-chip>
                    <v-chip v-if="w.wod_regime" size="x-small" color="primary" variant="tonal">
                      {{ w.wod_regime }}
                    </v-chip>
                    <v-chip v-if="w.wod_score_type" size="x-small" color="success" variant="tonal">
                      {{ w.wod_score_type }}
                    </v-chip>
                  </div>
                  <!-- Description -->
                  <p v-if="w.wod_description" class="text-caption text-medium-emphasis mt-1 mb-0" style="white-space: pre-wrap;">
                    {{ w.wod_description }}
                  </p>
                  <!-- Notes -->
                  <p v-if="w.notes" class="text-caption text-medium-emphasis mt-1 mb-0">
                    <strong>Notes:</strong> {{ w.notes }}
                  </p>
                </div>
              </v-list-item>
            </v-list>
          </div>

          <!-- Workout Notes -->
          <div v-if="session.workout.notes" class="mb-4">
            <p class="text-caption font-weight-bold mb-1 text-medium-emphasis">
              <v-icon size="x-small" class="mr-1">mdi-note-text</v-icon>
              Notes
            </p>
            <div class="text-body-2 pa-2 rounded" style="background-color: rgba(var(--v-theme-on-surface), 0.05);">
              <markdown-renderer :content="session.workout.notes" />
            </div>
          </div>
        </template>

        <!-- No Workout Assigned -->
        <template v-else-if="!session.workout_id">
          <v-alert type="info" variant="tonal" density="compact">
            No workout assigned to this session.
          </v-alert>
        </template>

        <!-- Workout Name Only (no full details loaded) -->
        <template v-else-if="session.workout_name">
          <div class="d-flex align-center mb-3">
            <v-icon color="warning" size="small" class="mr-2">mdi-dumbbell</v-icon>
            <h3 class="text-subtitle-1 font-weight-bold">Workout: {{ session.workout_name }}</h3>
          </div>
        </template>
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
          v-if="session?.workout_id && session?.status === 'scheduled'"
          color="warning"
          variant="flat"
          @click="quickLog"
        >
          <v-icon start>mdi-lightning-bolt</v-icon>
          Quick Log
        </v-btn>
        <v-btn
          v-if="canReserve"
          color="primary"
          variant="flat"
          @click="makeReservation"
        >
          <v-icon start>mdi-calendar-plus</v-icon>
          Reserve
        </v-btn>
        <v-btn
          v-if="canCancel"
          color="error"
          variant="outlined"
          @click="cancelReservation"
        >
          <v-icon start>mdi-calendar-remove</v-icon>
          Cancel
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  sessionId: {
    type: [Number, String],
    default: null
  },
  gymId: {
    type: [Number, String],
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'reserved', 'cancelled'])

const router = useRouter()

// State
const session = ref(null)
const loading = ref(false)
const error = ref('')

// Computed
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const canReserve = computed(() => {
  if (!session.value) return false
  return !session.value.current_user_status &&
         session.value.status === 'scheduled' &&
         session.value.available_spots > 0
})

const canCancel = computed(() => {
  if (!session.value) return false
  return session.value.current_user_status === 'reserved'
})

// Methods
function formatDateTime(dateTimeStr) {
  if (!dateTimeStr) return ''
  const date = new Date(dateTimeStr)
  return date.toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  })
}

function formatTime(dateTimeStr) {
  if (!dateTimeStr) return ''
  const date = new Date(dateTimeStr)
  return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
}

function formatDuration(seconds) {
  if (!seconds) return ''
  if (seconds < 60) return `${seconds}s`
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return secs > 0 ? `${mins}m ${secs}s` : `${mins}m`
}

async function fetchSession() {
  if (!props.sessionId || !props.gymId) return

  loading.value = true
  error.value = ''
  session.value = null

  try {
    const response = await axios.get(`/api/gyms/${props.gymId}/sessions/${props.sessionId}`)
    session.value = response.data
  } catch (err) {
    console.error('Failed to fetch session:', err)
    error.value = err.response?.data?.error || 'Failed to load session details'
  } finally {
    loading.value = false
  }
}

function close() {
  dialogVisible.value = false
}

function quickLog() {
  if (session.value?.workout_id) {
    router.push(`/workouts/log?template=${session.value.workout_id}`)
    close()
  }
}

async function makeReservation() {
  if (!session.value) return

  loading.value = true
  error.value = ''

  try {
    await axios.post(`/api/sessions/${session.value.id}/reserve`)
    emit('reserved', session.value)
    close()
  } catch (err) {
    console.error('Failed to make reservation:', err)
    error.value = err.response?.data?.error || 'Failed to make reservation'
  } finally {
    loading.value = false
  }
}

async function cancelReservation() {
  if (!session.value) return

  loading.value = true
  error.value = ''

  try {
    await axios.delete(`/api/sessions/${session.value.id}/reserve`)
    emit('cancelled', session.value)
    close()
  } catch (err) {
    console.error('Failed to cancel reservation:', err)
    error.value = err.response?.data?.error || 'Failed to cancel reservation'
  } finally {
    loading.value = false
  }
}

// Watch for dialog open
watch(() => props.modelValue, (newVal) => {
  if (newVal && props.sessionId && props.gymId) {
    fetchSession()
  }
})

// Watch for sessionId change while dialog is open
watch(() => props.sessionId, (newVal) => {
  if (props.modelValue && newVal && props.gymId) {
    fetchSession()
  }
})
</script>

<style scoped>
.gap-1 {
  gap: 4px;
}
.gap-2 {
  gap: 8px;
}
</style>
