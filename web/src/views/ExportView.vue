<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" density="compact" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-toolbar-title class="text-white font-weight-bold">Export Data</v-toolbar-title>
      <v-spacer />
      <v-btn icon="mdi-close" color="white" size="small" @click="$router.back()" />
    </v-app-bar>

    <v-container class="px-3 pb-1 pt-0" style="margin-top: 48px; margin-bottom: 70px">
      <!-- Success Alert -->
      <v-alert v-if="successMessage" type="success" closable @click:close="successMessage = null" class="mb-3">
        {{ successMessage }}
      </v-alert>

      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable @click:close="error = null" class="mb-3">
        {{ error }}
      </v-alert>

      <!-- Info Card -->
      <v-card elevation="0" rounded="lg" class="pa-4 mb-3" style="background: white">
        <div class="d-flex align-center mb-2">
          <v-icon color="#00bcd4" size="small" class="mr-2">mdi-information</v-icon>
          <h3 class="text-body-1 font-weight-bold" style="color: #1a1a1a">About Data Export</h3>
        </div>
        <p class="text-body-2 mb-0" style="color: #666">
          Export your data for backup, analysis, or migration. WODs and Movements export as CSV, User Workouts export as JSON with full nested data.
        </p>
      </v-card>

      <!-- WODs Export Section -->
      <v-card elevation="0" rounded="lg" class="pa-4 mb-3" style="background: white">
        <div class="d-flex align-center mb-3">
          <v-icon color="#ffc107" class="mr-2">mdi-fire</v-icon>
          <h2 class="text-h6 font-weight-bold" style="color: #1a1a1a">Export WODs</h2>
        </div>

        <v-checkbox
          v-model="exportOptions.wods.includeStandard"
          label="Include standard WODs (Fran, Murph, etc.)"
          color="#00bcd4"
          density="compact"
          hide-details
        />

        <v-checkbox
          v-model="exportOptions.wods.includeCustom"
          label="Include my custom WODs"
          color="#00bcd4"
          density="compact"
          hide-details
          class="mt-2"
        />

        <v-btn
          block
          size="large"
          color="#ffc107"
          rounded="lg"
          elevation="2"
          :disabled="!exportOptions.wods.includeStandard && !exportOptions.wods.includeCustom"
          :loading="exportingWods"
          class="mt-4 font-weight-bold"
          style="text-transform: none"
          @click="exportWODs"
        >
          <v-icon start>mdi-download</v-icon>
          Export WODs to CSV
        </v-btn>
      </v-card>

      <!-- Movements Export Section -->
      <v-card elevation="0" rounded="lg" class="pa-4 mb-3" style="background: white">
        <div class="d-flex align-center mb-3">
          <v-icon color="#00bcd4" class="mr-2">mdi-dumbbell</v-icon>
          <h2 class="text-h6 font-weight-bold" style="color: #1a1a1a">Export Movements</h2>
        </div>

        <v-checkbox
          v-model="exportOptions.movements.includeStandard"
          label="Include standard movements (Back Squat, Deadlift, etc.)"
          color="#00bcd4"
          density="compact"
          hide-details
        />

        <v-checkbox
          v-model="exportOptions.movements.includeCustom"
          label="Include my custom movements"
          color="#00bcd4"
          density="compact"
          hide-details
          class="mt-2"
        />

        <v-btn
          block
          size="large"
          color="#00bcd4"
          rounded="lg"
          elevation="2"
          :disabled="!exportOptions.movements.includeStandard && !exportOptions.movements.includeCustom"
          :loading="exportingMovements"
          class="mt-4 font-weight-bold"
          style="text-transform: none"
          @click="exportMovements"
        >
          <v-icon start>mdi-download</v-icon>
          Export Movements to CSV
        </v-btn>
      </v-card>

      <!-- User Workouts Export Section -->
      <v-card elevation="0" rounded="lg" class="pa-4 mb-3" style="background: white">
        <div class="d-flex align-center mb-3">
          <v-icon color="#4caf50" class="mr-2">mdi-clipboard-text</v-icon>
          <h2 class="text-h6 font-weight-bold" style="color: #1a1a1a">Export User Workouts</h2>
        </div>

        <p class="text-body-2 mb-3" style="color: #666">
          Export your workout history with full details (movements, WODs, performance data) in JSON format.
          Optionally filter by date range.
        </p>

        <v-text-field
          v-model="exportOptions.userWorkouts.startDate"
          label="Start Date (optional)"
          type="date"
          color="#00bcd4"
          density="compact"
          variant="outlined"
          hide-details
          class="mb-2"
        />

        <v-text-field
          v-model="exportOptions.userWorkouts.endDate"
          label="End Date (optional)"
          type="date"
          color="#00bcd4"
          density="compact"
          variant="outlined"
          hide-details
          class="mb-3"
        />

        <p v-if="dateRangeError" class="text-caption mb-2" style="color: #f44336">
          {{ dateRangeError }}
        </p>

        <v-btn
          block
          size="large"
          color="#4caf50"
          rounded="lg"
          elevation="2"
          :disabled="!!dateRangeError"
          :loading="exportingUserWorkouts"
          class="mt-2 font-weight-bold"
          style="text-transform: none"
          @click="exportUserWorkouts"
        >
          <v-icon start>mdi-download</v-icon>
          Export User Workouts to JSON
        </v-btn>
      </v-card>

      <!-- Export Format Info -->
      <v-card elevation="0" rounded="lg" class="pa-4 mb-3" style="background: white">
        <div class="d-flex align-center mb-2">
          <v-icon color="#00bcd4" size="small" class="mr-2">mdi-information</v-icon>
          <h3 class="text-body-1 font-weight-bold" style="color: #1a1a1a">Export Formats</h3>
        </div>
        <p class="text-body-2 mb-2" style="color: #666">
          <strong>CSV Format:</strong> WODs and Movements export as CSV files that can be opened in Excel, Google Sheets, or any spreadsheet application.
        </p>
        <p class="text-caption mb-2" style="color: #999">
          <strong>WODs CSV:</strong> name, source, type, regime, score_type, description, url, notes, is_standard, created_by_email<br>
          <strong>Movements CSV:</strong> name, type, description, is_standard, created_by_email
        </p>
        <p class="text-body-2 mb-1 mt-2" style="color: #666">
          <strong>JSON Format:</strong> User Workouts export as JSON with full nested data (movements, WODs, performance records).
        </p>
        <p class="text-caption mb-0" style="color: #999">
          <strong>User Workouts JSON:</strong> Complete workout history with metadata, movements, WODs, sets, and performance details.
        </p>
      </v-card>
    </v-container>

    <!-- Bottom Navigation -->
    <v-bottom-navigation fixed elevation="8" height="56" style="z-index: 5">
      <v-btn value="dashboard" to="/">
        <v-icon>mdi-view-dashboard</v-icon>
        <span>Dashboard</span>
      </v-btn>

      <v-btn value="log" to="/log">
        <v-icon>mdi-plus-circle</v-icon>
        <span>Log</span>
      </v-btn>

      <v-btn value="performance" to="/performance">
        <v-icon>mdi-chart-line</v-icon>
        <span>Performance</span>
      </v-btn>

      <v-btn value="profile" to="/profile">
        <v-icon>mdi-account</v-icon>
        <span>Profile</span>
      </v-btn>
    </v-bottom-navigation>
  </v-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import axios from '@/utils/axios'

// State
const exportingWods = ref(false)
const exportingMovements = ref(false)
const exportingUserWorkouts = ref(false)
const successMessage = ref(null)
const error = ref(null)

const exportOptions = ref({
  wods: {
    includeStandard: true,
    includeCustom: true
  },
  movements: {
    includeStandard: true,
    includeCustom: true
  },
  userWorkouts: {
    startDate: '',
    endDate: ''
  }
})

// Computed: Date range validation
const dateRangeError = computed(() => {
  const { startDate, endDate } = exportOptions.value.userWorkouts

  // Both must be provided together or neither
  if ((startDate && !endDate) || (!startDate && endDate)) {
    return 'Both start and end dates must be provided together, or leave both empty for all workouts.'
  }

  // Start date must be before or equal to end date
  if (startDate && endDate && startDate > endDate) {
    return 'Start date must be before or equal to end date.'
  }

  return null
})

// Export WODs
const exportWODs = async () => {
  exportingWods.value = true
  error.value = null
  successMessage.value = null

  try {
    const params = {
      include_standard: exportOptions.value.wods.includeStandard,
      include_custom: exportOptions.value.wods.includeCustom
    }

    const response = await axios.get('/api/export/wods', {
      params,
      responseType: 'blob'
    })

    // Create a download link
    const blob = new Blob([response.data], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `wods_export_${new Date().toISOString().split('T')[0]}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    successMessage.value = 'WODs exported successfully!'
  } catch (err) {
    console.error('Export WODs error:', err)
    error.value = err.response?.data?.error || 'Failed to export WODs. Please try again.'
  } finally {
    exportingWods.value = false
  }
}

// Export Movements
const exportMovements = async () => {
  exportingMovements.value = true
  error.value = null
  successMessage.value = null

  try {
    const params = {
      include_standard: exportOptions.value.movements.includeStandard,
      include_custom: exportOptions.value.movements.includeCustom
    }

    const response = await axios.get('/api/export/movements', {
      params,
      responseType: 'blob'
    })

    // Create a download link
    const blob = new Blob([response.data], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `movements_export_${new Date().toISOString().split('T')[0]}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    successMessage.value = 'Movements exported successfully!'
  } catch (err) {
    console.error('Export Movements error:', err)
    error.value = err.response?.data?.error || 'Failed to export movements. Please try again.'
  } finally {
    exportingMovements.value = false
  }
}

// Export User Workouts
const exportUserWorkouts = async () => {
  exportingUserWorkouts.value = true
  error.value = null
  successMessage.value = null

  try {
    const { startDate, endDate } = exportOptions.value.userWorkouts
    const params = {}

    // Add date range if provided
    if (startDate && endDate) {
      params.start_date = startDate
      params.end_date = endDate
    }

    const response = await axios.get('/api/export/user-workouts', {
      params,
      responseType: 'blob'
    })

    // Create a download link
    const blob = new Blob([response.data], { type: 'application/json' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url

    // Dynamic filename based on date range
    let filename = 'user_workouts_export.json'
    if (startDate && endDate) {
      filename = `user_workouts_${startDate}_to_${endDate}.json`
    } else {
      filename = `user_workouts_export_${new Date().toISOString().split('T')[0]}.json`
    }
    link.download = filename

    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    successMessage.value = 'User workouts exported successfully!'
  } catch (err) {
    console.error('Export User Workouts error:', err)
    error.value = err.response?.data?.error || 'Failed to export user workouts. Please try again.'
  } finally {
    exportingUserWorkouts.value = false
  }
}
</script>
