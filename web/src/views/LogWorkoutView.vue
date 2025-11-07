<template>
  <v-container>
    <!-- Header -->
    <v-row>
      <v-col cols="12">
        <v-app-bar color="primary" density="compact" flat>
          <v-btn icon="mdi-arrow-left" @click="$router.back()" />
          <v-toolbar-title>Log Workout</v-toolbar-title>
          <v-btn icon="mdi-dots-vertical" />
        </v-app-bar>
      </v-col>
    </v-row>

    <v-row class="mt-2">
      <v-col cols="12">
        <!-- Error Alert -->
        <v-alert v-if="error" type="error" closable @click:close="error = null" class="mb-4">
          {{ error }}
        </v-alert>

        <v-card elevation="2" rounded="lg">
          <v-card-text class="pa-6">
            <v-form>
              <!-- Workout Date -->
              <v-text-field
                v-model="workoutDate"
                label="Workout Date"
                type="date"
                prepend-inner-icon="mdi-calendar"
                variant="outlined"
              />

              <!-- Workout Type -->
              <div class="my-4">
                <label class="text-body-2 text-medium-emphasis mb-2 d-block">
                  Workout Type
                </label>
                <v-btn-toggle
                  v-model="workoutType"
                  color="primary"
                  variant="outlined"
                  divided
                  mandatory
                >
                  <v-btn value="named" prepend-icon="mdi-format-list-bulleted">
                    Named WOD
                  </v-btn>
                  <v-btn value="custom" prepend-icon="mdi-pencil">
                    Custom
                  </v-btn>
                </v-btn-toggle>
              </div>

              <!-- Select Movement -->
              <v-select
                v-model="selectedMovement"
                label="Select Movement"
                :items="movements"
                item-title="title"
                item-value="value"
                :loading="loading"
                prepend-inner-icon="mdi-dumbbell"
                variant="outlined"
                class="mt-4"
              />

              <v-btn
                variant="outlined"
                color="accent"
                block
                class="mt-2"
                prepend-icon="mdi-plus"
              >
                Add Custom Movement
              </v-btn>

              <!-- Movement Details -->
              <v-row class="mt-4">
                <v-col cols="6">
                  <v-text-field
                    v-model="weight"
                    label="Weight (lbs)"
                    type="number"
                    prepend-inner-icon="mdi-weight"
                    variant="outlined"
                  />
                </v-col>
                <v-col cols="6">
                  <v-text-field
                    v-model="sets"
                    label="Sets"
                    type="number"
                    prepend-inner-icon="mdi-format-list-numbered"
                    variant="outlined"
                  />
                </v-col>
              </v-row>

              <v-row>
                <v-col cols="6">
                  <v-text-field
                    v-model="reps"
                    label="Reps"
                    type="number"
                    prepend-inner-icon="mdi-numeric"
                    variant="outlined"
                  />
                </v-col>
                <v-col cols="6">
                  <v-btn-toggle
                    v-model="movementType"
                    color="accent"
                    variant="outlined"
                    divided
                    mandatory
                  >
                    <v-btn value="rx">Rx</v-btn>
                    <v-btn value="scaled">Scaled</v-btn>
                  </v-btn-toggle>
                </v-col>
              </v-row>

              <!-- Notes -->
              <v-textarea
                v-model="notes"
                label="Notes (Optional)"
                placeholder="How did it feel? Any observations..."
                variant="outlined"
                rows="3"
                class="mt-4"
              />

              <!-- Action Buttons -->
              <v-row class="mt-4">
                <v-col cols="12">
                  <v-btn
                    color="primary"
                    size="large"
                    block
                    :loading="saving"
                    :disabled="!selectedMovement"
                    @click="saveWorkout"
                  >
                    <v-icon start>mdi-content-save</v-icon>
                    Save Workout
                  </v-btn>
                </v-col>
                <v-col cols="6">
                  <v-btn variant="outlined" size="large" block @click="resetForm">
                    <v-icon start>mdi-refresh</v-icon>
                    Reset
                  </v-btn>
                </v-col>
                <v-col cols="6">
                  <v-btn variant="outlined" size="large" block @click="$router.back()">
                    <v-icon start>mdi-close</v-icon>
                    Cancel
                  </v-btn>
                </v-col>
              </v-row>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'

const router = useRouter()

// Form state
const workoutDate = ref(new Date().toISOString().split('T')[0])
const workoutType = ref('named')
const selectedMovement = ref(null)
const movements = ref([])
const weight = ref(null)
const sets = ref(null)
const reps = ref(null)
const movementType = ref('rx')
const notes = ref('')

// UI state
const saving = ref(false)
const loading = ref(false)
const error = ref(null)

// Fetch available movements
async function fetchMovements() {
  loading.value = true
  try {
    const response = await axios.get('/api/movements')
    // Transform movements for v-select (value and title)
    movements.value = response.data.movements.map((m) => ({
      value: m.id,
      title: m.name,
      type: m.type,
    }))
  } catch (err) {
    console.error('Failed to fetch movements:', err)
    error.value = 'Failed to load movements'
    // Fallback to hardcoded movements for demo
    movements.value = [
      { value: 1, title: 'Back Squat', type: 'weightlifting' },
      { value: 2, title: 'Deadlift', type: 'weightlifting' },
      { value: 3, title: 'Bench Press', type: 'weightlifting' },
    ]
  } finally {
    loading.value = false
  }
}

// Save workout
async function saveWorkout() {
  if (!selectedMovement.value) {
    return
  }

  saving.value = true
  error.value = null

  try {
    // Prepare workout data
    const workoutData = {
      workout_date: workoutDate.value,
      workout_type: workoutType.value === 'named' ? 'named_wod' : 'custom',
      notes: notes.value || undefined,
      movements: [
        {
          movement_id: selectedMovement.value,
          weight: weight.value ? parseFloat(weight.value) : undefined,
          sets: sets.value ? parseInt(sets.value) : undefined,
          reps: reps.value ? parseInt(reps.value) : undefined,
          is_rx: movementType.value === 'rx',
        },
      ],
    }

    const response = await axios.post('/api/workouts', workoutData)
    console.log('Workout saved successfully:', response.data)

    // Success! Navigate to workouts list
    router.push('/workouts')
  } catch (err) {
    console.error('Failed to save workout:', err)
    if (err.response) {
      error.value = err.response.data?.message || `Error ${err.response.status}: ${err.response.statusText}`
    } else if (err.request) {
      error.value = 'No response from server. Is the backend running?'
    } else {
      error.value = 'Failed to save workout: ' + err.message
    }
  } finally {
    saving.value = false
  }
}

// Reset form
function resetForm() {
  workoutDate.value = new Date().toISOString().split('T')[0]
  workoutType.value = 'named'
  selectedMovement.value = null
  weight.value = null
  sets.value = null
  reps.value = null
  movementType.value = 'rx'
  notes.value = ''
  error.value = null
}

// Load movements on mount
onMounted(() => {
  fetchMovements()
})
</script>
