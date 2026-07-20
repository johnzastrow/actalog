<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <!-- Success Alert -->
      <v-alert v-if="success" type="success" closable class="mb-4" @click:close="success = null">
        {{ success }}
      </v-alert>

      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable class="mb-4" @click:close="error = null">
        {{ error }}
      </v-alert>

      <!-- Loading State -->
      <div v-if="loadingWorkout" class="text-center py-8">
        <v-progress-circular indeterminate color="primary" size="64" />
        <p class="mt-4 text-body-2 text-medium-emphasis">Loading workout data...</p>
      </div>

      <v-form v-else @submit.prevent="logWorkout">
        <!-- Select Workout Template -->
        <div v-if="!isEditMode" class="mb-1">
          <label class="text-caption font-weight-bold d-block" >
            Workout Template *
          </label>
          <v-card elevation="0" rounded class="pa-1" bg-color="surface">
            <v-autocomplete
              v-model="selectedTemplateId"
              :items="workoutTemplates"
              item-title="name"
              item-value="id"
              :loading="loadingTemplates"
              :menu-props="{ location: 'top' }"
              placeholder="Search workout templates..."
              variant="plain"
              density="compact"
              hide-details
              clearable
              auto-select-first
              style="color: rgb(var(--v-theme-on-surface)); font-weight: 500"
              @update:model-value="onTemplateSelected"
            >
              <template #prepend-inner>
                <v-icon color="primary" size="small">mdi-magnify</v-icon>
              </template>
              <template #item="{ props, item }">
                <v-list-item v-bind="props" density="compact">
                  <template #prepend>
                    <v-icon
                      :color="item.raw.created_by ? 'primary' : 'warning'"
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

        <!-- Edit Workout Name (for ad-hoc workouts only) -->
        <div v-if="isEditMode && !selectedTemplateId" class="mb-1">
          <label class="text-caption font-weight-bold d-block" >
            Workout Name *
          </label>
          <v-card elevation="0" rounded class="pa-1" bg-color="surface">
            <v-text-field
              v-model="workoutName"
              variant="plain"
              density="compact"
              hide-details
              required
              placeholder="e.g., Morning Run, Upper Body, etc."
              style="color: rgb(var(--v-theme-on-surface)); font-weight: 500"
            />
          </v-card>
        </div>

        <!-- Selected Template Info -->
        <v-card
          v-if="selectedTemplate || (isEditMode && selectedTemplateId)"
          elevation="0"
          class="mb-1 pa-2"
          style="background-color: rgb(var(--v-theme-surface))"
        >
          <div class="d-flex align-center mb-2">
            <v-icon color="primary" class="mr-2">mdi-information-outline</v-icon>
            <span class="font-weight-bold" >
              {{ workoutName || selectedTemplate?.name }}
            </span>
          </div>
          <div v-if="selectedTemplate?.intro_warmup" class="text-caption text-medium-emphasis mb-1">
            <strong>Intro & Warmup:</strong> {{ truncateText(selectedTemplate.intro_warmup, 100) }}
          </div>
          <div v-if="selectedTemplate?.notes" class="text-caption text-medium-emphasis">
            <strong>Notes:</strong> {{ truncateText(selectedTemplate.notes, 100) }}
          </div>
          <div v-if="(selectedTemplate?.movements && selectedTemplate.movements.length > 0) || movementPerformance.length > 0" class="mt-2">
            <v-chip size="x-small" color="primary" class="mr-1">
              {{ selectedTemplate?.movements?.length || movementPerformance.length }} movement(s)
            </v-chip>
          </div>
          <div v-if="(selectedTemplate?.wods && selectedTemplate.wods.length > 0) || wodPerformance.length > 0" class="mt-1">
            <v-chip size="x-small" color="primary" class="mr-1">
              {{ selectedTemplate?.wods?.length || wodPerformance.length }} WOD(s)
            </v-chip>
          </div>
        </v-card>

        <!-- Workout Date -->
        <div class="mb-1">
          <label class="text-caption font-weight-bold d-block" >
            Workout Date *
          </label>
          <v-card elevation="0" rounded class="pa-1" bg-color="surface">
            <v-text-field
              v-model="workoutDate"
              type="date"
              append-inner-icon="mdi-calendar"
              variant="plain"
              density="compact"
              hide-details
              required
              style="color: rgb(var(--v-theme-on-surface)); font-weight: 500"
            />
          </v-card>
        </div>

        <!-- Movement Performance (if template has movements OR if editing and has movements) -->
        <div v-if="shouldShowMovements" class="mb-1">
          <label class="text-caption font-weight-bold d-block" >
            Movement Performance ({{ movementPerformance.length }} movements)
          </label>
          <v-card
            v-for="(movement, index) in movementPerformance"
            :key="index"
            elevation="0"
            class="mb-1 pa-2"
            style="background-color: rgb(var(--v-theme-surface));  border-radius: 8px"
          >
            <div class="d-flex align-center mb-2">
              <v-icon color="primary" size="small" class="mr-2">mdi-dumbbell</v-icon>
              <span class="font-weight-bold text-body-2" >
                {{ getMovementName(movement.movement_id) }}
              </span>
            </div>
            <v-row dense>
              <v-col cols="4">
                <v-text-field
                  v-model.number="movement.sets"
                  type="number"
                  label="Sets"
                  
                  density="compact"
                  hide-details
                  min="0"
                />
              </v-col>
              <v-col cols="4">
                <v-text-field
                  v-model.number="movement.reps"
                  type="number"
                  label="Reps"
                  
                  density="compact"
                  hide-details
                  min="0"
                />
              </v-col>
              <v-col cols="4">
                <v-text-field
                  v-model.number="movement.weight"
                  type="number"
                  label="Weight (lbs)"
                  
                  density="compact"
                  hide-details
                  min="0"
                  step="0.5"
                />
              </v-col>
            </v-row>
            <v-row dense class="mt-2">
              <v-col cols="6">
                <v-text-field
                  v-model.number="movement.time"
                  type="number"
                  label="Time (sec)"
                  
                  density="compact"
                  hide-details
                  min="0"
                />
              </v-col>
              <v-col cols="6">
                <v-text-field
                  v-model.number="movement.distance"
                  type="number"
                  label="Distance (m)"
                  
                  density="compact"
                  hide-details
                  min="0"
                  step="0.1"
                />
              </v-col>
            </v-row>
            <v-textarea
              v-model="movement.notes"
              density="compact"
              hide-details
              rows="2"
              class="mt-2"
              placeholder="Notes — how did this feel?"
            />
            <v-select
              v-model="movement.rpe"
              :items="RPE_OPTIONS"
              item-title="title"
              item-value="value"
              label="RPE (Effort)"
              density="compact"
              hide-details
              clearable
              class="mt-2"
            >
              <template #item="{ props, item }">
                <v-list-item v-bind="props" :subtitle="item.raw.description">
                  <template #prepend>
                    <v-icon :color="item.raw.color" size="small">mdi-circle</v-icon>
                  </template>
                </v-list-item>
              </template>
            </v-select>
          </v-card>
        </div>

        <!-- WOD Performance (if template has WODs OR if editing and has WODs) -->
        <div v-if="shouldShowWODs" class="mb-1">
          <label class="text-caption font-weight-bold d-block" >
            WOD Performance ({{ wodPerformance.length }} WODs)
          </label>
          <v-card
            v-for="(wod, index) in wodPerformance"
            :key="index"
            elevation="0"
            class="mb-1 pa-2"
            style="background-color: rgb(var(--v-theme-surface));  border-radius: 8px"
          >
            <div class="d-flex align-center mb-2">
              <v-icon color="primary" size="small" class="mr-2">mdi-flag-checkered</v-icon>
              <span class="font-weight-bold text-body-2" >
                {{ getWODName(wod.wod_id) }}
              </span>
            </div>

            <!-- Score Type (Read-only display) -->
            <v-text-field
              :model-value="wod.wod_score_type_display || wod.score_type"
              label="Score Type"
              
              density="compact"
              hide-details
              readonly
              class="mb-2"
              style="background-color: rgb(var(--v-theme-background))"
            />

            <!-- Time Score -->
            <div v-if="wod.score_type === 'Time (HH:MM:SS)'">
              <v-row dense>
                <v-col cols="4">
                  <v-text-field
                    v-model.number="wod.time_hours"
                    type="number"
                    label="Hours"
                    
                    density="compact"
                    hide-details
                    min="0"
                    max="23"
                  />
                </v-col>
                <v-col cols="4">
                  <v-text-field
                    v-model.number="wod.time_minutes"
                    type="number"
                    label="Minutes"
                    
                    density="compact"
                    hide-details
                    min="0"
                    max="59"
                  />
                </v-col>
                <v-col cols="4">
                  <v-text-field
                    v-model.number="wod.time_seconds"
                    type="number"
                    label="Seconds"
                    
                    density="compact"
                    hide-details
                    min="0"
                    max="59"
                  />
                </v-col>
              </v-row>
            </div>

            <!-- Rounds + Reps Score -->
            <div v-if="wod.score_type === 'Rounds+Reps'">
              <v-row dense>
                <v-col cols="6">
                  <v-text-field
                    v-model.number="wod.rounds"
                    type="number"
                    label="Rounds"
                    
                    density="compact"
                    hide-details
                    min="0"
                  />
                </v-col>
                <v-col cols="6">
                  <v-text-field
                    v-model.number="wod.reps"
                    type="number"
                    label="Reps"
                    
                    density="compact"
                    hide-details
                    min="0"
                  />
                </v-col>
              </v-row>
            </div>

            <!-- Max Weight Score -->
            <div v-if="wod.score_type === 'Max Weight'">
              <v-text-field
                v-model.number="wod.weight"
                type="number"
                label="Weight (lbs)"
                
                density="compact"
                hide-details
                min="0"
                step="0.5"
              />
            </div>

            <v-textarea
              v-model="wod.notes"
              density="compact"
              hide-details
              rows="2"
              class="mt-2"
              placeholder="Notes — how did this feel?"
            />
            <v-select
              v-model="wod.rpe"
              :items="RPE_OPTIONS"
              item-title="title"
              item-value="value"
              label="RPE (Effort)"
              density="compact"
              hide-details
              clearable
              class="mt-2"
            >
              <template #item="{ props, item }">
                <v-list-item v-bind="props" :subtitle="item.raw.description">
                  <template #prepend>
                    <v-icon :color="item.raw.color" size="small">mdi-circle</v-icon>
                  </template>
                </v-list-item>
              </template>
            </v-select>
          </v-card>
        </div>

        <!-- Total Time -->
        <div class="mb-1">
          <label class="text-caption font-weight-bold d-block" >
            Total Time (minutes)
          </label>
          <v-card elevation="0" rounded class="pa-1" bg-color="surface">
            <v-text-field
              v-model.number="totalTimeMinutes"
              type="number"
              min="0"
              append-inner-icon="mdi-clock-outline"
              variant="plain"
              density="compact"
              hide-details
              placeholder="e.g., 30"
              style="color: rgb(var(--v-theme-on-surface)); font-weight: 500"
            />
          </v-card>
        </div>

        <!-- Notes -->
        <div class="mb-1">
          <label class="text-caption font-weight-bold d-block" >
            Overall Notes
          </label>
          <v-card elevation="0" rounded class="pa-1" bg-color="surface">
            <v-textarea
              v-model="notes"
              variant="plain"
              density="compact"
              hide-details
              rows="3"
              placeholder="How did it feel? Any PRs? Modifications?"
              style="color: rgb(var(--v-theme-on-surface)); font-weight: 500"
            />
          </v-card>
        </div>

        <!-- Submit Button -->
        <v-btn
          type="submit"
          color="primary"
          size="large"
          block
          :loading="submitting"
          :disabled="(!isEditMode && !selectedTemplateId) || !workoutDate"
          class="font-weight-bold"
          style="text-transform: none"
        >
          <v-icon start>{{ isEditMode ? 'mdi-content-save' : 'mdi-check-circle' }}</v-icon>
          {{ isEditMode ? 'Update Workout' : 'Log Workout' }}
        </v-btn>

        <!-- Quick Browse Templates Button (hide if came from Quick Log with template) -->
        <v-btn
          v-if="!isEditMode && !route.query.template"
          
          color="primary"
          size="large"
          block
          class="mt-1 font-weight-bold"
          style="text-transform: none; border: 2px dashed rgb(var(--v-theme-primary))"
          @click="$router.push('/workouts')"
        >
          <v-icon start>mdi-view-list</v-icon>
          Browse Templates
        </v-btn>
      </v-form>
    </v-container>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick, inject } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/axios'
import { RPE_OPTIONS } from '@/utils/rpe'
import { useSettingsStore } from '@/stores/settings'
import { getTodayInTimezone } from '@/utils/timezone'

const refreshNotificationCount = inject('refreshNotificationCount', () => {})

const router = useRouter()
const route = useRoute()
const settingsStore = useSettingsStore()

// Edit mode state
const isEditMode = ref(false)
const editWorkoutId = ref(null)
const workoutName = ref('')
const editDataLoaded = ref(false)

// State
const workoutTemplates = ref([])
const selectedTemplateId = ref(null)
const workoutDate = ref(getTodayDate())
const totalTimeMinutes = ref(null)
const notes = ref('')
const movementPerformance = ref([])
const wodPerformance = ref([])

const loadingTemplates = ref(false)
const loadingWorkout = ref(false)
const submitting = ref(false)
const error = ref(null)
const success = ref(null)

// Computed selected template
const selectedTemplate = computed(() => {
  if (!selectedTemplateId.value) return null
  return workoutTemplates.value.find(t => t.id === selectedTemplateId.value)
})

// Computed properties for showing movement/WOD sections
const shouldShowMovements = computed(() => {
  if (isEditMode.value && editDataLoaded.value) {
    return movementPerformance.value.length > 0
  }
  return selectedTemplate.value && selectedTemplate.value.movements && selectedTemplate.value.movements.length > 0
})

const shouldShowWODs = computed(() => {
  if (isEditMode.value && editDataLoaded.value) {
    return wodPerformance.value.length > 0
  }
  return selectedTemplate.value && selectedTemplate.value.wods && selectedTemplate.value.wods.length > 0
})

// Watch for template selection to initialize performance arrays
watch(selectedTemplate, (newTemplate) => {
  // Don't initialize arrays in edit mode - we load them from existing workout data
  if (isEditMode.value) return

  if (newTemplate) {
    initializePerformanceArrays()
  }
})

// Get today's date in YYYY-MM-DD format using user's timezone
function getTodayDate() {
  return getTodayInTimezone(settingsStore.timezone)
}

// Initialize performance arrays based on template
function initializePerformanceArrays() {
  if (!selectedTemplate.value) return

  // Initialize movement performance
  if (selectedTemplate.value.movements && selectedTemplate.value.movements.length > 0) {
    movementPerformance.value = selectedTemplate.value.movements.map((m, index) => ({
      movement_id: m.movement_id,
      sets: null,
      reps: null,
      weight: null,
      time: null,
      distance: null,
      notes: '',
      rpe: null,
      order_index: index
    }))
  } else {
    movementPerformance.value = []
  }

  // Initialize WOD performance
  if (selectedTemplate.value.wods && selectedTemplate.value.wods.length > 0) {
    wodPerformance.value = selectedTemplate.value.wods.map((w, index) => {
      // Get the WOD's defined score_type (handle both nested and flattened formats)
      const wodScoreType = w.wod?.score_type || w.wod_score_type || ''
      let scoreTypeValue = null

      // Map score_type to form value
      if (wodScoreType.includes('Time')) {
        scoreTypeValue = 'Time (HH:MM:SS)'
      } else if (wodScoreType.includes('Rounds')) {
        scoreTypeValue = 'Rounds+Reps'
      } else if (wodScoreType.includes('Weight')) {
        scoreTypeValue = 'Max Weight'
      }

      return {
        wod_id: w.wod_id,
        score_type: scoreTypeValue, // Set from WOD definition
        wod_score_type_display: wodScoreType, // For display purposes
        score_value: null,
        time_hours: null,
        time_minutes: null,
        time_seconds: null,
        rounds: null,
        reps: null,
        weight: null,
        notes: '',
        rpe: null,
        order_index: index
      }
    })
  } else {
    wodPerformance.value = []
  }
}

// Get movement name by ID
function getMovementName(movementId) {
  // In edit mode, get from movementPerformance array
  if (isEditMode.value) {
    const movement = movementPerformance.value.find(m => m.movement_id === movementId)
    return movement?.movement_name || `Movement #${movementId}`
  }

  if (!selectedTemplate.value || !selectedTemplate.value.movements) return 'Movement'
  const movement = selectedTemplate.value.movements.find(m => m.movement_id === movementId)
  return movement?.movement?.name || 'Movement'
}

// Get WOD name by ID
function getWODName(wodId) {
  // In edit mode, get from wodPerformance array
  if (isEditMode.value) {
    const wod = wodPerformance.value.find(w => w.wod_id === wodId)
    return wod?.wod_name || `WOD #${wodId}`
  }

  if (!selectedTemplate.value || !selectedTemplate.value.wods) return 'WOD'
  const wod = selectedTemplate.value.wods.find(w => w.wod_id === wodId)
  // Handle both nested format (wod.name) and flattened format (wod_name)
  return wod?.wod?.name || wod?.wod_name || 'WOD'
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

    const standard = Array.isArray(standardRes.data.workouts) ? standardRes.data.workouts : []
    const custom = Array.isArray(customRes.data.workouts) ? customRes.data.workouts : []

    // Combine and sort (standard first, then custom)
    workoutTemplates.value = [...standard, ...custom]

    console.log('Fetched templates:', workoutTemplates.value.length)
  } catch (err) {
    console.error('Failed to fetch templates:', err)
    if (err.response) {
      error.value = err.response.data?.message || `Server error: ${err.response.status}`
    } else if (err.request) {
      error.value = 'Cannot connect to server. Please check if the server is running.'
    } else {
      error.value = err.message || 'Failed to load templates'
    }
  } finally {
    loadingTemplates.value = false
  }
}

// Handle template selection
async function onTemplateSelected(templateId) {
  if (!templateId) return

  // Fetch template details if needed
  try {
    const response = await axios.get(`/api/templates/${templateId}`)
    const template = response.data.template

    // Update the template in the list with full details
    const index = workoutTemplates.value.findIndex(t => t.id === templateId)
    if (index !== -1) {
      workoutTemplates.value[index] = template
    } else {
      // Template not in list, add it
      workoutTemplates.value.push(template)
    }

    // Initialize performance arrays with the template data
    initializePerformanceArrays()
  } catch (err) {
    console.error('Failed to fetch template details:', err)
  }
}

// Build movements payload
function buildMovementsPayload() {
  return movementPerformance.value
    .filter(m => m.sets || m.reps || m.weight || m.time || m.distance || m.notes || m.rpe)
    .map(m => ({
      movement_id: m.movement_id,
      sets: m.sets || null,
      reps: m.reps || null,
      weight: m.weight || null,
      time: m.time || null,
      distance: m.distance || null,
      notes: m.notes || '',
      rpe: m.rpe || null,
      order_index: m.order_index
    }))
}

// Build WODs payload
function buildWODsPayload() {
  return wodPerformance.value
    .filter(w => {
      // Must have a score_type
      if (!w.score_type) return false

      // Only include WODs where user actually entered score data
      if (w.score_type === 'Time (HH:MM:SS)') {
        return w.time_hours || w.time_minutes || w.time_seconds
      }
      if (w.score_type === 'Rounds+Reps') {
        return w.rounds != null
      }
      if (w.score_type === 'Max Weight') {
        return w.weight != null
      }

      // For other score types, include if notes or rpe is set
      return w.notes || w.rpe
    })
    .map(w => {
      const payload = {
        wod_id: w.wod_id,
        score_type: w.score_type,
        notes: w.notes || '',
        rpe: w.rpe || null,
        order_index: w.order_index
      }

      // Calculate time_seconds and score_value for Time-based WODs
      if (w.score_type === 'Time (HH:MM:SS)' && (w.time_hours || w.time_minutes || w.time_seconds)) {
        const totalSeconds = (w.time_hours || 0) * 3600 + (w.time_minutes || 0) * 60 + (w.time_seconds || 0)
        payload.time_seconds = totalSeconds
        const hours = Math.floor(totalSeconds / 3600)
        const mins = Math.floor((totalSeconds % 3600) / 60)
        const secs = totalSeconds % 60
        payload.score_value = `${String(hours).padStart(2, '0')}:${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
      }

      // Set rounds and reps for AMRAP
      if (w.score_type === 'Rounds+Reps') {
        payload.rounds = w.rounds || null
        payload.reps = w.reps || null
        payload.score_value = `${w.rounds || 0}+${w.reps || 0}`
      }

      // Set weight for Max Weight
      if (w.score_type === 'Max Weight') {
        payload.weight = w.weight || null
        payload.score_value = String(w.weight || 0)
      }

      return payload
    })
}

// Load existing workout for editing
async function loadWorkoutForEdit(workoutId) {
  loadingWorkout.value = true
  error.value = null

  try {
    const response = await axios.get(`/api/workouts/${workoutId}`)
    const workout = response.data

    console.log('Loaded workout for editing:', workout)
    console.log('Performance movements:', workout.performance_movements)
    console.log('Performance WODs:', workout.performance_wods)

    // Set basic fields
    editWorkoutId.value = workout.id
    workoutName.value = workout.workout_name
    selectedTemplateId.value = workout.workout_id
    workoutDate.value = workout.workout_date
    totalTimeMinutes.value = workout.total_time ? Math.floor(workout.total_time / 60) : null
    notes.value = workout.notes || ''

    // Load movement performance data
    if (workout.performance_movements && workout.performance_movements.length > 0) {
      movementPerformance.value = workout.performance_movements.map((m, index) => ({
        movement_id: m.movement_id,
        movement_name: m.movement?.name || 'Unknown Movement', // Store the name for display
        sets: m.sets,
        reps: m.reps,
        weight: m.weight,
        time: m.time_seconds,
        distance: m.distance,
        notes: m.notes || '',
        rpe: m.rpe || null,
        order_index: index
      }))
      console.log('Populated movementPerformance:', movementPerformance.value)
    }

    // Load WOD performance data
    if (workout.performance_wods && workout.performance_wods.length > 0) {
      wodPerformance.value = workout.performance_wods.map((w, index) => {
        // Convert total seconds to hours, minutes, seconds
        const timeHours = w.time_seconds ? Math.floor(w.time_seconds / 3600) : null
        const timeMinutes = w.time_seconds ? Math.floor((w.time_seconds % 3600) / 60) : null
        const timeSeconds = w.time_seconds ? w.time_seconds % 60 : null

        return {
          wod_id: w.wod_id,
          wod_name: w.wod?.name || 'Unknown WOD', // Store the name for display
          score_type: w.score_type,
          score_value: w.score_value,
          time_hours: timeHours,
          time_minutes: timeMinutes,
          time_seconds: timeSeconds,
          rounds: w.rounds,
          reps: w.reps,
          weight: w.weight,
          notes: w.notes || '',
          rpe: w.rpe || null,
          order_index: index
        }
      })
      console.log('Populated wodPerformance:', wodPerformance.value)
    }

    // Force reactivity update with nextTick
    await nextTick()

    // Reassign to trigger reactivity for v-for
    movementPerformance.value = [...movementPerformance.value]
    wodPerformance.value = [...wodPerformance.value]

    console.log('After nextTick - movementPerformance:', movementPerformance.value.length)
    console.log('After nextTick - wodPerformance:', wodPerformance.value.length)

    // Mark edit data as loaded to trigger section rendering
    editDataLoaded.value = true
    console.log('Edit data loaded flag set to true')
  } catch (err) {
    console.error('Failed to load workout for editing:', err)
    error.value = err.response?.data?.message || 'Failed to load workout'
  } finally {
    loadingWorkout.value = false
  }
}

// Log or update workout instance
async function logWorkout() {
  if (!isEditMode.value && (!selectedTemplateId.value || !workoutDate.value)) {
    error.value = 'Please select a template and date'
    return
  }

  if (isEditMode.value && !workoutDate.value) {
    error.value = 'Please select a date'
    return
  }

  submitting.value = true
  error.value = null
  success.value = null

  try {
    // Convert total time from minutes to seconds if provided
    const totalTimeSeconds = totalTimeMinutes.value ? totalTimeMinutes.value * 60 : null

    if (isEditMode.value) {
      // Update existing workout
      const payload = {
        total_time: totalTimeSeconds,
        notes: notes.value.trim() || null
      }

      // Add performance data if any
      const movements = buildMovementsPayload()
      const wods = buildWODsPayload()

      if (movements.length > 0) {
        payload.movements = movements
      }

      if (wods.length > 0) {
        payload.wods = wods
      }

      console.log('Updating workout with payload:', payload)

      await axios.put(`/api/workouts/${editWorkoutId.value}`, payload)

      success.value = 'Workout updated successfully!'
    } else {
      // Create new workout
      const payload = {
        workout_id: selectedTemplateId.value,
        workout_date: workoutDate.value,
        total_time: totalTimeSeconds,
        notes: notes.value.trim() || null
      }

      // Add performance data if any
      const movements = buildMovementsPayload()
      const wods = buildWODsPayload()

      if (movements.length > 0) {
        payload.movements = movements
      }

      if (wods.length > 0) {
        payload.wods = wods
      }

      console.log('Logging workout with payload:', payload)

      const response = await axios.post('/api/workouts', payload)

      console.log('Workout logged:', response.data)

      // If any movement or WOD was flagged as a PR, refresh the badge immediately
      const hasPR = response.data.performance_movements?.some(m => m.is_pr) ||
                    response.data.performance_wods?.some(w => w.is_pr)
      if (hasPR) {
        refreshNotificationCount()
      }

      success.value = 'Workout logged successfully!'
    }

    // Redirect to dashboard after short delay
    setTimeout(() => {
      if (isEditMode.value) {
        router.push(`/workouts/${editWorkoutId.value}`)
      } else {
        router.push('/dashboard')
      }
    }, 1500)
  } catch (err) {
    console.error('Failed to save workout:', err)
    error.value = err.response?.data?.message || 'Failed to save workout'
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

  // Check if editing existing workout
  const editIdFromQuery = route.query.edit
  if (editIdFromQuery) {
    isEditMode.value = true
    await loadWorkoutForEdit(parseInt(editIdFromQuery))
  }

  // Check if template ID is provided in query params (from WorkoutsView)
  const templateIdFromQuery = route.query.template
  if (templateIdFromQuery && !isEditMode.value) {
    selectedTemplateId.value = parseInt(templateIdFromQuery)
    await onTemplateSelected(selectedTemplateId.value)
  }
})
</script>

<style scoped>
/* Additional styles if needed */
</style>
