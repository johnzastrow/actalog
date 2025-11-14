<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" density="compact" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-btn icon="mdi-close" color="white" size="small" @click="handleBack" />
      <v-toolbar-title class="text-white font-weight-bold">
        {{ isEditMode ? 'Edit Movement' : 'Create Movement' }}
      </v-toolbar-title>
      <v-spacer />
      <v-btn
        icon="mdi-content-save"
        color="white"
        :loading="saving"
        @click="saveMovement"
      />
    </v-app-bar>

    <v-container class="pa-2" style="margin-top: 36px; margin-bottom: 70px">
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
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Basic Information</h2>

        <v-text-field
          v-model="movement.name"
          label="Movement Name *"
          placeholder="e.g., Barbell Back Squat"
          variant="outlined"
          density="compact"
          rounded="lg"
          :error-messages="validationErrors.name"
          class="mb-2"
          required
          autofocus
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-dumbbell</v-icon>
          </template>
        </v-text-field>

        <v-select
          v-model="movement.type"
          :items="movementTypes"
          label="Movement Type *"
          variant="outlined"
          density="compact"
          rounded="lg"
          :error-messages="validationErrors.type"
          class="mb-2"
          required
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-tag</v-icon>
          </template>
        </v-select>

        <v-select
          v-model="movement.difficulty"
          :items="difficultyLevels"
          label="Difficulty Level"
          variant="outlined"
          density="compact"
          rounded="lg"
          class="mb-2"
          clearable
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-signal</v-icon>
          </template>
        </v-select>

        <v-text-field
          v-model="movement.equipment"
          label="Equipment Required"
          placeholder="e.g., Barbell, Squat Rack"
          variant="outlined"
          density="compact"
          rounded="lg"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-tools</v-icon>
          </template>
        </v-text-field>

        <v-text-field
          v-model="movement.primaryMuscles"
          label="Primary Muscle Groups"
          placeholder="e.g., Quadriceps, Glutes, Core"
          variant="outlined"
          density="compact"
          rounded="lg"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-arm-flex</v-icon>
          </template>
        </v-text-field>

        <!-- Info Note -->
        <v-alert
          type="info"
          variant="tonal"
          density="compact"
        >
          <template #prepend>
            <v-icon size="small">mdi-information</v-icon>
          </template>
          Custom movements will be available only to you
        </v-alert>
      </v-card>

      <!-- Description & Instructions Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Description & Instructions</h2>

        <v-textarea
          v-model="movement.description"
          label="Movement Description *"
          placeholder="Describe the movement execution, key technique points, and setup..."
          variant="outlined"
          density="compact"
          rounded="lg"
          rows="4"
          auto-grow
          :error-messages="validationErrors.description"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-text</v-icon>
          </template>
        </v-textarea>

        <v-textarea
          v-model="movement.coachingCues"
          label="Coaching Cues"
          placeholder="e.g., Keep chest up, drive through heels, maintain neutral spine..."
          variant="outlined"
          density="compact"
          rounded="lg"
          rows="3"
          auto-grow
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-comment-text</v-icon>
          </template>
        </v-textarea>

        <v-text-field
          v-model="movement.videoUrl"
          label="Video Tutorial URL"
          placeholder="https://youtube.com/watch?v=..."
          variant="outlined"
          density="compact"
          rounded="lg"
          type="url"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-video</v-icon>
          </template>
        </v-text-field>

        <v-textarea
          v-model="movement.scalingOptions"
          label="Scaling/Modifications"
          placeholder="Beginner: Goblet squat&#10;Intermediate: Front squat&#10;Advanced: Pause squat"
          variant="outlined"
          density="compact"
          rounded="lg"
          rows="3"
          auto-grow
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-stairs</v-icon>
          </template>
        </v-textarea>
      </v-card>

      <!-- Actions Card -->
      <v-card elevation="0" rounded="lg" class="pa-3" style="background: white">
        <v-btn
          block
          color="#00bcd4"
          size="large"
          rounded="lg"
          :loading="saving"
          @click="saveMovement"
          style="text-transform: none; font-weight: 600"
        >
          <v-icon start>mdi-content-save</v-icon>
          {{ isEditMode ? 'Update Movement' : 'Create Movement' }}
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
          Delete Movement
        </v-btn>

        <v-btn
          block
          variant="text"
          size="large"
          rounded="lg"
          class="mt-2"
          @click="handleBack"
          style="text-transform: none"
        >
          Cancel
        </v-btn>
      </v-card>
    </v-container>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="text-h6" style="color: #e91e63">Delete Movement?</v-card-title>
        <v-card-text>
          <p style="color: #666">
            Are you sure you want to delete "{{ movement.name }}"? This action cannot be undone.
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn
            color="#e91e63"
            variant="flat"
            :loading="deleting"
            @click="deleteMovement"
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
const movement = ref({
  name: '',
  type: 'weightlifting',
  difficulty: '',
  equipment: '',
  primaryMuscles: '',
  description: '',
  coachingCues: '',
  videoUrl: '',
  scalingOptions: ''
})

const saving = ref(false)
const deleting = ref(false)
const successMessage = ref('')
const error = ref('')
const validationErrors = ref({})
const deleteDialog = ref(false)

const movementTypes = [
  { title: 'Weightlifting', value: 'weightlifting' },
  { title: 'Gymnastics', value: 'gymnastics' },
  { title: 'Cardio', value: 'cardio' },
  { title: 'Bodyweight', value: 'bodyweight' }
]

const difficultyLevels = [
  { title: 'Beginner', value: 'beginner' },
  { title: 'Intermediate', value: 'intermediate' },
  { title: 'Advanced', value: 'advanced' },
  { title: 'Elite', value: 'elite' }
]

// Computed
const isEditMode = computed(() => !!route.params.id)
const returnPath = computed(() => route.query.returnPath || '/movements')
const selectionMode = computed(() => route.query.select === 'true')

// Load existing movement if editing
async function loadMovement() {
  if (!route.params.id) return

  try {
    const response = await axios.get(`/api/movements/${route.params.id}`)
    const data = response.data.movement || response.data

    // Parse description if it contains structured data
    let parsedData = {}
    try {
      if (data.description && data.description.startsWith('__STRUCTURED__')) {
        const jsonStr = data.description.substring('__STRUCTURED__'.length)
        parsedData = JSON.parse(jsonStr)
      }
    } catch (e) {
      // Not structured data, use as plain description
      parsedData.description = data.description || ''
    }

    movement.value = {
      name: data.name || '',
      type: data.type || 'weightlifting',
      difficulty: parsedData.difficulty || '',
      equipment: parsedData.equipment || '',
      primaryMuscles: parsedData.primaryMuscles || '',
      description: parsedData.description || data.description || '',
      coachingCues: parsedData.coachingCues || '',
      videoUrl: parsedData.videoUrl || '',
      scalingOptions: parsedData.scalingOptions || ''
    }
  } catch (err) {
    console.error('Failed to load movement:', err)
    error.value = 'Failed to load movement'
  }
}

// Validate movement
function validateMovement() {
  validationErrors.value = {}
  let isValid = true

  if (!movement.value.name || movement.value.name.trim() === '') {
    validationErrors.value.name = 'Movement name is required'
    isValid = false
  }

  if (!movement.value.type) {
    validationErrors.value.type = 'Movement type is required'
    isValid = false
  }

  if (!movement.value.description || movement.value.description.trim() === '') {
    validationErrors.value.description = 'Description is required'
    isValid = false
  }

  return isValid
}

// Save movement
async function saveMovement() {
  if (!validateMovement()) return

  saving.value = true
  error.value = ''
  successMessage.value = ''

  try {
    // Build structured data
    const structuredData = {
      difficulty: movement.value.difficulty || null,
      equipment: movement.value.equipment?.trim() || null,
      primaryMuscles: movement.value.primaryMuscles?.trim() || null,
      description: movement.value.description.trim(),
      coachingCues: movement.value.coachingCues?.trim() || null,
      videoUrl: movement.value.videoUrl?.trim() || null,
      scalingOptions: movement.value.scalingOptions?.trim() || null
    }

    // Encode structured data with marker
    const descriptionField = '__STRUCTURED__' + JSON.stringify(structuredData)

    const payload = {
      name: movement.value.name.trim(),
      type: movement.value.type,
      description: descriptionField
    }

    let response
    if (isEditMode.value) {
      response = await axios.put(`/api/movements/${route.params.id}`, payload)
      successMessage.value = 'Movement updated successfully!'
    } else {
      response = await axios.post('/api/movements', payload)
      successMessage.value = 'Movement created successfully!'
    }

    // If in selection mode, return to caller with new movement selected
    if (selectionMode.value && response.data.movement) {
      setTimeout(() => {
        router.push({
          path: returnPath.value,
          query: { selectedMovement: response.data.movement.id }
        })
      }, 1000)
    } else if (!isEditMode.value) {
      // For new movements in non-selection mode, navigate to library
      setTimeout(() => {
        router.push('/movements')
      }, 1500)
    }
  } catch (err) {
    console.error('Failed to save movement:', err)
    if (err.response?.data?.message) {
      error.value = err.response.data.message
    } else {
      error.value = 'Failed to save movement. Please try again.'
    }
  } finally {
    saving.value = false
  }
}

// Confirm delete
function confirmDelete() {
  deleteDialog.value = true
}

// Delete movement
async function deleteMovement() {
  deleting.value = true
  error.value = ''

  try {
    await axios.delete(`/api/movements/${route.params.id}`)
    successMessage.value = 'Movement deleted successfully!'
    setTimeout(() => {
      router.push(returnPath.value)
    }, 1000)
  } catch (err) {
    console.error('Failed to delete movement:', err)
    error.value = 'Failed to delete movement. Please try again.'
    deleteDialog.value = false
  } finally {
    deleting.value = false
  }
}

// Handle back navigation
function handleBack() {
  const hasChanges = movement.value.name ||
                     movement.value.description ||
                     movement.value.difficulty ||
                     movement.value.equipment ||
                     movement.value.primaryMuscles ||
                     movement.value.coachingCues ||
                     movement.value.videoUrl ||
                     movement.value.scalingOptions

  if (hasChanges) {
    if (confirm('You have unsaved changes. Are you sure you want to leave?')) {
      router.push(returnPath.value)
    }
  } else {
    router.push(returnPath.value)
  }
}

// Initialize
onMounted(async () => {
  await loadMovement()
})
</script>
