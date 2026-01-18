<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <div class="d-flex align-center mb-4">
        <v-btn icon class="mr-2" @click="$router.back()">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <div>
          <h1 class="text-h5">My Reservations</h1>
          <div class="text-body-2 text-medium-emphasis">Upcoming class reservations</div>
        </div>
      </div>

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

      <!-- Upcoming Reservations -->
      <v-card elevation="0" rounded="lg" class="mb-4">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-calendar-check</v-icon>
          Upcoming Classes ({{ reservations.length }})
        </v-card-title>
        <v-divider />

        <div v-if="reservations.length > 0">
          <v-list>
            <v-list-item v-for="res in reservations" :key="res.id" class="py-3">
              <template #prepend>
                <v-avatar :color="getStatusColor(res.status)" class="mr-3">
                  <v-icon color="white">{{ getStatusIcon(res.status) }}</v-icon>
                </v-avatar>
              </template>

              <v-list-item-title class="font-weight-bold">
                {{ res.session_name }}
              </v-list-item-title>

              <v-list-item-subtitle>
                <div class="mt-1">
                  <v-icon size="small" class="mr-1">mdi-calendar</v-icon>
                  {{ formatDateTime(res.session_start_time) }}
                </div>
                <div>
                  <v-icon size="small" class="mr-1">mdi-domain</v-icon>
                  {{ res.organization_name }}
                </div>
                <div v-if="res.location_name">
                  <v-icon size="small" class="mr-1">mdi-map-marker</v-icon>
                  {{ res.location_name }}
                </div>
              </v-list-item-subtitle>

              <template #append>
                <div class="text-right">
                  <v-chip :color="getStatusColor(res.status)" size="small" class="mb-1">
                    {{ formatStatus(res.status) }}
                  </v-chip>
                  <div v-if="res.status === 'reserved'">
                    <v-btn color="error" variant="text" size="small" @click="cancelReservation(res)">
                      Cancel
                    </v-btn>
                  </div>
                </div>
              </template>
            </v-list-item>
          </v-list>
        </div>

        <v-card-text v-else>
          <v-alert type="info" variant="tonal" class="mb-0">
            <div class="d-flex align-center">
              <div class="flex-grow-1">
                You don't have any upcoming class reservations.
              </div>
              <v-btn color="primary" size="small" to="/schedule">
                Browse Classes
              </v-btn>
            </div>
          </v-alert>
        </v-card-text>
      </v-card>
    </v-container>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'

const loading = ref(false)
const error = ref(null)
const successMessage = ref(null)
const reservations = ref([])

async function fetchReservations() {
  loading.value = true
  error.value = null

  try {
    const response = await axios.get('/api/users/me/reservations/upcoming')
    reservations.value = response.data.reservations || []
  } catch (err) {
    console.error('Failed to fetch reservations:', err)
    error.value = err.response?.data?.error || 'Failed to fetch reservations'
  } finally {
    loading.value = false
  }
}

async function cancelReservation(res) {
  loading.value = true
  error.value = null

  try {
    await axios.delete(`/api/sessions/${res.session_id}/reserve`)
    successMessage.value = `Reservation cancelled for ${res.session_name}`
    await fetchReservations()
  } catch (err) {
    console.error('Failed to cancel reservation:', err)
    error.value = err.response?.data?.error || 'Failed to cancel reservation'
  } finally {
    loading.value = false
  }
}

function formatDateTime(dateTimeStr) {
  const date = new Date(dateTimeStr)
  return date.toLocaleString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  })
}

function getStatusColor(status) {
  switch (status) {
    case 'reserved': return 'primary'
    case 'checked_in': return 'success'
    case 'attended': return 'success'
    case 'cancelled': return 'grey'
    case 'no_show': return 'error'
    default: return 'grey'
  }
}

function getStatusIcon(status) {
  switch (status) {
    case 'reserved': return 'mdi-calendar-clock'
    case 'checked_in': return 'mdi-check-circle'
    case 'attended': return 'mdi-check-all'
    case 'cancelled': return 'mdi-cancel'
    case 'no_show': return 'mdi-account-off'
    default: return 'mdi-help-circle'
  }
}

function formatStatus(status) {
  switch (status) {
    case 'reserved': return 'Reserved'
    case 'checked_in': return 'Checked In'
    case 'attended': return 'Attended'
    case 'cancelled': return 'Cancelled'
    case 'no_show': return 'No Show'
    default: return status
  }
}

onMounted(() => {
  fetchReservations()
})
</script>

<style scoped>
.mobile-view-wrapper {
  min-height: 100vh;
  padding-bottom: 70px;
}
</style>
