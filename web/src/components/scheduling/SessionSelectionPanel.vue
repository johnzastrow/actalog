<template>
  <div class="session-selection-panel">
    <div class="panel-header mb-3">
      <div class="section-title">
        <v-icon size="small" class="mr-2">mdi-calendar-check</v-icon>
        Select Sessions
      </div>
    </div>

    <!-- Filters -->
    <div class="filter-section mb-3">
      <v-row dense>
        <v-col cols="6">
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
        <v-col cols="6">
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
      <v-row dense class="mt-2">
        <v-col cols="6">
          <v-select
            v-model="filters.daysOfWeek"
            :items="dayOptions"
            item-title="label"
            item-value="value"
            label="Days"
            variant="solo-filled"
            density="compact"
            rounded="lg"
            flat
            hide-details
            clearable
            multiple
            chips
            closable-chips
          />
        </v-col>
        <v-col cols="6">
          <v-select
            v-model="filters.timeOfDay"
            :items="timeOfDayOptions"
            item-title="label"
            item-value="value"
            label="Time of Day"
            variant="solo-filled"
            density="compact"
            rounded="lg"
            flat
            hide-details
            clearable
          />
        </v-col>
      </v-row>
      <v-row dense class="mt-2">
        <v-col cols="6">
          <v-text-field
            v-model="filters.startDate"
            label="From Date"
            type="date"
            variant="solo-filled"
            density="compact"
            rounded="lg"
            flat
            hide-details
            clearable
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            v-model="filters.endDate"
            label="To Date"
            type="date"
            variant="solo-filled"
            density="compact"
            rounded="lg"
            flat
            hide-details
            clearable
          />
        </v-col>
      </v-row>
      <v-row dense class="mt-2">
        <v-col cols="12">
          <v-btn-toggle
            v-model="filters.hasWorkout"
            density="compact"
            color="primary"
            variant="outlined"
            divided
          >
            <v-btn :value="null" size="small">All</v-btn>
            <v-btn :value="false" size="small">Missing Workout</v-btn>
            <v-btn :value="true" size="small">Has Workout</v-btn>
          </v-btn-toggle>
        </v-col>
      </v-row>
    </div>

    <!-- Select All / Stats -->
    <div class="d-flex align-center justify-space-between mb-2">
      <v-checkbox
        v-model="selectAll"
        label="Select All"
        density="compact"
        hide-details
        :indeterminate="selectedIds.length > 0 && selectedIds.length < filteredSessions.length"
      />
      <span class="text-caption text-medium-emphasis">
        {{ selectedIds.length }} of {{ filteredSessions.length }} selected
      </span>
    </div>

    <!-- Session List -->
    <div class="session-list">
      <template v-if="filteredSessions.length > 0">
        <div
          v-for="session in filteredSessions"
          :key="session.id"
          class="session-item"
          :class="{ 'selected': selectedIds.includes(session.id) }"
          @click="toggleSession(session.id)"
        >
          <v-checkbox
            :model-value="selectedIds.includes(session.id)"
            density="compact"
            hide-details
            @click.stop="toggleSession(session.id)"
          />
          <div class="session-info flex-grow-1">
            <div class="d-flex align-center gap-2">
              <span class="font-weight-medium">{{ session.name }}</span>
              <v-chip
                v-if="session.workout_name"
                size="x-small"
                color="success"
                variant="tonal"
              >
                {{ session.workout_name }}
              </v-chip>
              <v-chip
                v-else
                size="x-small"
                color="warning"
                variant="tonal"
              >
                No workout
              </v-chip>
            </div>
            <div class="session-meta text-caption text-medium-emphasis mt-1">
              <span>{{ formatDateTime(session.start_time) }}</span>
              <span v-if="session.location_name" class="ml-2">
                <v-icon size="x-small">mdi-map-marker</v-icon>
                {{ session.location_name }}
              </span>
              <span v-if="getCoachNames(session)" class="ml-2">
                <v-icon size="x-small">mdi-account</v-icon>
                {{ getCoachNames(session) }}
              </span>
            </div>
          </div>
        </div>
      </template>
      <div v-else class="no-results text-center py-4 text-medium-emphasis">
        <v-icon size="32" class="mb-2">mdi-calendar-blank</v-icon>
        <div>No sessions match the filters</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  sessions: {
    type: Array,
    default: () => []
  },
  templates: {
    type: Array,
    default: () => []
  },
  coaches: {
    type: Array,
    default: () => []
  },
  locations: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['selection-change'])

const selectedIds = ref([])

const filters = ref({
  templateId: null,
  daysOfWeek: [],
  timeOfDay: null,
  startDate: null,
  endDate: null,
  coachId: null,
  hasWorkout: null
})

const dayOptions = [
  { value: 0, label: 'Sun' },
  { value: 1, label: 'Mon' },
  { value: 2, label: 'Tue' },
  { value: 3, label: 'Wed' },
  { value: 4, label: 'Thu' },
  { value: 5, label: 'Fri' },
  { value: 6, label: 'Sat' }
]

const timeOfDayOptions = [
  { value: 'morning', label: 'Morning (5am-12pm)' },
  { value: 'afternoon', label: 'Afternoon (12pm-5pm)' },
  { value: 'evening', label: 'Evening (5pm-10pm)' }
]

const templateOptions = computed(() => {
  return [{ id: null, name: 'All Classes' }, ...props.templates]
})

const coachOptions = computed(() => {
  return [
    { id: null, name: 'All Coaches' },
    ...props.coaches.map(c => ({ id: c.user_id, name: c.user_name || c.user_email }))
  ]
})

const selectAll = computed({
  get() {
    return filteredSessions.value.length > 0 && selectedIds.value.length === filteredSessions.value.length
  },
  set(val) {
    if (val) {
      selectedIds.value = filteredSessions.value.map(s => s.id)
    } else {
      selectedIds.value = []
    }
    emit('selection-change', selectedIds.value)
  }
})

const filteredSessions = computed(() => {
  let result = props.sessions.filter(s => s.status === 'scheduled')

  // Template filter
  if (filters.value.templateId) {
    result = result.filter(s => s.template_id === filters.value.templateId)
  }

  // Days of week filter
  if (filters.value.daysOfWeek?.length > 0) {
    result = result.filter(s => {
      const day = new Date(s.start_time).getDay()
      return filters.value.daysOfWeek.includes(day)
    })
  }

  // Time of day filter
  if (filters.value.timeOfDay) {
    result = result.filter(s => {
      const hour = new Date(s.start_time).getHours()
      switch (filters.value.timeOfDay) {
        case 'morning':
          return hour >= 5 && hour < 12
        case 'afternoon':
          return hour >= 12 && hour < 17
        case 'evening':
          return hour >= 17 || hour < 5
        default:
          return true
      }
    })
  }

  // Date range filters
  if (filters.value.startDate) {
    const startDate = new Date(filters.value.startDate)
    startDate.setHours(0, 0, 0, 0)
    result = result.filter(s => new Date(s.start_time) >= startDate)
  }

  if (filters.value.endDate) {
    const endDate = new Date(filters.value.endDate)
    endDate.setHours(23, 59, 59, 999)
    result = result.filter(s => new Date(s.start_time) <= endDate)
  }

  // Coach filter
  if (filters.value.coachId) {
    result = result.filter(s =>
      (s.coaches || []).some(c => c.user_id === filters.value.coachId)
    )
  }

  // Has workout filter
  if (filters.value.hasWorkout !== null) {
    if (filters.value.hasWorkout) {
      result = result.filter(s => s.workout_id != null)
    } else {
      result = result.filter(s => s.workout_id == null)
    }
  }

  // Sort by date
  result.sort((a, b) => new Date(a.start_time) - new Date(b.start_time))

  return result
})

function toggleSession(id) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) {
    selectedIds.value.splice(idx, 1)
  } else {
    selectedIds.value.push(id)
  }
  emit('selection-change', [...selectedIds.value])
}

function formatDateTime(dateStr) {
  const d = new Date(dateStr)
  return d.toLocaleString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  })
}

function getCoachNames(session) {
  if (!session.coaches?.length) return ''
  return session.coaches.map(c => c.user_name || c.user_email).join(', ')
}

// Clear selection when filters change
watch(filters, () => {
  selectedIds.value = selectedIds.value.filter(id =>
    filteredSessions.value.some(s => s.id === id)
  )
  emit('selection-change', [...selectedIds.value])
}, { deep: true })

// Expose method to clear selection
defineExpose({
  clearSelection: () => {
    selectedIds.value = []
    emit('selection-change', [])
  }
})
</script>

<style scoped>
.session-selection-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.section-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #546e7a;
  display: flex;
  align-items: center;
}

.filter-section {
  background-color: #f8f9fa;
  border-radius: 8px;
  padding: 12px;
}

.session-list {
  flex: 1;
  overflow-y: auto;
  max-height: 350px;
}

.session-item {
  display: flex;
  align-items: flex-start;
  padding: 10px;
  border-radius: 8px;
  margin-bottom: 6px;
  background-color: #f8f9fa;
  cursor: pointer;
  transition: all 0.2s;
}

.session-item:hover {
  background-color: #e9ecef;
}

.session-item.selected {
  background-color: #e3f2fd;
  border: 1px solid #2196f3;
}

.session-info {
  margin-left: 8px;
}

.gap-2 {
  gap: 8px;
}

.no-results {
  color: #78909c;
}
</style>
