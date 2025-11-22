<template>
  <div style="background-color: #f5f7fa; min-height: 100vh">
    <v-container class="pa-2" style="margin-bottom: 70px; overflow-y: auto; max-height: calc(100vh - 70px)">
      <!-- Search and Filters Card -->
      <v-card elevation="0" rounded="lg" class="pa-2 mb-2" style="background: white">
        <v-text-field
          v-model="searchQuery"
          label="Search movements"
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
          <v-chip value="weightlifting" size="small">Weightlifting</v-chip>
          <v-chip value="gymnastics" size="small">Gymnastics</v-chip>
          <v-chip value="cardio" size="small">Cardio</v-chip>
          <v-chip value="bodyweight" size="small">Bodyweight</v-chip>
        </v-chip-group>
      </v-card>

      <!-- Error Alert -->
      <v-alert
        v-if="error"
        type="error"
        closable
        @click:close="error = ''"
        class="mb-3"
      >
        {{ error }}
      </v-alert>

      <!-- Loading State -->
      <div v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="#00bcd4" size="48" />
        <p class="text-body-2 mt-3" style="color: #666">Loading movements...</p>
      </div>

      <!-- Empty State -->
      <v-card
        v-else-if="filteredMovements.length === 0"
        elevation="0"
        rounded="lg"
        class="pa-6 text-center"
        style="background: white"
      >
        <v-icon size="64" color="#ccc">mdi-dumbbell</v-icon>
        <p class="text-h6 mt-3" style="color: #666">No movements found</p>
        <p class="text-body-2" style="color: #999">
          {{ searchQuery ? 'Try adjusting your search' : 'Create your first movement to get started' }}
        </p>
        <v-btn
          v-if="!searchQuery"
          color="#00bcd4"
          size="large"
          rounded="lg"
          class="mt-4"
          prepend-icon="mdi-plus"
          @click="createNewMovement"
          style="text-transform: none"
        >
          Create Movement
        </v-btn>
      </v-card>

      <!-- Movements List -->
      <div v-else>
        <v-card
          v-for="movement in filteredMovements"
          :key="movement.id"
          elevation="0"
          rounded="lg"
          class="pa-2 mb-2"
          style="background: white; border: 1px solid #e0e0e0"
          :ripple="true"
          @click="handleMovementClick(movement)"
        >
          <div class="d-flex align-center">
            <v-icon :color="getMovementTypeColor(movement.type)" class="mr-3">
              {{ getMovementTypeIcon(movement.type) }}
            </v-icon>
            <div style="flex: 1">
              <div class="d-flex align-center">
                <span class="text-body-1 font-weight-bold" style="color: #1a1a1a">
                  {{ movement.name }}
                </span>
                <v-chip
                  v-if="!movement.is_standard"
                  size="x-small"
                  color="teal"
                  class="ml-2"
                >
                  Custom
                </v-chip>
              </div>
              <p class="text-caption mb-0" style="color: #666">
                {{ movement.description }}
              </p>
              <v-chip size="x-small" :color="getMovementTypeColor(movement.type)" class="mt-1" variant="outlined">
                {{ capitalizeFirst(movement.type) }}
              </v-chip>
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
                @click.stop="openQuickLog(movement)"
              >
                <v-icon>mdi-lightning-bolt</v-icon>
                <v-tooltip activator="parent" location="top">Quick Log</v-tooltip>
              </v-btn>
              <v-btn
                v-if="!movement.is_standard"
                icon="mdi-pencil"
                size="small"
                variant="text"
                color="#00bcd4"
                @click.stop="editMovement(movement.id)"
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
      @click="createNewMovement"
    />

    <!-- Bottom Navigation (only when not in selection mode) -->
    <v-bottom-navigation
      v-if="!selectionMode"
      v-model="activeNav"
      name="bottom-navigation"
      color="#00bcd4"
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
          Quick Log {{ selectedMovement?.name }}
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

            <!-- Movement Performance Form -->
            <div class="mt-3 pa-3" style="background: #f5f5f5; border-radius: 8px">
              <div class="mb-2">
                <label class="text-caption">Sets</label>
                <v-text-field
                  v-model.number="quickLogData.movement.sets"
                  type="number"
                  variant="outlined"
                  density="compact"
                  hide-details
                  min="0"
                />
              </div>
              <div class="mb-2">
                <label class="text-caption">Reps</label>
                <v-text-field
                  v-model.number="quickLogData.movement.reps"
                  type="number"
                  variant="outlined"
                  density="compact"
                  hide-details
                  min="0"
                />
              </div>
              <div class="mb-2">
                <label class="text-caption">Weight (lbs)</label>
                <v-text-field
                  v-model.number="quickLogData.movement.weight"
                  type="number"
                  variant="outlined"
                  density="compact"
                  hide-details
                  min="0"
                  step="0.1"
                />
              </div>
              <div class="mb-2">
                <label class="text-caption">Time (seconds)</label>
                <v-text-field
                  v-model.number="quickLogData.movement.time"
                  type="number"
                  variant="outlined"
                  density="compact"
                  hide-details
                  min="0"
                />
              </div>
              <div class="mb-2">
                <label class="text-caption">Distance (meters)</label>
                <v-text-field
                  v-model.number="quickLogData.movement.distance"
                  type="number"
                  variant="outlined"
                  density="compact"
                  hide-details
                  min="0"
                  step="0.1"
                />
              </div>
              <div>
                <label class="text-caption">Notes</label>
                <v-textarea
                  v-model="quickLogData.movement.notes"
                  variant="outlined"
                  density="compact"
                  rows="2"
                  hide-details
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
import axios from '@/utils/axios'

const router = useRouter()
const route = useRoute()

// State
const movements = ref([])
const loading = ref(false)
const error = ref('')
const searchQuery = ref('')
const selectedType = ref('all')
const activeNav = ref('movements')

// Quick Log state
const quickLogDialog = ref(false)
const quickLogSubmitting = ref(false)
const selectedMovement = ref(null)
const quickLogData = ref({
  name: '',
  date: '',
  movement: {
    sets: null,
    reps: null,
    weight: null,
    time: null,
    distance: null,
    notes: ''
  }
})

// Check if in selection mode (opened from another screen)
const selectionMode = computed(() => route.query.select === 'true')

// Filtered movements
const filteredMovements = computed(() => {
  let filtered = movements.value

  // Filter by type
  if (selectedType.value !== 'all') {
    filtered = filtered.filter(m => m.type === selectedType.value)
  }

  // Filter by search query
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(m =>
      m.name.toLowerCase().includes(query) ||
      (m.description && m.description.toLowerCase().includes(query))
    )
  }

  return filtered
})

// Load movements
async function fetchMovements() {
  loading.value = true
  error.value = ''

  try {
    const response = await axios.get('/api/movements')
    movements.value = response.data.movements || []
  } catch (err) {
    console.error('Failed to fetch movements:', err)
    error.value = 'Failed to load movements. Please try again.'
  } finally {
    loading.value = false
  }
}

// Get movement type icon
function getMovementTypeIcon(type) {
  const icons = {
    weightlifting: 'mdi-weight-lifter',
    gymnastics: 'mdi-gymnastics',
    cardio: 'mdi-run',
    bodyweight: 'mdi-human'
  }
  return icons[type] || 'mdi-dumbbell'
}

// Get movement type color
function getMovementTypeColor(type) {
  const colors = {
    weightlifting: '#00bcd4',
    gymnastics: '#9c27b0',
    cardio: '#ff5722',
    bodyweight: '#4caf50'
  }
  return colors[type] || '#666'
}

// Capitalize first letter
function capitalizeFirst(str) {
  if (!str) return ''
  return str.charAt(0).toUpperCase() + str.slice(1)
}

// Handle movement click
function handleMovementClick(movement) {
  if (selectionMode.value) {
    // In selection mode, emit selection and go back
    // The parent component should handle this via route params or state
    router.push({
      path: route.query.returnPath || '/workouts/templates/create',
      query: { selectedMovement: movement.id }
    })
  } else {
    // In browse mode, navigate to movement detail
    router.push(`/movements/${movement.id}`)
  }
}

// Create new movement
function createNewMovement() {
  if (selectionMode.value) {
    // Pass return path so we can come back after creation
    router.push({
      path: '/movements/create',
      query: { returnPath: route.query.returnPath, select: 'true' }
    })
  } else {
    router.push('/movements/create')
  }
}

// Edit movement
function editMovement(id) {
  router.push(`/movements/${id}/edit`)
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
function openQuickLog(movement) {
  if (!movement) return

  selectedMovement.value = movement

  // Reset and pre-populate quick log data
  quickLogData.value = {
    name: formatQuickLogName(getTodayDate()),
    date: getTodayDate(),
    movement: {
      sets: null,
      reps: null,
      weight: null,
      time: null,
      distance: null,
      notes: ''
    }
  }

  quickLogDialog.value = true
}

// Close Quick Log
function closeQuickLog() {
  quickLogDialog.value = false
  selectedMovement.value = null
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

    // Add movement performance
    const m = quickLogData.value.movement
    if (m.sets || m.reps || m.weight || m.time || m.distance) {
      payload.movements.push({
        movement_id: selectedMovement.value.id,
        sets: m.sets || null,
        reps: m.reps || null,
        weight: m.weight || null,
        time_seconds: m.time || null,
        distance: m.distance || null,
        notes: m.notes || null
      })
    }

    await axios.post('/api/workouts', payload)

    // Close dialog and navigate to dashboard
    quickLogDialog.value = false
    selectedMovement.value = null
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
  fetchMovements()
})
</script>
