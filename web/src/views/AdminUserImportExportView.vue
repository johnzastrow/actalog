<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <!-- Page Title -->
      <div class="mb-4">
        <div class="d-flex align-center mb-2">
          <v-btn icon variant="text" size="small" @click="$router.push('/admin')">
            <v-icon>mdi-arrow-left</v-icon>
          </v-btn>
          <h1 style="color: #2c3e50; font-size: 24px; font-weight: 600" class="ml-2">User Import/Export</h1>
        </div>
        <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px">
          Import users from CSV, export user list, or send batch password reset emails
        </p>
      </div>

      <!-- Success Alert -->
      <v-alert v-if="successMessage" type="success" closable class="mb-4" @click:close="successMessage = null">
        {{ successMessage }}
      </v-alert>

      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable class="mb-4" @click:close="error = null">
        {{ error }}
      </v-alert>

      <!-- Tabs -->
      <v-tabs v-model="activeTab" bg-color="surface" class="mb-4" grow>
        <v-tab value="import">
          <v-icon start>mdi-upload</v-icon>
          Import
        </v-tab>
        <v-tab value="export">
          <v-icon start>mdi-download</v-icon>
          Export
        </v-tab>
        <v-tab value="reset">
          <v-icon start>mdi-email-send</v-icon>
          Password Reset
        </v-tab>
      </v-tabs>

      <v-tabs-window v-model="activeTab">
        <!-- Import Tab -->
        <v-tabs-window-item value="import">
          <!-- Step 1: CSV Format Info -->
          <v-card v-if="!previewResult" elevation="0" rounded="lg" class="pa-4 mb-4" bg-color="surface">
            <h2 class="text-h6 font-weight-bold mb-3">CSV Format</h2>
            <v-alert type="info" density="compact" class="text-caption">
              <strong>Required columns:</strong> email, name, password<br>
              <code class="text-caption">email,name,password</code><br>
              <code class="text-caption">user@example.com,John Doe,SecurePass123</code>
            </v-alert>
            <p class="text-caption text-medium-emphasis mt-3">
              All imported users will be created with role "user" and a permanent free subscription.
              Passwords must be at least 8 characters.
            </p>
          </v-card>

          <!-- Step 2: File Upload -->
          <v-card v-if="!previewResult" elevation="0" rounded="lg" class="pa-4 mb-4" bg-color="surface">
            <h2 class="text-h6 font-weight-bold mb-3">Upload CSV File</h2>

            <div
              class="upload-zone"
              :class="{ 'drag-over': dragOver }"
              @drop.prevent="handleDrop"
              @dragover.prevent="dragOver = true"
              @dragleave.prevent="dragOver = false"
              @click="$refs.fileInput.click()"
            >
              <v-icon size="64" :color="selectedFile ? 'primary' : '#ccc'">
                {{ selectedFile ? 'mdi-file-check' : 'mdi-cloud-upload' }}
              </v-icon>
              <p class="text-body-1 font-weight-bold mt-3" :style="{ color: selectedFile ? 'rgb(var(--v-theme-primary))' : 'rgb(var(--v-theme-on-surface))' }">
                {{ selectedFile ? selectedFile.name : 'Drop CSV file here or click to browse' }}
              </p>
              <p v-if="!selectedFile" class="text-caption text-disabled">
                Maximum file size: 10MB
              </p>
              <input
                ref="fileInput"
                type="file"
                accept=".csv"
                style="display: none"
                @change="handleFileSelect"
              >
            </div>

            <v-btn
              v-if="selectedFile"
              block
              size="large"
              color="primary"
              rounded="lg"
              elevation="2"
              :loading="loading"
              class="mt-4 font-weight-bold"
              style="text-transform: none"
              @click="previewImport"
            >
              <v-icon start>mdi-eye</v-icon>
              Preview Import
            </v-btn>
          </v-card>

          <!-- Step 3: Preview Results -->
          <template v-if="previewResult">
            <v-card elevation="0" rounded="lg" class="pa-4 mb-4" bg-color="surface">
              <h2 class="text-h6 font-weight-bold mb-3">Preview Results</h2>

              <v-row dense>
                <v-col cols="6" sm="3">
                  <div class="stat-box">
                    <p class="text-caption text-medium-emphasis">Total Rows</p>
                    <p class="text-h6 font-weight-bold">{{ previewResult.total_rows }}</p>
                  </div>
                </v-col>
                <v-col cols="6" sm="3">
                  <div class="stat-box">
                    <p class="text-caption text-medium-emphasis">Valid</p>
                    <p class="text-h6 font-weight-bold" style="color: rgb(var(--v-theme-success))">
                      {{ previewResult.valid_rows }}
                    </p>
                  </div>
                </v-col>
                <v-col cols="6" sm="3">
                  <div class="stat-box">
                    <p class="text-caption text-medium-emphasis">Duplicates</p>
                    <p class="text-h6 font-weight-bold" style="color: rgb(var(--v-theme-warning))">
                      {{ previewResult.duplicate_rows }}
                    </p>
                  </div>
                </v-col>
                <v-col cols="6" sm="3">
                  <div class="stat-box">
                    <p class="text-caption text-medium-emphasis">Invalid</p>
                    <p class="text-h6 font-weight-bold" style="color: rgb(var(--v-theme-error))">
                      {{ previewResult.invalid_rows }}
                    </p>
                  </div>
                </v-col>
              </v-row>
            </v-card>

            <!-- Row Details -->
            <v-card elevation="0" rounded="lg" class="pa-4 mb-4" bg-color="surface">
              <h3 class="text-subtitle-1 font-weight-bold mb-3">Row Details</h3>
              <v-table density="compact">
                <thead>
                  <tr>
                    <th class="text-left">Row</th>
                    <th class="text-left">Email</th>
                    <th class="text-left">Name</th>
                    <th class="text-left">Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in previewResult.rows" :key="row.row_number">
                    <td>{{ row.row_number }}</td>
                    <td>{{ row.email }}</td>
                    <td>{{ row.name }}</td>
                    <td>
                      <v-chip v-if="row.is_valid && !row.is_duplicate" size="x-small" color="success">Valid</v-chip>
                      <v-chip v-else-if="row.is_duplicate" size="x-small" color="warning">Duplicate</v-chip>
                      <v-chip v-else size="x-small" color="error">
                        {{ row.errors?.join(', ') || 'Invalid' }}
                      </v-chip>
                    </td>
                  </tr>
                </tbody>
              </v-table>
            </v-card>

            <!-- Import Options -->
            <v-card elevation="0" rounded="lg" class="pa-4 mb-4" bg-color="surface">
              <v-checkbox
                v-model="skipDuplicates"
                label="Skip duplicate emails (recommended)"
                density="compact"
                hide-details
              />

              <div class="d-flex gap-3 mt-4">
                <v-btn
                  variant="outlined"
                  rounded="lg"
                  @click="resetImport"
                >
                  <v-icon start>mdi-arrow-left</v-icon>
                  Back
                </v-btn>
                <v-btn
                  color="primary"
                  rounded="lg"
                  :loading="loading"
                  :disabled="previewResult.valid_rows === 0"
                  @click="confirmImport"
                >
                  <v-icon start>mdi-check</v-icon>
                  Import {{ previewResult.valid_rows }} Users
                </v-btn>
              </div>
            </v-card>
          </template>
        </v-tabs-window-item>

        <!-- Export Tab -->
        <v-tabs-window-item value="export">
          <v-card elevation="0" rounded="lg" class="pa-4" bg-color="surface">
            <h2 class="text-h6 font-weight-bold mb-3">Export Users to CSV</h2>
            <p class="text-body-2 text-medium-emphasis mb-4">
              Download a CSV file containing all user email addresses and names.
              This file can be used as a template for importing new users.
            </p>
            <v-alert type="info" density="compact" class="mb-4 text-caption">
              <strong>Export format:</strong> email, name<br>
              Passwords are NOT included in the export for security reasons.
            </v-alert>
            <v-btn
              color="primary"
              rounded="lg"
              :loading="loading"
              @click="exportUsers"
            >
              <v-icon start>mdi-download</v-icon>
              Download Users CSV
            </v-btn>
          </v-card>
        </v-tabs-window-item>

        <!-- Password Reset Tab -->
        <v-tabs-window-item value="reset">
          <v-card elevation="0" rounded="lg" class="pa-4 mb-4" bg-color="surface">
            <h2 class="text-h6 font-weight-bold mb-3">Batch Password Reset</h2>
            <p class="text-body-2 text-medium-emphasis mb-4">
              Select users from the list below to send them password reset emails.
              Users will receive an email with a link to set a new password.
            </p>

            <!-- Filters -->
            <v-row dense class="mb-4">
              <v-col cols="12" sm="6">
                <v-text-field
                  v-model="searchQuery"
                  label="Search by name or email"
                  prepend-inner-icon="mdi-magnify"
                  density="compact"
                  variant="outlined"
                  clearable
                  hide-details
                  @update:model-value="debouncedSearch"
                />
              </v-col>
              <v-col cols="6" sm="3">
                <v-text-field
                  v-model="createdAfter"
                  label="Created after"
                  type="date"
                  density="compact"
                  variant="outlined"
                  hide-details
                  @update:model-value="loadUsers"
                />
              </v-col>
              <v-col cols="6" sm="3">
                <v-text-field
                  v-model="createdBefore"
                  label="Created before"
                  type="date"
                  density="compact"
                  variant="outlined"
                  hide-details
                  @update:model-value="loadUsers"
                />
              </v-col>
            </v-row>

            <!-- User Selection Table -->
            <v-table density="compact">
              <thead>
                <tr>
                  <th style="width: 50px">
                    <v-checkbox
                      :model-value="allSelected"
                      :indeterminate="someSelected && !allSelected"
                      density="compact"
                      hide-details
                      @update:model-value="toggleSelectAll"
                    />
                  </th>
                  <th class="text-left">Email</th>
                  <th class="text-left">Name</th>
                  <th class="text-left">Created</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="loadingUsers">
                  <td colspan="4" class="text-center pa-4">
                    <v-progress-circular indeterminate size="24" />
                  </td>
                </tr>
                <tr v-else-if="users.length === 0">
                  <td colspan="4" class="text-center pa-4 text-medium-emphasis">
                    No users found
                  </td>
                </tr>
                <tr v-for="user in users" :key="user.id">
                  <td>
                    <v-checkbox
                      :model-value="selectedUserIds.includes(user.id)"
                      density="compact"
                      hide-details
                      @update:model-value="toggleUserSelection(user.id)"
                    />
                  </td>
                  <td>{{ user.email }}</td>
                  <td>{{ user.name }}</td>
                  <td>{{ formatDate(user.created_at) }}</td>
                </tr>
              </tbody>
            </v-table>

            <!-- Pagination -->
            <div class="d-flex justify-space-between align-center mt-4">
              <span class="text-caption text-medium-emphasis">
                Showing {{ users.length }} of {{ totalUsers }} users
              </span>
              <div class="d-flex gap-2">
                <v-btn
                  size="small"
                  variant="outlined"
                  :disabled="offset === 0"
                  @click="prevPage"
                >
                  Previous
                </v-btn>
                <v-btn
                  size="small"
                  variant="outlined"
                  :disabled="offset + limit >= totalUsers"
                  @click="nextPage"
                >
                  Next
                </v-btn>
              </div>
            </div>
          </v-card>

          <!-- Send Button -->
          <v-card elevation="0" rounded="lg" class="pa-4" bg-color="surface">
            <div class="d-flex align-center justify-space-between">
              <span class="text-body-2">
                {{ selectedUserIds.length }} user(s) selected
              </span>
              <v-btn
                color="primary"
                rounded="lg"
                :loading="loading"
                :disabled="selectedUserIds.length === 0"
                @click="sendPasswordResetEmails"
              >
                <v-icon start>mdi-email-send</v-icon>
                Send Reset Emails
              </v-btn>
            </div>
          </v-card>

          <!-- Reset Result -->
          <v-card v-if="resetResult" elevation="0" rounded="lg" class="pa-4 mt-4" bg-color="surface">
            <h3 class="text-subtitle-1 font-weight-bold mb-3">Result</h3>
            <v-row dense>
              <v-col cols="6">
                <div class="stat-box">
                  <p class="text-caption text-medium-emphasis">Successfully Sent</p>
                  <p class="text-h6 font-weight-bold" style="color: rgb(var(--v-theme-success))">
                    {{ resetResult.successfully_sent }}
                  </p>
                </div>
              </v-col>
              <v-col cols="6">
                <div class="stat-box">
                  <p class="text-caption text-medium-emphasis">Failed</p>
                  <p class="text-h6 font-weight-bold" style="color: rgb(var(--v-theme-error))">
                    {{ resetResult.failed_count }}
                  </p>
                </div>
              </v-col>
            </v-row>
            <v-alert v-if="resetResult.errors?.length" type="warning" density="compact" class="mt-3">
              <strong>Errors:</strong>
              <ul class="text-caption">
                <li v-for="(err, i) in resetResult.errors" :key="i">{{ err }}</li>
              </ul>
            </v-alert>
          </v-card>
        </v-tabs-window-item>
      </v-tabs-window>
    </v-container>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from '@/utils/axios'

// Tab state
const activeTab = ref('import')

// Common state
const loading = ref(false)
const error = ref(null)
const successMessage = ref(null)

// Import state
const selectedFile = ref(null)
const dragOver = ref(false)
const previewResult = ref(null)
const skipDuplicates = ref(true)
const fileInput = ref(null)

// Export state (none needed)

// Password reset state
const users = ref([])
const totalUsers = ref(0)
const loadingUsers = ref(false)
const selectedUserIds = ref([])
const searchQuery = ref('')
const createdAfter = ref('')
const createdBefore = ref('')
const limit = ref(50)
const offset = ref(0)
const resetResult = ref(null)

// Computed
const allSelected = computed(() =>
  users.value.length > 0 && users.value.every(u => selectedUserIds.value.includes(u.id))
)
const someSelected = computed(() =>
  selectedUserIds.value.length > 0
)

// Methods
const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (file) {
    selectedFile.value = file
    error.value = null
  }
}

const handleDrop = (event) => {
  dragOver.value = false
  const file = event.dataTransfer.files[0]
  if (file && file.name.endsWith('.csv')) {
    selectedFile.value = file
    error.value = null
  } else {
    error.value = 'Please upload a CSV file'
  }
}

const previewImport = async () => {
  if (!selectedFile.value) return

  loading.value = true
  error.value = null

  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)

    const response = await axios.post('/api/admin/users/import/preview', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    previewResult.value = response.data
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to preview import'
  } finally {
    loading.value = false
  }
}

const confirmImport = async () => {
  if (!selectedFile.value) return

  loading.value = true
  error.value = null

  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)
    formData.append('skip_duplicates', skipDuplicates.value.toString())

    const response = await axios.post('/api/admin/users/import/confirm', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })

    successMessage.value = `Successfully imported ${response.data.created_count} users`
    resetImport()
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to import users'
  } finally {
    loading.value = false
  }
}

const resetImport = () => {
  selectedFile.value = null
  previewResult.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const exportUsers = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await axios.get('/api/admin/users/export', {
      responseType: 'blob'
    })

    // Create download link
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', 'users_export.csv')
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)

    successMessage.value = 'Users exported successfully'
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to export users'
  } finally {
    loading.value = false
  }
}

const loadUsers = async () => {
  loadingUsers.value = true
  error.value = null

  try {
    const params = new URLSearchParams()
    if (searchQuery.value) params.append('search', searchQuery.value)
    if (createdAfter.value) params.append('created_after', createdAfter.value)
    if (createdBefore.value) params.append('created_before', createdBefore.value)
    params.append('limit', limit.value.toString())
    params.append('offset', offset.value.toString())

    const response = await axios.get(`/api/admin/users/filter?${params}`)
    users.value = response.data.users || []
    totalUsers.value = response.data.count || 0
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to load users'
  } finally {
    loadingUsers.value = false
  }
}

let searchTimeout = null
const debouncedSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    offset.value = 0
    loadUsers()
  }, 300)
}

const toggleSelectAll = () => {
  if (allSelected.value) {
    selectedUserIds.value = []
  } else {
    selectedUserIds.value = users.value.map(u => u.id)
  }
}

const toggleUserSelection = (userId) => {
  const index = selectedUserIds.value.indexOf(userId)
  if (index === -1) {
    selectedUserIds.value.push(userId)
  } else {
    selectedUserIds.value.splice(index, 1)
  }
}

const prevPage = () => {
  offset.value = Math.max(0, offset.value - limit.value)
  loadUsers()
}

const nextPage = () => {
  offset.value += limit.value
  loadUsers()
}

const sendPasswordResetEmails = async () => {
  if (selectedUserIds.value.length === 0) return

  loading.value = true
  error.value = null
  resetResult.value = null

  try {
    const response = await axios.post('/api/admin/users/batch-password-reset', {
      user_ids: selectedUserIds.value
    })
    resetResult.value = response.data
    selectedUserIds.value = []
    successMessage.value = `Password reset emails sent to ${response.data.successfully_sent} users`
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to send password reset emails'
  } finally {
    loading.value = false
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString()
}

// Load users when switching to reset tab
onMounted(() => {
  if (activeTab.value === 'reset') {
    loadUsers()
  }
})

// Watch tab changes
import { watch } from 'vue'
watch(activeTab, (newTab) => {
  if (newTab === 'reset' && users.value.length === 0) {
    loadUsers()
  }
})
</script>

<style scoped>
.mobile-view-wrapper {
  overflow-y: auto;
  overflow-x: hidden;
  padding-bottom: 80px;
}

.upload-zone {
  border: 2px dashed #ccc;
  border-radius: 12px;
  padding: 40px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.upload-zone:hover {
  border-color: rgb(var(--v-theme-primary));
  background: rgba(var(--v-theme-primary), 0.05);
}

.upload-zone.drag-over {
  border-color: rgb(var(--v-theme-primary));
  background: rgba(var(--v-theme-primary), 0.1);
}

.stat-box {
  background: rgba(var(--v-theme-on-surface), 0.05);
  border-radius: 8px;
  padding: 12px;
  text-align: center;
}

.gap-2 {
  gap: 8px;
}

.gap-3 {
  gap: 12px;
}
</style>
