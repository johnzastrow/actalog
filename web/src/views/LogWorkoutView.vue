<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto; padding-bottom: 100px">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-btn icon="mdi-arrow-left" color="white" @click="$router.back()" />
      <v-toolbar-title class="text-white font-weight-bold">Log Workout</v-toolbar-title>
      <v-btn icon="mdi-dots-vertical" color="white" />
    </v-app-bar>

    <v-container class="pa-3" style="margin-top: 56px; margin-bottom: 80px">
      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable @click:close="error = null" class="mb-4">
        {{ error }}
      </v-alert>

      <v-form>
        <!-- Workout Date -->
        <div class="mb-3">
          <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
            Workout Date
          </label>
          <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
            <v-text-field
              v-model="workoutDate"
              type="date"
              append-inner-icon="mdi-calendar"
              variant="plain"
              density="compact"
              hide-details
              style="color: #1a1a1a; font-weight: 500"
            />
          </v-card>
        </div>

        <!-- Workout Type -->
        <div class="mb-3">
          <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
            Workout Type
          </label>
          <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
            <v-btn-toggle
              v-model="workoutType"
              mandatory
              divided
              density="comfortable"
              style="width: 100%"
            >
              <v-btn
                value="named"
                style="flex: 1; text-transform: none; font-weight: 600"
                :style="
                  workoutType === 'named'
                    ? 'border: 2px solid #00bcd4; color: #00bcd4'
                    : 'border: 2px solid transparent; color: #666'
                "
                rounded="lg"
                prepend-icon="mdi-format-list-bulleted"
              >
                Named WOD
              </v-btn>
              <v-btn
                value="custom"
                style="flex: 1; text-transform: none; font-weight: 600"
                :style="
                  workoutType === 'custom'
                    ? 'border: 2px solid #00bcd4; color: #00bcd4'
                    : 'border: 2px solid transparent; color: #666'
                "
                rounded="lg"
                prepend-icon="mdi-pencil"
              >
                Custom
              </v-btn>
            </v-btn-toggle>
          </v-card>
        </div>

        <!-- Select Movement -->
        <div class="mb-3">
          <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
            Select Movement
          </label>
          <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
            <v-autocomplete
              v-model="selectedMovement"
              :items="movements"
              item-title="title"
              item-value="value"
              :loading="loading"
              placeholder="Type to search movements..."
              variant="plain"
              density="compact"
              hide-details
              clearable
              auto-select-first
              style="color: #1a1a1a; font-weight: 500"
            >
              <template #prepend-inner>
                <v-icon color="#00bcd4" size="small">mdi-magnify</v-icon>
              </template>
              <template #item="{ props, item }">
                <v-list-item v-bind="props" density="compact">
                  <template #prepend>
                    <v-icon
                      :color="item.raw.type === 'weightlifting' ? '#00bcd4' : '#666'"
                      size="small"
                    >
                      mdi-dumbbell
                    </v-icon>
                  </template>
                  <v-list-item-title class="text-caption">
                    {{ item.raw.title }}
                  </v-list-item-title>
                  <v-list-item-subtitle class="text-caption">
                    {{ item.raw.type }}
                  </v-list-item-subtitle>
                </v-list-item>
              </template>
            </v-autocomplete>
          </v-card>
        </div>

        <!-- Add Custom Movement Button -->
        <v-btn
          variant="outlined"
          color="#00bcd4"
          block
          class="mb-3"
          prepend-icon="mdi-plus"
          size="small"
          style="
            border: 2px dashed #00bcd4;
            text-transform: none;
            font-weight: 600;
            color: #00bcd4;
          "
          rounded="lg"
        >
          Add Custom Movement
        </v-btn>

        <!-- Movement Details Grid -->
        <v-row dense class="mb-1">
          <!-- Weight -->
          <v-col cols="6">
            <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
              Weight (lbs)
            </label>
            <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
              <v-text-field
                v-model="weight"
                type="number"
                placeholder="0"
                variant="plain"
                density="compact"
                hide-details
                append-inner-icon="mdi-weight-lifter"
                class="text-h6 text-center font-weight-medium"
                style="color: #1a1a1a"
              />
            </v-card>
          </v-col>

          <!-- Sets -->
          <v-col cols="6">
            <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
              Sets
            </label>
            <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
              <v-text-field
                v-model="sets"
                type="number"
                placeholder="0"
                variant="plain"
                density="compact"
                hide-details
                append-inner-icon="mdi-layers-triple"
                class="text-h6 text-center font-weight-medium"
                style="color: #1a1a1a"
              />
            </v-card>
          </v-col>
        </v-row>

        <v-row dense class="mb-3">
          <!-- Reps -->
          <v-col cols="6">
            <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
              Reps
            </label>
            <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
              <v-text-field
                v-model="reps"
                type="number"
                placeholder="0"
                variant="plain"
                density="compact"
                hide-details
                append-inner-icon="mdi-numeric"
                class="text-h6 text-center font-weight-medium"
                style="color: #1a1a1a"
              />
            </v-card>
          </v-col>

          <!-- Type (Rx/Scaled) -->
          <v-col cols="6">
            <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
              Type
            </label>
            <v-card elevation="0" rounded="lg" class="pa-1" style="background: white">
              <v-btn-toggle
                v-model="movementType"
                mandatory
                density="compact"
                style="width: 100%"
              >
                <v-btn
                  value="rx"
                  size="small"
                  style="flex: 1; text-transform: none; font-weight: 600"
                  :style="
                    movementType === 'rx'
                      ? 'background: #00bcd4; color: white'
                      : 'color: #1a1a1a'
                  "
                  rounded="lg"
                >
                  Rx
                </v-btn>
                <v-btn
                  value="scaled"
                  size="small"
                  style="flex: 1; text-transform: none; font-weight: 600"
                  :style="
                    movementType === 'scaled'
                      ? 'background: #00bcd4; color: white'
                      : 'color: #1a1a1a'
                  "
                  rounded="lg"
                >
                  Scaled
                </v-btn>
              </v-btn-toggle>
            </v-card>
          </v-col>
        </v-row>

        <!-- Notes -->
        <div class="mb-3">
          <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
            Notes (Optional)
          </label>
          <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
            <v-textarea
              v-model="notes"
              placeholder="How did it feel? Any observations..."
              variant="plain"
              rows="2"
              hide-details
              auto-grow
              style="color: #1a1a1a"
            />
          </v-card>
        </div>

        <!-- Save Button -->
        <v-btn
          color="#00bcd4"
          size="large"
          block
          :loading="saving"
          :disabled="!selectedMovement"
          @click="saveWorkout"
          rounded="lg"
          elevation="2"
          class="text-none font-weight-bold mb-2"
          style="background: #00bcd4; color: white"
        >
          <v-icon start>mdi-content-save</v-icon>
          Save Workout
        </v-btn>
      </v-form>
    </v-container>

    <!-- Bottom Navigation -->
    <v-bottom-navigation
      grow
      style="position: fixed; bottom: 0; background: white"
      elevation="8"
    >
      <v-btn value="dashboard" to="/dashboard">
        <v-icon>mdi-view-dashboard</v-icon>
        <span style="font-size: 10px">Dashboard</span>
      </v-btn>
      <v-btn value="performance" to="/performance">
        <v-icon>mdi-chart-line</v-icon>
        <span style="font-size: 10px">Performance</span>
      </v-btn>
      <v-btn
        value="log"
        style="position: relative; bottom: 20px"
      >
        <v-avatar color="#ffc107" size="56" style="box-shadow: 0 4px 8px rgba(0,0,0,0.2)">
          <v-icon color="white" size="32">mdi-plus</v-icon>
        </v-avatar>
      </v-btn>
      <v-btn value="workouts" to="/workouts">
        <v-icon>mdi-dumbbell</v-icon>
        <span style="font-size: 10px">Workouts</span>
      </v-btn>
      <v-btn value="profile" to="/profile">
        <v-icon>mdi-account</v-icon>
        <span style="font-size: 10px">Profile</span>
      </v-btn>
    </v-bottom-navigation>
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
