<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">

    <v-container class="pa-2" style="margin-bottom: 70px">
      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable @click:close="error = null" class="mb-4">
        {{ error }}
      </v-alert>

      <!-- Success Alert -->
      <v-alert v-if="success" type="success" closable @click:close="success = null" class="mb-4">
        {{ success }}
      </v-alert>

      <!-- Tabs for Standard vs Custom WODs -->
      <v-tabs v-model="activeWODTab" color="#00bcd4" class="mb-3" grow>
        <v-tab value="standard">
          <v-icon start>mdi-star</v-icon>
          Standard
        </v-tab>
        <v-tab value="custom">
          <v-icon start>mdi-account</v-icon>
          My WODs
        </v-tab>
      </v-tabs>

      <div>
        <!-- Loading State -->
        <div v-if="loading" class="text-center py-8">
          <v-progress-circular indeterminate color="#00bcd4" size="64" />
          <p class="mt-4 text-medium-emphasis">Loading WODs...</p>
        </div>

        <!-- Empty State -->
        <v-card
          v-else-if="!loading && displayedWODs.length === 0"
          elevation="0"
          rounded="lg"
          class="pa-8 text-center"
          style="background: white"
        >
          <v-icon size="64" color="#ccc">mdi-fire</v-icon>
          <p class="text-h6 mt-4" style="color: #2c3e50">
            {{ activeWODTab === 'standard' ? 'No standard WODs found' : 'No custom WODs yet' }}
          </p>
          <p class="text-body-2" style="color: #666">
            {{
              activeWODTab === 'standard'
                ? 'Standard CrossFit benchmarks will be seeded on first use'
                : 'Create your first custom WOD'
            }}
          </p>
          <v-btn
            v-if="activeWODTab === 'custom'"
            color="#00bcd4"
            class="mt-4"
            prepend-icon="mdi-plus"
            @click="$router.push('/wods/create')"
            rounded="lg"
            style="text-transform: none; font-weight: 600"
          >
            Create WOD
          </v-btn>
        </v-card>

        <!-- WODs List -->
        <div v-else>
          <v-card
            v-for="wod in displayedWODs"
            :key="wod.id"
            elevation="0"
            rounded="lg"
            class="mb-2 pa-3"
            style="background: white; cursor: pointer"
            @click="selectWOD(wod)"
          >
            <div class="d-flex align-center mb-2">
              <v-icon
                :color="wod.is_standard ? '#ffc107' : '#00bcd4'"
                class="mr-2"
                size="small"
              >
                {{ wod.is_standard ? 'mdi-star' : 'mdi-account' }}
              </v-icon>
              <div class="flex-grow-1">
                <div class="font-weight-bold text-body-1" style="color: #1a1a1a">
                  {{ wod.name }}
                </div>
                <div v-if="wod.type" class="text-caption" style="color: #666">
                  {{ formatWODType(wod.type) }}
                  <span v-if="wod.regime"> • {{ formatRegime(wod.regime) }}</span>
                </div>
              </div>
            </div>

            <!-- WOD Description -->
            <div v-if="wod.description" class="ml-7 mb-2 text-caption" style="color: #666">
              {{ truncateText(wod.description, 120) }}
            </div>

            <!-- WOD Metadata Chips -->
            <div class="ml-7 d-flex flex-wrap gap-1">
              <v-chip v-if="wod.score_type" size="x-small" color="#e0e0e0">
                <v-icon start size="x-small">mdi-timer</v-icon>
                {{ formatScoreType(wod.score_type) }}
              </v-chip>
              <v-chip v-if="wod.source" size="x-small" color="#e0e0e0">
                <v-icon start size="x-small">mdi-tag</v-icon>
                {{ wod.source }}
              </v-chip>
            </div>

            <!-- Action buttons for custom WODs -->
            <div v-if="!wod.is_standard" class="mt-2 d-flex gap-2">
              <v-btn
                size="x-small"
                variant="text"
                prepend-icon="mdi-pencil"
                @click.stop="editWOD(wod)"
                style="text-transform: none"
              >
                Edit
              </v-btn>
              <v-btn
                size="x-small"
                variant="text"
                prepend-icon="mdi-delete"
                color="error"
                @click.stop="deleteWOD(wod.id)"
                style="text-transform: none"
              >
                Delete
              </v-btn>
            </div>
          </v-card>
        </div>
      </div>
    </v-container>

    <!-- Create/Edit WOD Dialog -->
    <v-dialog v-model="createWODDialog" max-width="600" scrollable>
      <v-card>
        <v-card-title class="bg-primary text-white">
          <v-icon start>{{ editingWOD ? 'mdi-pencil' : 'mdi-plus' }}</v-icon>
          {{ editingWOD ? 'Edit WOD' : 'Create WOD' }}
        </v-card-title>
        <v-card-text class="pt-4">
          <v-text-field
            v-model="wodForm.name"
            label="WOD Name *"
            placeholder="e.g., My Custom Fran, Saturday Special"
            variant="outlined"
            density="comfortable"
            prepend-inner-icon="mdi-fire"
            required
          />

          <v-select
            v-model="wodForm.type"
            :items="wodTypes"
            label="WOD Type"
            variant="outlined"
            density="comfortable"
            prepend-inner-icon="mdi-tag"
            placeholder="Select type (optional)"
          />

          <v-select
            v-model="wodForm.regime"
            :items="regimes"
            label="Regime"
            variant="outlined"
            density="comfortable"
            prepend-inner-icon="mdi-dumbbell"
            placeholder="Select regime (optional)"
          />

          <v-select
            v-model="wodForm.score_type"
            :items="scoreTypes"
            label="Score Type"
            variant="outlined"
            density="comfortable"
            prepend-inner-icon="mdi-timer"
            placeholder="Select score type (optional)"
          />

          <v-text-field
            v-model="wodForm.source"
            label="Source"
            placeholder="e.g., CrossFit, Other Coach, Self-created"
            variant="outlined"
            density="comfortable"
            prepend-inner-icon="mdi-book"
          />

          <v-textarea
            v-model="wodForm.description"
            label="Description / Instructions *"
            placeholder="Describe the WOD in detail (e.g., 21-15-9 reps for time of Thrusters and Pull-ups)"
            variant="outlined"
            density="comfortable"
            rows="4"
            prepend-inner-icon="mdi-text"
            required
          />

          <v-text-field
            v-model="wodForm.url"
            label="Reference URL"
            placeholder="https://www.crossfit.com/workout/..."
            variant="outlined"
            density="comfortable"
            prepend-inner-icon="mdi-link"
            type="url"
          />

          <v-textarea
            v-model="wodForm.notes"
            label="Additional Notes"
            placeholder="Any additional notes or modifications..."
            variant="outlined"
            density="comfortable"
            rows="2"
            prepend-inner-icon="mdi-note-text"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="cancelWODForm" style="text-transform: none">
            Cancel
          </v-btn>
          <v-btn
            color="#00bcd4"
            :loading="creating"
            @click="saveWOD"
            style="text-transform: none"
          >
            {{ editingWOD ? 'Update' : 'Create' }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- WOD Detail Dialog -->
    <v-dialog v-model="detailDialog" max-width="600" scrollable>
      <v-card v-if="selectedWOD">
        <v-card-title class="bg-primary text-white">
          <v-icon start>mdi-fire</v-icon>
          {{ selectedWOD.name }}
        </v-card-title>
        <v-card-text class="pt-4">
          <!-- WOD Type and Regime -->
          <div v-if="selectedWOD.type || selectedWOD.regime" class="mb-3">
            <v-chip v-if="selectedWOD.type" color="teal" class="mr-2">
              {{ formatWODType(selectedWOD.type) }}
            </v-chip>
            <v-chip v-if="selectedWOD.regime" color="#00bcd4">
              {{ formatRegime(selectedWOD.regime) }}
            </v-chip>
          </div>

          <!-- Score Type and Source -->
          <div v-if="selectedWOD.score_type || selectedWOD.source" class="mb-3">
            <v-chip v-if="selectedWOD.score_type" size="small" color="#e0e0e0" class="mr-2">
              <v-icon start size="x-small">mdi-timer</v-icon>
              {{ formatScoreType(selectedWOD.score_type) }}
            </v-chip>
            <v-chip v-if="selectedWOD.source" size="small" color="#e0e0e0">
              <v-icon start size="x-small">mdi-tag</v-icon>
              {{ selectedWOD.source }}
            </v-chip>
          </div>

          <!-- Description -->
          <div v-if="selectedWOD.description" class="mb-3">
            <div class="text-caption font-weight-bold mb-1" style="color: #666">
              DESCRIPTION
            </div>
            <div class="text-body-2" style="color: #1a1a1a; white-space: pre-wrap">
              {{ selectedWOD.description }}
            </div>
          </div>

          <!-- Reference URL -->
          <div v-if="selectedWOD.url" class="mb-3">
            <div class="text-caption font-weight-bold mb-1" style="color: #666">
              REFERENCE
            </div>
            <a :href="selectedWOD.url" target="_blank" rel="noopener noreferrer" class="text-body-2">
              {{ selectedWOD.url }}
            </a>
          </div>

          <!-- Notes -->
          <div v-if="selectedWOD.notes" class="mb-3">
            <div class="text-caption font-weight-bold mb-1" style="color: #666">
              NOTES
            </div>
            <div class="text-body-2" style="color: #1a1a1a; white-space: pre-wrap">
              {{ selectedWOD.notes }}
            </div>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="detailDialog = false" style="text-transform: none">
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Bottom Navigation -->
    <v-bottom-navigation
      v-model="activeTab"
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
      <v-btn value="log" to="/dashboard?open=quick-log" style="position: relative; bottom: 20px">
        <v-avatar color="teal" size="56" style="box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2)">
          <v-icon color="white" size="32">mdi-plus</v-icon>
        </v-avatar>
      </v-btn>
      <v-btn value="workouts" to="/workouts">
        <v-icon>mdi-dumbbell</v-icon>
        <span style="font-size: 10px">Templates</span>
      </v-btn>
      <v-btn value="profile" to="/profile">
        <v-icon>mdi-account</v-icon>
        <span style="font-size: 10px">Profile</span>
      </v-btn>
    </v-bottom-navigation>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'

const router = useRouter()
const activeTab = ref(null) // No default active tab on WODs page
const activeWODTab = ref('standard')

// State
const standardWODs = ref([])
const customWODs = ref([])
const loading = ref(false)
const creating = ref(false)
const error = ref(null)
const success = ref(null)
const createWODDialog = ref(false)
const detailDialog = ref(false)
const selectedWOD = ref(null)
const editingWOD = ref(null)

// WOD form
const wodForm = ref({
  name: '',
  type: null,
  regime: null,
  score_type: null,
  source: '',
  description: '',
  url: '',
  notes: ''
})

// WOD Type options
const wodTypes = [
  { title: 'Benchmark', value: 'benchmark' },
  { title: 'Hero', value: 'hero' },
  { title: 'Girl', value: 'girl' },
  { title: 'Notables', value: 'notables' },
  { title: 'Games', value: 'games' },
  { title: 'Endurance', value: 'endurance' },
  { title: 'Self-created', value: 'self-created' }
]

// Regime options
const regimes = [
  { title: 'EMOM (Every Minute On the Minute)', value: 'emom' },
  { title: 'AMRAP (As Many Rounds As Possible)', value: 'amrap' },
  { title: 'For Time', value: 'for_time' },
  { title: 'Tabata', value: 'tabata' },
  { title: 'Chipper', value: 'chipper' },
  { title: 'Ladder', value: 'ladder' },
  { title: 'Death by...', value: 'death_by' }
]

// Score Type options
const scoreTypes = [
  { title: 'Time (HH:MM:SS)', value: 'time' },
  { title: 'Rounds + Reps', value: 'rounds_reps' },
  { title: 'Max Weight', value: 'max_weight' },
  { title: 'Total Reps', value: 'total_reps' },
  { title: 'Max Reps', value: 'max_reps' }
]

// Computed property for displayed WODs based on active tab
const displayedWODs = computed(() => {
  return activeWODTab.value === 'standard' ? standardWODs.value : customWODs.value
})

// Fetch WODs from API
async function fetchWODs() {
  loading.value = true
  error.value = null

  try {
    // Fetch standard WODs
    const standardResponse = await axios.get('/api/wods/standard')
    standardWODs.value = standardResponse.data.wods || []

    // Fetch user's custom WODs
    const customResponse = await axios.get('/api/wods/my-wods')
    customWODs.value = customResponse.data.wods || []

    console.log('Fetched WODs:', {
      standard: standardWODs.value.length,
      custom: customWODs.value.length
    })
  } catch (err) {
    console.error('Failed to fetch WODs:', err)
    if (err.response) {
      error.value = err.response.data?.message || `Error ${err.response.status}`
    } else if (err.request) {
      error.value = 'No response from server. Is the backend running?'
    } else {
      error.value = 'Failed to fetch WODs'
    }
  } finally {
    loading.value = false
  }
}

// Create or update WOD
async function saveWOD() {
  if (!wodForm.value.name.trim() || !wodForm.value.description.trim()) {
    error.value = 'WOD name and description are required'
    return
  }

  creating.value = true
  error.value = null
  success.value = null

  try {
    const payload = {
      name: wodForm.value.name.trim(),
      type: wodForm.value.type || null,
      regime: wodForm.value.regime || null,
      score_type: wodForm.value.score_type || null,
      source: wodForm.value.source.trim() || null,
      description: wodForm.value.description.trim(),
      url: wodForm.value.url.trim() || null,
      notes: wodForm.value.notes.trim() || null
    }

    if (editingWOD.value) {
      // Update existing WOD
      const response = await axios.put(`/api/wods/${editingWOD.value.id}`, payload)

      // Update in custom WODs list
      const index = customWODs.value.findIndex(w => w.id === editingWOD.value.id)
      if (index !== -1) {
        customWODs.value[index] = response.data.wod
      }

      success.value = 'WOD updated successfully!'
    } else {
      // Create new WOD
      const response = await axios.post('/api/wods', payload)

      // Add to custom WODs list
      customWODs.value.unshift(response.data.wod)

      success.value = 'WOD created successfully!'

      // Switch to custom tab to show new WOD
      activeWODTab.value = 'custom'
    }

    // Reset form and close dialog
    cancelWODForm()
  } catch (err) {
    console.error('Failed to save WOD:', err)
    error.value = err.response?.data?.message || 'Failed to save WOD'
  } finally {
    creating.value = false
  }
}

// Select WOD to view details
function selectWOD(wod) {
  selectedWOD.value = wod
  detailDialog.value = true
}

// Edit WOD
function editWOD(wod) {
  router.push(`/wods/${wod.id}/edit`)
}

// Delete WOD
async function deleteWOD(wodId) {
  if (!confirm('Are you sure you want to delete this WOD?')) {
    return
  }

  try {
    await axios.delete(`/api/wods/${wodId}`)

    // Remove from list
    customWODs.value = customWODs.value.filter(w => w.id !== wodId)

    success.value = 'WOD deleted successfully!'
  } catch (err) {
    console.error('Failed to delete WOD:', err)
    error.value = err.response?.data?.message || 'Failed to delete WOD'
  }
}

// Cancel WOD form
function cancelWODForm() {
  wodForm.value = {
    name: '',
    type: null,
    regime: null,
    score_type: null,
    source: '',
    description: '',
    url: '',
    notes: ''
  }
  editingWOD.value = null
  createWODDialog.value = false
}

// Format WOD type for display
function formatWODType(type) {
  if (!type) return ''
  return type.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')
}

// Format regime for display
function formatRegime(regime) {
  if (!regime) return ''
  const formattedRegime = regime.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')

  // Add full form for common acronyms
  if (regime === 'emom') return 'EMOM'
  if (regime === 'amrap') return 'AMRAP'
  if (regime === 'for_time') return 'For Time'

  return formattedRegime
}

// Format score type for display
function formatScoreType(scoreType) {
  if (!scoreType) return ''
  if (scoreType === 'time') return 'Time'
  if (scoreType === 'rounds_reps') return 'Rounds + Reps'
  if (scoreType === 'max_weight') return 'Max Weight'
  if (scoreType === 'total_reps') return 'Total Reps'
  if (scoreType === 'max_reps') return 'Max Reps'
  return scoreType.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')
}

// Utility function to truncate text
function truncateText(text, maxLength) {
  if (!text || text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

// Load WODs on mount
onMounted(() => {
  fetchWODs()
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
