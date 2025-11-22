<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" density="compact" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-toolbar-title class="text-white font-weight-bold">Import Data</v-toolbar-title>
      <v-spacer />
      <v-btn icon="mdi-close" color="white" size="small" @click="$router.back()" />
    </v-app-bar>

    <v-container class="px-3 pb-1 pt-0" style="margin-top: 48px; margin-bottom: 70px">
      <!-- Success Alert -->
      <v-alert v-if="importResult" type="success" closable @click:close="resetImport" class="mb-3">
        <strong>Import Complete!</strong><br>
        <span v-if="selectedEntity === 'user_workouts'">
          Workouts Created: {{ importResult.created_count }} |
          Movements Auto-Created: {{ importResult.movements_created || 0 }} |
          WODs Auto-Created: {{ importResult.wods_created || 0 }}
        </span>
        <span v-else>
          Created: {{ importResult.created_count }} | Updated: {{ importResult.updated_count }} | Skipped: {{ importResult.skipped_count }}
        </span>
      </v-alert>

      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable @click:close="error = null" class="mb-3">
        {{ error }}
      </v-alert>

      <!-- Step 1: Entity Type Selection -->
      <v-card v-if="!previewResult" elevation="0" rounded="lg" class="pa-4 mb-3" style="background: white">
        <h2 class="text-h6 font-weight-bold mb-3" style="color: #1a1a1a">1. Select Data Type</h2>

        <v-btn-toggle
          v-model="selectedEntity"
          color="#00bcd4"
          variant="outlined"
          divided
          mandatory
          class="mb-4 d-flex flex-wrap"
        >
          <v-btn value="wods" prepend-icon="mdi-fire">WODs</v-btn>
          <v-btn value="movements" prepend-icon="mdi-dumbbell">Movements</v-btn>
          <v-btn value="user_workouts" prepend-icon="mdi-clipboard-text">User Workouts</v-btn>
        </v-btn-toggle>

        <v-alert type="info" density="compact" class="text-caption">
          <template v-if="selectedEntity === 'wods'">
            <strong>WODs CSV Format:</strong><br>
            name, source, type, regime, score_type, description, url, notes, is_standard, created_by_email
          </template>
          <template v-else-if="selectedEntity === 'movements'">
            <strong>Movements CSV Format:</strong><br>
            name, type, description, is_standard, created_by_email
          </template>
          <template v-else>
            <strong>User Workouts JSON Format:</strong><br>
            JSON file exported from ActaLog with full workout history, including nested movements, WODs, and performance data.
          </template>
        </v-alert>
      </v-card>

      <!-- Step 2: File Upload -->
      <v-card v-if="!previewResult" elevation="0" rounded="lg" class="pa-4 mb-3" style="background: white">
        <h2 class="text-h6 font-weight-bold mb-3" style="color: #1a1a1a">
          2. Upload {{ selectedEntity === 'user_workouts' ? 'JSON' : 'CSV' }} File
        </h2>

        <div
          class="upload-zone"
          :class="{ 'drag-over': dragOver }"
          @drop.prevent="handleDrop"
          @dragover.prevent="dragOver = true"
          @dragleave.prevent="dragOver = false"
          @click="$refs.fileInput.click()"
        >
          <v-icon size="64" :color="selectedFile ? '#00bcd4' : '#ccc'">
            {{ selectedFile ? 'mdi-file-check' : 'mdi-cloud-upload' }}
          </v-icon>
          <p class="text-body-1 font-weight-bold mt-3" :style="{ color: selectedFile ? '#00bcd4' : '#1a1a1a' }">
            {{ selectedFile ? selectedFile.name : `Drop ${fileTypeLabel} file here or click to browse` }}
          </p>
          <p v-if="!selectedFile" class="text-caption" style="color: #999">
            Maximum file size: 10MB
          </p>
          <input
            ref="fileInput"
            type="file"
            :accept="fileAccept"
            style="display: none"
            @change="handleFileSelect"
          >
        </div>

        <v-btn
          v-if="selectedFile"
          block
          size="large"
          color="#00bcd4"
          rounded="lg"
          elevation="2"
          :loading="uploading"
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
        <!-- Summary Card -->
        <v-card elevation="0" rounded="lg" class="pa-4 mb-3" style="background: white">
          <h2 class="text-h6 font-weight-bold mb-3" style="color: #1a1a1a">3. Preview Results</h2>

          <v-row dense>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption" style="color: #666">
                  {{ selectedEntity === 'user_workouts' ? 'Total Workouts' : 'Total Rows' }}
                </p>
                <p class="text-h6 font-weight-bold" style="color: #1a1a1a">
                  {{ previewResult.total_workouts || previewResult.total_rows }}
                </p>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption" style="color: #666">
                  {{ selectedEntity === 'user_workouts' ? 'Valid Workouts' : 'Valid Rows' }}
                </p>
                <p class="text-h6 font-weight-bold" style="color: #4caf50">
                  {{ previewResult.valid_workouts || previewResult.valid_rows }}
                </p>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption" style="color: #666">
                  {{ selectedEntity === 'user_workouts' ? 'Invalid Workouts' : 'Invalid Rows' }}
                </p>
                <p class="text-h6 font-weight-bold" style="color: #f44336">
                  {{ previewResult.invalid_workouts || previewResult.invalid_rows }}
                </p>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption" style="color: #666">Duplicates</p>
                <p class="text-h6 font-weight-bold" style="color: #ff9800">
                  {{ previewResult.duplicate_workouts || previewResult.duplicate_rows }}
                </p>
              </div>
            </v-col>
          </v-row>
        </v-card>

        <!-- Import Options -->
        <v-card v-if="selectedEntity !== 'user_workouts'" elevation="0" rounded="lg" class="pa-4 mb-3" style="background: white">
          <h3 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Import Options</h3>

          <v-radio-group v-model="duplicateHandling" density="compact">
            <v-radio
              value="skip"
              label="Skip duplicates (only import new records)"
              color="#00bcd4"
            />
            <v-radio
              value="update"
              label="Update duplicates (overwrite existing records)"
              color="#00bcd4"
            />
            <v-radio
              value="cancel"
              label="Cancel import if duplicates found"
              color="#00bcd4"
            />
          </v-radio-group>
        </v-card>

        <!-- User Workouts Import Options -->
        <v-card v-if="selectedEntity === 'user_workouts'" elevation="0" rounded="lg" class="pa-4 mb-3" style="background: white">
          <h3 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Import Options</h3>

          <v-checkbox
            v-model="skipDuplicates"
            label="Skip duplicate workouts (based on workout date and WOD)"
            color="#00bcd4"
            density="compact"
            hide-details
          />

          <v-alert type="info" density="compact" class="text-caption mt-3">
            Missing movements and WODs will be automatically created during import.
          </v-alert>
        </v-card>

        <!-- Preview Table -->
        <v-card elevation="0" rounded="lg" class="pa-4 mb-3" style="background: white; overflow-x: auto">
          <h3 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Data Preview</h3>

          <v-data-table
            :headers="previewHeaders"
            :items="previewData"
            density="compact"
            class="preview-table"
            :item-value="selectedEntity === 'user_workouts' ? 'workout_number' : 'row_number'"
          >
            <template #item.row_number="{ item }">
              <v-chip size="x-small" :color="getRowColor(item)">
                {{ item.row_number }}
              </v-chip>
            </template>
            <template #item.workout_number="{ item }">
              <v-chip size="x-small" :color="getRowColor(item)">
                {{ item.workout_number }}
              </v-chip>
            </template>
            <template #item.name="{ item }">
              <span :class="{ 'text-red': !item.is_valid }">{{ item.name }}</span>
            </template>
            <template #item.errors="{ item }">
              <v-chip v-if="item.is_duplicate" size="x-small" color="orange">Duplicate</v-chip>
              <v-tooltip v-if="item.errors && item.errors.length > 0" location="top">
                <template #activator="{ props }">
                  <v-chip v-bind="props" size="x-small" color="error">
                    {{ item.errors.length }} error(s)
                  </v-chip>
                </template>
                <div v-for="(err, idx) in item.errors" :key="idx">• {{ err }}</div>
              </v-tooltip>
            </template>
          </v-data-table>

          <p v-if="previewData.length > 10" class="text-caption text-center mt-2" style="color: #999">
            Showing first 10 {{ selectedEntity === 'user_workouts' ? 'workouts' : 'rows' }} of {{ previewData.length }}
          </p>
        </v-card>

        <!-- Action Buttons -->
        <v-row dense>
          <v-col cols="6">
            <v-btn
              block
              size="large"
              variant="outlined"
              color="#666"
              rounded="lg"
              class="font-weight-bold"
              style="text-transform: none"
              @click="resetImport"
            >
              Cancel
            </v-btn>
          </v-col>
          <v-col cols="6">
            <v-btn
              block
              size="large"
              color="#4caf50"
              rounded="lg"
              elevation="2"
              :disabled="(previewResult.valid_workouts || previewResult.valid_rows || 0) === 0"
              :loading="confirming"
              class="font-weight-bold"
              style="text-transform: none"
              @click="confirmImport"
            >
              <v-icon start>mdi-check</v-icon>
              Confirm Import
            </v-btn>
          </v-col>
        </v-row>
      </template>
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
const selectedEntity = ref('wods')
const selectedFile = ref(null)
const dragOver = ref(false)
const uploading = ref(false)
const confirming = ref(false)
const previewResult = ref(null)
const importResult = ref(null)
const error = ref(null)
const duplicateHandling = ref('skip')
const skipDuplicates = ref(true)

// Computed
const fileAccept = computed(() => {
  return selectedEntity.value === 'user_workouts' ? '.json' : '.csv'
})

const fileTypeLabel = computed(() => {
  return selectedEntity.value === 'user_workouts' ? 'JSON' : 'CSV'
})

const previewData = computed(() => {
  if (!previewResult.value) return []

  // For user workouts, use workouts array; for CSV, use rows array
  const data = previewResult.value.workouts || previewResult.value.rows || []
  return data.slice(0, 10)
})

const previewHeaders = computed(() => {
  if (selectedEntity.value === 'wods') {
    return [
      { title: '#', key: 'row_number', width: 50 },
      { title: 'Name', key: 'name' },
      { title: 'Type', key: 'type' },
      { title: 'Regime', key: 'regime' },
      { title: 'Status', key: 'errors', width: 120 }
    ]
  } else if (selectedEntity.value === 'movements') {
    return [
      { title: '#', key: 'row_number', width: 50 },
      { title: 'Name', key: 'name' },
      { title: 'Type', key: 'type' },
      { title: 'Status', key: 'errors', width: 120 }
    ]
  } else {
    // user_workouts
    return [
      { title: '#', key: 'workout_number', width: 50 },
      { title: 'Date', key: 'workout_date' },
      { title: 'WOD', key: 'wod_name' },
      { title: 'Movements', key: 'movement_count' },
      { title: 'Status', key: 'errors', width: 120 }
    ]
  }
})

// Methods
const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (file) {
    selectedFile.value = file
  }
}

const handleDrop = (event) => {
  dragOver.value = false
  const file = event.dataTransfer.files[0]
  const expectedExtension = selectedEntity.value === 'user_workouts' ? '.json' : '.csv'

  if (file && file.name.endsWith(expectedExtension)) {
    selectedFile.value = file
  } else {
    error.value = `Please upload a valid ${fileTypeLabel.value} file`
  }
}

const previewImport = async () => {
  if (!selectedFile.value) return

  uploading.value = true
  error.value = null
  previewResult.value = null

  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)

    let endpoint
    if (selectedEntity.value === 'wods') {
      endpoint = '/api/import/wods/preview'
    } else if (selectedEntity.value === 'movements') {
      endpoint = '/api/import/movements/preview'
    } else {
      endpoint = '/api/import/user-workouts/preview'
    }

    const response = await axios.post(endpoint, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })

    previewResult.value = response.data
  } catch (err) {
    console.error('Preview error:', err)
    error.value = err.response?.data?.error || `Failed to preview import. Please check your ${fileTypeLabel.value} format.`
  } finally {
    uploading.value = false
  }
}

const confirmImport = async () => {
  if (!selectedFile.value || !previewResult.value) return

  confirming.value = true
  error.value = null

  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)

    let endpoint
    if (selectedEntity.value === 'wods') {
      endpoint = '/api/import/wods/confirm'
      formData.append('skip_duplicates', duplicateHandling.value === 'skip')
      formData.append('update_duplicates', duplicateHandling.value === 'update')
    } else if (selectedEntity.value === 'movements') {
      endpoint = '/api/import/movements/confirm'
      formData.append('skip_duplicates', duplicateHandling.value === 'skip')
      formData.append('update_duplicates', duplicateHandling.value === 'update')
    } else {
      endpoint = '/api/import/user-workouts/confirm'
      formData.append('skip_duplicates', skipDuplicates.value)
    }

    const response = await axios.post(endpoint, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })

    importResult.value = response.data
    previewResult.value = null
    selectedFile.value = null
  } catch (err) {
    console.error('Import error:', err)
    error.value = err.response?.data?.error || 'Failed to import data. Please try again.'
  } finally {
    confirming.value = false
  }
}

const resetImport = () => {
  previewResult.value = null
  importResult.value = null
  selectedFile.value = null
  error.value = null
  duplicateHandling.value = 'skip'
}

const getRowColor = (item) => {
  if (!item.is_valid) return 'error'
  if (item.is_duplicate) return 'warning'
  return 'success'
}
</script>

<style scoped>
.upload-zone {
  border: 2px dashed #ccc;
  border-radius: 12px;
  padding: 48px 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  background: #fafafa;
}

.upload-zone:hover, .upload-zone.drag-over {
  border-color: #00bcd4;
  background: #e0f7fa;
}

.stat-box {
  padding: 12px;
  border-radius: 8px;
  background: #f5f7fa;
  text-align: center;
}

.preview-table {
  font-size: 12px;
}

.text-red {
  color: #f44336;
}
</style>
