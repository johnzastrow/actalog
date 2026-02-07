<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <!-- User Info Badge -->
      <div class="d-flex align-center justify-end mb-2">
        <v-chip
          :color="isAdmin ? 'error' : 'info'"
          variant="tonal"
          size="small"
          class="mr-2"
        >
          <v-icon start size="small">mdi-shield-account</v-icon>
          {{ isAdmin ? 'Admin' : 'User' }}
        </v-chip>
        <span class="text-caption text-medium-emphasis">{{ authStore.user?.email }}</span>
      </div>

      <!-- Error Alert -->
      <v-alert
        v-if="error"
        type="error"
        closable
        class="mb-3"
        @click:close="error = ''"
      >
        {{ error }}
      </v-alert>

      <!-- Top Action Buttons -->
      <div class="d-flex gap-2 mb-3">
        <v-btn
          v-if="isEditMode && !hasUnsavedChanges"
          color="secondary"
          size="large"
          rounded="lg"
          style="text-transform: none; font-weight: 600"
          @click="openPreview"
        >
          <v-icon start>mdi-eye</v-icon>
          View
        </v-btn>
        <v-btn
          :block="!(isEditMode && !hasUnsavedChanges)"
          color="primary"
          size="large"
          rounded="lg"
          :loading="saving"
          style="text-transform: none; font-weight: 600"
          class="flex-grow-1"
          @click="saveTemplate"
        >
          <v-icon start>mdi-content-save</v-icon>
          Save Workout
        </v-btn>
      </div>

      <!-- Workout Type Selection (Admin Only) -->
      <v-card v-if="isAdmin" elevation="0" rounded="lg" class="pa-3 mb-2" bg-color="surface">
        <div class="d-flex align-center">
          <v-icon color="primary" size="small" class="mr-2">mdi-shield-account</v-icon>
          <span class="text-body-2 font-weight-medium mr-4">Workout Visibility:</span>
          <v-btn-toggle
            v-model="template.is_standard"
            density="compact"
            color="primary"
            mandatory
          >
            <v-btn :value="false" size="small">
              <v-icon start size="small">mdi-account</v-icon>
              Personal
            </v-btn>
            <v-btn :value="true" size="small">
              <v-icon start size="small">mdi-earth</v-icon>
              Standard
            </v-btn>
          </v-btn-toggle>
        </div>
        <p v-if="template.is_standard" class="text-caption text-medium-emphasis mt-2 mb-0">
          Standard workouts are available for all users and can be assigned to class sessions.
        </p>
      </v-card>

      <!-- Basic Information Card -->
      <v-card elevation="0" rounded="lg" class="pa-2 mb-2" bg-color="surface">
        <h2 class="text-body-1 font-weight-bold mb-3">Workout Details</h2>

        <v-text-field
          v-model="template.name"
          label="Workout Name"
          placeholder="e.g., Upper Body Strength"

          density="compact"
          rounded="lg"
          :error-messages="validationErrors.name"
          class="mb-2"
          required
        >
          <template #prepend-inner>
            <v-icon color="primary" size="small">mdi-dumbbell</v-icon>
          </template>
        </v-text-field>

        <v-textarea
          v-model="template.intro_warmup"
          label="Intro and Warmup (Optional)"
          placeholder="Add warmup instructions, setup notes, or introduction"
          hint="Markdown: **bold**, *italic*, [link](url), lists (* or 1.), > quotes"
          persistent-hint

          density="compact"
          rounded="lg"
          rows="3"
          auto-grow
        >
          <template #prepend-inner>
            <v-icon color="primary" size="small">mdi-run</v-icon>
          </template>
        </v-textarea>
      </v-card>

      <!-- Movements Card -->
      <v-card elevation="0" rounded="lg" class="pa-2 mb-2" bg-color="surface">
        <div class="d-flex align-center justify-space-between mb-3">
          <h2 class="text-body-1 font-weight-bold">Movements</h2>
          <v-btn
            size="small"
            color="primary"
            prepend-icon="mdi-plus"
            style="text-transform: none"
            @click="addMovement"
          >
            Add
          </v-btn>
        </div>

        <!-- Empty State -->
        <div v-if="template.movements.length === 0" class="text-center py-4">
          <v-icon size="48" color="surface-variant">mdi-dumbbell</v-icon>
          <p class="text-body-2 mt-2 text-medium-emphasis">No movements added yet</p>
          <p class="text-caption text-disabled">Tap "Add Movement" to get started</p>
        </div>

        <!-- Movements List -->
        <div v-else>
          <v-card
            v-for="(movement, index) in template.movements"
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

            <!-- Instructions first -->
            <v-textarea
              v-model="movement.instructions"
              label="Instructions (Optional)"
              placeholder="e.g., Setup, form cues, scaling options"
              hint="Markdown supported: **bold**, *italic*, - lists"
              persistent-hint
              density="compact"
              rounded="lg"
              rows="2"
              auto-grow
              class="mb-2"
            >
              <template #prepend-inner>
                <v-icon color="info" size="small">mdi-clipboard-text-outline</v-icon>
              </template>
            </v-textarea>

            <!-- Movement specifics -->
            <v-row dense>
              <v-col cols="4">
                <v-text-field
                  v-model.number="movement.sets"
                  label="Sets"
                  type="number"

                  density="compact"
                  rounded="lg"
                  min="1"
                  hide-details
                />
              </v-col>
              <v-col cols="4">
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
              <v-col cols="4">
                <v-text-field
                  v-model.number="movement.rest_seconds"
                  label="Rest (s)"
                  type="number"

                  density="compact"
                  rounded="lg"
                  min="0"
                  hide-details
                />
              </v-col>
            </v-row>

            <!-- Notes last -->
            <v-textarea
              v-model="movement.notes"
              label="Notes (Optional)"
              placeholder="e.g., Tempo: 3-1-1-0, RPE 8"

              density="compact"
              rounded="lg"
              rows="2"
              auto-grow
              class="mt-2"
              hide-details
            />
          </v-card>
        </div>
      </v-card>

      <!-- WODs Card -->
      <v-card elevation="0" rounded="lg" class="pa-2 mb-2" bg-color="surface">
        <div class="d-flex align-center justify-space-between mb-3">
          <h2 class="text-body-1 font-weight-bold">WODs</h2>
          <v-btn
            size="small"
            color="primary"
            prepend-icon="mdi-plus"
            style="text-transform: none"
            @click="addWOD"
          >
            Add
          </v-btn>
        </div>

        <!-- Empty State -->
        <div v-if="template.wods.length === 0" class="text-center py-4">
          <v-icon size="48" color="surface-variant">mdi-fire</v-icon>
          <p class="text-body-2 mt-2 text-medium-emphasis">No WODs added yet</p>
          <p class="text-caption text-disabled">Tap "Add WOD" to include benchmark workouts</p>
        </div>

        <!-- WODs List -->
        <div v-else>
          <v-card
            v-for="(wod, index) in template.wods"
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
                @click="removeWOD(index)"
              />
            </div>

            <v-autocomplete
              v-model="wod.wod_id"
              :items="availableWODs"
              item-title="name"
              item-value="id"
              label="Select WOD"
              
              density="compact"
              rounded="lg"
              :loading="loadingWODs"
              clearable
              auto-select-first
              class="mb-2"
            >
              <template #prepend-inner>
                <v-icon color="primary" size="small">mdi-magnify</v-icon>
              </template>
              <template #item="{ props, item }">
                <v-list-item v-bind="props">
                  <template #prepend>
                    <v-icon color="primary" size="small">mdi-fire</v-icon>
                  </template>
                  <template #subtitle>
                    <span class="text-caption">{{ item.raw.type }} - {{ item.raw.regime }}</span>
                  </template>
                </v-list-item>
              </template>
            </v-autocomplete>

            <v-textarea
              v-model="wod.instructions"
              label="Instructions (Optional)"
              placeholder="e.g., Setup, standards, scaling options"
              hint="Markdown supported: **bold**, *italic*, - lists"
              persistent-hint
              density="compact"
              rounded="lg"
              rows="2"
              auto-grow
              class="mb-2"
            >
              <template #prepend-inner>
                <v-icon color="info" size="small">mdi-clipboard-text-outline</v-icon>
              </template>
            </v-textarea>

            <v-textarea
              v-model="wod.notes"
              label="Notes (Optional)"
              placeholder="e.g., Time cap, scaling notes"
              density="compact"
              rounded="lg"
              rows="2"
              auto-grow
              hide-details
            />
          </v-card>
        </div>
      </v-card>

      <!-- Notes Card -->
      <v-card elevation="0" rounded="lg" class="pa-2 mb-2" bg-color="surface">
        <h2 class="text-body-1 font-weight-bold mb-3">Notes</h2>

        <v-textarea
          v-model="template.notes"
          label="Notes (Optional)"
          placeholder="Add any closing notes, cooldown instructions, or additional information"
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

      <!-- Actions Card -->
      <v-card elevation="0" rounded="lg" class="pa-3" bg-color="surface">
        <div class="d-flex gap-2 mb-2">
          <v-btn
            v-if="isEditMode && !hasUnsavedChanges"
            color="secondary"
            size="large"
            rounded="lg"
            style="text-transform: none; font-weight: 600"
            @click="openPreview"
          >
            <v-icon start>mdi-eye</v-icon>
            View
          </v-btn>
          <v-btn
            :block="!(isEditMode && !hasUnsavedChanges)"
            color="primary"
            size="large"
            rounded="lg"
            :loading="saving"
            style="text-transform: none; font-weight: 600"
            class="flex-grow-1"
            @click="saveTemplate"
          >
            <v-icon start>mdi-content-save</v-icon>
            Save Workout
          </v-btn>
        </div>

        <v-btn
          v-if="isEditMode"
          block
          color="error"
          variant="outlined"
          size="large"
          rounded="lg"
          :loading="deleting"
          style="text-transform: none; font-weight: 600"
          @click="confirmDelete"
        >
          <v-icon start>mdi-delete</v-icon>
          Delete Workout
        </v-btn>
      </v-card>
    </v-container>

    <!-- Workout Preview Dialog -->
    <DetailViewDialog
      v-model="previewDialog"
      type="template"
      :id="previewId"
    />

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="text-h6" style="color: rgb(var(--v-theme-error))">Delete Workout?</v-card-title>
        <v-card-text>
          <p class="text-medium-emphasis">
            Are you sure you want to delete "{{ template.name }}"? This action cannot be undone.
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn
            color="error"
            variant="flat"
            :loading="deleting"
            @click="deleteTemplate"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Save Confirmation Snackbar -->
    <v-snackbar
      v-model="snackbar"
      :color="snackbarColor"
      :timeout="4000"
      location="bottom"
    >
      {{ snackbarText }}
      <template #actions>
        <v-btn variant="text" @click="snackbar = false">Close</v-btn>
      </template>
    </v-snackbar>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/axios'
import { useAuthStore } from '@/stores/auth'
import DetailViewDialog from '@/components/DetailViewDialog.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// Helper to format today's date in long form
function getTodayLongDate() {
  const options = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' }
  return new Date().toLocaleDateString('en-US', options)
}

// Boilerplate markdown example for intro/warmup
const introBoilerplate = `## Warmup (10 minutes)

**General warmup:**
- 400m jog or 2 min row
- Dynamic stretches

### Movement Prep
1. First movement prep
2. Second movement prep
3. Third movement prep

*Focus on quality over speed*

---

> **Coach's Note:** Scale as needed based on athlete's ability level.

### Equipment Needed
- Barbell
- Plates
- Pull-up bar

[Link to demo video](https://example.com)`

// State
const template = ref({
  name: getTodayLongDate(),
  intro_warmup: introBoilerplate,
  notes: '',
  wods: [],
  movements: [],
  is_standard: false
})

const availableMovements = ref([])
const availableWODs = ref([])
const loadingMovements = ref(false)
const loadingWODs = ref(false)
const saving = ref(false)
const deleting = ref(false)
const error = ref('')
const validationErrors = ref({})
const deleteDialog = ref(false)
const previewDialog = ref(false)
const previewId = ref(null)
const savedState = ref(null) // Track saved state for change detection

// Snackbar state
const snackbar = ref(false)
const snackbarText = ref('')
const snackbarColor = ref('success')

// Computed
const isEditMode = computed(() => !!route.params.id)
const isAdmin = computed(() => authStore.user?.role === 'admin')

// Track unsaved changes
const hasUnsavedChanges = computed(() => {
  if (!savedState.value) return true // New workout always has "unsaved" changes
  return JSON.stringify(template.value) !== savedState.value
})

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

// Load WODs
async function fetchWODs() {
  loadingWODs.value = true
  try {
    const [standardRes, customRes] = await Promise.all([
      axios.get('/api/wods/standard'),
      axios.get('/api/wods/my-wods')
    ])
    const standard = Array.isArray(standardRes.data.wods) ? standardRes.data.wods : []
    const custom = Array.isArray(customRes.data.wods) ? customRes.data.wods : []
    availableWODs.value = [...standard, ...custom]
  } catch (err) {
    console.error('Failed to fetch WODs:', err)
    error.value = 'Failed to load WODs'
  } finally {
    loadingWODs.value = false
  }
}

// Load existing template if editing
async function loadTemplate() {
  if (!route.params.id) return

  try {
    const response = await axios.get(`/api/templates/${route.params.id}`)
    const data = response.data.template

    template.value = {
      name: data.name || '',
      workout_type: data.workout_type || 'strength',
      intro_warmup: data.intro_warmup || '',
      notes: data.notes || '',
      is_standard: data.created_by === null, // Standard if no creator
      wods: (data.wods || []).map(w => ({
        wod_id: w.wod_id,
        instructions: w.instructions || '',
        notes: w.notes || '',
        order_index: w.order_index
      })),
      movements: (data.movements || []).map(m => ({
        movement_id: m.movement_id,
        sets: m.sets || null,
        reps: m.reps || null,
        rest_seconds: m.rest_seconds || null,
        instructions: m.instructions || '',
        notes: m.notes || '',
        order_index: m.order_index
      }))
    }
    // Save initial state for change detection
    savedState.value = JSON.stringify(template.value)
  } catch (err) {
    console.error('Failed to load workout:', err)
    error.value = 'Failed to load workout'
  }
}

// Show snackbar notification
function showSnackbar(message, color = 'success') {
  snackbarText.value = message
  snackbarColor.value = color
  snackbar.value = true
}

// Open preview dialog
function openPreview() {
  if (route.params.id) {
    previewId.value = parseInt(route.params.id)
    previewDialog.value = true
  }
}

// Add new movement
function addMovement() {
  template.value.movements.push({
    movement_id: null,
    sets: null,
    reps: null,
    work_time: 60,
    instructions: '',
    notes: '',
    order_index: template.value.movements.length + 1
  })
}

// Remove movement
function removeMovement(index) {
  template.value.movements.splice(index, 1)
  // Update order indices
  template.value.movements.forEach((m, idx) => {
    m.order_index = idx + 1
  })
}

// Add new WOD
function addWOD() {
  template.value.wods.push({
    wod_id: null,
    instructions: '',
    notes: '',
    order_index: template.value.wods.length + 1
  })
}

// Remove WOD
function removeWOD(index) {
  template.value.wods.splice(index, 1)
  // Update order indices
  template.value.wods.forEach((w, idx) => {
    w.order_index = idx + 1
  })
}

// Validate workout
function validateTemplate() {
  validationErrors.value = {}
  let isValid = true

  if (!template.value.name || template.value.name.trim() === '') {
    validationErrors.value.name = 'Workout name is required'
    isValid = false
  }

  // Movements and WODs are optional - no validation required

  // Validate movements (only if any are added)
  for (let i = 0; i < template.value.movements.length; i++) {
    const movement = template.value.movements[i]
    if (!movement.movement_id) {
      error.value = `Movement #${i + 1}: Please select a movement`
      isValid = false
      break
    }
  }

  // Validate WODs
  for (let i = 0; i < template.value.wods.length; i++) {
    const wod = template.value.wods[i]
    if (!wod.wod_id) {
      error.value = `WOD #${i + 1}: Please select a WOD`
      isValid = false
      break
    }
  }

  return isValid
}

// Save workout
async function saveTemplate() {
  if (!validateTemplate()) return

  saving.value = true
  error.value = ''

  try {
    const payload = {
      name: template.value.name.trim(),
      workout_type: template.value.workout_type,
      intro_warmup: template.value.intro_warmup?.trim() || null,
      notes: template.value.notes?.trim() || null,
      is_standard: isAdmin.value ? template.value.is_standard : false,
      wods: template.value.wods.map((w, idx) => ({
        wod_id: w.wod_id,
        instructions: w.instructions?.trim() || '',
        notes: w.notes?.trim() || '',
        order_index: idx + 1
      })),
      movements: template.value.movements.map((m, idx) => ({
        movement_id: m.movement_id,
        sets: m.sets || null,
        reps: m.reps || null,
        work_time: m.work_time || null,
        instructions: m.instructions?.trim() || '',
        notes: m.notes?.trim() || '',
        order_index: idx + 1
      }))
    }

    if (isEditMode.value) {
      await axios.put(`/api/templates/${route.params.id}`, payload)
      showSnackbar('Workout saved successfully!')
      // Update saved state after successful save
      savedState.value = JSON.stringify(template.value)
    } else {
      const response = await axios.post('/api/templates', payload)
      showSnackbar('Workout created successfully!')
      // Redirect to edit mode with new ID
      setTimeout(() => {
        router.push(`/workouts/templates/${response.data.template.id}/edit`)
      }, 1500)
    }
  } catch (err) {
    console.error('Failed to save workout:', err)
    if (err.response?.data?.message) {
      showSnackbar(err.response.data.message, 'error')
    } else {
      showSnackbar('Failed to save workout. Please try again.', 'error')
    }
  } finally {
    saving.value = false
  }
}

// Confirm delete
function confirmDelete() {
  deleteDialog.value = true
}

// Delete workout
async function deleteTemplate() {
  deleting.value = true

  try {
    await axios.delete(`/api/templates/${route.params.id}`)
    showSnackbar('Workout deleted successfully!')
    setTimeout(() => {
      router.push('/workouts')
    }, 1000)
  } catch (err) {
    console.error('Failed to delete workout:', err)
    showSnackbar('Failed to delete workout. Please try again.', 'error')
    deleteDialog.value = false
  } finally {
    deleting.value = false
  }
}

// Initialize
onMounted(async () => {
  await Promise.all([fetchMovements(), fetchWODs()])
  await loadTemplate()

  // Handle selected movement from library
  if (route.query.selectedMovement) {
    const movementId = parseInt(route.query.selectedMovement)
    addMovement()
    template.value.movements[template.value.movements.length - 1].movement_id = movementId
    // Clear query param
    router.replace({ path: route.path })
  }

  // Handle selected WOD from library
  if (route.query.selectedWOD) {
    const wodId = parseInt(route.query.selectedWOD)
    addWOD()
    template.value.wods[template.value.wods.length - 1].wod_id = wodId
    // Clear query param
    router.replace({ path: route.path })
  }
})
</script>
