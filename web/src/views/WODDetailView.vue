<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" density="compact" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-btn icon="mdi-arrow-left" color="white" size="small" @click="router.back()" />
      <v-toolbar-title class="text-white font-weight-bold">
        WOD Details
      </v-toolbar-title>
      <v-spacer />
      <v-btn
        v-if="wod && !wod.is_standard"
        icon="mdi-pencil"
        color="white"
        @click="editWOD"
      />
    </v-app-bar>

    <v-container class="pa-2" style="margin-top: 36px; margin-bottom: 70px">
      <!-- Loading State -->
      <div v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="#00bcd4" size="48" />
        <p class="text-body-2 mt-3" style="color: #666">Loading WOD...</p>
      </div>

      <!-- Error State -->
      <v-alert v-else-if="error" type="error" class="mb-3">
        {{ error }}
      </v-alert>

      <!-- WOD Details -->
      <div v-else-if="wod">
        <!-- Header Card -->
        <v-card elevation="0" rounded="lg" class="pa-3 mb-2" style="background: white">
          <div class="d-flex align-center mb-3">
            <v-icon color="#ff5722" size="48" class="mr-3">mdi-fire</v-icon>
            <div style="flex: 1">
              <h1 class="text-h5 font-weight-bold" style="color: #1a1a1a">
                {{ wod.name }}
              </h1>
              <div class="d-flex align-center flex-wrap gap-2 mt-2">
                <v-chip size="small" color="#9c27b0" variant="flat">
                  {{ wod.type }}
                </v-chip>
                <v-chip size="small" color="#00bcd4" variant="flat">
                  {{ wod.regime }}
                </v-chip>
                <v-chip size="small" color="#4caf50" variant="flat">
                  {{ wod.score_type }}
                </v-chip>
                <v-chip v-if="!wod.is_standard" size="small" color="#ffc107">
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
            Workout Details
          </h2>

          <!-- Source -->
          <div class="mb-3">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Source</p>
            <v-chip size="small" color="#2196f3" variant="outlined">
              {{ wod.source }}
            </v-chip>
          </div>

          <!-- Description -->
          <div v-if="wod.description" class="mb-3">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Description</p>
            <p class="text-body-2" style="color: #1a1a1a; white-space: pre-wrap">
              {{ wod.description }}
            </p>
          </div>

          <!-- Notes -->
          <div v-if="wod.notes" class="mb-3">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Notes</p>
            <p class="text-body-2" style="color: #1a1a1a; white-space: pre-wrap">
              {{ wod.notes }}
            </p>
          </div>

          <!-- Video URL -->
          <div v-if="wod.url">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Video</p>
            <v-btn
              :href="wod.url"
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

        <!-- Workout Type Info Card -->
        <v-card elevation="0" rounded="lg" class="pa-3 mb-2" style="background: white">
          <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">
            <v-icon color="#00bcd4" size="small" class="mr-1">mdi-tag</v-icon>
            Workout Classification
          </h2>

          <div class="mb-2">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Type</p>
            <p class="text-body-2" style="color: #1a1a1a">
              {{ wod.type }}
            </p>
          </div>

          <div class="mb-2">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Regime</p>
            <p class="text-body-2" style="color: #1a1a1a">
              {{ wod.regime }}
            </p>
          </div>

          <div>
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Score Type</p>
            <p class="text-body-2" style="color: #1a1a1a">
              {{ wod.score_type }}
            </p>
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
              {{ formatDate(wod.created_at) }}
            </p>
          </div>

          <div>
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Last Updated</p>
            <p class="text-body-2" style="color: #1a1a1a">
              {{ formatDate(wod.updated_at) }}
            </p>
          </div>
        </v-card>
      </div>
    </v-container>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/axios'

const router = useRouter()
const route = useRoute()

const wod = ref(null)
const loading = ref(false)
const error = ref('')

// Load WOD details
async function fetchWOD() {
  loading.value = true
  error.value = ''

  try {
    const response = await axios.get(`/api/wods/${route.params.id}`)
    wod.value = response.data.wod || response.data
  } catch (err) {
    console.error('Failed to fetch WOD:', err)
    error.value = 'Failed to load WOD details. Please try again.'
  } finally {
    loading.value = false
  }
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

// Edit WOD
function editWOD() {
  router.push(`/wods/${route.params.id}/edit`)
}

// Initialize
onMounted(() => {
  fetchWOD()
})
</script>
