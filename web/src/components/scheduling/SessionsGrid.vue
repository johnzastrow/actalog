<template>
  <div class="sessions-grid">
    <!-- Header with date navigation -->
    <div class="d-flex align-center justify-space-between mb-4">
      <div class="d-flex align-center">
        <v-btn icon variant="text" size="small" @click="previousWeek">
          <v-icon>mdi-chevron-left</v-icon>
        </v-btn>
        <span class="text-body-1 mx-2 font-weight-medium">{{ formatDateRange }}</span>
        <v-btn icon variant="text" size="small" @click="nextWeek">
          <v-icon>mdi-chevron-right</v-icon>
        </v-btn>
      </div>
      <div class="d-flex gap-2">
        <v-btn variant="tonal" size="small" @click="goToToday">Today</v-btn>
        <v-btn color="primary" size="small" @click="$emit('create')">
          <v-icon start>mdi-plus</v-icon>
          Create Session
        </v-btn>
      </div>
    </div>

    <!-- Filters -->
    <div class="filter-section mb-4">
      <v-row dense>
        <v-col cols="6" sm="3">
          <v-select
            v-model="filters.templateId"
            :items="templateOptions"
            item-title="name"
            item-value="id"
            label="Class"
            variant="solo-filled"
            density="compact"
            rounded="lg"
            flat
            hide-details
            clearable
          />
        </v-col>
        <v-col cols="6" sm="3">
          <v-select
            v-model="filters.locationId"
            :items="locationOptions"
            item-title="name"
            item-value="id"
            label="Location"
            variant="solo-filled"
            density="compact"
            rounded="lg"
            flat
            hide-details
            clearable
          />
        </v-col>
        <v-col cols="6" sm="3">
          <v-select
            v-model="filters.status"
            :items="statusOptions"
            item-title="label"
            item-value="value"
            label="Status"
            variant="solo-filled"
            density="compact"
            rounded="lg"
            flat
            hide-details
            clearable
          />
        </v-col>
        <v-col cols="6" sm="3">
          <v-select
            v-model="filters.coachId"
            :items="coachOptions"
            item-title="name"
            item-value="id"
            label="Coach"
            variant="solo-filled"
            density="compact"
            rounded="lg"
            flat
            hide-details
            clearable
          />
        </v-col>
      </v-row>
      <div v-if="hasActiveFilters" class="mt-2 d-flex align-center gap-2">
        <span class="text-caption text-medium-emphasis">
          Showing {{ filteredSessions.length }} of {{ sessions.length }} sessions
        </span>
        <v-btn variant="text" size="x-small" color="primary" @click="clearFilters">
          Clear filters
        </v-btn>
      </div>
    </div>

    <!-- Data Table -->
    <v-data-table
      :headers="headers"
      :items="filteredSessions"
      :loading="loading"
      class="elevation-1 sessions-table"
      density="comfortable"
      :items-per-page="-1"
      hide-default-footer
    >
      <!-- Date/Time Column -->
      <template #item.start_time="{ item }">
        <div class="py-2">
          <div class="font-weight-medium">{{ formatDate(item.start_time) }}</div>
          <div class="text-caption text-medium-emphasis">{{ formatTime(item.start_time) }} - {{ formatTime(item.end_time) }}</div>
        </div>
      </template>

      <!-- Name Column (Editable) -->
      <template #item.name="{ item }">
        <v-text-field
          v-if="editingId === item.id && editingField === 'name'"
          v-model="editValue"
          density="compact"
          variant="outlined"
          hide-details
          autofocus
          @keyup.enter="saveEdit(item)"
          @keyup.escape="cancelEdit"
          @blur="saveEdit(item)"
        />
        <span v-else class="editable-cell" @click="startEdit(item, 'name', item.name)">
          {{ item.name }}
          <v-icon size="x-small" class="edit-icon ml-1">mdi-pencil</v-icon>
        </span>
      </template>

      <!-- Workout Column (Editable with autocomplete) -->
      <template #item.workout_name="{ item }">
        <v-autocomplete
          v-if="editingId === item.id && editingField === 'workout_id'"
          v-model="editValue"
          :items="props.workouts"
          item-title="name"
          item-value="id"
          density="compact"
          variant="outlined"
          hide-details
          clearable
          autofocus
          placeholder="Search workouts..."
          style="min-width: 180px"
          @update:model-value="saveEdit(item)"
          @blur="cancelEdit"
        />
        <div v-else class="d-flex align-center gap-1 workout-cell">
          <span
            v-if="item.workout_name"
            class="editable-cell flex-grow-1"
            @click="startEdit(item, 'workout_id', item.workout_id)"
          >
            {{ item.workout_name }}
            <v-icon size="x-small" class="edit-icon ml-1">mdi-pencil</v-icon>
          </span>
          <span
            v-else
            class="editable-cell text-caption text-medium-emphasis flex-grow-1"
            @click="startEdit(item, 'workout_id', item.workout_id)"
          >
            No workout
            <v-icon size="x-small" class="edit-icon ml-1">mdi-pencil</v-icon>
          </span>
          <v-btn
            v-if="item.workout_id"
            icon
            size="x-small"
            variant="text"
            color="primary"
            @click.stop="openWorkoutPreview(item.workout_id)"
          >
            <v-icon size="small">mdi-eye</v-icon>
            <v-tooltip activator="parent" location="top">View Workout Details</v-tooltip>
          </v-btn>
        </div>
      </template>

      <!-- Location Column (Editable) -->
      <template #item.location_id="{ item }">
        <v-select
          v-if="editingId === item.id && editingField === 'location_id'"
          v-model="editValue"
          :items="locations"
          item-title="name"
          item-value="id"
          density="compact"
          variant="outlined"
          hide-details
          clearable
          autofocus
          @update:model-value="saveEdit(item)"
          @blur="cancelEdit"
        />
        <span v-else class="editable-cell" @click="startEdit(item, 'location_id', item.location_id)">
          {{ item.location_name || 'No location' }}
          <v-icon size="x-small" class="edit-icon ml-1">mdi-pencil</v-icon>
        </span>
      </template>

      <!-- Capacity Column (Editable) -->
      <template #item.capacity="{ item }">
        <v-text-field
          v-if="editingId === item.id && editingField === 'capacity'"
          v-model.number="editValue"
          type="number"
          density="compact"
          variant="outlined"
          hide-details
          style="width: 80px"
          autofocus
          @keyup.enter="saveEdit(item)"
          @keyup.escape="cancelEdit"
          @blur="saveEdit(item)"
        />
        <div v-else class="editable-cell" @click="startEdit(item, 'capacity', item.capacity)">
          <span :class="{ 'text-error': item.reservation_count >= item.capacity }">
            {{ item.reservation_count }}/{{ item.capacity }}
          </span>
          <v-icon size="x-small" class="edit-icon ml-1">mdi-pencil</v-icon>
        </div>
      </template>

      <!-- Coaches Column (Editable) -->
      <template #item.coaches="{ item }">
        <div class="d-flex align-center flex-wrap gap-1">
          <v-chip
            v-for="coach in (item.coaches || [])"
            :key="coach.user_id"
            size="small"
            closable
            :color="coach.is_lead ? 'primary' : 'grey-lighten-1'"
            class="coach-chip"
            @click:close="removeCoach(item, coach.user_id)"
            @click="toggleLeadCoach(item, coach.user_id)"
          >
            <v-icon v-if="coach.is_lead" start size="small">mdi-star</v-icon>
            {{ coach.user_name || coach.user_email }}
            <v-tooltip activator="parent">
              {{ coach.is_lead ? 'Lead Coach - Click to change' : 'Click to make Lead Coach' }}
            </v-tooltip>
          </v-chip>
          <v-menu>
            <template #activator="{ props }">
              <v-btn v-bind="props" icon size="x-small" variant="text" color="primary">
                <v-icon>mdi-plus</v-icon>
              </v-btn>
            </template>
            <v-list density="compact">
              <v-list-subheader>Add Coach</v-list-subheader>
              <v-list-item
                v-for="coach in availableCoaches(item)"
                :key="coach.user_id"
                @click="addCoach(item, coach.user_id)"
              >
                <v-list-item-title>{{ coach.user_name || coach.user_email }}</v-list-item-title>
              </v-list-item>
              <v-list-item v-if="availableCoaches(item).length === 0" disabled>
                <v-list-item-title class="text-medium-emphasis">No available coaches</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-menu>
        </div>
      </template>

      <!-- Status Column -->
      <template #item.status="{ item }">
        <v-chip :color="getStatusColor(item.status)" size="small" variant="tonal">
          {{ item.status }}
        </v-chip>
      </template>

      <!-- Actions Column -->
      <template #item.actions="{ item }">
        <div class="d-flex gap-1">
          <v-btn icon size="x-small" variant="text" title="View Roster" @click="$emit('view-roster', item)">
            <v-icon>mdi-account-multiple</v-icon>
          </v-btn>
          <v-btn
            v-if="item.status === 'scheduled'"
            icon
            size="x-small"
            variant="text"
            color="error"
            title="Cancel Session"
            @click="$emit('cancel', item)"
          >
            <v-icon>mdi-cancel</v-icon>
          </v-btn>
          <v-btn
            v-if="item.status === 'scheduled' || item.status === 'in_progress'"
            icon
            size="x-small"
            variant="text"
            color="success"
            title="Complete Session"
            @click="$emit('complete', item)"
          >
            <v-icon>mdi-check-circle</v-icon>
          </v-btn>
        </div>
      </template>

      <!-- Empty state -->
      <template #no-data>
        <div class="text-center py-8">
          <v-icon size="48" color="grey-lighten-1">mdi-calendar-blank</v-icon>
          <p class="text-body-2 mt-2 text-medium-emphasis">No sessions scheduled for this period</p>
        </div>
      </template>
    </v-data-table>

    <!-- Workout Preview Dialog -->
    <WorkoutPreviewDialog
      v-model="workoutPreviewDialog"
      :workout-id="selectedWorkoutId"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import axios from '@/utils/axios'
import WorkoutPreviewDialog from './WorkoutPreviewDialog.vue'

const props = defineProps({
  gymId: {
    type: [Number, String],
    required: true
  },
  locations: {
    type: Array,
    default: () => []
  },
  coaches: {
    type: Array,
    default: () => []
  },
  templates: {
    type: Array,
    default: () => []
  },
  workouts: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['create', 'view-roster', 'cancel', 'complete', 'updated'])

const sessions = ref([])
const loading = ref(false)
const startDate = ref(new Date())
const editingId = ref(null)
const editingField = ref(null)
const editValue = ref(null)

// Workout preview dialog state
const workoutPreviewDialog = ref(false)
const selectedWorkoutId = ref(null)

// Filter state
const filters = ref({
  templateId: null,
  locationId: null,
  status: null,
  coachId: null
})

// Filter options
const templateOptions = computed(() => {
  return [{ id: null, name: 'All Classes' }, ...props.templates]
})

const locationOptions = computed(() => {
  return [{ id: null, name: 'All Locations' }, ...props.locations]
})

const statusOptions = [
  { value: null, label: 'All Statuses' },
  { value: 'scheduled', label: 'Scheduled' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'completed', label: 'Completed' },
  { value: 'cancelled', label: 'Cancelled' }
]

const coachOptions = computed(() => {
  return [
    { id: null, name: 'All Coaches' },
    ...props.coaches.map(c => ({ id: c.user_id, name: c.user_name || c.user_email }))
  ]
})

// Filtered sessions
const filteredSessions = computed(() => {
  let result = sessions.value

  if (filters.value.templateId) {
    result = result.filter(s => s.template_id === filters.value.templateId)
  }

  if (filters.value.locationId) {
    result = result.filter(s => s.location_id === filters.value.locationId)
  }

  if (filters.value.status) {
    result = result.filter(s => s.status === filters.value.status)
  }

  if (filters.value.coachId) {
    result = result.filter(s =>
      (s.coaches || []).some(c => c.user_id === filters.value.coachId)
    )
  }

  return result
})

const hasActiveFilters = computed(() => {
  return filters.value.templateId || filters.value.locationId ||
         filters.value.status || filters.value.coachId
})

function clearFilters() {
  filters.value = {
    templateId: null,
    locationId: null,
    status: null,
    coachId: null
  }
}

// Set start date to beginning of current week (Monday)
function initStartDate() {
  const d = new Date()
  const day = d.getDay()
  const diff = d.getDate() - day + (day === 0 ? -6 : 1)
  startDate.value = new Date(d.setDate(diff))
  startDate.value.setHours(0, 0, 0, 0)
}
initStartDate()

const endDate = computed(() => {
  const d = new Date(startDate.value)
  d.setDate(d.getDate() + 6)
  return d
})

const formatDateRange = computed(() => {
  const options = { month: 'short', day: 'numeric' }
  const start = startDate.value.toLocaleDateString('en-US', options)
  const end = endDate.value.toLocaleDateString('en-US', { ...options, year: 'numeric' })
  return `${start} - ${end}`
})

const headers = [
  { title: 'Date/Time', key: 'start_time', sortable: true, width: '160px' },
  { title: 'Name', key: 'name', sortable: true },
  { title: 'Workout', key: 'workout_name', sortable: true, width: '200px' },
  { title: 'Location', key: 'location_id', sortable: false, width: '150px' },
  { title: 'Capacity', key: 'capacity', sortable: false, width: '100px' },
  { title: 'Coaches', key: 'coaches', sortable: false, width: '200px' },
  { title: 'Status', key: 'status', sortable: true, width: '100px' },
  { title: 'Actions', key: 'actions', sortable: false, width: '120px', align: 'center' }
]

function formatDate(dateStr) {
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })
}

function formatTime(dateStr) {
  const d = new Date(dateStr)
  return d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
}

function getStatusColor(status) {
  const colors = {
    scheduled: 'info',
    in_progress: 'warning',
    completed: 'success',
    cancelled: 'error'
  }
  return colors[status] || 'grey'
}

function previousWeek() {
  const d = new Date(startDate.value)
  d.setDate(d.getDate() - 7)
  startDate.value = d
}

function nextWeek() {
  const d = new Date(startDate.value)
  d.setDate(d.getDate() + 7)
  startDate.value = d
}

function goToToday() {
  initStartDate()
}

async function fetchSessions() {
  if (!props.gymId) return

  loading.value = true
  try {
    const params = new URLSearchParams({
      start_date: startDate.value.toISOString().split('T')[0],
      end_date: endDate.value.toISOString().split('T')[0]
    })
    const response = await axios.get(`/api/gyms/${props.gymId}/sessions?${params}`)
    sessions.value = response.data.sessions || []

    // Load coaches for each session
    for (const session of sessions.value) {
      try {
        const coachRes = await axios.get(`/api/admin/sessions/${session.id}/coaches`)
        session.coaches = coachRes.data.coaches || []
      } catch (err) {
        console.warn(`Failed to load coaches for session ${session.id}:`, err)
        session.coaches = []
      }
    }
  } catch (err) {
    console.error('Failed to fetch sessions:', err)
    sessions.value = []
  } finally {
    loading.value = false
  }
}

function availableCoaches(session) {
  const assignedIds = (session.coaches || []).map(c => c.user_id)
  return props.coaches.filter(c => !assignedIds.includes(c.user_id))
}

function startEdit(item, field, value) {
  if (item.status !== 'scheduled') return
  editingId.value = item.id
  editingField.value = field
  editValue.value = value
}

function cancelEdit() {
  editingId.value = null
  editingField.value = null
  editValue.value = null
}

function openWorkoutPreview(workoutId) {
  selectedWorkoutId.value = workoutId
  workoutPreviewDialog.value = true
}

async function saveEdit(item) {
  if (editValue.value === item[editingField.value]) {
    cancelEdit()
    return
  }

  try {
    const updateData = {
      name: item.name,
      description: item.description,
      workout_id: item.workout_id,
      start_time: item.start_time,
      end_time: item.end_time,
      capacity: item.capacity,
      location_id: item.location_id
    }

    updateData[editingField.value] = editValue.value

    await axios.put(`/api/admin/gyms/${props.gymId}/sessions/${item.id}`, updateData)

    // Update local data
    item[editingField.value] = editValue.value
    if (editingField.value === 'location_id') {
      const loc = props.locations.find(l => l.id === editValue.value)
      item.location_name = loc?.name || null
    }
    if (editingField.value === 'workout_id') {
      const workout = props.workouts.find(w => w.id === editValue.value)
      item.workout_name = workout?.name || null
    }

    emit('updated')
  } catch (err) {
    console.error('Failed to update session:', err)
  } finally {
    cancelEdit()
  }
}

async function addCoach(session, coachUserId) {
  try {
    await axios.post(`/api/admin/sessions/${session.id}/coaches`, {
      user_id: coachUserId,
      is_lead: (session.coaches || []).length === 0
    })

    // Reload coaches for this session
    const coachRes = await axios.get(`/api/admin/sessions/${session.id}/coaches`)
    session.coaches = coachRes.data.coaches || []
    emit('updated')
  } catch (err) {
    console.error('Failed to add coach:', err)
  }
}

async function removeCoach(session, coachUserId) {
  try {
    await axios.delete(`/api/admin/sessions/${session.id}/coaches/${coachUserId}`)
    session.coaches = (session.coaches || []).filter(c => c.user_id !== coachUserId)
    emit('updated')
  } catch (err) {
    console.error('Failed to remove coach:', err)
  }
}

async function toggleLeadCoach(session, coachUserId) {
  // Only allow toggling for scheduled sessions
  if (session.status !== 'scheduled') return

  try {
    // Set this coach as lead by re-adding with is_lead=true
    // The API will update the existing assignment
    await axios.post(`/api/admin/sessions/${session.id}/coaches`, {
      user_id: coachUserId,
      is_lead: true
    })

    // Update local state - set clicked coach as lead, remove lead from others
    for (const coach of session.coaches || []) {
      coach.is_lead = coach.user_id === coachUserId
    }
    emit('updated')
  } catch (err) {
    console.error('Failed to toggle lead coach:', err)
  }
}

// Watch for gym changes and date changes
watch([() => props.gymId, startDate], () => {
  fetchSessions()
}, { immediate: true })

// Expose refresh method
defineExpose({ refresh: fetchSessions })
</script>

<style scoped>
.sessions-grid {
  width: 100%;
}

.gap-1 {
  gap: 4px;
}

.gap-2 {
  gap: 8px;
}

.editable-cell {
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.editable-cell:hover {
  background-color: rgba(0, 0, 0, 0.04);
}

.editable-cell .edit-icon {
  opacity: 0;
  transition: opacity 0.2s;
}

.editable-cell:hover .edit-icon {
  opacity: 0.5;
}

.sessions-table :deep(.v-data-table__td) {
  vertical-align: middle;
}

.filter-section {
  background-color: rgba(0, 0, 0, 0.02);
  border-radius: 8px;
  padding: 12px 16px;
}

.coach-chip {
  cursor: pointer;
  transition: transform 0.1s ease;
}

.coach-chip:hover {
  transform: scale(1.02);
}

.workout-cell {
  min-width: 120px;
}

.workout-cell .editable-cell {
  flex-grow: 1;
}
</style>
