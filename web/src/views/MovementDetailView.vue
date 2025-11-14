<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" density="compact" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-btn icon="mdi-arrow-left" color="white" size="small" @click="router.back()" />
      <v-toolbar-title class="text-white font-weight-bold">
        Movement Details
      </v-toolbar-title>
      <v-spacer />
      <v-btn
        v-if="movement && !movement.is_standard"
        icon="mdi-pencil"
        color="white"
        @click="editMovement"
      />
    </v-app-bar>

    <v-container class="pa-2" style="margin-top: 36px; margin-bottom: 70px">
      <!-- Loading State -->
      <div v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="#00bcd4" size="48" />
        <p class="text-body-2 mt-3" style="color: #666">Loading movement...</p>
      </div>

      <!-- Error State -->
      <v-alert v-else-if="error" type="error" class="mb-3">
        {{ error }}
      </v-alert>

      <!-- Movement Details -->
      <div v-else-if="movement">
        <!-- Header Card -->
        <v-card elevation="0" rounded="lg" class="pa-3 mb-2" style="background: white">
          <div class="d-flex align-center mb-3">
            <v-icon :color="getMovementTypeColor(movement.type)" size="48" class="mr-3">
              {{ getMovementTypeIcon(movement.type) }}
            </v-icon>
            <div style="flex: 1">
              <h1 class="text-h5 font-weight-bold" style="color: #1a1a1a">
                {{ movement.name }}
              </h1>
              <div class="d-flex align-center gap-2 mt-1">
                <v-chip size="small" :color="getMovementTypeColor(movement.type)" variant="flat">
                  {{ capitalizeFirst(movement.type) }}
                </v-chip>
                <v-chip v-if="!movement.is_standard" size="small" color="#ffc107">
                  Custom
                </v-chip>
              </div>
            </div>
          </div>
        </v-card>

        <!-- Details Card -->
        <v-card elevation="0" rounded="lg" class="pa-3 mb-2" style="background: white">
          <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">
            <v-icon color="#00bcd4" size="small" class="mr-1">mdi-information</v-icon>
            Information
          </h2>

          <!-- Description -->
          <div v-if="parsedData.description" class="mb-3">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Description</p>
            <p class="text-body-2" style="color: #1a1a1a; white-space: pre-wrap">
              {{ parsedData.description }}
            </p>
          </div>

          <!-- Difficulty -->
          <div v-if="parsedData.difficulty" class="mb-3">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Difficulty Level</p>
            <v-chip size="small" :color="getDifficultyColor(parsedData.difficulty)">
              {{ parsedData.difficulty }}
            </v-chip>
          </div>

          <!-- Equipment -->
          <div v-if="parsedData.equipment" class="mb-3">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Equipment Required</p>
            <p class="text-body-2" style="color: #1a1a1a">
              {{ parsedData.equipment }}
            </p>
          </div>

          <!-- Primary Muscles -->
          <div v-if="parsedData.primaryMuscles" class="mb-3">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Primary Muscle Groups</p>
            <p class="text-body-2" style="color: #1a1a1a">
              {{ parsedData.primaryMuscles }}
            </p>
          </div>

          <!-- Coaching Cues -->
          <div v-if="parsedData.coachingCues" class="mb-3">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Coaching Cues</p>
            <p class="text-body-2" style="color: #1a1a1a; white-space: pre-wrap">
              {{ parsedData.coachingCues }}
            </p>
          </div>

          <!-- Scaling Options -->
          <div v-if="parsedData.scalingOptions" class="mb-3">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Scaling/Modifications</p>
            <p class="text-body-2" style="color: #1a1a1a; white-space: pre-wrap">
              {{ parsedData.scalingOptions }}
            </p>
          </div>

          <!-- Video URL -->
          <div v-if="parsedData.videoUrl">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Video Tutorial</p>
            <v-btn
              :href="parsedData.videoUrl"
              target="_blank"
              color="#00bcd4"
              variant="outlined"
              prepend-icon="mdi-play-circle"
              size="small"
              rounded="lg"
              style="text-transform: none"
            >
              Watch Video
            </v-btn>
          </div>
        </v-card>

        <!-- Metadata Card -->
        <v-card elevation="0" rounded="lg" class="pa-3" style="background: white">
          <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">
            <v-icon color="#00bcd4" size="small" class="mr-1">mdi-clock</v-icon>
            Metadata
          </h2>

          <div class="mb-2">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Created</p>
            <p class="text-body-2" style="color: #1a1a1a">
              {{ formatDate(movement.created_at) }}
            </p>
          </div>

          <div>
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Last Updated</p>
            <p class="text-body-2" style="color: #1a1a1a">
              {{ formatDate(movement.updated_at) }}
            </p>
          </div>
        </v-card>
      </div>
    </v-container>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/axios'

const router = useRouter()
const route = useRoute()

const movement = ref(null)
const loading = ref(false)
const error = ref('')

// Parse structured data from description field
const parsedData = computed(() => {
  if (!movement.value) return {}

  const desc = movement.value.description || ''

  // Check if description contains structured data
  if (desc.startsWith('__STRUCTURED__')) {
    try {
      const jsonStr = desc.substring('__STRUCTURED__'.length)
      return JSON.parse(jsonStr)
    } catch (e) {
      console.error('Failed to parse structured data:', e)
      return { description: desc }
    }
  }

  // Legacy plain text description
  return { description: desc }
})

// Load movement details
async function fetchMovement() {
  loading.value = true
  error.value = ''

  try {
    const response = await axios.get(`/api/movements/${route.params.id}`)
    movement.value = response.data
  } catch (err) {
    console.error('Failed to fetch movement:', err)
    error.value = 'Failed to load movement details. Please try again.'
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

// Get difficulty color
function getDifficultyColor(difficulty) {
  const colors = {
    Beginner: '#4caf50',
    Intermediate: '#ffc107',
    Advanced: '#ff5722'
  }
  return colors[difficulty] || '#666'
}

// Capitalize first letter
function capitalizeFirst(str) {
  if (!str) return ''
  return str.charAt(0).toUpperCase() + str.slice(1)
}

// Format date
function formatDate(dateString) {
  if (!dateString) return 'N/A'
  // Parse as local date to avoid timezone conversion issues
  // Extract YYYY-MM-DD from the date string
  const datePart = dateString.split('T')[0]
  const [year, month, day] = datePart.split('-').map(Number)
  const date = new Date(year, month - 1, day) // Month is 0-indexed
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// Edit movement
function editMovement() {
  router.push(`/movements/${route.params.id}/edit`)
}

// Initialize
onMounted(() => {
  fetchMovement()
})
</script>
