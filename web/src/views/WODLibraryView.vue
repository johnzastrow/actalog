<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <!-- Search and Filters Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <v-text-field
          v-model="searchQuery"
          label="Search WODs"
          placeholder="Search by name..."
          variant="outlined"
          density="compact"
          rounded="lg"
          clearable
          hide-details
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-magnify</v-icon>
          </template>
        </v-text-field>

        <v-chip-group
          v-model="selectedType"
          selected-class="text-white"
          color="#00bcd4"
          class="mt-2"
          mandatory
        >
          <v-chip value="all" size="small">All</v-chip>
          <v-chip value="Girl" size="small">Girl</v-chip>
          <v-chip value="Hero" size="small">Hero</v-chip>
          <v-chip value="Benchmark" size="small">Benchmark</v-chip>
          <v-chip value="Games" size="small">Games</v-chip>
        </v-chip-group>
      </v-card>

      <!-- Error Alert -->
      <v-alert
        v-if="wodsStore.error"
        type="error"
        closable
        @click:close="wodsStore.error = ''"
        class="mb-3"
      >
        {{ wodsStore.error }}
      </v-alert>

      <!-- Loading State -->
      <div v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="#00bcd4" size="48" />
        <p class="text-body-2 mt-3" style="color: #666">Loading WODs...</p>
      </div>

      <!-- Empty State -->
      <v-card
        v-else-if="filteredWODs.length === 0"
        elevation="0"
        rounded="lg"
        class="pa-6 text-center"
        style="background: white"
      >
        <v-icon size="64" color="#ccc">mdi-fire</v-icon>
        <p class="text-h6 mt-3" style="color: #666">No WODs found</p>
        <p class="text-body-2" style="color: #999">
          {{ searchQuery ? 'Try adjusting your search' : 'Create your first WOD to get started' }}
        </p>
        <v-btn
          v-if="!searchQuery"
          color="#00bcd4"
          size="large"
          rounded="lg"
          class="mt-4"
          prepend-icon="mdi-plus"
          @click="createNewWOD"
          style="text-transform: none"
        >
          Create WOD
        </v-btn>
      </v-card>

      <!-- WODs List -->
      <div v-else>
        <v-card
          v-for="wod in filteredWODs"
          :key="wod.id"
          elevation="0"
          rounded="lg"
          class="pa-3 mb-2"
          style="background: white; border: 1px solid #e0e0e0"
          :ripple="true"
          @click="handleWODClick(wod)"
        >
          <div class="d-flex align-center">
            <v-icon color="#ff5722" class="mr-3" size="32">mdi-fire</v-icon>
            <div style="flex: 1">
              <div class="d-flex align-center mb-1">
                <span class="text-body-1 font-weight-bold" style="color: #1a1a1a">
                  {{ wod.name }}
                </span>
                <v-chip
                  v-if="!wod.is_standard"
                  size="x-small"
                  color="teal"
                  class="ml-2"
                >
                  Custom
                </v-chip>
              </div>
              <div class="d-flex flex-wrap gap-1 mb-1">
                <v-chip size="x-small" color="#9c27b0" variant="outlined">
                  {{ wod.type }}
                </v-chip>
                <v-chip size="x-small" color="#00bcd4" variant="outlined">
                  {{ wod.regime }}
                </v-chip>
                <v-chip size="x-small" color="#4caf50" variant="outlined">
                  {{ wod.score_type }}
                </v-chip>
              </div>
              <p class="text-caption mb-0" style="color: #666; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;">
                {{ wod.description }}
              </p>
            </div>
            <div v-if="selectionMode">
              <v-icon color="#00bcd4">mdi-chevron-right</v-icon>
            </div>
            <div v-else class="d-flex gap-1">
              <v-btn
                icon="mdi-lightning-bolt"
                size="small"
                variant="text"
                color="teal"
                @click.stop="openQuickLog(wod)"
              >
                <v-icon>mdi-lightning-bolt</v-icon>
                <v-tooltip activator="parent" location="top">Quick Log</v-tooltip>
              </v-btn>
              <v-btn
                v-if="!wod.is_standard"
                icon="mdi-pencil"
                size="small"
                variant="text"
                color="#00bcd4"
                @click.stop="editWOD(wod.id)"
              >
                <v-icon>mdi-pencil</v-icon>
                <v-tooltip activator="parent" location="top">Edit</v-tooltip>
              </v-btn>
            </div>
          </div>
        </v-card>
      </div>
    </v-container>

    <!-- FAB for Create (when not in selection mode) -->
    <v-btn
      v-if="!selectionMode && !loading"
      icon="mdi-plus"
      size="x-large"
      color="teal"
      elevation="8"
      style="position: fixed; bottom: 80px; right: 16px; z-index: 5"
      @click="createNewWOD"
    />

    <!-- Bottom Navigation (only when not in selection mode) -->
    <v-bottom-navigation
      v-if="!selectionMode"
      v-model="activeNav"
      grow
      style="position: fixed; bottom: 0; width: 100%; z-index: 5; background: white"
      elevation="8"
    >
      <v-btn value="dashboard" @click="$router.push('/dashboard')">
        <v-icon>mdi-view-dashboard</v-icon>
        <span>Dashboard</span>
      </v-btn>

      <v-btn value="workouts" @click="$router.push('/workouts')">
        <v-icon>mdi-clipboard-text</v-icon>
        <span>Workouts</span>
      </v-btn>

      <v-btn value="performance" @click="$router.push('/performance')">
        <v-icon>mdi-chart-line</v-icon>
        <span>Performance</span>
      </v-btn>

      <v-btn value="profile" @click="$router.push('/profile')">
        <v-icon>mdi-account</v-icon>
        <span>Profile</span>
      </v-btn>
    </v-bottom-navigation>

    <!-- Quick Log Dialog -->
    <v-dialog v-model="quickLogDialog" max-width="500px">
      <v-card>
        <v-card-title class="text-h6 font-weight-bold" style="background: #00bcd4; color: white">
          <v-icon color="white" class="mr-2">mdi-lightning-bolt</v-icon>
          Quick Log {{ selectedWOD?.name }}
        </v-card-title>

        <v-card-text class="pa-2">
          <v-form ref="quickLogForm" @submit.prevent="submitQuickLog">
            <!-- Date -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" style="color: #1a1a1a">Date *</label>
              <v-text-field
                v-model="quickLogData.date"
                type="date"
                variant="outlined"
                density="compact"
                hide-details
                required
              />
            </div>

            <!-- Workout Name -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" style="color: #1a1a1a">Workout Name *</label>
              <v-text-field
                v-model="quickLogData.name"
                variant="outlined"
                density="compact"
                placeholder="e.g., Morning Run, Upper Body, etc."
                hide-details
                required
              />
            </div>

            <!-- WOD Performance Form -->
            <div class="mt-3 pa-3" style="background: #f5f5f5; border-radius: 8px">
              <div class="mb-2">
                <label class="text-caption">Score Type (from WOD)</label>
                <v-text-field
                  v-model="quickLogData.wod.scoreType"
                  variant="outlined"
                  density="compact"
                  hide-details
                  readonly
                  bg-color="#e0e0e0"
                />
              </div>

              <!-- Time-based WOD fields -->
              <div v-if="quickLogData.wod.scoreType === 'Time (HH:MM:SS)'">
                <label class="text-caption d-block mb-1">Time (HH:MM:SS) *</label>
                <div class="d-flex gap-2 mb-2">
                  <v-text-field
                    v-model.number="quickLogData.wod.timeHours"
                    type="number"
                    variant="outlined"
                    density="compact"
                    hide-details
                    min="0"
                    max="23"
                    placeholder="HH"
                    style="flex: 1"
                  />
                  <span class="align-self-center">:</span>
                  <v-text-field
                    v-model.number="quickLogData.wod.timeMinutes"
                    type="number"
                    variant="outlined"
                    density="compact"
                    hide-details
                    min="0"
                    max="59"
                    placeholder="MM"
                    style="flex: 1"
                  />
                  <span class="align-self-center">:</span>
                  <v-text-field
                    v-model.number="quickLogData.wod.timeSecondsInput"
                    type="number"
                    variant="outlined"
                    density="compact"
                    hide-details
                    min="0"
                    max="59"
                    placeholder="SS"
                    style="flex: 1"
                  />
                </div>
              </div>

              <!-- Rounds+Reps WOD fields -->
              <template v-if="quickLogData.wod.scoreType === 'Rounds+Reps'">
                <div class="mb-2">
                  <label class="text-caption">Rounds *</label>
                  <v-text-field
                    v-model.number="quickLogData.wod.rounds"
                    type="number"
                    variant="outlined"
                    density="compact"
                    hide-details
                    min="0"
                    placeholder="e.g., 10"
                  />
                </div>
                <div class="mb-2">
                  <label class="text-caption">Reps (optional)</label>
                  <v-text-field
                    v-model.number="quickLogData.wod.reps"
                    type="number"
                    variant="outlined"
                    density="compact"
                    hide-details
                    min="0"
                    placeholder="e.g., 15"
                  />
                </div>
              </template>

              <!-- Max Weight WOD field -->
              <div v-if="quickLogData.wod.scoreType === 'Max Weight'" class="mb-2">
                <label class="text-caption">Weight (lbs) *</label>
                <v-text-field
                  v-model.number="quickLogData.wod.weight"
                  type="number"
                  variant="outlined"
                  density="compact"
                  hide-details
                  min="0"
                  step="0.5"
                  placeholder="e.g., 225"
                />
              </div>

              <!-- Notes field (always shown) -->
              <div>
                <label class="text-caption">Notes</label>
                <v-textarea
                  v-model="quickLogData.wod.notes"
                  variant="outlined"
                  density="compact"
                  rows="2"
                  hide-details
                  placeholder="How did it feel?"
                />
              </div>
            </div>
          </v-form>
        </v-card-text>

        <v-card-actions class="pa-2 pt-0">
          <v-btn variant="text" @click="closeQuickLog">Cancel</v-btn>
          <v-spacer />
          <v-btn
            color="teal"
            variant="elevated"
            :loading="quickLogSubmitting"
            :disabled="!quickLogData.name || !quickLogData.date"
            @click="submitQuickLog"
          >
            <v-icon start>mdi-check</v-icon>
            Log Workout
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useWodsStore } from '@/stores/wods'
import axios from '@/utils/axios'

const router = useRouter()
const route = useRoute()
const wodsStore = useWodsStore()

// State
const searchQuery = ref('')
const selectedType = ref('all')
const activeNav = ref('wods')

// Quick Log state
const quickLogDialog = ref(false)
const quickLogSubmitting = ref(false)
const selectedWOD = ref(null)
const quickLogData = ref({
  name: '',
  date: '',
  wod: {
    scoreType: '',
    timeHours: 0,
    timeMinutes: 0,
    timeSecondsInput: 0,
    rounds: null,
    reps: null,
    weight: null,
    notes: ''
  }
})

// Check if in selection mode
const selectionMode = computed(() => route.query.select === 'true')

// Computed properties from store
const loading = computed(() => wodsStore.loading)
const error = computed(() => wodsStore.error)

// Filtered WODs
const filteredWODs = computed(() => {
  let filtered = wodsStore.wods

  // Filter by type
  if (selectedType.value !== 'all') {
    filtered = wodsStore.filterByType(selectedType.value)
  }

  // Filter by search query
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(w =>
      w.name.toLowerCase().includes(query) ||
      (w.description && w.description.toLowerCase().includes(query))
    )
  }

  return filtered
})

// Load WODs
async function fetchWODs() {
  await wodsStore.fetchWods()
}

// Handle WOD click
function handleWODClick(wod) {
  if (selectionMode.value) {
    // In selection mode, return with selected WOD
    router.push({
      path: route.query.returnPath || '/workouts/templates/create',
      query: { selectedWOD: wod.id }
    })
  } else {
    // In browse mode, navigate to WOD detail
    router.push(`/wods/${wod.id}`)
  }
}

// Create new WOD
function createNewWOD() {
  if (selectionMode.value) {
    router.push({
      path: '/wods/create',
      query: { returnPath: route.query.returnPath, select: 'true' }
    })
  } else {
    router.push('/wods/create')
  }
}

// Edit WOD
function editWOD(id) {
  router.push(`/wods/${id}/edit`)
}

// Handle back navigation
function handleBack() {
  if (selectionMode.value && route.query.returnPath) {
    router.push(route.query.returnPath)
  } else {
    router.back()
  }
}

// Helper functions for Quick Log
function getTodayDate() {
  const today = new Date()
  return today.toISOString().split('T')[0]
}

function formatQuickLogName(date) {
  if (!date) return 'Workout'
  const d = new Date(date + 'T00:00:00')
  const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
  return `${days[d.getDay()]} Workout`
}

// Open Quick Log
function openQuickLog(wod) {
  if (!wod) return

  selectedWOD.value = wod

  // Reset and pre-populate quick log data
  quickLogData.value = {
    name: formatQuickLogName(getTodayDate()),
    date: getTodayDate(),
    wod: {
      scoreType: wod.score_type,
      timeHours: 0,
      timeMinutes: 0,
      timeSecondsInput: 0,
      rounds: null,
      reps: null,
      weight: null,
      notes: ''
    }
  }

  quickLogDialog.value = true
}

// Close Quick Log
function closeQuickLog() {
  quickLogDialog.value = false
  selectedWOD.value = null
}

// Submit Quick Log
async function submitQuickLog() {
  quickLogSubmitting.value = true

  try {
    const payload = {
      workout_date: quickLogData.value.date,
      workout_name: quickLogData.value.name,
      total_time: null,
      notes: null,
      movements: [],
      wods: []
    }

    // Add WOD performance
    const w = quickLogData.value.wod
    const wodPerformance = {
      wod_id: selectedWOD.value.id,
      notes: w.notes || null
    }

    // Handle different score types
    if (w.scoreType === 'Time (HH:MM:SS)') {
      // Convert HH:MM:SS to total seconds
      const totalSeconds = (w.timeHours || 0) * 3600 + (w.timeMinutes || 0) * 60 + (w.timeSecondsInput || 0)
      if (totalSeconds > 0) {
        wodPerformance.time_seconds = totalSeconds
      }
    } else if (w.scoreType === 'Rounds+Reps') {
      if (w.rounds !== null) {
        wodPerformance.rounds = w.rounds
      }
      if (w.reps !== null) {
        wodPerformance.reps = w.reps
      }
    } else if (w.scoreType === 'Max Weight') {
      if (w.weight !== null) {
        wodPerformance.weight = w.weight
      }
    }

    // Only add if there's actual performance data
    if (wodPerformance.time_seconds || wodPerformance.rounds !== undefined || wodPerformance.weight !== undefined) {
      payload.wods.push(wodPerformance)
    }

    await axios.post('/api/workouts', payload)

    // Close dialog and navigate to dashboard
    quickLogDialog.value = false
    selectedWOD.value = null
    router.push('/dashboard')
  } catch (err) {
    console.error('Failed to log workout:', err)
    alert(err.response?.data?.message || 'Failed to log workout')
  } finally {
    quickLogSubmitting.value = false
  }
}

// Initialize
onMounted(() => {
  fetchWODs()
})
</script>
