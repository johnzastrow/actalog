<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
    <div class="d-flex align-center mb-4">
      <v-btn icon class="mr-2" @click="$router.back()">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <div>
        <h1 class="text-h5">Data Change Logs</h1>
        <div class="text-body-2 text-medium-emphasis">View history of edits and deletions</div>
      </div>
    </div>

    <!-- Loading State -->
    <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />

    <!-- Error Alert -->
    <v-alert v-if="error" type="error" variant="tonal" closable class="mb-4" @click:close="error = null">
      {{ error }}
    </v-alert>

    <!-- Filters Card -->
    <v-card elevation="0" rounded="lg" class="mb-4">
      <v-card-text>
        <v-row>
          <v-col cols="12" md="3">
            <v-select
              v-model="filterEntityType"
              :items="entityTypes"
              label="Entity Type"
              clearable
              variant="outlined"
              density="compact"
              hide-details
              @update:model-value="loadLogs"
            />
          </v-col>
          <v-col cols="12" md="3">
            <v-select
              v-model="filterOperation"
              :items="operations"
              label="Operation"
              clearable
              variant="outlined"
              density="compact"
              hide-details
              @update:model-value="loadLogs"
            />
          </v-col>
          <v-col cols="12" md="3">
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
          <v-col cols="12" md="3">
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

    <!-- Data Change Logs Table -->
    <v-card elevation="0" rounded="lg">
      <v-card-title class="d-flex align-center">
        <v-icon class="mr-2">mdi-history</v-icon>
        Data Change Logs ({{ total }})
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
        <!-- eslint-disable-next-line vue/valid-v-slot -->
        <template #item.created_at="{ item }">
          <div class="text-body-2">
            {{ formatDate(item.created_at) }}
          </div>
        </template>

        <!-- Entity Type Column -->
        <!-- eslint-disable-next-line vue/valid-v-slot -->
        <template #item.entity_type="{ item }">
          <v-chip :color="getEntityColor(item.entity_type)" size="small">
            {{ formatEntityType(item.entity_type) }}
          </v-chip>
        </template>

        <!-- Entity Name Column -->
        <!-- eslint-disable-next-line vue/valid-v-slot -->
        <template #item.entity_name="{ item }">
          <div class="text-body-2">
            {{ item.entity_name || `#${item.entity_id}` }}
          </div>
        </template>

        <!-- Operation Column -->
        <!-- eslint-disable-next-line vue/valid-v-slot -->
        <template #item.operation="{ item }">
          <v-chip :color="getOperationColor(item.operation)" size="small" variant="outlined">
            <v-icon start size="small">{{ getOperationIcon(item.operation) }}</v-icon>
            {{ formatOperation(item.operation) }}
          </v-chip>
        </template>

        <!-- User Column -->
        <!-- eslint-disable-next-line vue/valid-v-slot -->
        <template #item.user_email="{ item }">
          <div v-if="item.user_email" class="text-body-2">
            {{ item.user_email }}
          </div>
          <span v-else class="text-body-2 text-medium-emphasis">System</span>
        </template>

        <!-- Details Column -->
        <!-- eslint-disable-next-line vue/valid-v-slot -->
        <template #item.details="{ item }">
          <v-btn
            icon
            size="small"
            variant="text"
            @click="showDetails(item)"
          >
            <v-icon>mdi-eye</v-icon>
          </v-btn>
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
    <v-dialog v-model="detailsDialog" max-width="800">
      <v-card v-if="selectedLog">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-information</v-icon>
          Data Change Details
        </v-card-title>
        <v-divider />
        <v-card-text>
          <!-- Summary Info -->
          <v-list density="compact">
            <v-list-item>
              <template #prepend>
                <v-icon size="small" class="mr-2">mdi-clock</v-icon>
              </template>
              <v-list-item-title class="text-subtitle-2">Timestamp</v-list-item-title>
              <v-list-item-subtitle>{{ formatFullDate(selectedLog.created_at) }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item>
              <template #prepend>
                <v-icon size="small" class="mr-2">mdi-database</v-icon>
              </template>
              <v-list-item-title class="text-subtitle-2">Entity</v-list-item-title>
              <v-list-item-subtitle>
                <v-chip :color="getEntityColor(selectedLog.entity_type)" size="small" class="mr-2">
                  {{ formatEntityType(selectedLog.entity_type) }}
                </v-chip>
                {{ selectedLog.entity_name || `#${selectedLog.entity_id}` }}
              </v-list-item-subtitle>
            </v-list-item>

            <v-list-item>
              <template #prepend>
                <v-icon size="small" class="mr-2">mdi-pencil</v-icon>
              </template>
              <v-list-item-title class="text-subtitle-2">Operation</v-list-item-title>
              <v-list-item-subtitle>
                <v-chip :color="getOperationColor(selectedLog.operation)" size="small">
                  {{ formatOperation(selectedLog.operation) }}
                </v-chip>
              </v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="selectedLog.user_email">
              <template #prepend>
                <v-icon size="small" class="mr-2">mdi-account</v-icon>
              </template>
              <v-list-item-title class="text-subtitle-2">User</v-list-item-title>
              <v-list-item-subtitle>{{ selectedLog.user_email }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="selectedLog.ip_address">
              <template #prepend>
                <v-icon size="small" class="mr-2">mdi-ip</v-icon>
              </template>
              <v-list-item-title class="text-subtitle-2">IP Address</v-list-item-title>
              <v-list-item-subtitle><code>{{ selectedLog.ip_address }}</code></v-list-item-subtitle>
            </v-list-item>
          </v-list>

          <v-divider class="my-3" />

          <!-- Before/After Comparison -->
          <h4 class="text-subtitle-1 mb-3">Changes</h4>

          <v-row>
            <!-- Before Values -->
            <v-col cols="12" md="6">
              <v-card variant="outlined" class="pa-3" color="error">
                <div class="d-flex align-center mb-2">
                  <v-icon color="error" size="small" class="mr-2">mdi-minus-circle</v-icon>
                  <span class="text-subtitle-2 text-error">Before</span>
                </div>
                <pre class="before-after-pre">{{ formatJSON(selectedLog.before_values) }}</pre>
              </v-card>
            </v-col>

            <!-- After Values -->
            <v-col cols="12" md="6">
              <v-card variant="outlined" class="pa-3" :color="selectedLog.operation === 'delete' ? 'grey' : 'success'">
                <div class="d-flex align-center mb-2">
                  <v-icon :color="selectedLog.operation === 'delete' ? 'grey' : 'success'" size="small" class="mr-2">
                    {{ selectedLog.operation === 'delete' ? 'mdi-delete' : 'mdi-plus-circle' }}
                  </v-icon>
                  <span class="text-subtitle-2" :class="selectedLog.operation === 'delete' ? 'text-grey' : 'text-success'">
                    {{ selectedLog.operation === 'delete' ? 'Deleted' : 'After' }}
                  </span>
                </div>
                <pre class="before-after-pre">{{ selectedLog.operation === 'delete' ? 'Record deleted' : formatJSON(selectedLog.after_values) }}</pre>
              </v-card>
            </v-col>
          </v-row>

          <!-- Diff View -->
          <div v-if="selectedLog.operation === 'update' && changedFields.length > 0" class="mt-4">
            <h4 class="text-subtitle-1 mb-3">Changed Fields</h4>
            <v-table density="compact">
              <thead>
                <tr>
                  <th>Field</th>
                  <th>Old Value</th>
                  <th>New Value</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="field in changedFields" :key="field.name">
                  <td class="font-weight-medium">{{ field.name }}</td>
                  <td class="text-error"><code>{{ field.oldValue }}</code></td>
                  <td class="text-success"><code>{{ field.newValue }}</code></td>
                </tr>
              </tbody>
            </v-table>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="detailsDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from '@/utils/axios'

const loading = ref(false)
const error = ref(null)
const logs = ref([])
const total = ref(0)
const limit = ref(50)
const offset = ref(0)

const filterEntityType = ref(null)
const filterOperation = ref(null)
const filterUserEmail = ref('')

const detailsDialog = ref(false)
const selectedLog = ref(null)

const headers = [
  { title: 'Timestamp', value: 'created_at', sortable: false },
  { title: 'Entity Type', value: 'entity_type', sortable: false },
  { title: 'Entity Name', value: 'entity_name', sortable: false },
  { title: 'Operation', value: 'operation', sortable: false },
  { title: 'User', value: 'user_email', sortable: false },
  { title: 'Details', value: 'details', sortable: false, align: 'end' }
]

const entityTypes = [
  { title: 'All Types', value: null },
  { title: 'WOD', value: 'wod' },
  { title: 'Movement', value: 'movement' },
  { title: 'Workout', value: 'workout' },
  { title: 'User Workout', value: 'user_workout' },
  { title: 'Performance', value: 'performance' },
  { title: 'PR', value: 'pr' }
]

const operations = [
  { title: 'All Operations', value: null },
  { title: 'Update', value: 'update' },
  { title: 'Delete', value: 'delete' }
]

// Helper to parse JSON values (they come as strings from the API)
function parseValues(val) {
  if (!val) return null
  if (typeof val === 'object') return val
  try {
    return JSON.parse(val)
  } catch {
    return null
  }
}

// Computed property for changed fields in update operations
const changedFields = computed(() => {
  if (!selectedLog.value || selectedLog.value.operation !== 'update') return []

  const before = parseValues(selectedLog.value.before_values) || {}
  const after = parseValues(selectedLog.value.after_values) || {}
  const changes = []

  // Get all keys from both objects
  const allKeys = new Set([...Object.keys(before), ...Object.keys(after)])

  for (const key of allKeys) {
    // Skip certain metadata fields
    if (['updated_at', 'created_at', 'id'].includes(key)) continue

    const oldVal = before[key]
    const newVal = after[key]

    if (JSON.stringify(oldVal) !== JSON.stringify(newVal)) {
      changes.push({
        name: formatFieldName(key),
        oldValue: formatValue(oldVal),
        newValue: formatValue(newVal)
      })
    }
  }

  return changes
})

function formatFieldName(name) {
  return name
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

function formatValue(value) {
  if (value === null || value === undefined) return '-'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

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

function formatEntityType(entityType) {
  return entityType
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

function formatOperation(operation) {
  return operation.charAt(0).toUpperCase() + operation.slice(1)
}

function getEntityColor(entityType) {
  const colors = {
    wod: 'blue',
    movement: 'green',
    workout: 'purple',
    user_workout: 'purple',
    performance: 'orange',
    pr: 'amber'
  }
  return colors[entityType] || 'grey'
}

function getOperationColor(operation) {
  return operation === 'delete' ? 'error' : 'warning'
}

function getOperationIcon(operation) {
  return operation === 'delete' ? 'mdi-delete' : 'mdi-pencil'
}

function formatJSON(val) {
  if (!val) return 'No data'
  try {
    const parsed = parseValues(val)
    if (!parsed) return 'No data'
    return JSON.stringify(parsed, null, 2)
  } catch {
    return String(val)
  }
}

async function loadLogs() {
  loading.value = true
  error.value = null
  try {
    const params = new URLSearchParams({
      limit: limit.value,
      offset: offset.value
    })

    if (filterEntityType.value) {
      params.append('entity_type', filterEntityType.value)
    }
    if (filterOperation.value) {
      params.append('operation', filterOperation.value)
    }
    if (filterUserEmail.value) {
      params.append('user_email', filterUserEmail.value)
    }

    const response = await axios.get(`/api/admin/data-change-logs?${params}`)
    logs.value = response.data.logs || []
    total.value = response.data.total || 0
  } catch (e) {
    console.error('Failed to load data change logs:', e)
    error.value = e.response?.data?.message || 'Failed to load data change logs'
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

.before-after-pre {
  background-color: rgba(0, 0, 0, 0.03);
  padding: 12px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.8em;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 300px;
  overflow-y: auto;
  margin: 0;
}

code {
  background-color: rgba(0, 0, 0, 0.05);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.875em;
}
</style>
