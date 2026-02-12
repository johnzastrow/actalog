<template>
  <v-container class="pa-4">
    <!-- Page Header -->
    <div class="d-flex align-center mb-4">
      <v-icon icon="mdi-whistle" class="mr-2" color="primary" size="28"></v-icon>
      <div>
        <h1 class="text-h5 font-weight-bold">Coach Dashboard</h1>
        <p class="text-body-2 text-medium-emphasis mb-0">
          Manage your assigned class sessions
        </p>
      </div>
    </div>

    <!-- Alerts -->
    <v-alert
      v-if="error"
      type="error"
      closable
      class="mb-4"
      @click:close="error = null"
    >
      {{ error }}
    </v-alert>

    <v-alert
      v-if="successMessage"
      type="success"
      closable
      class="mb-4"
      @click:close="successMessage = null"
    >
      {{ successMessage }}
    </v-alert>

    <!-- Loading State -->
    <div v-if="loading" class="d-flex justify-center py-8">
      <v-progress-circular indeterminate color="primary"></v-progress-circular>
    </div>

    <!-- No Sessions -->
    <v-card v-else-if="sessions.length === 0" class="text-center pa-8">
      <v-icon icon="mdi-calendar-blank" size="64" color="grey-lighten-1" class="mb-4"></v-icon>
      <h2 class="text-h6 mb-2">No Upcoming Sessions</h2>
      <p class="text-body-2 text-medium-emphasis">
        You don't have any sessions assigned to you yet.
      </p>
    </v-card>

    <!-- Sessions List -->
    <div v-else>
      <h2 class="text-h6 mb-3">My Sessions ({{ sessions.length }})</h2>

      <v-card
        v-for="session in sessions"
        :key="session.id"
        class="mb-3"
        :class="getSessionCardClass(session)"
      >
        <v-card-text>
          <div class="d-flex justify-space-between align-start">
            <div>
              <h3 class="text-subtitle-1 font-weight-bold">{{ session.name }}</h3>
              <div class="text-body-2 text-medium-emphasis">
                <v-icon icon="mdi-clock-outline" size="14" class="mr-1"></v-icon>
                {{ formatDateTime(session.start_time) }} - {{ formatTime(session.end_time) }}
              </div>
              <div v-if="session.location_name" class="text-body-2 text-medium-emphasis">
                <v-icon icon="mdi-map-marker" size="14" class="mr-1"></v-icon>
                {{ session.location_name }}
              </div>
              <div class="text-body-2 text-medium-emphasis">
                <v-icon icon="mdi-account-group" size="14" class="mr-1"></v-icon>
                {{ session.reservation_count || 0 }} / {{ session.capacity }} enrolled
              </div>
            </div>
            <div class="text-right">
              <v-chip
                :color="getStatusColor(session.status)"
                size="small"
                class="mb-2"
              >
                {{ formatStatus(session.status) }}
              </v-chip>
              <div class="d-flex flex-column ga-1">
                <v-btn
                  v-if="session.status === 'scheduled' || session.status === 'in_progress'"
                  color="primary"
                  variant="outlined"
                  size="small"
                  @click="openRoster(session)"
                >
                  <v-icon icon="mdi-clipboard-list" size="16" class="mr-1"></v-icon>
                  View Roster
                </v-btn>
                <v-btn
                  v-if="session.status === 'in_progress'"
                  color="success"
                  variant="outlined"
                  size="small"
                  @click="confirmCompleteSession(session)"
                >
                  <v-icon icon="mdi-check-circle" size="16" class="mr-1"></v-icon>
                  Complete
                </v-btn>
              </div>
            </div>
          </div>
        </v-card-text>
      </v-card>
    </div>

    <!-- Roster Dialog -->
    <v-dialog v-model="rosterDialog" max-width="600">
      <v-card>
        <v-card-title class="d-flex align-center">
          <v-icon icon="mdi-clipboard-list" class="mr-2"></v-icon>
          Session Roster
          <v-spacer></v-spacer>
          <v-btn icon="mdi-close" variant="text" @click="rosterDialog = false"></v-btn>
        </v-card-title>
        <v-card-subtitle v-if="currentSession">
          {{ currentSession.name }} - {{ formatDateTime(currentSession.start_time) }}
        </v-card-subtitle>
        <v-divider></v-divider>
        <v-card-text>
          <div v-if="rosterLoading" class="d-flex justify-center py-4">
            <v-progress-circular indeterminate color="primary"></v-progress-circular>
          </div>
          <div v-else-if="roster.length === 0" class="text-center py-4 text-medium-emphasis">
            No reservations for this session
          </div>
          <v-list v-else>
            <v-list-item
              v-for="res in roster"
              :key="res.id"
              :title="res.user_name || res.user_email"
              :subtitle="res.user_email"
            >
              <template v-slot:prepend>
                <v-avatar :color="getReservationAvatarColor(res.status)" size="40">
                  <span class="text-body-2 font-weight-medium">
                    {{ getInitials(res.user_name || res.user_email) }}
                  </span>
                </v-avatar>
              </template>
              <template v-slot:append>
                <div class="d-flex align-center ga-2">
                  <v-chip
                    :color="getReservationStatusColor(res.status)"
                    size="small"
                  >
                    {{ formatReservationStatus(res.status) }}
                  </v-chip>
                  <v-btn
                    v-if="res.status === 'reserved'"
                    color="success"
                    variant="tonal"
                    size="small"
                    :loading="checkInLoading === res.id"
                    @click="checkInReservation(res)"
                  >
                    Check In
                  </v-btn>
                  <v-btn
                    v-if="res.status === 'reserved' || res.status === 'checked_in'"
                    color="error"
                    variant="tonal"
                    size="small"
                    :loading="noShowLoading === res.id"
                    @click="markNoShow(res)"
                  >
                    No Show
                  </v-btn>
                </div>
              </template>
            </v-list-item>
          </v-list>
        </v-card-text>
        <v-divider></v-divider>
        <v-card-actions>
          <v-btn
            v-if="currentSession && (currentSession.status === 'scheduled' || currentSession.status === 'in_progress')"
            color="success"
            variant="flat"
            @click="confirmCompleteSession(currentSession)"
          >
            <v-icon icon="mdi-check-circle" class="mr-1"></v-icon>
            Complete Session
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn variant="text" @click="rosterDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Complete Session Confirmation Dialog -->
    <v-dialog v-model="completeDialog" max-width="400">
      <v-card>
        <v-card-title>Complete Session?</v-card-title>
        <v-card-text>
          Are you sure you want to mark this session as completed? This will finalize attendance records.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" @click="completeDialog = false">Cancel</v-btn>
          <v-btn
            color="success"
            variant="flat"
            :loading="completeLoading"
            @click="completeSession"
          >
            Complete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'

// State
const loading = ref(false)
const error = ref(null)
const successMessage = ref(null)
const sessions = ref([])

// Roster dialog
const rosterDialog = ref(false)
const rosterLoading = ref(false)
const roster = ref([])
const currentSession = ref(null)
const checkInLoading = ref(null)
const noShowLoading = ref(null)

// Complete dialog
const completeDialog = ref(false)
const completeLoading = ref(false)
const sessionToComplete = ref(null)

// Load coach sessions
const loadSessions = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await axios.get('/api/coaches/me/sessions')
    sessions.value = response.data.sessions || []
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to load sessions'
  } finally {
    loading.value = false
  }
}

// Open roster dialog
const openRoster = async (session) => {
  currentSession.value = session
  rosterDialog.value = true
  rosterLoading.value = true
  roster.value = []

  try {
    const response = await axios.get(`/api/coaches/sessions/${session.id}/roster`)
    roster.value = response.data.roster || []
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to load roster'
    rosterDialog.value = false
  } finally {
    rosterLoading.value = false
  }
}

// Check in reservation
const checkInReservation = async (res) => {
  checkInLoading.value = res.id
  try {
    await axios.post(`/api/coaches/sessions/${currentSession.value.id}/check-in/${res.id}`)
    res.status = 'checked_in'
    successMessage.value = `Checked in ${res.user_name || res.user_email}`
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to check in'
  } finally {
    checkInLoading.value = null
  }
}

// Mark no show
const markNoShow = async (res) => {
  noShowLoading.value = res.id
  try {
    await axios.post(`/api/coaches/sessions/${currentSession.value.id}/no-show/${res.id}`)
    res.status = 'no_show'
    successMessage.value = `Marked ${res.user_name || res.user_email} as no-show`
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to mark no-show'
  } finally {
    noShowLoading.value = null
  }
}

// Confirm complete session
const confirmCompleteSession = (session) => {
  sessionToComplete.value = session
  completeDialog.value = true
}

// Complete session
const completeSession = async () => {
  if (!sessionToComplete.value) return

  completeLoading.value = true
  try {
    await axios.post(`/api/coaches/sessions/${sessionToComplete.value.id}/complete`)
    sessionToComplete.value.status = 'completed'
    successMessage.value = `Session "${sessionToComplete.value.name}" completed`
    completeDialog.value = false
    rosterDialog.value = false
    await loadSessions()
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to complete session'
  } finally {
    completeLoading.value = false
    sessionToComplete.value = null
  }
}

// Helper functions
const formatDateTime = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  })
}

const formatTime = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit'
  })
}

const getStatusColor = (status) => {
  const colors = {
    scheduled: 'primary',
    in_progress: 'warning',
    completed: 'success',
    cancelled: 'grey'
  }
  return colors[status] || 'grey'
}

const formatStatus = (status) => {
  const labels = {
    scheduled: 'Scheduled',
    in_progress: 'In Progress',
    completed: 'Completed',
    cancelled: 'Cancelled'
  }
  return labels[status] || status
}

const getSessionCardClass = (session) => {
  if (session.status === 'cancelled') return 'bg-grey-lighten-3'
  if (session.status === 'in_progress') return 'border-warning'
  return ''
}

const getReservationStatusColor = (status) => {
  const colors = {
    reserved: 'primary',
    checked_in: 'success',
    attended: 'success',
    cancelled: 'grey',
    no_show: 'error'
  }
  return colors[status] || 'grey'
}

const getReservationAvatarColor = (status) => {
  const colors = {
    reserved: 'blue-lighten-4',
    checked_in: 'green-lighten-4',
    attended: 'green-lighten-4',
    cancelled: 'grey-lighten-3',
    no_show: 'red-lighten-4'
  }
  return colors[status] || 'grey-lighten-3'
}

const formatReservationStatus = (status) => {
  const labels = {
    reserved: 'Reserved',
    checked_in: 'Checked In',
    attended: 'Attended',
    cancelled: 'Cancelled',
    no_show: 'No Show'
  }
  return labels[status] || status
}

const getInitials = (name) => {
  if (!name) return '?'
  const parts = name.split(' ')
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase()
  }
  return name.substring(0, 2).toUpperCase()
}

onMounted(() => {
  loadSessions()
})
</script>
