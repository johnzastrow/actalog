<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" density="compact" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-btn icon="mdi-arrow-left" color="white" size="small" @click="handleBack" />
      <v-toolbar-title class="text-white font-weight-bold">
        {{ selectionMode ? 'Select Movement' : 'Movement Library' }}
      </v-toolbar-title>
      <v-spacer />
      <v-btn
        v-if="!selectionMode"
        icon="mdi-plus"
        color="white"
        @click="createNewMovement"
      />
    </v-app-bar>

    <v-container class="pa-2" style="margin-top: 36px; margin-bottom: 70px">
      <!-- Search and Filters Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
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
          class="pa-3 mb-2"
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
                  color="#ffc107"
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
            <v-icon v-if="selectionMode" color="#00bcd4">mdi-chevron-right</v-icon>
            <v-btn
              v-else-if="!movement.is_standard"
              icon="mdi-pencil"
              size="small"
              variant="text"
              color="#00bcd4"
              @click.stop="editMovement(movement.id)"
            />
          </div>
        </v-card>
      </div>
    </v-container>

    <!-- FAB for Create (when not in selection mode) -->
    <v-btn
      v-if="!selectionMode && !loading"
      icon="mdi-plus"
      size="x-large"
      color="#ffc107"
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
      style="position: fixed; bottom: 0; width: 100%; z-index: 5"
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
  </v-container>
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

// Initialize
onMounted(() => {
  fetchMovements()
})
</script>
