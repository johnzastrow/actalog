<template>
  <v-container fluid class="pa-4">
    <!-- Header -->
    <div class="d-flex align-center mb-4">
      <v-btn icon @click="$router.back()" class="mr-2">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <div>
        <h1 class="text-h5">Database Backups</h1>
        <div class="text-body-2 text-medium-emphasis">Create, download, and restore database backups</div>
      </div>
      <v-spacer />
      <v-btn
        color="primary"
        prepend-icon="mdi-database-export"
        :loading="creating"
        :disabled="creating || loading"
        @click="createBackup"
      >
        Create Backup
      </v-btn>
    </div>

    <!-- Loading State -->
    <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />

    <!-- Error Alert -->
    <v-alert v-if="error" type="error" variant="tonal" closable @click:close="error = null" class="mb-4">
      {{ error }}
    </v-alert>

    <!-- Success Alert -->
    <v-alert v-if="successMessage" type="success" variant="tonal" closable @click:close="successMessage = null" class="mb-4">
      {{ successMessage }}
    </v-alert>

    <!-- Backups Table -->
    <v-card elevation="0" rounded="lg">
      <v-card-title class="d-flex align-center">
        <v-icon class="mr-2">mdi-database</v-icon>
        Available Backups ({{ backups.length }})
      </v-card-title>

      <v-divider />

      <!-- Empty State -->
      <v-card-text v-if="!loading && backups.length === 0" class="text-center py-12">
        <v-icon size="64" color="grey-lighten-1" class="mb-4">mdi-database-off</v-icon>
        <div class="text-h6 mb-2">No Backups Found</div>
        <div class="text-body-2 text-medium-emphasis mb-4">
          Create your first backup to protect your data
        </div>
        <v-btn color="primary" prepend-icon="mdi-database-export" @click="createBackup">
          Create Backup
        </v-btn>
      </v-card-text>

      <!-- Backups List -->
      <v-data-table
        v-else
        :headers="headers"
        :items="backups"
        :loading="loading"
        :items-per-page="20"
        hide-default-footer
        class="elevation-0"
      >
        <!-- Filename Column -->
        <template #item.filename="{ item }">
          <div class="d-flex align-center">
            <v-icon class="mr-2" color="primary">mdi-file-document</v-icon>
            <div>
              <div class="font-weight-medium">{{ item.filename }}</div>
              <div class="text-caption text-medium-emphasis">{{ formatFileSize(item.file_size) }}</div>
            </div>
          </div>
        </template>

        <!-- Created At Column -->
        <template #item.created_at="{ item }">
          <div>
            <div class="text-body-2">{{ formatDate(item.created_at) }}</div>
            <div class="text-caption text-medium-emphasis">{{ formatTime(item.created_at) }}</div>
          </div>
        </template>

        <!-- Created By Column -->
        <template #item.created_by_email="{ item }">
          <div class="d-flex align-center">
            <v-avatar size="24" color="primary" class="mr-2">
              <span class="text-white text-caption">{{ item.created_by_email[0].toUpperCase() }}</span>
            </v-avatar>
            <div class="text-body-2">{{ item.created_by_email }}</div>
          </div>
        </template>

        <!-- Stats Column -->
        <template #item.stats="{ item }">
          <div class="text-body-2">
            <div>Users: {{ item.total_users }}</div>
            <div>Workouts: {{ item.total_workouts }}</div>
            <div class="text-caption text-medium-emphasis">
              v{{ item.version }}
            </div>
          </div>
        </template>

        <!-- Actions Column -->
        <template #item.actions="{ item }">
          <div class="d-flex gap-1">
            <!-- Download Button -->
            <v-tooltip text="Download Backup" location="top">
              <template #activator="{ props }">
                <v-btn
                  v-bind="props"
                  icon
                  size="small"
                  variant="text"
                  @click="downloadBackup(item)"
                >
                  <v-icon color="primary">mdi-download</v-icon>
                </v-btn>
              </template>
            </v-tooltip>

            <!-- Restore Button -->
            <v-tooltip text="Restore Backup" location="top">
              <template #activator="{ props }">
                <v-btn
                  v-bind="props"
                  icon
                  size="small"
                  variant="text"
                  @click="showRestoreDialog(item)"
                >
                  <v-icon color="warning">mdi-database-import</v-icon>
                </v-btn>
              </template>
            </v-tooltip>

            <!-- Delete Button -->
            <v-tooltip text="Delete Backup" location="top">
              <template #activator="{ props }">
                <v-btn
                  v-bind="props"
                  icon
                  size="small"
                  variant="text"
                  @click="showDeleteDialog(item)"
                >
                  <v-icon color="error">mdi-delete</v-icon>
                </v-btn>
              </template>
            </v-tooltip>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <!-- Restore Confirmation Dialog -->
    <v-dialog v-model="restoreDialog" max-width="500">
      <v-card>
        <v-card-title class="text-h6 d-flex align-center">
          <v-icon color="warning" class="mr-2">mdi-alert</v-icon>
          Confirm Restore
        </v-card-title>
        <v-divider />
        <v-card-text class="pt-4">
          <v-alert type="warning" variant="tonal" class="mb-4">
            <div class="font-weight-bold mb-2">WARNING: This action is irreversible!</div>
            <div>Restoring this backup will:</div>
            <ul class="mt-2">
              <li>Delete ALL current data</li>
              <li>Replace with data from: <strong>{{ selectedBackup?.filename }}</strong></li>
              <li>Created: <strong>{{ formatDate(selectedBackup?.created_at) }}</strong></li>
              <li>Contains: <strong>{{ selectedBackup?.total_users }} users, {{ selectedBackup?.total_workouts }} workouts</strong></li>
            </ul>
          </v-alert>
          <div class="text-body-2 text-medium-emphasis">
            Are you absolutely sure you want to proceed? All users will need to log in again after the restore.
          </div>
        </v-card-text>
        <v-divider />
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="restoreDialog = false">Cancel</v-btn>
          <v-btn
            color="warning"
            variant="flat"
            :loading="restoring"
            @click="confirmRestore"
          >
            Restore Backup
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="500">
      <v-card>
        <v-card-title class="text-h6 d-flex align-center">
          <v-icon color="error" class="mr-2">mdi-delete</v-icon>
          Confirm Delete
        </v-card-title>
        <v-divider />
        <v-card-text class="pt-4">
          <div class="text-body-2 mb-4">
            Are you sure you want to delete this backup?
          </div>
          <v-alert type="info" variant="tonal">
            <div class="font-weight-bold">{{ selectedBackup?.filename }}</div>
            <div class="text-caption mt-1">
              Created: {{ formatDate(selectedBackup?.created_at) }} by {{ selectedBackup?.created_by_email }}
            </div>
            <div class="text-caption">
              Size: {{ formatFileSize(selectedBackup?.file_size) }}
            </div>
          </v-alert>
        </v-card-text>
        <v-divider />
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn
            color="error"
            variant="flat"
            :loading="deleting"
            @click="confirmDelete"
          >
            Delete Backup
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
const creating = ref(false)
const restoring = ref(false)
const deleting = ref(false)
const error = ref(null)
const successMessage = ref(null)
const backups = ref([])
const restoreDialog = ref(false)
const deleteDialog = ref(false)
const selectedBackup = ref(null)

// Table headers
const headers = [
  { title: 'Filename', key: 'filename', sortable: true },
  { title: 'Created', key: 'created_at', sortable: true },
  { title: 'Created By', key: 'created_by_email', sortable: true },
  { title: 'Stats', key: 'stats', sortable: false },
  { title: 'Actions', key: 'actions', sortable: false, align: 'end' }
]

// Fetch backups from API
const fetchBackups = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await axios.get('/api/admin/backups')
    backups.value = response.data.backups || []
  } catch (err) {
    console.error('Failed to fetch backups:', err)
    error.value = err.response?.data?.error || 'Failed to load backups'
  } finally {
    loading.value = false
  }
}

// Create new backup
const createBackup = async () => {
  creating.value = true
  error.value = null
  successMessage.value = null

  try {
    const response = await axios.post('/api/admin/backups')
    successMessage.value = `Backup created successfully: ${response.data.filename}`
    await fetchBackups()
  } catch (err) {
    console.error('Failed to create backup:', err)
    error.value = err.response?.data?.error || 'Failed to create backup'
  } finally {
    creating.value = false
  }
}

// Download backup
const downloadBackup = async (backup) => {
  try {
    const response = await axios.get(`/api/admin/backups/${backup.filename}`, {
      responseType: 'blob'
    })

    // Create download link
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', backup.filename)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    successMessage.value = `Backup downloaded: ${backup.filename}`
  } catch (err) {
    console.error('Failed to download backup:', err)
    error.value = err.response?.data?.error || 'Failed to download backup'
  }
}

// Show restore dialog
const showRestoreDialog = (backup) => {
  selectedBackup.value = backup
  restoreDialog.value = true
}

// Confirm restore
const confirmRestore = async () => {
  restoring.value = true
  error.value = null
  successMessage.value = null

  try {
    await axios.post(`/api/admin/backups/${selectedBackup.value.filename}/restore`, {
      confirm: true
    })

    successMessage.value = 'Backup restored successfully! Please log in again.'
    restoreDialog.value = false
    selectedBackup.value = null

    // Wait a moment, then redirect to login
    setTimeout(() => {
      window.location.href = '/login'
    }, 2000)
  } catch (err) {
    console.error('Failed to restore backup:', err)
    error.value = err.response?.data?.error || 'Failed to restore backup'
  } finally {
    restoring.value = false
  }
}

// Show delete dialog
const showDeleteDialog = (backup) => {
  selectedBackup.value = backup
  deleteDialog.value = true
}

// Confirm delete
const confirmDelete = async () => {
  deleting.value = true
  error.value = null
  successMessage.value = null

  try {
    await axios.delete(`/api/admin/backups/${selectedBackup.value.filename}`)
    successMessage.value = `Backup deleted: ${selectedBackup.value.filename}`
    deleteDialog.value = false
    selectedBackup.value = null
    await fetchBackups()
  } catch (err) {
    console.error('Failed to delete backup:', err)
    error.value = err.response?.data?.error || 'Failed to delete backup'
  } finally {
    deleting.value = false
  }
}

// Fetch backups on mount
onMounted(() => {
  fetchBackups()
})

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

function formatFileSize(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}
</script>

<style scoped>
.gap-1 {
  gap: 4px;
}

.gap-2 {
  gap: 8px;
}
</style>
