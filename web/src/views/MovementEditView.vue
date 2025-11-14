<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" density="compact" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-btn icon="mdi-arrow-left" color="white" size="small" @click="handleBack" />
      <v-toolbar-title class="text-white font-weight-bold">
        {{ isEditMode ? 'Edit Movement' : 'Create Custom Movement' }}
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

      <!-- Info Alert -->
      <v-alert
        v-if="!isEditMode"
        type="info"
        variant="tonal"
        class="mb-3"
      >
        <div class="text-body-2">
          <strong>Create a custom movement</strong> to track exercises not in the standard library.
          Once created, it will be available in all your workouts and WODs.
        </div>
      </v-alert>

      <!-- Basic Information Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Movement Details</h2>

        <v-text-field
          v-model="movement.name"
          label="Movement Name"
          placeholder="e.g., Banded Pull-ups, Assault Bike"
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

        <v-select
          v-model="movement.movement_type"
          :items="movementTypes"
          label="Movement Type"
          variant="outlined"
          density="compact"
          rounded="lg"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-tag</v-icon>
          </template>
        </v-select>

        <v-select
          v-model="movement.primary_muscle_group"
          :items="muscleGroups"
          label="Primary Muscle Group"
          variant="outlined"
          density="compact"
          rounded="lg"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-arm-flex</v-icon>
          </template>
        </v-select>

        <v-textarea
          v-model="movement.description"
          label="Description (Optional)"
          placeholder="Add any notes, form cues, or scaling options"
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

        <v-row dense>
          <v-col cols="6">
            <v-switch
              v-model="movement.is_weighted"
              label="Weighted Movement"
              color="#00bcd4"
              density="compact"
              hide-details
            />
          </v-col>
          <v-col cols="6">
            <v-switch
              v-model="movement.is_timed"
              label="Timed Movement"
              color="#00bcd4"
              density="compact"
              hide-details
            />
          </v-col>
        </v-row>
      </v-card>

      <!-- Video/Image Card (Optional) -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Media (Optional)</h2>

        <v-text-field
          v-model="movement.video_url"
          label="Video URL"
          placeholder="https://youtube.com/watch?v=..."
          variant="outlined"
          density="compact"
          rounded="lg"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-video</v-icon>
          </template>
        </v-text-field>

        <v-text-field
          v-model="movement.demo_image_url"
          label="Demo Image URL"
          placeholder="https://example.com/image.jpg"
          variant="outlined"
          density="compact"
          rounded="lg"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-image</v-icon>
          </template>
        </v-text-field>
      </v-card>

      <!-- Equipment Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Equipment</h2>

        <v-combobox
          v-model="movement.equipment"
          :items="commonEquipment"
          label="Equipment Required"
          placeholder="Select or type custom equipment"
          variant="outlined"
          density="compact"
          rounded="lg"
          multiple
          chips
          closable-chips
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-tools</v-icon>
          </template>
        </v-combobox>
      </v-card>

      <!-- Standard Weights Card -->
      <v-card
        v-if="movement.is_weighted"
        elevation="0"
        rounded="lg"
        class="pa-3 mb-3"
        style="background: white"
      >
        <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Standard Weights (Optional)</h2>
        <p class="text-caption mb-3" style="color: #666">
          Define typical RX and scaled weights for this movement
        </p>

        <v-row dense>
          <v-col cols="6">
            <v-text-field
              v-model.number="movement.rx_weight_male"
              label="RX Male (lbs)"
              type="number"
              variant="outlined"
              density="compact"
              rounded="lg"
              min="0"
            />
          </v-col>
          <v-col cols="6">
            <v-text-field
              v-model.number="movement.rx_weight_female"
              label="RX Female (lbs)"
              type="number"
              variant="outlined"
              density="compact"
              rounded="lg"
              min="0"
            />
          </v-col>
        </v-row>

        <v-row dense>
          <v-col cols="6">
            <v-text-field
              v-model.number="movement.scaled_weight_male"
              label="Scaled Male (lbs)"
              type="number"
              variant="outlined"
              density="compact"
              rounded="lg"
              min="0"
            />
          </v-col>
          <v-col cols="6">
            <v-text-field
              v-model.number="movement.scaled_weight_female"
              label="Scaled Female (lbs)"
              type="number"
              variant="outlined"
              density="compact"
              rounded="lg"
              min="0"
            />
          </v-col>
        </v-row>
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
          v-if="isEditMode && movement.is_custom"
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

        <v-alert
          v-if="isEditMode && !movement.is_custom"
          type="warning"
          variant="tonal"
          class="mt-2"
        >
          <div class="text-caption">
            Standard movements cannot be deleted. You can only edit custom movements you created.
          </div>
        </v-alert>
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
          <p class="text-caption mt-2" style="color: #999">
            Note: This will not delete historical workout data using this movement.
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
  movement_type: 'weightlifting',
  primary_muscle_group: 'full_body',
  description: '',
  is_weighted: true,
  is_timed: false,
  is_custom: true,
  equipment: [],
  video_url: '',
  demo_image_url: '',
  rx_weight_male: null,
  rx_weight_female: null,
  scaled_weight_male: null,
  scaled_weight_female: null
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
  { title: 'Bodyweight', value: 'bodyweight' },
  { title: 'Cardio', value: 'cardio' },
  { title: 'Olympic Lifting', value: 'olympic_lifting' },
  { title: 'Powerlifting', value: 'powerlifting' },
  { title: 'Plyometric', value: 'plyometric' },
  { title: 'Mobility', value: 'mobility' }
]

const muscleGroups = [
  { title: 'Full Body', value: 'full_body' },
  { title: 'Upper Body', value: 'upper_body' },
  { title: 'Lower Body', value: 'lower_body' },
  { title: 'Core', value: 'core' },
  { title: 'Back', value: 'back' },
  { title: 'Chest', value: 'chest' },
  { title: 'Shoulders', value: 'shoulders' },
  { title: 'Arms', value: 'arms' },
  { title: 'Legs', value: 'legs' },
  { title: 'Glutes', value: 'glutes' }
]

const commonEquipment = [
  'Barbell',
  'Dumbbells',
  'Kettlebell',
  'Pull-up Bar',
  'Rings',
  'Rower',
  'Assault Bike',
  'Jump Rope',
  'Box',
  'Medicine Ball',
  'Slam Ball',
  'Resistance Bands',
  'TRX',
  'Sled',
  'None (Bodyweight)'
]

// Computed
const isEditMode = computed(() => !!route.params.id)

// Load existing movement if editing
async function loadMovement() {
  if (!route.params.id) return

  try {
    const response = await axios.get(`/api/movements/${route.params.id}`)
    const data = response.data.movement

    movement.value = {
      name: data.name || '',
      movement_type: data.type || 'weightlifting',
      primary_muscle_group: data.primary_muscle_group || 'full_body',
      description: data.description || '',
      is_weighted: data.is_weighted || false,
      is_timed: data.is_timed || false,
      is_custom: data.is_custom || false,
      equipment: data.equipment || [],
      video_url: data.video_url || '',
      demo_image_url: data.demo_image_url || '',
      rx_weight_male: data.rx_weight_male || null,
      rx_weight_female: data.rx_weight_female || null,
      scaled_weight_male: data.scaled_weight_male || null,
      scaled_weight_female: data.scaled_weight_female || null
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

  return isValid
}

// Save movement
async function saveMovement() {
  if (!validateMovement()) return

  saving.value = true
  error.value = ''
  successMessage.value = ''

  try {
    const payload = {
      name: movement.value.name.trim(),
      type: movement.value.movement_type,
      primary_muscle_group: movement.value.primary_muscle_group,
      description: movement.value.description?.trim() || null,
      is_weighted: movement.value.is_weighted,
      is_timed: movement.value.is_timed,
      equipment: movement.value.equipment,
      video_url: movement.value.video_url?.trim() || null,
      demo_image_url: movement.value.demo_image_url?.trim() || null,
      rx_weight_male: movement.value.rx_weight_male || null,
      rx_weight_female: movement.value.rx_weight_female || null,
      scaled_weight_male: movement.value.scaled_weight_male || null,
      scaled_weight_female: movement.value.scaled_weight_female || null
    }

    if (isEditMode.value) {
      await axios.put(`/api/movements/${route.params.id}`, payload)
      successMessage.value = 'Movement updated successfully!'
    } else {
      const response = await axios.post('/api/movements', payload)
      successMessage.value = 'Movement created successfully!'
      // Redirect to edit mode with new ID
      setTimeout(() => {
        router.push(`/movements/${response.data.movement.id}/edit`)
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
  if (!movement.value.is_custom) {
    error.value = 'Standard movements cannot be deleted'
    return
  }
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
      router.push('/workouts')
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
  if (movement.value.name) {
    if (confirm('You have unsaved changes. Are you sure you want to leave?')) {
      router.back()
    }
  } else {
    router.back()
  }
}

// Initialize
onMounted(async () => {
  await loadMovement()
})
</script>
