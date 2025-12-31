<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <div class="d-flex align-center mb-4">
        <v-btn icon class="mr-2" @click="$router.back()">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <div>
          <h1 class="text-h5">Email Logs</h1>
          <div class="text-body-2 text-medium-emphasis">View history of sent emails</div>
        </div>
        <v-spacer />
        <v-btn
          color="secondary"
          prepend-icon="mdi-email-cog"
          :to="{ name: 'admin-email-settings' }"
          class="mr-2"
        >
          Email Settings
        </v-btn>
        <v-btn
          color="warning"
          prepend-icon="mdi-delete-sweep"
          :loading="cleaning"
          @click="showCleanupDialog"
        >
          Cleanup
        </v-btn>
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

      <!-- Filters Card -->
      <v-card elevation="0" rounded="lg" class="mb-4">
        <v-expansion-panels variant="accordion">
          <v-expansion-panel>
            <v-expansion-panel-title>
              <v-icon class="mr-2">mdi-filter</v-icon>
              Filters
              <v-chip v-if="hasActiveFilters" size="small" color="primary" class="ml-2">
                Active
              </v-chip>
            </v-expansion-panel-title>
            <v-expansion-panel-text>
              <v-row>
                <v-col cols="12" md="3">
                  <v-select
                    v-model="filters.type"
                    label="Email Type"
                    :items="emailTypes"
                    clearable
                    variant="outlined"
                    density="compact"
                  />
                </v-col>
                <v-col cols="12" md="3">
                  <v-text-field
                    v-model="filters.recipient"
                    label="Recipient"
                    clearable
                    variant="outlined"
                    density="compact"
                    prepend-inner-icon="mdi-email-search"
                  />
                </v-col>
                <v-col cols="12" md="3">
                  <v-select
                    v-model="filters.success"
                    label="Status"
                    :items="statusOptions"
                    clearable
                    variant="outlined"
                    density="compact"
                  />
                </v-col>
                <v-col cols="12" md="3">
                  <v-btn color="primary" block @click="applyFilters">
                    Apply Filters
                  </v-btn>
                </v-col>
              </v-row>
            </v-expansion-panel-text>
          </v-expansion-panel>
        </v-expansion-panels>
      </v-card>

      <!-- Email Logs Table -->
      <v-card elevation="0" rounded="lg">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-email-multiple</v-icon>
          Email History ({{ totalCount }})
        </v-card-title>

        <v-divider />

        <!-- Empty State -->
        <v-card-text v-if="!loading && logs.length === 0" class="text-center py-12">
          <v-icon size="64" color="grey-lighten-1" class="mb-4">mdi-email-off</v-icon>
          <div class="text-h6 mb-2">No Email Logs Found</div>
          <div class="text-body-2 text-medium-emphasis mb-4">
            Email logs will appear here after emails are sent
          </div>
        </v-card-text>

        <!-- Logs List -->
        <v-data-table
          v-else
          :headers="headers"
          :items="logs"
          :loading="loading"
          :items-per-page="itemsPerPage"
          @update:items-per-page="updateItemsPerPage"
          :items-per-page-options="[10, 20, 50, 100]"
          class="elevation-0"
        >
          <!-- Timestamp Column -->
          <template #item.created_at="{ item }">
            <div>
              <div class="text-body-2">{{ formatDate(item.created_at) }}</div>
              <div class="text-caption text-medium-emphasis">{{ formatTime(item.created_at) }}</div>
            </div>
          </template>

          <!-- Type Column -->
          <template #item.email_type="{ item }">
            <v-chip
              :color="getTypeColor(item.email_type)"
              size="small"
              variant="flat"
            >
              {{ formatType(item.email_type) }}
            </v-chip>
          </template>

          <!-- Recipient Column -->
          <template #item.recipient_email="{ item }">
            <div class="d-flex align-center">
              <v-icon size="small" class="mr-2">mdi-email</v-icon>
              <span>{{ item.recipient_email }}</span>
            </div>
          </template>

          <!-- Subject Column -->
          <template #item.subject="{ item }">
            <div class="text-truncate" style="max-width: 200px;" :title="item.subject">
              {{ item.subject }}
            </div>
          </template>

          <!-- Status Column -->
          <template #item.success="{ item }">
            <v-chip
              :color="item.success ? 'success' : 'error'"
              size="small"
              variant="flat"
            >
              <v-icon size="small" start>
                {{ item.success ? 'mdi-check' : 'mdi-close' }}
              </v-icon>
              {{ item.success ? 'Sent' : 'Failed' }}
            </v-chip>
          </template>

          <!-- Actions Column -->
          <template #item.actions="{ item }">
            <v-tooltip text="View Details" location="top">
              <template #activator="{ props }">
                <v-btn
                  v-bind="props"
                  icon
                  size="small"
                  variant="text"
                  @click="showLogDetails(item)"
                >
                  <v-icon>mdi-eye</v-icon>
                </v-btn>
              </template>
            </v-tooltip>
          </template>
        </v-data-table>

        <!-- Pagination -->
        <v-divider />
        <v-card-actions>
          <v-btn
            :disabled="offset === 0"
            @click="previousPage"
          >
            Previous
          </v-btn>
          <v-spacer />
          <span class="text-body-2 text-medium-emphasis">
            Showing {{ offset + 1 }} - {{ Math.min(offset + itemsPerPage, totalCount) }} of {{ totalCount }}
          </span>
          <v-spacer />
          <v-btn
            :disabled="offset + itemsPerPage >= totalCount"
            @click="nextPage"
          >
            Next
          </v-btn>
        </v-card-actions>
      </v-card>

      <!-- Log Details Dialog -->
      <v-dialog v-model="detailsDialog" max-width="700">
        <v-card v-if="selectedLog">
          <v-card-title class="d-flex align-center">
            <v-icon class="mr-2">mdi-email-open</v-icon>
            Email Log Details
          </v-card-title>
          <v-divider />
          <v-card-text>
            <v-row>
              <v-col cols="12" md="6">
                <div class="detail-item">
                  <div class="detail-label">Recipient</div>
                  <div class="detail-value">{{ selectedLog.recipient_email }}</div>
                </div>
              </v-col>
              <v-col cols="12" md="6">
                <div class="detail-item">
                  <div class="detail-label">Type</div>
                  <v-chip :color="getTypeColor(selectedLog.email_type)" size="small" variant="flat">
                    {{ formatType(selectedLog.email_type) }}
                  </v-chip>
                </div>
              </v-col>
              <v-col cols="12">
                <div class="detail-item">
                  <div class="detail-label">Subject</div>
                  <div class="detail-value">{{ selectedLog.subject }}</div>
                </div>
              </v-col>
              <v-col cols="12" md="6">
                <div class="detail-item">
                  <div class="detail-label">Status</div>
                  <v-chip :color="selectedLog.success ? 'success' : 'error'" size="small" variant="flat">
                    {{ selectedLog.success ? 'Sent' : 'Failed' }}
                  </v-chip>
                </div>
              </v-col>
              <v-col cols="12" md="6">
                <div class="detail-item">
                  <div class="detail-label">Sent At</div>
                  <div class="detail-value">{{ formatDateTime(selectedLog.created_at) }}</div>
                </div>
              </v-col>
              <v-col v-if="selectedLog.sent_by_user_email" cols="12">
                <div class="detail-item">
                  <div class="detail-label">Sent By</div>
                  <div class="detail-value">{{ selectedLog.sent_by_user_email }}</div>
                </div>
              </v-col>
              <v-col v-if="selectedLog.error_message" cols="12">
                <v-alert type="error" variant="tonal">
                  <strong>Error:</strong> {{ selectedLog.error_message }}
                </v-alert>
              </v-col>
            </v-row>

            <!-- Debug Info -->
            <v-expansion-panels v-if="parsedDebugInfo" variant="accordion" class="mt-4">
              <v-expansion-panel>
                <v-expansion-panel-title>
                  <v-icon class="mr-2">mdi-bug</v-icon>
                  SMTP Debug Information
                </v-expansion-panel-title>
                <v-expansion-panel-text>
                  <div class="debug-info">
                    <div class="mb-3">
                      <strong>Connection:</strong> {{ parsedDebugInfo.connection_host }}:{{ parsedDebugInfo.connection_port }}
                      ({{ parsedDebugInfo.connection_tls }})
                      <span class="text-caption ml-2">{{ parsedDebugInfo.connection_time }}</span>
                    </div>
                    <div class="mb-3">
                      <strong>Authentication:</strong> {{ parsedDebugInfo.auth_method || 'PLAIN' }}
                      <span class="text-caption ml-2">{{ parsedDebugInfo.auth_time }}</span>
                    </div>
                    <div class="mb-3">
                      <strong>Send:</strong>
                      <span class="text-caption ml-2">{{ parsedDebugInfo.send_time }}</span>
                    </div>
                    <div class="mb-3">
                      <strong>Total Duration:</strong> {{ parsedDebugInfo.total_duration }}
                    </div>
                    <div v-if="parsedDebugInfo.smtp_responses && parsedDebugInfo.smtp_responses.length > 0">
                      <strong>SMTP Responses:</strong>
                      <v-sheet class="pa-2 mt-2 rounded" color="grey-darken-4">
                        <code class="smtp-responses">
                          <div v-for="(response, index) in parsedDebugInfo.smtp_responses" :key="index">
                            {{ response }}
                          </div>
                        </code>
                      </v-sheet>
                    </div>
                  </div>
                </v-expansion-panel-text>
              </v-expansion-panel>
            </v-expansion-panels>
          </v-card-text>
          <v-divider />
          <v-card-actions>
            <v-spacer />
            <v-btn variant="text" @click="detailsDialog = false">Close</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>

      <!-- Cleanup Confirmation Dialog -->
      <v-dialog v-model="cleanupDialog" max-width="500">
        <v-card>
          <v-card-title class="d-flex align-center">
            <v-icon color="warning" class="mr-2">mdi-delete-sweep</v-icon>
            Cleanup Old Email Logs
          </v-card-title>
          <v-divider />
          <v-card-text class="pt-4">
            <v-alert type="info" variant="tonal" class="mb-4">
              This will delete all email logs older than 90 days.
            </v-alert>
            <div class="text-body-2">
              Are you sure you want to proceed? This action cannot be undone.
            </div>
          </v-card-text>
          <v-divider />
          <v-card-actions>
            <v-spacer />
            <v-btn variant="text" @click="cleanupDialog = false">Cancel</v-btn>
            <v-btn
              color="warning"
              variant="flat"
              :loading="cleaning"
              @click="confirmCleanup"
            >
              Cleanup Logs
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-container>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from '@/utils/axios'

// State
const loading = ref(false)
const cleaning = ref(false)
const error = ref(null)
const successMessage = ref(null)
const logs = ref([])
const totalCount = ref(0)
const offset = ref(0)
const itemsPerPage = ref(20)
const detailsDialog = ref(false)
const cleanupDialog = ref(false)
const selectedLog = ref(null)

// Filters
const filters = ref({
  type: null,
  recipient: null,
  success: null
})

const emailTypes = [
  { title: 'Test', value: 'test' },
  { title: 'Password Reset', value: 'password_reset' },
  { title: 'Verification', value: 'verification' },
  { title: 'Notification', value: 'notification' }
]

const statusOptions = [
  { title: 'Sent', value: 'true' },
  { title: 'Failed', value: 'false' }
]

// Table headers
const headers = [
  { title: 'Timestamp', key: 'created_at', sortable: false },
  { title: 'Type', key: 'email_type', sortable: false },
  { title: 'Recipient', key: 'recipient_email', sortable: false },
  { title: 'Subject', key: 'subject', sortable: false },
  { title: 'Status', key: 'success', sortable: false },
  { title: 'Actions', key: 'actions', sortable: false, align: 'end' }
]

// Computed
const hasActiveFilters = computed(() => {
  return filters.value.type || filters.value.recipient || filters.value.success !== null
})

const parsedDebugInfo = computed(() => {
  if (!selectedLog.value?.debug_info) return null
  try {
    return JSON.parse(selectedLog.value.debug_info)
  } catch {
    return null
  }
})

// Methods
const fetchLogs = async () => {
  loading.value = true
  error.value = null

  try {
    const params = new URLSearchParams()
    params.append('limit', itemsPerPage.value.toString())
    params.append('offset', offset.value.toString())

    if (filters.value.type) {
      params.append('type', filters.value.type)
    }
    if (filters.value.recipient) {
      params.append('recipient', filters.value.recipient)
    }
    if (filters.value.success !== null) {
      params.append('success', filters.value.success)
    }

    const response = await axios.get(`/api/admin/email-logs?${params.toString()}`)
    logs.value = response.data.logs || []
    totalCount.value = response.data.count || 0
  } catch (err) {
    console.error('Failed to fetch email logs:', err)
    error.value = err.response?.data?.error || 'Failed to load email logs'
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  offset.value = 0
  fetchLogs()
}

const previousPage = () => {
  if (offset.value > 0) {
    offset.value = Math.max(0, offset.value - itemsPerPage.value)
    fetchLogs()
  }
}

const nextPage = () => {
  if (offset.value + itemsPerPage.value < totalCount.value) {
    offset.value += itemsPerPage.value
    fetchLogs()
  }
}

const updateItemsPerPage = (value) => {
  itemsPerPage.value = value
  offset.value = 0
  fetchLogs()
}

const showLogDetails = (log) => {
  selectedLog.value = log
  detailsDialog.value = true
}

const showCleanupDialog = () => {
  cleanupDialog.value = true
}

const confirmCleanup = async () => {
  cleaning.value = true
  error.value = null
  successMessage.value = null

  try {
    const response = await axios.post('/api/admin/email-logs/cleanup')
    successMessage.value = `Cleaned up ${response.data.deleted_count} old email logs`
    cleanupDialog.value = false
    fetchLogs()
  } catch (err) {
    console.error('Failed to cleanup email logs:', err)
    error.value = err.response?.data?.error || 'Failed to cleanup email logs'
  } finally {
    cleaning.value = false
  }
}

// Format functions
function formatDate(dateString) {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

function formatTime(dateString) {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

function formatDateTime(dateString) {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function formatType(type) {
  const typeMap = {
    'test': 'Test',
    'password_reset': 'Password Reset',
    'verification': 'Verification',
    'notification': 'Notification'
  }
  return typeMap[type] || type
}

function getTypeColor(type) {
  const colorMap = {
    'test': 'info',
    'password_reset': 'warning',
    'verification': 'primary',
    'notification': 'success'
  }
  return colorMap[type] || 'grey'
}

// Fetch logs on mount
onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
.detail-item {
  margin-bottom: 8px;
}

.detail-label {
  font-size: 0.75rem;
  color: rgba(var(--v-theme-on-surface), 0.6);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-value {
  font-size: 1rem;
}

.debug-info {
  font-size: 0.875rem;
}

.smtp-responses {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.75rem;
  line-height: 1.4;
  color: #4caf50;
  display: block;
}
</style>
