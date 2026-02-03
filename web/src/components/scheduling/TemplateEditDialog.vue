<template>
  <v-dialog v-model="dialogOpen" max-width="900" scrollable>
    <v-card>
      <!-- Header -->
      <v-card-title class="d-flex align-center pa-4">
        <v-btn icon variant="text" class="mr-2" @click="close">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <div>
          <div class="text-h6">{{ isEditing ? 'Edit Template' : 'Create Template' }}</div>
          <div class="text-caption text-medium-emphasis">
            {{ template?.name || 'New Template' }}
          </div>
        </div>
        <v-spacer />
        <v-avatar :color="form.color || '#00bcd4'" size="32">
          <v-icon color="white" size="small">mdi-dumbbell</v-icon>
        </v-avatar>
      </v-card-title>

      <v-divider />

      <!-- Tabs -->
      <v-tabs v-model="activeTab" bg-color="surface">
        <v-tab value="details">Details</v-tab>
        <v-tab value="schedule">Schedule</v-tab>
        <v-tab value="preview">Preview</v-tab>
      </v-tabs>

      <v-divider />

      <v-card-text class="pa-0">
        <v-window v-model="activeTab">
          <!-- Details Tab -->
          <v-window-item value="details">
            <div class="pa-4">
              <v-text-field
                v-model="form.name"
                label="Template Name"
                variant="outlined"
                density="comfortable"
                required
                class="mb-3"
              />

              <v-textarea
                v-model="form.description"
                label="Description"
                variant="outlined"
                density="comfortable"
                rows="2"
                class="mb-3"
              />

              <div class="d-flex gap-3 mb-3">
                <v-text-field
                  v-model.number="form.duration_minutes"
                  label="Duration (minutes)"
                  type="number"
                  variant="outlined"
                  density="comfortable"
                  style="flex: 1"
                />
                <v-text-field
                  v-model.number="form.default_capacity"
                  label="Default Capacity"
                  type="number"
                  variant="outlined"
                  density="comfortable"
                  style="flex: 1"
                />
              </div>

              <v-text-field
                v-model="form.color"
                label="Color"
                variant="outlined"
                density="comfortable"
                class="mb-3"
              >
                <template #prepend-inner>
                  <div
                    :style="{ backgroundColor: form.color, width: '24px', height: '24px', borderRadius: '4px' }"
                  ></div>
                </template>
              </v-text-field>

              <v-select
                v-model="form.workout_id"
                :items="workouts"
                item-title="name"
                item-value="id"
                label="Default Workout (optional)"
                variant="outlined"
                density="comfortable"
                clearable
              />

              <v-switch
                v-if="isEditing"
                v-model="form.is_active"
                label="Active"
                color="primary"
              />
            </div>
          </v-window-item>

          <!-- Schedule Tab -->
          <v-window-item value="schedule">
            <div class="pa-4">
              <div class="d-flex align-center justify-space-between mb-4">
                <div class="text-subtitle-1">Schedule Slots</div>
                <v-btn color="primary" size="small" @click="addSlot">
                  <v-icon start>mdi-plus</v-icon>
                  Add Slot
                </v-btn>
              </div>

              <v-alert v-if="slots.length === 0" type="info" variant="tonal">
                No schedule slots configured. Add a slot to define when classes occur.
              </v-alert>

              <div v-for="(slot, index) in slots" :key="slot.id || index" class="mb-4">
                <v-card variant="outlined">
                  <v-card-title class="d-flex align-center py-2 px-4">
                    <span class="text-body-2">Slot {{ index + 1 }}</span>
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
                  </v-card-title>
                  <v-divider />
                  <v-card-text>
                    <RecurrencePatternEditor
                      :model-value="slots[index]"
                      @update:model-value="onSlotUpdate(index, $event)"
                    />

                    <v-select
                      v-model="slots[index].location_id"
                      :items="locations"
                      item-title="name"
                      item-value="id"
                      label="Location (optional)"
                      variant="outlined"
                      density="compact"
                      clearable
                      class="mt-4"
                    />

                    <v-text-field
                      v-model.number="slots[index].override_capacity"
                      label="Capacity Override (optional)"
                      type="number"
                      variant="outlined"
                      density="compact"
                      class="mt-3"
                    />
                  </v-card-text>
                </v-card>
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

      <v-divider />

      <v-card-actions class="pa-4">
        <v-btn variant="text" @click="close">Cancel</v-btn>
        <v-spacer />
        <v-btn
          color="primary"
          :loading="saving"
          @click="save"
        >
          {{ isEditing ? 'Save Changes' : 'Create Template' }}
        </v-btn>
      </v-card-actions>
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

// Form state
const form = ref({
  name: '',
  description: '',
  duration_minutes: 60,
  default_capacity: 20,
  color: '#00bcd4',
  workout_id: null,
  is_active: true
})

// Schedule slots
const slots = ref([])

// Preview
const previewDates = ref([])

// Session edit
const sessionEditOpen = ref(false)
const selectedSession = ref(null)

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
      is_active: newTemplate.is_active !== false
    }

    // Load slots if editing
    if (newTemplate.id) {
      await loadSlots(newTemplate.id)
    } else {
      slots.value = []
    }
  } else {
    resetForm()
  }
}, { immediate: true })

function resetForm() {
  form.value = {
    name: '',
    description: '',
    duration_minutes: 60,
    default_capacity: 20,
    color: '#00bcd4',
    workout_id: null,
    is_active: true
  }
  slots.value = []
  previewDates.value = []
  activeTab.value = 'details'
}

async function loadSlots(templateId) {
  try {
    const response = await axios.get(`/api/admin/scheduling/templates/${templateId}/slots?include_inactive=true`)
    slots.value = (response.data.slots || []).map(slot => ({
      ...slot,
      // Convert single day_of_week to days_of_week array for multi-day UI
      days_of_week: [slot.day_of_week ?? 1],
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
        // Create new slot(s) - one per selected day
        const daysToCreate = slot.days_of_week?.length > 0
          ? slot.days_of_week
          : [slot.day_of_week ?? 1]

        for (const dayOfWeek of daysToCreate) {
          const slotData = {
            day_of_week: dayOfWeek,
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
        }
      } else if (slot.id) {
        // Update existing slot (single day only - multi-day editing not supported for existing slots)
        const slotData = {
          ...slot,
          day_of_week: slot.days_of_week?.[0] ?? slot.day_of_week ?? 1,
          recurrence_end_date: slot.recurrence_end_date || null,
          effective_start_date: slot.effective_start_date || null
        }
        delete slotData._isNew
        delete slotData._deleted
        delete slotData.days_of_week // Backend doesn't support array
        await axios.put(`/api/admin/scheduling/templates/${templateId}/slots/${slot.id}`, slotData)
      }
    }

    emit('saved')
    close()
  } catch (err) {
    console.error('Failed to save template:', err)
  } finally {
    saving.value = false
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
.gap-3 {
  gap: 12px;
}
</style>
