<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-btn icon="mdi-arrow-left" color="white" @click="$router.back()" />
      <v-toolbar-title class="text-white font-weight-bold">Log Workout</v-toolbar-title>
      <v-spacer />
    </v-app-bar>

    <v-container class="pa-3" style="margin-top: 56px; margin-bottom: 100px">
      <!-- Success Alert -->
      <v-alert v-if="success" type="success" closable @click:close="success = null" class="mb-4">
        {{ success }}
      </v-alert>

      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable @click:close="error = null" class="mb-4">
        {{ error }}
      </v-alert>

      <v-form @submit.prevent="logWorkout">
        <!-- Select Workout Template -->
        <div class="mb-3">
          <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
            Workout Template
          </label>
          <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
            <v-autocomplete
              v-model="selectedTemplateId"
              :items="workoutTemplates"
              item-title="name"
              item-value="id"
              :loading="loadingTemplates"
              placeholder="Search workout templates..."
              variant="plain"
              density="compact"
              hide-details
              clearable
              auto-select-first
              style="color: #1a1a1a; font-weight: 500"
              @update:model-value="onTemplateSelected"
            >
              <template #prepend-inner>
                <v-icon color="#00bcd4" size="small">mdi-magnify</v-icon>
              </template>
              <template #item="{ props, item }">
                <v-list-item v-bind="props" density="compact">
                  <template #prepend>
                    <v-icon
                      :color="item.raw.created_by ? '#00bcd4' : '#ffc107'"
                      size="small"
                    >
                      {{ item.raw.created_by ? 'mdi-account' : 'mdi-star' }}
                    </v-icon>
                  </template>
                  <v-list-item-title class="text-body-2">
                    {{ item.raw.name }}
                  </v-list-item-title>
                  <v-list-item-subtitle v-if="item.raw.notes" class="text-caption">
                    {{ truncateText(item.raw.notes, 40) }}
                  </v-list-item-subtitle>
                </v-list-item>
              </template>
            </v-autocomplete>
          </v-card>
        </div>

        <!-- Selected Template Info -->
        <v-card
          v-if="selectedTemplate"
          elevation="0"
          rounded="lg"
          class="mb-3 pa-3"
          style="background: #e3f2fd; border: 2px solid #00bcd4"
        >
          <div class="d-flex align-center mb-2">
            <v-icon color="#00bcd4" class="mr-2">mdi-information-outline</v-icon>
            <span class="font-weight-bold" style="color: #1a1a1a">
              {{ selectedTemplate.name }}
            </span>
          </div>
          <div v-if="selectedTemplate.notes" class="text-caption" style="color: #666">
            {{ selectedTemplate.notes }}
          </div>
          <div v-if="selectedTemplate.movements && selectedTemplate.movements.length > 0" class="mt-2">
            <v-chip size="x-small" color="#00bcd4" class="mr-1">
              {{ selectedTemplate.movements.length }} movement(s)
            </v-chip>
          </div>
          <div v-if="selectedTemplate.wods && selectedTemplate.wods.length > 0" class="mt-1">
            <v-chip size="x-small" color="#ffc107" class="mr-1">
              {{ selectedTemplate.wods.length }} WOD(s)
            </v-chip>
          </div>
        </v-card>

        <!-- Workout Date -->
        <div class="mb-3">
          <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
            Workout Date *
          </label>
          <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
            <v-text-field
              v-model="workoutDate"
              type="date"
              append-inner-icon="mdi-calendar"
              variant="plain"
              density="compact"
              hide-details
              required
              style="color: #1a1a1a; font-weight: 500"
            />
          </v-card>
        </div>

        <!-- Workout Type -->
        <div class="mb-3">
          <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
            Workout Type
          </label>
          <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
            <v-select
              v-model="workoutType"
              :items="workoutTypes"
              variant="plain"
              density="compact"
              hide-details
              placeholder="Select type (optional)"
              style="color: #1a1a1a; font-weight: 500"
            >
              <template #prepend-inner>
                <v-icon color="#00bcd4" size="small">mdi-format-list-bulleted</v-icon>
              </template>
            </v-select>
          </v-card>
        </div>

        <!-- Total Time -->
        <div class="mb-3">
          <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
            Total Time (minutes)
          </label>
          <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
            <v-text-field
              v-model.number="totalTimeMinutes"
              type="number"
              min="0"
              append-inner-icon="mdi-clock-outline"
              variant="plain"
              density="compact"
              hide-details
              placeholder="e.g., 30"
              style="color: #1a1a1a; font-weight: 500"
            />
          </v-card>
        </div>

        <!-- Notes -->
        <div class="mb-3">
          <label class="text-caption font-weight-bold mb-1 d-block" style="color: #1a1a1a">
            Notes
          </label>
          <v-card elevation="0" rounded="lg" class="pa-2" style="background: white">
            <v-textarea
              v-model="notes"
              variant="plain"
              density="compact"
              hide-details
              rows="3"
              placeholder="How did it feel? Any PRs? Modifications?"
              style="color: #1a1a1a; font-weight: 500"
            />
          </v-card>
        </div>

        <!-- Submit Button -->
        <v-btn
          type="submit"
          color="#00bcd4"
          size="large"
          block
          rounded="lg"
          :loading="submitting"
          :disabled="!selectedTemplateId || !workoutDate"
          class="font-weight-bold"
          style="text-transform: none"
        >
          <v-icon start>mdi-check-circle</v-icon>
          Log Workout
        </v-btn>

        <!-- Quick Browse Templates Button -->
        <v-btn
          variant="outlined"
          color="#00bcd4"
          size="large"
          block
          rounded="lg"
          class="mt-2 font-weight-bold"
          style="text-transform: none; border: 2px dashed #00bcd4"
          @click="$router.push('/workouts')"
        >
          <v-icon start>mdi-view-list</v-icon>
          Browse Templates
        </v-btn>
      </v-form>
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
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/axios'

const router = useRouter()
const route = useRoute()
const activeTab = ref('log')

// State
const workoutTemplates = ref([])
const selectedTemplateId = ref(null)
const workoutDate = ref(getTodayDate())
const workoutType = ref(null)
const totalTimeMinutes = ref(null)
const notes = ref('')

const loadingTemplates = ref(false)
const submitting = ref(false)
const error = ref(null)
const success = ref(null)

// Workout types
const workoutTypes = [
  { title: 'Strength', value: 'strength' },
  { title: 'Metcon', value: 'metcon' },
  { title: 'Cardio', value: 'cardio' },
  { title: 'Mixed', value: 'mixed' },
  { title: 'Skill Work', value: 'skill' }
]

// Computed selected template
const selectedTemplate = computed(() => {
  if (!selectedTemplateId.value) return null
  return workoutTemplates.value.find(t => t.id === selectedTemplateId.value)
})

// Get today's date in YYYY-MM-DD format
function getTodayDate() {
  const today = new Date()
  const year = today.getFullYear()
  const month = String(today.getMonth() + 1).padStart(2, '0')
  const day = String(today.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// Fetch workout templates
async function fetchTemplates() {
  loadingTemplates.value = true
  error.value = null

  try {
    // Fetch both standard and custom templates
    const [standardRes, customRes] = await Promise.all([
      axios.get('/api/workouts/standard'),
      axios.get('/api/workouts/my-templates')
    ])

    const standard = standardRes.data.workouts || []
    const custom = customRes.data.workouts || []

    // Combine and sort (standard first, then custom)
    workoutTemplates.value = [...standard, ...custom]

    console.log('Fetched templates:', workoutTemplates.value.length)
  } catch (err) {
    console.error('Failed to fetch templates:', err)
    error.value = err.response?.data?.message || 'Failed to load templates'
  } finally {
    loadingTemplates.value = false
  }
}

// Handle template selection
async function onTemplateSelected(templateId) {
  if (!templateId) return

  // Fetch template details if needed
  try {
    const response = await axios.get(`/api/workouts/${templateId}`)
    const template = response.data.workout

    // Update the template in the list with full details
    const index = workoutTemplates.value.findIndex(t => t.id === templateId)
    if (index !== -1) {
      workoutTemplates.value[index] = template
    }
  } catch (err) {
    console.error('Failed to fetch template details:', err)
  }
}

// Log workout instance
async function logWorkout() {
  if (!selectedTemplateId.value || !workoutDate.value) {
    error.value = 'Please select a template and date'
    return
  }

  submitting.value = true
  error.value = null
  success.value = null

  try {
    // Convert total time from minutes to seconds if provided
    const totalTimeSeconds = totalTimeMinutes.value ? totalTimeMinutes.value * 60 : null

    const payload = {
      workout_id: selectedTemplateId.value,
      workout_date: workoutDate.value,
      workout_type: workoutType.value || null,
      total_time: totalTimeSeconds,
      notes: notes.value.trim() || null
    }

    const response = await axios.post('/api/user-workouts', payload)

    console.log('Workout logged:', response.data)

    success.value = 'Workout logged successfully!'

    // Reset form
    selectedTemplateId.value = null
    workoutType.value = null
    totalTimeMinutes.value = null
    notes.value = ''
    workoutDate.value = getTodayDate()

    // Redirect to dashboard after short delay
    setTimeout(() => {
      router.push('/dashboard')
    }, 1500)
  } catch (err) {
    console.error('Failed to log workout:', err)
    error.value = err.response?.data?.message || 'Failed to log workout'
  } finally {
    submitting.value = false
  }
}

// Utility function to truncate text
function truncateText(text, maxLength) {
  if (!text || text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

// Load templates on mount
onMounted(async () => {
  await fetchTemplates()

  // Check if template ID is provided in query params (from WorkoutsView)
  const templateIdFromQuery = route.query.template
  if (templateIdFromQuery) {
    selectedTemplateId.value = parseInt(templateIdFromQuery)
    await onTemplateSelected(selectedTemplateId.value)
  }
})
</script>

<style scoped>
/* Additional styles if needed */
</style>
