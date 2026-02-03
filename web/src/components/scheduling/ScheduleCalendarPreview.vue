<template>
  <div class="schedule-calendar-preview">
    <div class="text-subtitle-2 mb-3 d-flex align-center justify-space-between">
      <span>Schedule Preview</span>
      <v-chip size="small" variant="tonal" color="primary">
        {{ scheduledDates.length }} sessions
      </v-chip>
    </div>

    <div class="calendars-container" :class="{ 'stacked': !horizontal }">
      <div
        v-for="(month, index) in months"
        :key="index"
        class="calendar-month"
      >
        <!-- Month Header -->
        <div class="month-header text-body-2 font-weight-medium mb-2 text-center">
          {{ formatMonthYear(month) }}
        </div>

        <!-- Day Labels -->
        <div class="day-labels d-flex mb-1">
          <div
            v-for="day in dayLabels"
            :key="day"
            class="day-label text-caption text-center text-medium-emphasis"
          >
            {{ day }}
          </div>
        </div>

        <!-- Calendar Days -->
        <div
          v-for="(week, weekIndex) in getCalendarWeeks(month)"
          :key="weekIndex"
          class="d-flex"
        >
          <div
            v-for="(day, dayIndex) in week"
            :key="dayIndex"
            class="calendar-day text-center"
            :class="{
              'has-session': day && hasSession(day.fullDate),
              'is-today': day && isToday(day.fullDate),
              'empty': !day
            }"
            @click="day && onDayClick(day.fullDate)"
          >
            <span v-if="day" class="day-number text-caption">{{ day.date }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Legend -->
    <div class="legend d-flex align-center gap-4 mt-3">
      <div class="d-flex align-center gap-1">
        <div class="legend-dot has-session"></div>
        <span class="text-caption text-medium-emphasis">Scheduled</span>
      </div>
      <div class="d-flex align-center gap-1">
        <div class="legend-dot is-today"></div>
        <span class="text-caption text-medium-emphasis">Today</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  scheduledDates: {
    type: Array,
    default: () => []
  },
  templateColor: {
    type: String,
    default: '#00bcd4'
  },
  monthCount: {
    type: Number,
    default: 3
  },
  horizontal: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['dayClick'])

const dayLabels = ['S', 'M', 'T', 'W', 'T', 'F', 'S']
const today = new Date()

// Generate months array
const months = computed(() => {
  const result = []
  const startMonth = new Date(today.getFullYear(), today.getMonth(), 1)

  for (let i = 0; i < props.monthCount; i++) {
    const month = new Date(startMonth)
    month.setMonth(month.getMonth() + i)
    result.push(month)
  }

  return result
})

// Convert scheduled dates to Date objects for comparison
const scheduledDateSet = computed(() => {
  const set = new Set()
  props.scheduledDates.forEach(dateStr => {
    // Handle both RFC3339 and date-only formats
    const date = new Date(dateStr)
    const key = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`
    set.add(key)
  })
  return set
})

function formatMonthYear(date) {
  return date.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
}

function getCalendarWeeks(monthDate) {
  const year = monthDate.getFullYear()
  const month = monthDate.getMonth()

  const firstDay = new Date(year, month, 1)
  const lastDay = new Date(year, month + 1, 0)

  // Start from Sunday (0)
  const firstDayOfWeek = firstDay.getDay()

  const weeks = []
  let currentWeek = []

  // Fill in empty days at the start
  for (let i = 0; i < firstDayOfWeek; i++) {
    currentWeek.push(null)
  }

  // Fill in the days of the month
  for (let date = 1; date <= lastDay.getDate(); date++) {
    const dayDate = new Date(year, month, date)
    currentWeek.push({
      date,
      fullDate: dayDate
    })

    if (currentWeek.length === 7) {
      weeks.push(currentWeek)
      currentWeek = []
    }
  }

  // Fill in remaining days
  if (currentWeek.length > 0) {
    while (currentWeek.length < 7) {
      currentWeek.push(null)
    }
    weeks.push(currentWeek)
  }

  return weeks
}

function hasSession(date) {
  if (!date) return false
  const key = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`
  return scheduledDateSet.value.has(key)
}

function isToday(date) {
  if (!date) return false
  return (
    date.getDate() === today.getDate() &&
    date.getMonth() === today.getMonth() &&
    date.getFullYear() === today.getFullYear()
  )
}

function onDayClick(date) {
  emit('dayClick', date)
}
</script>

<style scoped>
.schedule-calendar-preview {
  padding: 8px;
}

.calendars-container {
  display: flex;
  gap: 16px;
  overflow-x: auto;
}

.calendars-container.stacked {
  flex-direction: column;
}

.calendar-month {
  min-width: 200px;
  flex: 1;
}

.day-labels {
  border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
  padding-bottom: 4px;
}

.day-label {
  flex: 1;
  font-size: 10px;
}

.calendar-day {
  flex: 1;
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-radius: 4px;
  margin: 1px;
  transition: background-color 0.2s;
}

.calendar-day:hover:not(.empty) {
  background-color: rgba(var(--v-theme-primary), 0.1);
}

.calendar-day.has-session {
  background-color: v-bind(templateColor);
  color: white;
}

.calendar-day.has-session .day-number {
  color: white;
  font-weight: 600;
}

.calendar-day.is-today:not(.has-session) {
  border: 2px solid rgba(var(--v-theme-primary), 0.5);
}

.calendar-day.empty {
  cursor: default;
}

.day-number {
  font-size: 11px;
}

.legend {
  padding-top: 8px;
  border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}

.legend-dot {
  width: 12px;
  height: 12px;
  border-radius: 2px;
}

.legend-dot.has-session {
  background-color: v-bind(templateColor);
}

.legend-dot.is-today {
  border: 2px solid rgba(var(--v-theme-primary), 0.5);
  background-color: transparent;
}

.gap-1 {
  gap: 4px;
}

.gap-4 {
  gap: 16px;
}
</style>
