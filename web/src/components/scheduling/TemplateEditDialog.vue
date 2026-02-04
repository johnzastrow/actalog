<template>
  <v-dialog v-model="dialogOpen" max-width="900">
    <v-card class="d-flex flex-column template-edit-dialog" style="max-height: 90vh;">
      <!-- Header - Fixed -->
      <div class="dialog-header pa-4 d-flex align-center">
        <v-btn icon variant="text" color="white" class="mr-2" @click="close">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <div class="flex-grow-1">
          <div class="text-h6 font-weight-bold text-white">{{ isEditing ? 'Edit Class' : 'Create Class' }}</div>
          <div class="text-caption text-white" style="opacity: 0.8;">
            {{ template?.name || 'New Class' }}
          </div>
        </div>
        <v-btn icon variant="text" color="white" @click="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </div>

      <!-- Tabs - Fixed -->
      <v-tabs v-model="activeTab" bg-color="grey-lighten-4" class="flex-shrink-0 px-2" slider-color="primary">
        <v-tab value="details" class="text-none">
          <v-icon start size="small">mdi-text-box</v-icon>
          Details
        </v-tab>
        <v-tab value="schedule" class="text-none">
          <v-icon start size="small">mdi-calendar-clock</v-icon>
          Schedule
        </v-tab>
        <v-tab value="preview" class="text-none">
          <v-icon start size="small">mdi-calendar-month</v-icon>
          Preview
        </v-tab>
      </v-tabs>

      <!-- Scrollable Content -->
      <v-card-text class="pa-0 flex-grow-1 overflow-y-auto bg-grey-lighten-4">
        <v-window v-model="activeTab">
          <!-- Details Tab -->
          <v-window-item value="details">
            <div class="pa-4">
              <!-- Basic Info Section -->
              <div class="form-section mb-4">
                <div class="section-title">
                  <v-icon size="small" class="mr-2">mdi-information</v-icon>
                  Basic Information
                </div>
                <v-text-field
                  v-model="form.name"
                  label="Class Name"
                  variant="solo-filled"
                  density="comfortable"
                  rounded="lg"
                  flat
                  required
                  class="mb-3"
                />

                <v-textarea
                  v-model="form.description"
                  label="Description"
                  variant="solo-filled"
                  density="comfortable"
                  rounded="lg"
                  flat
                  rows="2"
                  class="mb-3"
                />

                <v-row dense>
                  <v-col cols="6">
                    <v-text-field
                      v-model.number="form.duration_minutes"
                      label="Duration (minutes)"
                      type="number"
                      variant="solo-filled"
                      density="comfortable"
                      rounded="lg"
                      flat
                    />
                  </v-col>
                  <v-col cols="6">
                    <v-text-field
                      v-model.number="form.default_capacity"
                      label="Default Capacity"
                      type="number"
                      variant="solo-filled"
                      density="comfortable"
                      rounded="lg"
                      flat
                    />
                  </v-col>
                </v-row>
              </div>

              <!-- Appearance Section -->
              <div class="form-section mb-4">
                <div class="section-title">
                  <v-icon size="small" class="mr-2">mdi-palette</v-icon>
                  Appearance
                </div>
                <div class="d-flex align-center gap-3 mb-2">
                  <div class="text-body-2 text-medium-emphasis">Class Color</div>
                  <v-avatar :color="form.color || '#00bcd4'" size="32">
                    <v-icon color="white" size="small">mdi-dumbbell</v-icon>
                  </v-avatar>
                </div>
                <v-color-picker
                  v-model="form.color"
                  :swatches="colorSwatches"
                  show-swatches
                  hide-inputs
                  hide-canvas
                  swatches-max-height="120"
                  elevation="0"
                  rounded="lg"
                  class="color-picker-compact"
                />
              </div>

              <!-- Defaults Section -->
              <div class="form-section mb-4">
                <div class="section-title">
                  <v-icon size="small" class="mr-2">mdi-cog</v-icon>
                  Default Settings
                </div>
                <v-select
                  v-model="form.workout_id"
                  :items="workouts"
                  item-title="name"
                  item-value="id"
                  label="Default Workout (optional)"
                  variant="solo-filled"
                  density="comfortable"
                  rounded="lg"
                  flat
                  clearable
                  class="mb-3"
                />

                <v-select
                  v-model="form.default_location_id"
                  :items="locations"
                  item-title="name"
                  item-value="id"
                  label="Default Location (optional)"
                  variant="solo-filled"
                  density="comfortable"
                  rounded="lg"
                  flat
                  clearable
                />
              </div>

              <!-- Coaches Section -->
              <div class="form-section mb-4">
                <div class="section-title">
                  <v-icon size="small" class="mr-2">mdi-account-tie</v-icon>
                  Default Coaches
                </div>
                <div class="d-flex flex-wrap gap-2 mb-3">
                  <v-chip
                    v-for="coach in defaultCoaches"
                    :key="coach.user_id"
                    closable
                    :color="coach.is_lead ? 'primary' : 'grey-lighten-1'"
                    :variant="coach.is_lead ? 'flat' : 'flat'"
                    rounded="lg"
                    @click:close="removeCoach(coach.user_id)"
                    @click="toggleLeadCoach(coach.user_id)"
                  >
                    <v-icon v-if="coach.is_lead" start size="small">mdi-star</v-icon>
                    {{ coach.user_name || coach.user_email }}
                  </v-chip>
                  <v-chip v-if="defaultCoaches.length === 0" variant="tonal" color="grey" size="small" rounded="lg">
                    No coaches assigned
                  </v-chip>
                </div>
                <v-autocomplete
                  v-model="selectedCoach"
                  v-model:search="coachSearch"
                  :items="coachSearchResults"
                  :loading="searchingCoaches"
                  item-title="displayName"
                  item-value="id"
                  label="Add Coach"
                  variant="solo-filled"
                  density="comfortable"
                  rounded="lg"
                  flat
                  clearable
                  no-filter
                  placeholder="Search by name or email..."
                  hide-no-data
                  @update:model-value="onCoachSelected"
                >
                  <template #item="{ item, props: itemProps }">
                    <v-list-item v-bind="itemProps">
                      <template #subtitle>{{ item.raw.email }}</template>
                    </v-list-item>
                  </template>
                </v-autocomplete>
                <div class="text-caption text-medium-emphasis mt-1">
                  Click a coach chip to toggle lead status
                </div>
              </div>

              <!-- Status Section -->
              <div v-if="isEditing" class="form-section">
                <div class="section-title">
                  <v-icon size="small" class="mr-2">mdi-toggle-switch</v-icon>
                  Status
                </div>
                <v-switch
                  v-model="form.is_active"
                  label="Active"
                  color="primary"
                  hide-details
                  inset
                />
              </div>
            </div>
          </v-window-item>

          <!-- Schedule Tab -->
          <v-window-item value="schedule">
            <div class="pa-4">
              <div class="d-flex align-center justify-space-between mb-4">
                <div class="text-subtitle-1 font-weight-medium">Schedule Slots</div>
                <v-btn color="primary" size="small" rounded="lg" @click="addSlot">
                  <v-icon start>mdi-plus</v-icon>
                  Add Slot
                </v-btn>
              </div>

              <v-alert v-if="slots.length === 0" type="info" variant="tonal" rounded="lg">
                No schedule slots configured. Add a slot to define when classes occur.
              </v-alert>

              <div v-for="(slot, index) in slots" :key="slot.id || index" class="mb-4">
                <div class="slot-card">
                  <div class="slot-header d-flex align-center">
                    <v-icon size="small" class="mr-2">mdi-clock-outline</v-icon>
                    <span class="text-body-2 font-weight-medium">Slot {{ index + 1 }}</span>
                    <v-spacer />
                    <v-btn
                      icon
                      variant="text"
                      size="small"
                      color="error"
                      @click="removeSlot(index)"
                    >
                      <v-icon>mdi-delete</v-icon>
                    </v-btn>
                  </div>
                  <div class="slot-content">
                    <RecurrencePatternEditor
                      :model-value="slots[index]"
                      @update:model-value="onSlotUpdate(index, $event)"
                    />

                    <v-text-field
                      v-model.number="slots[index].override_capacity"
                      label="Capacity Override (optional)"
                      type="number"
                      variant="solo-filled"
                      density="comfortable"
                      rounded="lg"
                      flat
                      class="mt-4"
                    />
                  </div>
                </div>
              </div>
            </div>
          </v-window-item>

          <!-- Preview Tab -->
          <v-window-item value="preview">
            <div class="pa-4">
              <v-btn
                color="primary"
                variant="tonal"
                class="mb-4"
                rounded="lg"
                :loading="loadingPreview"
                @click="loadPreview"
              >
                <v-icon start>mdi-refresh</v-icon>
                Generate Preview
              </v-btn>

              <ScheduleCalendarPreview
                :scheduled-dates="previewDates"
                :template-color="form.color || '#00bcd4'"
                :month-count="3"
                :horizontal="!isMobile"
                @day-click="onPreviewDayClick"
              />
            </div>
          </v-window-item>
        </v-window>
      </v-card-text>

      <!-- Footer -->
      <div class="dialog-footer pa-4 d-flex align-center">
        <v-btn variant="text" rounded="lg" @click="close">Cancel</v-btn>
        <v-spacer />
        <v-btn
          color="primary"
          rounded="lg"
          :loading="saving"
          @click="save"
        >
          <v-icon start>mdi-content-save</v-icon>
          {{ isEditing ? 'Save Changes' : 'Create Class' }}
        </v-btn>
      </div>
    </v-card>

    <!-- Session Edit Dialog -->
    <SessionEditDialog
      v-model="sessionEditOpen"
      :session="selectedSession"
      :locations="locations"
      :workouts="workouts"
      @save="onSessionSave"
      @cancel="onSessionCancel"
    />
  </v-dialog>

  <!-- Success Snackbar (outside dialog so it persists after close) -->
  <v-snackbar
    v-model="showSuccessSnackbar"
    color="success"
    :timeout="3000"
    location="bottom"
  >
    <v-icon class="mr-2">mdi-check-circle</v-icon>
    Class saved successfully
    <template #actions>
      <v-btn variant="text" @click="showSuccessSnackbar = false">Close</v-btn>
    </template>
  </v-snackbar>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useDisplay } from 'vuetify'
import axios from '@/utils/axios'
import RecurrencePatternEditor from './RecurrencePatternEditor.vue'
import ScheduleCalendarPreview from './ScheduleCalendarPreview.vue'
import SessionEditDialog from './SessionEditDialog.vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  template: {
    type: Object,
    default: null
  },
  gymId: {
    type: [Number, String],
    required: true
  },
  locations: {
    type: Array,
    default: () => []
  },
  workouts: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue', 'saved'])

const { mobile: isMobile } = useDisplay()

const activeTab = ref('details')
const saving = ref(false)
const loadingPreview = ref(false)

// Color swatches for picker
const colorSwatches = [
  ['#00bcd4', '#03a9f4', '#2196f3', '#3f51b5', '#673ab7'],
  ['#9c27b0', '#e91e63', '#f44336', '#ff5722', '#ff9800'],
  ['#ffc107', '#cddc39', '#8bc34a', '#4caf50', '#009688'],
  ['#795548', '#607d8b', '#9e9e9e', '#455a64', '#263238']
]

// Form state
const form = ref({
  name: '',
  description: '',
  duration_minutes: 60,
  default_capacity: 20,
  color: '#00bcd4',
  workout_id: null,
  default_location_id: null,
  is_active: true
})

// Default coaches
const defaultCoaches = ref([])
const originalCoachesSnapshot = ref([]) // Track original coaches for save comparison
const selectedCoach = ref(null)
const coachSearch = ref('')
const coachSearchResults = ref([])
const searchingCoaches = ref(false)
let coachSearchTimeout = null

// Schedule slots
const slots = ref([])

// Preview
const previewDates = ref([])

// Session edit
const sessionEditOpen = ref(false)
const selectedSession = ref(null)

// Success snackbar
const showSuccessSnackbar = ref(false)

const dialogOpen = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const isEditing = computed(() => !!props.template?.id)

// Watch template changes
watch(() => props.template, async (newTemplate) => {
  if (newTemplate) {
    form.value = {
      name: newTemplate.name || '',
      description: newTemplate.description || '',
      duration_minutes: newTemplate.duration_minutes || 60,
      default_capacity: newTemplate.default_capacity || 20,
      color: newTemplate.color || '#00bcd4',
      workout_id: newTemplate.workout_id || null,
      default_location_id: newTemplate.default_location_id || null,
      is_active: newTemplate.is_active !== false
    }

    // Load slots and coaches if editing
    if (newTemplate.id) {
      await Promise.all([
        loadSlots(newTemplate.id),
        loadCoaches(newTemplate.id)
      ])
    } else {
      slots.value = []
      defaultCoaches.value = []
      originalCoachesSnapshot.value = []
    }
  } else {
    resetForm()
  }
}, { immediate: true })

// Reset form when dialog opens for a new class
watch(dialogOpen, (isOpen) => {
  if (isOpen && !props.template) {
    resetForm()
  }
})

// Watch coach search for autocomplete
watch(coachSearch, (search) => {
  if (coachSearchTimeout) {
    clearTimeout(coachSearchTimeout)
  }
  if (!search || search.length < 2) {
    coachSearchResults.value = []
    return
  }
  coachSearchTimeout = setTimeout(() => {
    searchCoaches(search)
  }, 300)
})

// Auto-generate preview when switching to Preview tab
watch(activeTab, (newTab) => {
  if (newTab === 'preview' && slots.value.length > 0) {
    if (props.template?.id) {
      loadPreview()
    } else {
      generateLocalPreview()
    }
  }
})

function resetForm() {
  form.value = {
    name: '',
    description: '',
    duration_minutes: 60,
    default_capacity: 20,
    color: '#00bcd4',
    workout_id: null,
    default_location_id: null,
    is_active: true
  }
  slots.value = []
  defaultCoaches.value = []
  originalCoachesSnapshot.value = []
  previewDates.value = []
  activeTab.value = 'details'
  coachSearch.value = ''
  coachSearchResults.value = []
}

async function loadCoaches(templateId) {
  try {
    const response = await axios.get(`/api/admin/scheduling/templates/${templateId}/coaches`)
    const coaches = response.data.coaches || []
    defaultCoaches.value = coaches
    // Save a snapshot to compare on save (deep copy)
    originalCoachesSnapshot.value = coaches.map(c => ({ ...c }))
  } catch (err) {
    console.error('Failed to load coaches:', err)
    defaultCoaches.value = []
    originalCoachesSnapshot.value = []
  }
}

async function searchCoaches(search) {
  searchingCoaches.value = true
  try {
    const response = await axios.get('/api/admin/user-management/filter', {
      params: { q: search }
    })
    // Filter out already assigned coaches
    const assignedIds = new Set(defaultCoaches.value.map(c => c.user_id))
    coachSearchResults.value = (response.data.users || [])
      .filter(u => !assignedIds.has(u.id))
      .map(u => ({
        id: u.id,
        email: u.email,
        name: u.name || '',
        displayName: u.name ? `${u.name} (${u.email})` : u.email
      }))
  } catch (err) {
    console.error('Failed to search coaches:', err)
    coachSearchResults.value = []
  } finally {
    searchingCoaches.value = false
  }
}

function onCoachSelected(userId) {
  if (!userId) return

  const coach = coachSearchResults.value.find(c => c.id === userId)
  if (coach) {
    // Add to local list (will be saved when template is saved)
    defaultCoaches.value.push({
      user_id: coach.id,
      user_name: coach.name,
      user_email: coach.email,
      is_lead: defaultCoaches.value.length === 0, // First coach becomes lead
      _isNew: true
    })
  }

  // Clear selection
  selectedCoach.value = null
  coachSearch.value = ''
  coachSearchResults.value = []
}

function removeCoach(userId) {
  const index = defaultCoaches.value.findIndex(c => c.user_id === userId)
  if (index !== -1) {
    const coach = defaultCoaches.value[index]
    if (coach._isNew) {
      // Just remove from local list
      defaultCoaches.value.splice(index, 1)
    } else {
      // Mark for deletion
      coach._deleted = true
      defaultCoaches.value.splice(index, 1)
    }

    // If we removed the lead, make first remaining coach the lead
    if (coach.is_lead && defaultCoaches.value.length > 0) {
      defaultCoaches.value[0].is_lead = true
    }
  }
}

function toggleLeadCoach(userId) {
  // Set this coach as lead, remove lead from others
  for (const coach of defaultCoaches.value) {
    coach.is_lead = coach.user_id === userId
  }
}

async function loadSlots(templateId) {
  try {
    const response = await axios.get(`/api/admin/scheduling/templates/${templateId}/slots?include_inactive=true`)
    slots.value = (response.data.slots || []).map(slot => ({
      ...slot,
      // Use days_of_week array from backend, fallback to single day_of_week
      days_of_week: slot.days_of_week?.length > 0 ? [...slot.days_of_week] : [slot.day_of_week ?? 1],
      // Convert dates to string format for the editor
      recurrence_end_date: slot.recurrence_end_date ? slot.recurrence_end_date.split('T')[0] : '',
      effective_start_date: slot.effective_start_date ? slot.effective_start_date.split('T')[0] : ''
    }))
  } catch (err) {
    console.error('Failed to load slots:', err)
    slots.value = []
  }
}

function addSlot() {
  slots.value.push({
    days_of_week: [1], // Monday (array for multi-day support)
    day_of_week: 1, // Legacy single day support
    start_time: '09:00',
    recurrence_interval: 1,
    recurrence_end_type: 'none',
    recurrence_end_count: null,
    recurrence_end_date: '',
    effective_start_date: '',
    location_id: null,
    override_capacity: null,
    is_active: true,
    _isNew: true
  })
}

function removeSlot(index) {
  const slot = slots.value[index]
  if (slot._isNew) {
    slots.value.splice(index, 1)
  } else {
    // Mark for deletion
    slot._deleted = true
    slots.value.splice(index, 1)
  }
}

function onSlotUpdate(index, updatedSlot) {
  slots.value[index] = {
    ...slots.value[index],
    ...updatedSlot
  }
  // Auto-regenerate preview when slots change
  if (activeTab.value === 'preview') {
    generateLocalPreview()
  }
}

async function loadPreview() {
  if (!props.template?.id) {
    // For new templates, generate preview from local slots
    generateLocalPreview()
    return
  }

  loadingPreview.value = true
  try {
    const response = await axios.get(`/api/admin/scheduling/templates/${props.template.id}/preview-schedule?months=3`)
    previewDates.value = response.data.dates || []
  } catch (err) {
    console.error('Failed to load preview:', err)
    previewDates.value = []
  } finally {
    loadingPreview.value = false
  }
}

function generateLocalPreview() {
  // Generate preview from local slots (for new templates)
  const dates = []
  const today = new Date()
  const endDate = new Date(today)
  endDate.setMonth(endDate.getMonth() + 3)

  for (const slot of slots.value) {
    if (slot._deleted) continue

    const interval = slot.recurrence_interval || 1
    // Support both multi-day array and legacy single day
    const daysOfWeek = slot.days_of_week?.length > 0
      ? slot.days_of_week
      : [slot.day_of_week ?? 0]
    const [hours, minutes] = (slot.start_time || '09:00').split(':').map(Number)

    // Generate dates for each selected day
    for (const dayOfWeek of daysOfWeek) {
      // Find first matching day
      let date = new Date(today)
      while (date.getDay() !== dayOfWeek) {
        date.setDate(date.getDate() + 1)
      }

      // Effective start date
      if (slot.effective_start_date) {
        const effectiveStart = new Date(slot.effective_start_date)
        if (effectiveStart > date) {
          date = new Date(effectiveStart)
          while (date.getDay() !== dayOfWeek) {
            date.setDate(date.getDate() + 1)
          }
        }
      }

      // Generate dates
      let count = 0
      const maxCount = slot.recurrence_end_type === 'count' ? (slot.recurrence_end_count || 100) : 100
      const endDateLimit = slot.recurrence_end_type === 'date' && slot.recurrence_end_date
        ? new Date(slot.recurrence_end_date)
        : endDate

      while (date <= endDateLimit && date <= endDate && count < maxCount) {
        const sessionDate = new Date(date)
        sessionDate.setHours(hours, minutes, 0, 0)
        dates.push(sessionDate.toISOString())
        count++
        date.setDate(date.getDate() + 7 * interval)
      }
    }
  }

  previewDates.value = dates
}

function onPreviewDayClick(date) {
  // Could open session details or create session for that date
  // TODO: Open session edit dialog for this date
}

function close() {
  dialogOpen.value = false
}

async function save() {
  if (!form.value.name?.trim()) {
    // Show error
    return
  }

  saving.value = true
  try {
    let templateId = props.template?.id

    // Save template
    if (isEditing.value) {
      await axios.put(`/api/admin/scheduling/templates/${templateId}`, form.value)
    } else {
      const response = await axios.post(`/api/admin/gyms/${props.gymId}/templates`, form.value)
      templateId = response.data.id
    }

    // Save slots
    for (const slot of slots.value) {
      if (slot._deleted && slot.id) {
        // Delete existing slot
        await axios.delete(`/api/admin/scheduling/templates/${templateId}/slots/${slot.id}`)
      } else if (slot._isNew) {
        // Create new slot with days_of_week array
        const slotData = {
          days_of_week: slot.days_of_week?.length > 0 ? slot.days_of_week : [slot.day_of_week ?? 1],
          day_of_week: slot.days_of_week?.[0] ?? slot.day_of_week ?? 1,
          start_time: slot.start_time,
          recurrence_interval: slot.recurrence_interval,
          recurrence_end_type: slot.recurrence_end_type,
          recurrence_end_count: slot.recurrence_end_count,
          recurrence_end_date: slot.recurrence_end_date || null,
          effective_start_date: slot.effective_start_date || null,
          location_id: slot.location_id,
          override_capacity: slot.override_capacity,
          is_active: slot.is_active !== false
        }
        await axios.post(`/api/admin/scheduling/templates/${templateId}/slots`, slotData)
      } else if (slot.id) {
        // Update existing slot with days_of_week array
        const slotData = {
          days_of_week: slot.days_of_week?.length > 0 ? slot.days_of_week : [slot.day_of_week ?? 1],
          day_of_week: slot.days_of_week?.[0] ?? slot.day_of_week ?? 1,
          start_time: slot.start_time,
          recurrence_interval: slot.recurrence_interval,
          recurrence_end_type: slot.recurrence_end_type,
          recurrence_end_count: slot.recurrence_end_count,
          recurrence_end_date: slot.recurrence_end_date || null,
          effective_start_date: slot.effective_start_date || null,
          location_id: slot.location_id,
          override_capacity: slot.override_capacity,
          is_active: slot.is_active !== false
        }
        await axios.put(`/api/admin/scheduling/templates/${templateId}/slots/${slot.id}`, slotData)
      }
    }

    // Save coaches
    // Use the snapshot of coaches loaded when dialog opened
    const originalCoaches = originalCoachesSnapshot.value
    const originalCoachIds = new Set(originalCoaches.map(c => c.user_id))
    const currentCoachIds = new Set(defaultCoaches.value.map(c => c.user_id))

    // Delete coaches that were removed
    for (const original of originalCoaches) {
      if (!currentCoachIds.has(original.user_id)) {
        await axios.delete(`/api/admin/scheduling/templates/${templateId}/coaches/${original.user_id}`)
      }
    }

    // Add new coaches and update lead status
    for (const coach of defaultCoaches.value) {
      if (coach._isNew || !originalCoachIds.has(coach.user_id)) {
        // Add new coach
        await axios.post(`/api/admin/scheduling/templates/${templateId}/coaches`, {
          user_id: coach.user_id,
          is_lead: coach.is_lead
        })
      } else {
        // Update lead status if changed
        const original = originalCoaches.find(c => c.user_id === coach.user_id)
        if (original && original.is_lead !== coach.is_lead) {
          // Re-add with new lead status (API handles upsert)
          await axios.post(`/api/admin/scheduling/templates/${templateId}/coaches`, {
            user_id: coach.user_id,
            is_lead: coach.is_lead
          })
        }
      }
    }

    // Show success snackbar
    showSuccessSnackbar.value = true

    console.log('[TemplateEditDialog] Emitting saved event')
    emit('saved')

    // Reload data from database to refresh the form
    await reloadTemplateData(templateId)
  } catch (err) {
    console.error('Failed to save template:', err)
  } finally {
    saving.value = false
  }
}

async function reloadTemplateData(templateId) {
  try {
    // Fetch the updated template from the database
    const response = await axios.get(`/api/admin/scheduling/templates/${templateId}`)
    const template = response.data

    // Update form with fresh data
    form.value = {
      name: template.name || '',
      description: template.description || '',
      duration_minutes: template.duration_minutes || 60,
      default_capacity: template.default_capacity || 20,
      color: template.color || '#00bcd4',
      workout_id: template.workout_id || null,
      default_location_id: template.default_location_id || null,
      is_active: template.is_active !== false
    }

    // Reload slots and coaches
    await Promise.all([
      loadSlots(templateId),
      loadCoaches(templateId)
    ])
  } catch (err) {
    console.error('Failed to reload template data:', err)
  }
}

function onSessionSave(sessionData) {
  // Handle session save - TODO: implement
}

function onSessionCancel(session) {
  // Handle session cancel - TODO: implement
}
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
.gap-3 {
  gap: 12px;
}

/* Modern Dialog Styling */
.template-edit-dialog {
  border-radius: 16px !important;
  overflow: hidden;
}

.dialog-header {
  background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
  color: white;
}

.dialog-footer {
  background-color: #f5f5f5;
  border-top: 1px solid #e0e0e0;
}

/* Form Sections */
.form-section {
  background-color: white;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.section-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #546e7a;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
}

/* Schedule Slot Cards */
.slot-card {
  background-color: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.slot-header {
  background-color: #f8f9fa;
  padding: 12px 16px;
  border-bottom: 1px solid #e9ecef;
}

.slot-content {
  padding: 16px;
}

/* Compact Color Picker */
.color-picker-compact {
  width: 100% !important;
  max-width: 100% !important;
}

.color-picker-compact :deep(.v-color-picker-swatches) {
  max-height: none !important;
}
</style>
