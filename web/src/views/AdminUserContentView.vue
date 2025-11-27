<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <div class="d-flex align-center mb-4">
        <v-btn icon @click="$router.back()" class="mr-2">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <div>
          <h1 class="text-h5">User-Created Content</h1>
          <div class="text-body-2 text-medium-emphasis">View and promote user-created WODs, Movements, and Workouts to standard</div>
        </div>
      </div>

      <!-- Loading State -->
      <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />

      <!-- Error Alert -->
      <v-alert v-if="error" type="error" variant="tonal" closable @click:close="error = null" class="mb-4">
        {{ error }}
      </v-alert>

      <!-- Success Alert -->
      <v-alert v-if="successMessage" type="success" variant="tonal" closable @click:close="successMessage = null" class="mb-4">
        {{ successMessage }}
      </v-alert>

      <!-- Content Type Tabs -->
      <v-tabs v-model="activeTab" color="primary" class="mb-4">
        <v-tab value="wods">
          <v-icon start>mdi-lightning-bolt</v-icon>
          WODs ({{ wodCount }})
        </v-tab>
        <v-tab value="movements">
          <v-icon start>mdi-dumbbell</v-icon>
          Movements ({{ movementCount }})
        </v-tab>
        <v-tab value="workouts">
          <v-icon start>mdi-clipboard-list</v-icon>
          Workouts ({{ workoutCount }})
        </v-tab>
      </v-tabs>

      <!-- WODs Tab -->
      <v-window v-model="activeTab">
        <v-window-item value="wods">
          <v-card elevation="0" rounded="lg">
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2">mdi-lightning-bolt</v-icon>
              User-Created WODs
            </v-card-title>

            <v-divider />

            <!-- WOD Filters -->
            <div class="pa-4 d-flex flex-wrap gap-3">
              <v-text-field
                v-model="wodSearch"
                prepend-inner-icon="mdi-magnify"
                label="Search by name"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                style="max-width: 250px"
                @update:model-value="debounceLoadWODs"
              />
              <v-select
                v-model="wodScoreTypeFilter"
                :items="scoreTypeOptions"
                label="Score Type"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                style="max-width: 180px"
                @update:model-value="resetAndLoadWODs"
              />
              <v-text-field
                v-model="wodCreatorFilter"
                prepend-inner-icon="mdi-account"
                label="Creator email"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                style="max-width: 200px"
                @update:model-value="debounceLoadWODs"
              />
            </div>

            <v-divider />

            <v-data-table
              :headers="wodHeaders"
              :items="wods"
              :loading="loading"
              hide-default-footer
              class="elevation-0"
            >
              <!-- Name Column -->
              <template #item.name="{ item }">
                <div>
                  <div class="font-weight-medium">{{ item.name }}</div>
                  <div class="text-caption text-medium-emphasis">{{ item.type }} - {{ item.regime }}</div>
                </div>
              </template>

              <!-- Creator Column -->
              <template #item.creator="{ item }">
                <div class="d-flex align-center">
                  <v-avatar size="24" color="primary" class="mr-2">
                    <span class="text-white text-caption">{{ (item.creator_email || '?')[0].toUpperCase() }}</span>
                  </v-avatar>
                  <div>
                    <div class="text-body-2">{{ item.creator_name || 'Unknown' }}</div>
                    <div class="text-caption text-medium-emphasis">{{ item.creator_email }}</div>
                  </div>
                </div>
              </template>

              <!-- Score Type Column -->
              <template #item.score_type="{ item }">
                <v-chip size="small" :color="getScoreTypeColor(item.score_type)">
                  {{ item.score_type || 'N/A' }}
                </v-chip>
              </template>

              <!-- Created At Column -->
              <template #item.created_at="{ item }">
                {{ formatDate(item.created_at) }}
              </template>

              <!-- Actions Column -->
              <template #item.actions="{ item }">
                <v-btn
                  icon
                  size="small"
                  variant="text"
                  color="primary"
                  @click="openCopyDialog('wod', item)"
                >
                  <v-icon>mdi-swap-horizontal</v-icon>
                  <v-tooltip activator="parent" location="top">Convert to Standard</v-tooltip>
                </v-btn>
                <v-btn
                  icon
                  size="small"
                  variant="text"
                  @click="viewDetails('wod', item)"
                >
                  <v-icon>mdi-eye</v-icon>
                  <v-tooltip activator="parent" location="top">View Details</v-tooltip>
                </v-btn>
              </template>
            </v-data-table>

            <!-- Pagination -->
            <v-divider />
            <div class="d-flex align-center justify-space-between pa-4">
              <div class="text-body-2 text-medium-emphasis">
                Showing {{ wodOffset + 1 }} to {{ Math.min(wodOffset + limit, wodCount) }} of {{ wodCount }} WODs
              </div>
              <div class="d-flex gap-2">
                <v-btn variant="outlined" size="small" :disabled="wodOffset === 0" @click="previousWodPage">
                  <v-icon start>mdi-chevron-left</v-icon>
                  Previous
                </v-btn>
                <v-btn variant="outlined" size="small" :disabled="wodOffset + limit >= wodCount" @click="nextWodPage">
                  Next
                  <v-icon end>mdi-chevron-right</v-icon>
                </v-btn>
              </div>
            </div>
          </v-card>
        </v-window-item>

        <!-- Movements Tab -->
        <v-window-item value="movements">
          <v-card elevation="0" rounded="lg">
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2">mdi-dumbbell</v-icon>
              User-Created Movements
            </v-card-title>

            <v-divider />

            <!-- Movement Filters -->
            <div class="pa-4 d-flex flex-wrap gap-3">
              <v-text-field
                v-model="movementSearch"
                prepend-inner-icon="mdi-magnify"
                label="Search by name"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                style="max-width: 250px"
                @update:model-value="debounceLoadMovements"
              />
              <v-select
                v-model="movementTypeFilter"
                :items="movementTypeOptions"
                label="Type"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                style="max-width: 160px"
                @update:model-value="resetAndLoadMovements"
              />
              <v-text-field
                v-model="movementCreatorFilter"
                prepend-inner-icon="mdi-account"
                label="Creator email"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                style="max-width: 200px"
                @update:model-value="debounceLoadMovements"
              />
            </div>

            <v-divider />

            <v-data-table
              :headers="movementHeaders"
              :items="movements"
              :loading="loading"
              hide-default-footer
              class="elevation-0"
            >
              <!-- Name Column -->
              <template #item.name="{ item }">
                <div>
                  <div class="font-weight-medium">{{ item.name }}</div>
                  <div class="text-caption text-medium-emphasis">{{ item.description || 'No description' }}</div>
                </div>
              </template>

              <!-- Creator Column -->
              <template #item.creator="{ item }">
                <div class="d-flex align-center">
                  <v-avatar size="24" color="primary" class="mr-2">
                    <span class="text-white text-caption">{{ (item.creator_email || '?')[0].toUpperCase() }}</span>
                  </v-avatar>
                  <div>
                    <div class="text-body-2">{{ item.creator_name || 'Unknown' }}</div>
                    <div class="text-caption text-medium-emphasis">{{ item.creator_email }}</div>
                  </div>
                </div>
              </template>

              <!-- Type Column -->
              <template #item.type="{ item }">
                <v-chip size="small" :color="getMovementTypeColor(item.type)">
                  {{ item.type }}
                </v-chip>
              </template>

              <!-- Created At Column -->
              <template #item.created_at="{ item }">
                {{ formatDate(item.created_at) }}
              </template>

              <!-- Actions Column -->
              <template #item.actions="{ item }">
                <v-btn
                  icon
                  size="small"
                  variant="text"
                  color="primary"
                  @click="openCopyDialog('movement', item)"
                >
                  <v-icon>mdi-swap-horizontal</v-icon>
                  <v-tooltip activator="parent" location="top">Convert to Standard</v-tooltip>
                </v-btn>
                <v-btn
                  icon
                  size="small"
                  variant="text"
                  @click="viewDetails('movement', item)"
                >
                  <v-icon>mdi-eye</v-icon>
                  <v-tooltip activator="parent" location="top">View Details</v-tooltip>
                </v-btn>
              </template>
            </v-data-table>

            <!-- Movement Pagination -->
            <v-divider />
            <div class="d-flex align-center justify-space-between pa-4">
              <div class="text-body-2 text-medium-emphasis">
                Showing {{ movementOffset + 1 }} to {{ Math.min(movementOffset + limit, movementCount) }} of {{ movementCount }} movements
              </div>
              <div class="d-flex gap-2">
                <v-btn variant="outlined" size="small" :disabled="movementOffset === 0" @click="previousMovementPage">
                  <v-icon start>mdi-chevron-left</v-icon>
                  Previous
                </v-btn>
                <v-btn variant="outlined" size="small" :disabled="movementOffset + limit >= movementCount" @click="nextMovementPage">
                  Next
                  <v-icon end>mdi-chevron-right</v-icon>
                </v-btn>
              </div>
            </div>
          </v-card>
        </v-window-item>

        <!-- Workouts Tab -->
        <v-window-item value="workouts">
          <v-card elevation="0" rounded="lg">
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2">mdi-clipboard-list</v-icon>
              User-Created Workout Templates
            </v-card-title>

            <v-divider />

            <!-- Workout Filters -->
            <div class="pa-4 d-flex flex-wrap gap-3">
              <v-text-field
                v-model="workoutSearch"
                prepend-inner-icon="mdi-magnify"
                label="Search by name/notes"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                style="max-width: 250px"
                @update:model-value="debounceLoadWorkouts"
              />
              <v-text-field
                v-model="workoutCreatorFilter"
                prepend-inner-icon="mdi-account"
                label="Creator email"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                style="max-width: 200px"
                @update:model-value="debounceLoadWorkouts"
              />
            </div>

            <v-divider />

            <v-data-table
              :headers="workoutHeaders"
              :items="workouts"
              :loading="loading"
              hide-default-footer
              class="elevation-0"
            >
              <!-- Name Column -->
              <template #item.name="{ item }">
                <div>
                  <div class="font-weight-medium">{{ item.name }}</div>
                  <div class="text-caption text-medium-emphasis">{{ item.notes || 'No notes' }}</div>
                </div>
              </template>

              <!-- Creator Column -->
              <template #item.creator="{ item }">
                <div class="d-flex align-center">
                  <v-avatar size="24" color="primary" class="mr-2">
                    <span class="text-white text-caption">{{ (item.creator_email || '?')[0].toUpperCase() }}</span>
                  </v-avatar>
                  <div>
                    <div class="text-body-2">{{ item.creator_name || 'Unknown' }}</div>
                    <div class="text-caption text-medium-emphasis">{{ item.creator_email }}</div>
                  </div>
                </div>
              </template>

              <!-- Created At Column -->
              <template #item.created_at="{ item }">
                {{ formatDate(item.created_at) }}
              </template>

              <!-- Actions Column -->
              <template #item.actions="{ item }">
                <v-btn
                  icon
                  size="small"
                  variant="text"
                  color="primary"
                  @click="openCopyDialog('workout', item)"
                >
                  <v-icon>mdi-swap-horizontal</v-icon>
                  <v-tooltip activator="parent" location="top">Convert to Standard</v-tooltip>
                </v-btn>
                <v-btn
                  icon
                  size="small"
                  variant="text"
                  @click="viewDetails('workout', item)"
                >
                  <v-icon>mdi-eye</v-icon>
                  <v-tooltip activator="parent" location="top">View Details</v-tooltip>
                </v-btn>
              </template>
            </v-data-table>

            <!-- Pagination -->
            <v-divider />
            <div class="d-flex align-center justify-space-between pa-4">
              <div class="text-body-2 text-medium-emphasis">
                Showing {{ workoutOffset + 1 }} to {{ Math.min(workoutOffset + limit, workoutCount) }} of {{ workoutCount }} workouts
              </div>
              <div class="d-flex gap-2">
                <v-btn variant="outlined" size="small" :disabled="workoutOffset === 0" @click="previousWorkoutPage">
                  <v-icon start>mdi-chevron-left</v-icon>
                  Previous
                </v-btn>
                <v-btn variant="outlined" size="small" :disabled="workoutOffset + limit >= workoutCount" @click="nextWorkoutPage">
                  Next
                  <v-icon end>mdi-chevron-right</v-icon>
                </v-btn>
              </div>
            </div>
          </v-card>
        </v-window-item>
      </v-window>

      <!-- Convert to Standard Dialog -->
      <v-dialog v-model="copyDialog" max-width="500">
        <v-card>
          <v-card-title class="d-flex align-center">
            <v-icon class="mr-2" color="primary">mdi-swap-horizontal</v-icon>
            Convert to Standard
          </v-card-title>
          <v-divider />
          <v-card-text>
            <div class="mb-4">
              Convert <strong>{{ selectedItem?.name }}</strong> to become a standard {{ copyType }}?
            </div>
            <v-text-field
              v-model="newName"
              label="New Name"
              hint="Enter a name for the standard version (leave empty to use original)"
              persistent-hint
              variant="outlined"
              :placeholder="selectedItem?.name"
            />
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn @click="copyDialog = false">Cancel</v-btn>
            <v-btn color="primary" @click="confirmCopy" :loading="actionLoading">
              <v-icon start>mdi-check</v-icon>
              Convert to Standard
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>

      <!-- Details Dialog -->
      <v-dialog v-model="detailsDialog" max-width="700">
        <v-card v-if="selectedItem">
          <v-card-title class="d-flex align-center">
            <v-icon class="mr-2">mdi-information</v-icon>
            {{ detailsType === 'wod' ? 'WOD' : detailsType === 'movement' ? 'Movement' : 'Workout' }} Details
          </v-card-title>
          <v-divider />
          <v-card-text>
            <v-list>
              <v-list-item>
                <v-list-item-title class="text-subtitle-2">Name</v-list-item-title>
                <v-list-item-subtitle>{{ selectedItem.name }}</v-list-item-subtitle>
              </v-list-item>

              <v-list-item v-if="detailsType === 'wod'">
                <v-list-item-title class="text-subtitle-2">Type</v-list-item-title>
                <v-list-item-subtitle>{{ selectedItem.type }} - {{ selectedItem.regime }}</v-list-item-subtitle>
              </v-list-item>

              <v-list-item v-if="detailsType === 'wod'">
                <v-list-item-title class="text-subtitle-2">Score Type</v-list-item-title>
                <v-list-item-subtitle>{{ selectedItem.score_type || 'Not specified' }}</v-list-item-subtitle>
              </v-list-item>

              <v-list-item v-if="detailsType === 'wod' && selectedItem.description">
                <v-list-item-title class="text-subtitle-2">Description</v-list-item-title>
                <v-list-item-subtitle style="white-space: pre-wrap">{{ selectedItem.description }}</v-list-item-subtitle>
              </v-list-item>

              <v-list-item v-if="detailsType === 'movement'">
                <v-list-item-title class="text-subtitle-2">Type</v-list-item-title>
                <v-list-item-subtitle>{{ selectedItem.type }}</v-list-item-subtitle>
              </v-list-item>

              <v-list-item v-if="selectedItem.description">
                <v-list-item-title class="text-subtitle-2">Description</v-list-item-title>
                <v-list-item-subtitle>{{ selectedItem.description }}</v-list-item-subtitle>
              </v-list-item>

              <v-list-item v-if="selectedItem.notes">
                <v-list-item-title class="text-subtitle-2">Notes</v-list-item-title>
                <v-list-item-subtitle>{{ selectedItem.notes }}</v-list-item-subtitle>
              </v-list-item>

              <v-list-item>
                <v-list-item-title class="text-subtitle-2">Created By</v-list-item-title>
                <v-list-item-subtitle>
                  {{ selectedItem.creator_name || 'Unknown' }} ({{ selectedItem.creator_email }})
                </v-list-item-subtitle>
              </v-list-item>

              <v-list-item>
                <v-list-item-title class="text-subtitle-2">Created At</v-list-item-title>
                <v-list-item-subtitle>{{ formatDateTime(selectedItem.created_at) }}</v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn @click="detailsDialog = false">Close</v-btn>
            <v-btn color="primary" @click="openCopyDialogFromDetails">
              <v-icon start>mdi-swap-horizontal</v-icon>
              Convert to Standard
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-container>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'

// State
const loading = ref(false)
const error = ref(null)
const successMessage = ref(null)
const actionLoading = ref(false)
const activeTab = ref('wods')

// WODs
const wods = ref([])
const wodCount = ref(0)
const wodOffset = ref(0)
const wodSearch = ref('')
const wodScoreTypeFilter = ref(null)
const wodCreatorFilter = ref('')

// Movements
const movements = ref([])
const movementCount = ref(0)
const movementOffset = ref(0)
const movementSearch = ref('')
const movementTypeFilter = ref(null)
const movementCreatorFilter = ref('')

// Workouts
const workouts = ref([])
const workoutCount = ref(0)
const workoutOffset = ref(0)
const workoutSearch = ref('')
const workoutCreatorFilter = ref('')

const limit = 50

// Filter Options
const scoreTypeOptions = [
  'Time (HH:MM:SS)',
  'Rounds+Reps',
  'Max Weight'
]

const movementTypeOptions = [
  'weightlifting',
  'bodyweight',
  'cardio',
  'gymnastics',
  'other'
]

// Debounce timers
let wodDebounceTimer = null
let movementDebounceTimer = null
let workoutDebounceTimer = null

// Dialogs
const copyDialog = ref(false)
const detailsDialog = ref(false)
const selectedItem = ref(null)
const copyType = ref('')
const detailsType = ref('')
const newName = ref('')

// Table Headers
const wodHeaders = [
  { title: 'Actions', value: 'actions', sortable: false, width: '100px' },
  { title: 'Name', value: 'name', sortable: false },
  { title: 'Creator', value: 'creator', sortable: false },
  { title: 'Score Type', value: 'score_type', sortable: false },
  { title: 'Created', value: 'created_at', sortable: false }
]

const movementHeaders = [
  { title: 'Actions', value: 'actions', sortable: false, width: '100px' },
  { title: 'Name', value: 'name', sortable: false },
  { title: 'Creator', value: 'creator', sortable: false },
  { title: 'Type', value: 'type', sortable: false },
  { title: 'Created', value: 'created_at', sortable: false }
]

const workoutHeaders = [
  { title: 'Actions', value: 'actions', sortable: false, width: '100px' },
  { title: 'Name', value: 'name', sortable: false },
  { title: 'Creator', value: 'creator', sortable: false },
  { title: 'Created', value: 'created_at', sortable: false }
]

// Build query params
function buildWodParams() {
  const params = new URLSearchParams()
  params.append('limit', limit)
  params.append('offset', wodOffset.value)
  if (wodSearch.value) params.append('search', wodSearch.value)
  if (wodScoreTypeFilter.value) params.append('score_type', wodScoreTypeFilter.value)
  if (wodCreatorFilter.value) params.append('creator', wodCreatorFilter.value)
  return params.toString()
}

function buildMovementParams() {
  const params = new URLSearchParams()
  params.append('limit', limit)
  params.append('offset', movementOffset.value)
  if (movementSearch.value) params.append('search', movementSearch.value)
  if (movementTypeFilter.value) params.append('type', movementTypeFilter.value)
  if (movementCreatorFilter.value) params.append('creator', movementCreatorFilter.value)
  return params.toString()
}

function buildWorkoutParams() {
  const params = new URLSearchParams()
  params.append('limit', limit)
  params.append('offset', workoutOffset.value)
  if (workoutSearch.value) params.append('search', workoutSearch.value)
  if (workoutCreatorFilter.value) params.append('creator', workoutCreatorFilter.value)
  return params.toString()
}

// Load data
async function loadWODs() {
  loading.value = true
  try {
    const response = await axios.get(`/api/admin/user-created/wods?${buildWodParams()}`)
    wods.value = response.data.data || []
    wodCount.value = response.data.count || 0
  } catch (e) {
    error.value = e.response?.data?.error || 'Failed to load WODs'
  } finally {
    loading.value = false
  }
}

async function loadMovements() {
  loading.value = true
  try {
    const response = await axios.get(`/api/admin/user-created/movements?${buildMovementParams()}`)
    movements.value = response.data.data || []
    movementCount.value = response.data.count || 0
  } catch (e) {
    error.value = e.response?.data?.error || 'Failed to load movements'
  } finally {
    loading.value = false
  }
}

async function loadWorkouts() {
  loading.value = true
  try {
    const response = await axios.get(`/api/admin/user-created/workouts?${buildWorkoutParams()}`)
    workouts.value = response.data.data || []
    workoutCount.value = response.data.count || 0
  } catch (e) {
    error.value = e.response?.data?.error || 'Failed to load workouts'
  } finally {
    loading.value = false
  }
}

async function loadAll() {
  await Promise.all([loadWODs(), loadMovements(), loadWorkouts()])
}

// Debounce functions
function debounceLoadWODs() {
  clearTimeout(wodDebounceTimer)
  wodDebounceTimer = setTimeout(() => {
    wodOffset.value = 0
    loadWODs()
  }, 300)
}

function debounceLoadMovements() {
  clearTimeout(movementDebounceTimer)
  movementDebounceTimer = setTimeout(() => {
    movementOffset.value = 0
    loadMovements()
  }, 300)
}

function debounceLoadWorkouts() {
  clearTimeout(workoutDebounceTimer)
  workoutDebounceTimer = setTimeout(() => {
    workoutOffset.value = 0
    loadWorkouts()
  }, 300)
}

// Reset and load (for dropdown filters)
function resetAndLoadWODs() {
  wodOffset.value = 0
  loadWODs()
}

function resetAndLoadMovements() {
  movementOffset.value = 0
  loadMovements()
}

function resetAndLoadWorkouts() {
  workoutOffset.value = 0
  loadWorkouts()
}

// Pagination
function previousWodPage() {
  wodOffset.value = Math.max(0, wodOffset.value - limit)
  loadWODs()
}

function nextWodPage() {
  wodOffset.value += limit
  loadWODs()
}

function previousMovementPage() {
  movementOffset.value = Math.max(0, movementOffset.value - limit)
  loadMovements()
}

function nextMovementPage() {
  movementOffset.value += limit
  loadMovements()
}

function previousWorkoutPage() {
  workoutOffset.value = Math.max(0, workoutOffset.value - limit)
  loadWorkouts()
}

function nextWorkoutPage() {
  workoutOffset.value += limit
  loadWorkouts()
}

// Copy to Standard
function openCopyDialog(type, item) {
  copyType.value = type
  selectedItem.value = item
  newName.value = item.name
  copyDialog.value = true
}

function openCopyDialogFromDetails() {
  detailsDialog.value = false
  copyType.value = detailsType.value
  newName.value = selectedItem.value.name
  copyDialog.value = true
}

async function confirmCopy() {
  actionLoading.value = true
  try {
    const endpoint = `/api/admin/user-created/${copyType.value}s/${selectedItem.value.id}/copy-to-standard`
    await axios.post(endpoint, { new_name: newName.value || selectedItem.value.name })
    successMessage.value = `Successfully converted "${selectedItem.value.name}" to standard ${copyType.value}s`
    copyDialog.value = false
    // Reload the appropriate list
    if (copyType.value === 'wod') await loadWODs()
    else if (copyType.value === 'movement') await loadMovements()
    else await loadWorkouts()
  } catch (e) {
    error.value = e.response?.data?.error || `Failed to convert ${copyType.value} to standard`
  } finally {
    actionLoading.value = false
  }
}

// View Details
function viewDetails(type, item) {
  detailsType.value = type
  selectedItem.value = item
  detailsDialog.value = true
}

// Formatters
function formatDate(dateStr) {
  if (!dateStr) return 'N/A'
  return new Date(dateStr).toLocaleDateString()
}

function formatDateTime(dateStr) {
  if (!dateStr) return 'N/A'
  return new Date(dateStr).toLocaleString()
}

function getScoreTypeColor(scoreType) {
  switch (scoreType) {
    case 'Time (HH:MM:SS)': return 'blue'
    case 'Rounds+Reps': return 'green'
    case 'Max Weight': return 'orange'
    default: return 'grey'
  }
}

function getMovementTypeColor(type) {
  switch (type) {
    case 'weightlifting': return 'orange'
    case 'bodyweight': return 'green'
    case 'cardio': return 'blue'
    case 'gymnastics': return 'purple'
    default: return 'grey'
  }
}

onMounted(() => {
  loadAll()
})
</script>
