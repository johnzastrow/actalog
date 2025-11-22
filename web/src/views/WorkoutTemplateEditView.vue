<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">

    <v-container class="pa-2" style=" margin-bottom: 70px">
      <!-- Error/Success Alert -->
      <v-alert
        v-if="successMessage"
        type="success"
        closable
        @click:close="successMessage = ''"
        class="mb-3"
      >
        {{ successMessage }}
      </v-alert>

      <v-alert
        v-if="error"
        type="error"
        closable
        @click:close="error = ''"
        class="mb-3"
      >
        {{ error }}
      </v-alert>

      <!-- Basic Information Card -->
      <v-card elevation="0" rounded="lg" class="pa-2 mb-2" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Template Details</h2>

        <v-text-field
          v-model="template.name"
          label="Template Name"
          placeholder="e.g., Upper Body Strength"
          variant="outlined"
          density="compact"
          rounded="lg"
          :error-messages="validationErrors.name"
          class="mb-2"
          required
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-dumbbell</v-icon>
          </template>
        </v-text-field>

        <v-textarea
          v-model="template.description"
          label="Description (Optional)"
          placeholder="Add any notes or instructions for this template"
          variant="outlined"
          density="compact"
          rounded="lg"
          rows="3"
          auto-grow
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-text</v-icon>
          </template>
        </v-textarea>
      </v-card>

      <!-- Movements Card -->
      <v-card elevation="0" rounded="lg" class="pa-2 mb-2" style="background: white">
        <div class="d-flex align-center justify-space-between mb-3">
          <h2 class="text-body-1 font-weight-bold" style="color: #1a1a1a">Movements</h2>
          <div class="d-flex gap-2">
            <v-btn
              size="small"
              color="#9c27b0"
              variant="outlined"
              prepend-icon="mdi-library"
              @click="browseMovements"
              style="text-transform: none"
            >
              Browse
            </v-btn>
            <v-btn
              size="small"
              color="#00bcd4"
              prepend-icon="mdi-plus"
              @click="addMovement"
              style="text-transform: none"
            >
              Add
            </v-btn>
          </div>
        </div>

        <!-- Empty State -->
        <div v-if="template.movements.length === 0" class="text-center py-4">
          <v-icon size="48" color="#ccc">mdi-dumbbell</v-icon>
          <p class="text-body-2 mt-2" style="color: #666">No movements added yet</p>
          <p class="text-caption" style="color: #999">Tap "Add Movement" to get started</p>
        </div>

        <!-- Movements List -->
        <div v-else>
          <v-card
            v-for="(movement, index) in template.movements"
            :key="index"
            elevation="0"
            rounded="lg"
            class="mb-2 pa-3"
            style="border: 1px solid #e0e0e0"
          >
            <div class="d-flex align-center mb-2">
              <v-icon color="#00bcd4" class="mr-2">mdi-drag-vertical</v-icon>
              <span class="font-weight-bold text-body-2" style="color: #666">#{{ index + 1 }}</span>
              <v-spacer />
              <v-btn
                icon="mdi-delete"
                size="small"
                variant="text"
                color="#e91e63"
                @click="removeMovement(index)"
              />
            </div>

            <v-autocomplete
              v-model="movement.movement_id"
              :items="availableMovements"
              item-title="name"
              item-value="id"
              label="Select Movement"
              variant="outlined"
              density="compact"
              rounded="lg"
              :loading="loadingMovements"
              clearable
              auto-select-first
              class="mb-2"
            >
              <template #prepend-inner>
                <v-icon color="#00bcd4" size="small">mdi-magnify</v-icon>
              </template>
            </v-autocomplete>

            <v-row dense>
              <v-col cols="4">
                <v-text-field
                  v-model.number="movement.sets"
                  label="Sets"
                  type="number"
                  variant="outlined"
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
                  variant="outlined"
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
                  variant="outlined"
                  density="compact"
                  rounded="lg"
                  min="0"
                  hide-details
                />
              </v-col>
            </v-row>

            <v-textarea
              v-model="movement.notes"
              label="Notes (Optional)"
              placeholder="e.g., Tempo: 3-1-1-0, RPE 8"
              variant="outlined"
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
      <v-card elevation="0" rounded="lg" class="pa-2 mb-2" style="background: white">
        <div class="d-flex align-center justify-space-between mb-3">
          <h2 class="text-body-1 font-weight-bold" style="color: #1a1a1a">WODs</h2>
          <div class="d-flex gap-2">
            <v-btn
              size="small"
              color="#9c27b0"
              variant="outlined"
              prepend-icon="mdi-library"
              @click="browseWODs"
              style="text-transform: none"
            >
              Browse
            </v-btn>
            <v-btn
              size="small"
              color="teal"
              prepend-icon="mdi-plus"
              @click="addWOD"
              style="text-transform: none"
            >
              Add
            </v-btn>
          </div>
        </div>

        <!-- Empty State -->
        <div v-if="template.wods.length === 0" class="text-center py-4">
          <v-icon size="48" color="#ccc">mdi-fire</v-icon>
          <p class="text-body-2 mt-2" style="color: #666">No WODs added yet</p>
          <p class="text-caption" style="color: #999">Tap "Add WOD" to include benchmark workouts</p>
        </div>

        <!-- WODs List -->
        <div v-else>
          <v-card
            v-for="(wod, index) in template.wods"
            :key="index"
            elevation="0"
            rounded="lg"
            class="mb-2 pa-3"
            style="border: 1px solid #e0e0e0"
          >
            <div class="d-flex align-center mb-2">
              <v-icon color="teal" class="mr-2">mdi-drag-vertical</v-icon>
              <span class="font-weight-bold text-body-2" style="color: #666">#{{ index + 1 }}</span>
              <v-spacer />
              <v-btn
                icon="mdi-delete"
                size="small"
                variant="text"
                color="#e91e63"
                @click="removeWOD(index)"
              />
            </div>

            <v-autocomplete
              v-model="wod.wod_id"
              :items="availableWODs"
              item-title="name"
              item-value="id"
              label="Select WOD"
              variant="outlined"
              density="compact"
              rounded="lg"
              :loading="loadingWODs"
              clearable
              auto-select-first
              class="mb-2"
            >
              <template #prepend-inner>
                <v-icon color="teal" size="small">mdi-magnify</v-icon>
              </template>
              <template #item="{ props, item }">
                <v-list-item v-bind="props">
                  <template #prepend>
                    <v-icon color="teal" size="small">mdi-fire</v-icon>
                  </template>
                  <template #subtitle>
                    <span class="text-caption">{{ item.raw.type }} - {{ item.raw.regime }}</span>
                  </template>
                </v-list-item>
              </template>
            </v-autocomplete>

            <v-textarea
              v-model="wod.notes"
              label="Notes (Optional)"
              placeholder="e.g., Scaling options, time cap"
              variant="outlined"
              density="compact"
              rounded="lg"
              rows="2"
              auto-grow
              hide-details
            />
          </v-card>
        </div>
      </v-card>

      <!-- Actions Card -->
      <v-card elevation="0" rounded="lg" class="pa-3" style="background: white">
        <v-btn
          block
          color="#00bcd4"
          size="large"
          rounded="lg"
          :loading="saving"
          @click="saveTemplate"
          style="text-transform: none; font-weight: 600"
        >
          <v-icon start>mdi-content-save</v-icon>
          {{ isEditMode ? 'Update Template' : 'Create Template' }}
        </v-btn>

        <v-btn
          v-if="isEditMode"
          block
          color="#e91e63"
          variant="outlined"
          size="large"
          rounded="lg"
          class="mt-2"
          :loading="deleting"
          @click="confirmDelete"
          style="text-transform: none; font-weight: 600"
        >
          <v-icon start>mdi-delete</v-icon>
          Delete Template
        </v-btn>
      </v-card>
    </v-container>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="text-h6" style="color: #e91e63">Delete Template?</v-card-title>
        <v-card-text>
          <p style="color: #666">
            Are you sure you want to delete "{{ template.name }}"? This action cannot be undone.
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn
            color="#e91e63"
            variant="flat"
            :loading="deleting"
            @click="deleteTemplate"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/axios'

const router = useRouter()
const route = useRoute()

// State
const template = ref({
  name: '',
  description: '',
  wods: [],
  movements: []
})

const availableMovements = ref([])
const availableWODs = ref([])
const loadingMovements = ref(false)
const loadingWODs = ref(false)
const saving = ref(false)
const deleting = ref(false)
const successMessage = ref('')
const error = ref('')
const validationErrors = ref({})
const deleteDialog = ref(false)

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

// Load WODs
async function fetchWODs() {
  loadingWODs.value = true
  try {
    const response = await axios.get('/api/wods?limit=1000')
    availableWODs.value = response.data.wods || []
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
      description: data.description || '',
      wods: (data.wods || []).map(w => ({
        wod_id: w.wod_id,
        notes: w.notes || '',
        order_index: w.order_index
      })),
      movements: (data.movements || []).map(m => ({
        movement_id: m.movement_id,
        sets: m.sets || null,
        reps: m.reps || null,
        rest_seconds: m.rest_seconds || null,
        notes: m.notes || '',
        order_index: m.order_index
      }))
    }
  } catch (err) {
    console.error('Failed to load template:', err)
    error.value = 'Failed to load template'
  }
}

// Add new movement
function addMovement() {
  template.value.movements.push({
    movement_id: null,
    sets: null,
    reps: null,
    work_time: 60,
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

// Browse movements library
function browseMovements() {
  const currentPath = route.path
  router.push({
    path: '/movements',
    query: { select: 'true', returnPath: currentPath }
  })
}

// Add new WOD
function addWOD() {
  template.value.wods.push({
    wod_id: null,
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

// Browse WODs library
function browseWODs() {
  const currentPath = route.path
  router.push({
    path: '/wods',
    query: { select: 'true', returnPath: currentPath }
  })
}

// Validate template
function validateTemplate() {
  validationErrors.value = {}
  let isValid = true

  if (!template.value.name || template.value.name.trim() === '') {
    validationErrors.value.name = 'Template name is required'
    isValid = false
  }

  if (template.value.movements.length === 0 && template.value.wods.length === 0) {
    error.value = 'Please add at least one movement or WOD'
    isValid = false
  }

  // Validate movements
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

// Save template
async function saveTemplate() {
  if (!validateTemplate()) return

  saving.value = true
  error.value = ''
  successMessage.value = ''

  try {
    const payload = {
      name: template.value.name.trim(),
      workout_type: template.value.workout_type,
      description: template.value.description?.trim() || null,
      wods: template.value.wods.map((w, idx) => ({
        wod_id: w.wod_id,
        notes: w.notes?.trim() || null,
        order_index: idx + 1
      })),
      movements: template.value.movements.map((m, idx) => ({
        movement_id: m.movement_id,
        sets: m.sets || null,
        reps: m.reps || null,
        work_time: m.work_time || null,
        notes: m.notes?.trim() || null,
        order_index: idx + 1
      }))
    }

    if (isEditMode.value) {
      await axios.put(`/api/templates/${route.params.id}`, payload)
      successMessage.value = 'Template updated successfully!'
    } else {
      const response = await axios.post('/api/templates', payload)
      successMessage.value = 'Template created successfully!'
      // Redirect to edit mode with new ID
      setTimeout(() => {
        router.push(`/workouts/templates/${response.data.template.id}/edit`)
      }, 1500)
    }
  } catch (err) {
    console.error('Failed to save template:', err)
    if (err.response?.data?.message) {
      error.value = err.response.data.message
    } else {
      error.value = 'Failed to save template. Please try again.'
    }
  } finally {
    saving.value = false
  }
}

// Confirm delete
function confirmDelete() {
  deleteDialog.value = true
}

// Delete template
async function deleteTemplate() {
  deleting.value = true
  error.value = ''

  try {
    await axios.delete(`/api/templates/${route.params.id}`)
    successMessage.value = 'Template deleted successfully!'
    setTimeout(() => {
      router.push('/workouts')
    }, 1000)
  } catch (err) {
    console.error('Failed to delete template:', err)
    error.value = 'Failed to delete template. Please try again.'
    deleteDialog.value = false
  } finally {
    deleting.value = false
  }
}

// Handle back navigation
function handleBack() {
  if (template.value.wods.length > 0 || template.value.movements.length > 0 || template.value.name) {
    if (confirm('You have unsaved changes. Are you sure you want to leave?')) {
      router.back()
    }
  } else {
    router.back()
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
