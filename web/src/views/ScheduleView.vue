<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <div class="d-flex align-center mb-4">
        <v-btn icon class="mr-2" @click="$router.back()">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <div>
          <h1 class="text-h5">Class Schedule</h1>
          <div class="text-body-2 text-medium-emphasis">Book your classes</div>
        </div>
      </div>

      <!-- Organization Selector -->
      <v-select
        v-model="selectedOrgId"
        :items="organizations"
        item-title="name"
        item-value="id"
        label="Select Gym"
        variant="outlined"
        density="comfortable"
        class="mb-4"
        @update:model-value="fetchSessions"
      />

      <!-- Date Navigation -->
      <v-card elevation="0" rounded="lg" class="mb-4 pa-3">
        <div class="d-flex align-center justify-space-between">
          <v-btn icon variant="text" @click="previousWeek">
            <v-icon>mdi-chevron-left</v-icon>
          </v-btn>
          <div class="text-center">
            <div class="text-subtitle-1 font-weight-bold">
              {{ formatDateRange(startDate, endDate) }}
            </div>
          </div>
          <v-btn icon variant="text" @click="nextWeek">
            <v-icon>mdi-chevron-right</v-icon>
          </v-btn>
        </div>
      </v-card>

      <!-- Loading State -->
      <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />

      <!-- Error Alert -->
      <v-alert v-if="error" type="error" variant="tonal" closable class="mb-4" @click:close="error = null">
        {{ error }}
      </v-alert>

      <!-- Success Alert -->
      <v-alert v-if="successMessage" type="success" variant="tonal" closable class="mb-4" @click:close="successMessage = null">
        {{ successMessage }}
      </v-alert>

      <!-- Sessions by Day -->
      <div v-if="!loading && selectedOrgId">
        <div v-for="day in groupedSessions" :key="day.date" class="mb-4">
          <div class="text-subtitle-2 font-weight-bold mb-2 px-2">
            {{ formatDayHeader(day.date) }}
          </div>

          <v-card v-for="session in day.sessions" :key="session.id"
                  elevation="0" rounded="lg" class="mb-2" :class="getSessionCardClass(session)">
            <v-card-text class="py-3">
              <div class="d-flex align-center">
                <div class="flex-grow-1">
                  <div class="d-flex align-center">
                    <span class="text-subtitle-1 font-weight-bold">{{ session.name }}</span>
                    <v-chip v-if="session.status === 'cancelled'" color="error" size="x-small" class="ml-2">
                      Cancelled
                    </v-chip>
                  </div>
                  <div class="text-body-2 text-medium-emphasis">
                    {{ formatTime(session.start_time) }} - {{ formatTime(session.end_time) }}
                  </div>
                  <div v-if="session.location_name" class="text-caption text-medium-emphasis">
                    <v-icon size="small" class="mr-1">mdi-map-marker</v-icon>
                    {{ session.location_name }}
                  </div>
                  <div class="text-caption mt-1">
                    <v-icon size="small" class="mr-1">mdi-account-multiple</v-icon>
                    {{ session.reservation_count }}/{{ session.capacity }} spots
                    <span v-if="session.available_spots > 0" class="text-success">
                      ({{ session.available_spots }} available)
                    </span>
                    <span v-else class="text-error">(Full)</span>
                  </div>
                </div>

                <div>
                  <!-- User's reservation status -->
                  <v-chip v-if="session.current_user_status === 'reserved'" color="primary" size="small" class="mr-2">
                    Reserved
                  </v-chip>
                  <v-chip v-else-if="session.current_user_status === 'checked_in'" color="success" size="small" class="mr-2">
                    Checked In
                  </v-chip>

                  <!-- Action buttons -->
                  <v-btn v-if="!session.current_user_status && session.status === 'scheduled' && session.available_spots > 0"
                         color="primary" size="small" @click="makeReservation(session)">
                    Reserve
                  </v-btn>
                  <v-btn v-else-if="session.current_user_status === 'reserved'"
                         color="error" variant="outlined" size="small" @click="cancelReservation(session)">
                    Cancel
                  </v-btn>
                </div>
              </div>
            </v-card-text>
          </v-card>

          <v-alert v-if="day.sessions.length === 0" type="info" variant="tonal" density="compact">
            No classes scheduled
          </v-alert>
        </div>
      </div>

      <!-- No Organization Selected -->
      <v-alert v-if="!selectedOrgId && !loading" type="info" variant="tonal">
        Please select a gym to view the class schedule.
      </v-alert>
    </v-container>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from '@/utils/axios'

const loading = ref(false)
const error = ref(null)
const successMessage = ref(null)
const organizations = ref([])
const selectedOrgId = ref(null)
const sessions = ref([])
const startDate = ref(new Date())
const endDate = ref(new Date())

// Initialize dates
onMounted(() => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  startDate.value = today

  const end = new Date(today)
  end.setDate(end.getDate() + 6)
  endDate.value = end

  fetchOrganizations()
})

async function fetchOrganizations() {
  try {
    const response = await axios.get('/api/admin/organizations')
    organizations.value = response.data.organizations || []

    // Auto-select first organization if available
    if (organizations.value.length > 0 && !selectedOrgId.value) {
      selectedOrgId.value = organizations.value[0].id
      fetchSessions()
    }
  } catch (err) {
    console.error('Failed to fetch organizations:', err)
  }
}

async function fetchSessions() {
  if (!selectedOrgId.value) return

  loading.value = true
  error.value = null

  try {
    const start = formatDateParam(startDate.value)
    const end = formatDateParam(endDate.value)
    const response = await axios.get(`/api/gyms/${selectedOrgId.value}/sessions?start_date=${start}&end_date=${end}`)
    sessions.value = response.data.sessions || []
  } catch (err) {
    console.error('Failed to fetch sessions:', err)
    error.value = err.response?.data?.error || 'Failed to fetch sessions'
  } finally {
    loading.value = false
  }
}

async function makeReservation(session) {
  loading.value = true
  error.value = null

  try {
    await axios.post(`/api/sessions/${session.id}/reserve`)
    successMessage.value = `Successfully reserved a spot in ${session.name}`
    await fetchSessions()
  } catch (err) {
    console.error('Failed to make reservation:', err)
    error.value = err.response?.data?.error || 'Failed to make reservation'
  } finally {
    loading.value = false
  }
}

async function cancelReservation(session) {
  loading.value = true
  error.value = null

  try {
    await axios.delete(`/api/sessions/${session.id}/reserve`)
    successMessage.value = `Reservation cancelled for ${session.name}`
    await fetchSessions()
  } catch (err) {
    console.error('Failed to cancel reservation:', err)
    error.value = err.response?.data?.error || 'Failed to cancel reservation'
  } finally {
    loading.value = false
  }
}

function previousWeek() {
  const newStart = new Date(startDate.value)
  newStart.setDate(newStart.getDate() - 7)
  startDate.value = newStart

  const newEnd = new Date(endDate.value)
  newEnd.setDate(newEnd.getDate() - 7)
  endDate.value = newEnd

  fetchSessions()
}

function nextWeek() {
  const newStart = new Date(startDate.value)
  newStart.setDate(newStart.getDate() + 7)
  startDate.value = newStart

  const newEnd = new Date(endDate.value)
  newEnd.setDate(newEnd.getDate() + 7)
  endDate.value = newEnd

  fetchSessions()
}

const groupedSessions = computed(() => {
  const groups = {}
  const days = []

  // Create day entries for the date range
  const current = new Date(startDate.value)
  while (current <= endDate.value) {
    const dateStr = formatDateParam(current)
    groups[dateStr] = []
    days.push(dateStr)
    current.setDate(current.getDate() + 1)
  }

  // Group sessions by date
  sessions.value.forEach(session => {
    const sessionDate = new Date(session.start_time)
    const dateStr = formatDateParam(sessionDate)
    if (groups[dateStr]) {
      groups[dateStr].push(session)
    }
  })

  return days.map(date => ({
    date,
    sessions: groups[date] || []
  }))
})

function formatDateParam(date) {
  return date.toISOString().split('T')[0]
}

function formatDateRange(start, end) {
  const startStr = start.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  const endStr = end.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  return `${startStr} - ${endStr}`
}

function formatDayHeader(dateStr) {
  const date = new Date(dateStr + 'T00:00:00')
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  const tomorrow = new Date(today)
  tomorrow.setDate(tomorrow.getDate() + 1)

  if (date.toDateString() === today.toDateString()) {
    return 'Today'
  } else if (date.toDateString() === tomorrow.toDateString()) {
    return 'Tomorrow'
  }

  return date.toLocaleDateString('en-US', { weekday: 'long', month: 'short', day: 'numeric' })
}

function formatTime(dateTimeStr) {
  const date = new Date(dateTimeStr)
  return date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
}

function getSessionCardClass(session) {
  if (session.status === 'cancelled') {
    return 'bg-grey-lighten-3'
  }
  if (session.current_user_status === 'reserved' || session.current_user_status === 'checked_in') {
    return 'border-primary'
  }
  return ''
}
</script>

<style scoped>
.mobile-view-wrapper {
  min-height: 100vh;
  padding-bottom: 70px;
}

.border-primary {
  border-left: 4px solid rgb(var(--v-theme-primary));
}
</style>
