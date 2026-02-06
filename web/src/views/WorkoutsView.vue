<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable class="mb-4" @click:close="error = null">
        {{ error }}
      </v-alert>

      <!-- Tabs for Standard vs Custom Templates -->
      <v-tabs v-model="activeTemplateTab" color="primary" class="mb-3" grow>
        <v-tab value="standard">
          <v-icon start>mdi-star</v-icon>
          Standard
        </v-tab>
        <v-tab value="custom">
          <v-icon start>mdi-account</v-icon>
          My Workouts
        </v-tab>
      </v-tabs>

      <!-- Quick Links -->
      <v-card elevation="0" rounded class="pa-2 mb-1" bg-color="surface">
        <h3 class="text-body-2 font-weight-bold mb-1 text-medium-emphasis">Quick Links</h3>
        <v-row dense>
          <v-col cols="6">
            <v-btn
              block
              
              color="primary"
              rounded
              style="text-transform: none"
              @click="$router.push('/wods')"
            >
              <v-icon start size="small">mdi-fire</v-icon>
              WOD Library
            </v-btn>
          </v-col>
          <v-col cols="6">
            <v-btn
              block
              
              color="primary"
              rounded
              style="text-transform: none"
              @click="$router.push('/movements')"
            >
              <v-icon start size="small">mdi-weight-lifter</v-icon>
              Movements
            </v-btn>
          </v-col>
        </v-row>
      </v-card>

      <!-- Create Template Button for My Templates Tab -->
      <v-card
        v-if="activeTemplateTab === 'custom' && !loading"
        elevation="0"
        rounded
        class="pa-2 mb-1"
        bg-color="surface"
      >
        <v-btn
          block
          color="primary"
          variant="flat"
          prepend-icon="mdi-plus"
          rounded
          style="text-transform: none; font-weight: 600"
          @click="$router.push('/workouts/templates/create')"
        >
          Create Workout
        </v-btn>
      </v-card>

      <div>
        <!-- Loading State -->
        <div v-if="loading" class="text-center py-8">
          <v-progress-circular indeterminate color="primary" size="64" />
          <p class="mt-4 text-medium-emphasis">Loading workouts...</p>
        </div>

        <!-- Empty State -->
        <v-card
          v-else-if="!loading && displayedTemplates.length === 0"
          elevation="0"
          rounded
          class="pa-3 text-center"
          bg-color="surface"
        >
          <v-icon size="64" color="surface-variant">mdi-clipboard-text-outline</v-icon>
          <p class="text-h6 mt-4 text-high-emphasis">
            {{ activeTemplateTab === 'standard' ? 'No standard workouts found' : 'No custom workouts yet' }}
          </p>
          <p class="text-body-2 text-medium-emphasis">
            {{
              activeTemplateTab === 'standard'
                ? 'Standard workouts will be seeded on first use'
                : 'Create your first custom workout'
            }}
          </p>
          <v-btn
            v-if="activeTemplateTab === 'custom'"
            color="primary"
            class="mt-4"
            prepend-icon="mdi-plus"
            rounded
            style="text-transform: none; font-weight: 600"
            @click="$router.push('/workouts/templates/create')"
          >
            Create Workout
          </v-btn>
        </v-card>

        <!-- Templates List -->
        <div v-else>
          <v-card
            v-for="template in displayedTemplates"
            :key="template.id"
            elevation="0"
            rounded
            class="mb-1 pa-2"
            style="background-color: rgb(var(--v-theme-surface)); cursor: pointer"
            @click="selectTemplate(template)"
          >
            <div class="d-flex align-center mb-1">
              <v-icon
                :color="template.created_by ? 'primary' : 'warning'"
                class="mr-2"
                size="small"
              >
                {{ template.created_by ? 'mdi-account' : 'mdi-star' }}
              </v-icon>
              <div class="flex-grow-1">
                <div class="font-weight-bold text-body-1" >
                  {{ template.name }}
                </div>
                <div v-if="template.intro_warmup" class="text-caption text-medium-emphasis">
                  {{ truncateText(template.intro_warmup, 60) }}
                </div>
                <div v-else-if="template.notes" class="text-caption text-medium-emphasis">
                  {{ truncateText(template.notes, 60) }}
                </div>
              </div>
              <v-btn
                icon="mdi-lightning-bolt"
                color="primary"
                variant="flat"
                size="small"
                style="background-color: rgb(var(--v-theme-primary))"
                @click.stop="logWorkoutFromTemplate(template.id)"
              >
                <v-icon color="white">mdi-lightning-bolt</v-icon>
                <v-tooltip activator="parent" location="top">Quick Log</v-tooltip>
              </v-btn>
            </div>

            <!-- Display movements count -->
            <div v-if="template.movements && template.movements.length > 0" class="ml-7 mt-2">
              <v-chip size="small" color="#e0e0e0" class="mr-1">
                <v-icon start size="x-small">mdi-weight-lifter</v-icon>
                {{ template.movements.length }} movement{{ template.movements.length > 1 ? 's' : '' }}
              </v-chip>
            </div>

            <!-- Display WODs count -->
            <div v-if="template.wods && template.wods.length > 0" class="ml-7 mt-1">
              <v-chip size="small" color="#e0e0e0" class="mr-1">
                <v-icon start size="x-small">mdi-fire</v-icon>
                {{ template.wods.length }} WOD{{ template.wods.length > 1 ? 's' : '' }}
              </v-chip>
            </div>

            <!-- Action buttons for custom templates -->
            <div v-if="template.created_by" class="mt-2 d-flex gap-2">
              <v-btn
                size="x-small"
                variant="text"
                prepend-icon="mdi-pencil"
                style="text-transform: none"
                @click.stop="editTemplate(template)"
              >
                Edit
              </v-btn>
              <v-btn
                size="x-small"
                variant="text"
                prepend-icon="mdi-delete"
                color="error"
                style="text-transform: none"
                @click.stop="deleteTemplate(template.id)"
              >
                Delete
              </v-btn>
            </div>
          </v-card>
        </div>
      </div>
    </v-container>

    <!-- Detail View Dialog -->
    <DetailViewDialog
      v-model="showDetailDialog"
      type="template"
      :id="selectedTemplateId"
    />

    <!-- Create Template Dialog -->
    <v-dialog v-model="createTemplateDialog" max-width="600">
      <v-card>
        <v-card-title class="bg-primary text-white">
          <v-icon start>mdi-plus</v-icon>
          Create Workout
        </v-card-title>
        <v-card-text class="pt-4">
          <v-text-field
            v-model="newTemplate.name"
            label="Workout Name"
            placeholder="e.g., Monday Strength, Hero WOD"

            density="comfortable"
            prepend-inner-icon="mdi-clipboard-text"
            required
          />
          <v-textarea
            v-model="newTemplate.notes"
            label="Description / Notes"
            placeholder="Describe the workout..."

            density="comfortable"
            rows="3"
            prepend-inner-icon="mdi-note-text"
          />
          <v-alert type="info" density="compact" class="mt-2">
            After creating the workout, you can add movements and WODs in the details view.
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn style="text-transform: none" @click="createTemplateDialog = false">
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            :loading="creating"
            style="text-transform: none"
            @click="createTemplate"
          >
            Create
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Bottom Navigation -->
    <v-bottom-navigation
      v-model="activeTab"
      grow
      style="position: fixed; bottom: 0; background-color: rgb(var(--v-theme-surface))"
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
        <v-avatar color="primary" size="56" style="box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2)">
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'
import DetailViewDialog from '@/components/DetailViewDialog.vue'

const router = useRouter()
const activeTab = ref('workouts')
const activeTemplateTab = ref('standard')

// State
const standardTemplates = ref([])
const customTemplates = ref([])
const loading = ref(false)
const creating = ref(false)
const error = ref(null)
const createTemplateDialog = ref(false)

// Detail Dialog state
const showDetailDialog = ref(false)
const selectedTemplateId = ref(null)

// New template form
const newTemplate = ref({
  name: '',
  notes: ''
})

// Computed property for displayed templates based on active tab
const displayedTemplates = computed(() => {
  return activeTemplateTab.value === 'standard' ? standardTemplates.value : customTemplates.value
})

// Fetch workout templates from API
async function fetchTemplates() {
  loading.value = true
  error.value = null

  try {
    // Fetch standard templates
    const standardResponse = await axios.get('/api/workouts/standard')
    standardTemplates.value = standardResponse.data.workouts || []

    // Fetch user's custom templates
    const customResponse = await axios.get('/api/workouts/my-templates')
    customTemplates.value = customResponse.data.workouts || []

    console.log('Fetched templates:', {
      standard: standardTemplates.value.length,
      custom: customTemplates.value.length
    })
  } catch (err) {
    console.error('Failed to fetch templates:', err)
    if (err.response) {
      error.value = err.response.data?.message || `Error ${err.response.status}`
    } else if (err.request) {
      error.value = 'No response from server. Is the backend running?'
    } else {
      error.value = 'Failed to fetch workouts'
    }
  } finally {
    loading.value = false
  }
}

// Create new workout
async function createTemplate() {
  if (!newTemplate.value.name.trim()) {
    error.value = 'Workout name is required'
    return
  }

  creating.value = true
  error.value = null

  try {
    const response = await axios.post('/api/workouts', {
      name: newTemplate.value.name.trim(),
      notes: newTemplate.value.notes.trim() || null
    })

    // Add to custom workouts list
    customTemplates.value.unshift(response.data.workout)

    // Reset form and close dialog
    newTemplate.value = { name: '', notes: '' }
    createTemplateDialog.value = false

    // Switch to custom tab to show new workout
    activeTemplateTab.value = 'custom'
  } catch (err) {
    console.error('Failed to create workout:', err)
    error.value = err.response?.data?.message || 'Failed to create workout'
  } finally {
    creating.value = false
  }
}

// Select template to view details
function selectTemplate(template) {
  selectedTemplateId.value = template.id
  showDetailDialog.value = true
}

// Log workout from template
function logWorkoutFromTemplate(templateId) {
  console.log('Log workout from template:', templateId)
  router.push(`/workouts/log?template=${templateId}`)
}

// Edit template
function editTemplate(template) {
  router.push(`/workouts/templates/${template.id}/edit`)
}

// Delete workout
async function deleteTemplate(templateId) {
  if (!confirm('Are you sure you want to delete this workout?')) {
    return
  }

  try {
    await axios.delete(`/api/templates/${templateId}`)

    // Remove from list
    customTemplates.value = customTemplates.value.filter(t => t.id !== templateId)
  } catch (err) {
    console.error('Failed to delete workout:', err)
    error.value = err.response?.data?.message || 'Failed to delete workout'
  }
}

// Utility function to truncate text
function truncateText(text, maxLength) {
  if (!text || text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

// Load templates on mount
onMounted(() => {
  fetchTemplates()
})
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
</style>
