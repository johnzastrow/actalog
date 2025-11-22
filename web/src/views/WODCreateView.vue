<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">

    <v-container class="pa-2" style=" margin-bottom: 70px">
      <v-alert v-if="successMessage" type="success" closable @click:close="successMessage = ''" class="mb-3">
        {{ successMessage }}
      </v-alert>

      <v-alert v-if="error" type="error" closable @click:close="error = ''" class="mb-3">
        {{ error }}
      </v-alert>

      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">WOD Details</h2>

        <v-text-field
          v-model="wod.name"
          label="WOD Name"
          placeholder="e.g., Fran, Murph, My Custom WOD"
          variant="outlined"
          density="compact"
          rounded="lg"
          :error-messages="validationErrors.name"
          class="mb-2"
          required
          autofocus
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-fire</v-icon>
          </template>
        </v-text-field>

        <v-select
          v-model="wod.source"
          :items="sources"
          label="Source"
          variant="outlined"
          density="compact"
          rounded="lg"
          :error-messages="validationErrors.source"
          class="mb-2"
          required
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-source-branch</v-icon>
          </template>
        </v-select>

        <v-select
          v-model="wod.type"
          :items="wodTypes"
          label="WOD Type"
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
          v-model="wod.regime"
          :items="regimes"
          label="Regime"
          variant="outlined"
          density="compact"
          rounded="lg"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-clock-outline</v-icon>
          </template>
        </v-select>

        <v-select
          v-model="wod.score_type"
          :items="scoreTypes"
          label="Score Type"
          variant="outlined"
          density="compact"
          rounded="lg"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-trophy</v-icon>
          </template>
        </v-select>

        <v-textarea
          v-model="wod.description"
          label="Description"
          placeholder="Describe the WOD workout, movements, and instructions"
          variant="outlined"
          density="compact"
          rounded="lg"
          rows="6"
          auto-grow
          :error-messages="validationErrors.description"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-text</v-icon>
          </template>
        </v-textarea>

        <v-text-field
          v-model="wod.url"
          label="Video URL (Optional)"
          placeholder="https://youtube.com/..."
          variant="outlined"
          density="compact"
          rounded="lg"
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-video</v-icon>
          </template>
        </v-text-field>

        <v-textarea
          v-model="wod.notes"
          label="Notes (Optional)"
          placeholder="Any additional notes or scaling options"
          variant="outlined"
          density="compact"
          rounded="lg"
          rows="3"
          auto-grow
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-note</v-icon>
          </template>
        </v-textarea>
      </v-card>

      <v-card elevation="0" rounded="lg" class="pa-3" style="background: white">
        <v-btn
          block
          color="#00bcd4"
          size="large"
          rounded="lg"
          :loading="saving"
          @click="saveWOD"
          style="text-transform: none; font-weight: 600"
        >
          <v-icon start>mdi-content-save</v-icon>
          {{ isEditMode ? 'Update WOD' : 'Create WOD' }}
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
          Delete WOD
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

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="text-h6" style="color: #e91e63">Delete WOD?</v-card-title>
        <v-card-text>
          <p style="color: #666">
            Are you sure you want to delete "{{ wod.name }}"? This action cannot be undone.
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="#e91e63" variant="flat" :loading="deleting" @click="deleteWOD">
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

const wod = ref({
  name: '',
  source: 'Self-recorded',
  type: 'Benchmark',
  regime: 'Fastest Time',
  score_type: 'Time (HH:MM:SS)',
  description: '',
  url: '',
  notes: ''
})

const saving = ref(false)
const deleting = ref(false)
const successMessage = ref('')
const error = ref('')
const validationErrors = ref({})
const deleteDialog = ref(false)

const sources = ['CrossFit', 'Other Coach', 'Self-recorded']
const wodTypes = ['Benchmark', 'Hero', 'Girl', 'Games', 'Notables', 'Endurance', 'Self-created']
const regimes = ['EMOM', 'AMRAP', 'Fastest Time', 'Slowest Round', 'Get Stronger', 'Skills']
const scoreTypes = ['Time (HH:MM:SS)', 'Rounds+Reps', 'Max Weight']

const isEditMode = computed(() => !!route.params.id)
const returnPath = computed(() => route.query.returnPath || '/wods')
const selectionMode = computed(() => route.query.select === 'true')

async function loadWOD() {
  if (!route.params.id) return

  try {
    const response = await axios.get(`/api/wods/${route.params.id}`)
    const data = response.data.wod || response.data

    wod.value = {
      name: data.name || '',
      source: data.source || 'Self-recorded',
      type: data.type || 'Benchmark',
      regime: data.regime || 'Fastest Time',
      score_type: data.score_type || 'Time',
      description: data.description || '',
      url: data.url || '',
      notes: data.notes || ''
    }
  } catch (err) {
    console.error('Failed to load WOD:', err)
    error.value = 'Failed to load WOD'
  }
}

function validateWOD() {
  validationErrors.value = {}
  let isValid = true

  if (!wod.value.name || wod.value.name.trim() === '') {
    validationErrors.value.name = 'WOD name is required'
    isValid = false
  }

  if (!wod.value.source || wod.value.source.trim() === '') {
    validationErrors.value.source = 'Source is required'
    isValid = false
  }

  if (!wod.value.description || wod.value.description.trim() === '') {
    validationErrors.value.description = 'Description is required'
    isValid = false
  }

  return isValid
}

async function saveWOD() {
  if (!validateWOD()) return

  saving.value = true
  error.value = ''
  successMessage.value = ''

  try {
    const payload = {
      name: wod.value.name.trim(),
      source: wod.value.source.trim(),
      type: wod.value.type,
      regime: wod.value.regime,
      score_type: wod.value.score_type,
      description: wod.value.description.trim(),
      url: wod.value.url?.trim() || null,
      notes: wod.value.notes?.trim() || null
    }

    let response
    if (isEditMode.value) {
      response = await axios.put(`/api/wods/${route.params.id}`, payload)
      successMessage.value = 'WOD updated successfully!'
    } else {
      response = await axios.post('/api/wods', payload)
      successMessage.value = 'WOD created successfully!'
    }

    if (selectionMode.value && response.data.wod) {
      setTimeout(() => {
        router.push({
          path: returnPath.value,
          query: { selectedWOD: response.data.wod.id }
        })
      }, 1000)
    } else if (!isEditMode.value) {
      setTimeout(() => {
        router.push('/wods')
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

function confirmDelete() {
  deleteDialog.value = true
}

async function deleteWOD() {
  deleting.value = true
  error.value = ''

  try {
    await axios.delete(`/api/wods/${route.params.id}`)
    successMessage.value = 'WOD deleted successfully!'
    setTimeout(() => {
      router.push(returnPath.value)
    }, 1000)
  } catch (err) {
    console.error('Failed to delete WOD:', err)
    error.value = 'Failed to delete WOD. Please try again.'
    deleteDialog.value = false
  } finally {
    deleting.value = false
  }
}

function handleBack() {
  if (wod.value.name || wod.value.description || wod.value.source !== 'Self-recorded') {
    if (confirm('You have unsaved changes. Are you sure you want to leave?')) {
      router.push(returnPath.value)
    }
  } else {
    router.push(returnPath.value)
  }
}

onMounted(async () => {
  await loadWOD()
})
</script>
