<template>
  <div class="workout-assignment-tab">
    <div class="section-header d-flex align-center justify-space-between mb-4">
      <div class="section-title">
        <v-icon size="small" class="mr-2">mdi-dumbbell</v-icon>
        Assign Workout to Sessions
      </div>
      <v-btn
        color="primary"
        variant="tonal"
        size="small"
        prepend-icon="mdi-plus"
        rounded="lg"
        @click="createNewWorkout"
      >
        New Workout
      </v-btn>
    </div>

    <!-- Two Column Layout -->
    <v-row>
      <!-- Left Panel: Workout Selection -->
      <v-col cols="12" md="5">
        <div class="panel-card">
          <WorkoutSearchPanel
            :workouts="workouts"
            :selected-id="selectedWorkoutId"
            @select="onWorkoutSelect"
            @preview="openWorkoutPreview"
          />
        </div>
      </v-col>

      <!-- Right Panel: Session Selection -->
      <v-col cols="12" md="7">
        <div class="panel-card">
          <SessionSelectionPanel
            ref="sessionPanel"
            :sessions="sessions"
            :templates="templates"
            :coaches="coaches"
            :locations="locations"
            @selection-change="onSessionSelectionChange"
          />
        </div>
      </v-col>
    </v-row>

    <!-- Action Bar -->
    <div class="action-bar mt-4 pa-3">
      <div class="d-flex align-center justify-space-between">
        <div class="selection-summary">
          <template v-if="selectedWorkoutId && selectedSessionIds.length > 0">
            <v-chip color="primary" variant="tonal" class="mr-2">
              <v-icon start size="small">mdi-dumbbell</v-icon>
              {{ selectedWorkoutName }}
            </v-chip>
            <v-icon size="small" class="mx-2">mdi-arrow-right</v-icon>
            <v-chip color="info" variant="tonal">
              <v-icon start size="small">mdi-calendar</v-icon>
              {{ selectedSessionIds.length }} session{{ selectedSessionIds.length !== 1 ? 's' : '' }}
            </v-chip>
          </template>
          <span v-else class="text-medium-emphasis">
            Select a workout and sessions to assign
          </span>
        </div>
        <v-btn
          color="primary"
          size="large"
          rounded="lg"
          :disabled="!canApply"
          :loading="applying"
          @click="showConfirmDialog = true"
        >
          <v-icon start>mdi-check</v-icon>
          Apply Workout to {{ selectedSessionIds.length }} Session{{ selectedSessionIds.length !== 1 ? 's' : '' }}
        </v-btn>
      </div>
    </div>

    <!-- Confirmation Dialog -->
    <v-dialog v-model="showConfirmDialog" max-width="500">
      <v-card>
        <v-card-title class="d-flex align-center">
          <v-icon color="primary" class="mr-2">mdi-dumbbell</v-icon>
          Confirm Workout Assignment
        </v-card-title>
        <v-card-text>
          <p>
            You are about to assign the workout
            <strong>{{ selectedWorkoutName }}</strong>
            to <strong>{{ selectedSessionIds.length }}</strong> session{{ selectedSessionIds.length !== 1 ? 's' : '' }}.
          </p>
          <p class="text-caption text-medium-emphasis mt-2">
            This will replace any existing workout assignments on these sessions.
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="showConfirmDialog = false">Cancel</v-btn>
          <v-btn
            color="primary"
            variant="flat"
            :loading="applying"
            @click="applyWorkout"
          >
            Confirm
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Workout Preview Dialog -->
    <DetailViewDialog
      v-model="workoutPreviewDialog"
      type="template"
      :id="previewWorkoutId"
      :handle-quick-log-externally="true"
    />

    <!-- Snackbar -->
    <v-snackbar
      v-model="snackbar"
      :color="snackbarColor"
      :timeout="4000"
      location="bottom"
      :style="{ marginBottom: '80px' }"
    >
      {{ snackbarText }}
      <template #actions>
        <v-btn variant="text" @click="snackbar = false">Close</v-btn>
      </template>
    </v-snackbar>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'
import WorkoutSearchPanel from './WorkoutSearchPanel.vue'
import SessionSelectionPanel from './SessionSelectionPanel.vue'
import DetailViewDialog from '@/components/DetailViewDialog.vue'

const router = useRouter()

const props = defineProps({
  gymId: {
    type: [Number, String],
    required: true
  },
  workouts: {
    type: Array,
    default: () => []
  },
  sessions: {
    type: Array,
    default: () => []
  },
  templates: {
    type: Array,
    default: () => []
  },
  coaches: {
    type: Array,
    default: () => []
  },
  locations: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['applied'])

const sessionPanel = ref(null)
const selectedWorkoutId = ref(null)
const selectedSessionIds = ref([])
const showConfirmDialog = ref(false)
const applying = ref(false)

// Workout preview
const workoutPreviewDialog = ref(false)
const previewWorkoutId = ref(null)

// Snackbar
const snackbar = ref(false)
const snackbarText = ref('')
const snackbarColor = ref('success')

const selectedWorkoutName = computed(() => {
  if (!selectedWorkoutId.value) return ''
  const workout = props.workouts.find(w => w.id === selectedWorkoutId.value)
  return workout?.name || 'Unknown'
})

const canApply = computed(() => {
  return selectedWorkoutId.value && selectedSessionIds.value.length > 0
})

function onWorkoutSelect(id) {
  selectedWorkoutId.value = id
}

function onSessionSelectionChange(ids) {
  selectedSessionIds.value = ids
}

function openWorkoutPreview(id) {
  previewWorkoutId.value = id
  workoutPreviewDialog.value = true
}

function createNewWorkout() {
  router.push('/workouts/templates/create')
}

function showSnackbar(message, color = 'success') {
  snackbarText.value = message
  snackbarColor.value = color
  snackbar.value = true
}

async function applyWorkout() {
  if (!canApply.value) return

  applying.value = true
  try {
    const response = await axios.put(
      `/api/admin/gyms/${props.gymId}/sessions/batch-workout`,
      {
        session_ids: selectedSessionIds.value,
        workout_id: selectedWorkoutId.value
      }
    )

    const count = response.data.updated_count || 0
    showSnackbar(`Successfully assigned workout to ${count} session${count !== 1 ? 's' : ''}`)
    showConfirmDialog.value = false

    // Clear selections
    selectedSessionIds.value = []
    if (sessionPanel.value) {
      sessionPanel.value.clearSelection()
    }

    // Notify parent to refresh data
    emit('applied')
  } catch (err) {
    console.error('Failed to apply workout:', err)
    showSnackbar(err.response?.data?.error || 'Failed to apply workout', 'error')
  } finally {
    applying.value = false
  }
}
</script>

<style scoped>
.workout-assignment-tab {
  width: 100%;
}

.section-title {
  font-size: 1rem;
  font-weight: 600;
  color: #2c3e50;
  display: flex;
  align-items: center;
}

.panel-card {
  background-color: white;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  height: 100%;
  min-height: 500px;
}

.action-bar {
  background-color: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.selection-summary {
  display: flex;
  align-items: center;
}

@media (max-width: 960px) {
  .panel-card {
    min-height: auto;
    margin-bottom: 16px;
  }
}
</style>
