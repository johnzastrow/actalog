<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <!-- Success Alert -->
      <v-alert v-if="importResult" type="success" closable class="mb-3" @click:close="resetImport">
        <strong>Import Complete!</strong><br>
        <span v-if="selectedEntity === 'wodify'">
          Workouts Created: {{ importResult.workouts_created }} |
          Performances: {{ importResult.performances_created }} |
          Movements Created: {{ importResult.movements_created }} |
          WODs Created: {{ importResult.wods_created }} |
          PRs Flagged: {{ importResult.prs_flagged || 0 }}
        </span>
        <span v-else-if="selectedEntity === 'user_workouts'">
          Workouts Created: {{ importResult.created_count }} |
          Movements Auto-Created: {{ importResult.movements_created || 0 }} |
          WODs Auto-Created: {{ importResult.wods_created || 0 }}
        </span>
        <span v-else>
          Created: {{ importResult.created_count }} | Updated: {{ importResult.updated_count }} | Skipped: {{ importResult.skipped_count }}
        </span>
      </v-alert>

      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable class="mb-3" @click:close="error = null">
        {{ error }}
      </v-alert>

      <!-- Step 1: Entity Type Selection -->
      <v-card v-if="!previewResult" elevation="0" rounded="lg" class="pa-4 mb-3" bg-color="surface">
        <h2 class="text-h6 font-weight-bold mb-3" >1. Select Data Type</h2>

        <div class="d-flex flex-wrap mb-4" style="gap: 8px">
          <v-btn
            size="small"
            :color="selectedEntity === 'wods' ? 'primary' : 'default'"
            :variant="selectedEntity === 'wods' ? 'flat' : 'outlined'"
            rounded="lg"
            :elevation="selectedEntity === 'wods' ? 2 : 0"
            class="font-weight-bold"
            style="text-transform: none; padding: 8px 16px; flex: 1 1 calc(50% - 4px); min-width: 140px"
            @click="selectedEntity = 'wods'"
          >
            <v-icon start size="small">mdi-fire</v-icon>
            WODs
          </v-btn>
          <v-btn
            size="small"
            :color="selectedEntity === 'movements' ? 'primary' : 'default'"
            :variant="selectedEntity === 'movements' ? 'flat' : 'outlined'"
            rounded="lg"
            :elevation="selectedEntity === 'movements' ? 2 : 0"
            class="font-weight-bold"
            style="text-transform: none; padding: 8px 16px; flex: 1 1 calc(50% - 4px); min-width: 140px"
            @click="selectedEntity = 'movements'"
          >
            <v-icon start size="small">mdi-dumbbell</v-icon>
            Movements
          </v-btn>
          <v-btn
            size="small"
            :color="selectedEntity === 'user_workouts' ? 'primary' : 'default'"
            :variant="selectedEntity === 'user_workouts' ? 'flat' : 'outlined'"
            rounded="lg"
            :elevation="selectedEntity === 'user_workouts' ? 2 : 0"
            class="font-weight-bold"
            style="text-transform: none; padding: 8px 16px; flex: 1 1 calc(50% - 4px); min-width: 140px"
            @click="selectedEntity = 'user_workouts'"
          >
            <v-icon start size="small">mdi-clipboard-text</v-icon>
            User Workouts
          </v-btn>
          <v-btn
            size="small"
            :color="selectedEntity === 'wodify' ? 'primary' : 'default'"
            :variant="selectedEntity === 'wodify' ? 'flat' : 'outlined'"
            rounded="lg"
            :elevation="selectedEntity === 'wodify' ? 2 : 0"
            class="font-weight-bold"
            style="text-transform: none; padding: 8px 16px; flex: 1 1 calc(50% - 4px); min-width: 140px"
            @click="selectedEntity = 'wodify'"
          >
            <v-icon start size="small">mdi-file-chart</v-icon>
            Wodify
          </v-btn>
        </div>

        <v-alert type="info" density="compact" class="text-caption">
          <template v-if="selectedEntity === 'wods'">
            <strong>WODs CSV Format:</strong><br>
            name, source, type, regime, score_type, description, url, notes, is_standard, created_by_email
          </template>
          <template v-else-if="selectedEntity === 'movements'">
            <strong>Movements CSV Format:</strong><br>
            name, type, description, is_standard, created_by_email
          </template>
          <template v-else-if="selectedEntity === 'wodify'">
            <strong>Wodify Performance Export CSV:</strong><br>
            Export your performance data from Wodify. The import will automatically group performances by date and create workouts, movements, and WODs as needed.
          </template>
          <template v-else>
            <strong>User Workouts JSON Format:</strong><br>
            JSON file exported from ActaLog with full workout history, including nested movements, WODs, and performance data.
          </template>
        </v-alert>
      </v-card>

      <!-- Step 2: File Upload -->
      <v-card v-if="!previewResult" elevation="0" rounded="lg" class="pa-4 mb-3" bg-color="surface">
        <h2 class="text-h6 font-weight-bold mb-3" >
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
          <v-icon size="64" :color="selectedFile ? 'primary' : '#ccc'">
            {{ selectedFile ? 'mdi-file-check' : 'mdi-cloud-upload' }}
          </v-icon>
          <p class="text-body-1 font-weight-bold mt-3" :style="{ color: selectedFile ? 'rgb(var(--v-theme-primary))' : 'rgb(var(--v-theme-on-surface))' }">
            {{ selectedFile ? selectedFile.name : `Drop ${fileTypeLabel} file here or click to browse` }}
          </p>
          <p v-if="!selectedFile" class="text-caption text-disabled">
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
          color="primary"
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
        <v-card elevation="0" rounded="lg" class="pa-4 mb-3" bg-color="surface">
          <h2 class="text-h6 font-weight-bold mb-3" >3. Preview Results</h2>

          <!-- Wodify Import Stats -->
          <v-row v-if="isWodifyImport" dense>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption text-medium-emphasis">Total Rows</p>
                <p class="text-h6 font-weight-bold" >
                  {{ previewResult.total_rows }}
                </p>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption text-medium-emphasis">Valid Rows</p>
                <p class="text-h6 font-weight-bold" style="color: rgb(var(--v-theme-success))">
                  {{ previewResult.valid_rows }}
                </p>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption text-medium-emphasis">Workout Dates</p>
                <p class="text-h6 font-weight-bold" style="color: rgb(var(--v-theme-primary))">
                  {{ previewResult.unique_workout_dates }}
                </p>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption text-medium-emphasis">Workouts to Create</p>
                <p class="text-h6 font-weight-bold" style="color: rgb(var(--v-theme-primary))">
                  {{ previewResult.user_workouts_to_create }}
                </p>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption text-medium-emphasis">New Movements</p>
                <p class="text-h6 font-weight-bold" style="color: #ff9800">
                  {{ previewResult.movements_to_create }}
                </p>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption text-medium-emphasis">New WODs</p>
                <p class="text-h6 font-weight-bold" style="color: #ff9800">
                  {{ previewResult.wods_to_create }}
                </p>
              </div>
            </v-col>
          </v-row>

          <!-- Standard Import Stats -->
          <v-row v-else dense>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption text-medium-emphasis">
                  {{ selectedEntity === 'user_workouts' ? 'Total Workouts' : 'Total Rows' }}
                </p>
                <p class="text-h6 font-weight-bold" >
                  {{ previewResult.total_workouts || previewResult.total_rows }}
                </p>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption text-medium-emphasis">
                  {{ selectedEntity === 'user_workouts' ? 'Valid Workouts' : 'Valid Rows' }}
                </p>
                <p class="text-h6 font-weight-bold" style="color: rgb(var(--v-theme-success))">
                  {{ previewResult.valid_workouts || previewResult.valid_rows }}
                </p>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption text-medium-emphasis">
                  {{ selectedEntity === 'user_workouts' ? 'Invalid Workouts' : 'Invalid Rows' }}
                </p>
                <p class="text-h6 font-weight-bold" style="color: #f44336">
                  {{ previewResult.invalid_workouts || previewResult.invalid_rows }}
                </p>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="stat-box">
                <p class="text-caption text-medium-emphasis">Duplicates</p>
                <p class="text-h6 font-weight-bold" style="color: #ff9800">
                  {{ previewResult.duplicate_workouts || previewResult.duplicate_rows }}
                </p>
              </div>
            </v-col>
          </v-row>
        </v-card>

        <!-- Import Options -->
        <v-card v-if="selectedEntity !== 'user_workouts' && selectedEntity !== 'wodify'" elevation="0" rounded="lg" class="pa-4 mb-3" bg-color="surface">
          <h3 class="text-body-1 font-weight-bold mb-3" >Import Options</h3>

          <v-radio-group v-model="duplicateHandling" density="compact">
            <v-radio
              value="skip"
              label="Skip duplicates (only import new records)"
              color="primary"
            />
            <v-radio
              value="update"
              label="Update duplicates (overwrite existing records)"
              color="primary"
            />
            <v-radio
              value="cancel"
              label="Cancel import if duplicates found"
              color="primary"
            />
          </v-radio-group>
        </v-card>

        <!-- User Workouts Import Options -->
        <v-card v-if="selectedEntity === 'user_workouts'" elevation="0" rounded="lg" class="pa-4 mb-3" bg-color="surface">
          <h3 class="text-body-1 font-weight-bold mb-3" >Import Options</h3>

          <v-checkbox
            v-model="skipDuplicates"
            label="Skip duplicate workouts (based on workout date and WOD)"
            color="primary"
            density="compact"
            hide-details
          />

          <v-alert type="info" density="compact" class="text-caption mt-3">
            Missing movements and WODs will be automatically created during import.
          </v-alert>
        </v-card>

        <!-- Wodify New Entities -->
        <v-card v-if="isWodifyImport && (previewResult.new_movements?.length > 0 || previewResult.new_wods?.length > 0)" elevation="0" rounded="lg" class="pa-4 mb-3" bg-color="surface">
          <h3 class="text-body-1 font-weight-bold mb-3" >New Entities to Create</h3>

          <div v-if="previewResult.new_movements?.length > 0" class="mb-3">
            <p class="text-caption font-weight-bold mb-2 text-medium-emphasis">Movements ({{ previewResult.new_movements.length }})</p>
            <v-chip v-for="(movement, idx) in previewResult.new_movements" :key="'movement-' + idx" size="small" class="ma-1" color="primary">
              {{ movement }}
            </v-chip>
          </div>

          <div v-if="previewResult.new_wods?.length > 0">
            <p class="text-caption font-weight-bold mb-2 text-medium-emphasis">WODs ({{ previewResult.new_wods.length }})</p>
            <v-chip v-for="(wod, idx) in previewResult.new_wods" :key="'wod-' + idx" size="small" class="ma-1" color="warning">
              {{ wod }}
            </v-chip>
          </div>
        </v-card>

        <!-- Wodify Workout Summary -->
        <v-card v-if="isWodifyImport && previewResult.workout_summary?.length > 0" elevation="0" rounded="lg" class="pa-4 mb-3" style="background-color: rgb(var(--v-theme-surface)); overflow-x: auto">
          <h3 class="text-body-1 font-weight-bold mb-3" >Workout Summary</h3>

          <v-data-table
            :headers="[
              { title: 'Date', key: 'date' },
              { title: 'Movements', key: 'movement_count' },
              { title: 'WODs', key: 'wod_count' },
              { title: 'Component Types', key: 'component_types' },
              { title: 'Has PRs', key: 'has_prs' }
            ]"
            :items="previewResult.workout_summary"
            density="compact"
            class="preview-table"
            :items-per-page="10"
          >
            <template #item.has_prs="{ item }">
              <v-icon v-if="item.has_prs" color="gold" size="small">mdi-trophy</v-icon>
              <span v-else class="text-caption text-disabled">—</span>
            </template>
          </v-data-table>

          <p v-if="previewResult.workout_summary.length > 10" class="text-caption text-center mt-2 text-disabled">
            Showing first 10 workouts of {{ previewResult.workout_summary.length }}
          </p>
        </v-card>

        <!-- Standard Preview Table -->
        <v-card v-if="!isWodifyImport" elevation="0" rounded="lg" class="pa-4 mb-3" style="background-color: rgb(var(--v-theme-surface)); overflow-x: auto">
          <h3 class="text-body-1 font-weight-bold mb-3" >Data Preview</h3>

          <v-data-table
            :headers="previewHeaders"
            :items="previewData"
            density="compact"
            class="preview-table"
            :item-value="selectedEntity === 'user_workouts' ? 'workout_number' : 'row_number'"
            :items-per-page="10"
            :items-per-page-options="[10, 25, 50, 100]"
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
        </v-card>

        <!-- Action Buttons -->
        <v-row dense>
          <v-col cols="6">
            <v-btn
              block
              size="large"
              
              color="medium-emphasis"
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
              color="success"
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
  </div>
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

const isWodifyImport = computed(() => {
  return selectedEntity.value === 'wodify'
})

const previewData = computed(() => {
  if (!previewResult.value) return []

  // For user workouts, use workouts array; for CSV, use rows array
  return previewResult.value.workouts || previewResult.value.rows || []
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
    } else if (selectedEntity.value === 'wodify') {
      endpoint = '/api/import/wodify/preview'
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
    } else if (selectedEntity.value === 'wodify') {
      endpoint = '/api/import/wodify/confirm'
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
  border-color: rgb(var(--v-theme-primary));
  background-color: rgb(var(--v-theme-surface));
}

.stat-box {
  padding: 12px;
  border-radius: 8px;
  background-color: rgb(var(--v-theme-background));
  text-align: center;
}

.preview-table {
  font-size: 12px;
}

.text-red {
  color: #f44336;
}
</style>
