<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <!-- Error/Success Alert -->
      <v-alert
        v-if="successMessage"
        type="success"
        closable
        class="mb-3"
        @click:close="successMessage = ''"
      >
        {{ successMessage }}
      </v-alert>

      <v-alert
        v-if="error"
        type="error"
        closable
        class="mb-3"
        @click:close="error = ''"
      >
        {{ error }}
      </v-alert>

      <!-- Basic Information Card -->
      <v-card elevation="0" rounded="lg" class="pa-2 mb-2" bg-color="surface">
        <h2 class="text-body-1 font-weight-bold mb-3" >WOD Details</h2>

        <v-text-field
          v-model="wod.name"
          label="WOD Name"
          placeholder="e.g., Fran, Murph, Helen"
          
          density="compact"
          rounded="lg"
          :error-messages="validationErrors.name"
          class="mb-2"
          required
        >
          <template #prepend-inner>
            <v-icon color="primary" size="small">mdi-fire</v-icon>
          </template>
        </v-text-field>

        <v-select
          v-model="wod.wod_type"
          :items="wodTypes"
          label="WOD Type"
          
          density="compact"
          rounded="lg"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="primary" size="small">mdi-tag</v-icon>
          </template>
        </v-select>

        <v-select
          v-model="wod.scoring_type"
          :items="scoringTypes"
          label="Scoring Type"
          
          density="compact"
          rounded="lg"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="primary" size="small">mdi-trophy</v-icon>
          </template>
        </v-select>

        <v-row dense class="mb-2">
          <v-col cols="6">
            <v-text-field
              v-model.number="wod.time_cap"
              label="Time Cap (minutes)"
              type="number"
              
              density="compact"
              rounded="lg"
              min="0"
              hint="0 for no time cap"
              persistent-hint
            />
          </v-col>
          <v-col cols="6">
            <v-text-field
              v-model.number="wod.rounds"
              label="Rounds"
              type="number"
              
              density="compact"
              rounded="lg"
              min="1"
              hint="Leave empty for AMRAP"
              persistent-hint
            />
          </v-col>
        </v-row>

        <v-switch
          v-model="wod.is_benchmark"
          label="Benchmark WOD"
          color="primary"
          density="compact"
          hide-details
          class="mb-2"
        />

        <v-textarea
          v-model="wod.description"
          label="Description"
          placeholder="Describe the WOD, modifications, or notes"
          hint="Markdown: **bold**, *italic*, [link](url), lists (* or 1.), > quotes"
          persistent-hint

          density="compact"
          rounded="lg"
          rows="3"
          auto-grow
        >
          <template #prepend-inner>
            <v-icon color="primary" size="small">mdi-text</v-icon>
          </template>
        </v-textarea>
      </v-card>

      <!-- Movements Card -->
      <v-card elevation="0" rounded="lg" class="pa-2 mb-2" bg-color="surface">
        <div class="d-flex align-center justify-space-between mb-3">
          <h2 class="text-body-1 font-weight-bold" >WOD Movements</h2>
          <v-btn
            size="small"
            color="primary"
            prepend-icon="mdi-plus"
            style="text-transform: none"
            @click="addMovement"
          >
            Add Movement
          </v-btn>
        </div>

        <!-- Empty State -->
        <div v-if="wod.movements.length === 0" class="text-center py-4">
          <v-icon size="48" color="surface-variant">mdi-dumbbell</v-icon>
          <p class="text-body-2 mt-2 text-medium-emphasis">No movements added yet</p>
          <p class="text-caption text-disabled">Tap "Add Movement" to get started</p>
        </div>

        <!-- Movements List -->
        <div v-else>
          <v-card
            v-for="(movement, index) in wod.movements"
            :key="index"
            elevation="0"
            rounded="lg"
            class="mb-2 pa-3"
          >
            <div class="d-flex align-center mb-2">
              <v-icon color="primary" class="mr-2">mdi-drag-vertical</v-icon>
              <span class="font-weight-bold text-body-2 text-medium-emphasis">#{{ index + 1 }}</span>
              <v-spacer />
              <v-btn
                icon="mdi-delete"
                size="small"
                variant="text"
                color="error"
                @click="removeMovement(index)"
              />
            </div>

            <v-autocomplete
              v-model="movement.movement_id"
              :items="availableMovements"
              item-title="name"
              item-value="id"
              label="Select Movement"
              
              density="compact"
              rounded="lg"
              :loading="loadingMovements"
              clearable
              auto-select-first
              class="mb-2"
            >
              <template #prepend-inner>
                <v-icon color="primary" size="small">mdi-magnify</v-icon>
              </template>
            </v-autocomplete>

            <v-row dense>
              <v-col cols="6">
                <v-text-field
                  v-model.number="movement.reps"
                  label="Reps"
                  type="number"
                  
                  density="compact"
                  rounded="lg"
                  min="1"
                  hide-details
                />
              </v-col>
              <v-col cols="6">
                <v-text-field
                  v-model.number="movement.rx_weight"
                  label="RX Weight (lbs)"
                  type="number"
                  
                  density="compact"
                  rounded="lg"
                  min="0"
                  hide-details
                />
              </v-col>
            </v-row>

            <v-row dense class="mt-2">
              <v-col cols="6">
                <v-text-field
                  v-model.number="movement.scaled_weight"
                  label="Scaled Weight (lbs)"
                  type="number"
                  
                  density="compact"
                  rounded="lg"
                  min="0"
                  hide-details
                />
              </v-col>
              <v-col cols="6">
                <v-text-field
                  v-model.number="movement.distance_meters"
                  label="Distance (m)"
                  type="number"
                  
                  density="compact"
                  rounded="lg"
                  min="0"
                  hide-details
                />
              </v-col>
            </v-row>

            <v-text-field
              v-model="movement.notes"
              label="Notes (Optional)"
              placeholder="e.g., chest-to-bar, unbroken"
              
              density="compact"
              rounded="lg"
              class="mt-2"
              hide-details
            />
          </v-card>
        </div>
      </v-card>

      <!-- Actions Card -->
      <v-card elevation="0" rounded="lg" class="pa-3" bg-color="surface">
        <v-btn
          block
          color="primary"
          size="large"
          rounded="lg"
          :loading="saving"
          style="text-transform: none; font-weight: 600"
          @click="saveWOD"
        >
          <v-icon start>mdi-content-save</v-icon>
          {{ isEditMode ? 'Update WOD' : 'Create WOD' }}
        </v-btn>

        <v-btn
          v-if="isEditMode"
          block
          color="error"
          
          size="large"
          rounded="lg"
          class="mt-2"
          :loading="deleting"
          style="text-transform: none; font-weight: 600"
          @click="confirmDelete"
        >
          <v-icon start>mdi-delete</v-icon>
          Delete WOD
        </v-btn>
      </v-card>
    </v-container>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="text-h6" style="color: rgb(var(--v-theme-error))">Delete WOD?</v-card-title>
        <v-card-text>
          <p class="text-medium-emphasis">
            Are you sure you want to delete "{{ wod.name }}"? This action cannot be undone.
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn
            color="error"
            variant="flat"
            :loading="deleting"
            @click="deleteWOD"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/axios'

const router = useRouter()
const route = useRoute()

// State
const wod = ref({
  name: '',
  wod_type: 'for_time',
  scoring_type: 'time',
  time_cap: 0,
  rounds: null,
  is_benchmark: false,
  description: '',
  movements: []
})

const availableMovements = ref([])
const loadingMovements = ref(false)
const saving = ref(false)
const deleting = ref(false)
const successMessage = ref('')
const error = ref('')
const validationErrors = ref({})
const deleteDialog = ref(false)

const wodTypes = [
  { title: 'For Time', value: 'for_time' },
  { title: 'AMRAP', value: 'amrap' },
  { title: 'EMOM', value: 'emom' },
  { title: 'Chipper', value: 'chipper' },
  { title: 'Ladder', value: 'ladder' },
  { title: 'Tabata', value: 'tabata' }
]

const scoringTypes = [
  { title: 'Time', value: 'time' },
  { title: 'Rounds + Reps', value: 'rounds_reps' },
  { title: 'Reps', value: 'reps' },
  { title: 'Weight', value: 'weight' },
  { title: 'Calories', value: 'calories' }
]

// Computed
const isEditMode = computed(() => !!route.params.id)

// Load movements
async function fetchMovements() {
  loadingMovements.value = true
  try {
    const response = await axios.get('/api/movements')
    availableMovements.value = response.data.movements || []
  } catch (err) {
    console.error('Failed to fetch movements:', err)
    error.value = 'Failed to load movements'
  } finally {
    loadingMovements.value = false
  }
}

// Load existing WOD if editing
async function loadWOD() {
  if (!route.params.id) return

  try {
    const response = await axios.get(`/api/wods/${route.params.id}`)
    const data = response.data.wod

    wod.value = {
      name: data.name || '',
      wod_type: data.wod_type || 'for_time',
      scoring_type: data.scoring_type || 'time',
      time_cap: data.time_cap || 0,
      rounds: data.rounds || null,
      is_benchmark: data.is_benchmark || false,
      description: data.description || '',
      movements: (data.movements || []).map(m => ({
        movement_id: m.movement_id,
        reps: m.reps || null,
        rx_weight: m.rx_weight || null,
        scaled_weight: m.scaled_weight || null,
        distance_meters: m.distance_meters || null,
        notes: m.notes || '',
        order_index: m.order_index
      }))
    }
  } catch (err) {
    console.error('Failed to load WOD:', err)
    error.value = 'Failed to load WOD'
  }
}

// Add new movement
function addMovement() {
  wod.value.movements.push({
    movement_id: null,
    reps: null,
    rx_weight: null,
    scaled_weight: null,
    distance_meters: null,
    notes: '',
    order_index: wod.value.movements.length + 1
  })
}

// Remove movement
function removeMovement(index) {
  wod.value.movements.splice(index, 1)
  // Update order indices
  wod.value.movements.forEach((m, idx) => {
    m.order_index = idx + 1
  })
}

// Validate WOD
function validateWOD() {
  validationErrors.value = {}
  let isValid = true

  if (!wod.value.name || wod.value.name.trim() === '') {
    validationErrors.value.name = 'WOD name is required'
    isValid = false
  }

  if (wod.value.movements.length === 0) {
    error.value = 'Please add at least one movement'
    isValid = false
  }

  for (let i = 0; i < wod.value.movements.length; i++) {
    const movement = wod.value.movements[i]
    if (!movement.movement_id) {
      error.value = `Movement #${i + 1}: Please select a movement`
      isValid = false
      break
    }
  }

  return isValid
}

// Save WOD
async function saveWOD() {
  if (!validateWOD()) return

  saving.value = true
  error.value = ''
  successMessage.value = ''

  try {
    const payload = {
      name: wod.value.name.trim(),
      wod_type: wod.value.wod_type,
      scoring_type: wod.value.scoring_type,
      time_cap: wod.value.time_cap || 0,
      rounds: wod.value.rounds || null,
      is_benchmark: wod.value.is_benchmark,
      description: wod.value.description?.trim() || null,
      movements: wod.value.movements.map((m, idx) => ({
        movement_id: m.movement_id,
        reps: m.reps || null,
        rx_weight: m.rx_weight || null,
        scaled_weight: m.scaled_weight || null,
        distance_meters: m.distance_meters || null,
        notes: m.notes?.trim() || null,
        order_index: idx + 1
      }))
    }

    if (isEditMode.value) {
      await axios.put(`/api/wods/${route.params.id}`, payload)
      successMessage.value = 'WOD updated successfully!'
    } else {
      const response = await axios.post('/api/wods', payload)
      successMessage.value = 'WOD created successfully!'
      // Redirect to edit mode with new ID
      setTimeout(() => {
        router.push(`/wods/${response.data.wod.id}/edit`)
      }, 1500)
    }
  } catch (err) {
    console.error('Failed to save WOD:', err)
    if (err.response?.data?.message) {
      error.value = err.response.data.message
    } else {
      error.value = 'Failed to save WOD. Please try again.'
    }
  } finally {
    saving.value = false
  }
}

// Confirm delete
function confirmDelete() {
  deleteDialog.value = true
}

// Delete WOD
async function deleteWOD() {
  deleting.value = true
  error.value = ''

  try {
    await axios.delete(`/api/wods/${route.params.id}`)
    successMessage.value = 'WOD deleted successfully!'
    setTimeout(() => {
      router.push('/wods')
    }, 1000)
  } catch (err) {
    console.error('Failed to delete WOD:', err)
    error.value = 'Failed to delete WOD. Please try again.'
    deleteDialog.value = false
  } finally {
    deleting.value = false
  }
}

// Initialize
onMounted(async () => {
  await fetchMovements()
  await loadWOD()
})
</script>
