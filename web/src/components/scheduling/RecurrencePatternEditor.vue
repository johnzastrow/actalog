<template>
  <div class="recurrence-editor">
    <div class="section-label mb-3">
      <v-icon size="small" class="mr-2">mdi-repeat</v-icon>
      Recurrence Pattern
    </div>

    <!-- Frequency -->
    <div class="d-flex align-center gap-3 mb-4">
      <span class="text-body-2">Every</span>
      <v-text-field
        v-model.number="interval"
        type="number"
        min="1"
        max="12"
        density="compact"
        variant="solo-filled"
        rounded="lg"
        flat
        hide-details
        style="max-width: 80px"
      />
      <span class="text-body-2">week(s)</span>
      <v-chip
        :color="interval === 1 ? 'primary' : 'secondary'"
        size="small"
        variant="flat"
        rounded="lg"
      >
        {{ interval === 1 ? 'Weekly' : interval === 2 ? 'Bi-weekly' : `Every ${interval} weeks` }}
      </v-chip>
    </div>

    <!-- Days of Week (multiple selection) -->
    <div class="mb-4">
      <DayOfWeekSelector
        :modelValue="daysOfWeek"
        @update:modelValue="onDaysChange"
        label="On"
        :multiple="true"
      />
    </div>

    <!-- Time -->
    <v-text-field
      v-model="startTime"
      label="Start Time"
      type="time"
      density="compact"
      variant="solo-filled"
      rounded="lg"
      flat
      class="mb-4"
      hide-details
    />

    <!-- Effective Start Date -->
    <v-text-field
      v-model="effectiveStartDate"
      label="Start Date (optional)"
      type="date"
      density="compact"
      variant="solo-filled"
      rounded="lg"
      flat
      class="mb-4"
      hide-details
      hint="When this schedule takes effect"
    />

    <!-- End Options -->
    <div class="text-caption text-medium-emphasis mb-2">End</div>
    <v-radio-group v-model="endType" hide-details class="mt-0">
      <v-radio value="none" label="Never" />
      <v-radio value="count">
        <template #label>
          <div class="d-flex align-center gap-2">
            <span>After</span>
            <v-text-field
              v-model.number="endCount"
              type="number"
              min="1"
              max="100"
              density="compact"
              variant="solo-filled"
              rounded="lg"
              flat
              hide-details
              :disabled="endType !== 'count'"
              style="max-width: 80px"
            />
            <span>occurrences</span>
          </div>
        </template>
      </v-radio>
      <v-radio value="date">
        <template #label>
          <div class="d-flex align-center gap-2">
            <span>On</span>
            <v-text-field
              v-model="endDate"
              type="date"
              density="compact"
              variant="solo-filled"
              rounded="lg"
              flat
              hide-details
              :disabled="endType !== 'date'"
              style="max-width: 160px"
            />
          </div>
        </template>
      </v-radio>
    </v-radio-group>

    <!-- Summary Text -->
    <v-alert type="info" variant="tonal" density="compact" class="mt-4" rounded="lg">
      <div class="text-body-2">{{ summaryTextValue }}</div>
    </v-alert>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import DayOfWeekSelector from './DayOfWeekSelector.vue'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({
      days_of_week: [1], // Array of days (Monday default)
      day_of_week: 1, // Legacy single day support
      start_time: '09:00',
      recurrence_interval: 1,
      recurrence_end_type: 'none',
      recurrence_end_count: null,
      recurrence_end_date: null,
      effective_start_date: null
    })
  }
})

const emit = defineEmits(['update:modelValue'])

const dayNames = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']

// Local state - support both legacy single day and new multi-day
const daysOfWeek = ref(
  props.modelValue.days_of_week?.length > 0
    ? [...props.modelValue.days_of_week]
    : (props.modelValue.day_of_week !== undefined ? [props.modelValue.day_of_week] : [1])
)
const startTime = ref(props.modelValue.start_time ?? '09:00')
const interval = ref(props.modelValue.recurrence_interval ?? 1)
const endType = ref(props.modelValue.recurrence_end_type ?? 'none')
const endCount = ref(props.modelValue.recurrence_end_count ?? 10)
const endDate = ref(props.modelValue.recurrence_end_date ?? '')
const effectiveStartDate = ref(props.modelValue.effective_start_date ?? '')

// Flag to track when we're syncing from props (should not emit back)
let isSyncingFromProps = false

// Handler for days change from child component
function onDaysChange(newDays) {
  const newArray = Array.isArray(newDays) ? [...newDays] : [newDays]

  // Check if actually changed to prevent unnecessary updates
  if (arraysEqual(daysOfWeek.value, newArray)) return

  daysOfWeek.value = newArray
}

// Format selected days for summary
function formatSelectedDays(days) {
  if (!days || days.length === 0) {
    return 'no days selected'
  }
  // Sort days and get names
  const sortedDays = [...days].sort((a, b) => a - b)
  const names = sortedDays.map(d => dayNames[d])
  if (names.length === 1) {
    return names[0]
  } else if (names.length === 2) {
    return `${names[0]} and ${names[1]}`
  } else {
    return names.slice(0, -1).join(', ') + ', and ' + names[names.length - 1]
  }
}

// Computed summary text - using a computed ensures reactivity
// We create primitive dependencies to guarantee Vue tracks the changes
const summaryTextValue = computed(() => {
  // Access array as primitive string to ensure reactivity on content changes
  const daysKey = JSON.stringify(daysOfWeek.value)
  const days = JSON.parse(daysKey)
  const intervalVal = interval.value
  const startTimeVal = startTime.value
  const endTypeVal = endType.value
  const endCountVal = endCount.value
  const endDateVal = endDate.value
  const effectiveStartDateVal = effectiveStartDate.value

  let text = `Every ${intervalVal === 1 ? '' : intervalVal + ' '}week${intervalVal > 1 ? 's' : ''} on ${formatSelectedDays(days)} at ${formatTime(startTimeVal)}`

  if (effectiveStartDateVal) {
    text += `, starting ${formatDateDisplay(effectiveStartDateVal)}`
  }

  if (endTypeVal === 'count') {
    text += `, for ${endCountVal} occurrence${endCountVal > 1 ? 's' : ''}`
  } else if (endTypeVal === 'date' && endDateVal) {
    text += `, until ${formatDateDisplay(endDateVal)}`
  }

  return text
})

// Keep computed for backwards compatibility
const selectedDaysText = computed(() => formatSelectedDays(daysOfWeek.value))

function formatTime(time) {
  if (!time) return ''
  const [hours, minutes] = time.split(':')
  const h = parseInt(hours)
  const ampm = h >= 12 ? 'PM' : 'AM'
  const h12 = h % 12 || 12
  return `${h12}:${minutes} ${ampm}`
}

function formatDateDisplay(dateStr) {
  if (!dateStr) return ''
  const date = new Date(dateStr + 'T00:00:00')
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

// Watch for changes and emit
function emitUpdate() {
  // Skip emit when syncing from props to prevent circular loops
  if (isSyncingFromProps) return

  emit('update:modelValue', {
    days_of_week: [...daysOfWeek.value], // Array of selected days
    day_of_week: daysOfWeek.value[0] ?? 1, // Legacy support: first day
    start_time: startTime.value,
    recurrence_interval: interval.value,
    recurrence_end_type: endType.value,
    recurrence_end_count: endType.value === 'count' ? endCount.value : null,
    recurrence_end_date: endType.value === 'date' ? endDate.value : null,
    effective_start_date: effectiveStartDate.value || null
  })
}

watch([daysOfWeek, startTime, interval, endType, endCount, endDate, effectiveStartDate], emitUpdate, { deep: true })

// Helper to check if arrays are equal
function arraysEqual(a, b) {
  if (!Array.isArray(a) || !Array.isArray(b)) return false
  if (a.length !== b.length) return false
  const sortedA = [...a].sort()
  const sortedB = [...b].sort()
  return sortedA.every((val, i) => val === sortedB[i])
}

// Watch for prop changes
watch(() => props.modelValue, (newVal) => {
  // Set flag to prevent emitting back to parent while syncing
  isSyncingFromProps = true

  // Support both new multi-day and legacy single day
  const newDays = newVal.days_of_week?.length > 0
    ? newVal.days_of_week
    : (newVal.day_of_week !== undefined ? [newVal.day_of_week] : [1])

  // Only update if values actually changed to prevent circular updates
  if (!arraysEqual(daysOfWeek.value, newDays)) {
    daysOfWeek.value = [...newDays]
  }
  if (startTime.value !== (newVal.start_time ?? '09:00')) {
    startTime.value = newVal.start_time ?? '09:00'
  }
  if (interval.value !== (newVal.recurrence_interval ?? 1)) {
    interval.value = newVal.recurrence_interval ?? 1
  }
  if (endType.value !== (newVal.recurrence_end_type ?? 'none')) {
    endType.value = newVal.recurrence_end_type ?? 'none'
  }
  if (endCount.value !== (newVal.recurrence_end_count ?? 10)) {
    endCount.value = newVal.recurrence_end_count ?? 10
  }
  if (endDate.value !== (newVal.recurrence_end_date ?? '')) {
    endDate.value = newVal.recurrence_end_date ?? ''
  }
  if (effectiveStartDate.value !== (newVal.effective_start_date ?? '')) {
    effectiveStartDate.value = newVal.effective_start_date ?? ''
  }

  // Reset flag after all sync effects complete
  nextTick(() => {
    isSyncingFromProps = false
  })
}, { deep: true })
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
.gap-3 {
  gap: 12px;
}
.recurrence-editor {
  background-color: #f8f9fa;
  border-radius: 12px;
  padding: 16px;
}
.section-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #546e7a;
  display: flex;
  align-items: center;
}
</style>
