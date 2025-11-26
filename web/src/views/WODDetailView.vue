<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">

    <v-container class="pa-2" style=" margin-bottom: 70px">
      <!-- Back Button -->
      <v-btn
        variant="text"
        color="#00bcd4"
        class="mb-2"
        @click="router.back()"
      >
        <v-icon start>mdi-arrow-left</v-icon>
        Back
      </v-btn>

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
                <v-chip v-if="!wod.is_standard" size="small" color="teal">
                  Custom
                </v-chip>
              </div>
            </div>
            <!-- Edit Button (visible to admins or WOD owner) -->
            <v-btn
              v-if="canEdit"
              icon
              variant="text"
              color="#00bcd4"
              @click="editWOD"
            >
              <v-icon>mdi-pencil</v-icon>
              <v-tooltip activator="parent" location="bottom">Edit WOD</v-tooltip>
            </v-btn>
          </div>
        </v-card>

        <!-- Quick Log Button -->
        <v-btn
          block
          size="large"
          color="teal"
          rounded="lg"
          elevation="2"
          class="mb-3 font-weight-bold"
          style="text-transform: none"
          @click="openQuickLog"
        >
          <v-icon start>mdi-lightning-bolt</v-icon>
          Quick Log {{ wod.name }}
        </v-btn>

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

          <!-- Description with markdown -->
          <div v-if="wod.description" class="mb-3">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Description</p>
            <div class="text-body-2" style="color: #1a1a1a">
              <MarkdownRenderer :content="wod.description" />
            </div>
          </div>

          <!-- Notes with markdown -->
          <div v-if="wod.notes" class="mb-3">
            <p class="text-caption font-weight-bold mb-1" style="color: #666">Notes</p>
            <div class="text-body-2" style="color: #1a1a1a">
              <MarkdownRenderer :content="wod.notes" />
            </div>
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

    <!-- Quick Log Dialog -->
    <v-dialog v-model="quickLogDialog" max-width="500px">
      <v-card>
        <v-card-title class="text-h6 font-weight-bold" style="background: #00bcd4; color: white">
          <v-icon color="white" class="mr-2">mdi-lightning-bolt</v-icon>
          Quick Log {{ wod?.name }}
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
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/axios'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// Computed: can current user edit this WOD?
const canEdit = computed(() => {
  if (!wod.value || !authStore.user) return false
  // Admin can edit any WOD (including standard)
  if (authStore.user.role === 'admin') return true
  // Non-standard WOD owned by current user
  if (!wod.value.is_standard && wod.value.created_by === authStore.user.id) return true
  return false
})

const wod = ref(null)
const loading = ref(false)
const error = ref('')

// Quick Log state
const quickLogDialog = ref(false)
const quickLogSubmitting = ref(false)
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
function openQuickLog() {
  if (!wod.value) return

  // Reset and pre-populate quick log data
  quickLogData.value = {
    name: formatQuickLogName(getTodayDate()),
    date: getTodayDate(),
    wod: {
      scoreType: wod.value.score_type,
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
      wod_id: wod.value.id,
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
  fetchWOD()
})
</script>
