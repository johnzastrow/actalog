<template>
  <v-container fluid class="pa-4">
    <div class="d-flex align-center mb-4">
      <v-btn icon @click="$router.back()" class="mr-2">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <div>
        <h1 class="text-h5">Audit Logs</h1>
        <div class="text-body-2 text-medium-emphasis">View security events and user activity</div>
      </div>
    </div>

    <!-- Loading State -->
    <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />

    <!-- Error Alert -->
    <v-alert v-if="error" type="error" variant="tonal" closable @click:close="error = null" class="mb-4">
      {{ error }}
    </v-alert>

    <!-- Filters Card -->
    <v-card elevation="0" rounded="lg" class="mb-4">
      <v-card-text>
        <v-row>
          <v-col cols="12" md="4">
            <v-select
              v-model="filterEventType"
              :items="eventTypes"
              label="Event Type"
              clearable
              variant="outlined"
              density="compact"
              hide-details
              @update:model-value="loadLogs"
            />
          </v-col>
          <v-col cols="12" md="4">
            <v-text-field
              v-model="filterUserEmail"
              label="User Email"
              clearable
              variant="outlined"
              density="compact"
              prepend-inner-icon="mdi-magnify"
              hide-details
              @keyup.enter="loadLogs"
            />
          </v-col>
          <v-col cols="12" md="4">
            <v-btn
              color="primary"
              block
              @click="loadLogs"
            >
              <v-icon start>mdi-filter</v-icon>
              Apply Filters
            </v-btn>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <!-- Audit Logs Table -->
    <v-card elevation="0" rounded="lg">
      <v-card-title class="d-flex align-center">
        <v-icon class="mr-2">mdi-history</v-icon>
        Audit Logs ({{ total }})
      </v-card-title>

      <v-divider />

      <v-data-table
        :headers="headers"
        :items="logs"
        :loading="loading"
        :items-per-page="limit"
        hide-default-footer
        class="elevation-0"
      >
        <!-- Timestamp Column -->
        <template #item.created_at="{ item }">
          <div class="text-body-2">
            {{ formatDate(item.created_at) }}
          </div>
        </template>

        <!-- Event Type Column -->
        <template #item.event_type="{ item }">
          <v-chip :color="getEventColor(item.event_type)" size="small">
            {{ formatEventType(item.event_type) }}
          </v-chip>
        </template>

        <!-- User Column -->
        <template #item.user_email="{ item }">
          <div v-if="item.user_email" class="text-body-2">
            {{ item.user_email }}
          </div>
          <span v-else class="text-body-2 text-medium-emphasis">System</span>
        </template>

        <!-- Target Column -->
        <template #item.target_user_email="{ item }">
          <div v-if="item.target_user_email" class="text-body-2">
            {{ item.target_user_email }}
          </div>
          <span v-else class="text-medium-emphasis">-</span>
        </template>

        <!-- IP Address Column -->
        <template #item.ip_address="{ item }">
          <code v-if="item.ip_address" class="text-caption">{{ item.ip_address }}</code>
          <span v-else class="text-medium-emphasis">-</span>
        </template>

        <!-- Details Column -->
        <template #item.details="{ item }">
          <v-tooltip v-if="item.details" location="top">
            <template #activator="{ props }">
              <v-btn
                v-bind="props"
                icon
                size="small"
                variant="text"
                @click="showDetails(item)"
              >
                <v-icon>mdi-information-outline</v-icon>
              </v-btn>
            </template>
            View Details
          </v-tooltip>
          <span v-else class="text-medium-emphasis">-</span>
        </template>
      </v-data-table>

      <!-- Pagination -->
      <v-divider />
      <div class="d-flex align-center justify-space-between pa-4">
        <div class="text-body-2 text-medium-emphasis">
          Showing {{ offset + 1 }} to {{ Math.min(offset + limit, total) }} of {{ total }} logs
        </div>
        <div class="d-flex gap-2">
          <v-btn
            variant="outlined"
            size="small"
            :disabled="offset === 0"
            @click="previousPage"
          >
            <v-icon start>mdi-chevron-left</v-icon>
            Previous
          </v-btn>
          <v-btn
            variant="outlined"
            size="small"
            :disabled="offset + limit >= total"
            @click="nextPage"
          >
            Next
            <v-icon end>mdi-chevron-right</v-icon>
          </v-btn>
        </div>
      </div>
    </v-card>

    <!-- Details Dialog -->
    <v-dialog v-model="detailsDialog" max-width="600">
      <v-card v-if="selectedLog">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-information</v-icon>
          Audit Log Details
        </v-card-title>
        <v-divider />
        <v-card-text>
          <v-list>
            <v-list-item>
              <v-list-item-title class="text-subtitle-2">Event Type</v-list-item-title>
              <v-list-item-subtitle>
                <v-chip :color="getEventColor(selectedLog.event_type)" size="small" class="mt-1">
                  {{ formatEventType(selectedLog.event_type) }}
                </v-chip>
              </v-list-item-subtitle>
            </v-list-item>

            <v-list-item>
              <v-list-item-title class="text-subtitle-2">Timestamp</v-list-item-title>
              <v-list-item-subtitle>{{ formatFullDate(selectedLog.created_at) }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="selectedLog.user_email">
              <v-list-item-title class="text-subtitle-2">User</v-list-item-title>
              <v-list-item-subtitle>{{ selectedLog.user_email }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="selectedLog.target_user_email">
              <v-list-item-title class="text-subtitle-2">Target User</v-list-item-title>
              <v-list-item-subtitle>{{ selectedLog.target_user_email }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="selectedLog.ip_address">
              <v-list-item-title class="text-subtitle-2">IP Address</v-list-item-title>
              <v-list-item-subtitle><code>{{ selectedLog.ip_address }}</code></v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="selectedLog.user_agent">
              <v-list-item-title class="text-subtitle-2">User Agent</v-list-item-title>
              <v-list-item-subtitle class="text-caption">{{ selectedLog.user_agent }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="selectedLog.details">
              <v-list-item-title class="text-subtitle-2">Details</v-list-item-title>
              <v-list-item-subtitle class="white-space-pre-wrap">{{ selectedLog.details }}</v-list-item-subtitle>
            </v-list-item>
          </v-list>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="detailsDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'

const loading = ref(false)
const error = ref(null)
const logs = ref([])
const total = ref(0)
const limit = ref(50)
const offset = ref(0)

const filterEventType = ref(null)
const filterUserEmail = ref('')

const detailsDialog = ref(false)
const selectedLog = ref(null)

const headers = [
  { title: 'Timestamp', value: 'created_at', sortable: false },
  { title: 'Event', value: 'event_type', sortable: false },
  { title: 'User', value: 'user_email', sortable: false },
  { title: 'Target', value: 'target_user_email', sortable: false },
  { title: 'IP Address', value: 'ip_address', sortable: false },
  { title: 'Details', value: 'details', sortable: false, align: 'end' }
]

const eventTypes = [
  { title: 'All Events', value: null },
  { title: 'Login Success', value: 'login_success' },
  { title: 'Login Failed', value: 'login_failed' },
  { title: 'Account Locked (Auto)', value: 'account_locked_auto' },
  { title: 'Account Unlocked', value: 'account_unlocked' },
  { title: 'Account Disabled', value: 'account_disabled' },
  { title: 'Account Enabled', value: 'account_enabled' },
  { title: 'Role Changed', value: 'role_changed' },
  { title: 'Password Changed', value: 'password_changed' },
  { title: 'Password Reset Requested', value: 'password_reset_requested' },
  { title: 'Password Reset Completed', value: 'password_reset_completed' },
  { title: 'Email Verification Sent', value: 'email_verification_sent' },
  { title: 'Email Verified', value: 'email_verified' },
  { title: 'Registered', value: 'registered' }
]

function formatDate(dateString) {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now - date
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`

  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
}

function formatFullDate(dateString) {
  const date = new Date(dateString)
  return date.toLocaleString()
}

function formatEventType(eventType) {
  return eventType
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

function getEventColor(eventType) {
  if (eventType.includes('success') || eventType.includes('enabled') || eventType.includes('verified')) {
    return 'success'
  }
  if (eventType.includes('failed') || eventType.includes('locked') || eventType.includes('disabled')) {
    return 'error'
  }
  if (eventType.includes('changed') || eventType.includes('unlocked')) {
    return 'warning'
  }
  return 'default'
}

async function loadLogs() {
  loading.value = true
  error.value = null
  try {
    const params = new URLSearchParams({
      limit: limit.value,
      offset: offset.value
    })

    if (filterEventType.value) {
      params.append('event_type', filterEventType.value)
    }
    if (filterUserEmail.value) {
      params.append('user_email', filterUserEmail.value)
    }

    const response = await axios.get(`/api/admin/audit-logs?${params}`)
    logs.value = response.data.logs || []
    total.value = response.data.total || 0
  } catch (e) {
    console.error('Failed to load audit logs:', e)
    error.value = e.response?.data?.message || 'Failed to load audit logs'
  } finally {
    loading.value = false
  }
}

function showDetails(log) {
  selectedLog.value = log
  detailsDialog.value = true
}

function previousPage() {
  if (offset.value >= limit.value) {
    offset.value -= limit.value
    loadLogs()
  }
}

function nextPage() {
  if (offset.value + limit.value < total.value) {
    offset.value += limit.value
    loadLogs()
  }
}

onMounted(() => {
  loadLogs()
})
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}

.white-space-pre-wrap {
  white-space: pre-wrap;
}

code {
  background-color: rgba(0, 0, 0, 0.05);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.875em;
}
</style>
