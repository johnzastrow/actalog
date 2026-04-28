<template>
  <v-card elevation="0" rounded="lg" class="pa-3" bg-color="surface">
    <!-- Calendar Header -->
    <div class="d-flex align-center justify-space-between mb-3">
      <v-btn
        icon
        size="small"
        variant="text"
        @click="previousMonth"
      >
        <v-icon>mdi-chevron-left</v-icon>
      </v-btn>
      <div class="text-body-1 font-weight-medium">{{ currentMonthYear }}</div>
      <v-btn
        icon
        size="small"
        variant="text"
        :disabled="isCurrentMonth"
        @click="nextMonth"
      >
        <v-icon>mdi-chevron-right</v-icon>
      </v-btn>
    </div>

    <!-- Day Labels -->
    <div class="d-flex mb-2">
      <div
        v-for="day in dayLabels"
        :key="day"
        class="flex-grow-1 text-center text-caption font-weight-bold text-medium-emphasis"
      >
        {{ day }}
      </div>
    </div>

    <!-- Calendar Days -->
    <div
      v-for="(week, weekIndex) in calendarWeeks"
      :key="weekIndex"
      class="d-flex mb-2"
    >
      <div
        v-for="(day, dayIndex) in week"
        :key="dayIndex"
        class="flex-grow-1 text-center"
      >
        <v-btn
          v-if="day"
          icon
          size="small"
          variant="flat"
          :color="getDayColor(day)"
          :class="{ 'elevation-2': isToday(day) }"
          @click="selectDay(day)"
        >
          <span class="text-caption" :style="{ color: getDayTextColor(day) }">
            {{ day.date }}
          </span>
        </v-btn>
      </div>
    </div>
  </v-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useTheme } from 'vuetify'

const theme = useTheme()

// Get primary color from theme
const primaryColor = computed(() => {
  return theme.current.value.colors.primary || '#00bcd4'
})

const props = defineProps({
  workoutDates: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['daySelected', 'monthChanged'])

onMounted(() => {
  emit('monthChanged', new Date(currentDate.value))
})

const dayLabels = ['M', 'T', 'W', 'T', 'F', 'S', 'S']

const currentDate = ref(new Date())
const today = new Date()

const currentMonthYear = computed(() => {
  return currentDate.value.toLocaleDateString('en-US', {
    month: 'long',
    year: 'numeric'
  })
})

const isCurrentMonth = computed(() => {
  return (
    currentDate.value.getMonth() === today.getMonth() &&
    currentDate.value.getFullYear() === today.getFullYear()
  )
})

const calendarWeeks = computed(() => {
  const year = currentDate.value.getFullYear()
  const month = currentDate.value.getMonth()

  const firstDay = new Date(year, month, 1)
  const lastDay = new Date(year, month + 1, 0)

  // Adjust for Monday start (0 = Sunday, we want 0 = Monday)
  let firstDayOfWeek = firstDay.getDay() - 1
  if (firstDayOfWeek === -1) firstDayOfWeek = 6

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
      fullDate: dayDate,
      hasWorkout: hasWorkoutOnDate(dayDate)
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
})

function hasWorkoutOnDate(date) {
  // Create YYYY-MM-DD string from the calendar date
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const dateStr = `${year}-${month}-${day}`

  return props.workoutDates.some(workoutDate => {
    // Extract YYYY-MM-DD from workout date string
    const workoutDateStr = workoutDate.split('T')[0]
    return workoutDateStr === dateStr
  })
}

function isToday(day) {
  if (!day) return false
  const d = day.fullDate
  return (
    d.getDate() === today.getDate() &&
    d.getMonth() === today.getMonth() &&
    d.getFullYear() === today.getFullYear()
  )
}

function getDayColor(day) {
  if (!day) return 'transparent'
  if (day.hasWorkout) return primaryColor.value
  return 'transparent'  // No background for non-workout days - better contrast on all themes
}

function getDayTextColor(day) {
  if (!day) return 'rgb(var(--v-theme-on-surface), 0.4)'
  if (day.hasWorkout) return '#ffffff'
  return 'rgb(var(--v-theme-on-surface), 0.5)'  // Theme-aware text color
}

function previousMonth() {
  const newDate = new Date(currentDate.value)
  newDate.setDate(1)
  newDate.setMonth(newDate.getMonth() - 1)
  currentDate.value = newDate
  emit('monthChanged', new Date(newDate))
}

function nextMonth() {
  if (!isCurrentMonth.value) {
    const newDate = new Date(currentDate.value)
    newDate.setDate(1)
    newDate.setMonth(newDate.getMonth() + 1)
    currentDate.value = newDate
    emit('monthChanged', new Date(newDate))
  }
}

function selectDay(day) {
  if (day) {
    emit('daySelected', day.fullDate)
  }
}
</script>

<style scoped>
.v-btn {
  min-width: 32px !important;
  width: 32px;
  height: 32px;
}
</style>
