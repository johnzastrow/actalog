<template>
  <v-card variant="outlined" class="pa-4">
    <div class="text-subtitle-2 mb-3">Recurrence Pattern</div>

    <!-- Frequency -->
    <div class="d-flex align-center gap-3 mb-4">
      <span class="text-body-2">Every</span>
      <v-text-field
        v-model.number="interval"
        type="number"
        min="1"
        max="12"
        density="compact"
        variant="outlined"
        hide-details
        style="max-width: 80px"
      />
      <span class="text-body-2">week(s)</span>
      <v-chip
        :color="interval === 1 ? 'primary' : 'secondary'"
        size="small"
        variant="tonal"
      >
        {{ interval === 1 ? 'Weekly' : interval === 2 ? 'Bi-weekly' : `Every ${interval} weeks` }}
      </v-chip>
    </div>

    <!-- Day of Week -->
    <div class="mb-4">
      <DayOfWeekSelector
        v-model="dayOfWeek"
        label="On"
        :multiple="false"
      />
    </div>

    <!-- Time -->
    <v-text-field
      v-model="startTime"
      label="Start Time"
      type="time"
      density="compact"
      variant="outlined"
      class="mb-4"
      hide-details
    />

    <!-- Effective Start Date -->
    <v-text-field
      v-model="effectiveStartDate"
      label="Start Date (optional)"
      type="date"
      density="compact"
      variant="outlined"
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
              variant="outlined"
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
              variant="outlined"
              hide-details
              :disabled="endType !== 'date'"
              style="max-width: 160px"
            />
          </div>
        </template>
      </v-radio>
    </v-radio-group>

    <!-- Summary Text -->
    <v-alert type="info" variant="tonal" density="compact" class="mt-4">
      <div class="text-body-2">{{ summaryText }}</div>
    </v-alert>
  </v-card>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import DayOfWeekSelector from './DayOfWeekSelector.vue'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({
      day_of_week: 0,
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

// Local state
const dayOfWeek = ref(props.modelValue.day_of_week ?? 0)
const startTime = ref(props.modelValue.start_time ?? '09:00')
const interval = ref(props.modelValue.recurrence_interval ?? 1)
const endType = ref(props.modelValue.recurrence_end_type ?? 'none')
const endCount = ref(props.modelValue.recurrence_end_count ?? 10)
const endDate = ref(props.modelValue.recurrence_end_date ?? '')
const effectiveStartDate = ref(props.modelValue.effective_start_date ?? '')

// Summary text
const summaryText = computed(() => {
  let text = `Every ${interval.value === 1 ? '' : interval.value + ' '}week${interval.value > 1 ? 's' : ''} on ${dayNames[dayOfWeek.value]} at ${formatTime(startTime.value)}`

  if (effectiveStartDate.value) {
    text += `, starting ${formatDateDisplay(effectiveStartDate.value)}`
  }

  if (endType.value === 'count') {
    text += `, for ${endCount.value} occurrence${endCount.value > 1 ? 's' : ''}`
  } else if (endType.value === 'date' && endDate.value) {
    text += `, until ${formatDateDisplay(endDate.value)}`
  }

  return text
})

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
  emit('update:modelValue', {
    day_of_week: dayOfWeek.value,
    start_time: startTime.value,
    recurrence_interval: interval.value,
    recurrence_end_type: endType.value,
    recurrence_end_count: endType.value === 'count' ? endCount.value : null,
    recurrence_end_date: endType.value === 'date' ? endDate.value : null,
    effective_start_date: effectiveStartDate.value || null
  })
}

watch([dayOfWeek, startTime, interval, endType, endCount, endDate, effectiveStartDate], emitUpdate)

// Watch for prop changes
watch(() => props.modelValue, (newVal) => {
  dayOfWeek.value = newVal.day_of_week ?? 0
  startTime.value = newVal.start_time ?? '09:00'
  interval.value = newVal.recurrence_interval ?? 1
  endType.value = newVal.recurrence_end_type ?? 'none'
  endCount.value = newVal.recurrence_end_count ?? 10
  endDate.value = newVal.recurrence_end_date ?? ''
  effectiveStartDate.value = newVal.effective_start_date ?? ''
}, { deep: true })
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
.gap-3 {
  gap: 12px;
}
</style>
