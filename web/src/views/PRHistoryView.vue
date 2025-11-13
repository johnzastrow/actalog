<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-toolbar-title class="text-white font-weight-bold">Personal Records</v-toolbar-title>
      <v-spacer />
      <v-btn icon="mdi-refresh" color="white" variant="text" @click="fetchPRs" :loading="loading" />
    </v-app-bar>

    <v-container class="pa-3" style="margin-top: 56px; margin-bottom: 70px">
      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable @click:close="error = null" class="mb-4">
        {{ error }}
      </v-alert>

      <!-- Loading State -->
      <div v-if="loading && personalRecords.length === 0" class="text-center py-8">
        <v-progress-circular indeterminate color="#00bcd4" size="64" />
        <p class="mt-4 text-medium-emphasis">Loading personal records...</p>
      </div>

      <!-- Empty State -->
      <v-card
        v-else-if="!loading && personalRecords.length === 0"
        elevation="0"
        rounded="lg"
        class="pa-8 text-center"
        style="background: white"
      >
        <v-icon size="64" color="#ffc107">mdi-trophy-outline</v-icon>
        <p class="text-h6 mt-4" style="color: #2c3e50">No Personal Records Yet</p>
        <p class="text-body-2" style="color: #666">
          Start logging workouts to track your personal records
        </p>
        <v-btn
          color="#00bcd4"
          class="mt-4"
          prepend-icon="mdi-plus"
          to="/workouts/log"
          rounded="lg"
          style="text-transform: none; font-weight: 600"
        >
          Log a Workout
        </v-btn>
      </v-card>

      <!-- Personal Records List -->
      <div v-else>
        <!-- Recent PRs Section -->
        <v-card elevation="0" rounded="lg" class="mb-3 pa-3" style="background: white">
          <div class="d-flex align-center mb-2">
            <v-icon color="#ffc107" class="mr-2">mdi-trophy</v-icon>
            <h2 class="text-body-1 font-weight-bold" style="color: #1a1a1a">Recent PRs</h2>
          </div>

          <div v-if="recentPRs.length === 0" class="text-caption text-center pa-4" style="color: #666">
            No recent PRs. Keep pushing!
          </div>

          <div v-else>
            <div
              v-for="pr in recentPRs"
              :key="pr.id"
              class="pa-2 mb-2"
              style="border-left: 3px solid #ffc107; background: #fff9e6; border-radius: 4px"
            >
              <div class="d-flex align-center">
                <v-icon color="#ffc107" size="small" class="mr-2">mdi-trophy</v-icon>
                <div class="flex-grow-1">
                  <div class="font-weight-bold text-body-2" style="color: #1a1a1a">
                    {{ pr.movement?.name || 'Movement' }}
                  </div>
                  <div class="text-caption" style="color: #666">
                    <span v-if="pr.weight">{{ pr.weight }} lbs</span>
                    <span v-if="pr.reps"> × {{ pr.reps }} reps</span>
                    <span v-if="pr.sets"> ({{ pr.sets }} sets)</span>
                  </div>
                </div>
                <div class="text-caption" style="color: #999">
                  {{ formatRelativeDate(pr.created_at) }}
                </div>
              </div>
            </div>
          </div>
        </v-card>

        <!-- All Personal Records -->
        <v-card elevation="0" rounded="lg" class="pa-3" style="background: white">
          <div class="d-flex align-center mb-3">
            <v-icon color="#00bcd4" class="mr-2">mdi-chart-line</v-icon>
            <h2 class="text-body-1 font-weight-bold" style="color: #1a1a1a">All Records</h2>
          </div>

          <div
            v-for="record in personalRecords"
            :key="record.movement_id"
            class="mb-3 pa-3"
            style="border: 1px solid #e0e0e0; border-radius: 8px"
          >
            <div class="font-weight-bold text-body-2 mb-2" style="color: #1a1a1a">
              {{ record.movement_name }}
            </div>

            <v-row dense>
              <v-col v-if="record.max_weight" cols="4">
                <div class="text-caption" style="color: #666">Max Weight</div>
                <div class="font-weight-bold" style="color: #00bcd4">
                  {{ record.max_weight }} lbs
                </div>
              </v-col>

              <v-col v-if="record.max_reps" cols="4">
                <div class="text-caption" style="color: #666">Max Reps</div>
                <div class="font-weight-bold" style="color: #00bcd4">
                  {{ record.max_reps }}
                </div>
              </v-col>

              <v-col v-if="record.best_time" cols="4">
                <div class="text-caption" style="color: #666">Best Time</div>
                <div class="font-weight-bold" style="color: #00bcd4">
                  {{ formatTime(record.best_time) }}
                </div>
              </v-col>
            </v-row>

            <div class="text-caption mt-2" style="color: #999">
              Achieved on {{ formatDate(record.workout_date) }}
            </div>
          </div>
        </v-card>
      </div>
    </v-container>

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
      <v-btn value="log" to="/workouts/log" style="position: relative; bottom: 20px">
        <v-avatar color="#ffc107" size="56" style="box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2)">
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
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'

const activeTab = ref('performance')

// State
const personalRecords = ref([])
const recentPRs = ref([])
const loading = ref(false)
const error = ref(null)

// Fetch PRs from API
async function fetchPRs() {
  loading.value = true
  error.value = null

  try {
    // Fetch all personal records
    const recordsResponse = await axios.get('/api/prs')
    personalRecords.value = recordsResponse.data.prs || []

    // Fetch recent PR movements
    const recentResponse = await axios.get('/api/pr-movements?limit=5')
    recentPRs.value = recentResponse.data.movements || []

    console.log('Personal records:', personalRecords.value)
    console.log('Recent PRs:', recentPRs.value)
  } catch (err) {
    console.error('Failed to fetch PRs:', err)
    if (err.response) {
      error.value = err.response.data?.message || `Error ${err.response.status}`
    } else if (err.request) {
      error.value = 'No response from server. Is the backend running?'
    } else {
      error.value = 'Failed to fetch personal records'
    }
  } finally {
    loading.value = false
  }
}

// Format date for display
function formatDate(dateString) {
  // Parse as local date to avoid timezone conversion issues
  // Extract YYYY-MM-DD from the date string
  const datePart = dateString.split('T')[0]
  const [year, month, day] = datePart.split('-').map(Number)
  const date = new Date(year, month - 1, day) // Month is 0-indexed
  const options = {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }
  return date.toLocaleDateString('en-US', options)
}

function formatRelativeDate(dateString) {
  // Parse as local date to avoid timezone conversion issues
  // Extract YYYY-MM-DD from the date string
  const datePart = dateString.split('T')[0]
  const [year, month, day] = datePart.split('-').map(Number)
  const date = new Date(year, month - 1, day) // Month is 0-indexed

  const today = new Date()
  const dateOnly = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  const todayOnly = new Date(today.getFullYear(), today.getMonth(), today.getDate())
  const diffTime = todayOnly - dateOnly
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return 'Today'
  if (diffDays === 1) return 'Yesterday'
  if (diffDays <= 7) return `${diffDays} days ago`
  return formatDate(dateString)
}

function formatTime(seconds) {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

// Load PRs on mount
onMounted(() => {
  fetchPRs()
})
</script>
