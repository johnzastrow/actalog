<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <!-- Page Title -->
      <div class="mb-4">
        <h1 style="color: #2c3e50; font-size: 24px; font-weight: 600">
          <v-icon color="primary" size="28" class="mr-2">mdi-database-check</v-icon>
          Data Quality Checks
        </h1>
        <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px">
          Scan for duplicate records and data quality issues across the database
        </p>
      </div>

      <!-- Tabs -->
      <v-tabs v-model="activeTab" bg-color="transparent" color="primary" class="mb-4">
        <v-tab value="overview">
          <v-icon start>mdi-view-dashboard</v-icon>
          Overview
        </v-tab>
        <v-tab value="duplicates">
          <v-icon start>mdi-content-copy</v-icon>
          Duplicates
        </v-tab>
        <v-tab value="issues">
          <v-icon start>mdi-alert-circle</v-icon>
          Data Issues
        </v-tab>
      </v-tabs>

      <v-window v-model="activeTab">
        <!-- Overview Tab -->
        <v-window-item value="overview">
          <v-card elevation="0" rounded="lg" class="pa-4 mb-4">
            <div class="d-flex align-center mb-4">
              <h3 style="color: #2c3e50; font-size: 18px; font-weight: 600">Full Database Scan</h3>
              <v-spacer></v-spacer>
              <v-btn
                color="primary"
                :loading="scanning"
                :disabled="scanning"
                prepend-icon="mdi-magnify-scan"
                @click="runFullScan"
              >
                Run Full Scan
              </v-btn>
            </div>

            <v-alert v-if="!scanned" type="info" variant="tonal" class="mb-4">
              Click "Run Full Scan" to check for duplicates and data quality issues.
            </v-alert>

            <!-- Summary Cards -->
            <v-row v-if="scanned">
              <v-col cols="12" md="6">
                <v-card elevation="1" rounded="lg" class="pa-4" :color="totalDuplicates > 0 ? 'warning' : 'success'">
                  <div class="d-flex align-center">
                    <v-icon :color="totalDuplicates > 0 ? 'warning-darken-2' : 'success-darken-2'" size="40" class="mr-3">
                      {{ totalDuplicates > 0 ? 'mdi-content-copy' : 'mdi-check-circle' }}
                    </v-icon>
                    <div>
                      <div style="font-size: 28px; font-weight: 700">{{ totalDuplicates }}</div>
                      <div style="font-size: 14px; opacity: 0.8">Duplicate Records</div>
                    </div>
                  </div>
                </v-card>
              </v-col>
              <v-col cols="12" md="6">
                <v-card elevation="1" rounded="lg" class="pa-4" :color="totalIssues > 0 ? 'error' : 'success'">
                  <div class="d-flex align-center">
                    <v-icon :color="totalIssues > 0 ? 'error-darken-2' : 'success-darken-2'" size="40" class="mr-3">
                      {{ totalIssues > 0 ? 'mdi-alert-circle' : 'mdi-check-circle' }}
                    </v-icon>
                    <div>
                      <div style="font-size: 28px; font-weight: 700">{{ totalIssues }}</div>
                      <div style="font-size: 14px; opacity: 0.8">Data Quality Issues</div>
                    </div>
                  </div>
                </v-card>
              </v-col>
            </v-row>

            <!-- Duplicate Summary by Entity -->
            <div v-if="scanned && duplicateSummary" class="mt-4">
              <h4 style="color: #2c3e50; font-size: 16px; font-weight: 600; margin-bottom: 12px">
                Duplicates by Entity Type
              </h4>
              <v-row dense>
                <v-col v-for="(data, entity) in duplicateSummary.by_entity" :key="entity" cols="12" sm="6" md="4">
                  <v-card
                    elevation="0"
                    rounded="lg"
                    class="pa-3"
                    style="border: 1px solid rgba(0,0,0,0.12)"
                    :class="{ 'entity-card-warning': data.total_duplicates > 0 }"
                  >
                    <div class="d-flex align-center justify-space-between">
                      <div>
                        <div style="font-size: 12px; text-transform: uppercase; opacity: 0.7">{{ formatEntityName(entity) }}</div>
                        <div style="font-size: 20px; font-weight: 600">{{ data.total_duplicates }}</div>
                        <div style="font-size: 11px; opacity: 0.6">{{ (data.groups || []).length }} groups</div>
                      </div>
                      <v-btn
                        v-if="data.total_duplicates > 0"
                        variant="text"
                        color="primary"
                        size="small"
                        @click="viewEntityDuplicates(entity)"
                      >
                        View
                      </v-btn>
                    </div>
                  </v-card>
                </v-col>
              </v-row>
            </div>

            <!-- Issue Summary by Type -->
            <div v-if="scanned" class="mt-4">
              <h4 style="color: #2c3e50; font-size: 16px; font-weight: 600; margin-bottom: 12px">
                Data Quality Checks
              </h4>
              <v-row dense>
                <v-col v-for="checkType in qualityCheckTypes" :key="checkType.value" cols="12" sm="6" md="6" lg="3">
                  <v-card
                    elevation="0"
                    rounded="lg"
                    class="pa-3"
                    style="border: 1px solid rgba(0,0,0,0.12)"
                    :class="{
                      'entity-card-error': getIssueCount(checkType.value) > 0 && checkType.severity === 'error',
                      'entity-card-warning': getIssueCount(checkType.value) > 0 && checkType.severity === 'warning'
                    }"
                  >
                    <div class="d-flex align-center mb-2">
                      <v-icon
                        :color="getIssueCount(checkType.value) > 0 ? checkType.severity : 'grey'"
                        size="24"
                        class="mr-2"
                      >
                        {{ checkType.icon }}
                      </v-icon>
                      <div style="font-size: 13px; font-weight: 600">{{ checkType.label }}</div>
                    </div>
                    <div style="font-size: 11px; opacity: 0.6; margin-bottom: 8px">{{ checkType.description }}</div>
                    <div class="d-flex align-center justify-space-between">
                      <div style="font-size: 24px; font-weight: 700">{{ getIssueCount(checkType.value) }}</div>
                      <v-btn
                        v-if="getIssueCount(checkType.value) > 0"
                        variant="text"
                        :color="checkType.severity"
                        size="small"
                        @click="viewIssuesByType(checkType.value)"
                      >
                        View
                      </v-btn>
                      <v-icon v-else color="success" size="20">mdi-check-circle</v-icon>
                    </div>
                  </v-card>
                </v-col>
              </v-row>
            </div>
          </v-card>
        </v-window-item>

        <!-- Duplicates Tab -->
        <v-window-item value="duplicates">
          <v-card elevation="0" rounded="lg" class="pa-4 mb-4">
            <!-- Entity Type Filter -->
            <div class="d-flex align-center mb-4 flex-wrap gap-2">
              <v-chip-group v-model="selectedEntityType" mandatory selected-class="text-primary">
                <v-chip v-for="et in entityTypes" :key="et.value" :value="et.value" filter>
                  {{ et.label }}
                </v-chip>
              </v-chip-group>
              <v-spacer></v-spacer>
              <v-btn
                color="primary"
                variant="outlined"
                :loading="scanningEntity"
                :disabled="scanningEntity"
                prepend-icon="mdi-refresh"
                @click="scanSelectedEntity"
              >
                Scan {{ formatEntityName(selectedEntityType) }}
              </v-btn>
            </div>

            <!-- Duplicate Groups -->
            <div v-if="entityDuplicates">
              <v-alert v-if="!entityDuplicates.groups || entityDuplicates.groups.length === 0" type="success" variant="tonal" class="mb-3">
                <v-icon start>mdi-check-circle</v-icon>
                No duplicates found in {{ formatEntityName(selectedEntityType) }}.
              </v-alert>

              <div v-else>
                <v-alert type="warning" variant="tonal" class="mb-3">
                  <v-icon start>mdi-content-copy</v-icon>
                  <strong>{{ entityDuplicates.total_duplicates }} duplicates in {{ (entityDuplicates.groups || []).length }} groups</strong>
                  <span class="ml-2 text-caption">(scan took {{ entityDuplicates.scan_time }})</span>
                </v-alert>

                <!-- Duplicate Group List -->
                <v-expansion-panels variant="accordion" class="mb-4">
                  <v-expansion-panel v-for="(group, idx) in entityDuplicates.groups" :key="idx">
                    <v-expansion-panel-title>
                      <div class="d-flex align-center gap-2">
                        <v-chip color="warning" size="small">{{ group.count }}</v-chip>
                        <span class="font-weight-medium">{{ group.key }}</span>
                      </div>
                    </v-expansion-panel-title>
                    <v-expansion-panel-text>
                      <!-- Record Selection for Merge -->
                      <v-alert type="info" variant="tonal" density="compact" class="mb-3">
                        Select which record to <strong>keep</strong>. Others will be merged into it.
                      </v-alert>

                      <v-radio-group v-model="selectedKeepIds[group.key]" class="mt-0">
                        <v-card
                          v-for="record in group.records"
                          :key="record.id"
                          elevation="1"
                          rounded="lg"
                          class="pa-3 mb-2"
                          :style="{ borderLeft: selectedKeepIds[group.key] === record.id ? '4px solid rgb(var(--v-theme-primary))' : '4px solid transparent' }"
                        >
                          <v-radio :value="record.id">
                            <template #label>
                              <div class="ml-2" style="flex: 1">
                                <div class="d-flex align-center">
                                  <strong>ID: {{ record.id }}</strong>
                                  <span v-if="record.name" class="ml-2">- {{ record.name }}</span>
                                  <span v-if="record.email" class="ml-2">- {{ record.email }}</span>
                                  <v-chip v-if="record.fk_count > 0" color="info" size="x-small" class="ml-2">
                                    {{ record.fk_count }} references
                                  </v-chip>
                                </div>
                                <div style="font-size: 12px; opacity: 0.7">
                                  Created: {{ formatDateTime(record.created_at) }}
                                </div>
                                <div v-if="record.details" style="font-size: 11px; opacity: 0.6">
                                  <span v-for="(val, key) in record.details" :key="key" class="mr-2">
                                    {{ key }}: {{ val }}
                                  </span>
                                </div>
                              </div>
                            </template>
                          </v-radio>
                        </v-card>
                      </v-radio-group>

                      <!-- Merge Button -->
                      <div class="d-flex justify-end mt-3">
                        <v-btn
                          color="warning"
                          variant="outlined"
                          :disabled="!selectedKeepIds[group.key]"
                          prepend-icon="mdi-merge"
                          @click="previewMerge(group)"
                        >
                          Preview Merge
                        </v-btn>
                      </div>
                    </v-expansion-panel-text>
                  </v-expansion-panel>
                </v-expansion-panels>
              </div>
            </div>

            <v-alert v-else type="info" variant="tonal">
              Select an entity type and click "Scan" to check for duplicates.
            </v-alert>
          </v-card>
        </v-window-item>

        <!-- Data Issues Tab -->
        <v-window-item value="issues">
          <v-card elevation="0" rounded="lg" class="pa-4 mb-4">
            <div class="d-flex align-center mb-4">
              <h3 style="color: #2c3e50; font-size: 18px; font-weight: 600">Data Quality Issues</h3>
              <v-spacer></v-spacer>
              <v-btn
                color="primary"
                variant="outlined"
                :loading="scanningIssues"
                :disabled="scanningIssues"
                prepend-icon="mdi-magnify-scan"
                @click="scanDataQuality"
              >
                Scan Issues
              </v-btn>
            </div>

            <!-- Quality Check Type Summary -->
            <div v-if="issuesSummary" class="mb-4">
              <div class="d-flex flex-wrap gap-2">
                <v-chip
                  v-for="checkType in qualityCheckTypes"
                  :key="checkType.value"
                  :color="getIssueCount(checkType.value) > 0 ? checkType.severity : 'default'"
                  :variant="selectedIssueTypeFilter === checkType.value ? 'flat' : 'outlined'"
                  @click="selectedIssueTypeFilter = selectedIssueTypeFilter === checkType.value ? null : checkType.value"
                  style="cursor: pointer"
                >
                  <v-icon start size="16">{{ checkType.icon }}</v-icon>
                  {{ checkType.label }}
                  <v-badge
                    :content="getIssueCount(checkType.value)"
                    :color="getIssueCount(checkType.value) > 0 ? checkType.severity : 'grey'"
                    inline
                    class="ml-2"
                  />
                </v-chip>
              </div>
            </div>

            <!-- Active Filter Indicator -->
            <div v-if="selectedIssueTypeFilter" class="mb-4">
              <v-chip
                color="primary"
                closable
                @click:close="selectedIssueTypeFilter = null"
              >
                <v-icon start>mdi-filter</v-icon>
                Showing: {{ formatIssueType(selectedIssueTypeFilter) }}
              </v-chip>
            </div>

            <v-alert v-if="!issuesSummary" type="info" variant="tonal" class="mb-4">
              Click "Scan Issues" to check for data quality problems.
            </v-alert>

            <div v-else-if="issuesSummary.total_count === 0">
              <v-alert type="success" variant="tonal">
                <v-icon start>mdi-check-circle</v-icon>
                All quality checks passed - no issues found!
              </v-alert>
            </div>

            <div v-else>
              <v-alert :type="filteredIssues.length > 0 ? 'error' : 'success'" variant="tonal" class="mb-4">
                <v-icon start>{{ filteredIssues.length > 0 ? 'mdi-alert-circle' : 'mdi-check-circle' }}</v-icon>
                <strong>{{ filteredIssues.length }} issues{{ selectedIssueTypeFilter ? ' (filtered)' : '' }}</strong>
                <span class="ml-2 text-caption">(scan took {{ issuesSummary.scan_time }})</span>
              </v-alert>

              <!-- Issues List -->
              <v-data-table
                v-if="filteredIssues.length > 0"
                :headers="issueHeaders"
                :items="filteredIssues"
                :items-per-page="10"
                class="elevation-1"
              >
                <template #item.severity="{ value }">
                  <v-chip
                    :color="value === 'error' ? 'error' : value === 'warning' ? 'warning' : 'info'"
                    size="small"
                  >
                    {{ value }}
                  </v-chip>
                </template>
                <template #item.entity_type="{ value }">
                  {{ formatEntityName(value) }}
                </template>
                <template #item.issue_type="{ value }">
                  {{ formatIssueType(value) }}
                </template>
              </v-data-table>
            </div>
          </v-card>
        </v-window-item>
      </v-window>

      <!-- Merge Preview Dialog -->
      <v-dialog v-model="mergePreviewDialog" max-width="600">
        <v-card>
          <v-card-title style="background-color: rgb(var(--v-theme-warning)); color: white">
            <v-icon color="white" class="mr-2">mdi-merge</v-icon>
            Merge Preview
          </v-card-title>
          <v-card-text class="pt-4">
            <div v-if="mergePreview">
              <v-alert v-if="!mergePreview.can_proceed" type="error" variant="tonal" class="mb-4">
                <v-icon start>mdi-alert</v-icon>
                Merge cannot proceed. See warnings below.
              </v-alert>

              <div class="mb-4">
                <strong>Entity Type:</strong> {{ formatEntityName(mergePreview.entity_type) }}
              </div>
              <div class="mb-4">
                <strong>Keep Record ID:</strong> {{ mergePreview.keep_id }}
              </div>
              <div class="mb-4">
                <strong>Delete Record IDs:</strong> {{ mergePreview.delete_ids.join(', ') }}
              </div>

              <!-- FK Updates -->
              <div v-if="mergePreview.fk_updates && Object.keys(mergePreview.fk_updates).length > 0" class="mb-4">
                <strong>FK References to Update:</strong>
                <v-list density="compact" class="mt-2">
                  <v-list-item v-for="(count, table) in mergePreview.fk_updates" :key="table">
                    <v-list-item-title>{{ table }}: {{ count }} records</v-list-item-title>
                  </v-list-item>
                </v-list>
              </div>

              <!-- Warnings -->
              <div v-if="mergePreview.warnings && mergePreview.warnings.length > 0">
                <v-alert type="warning" variant="tonal">
                  <strong>Warnings:</strong>
                  <ul class="mt-2 mb-0">
                    <li v-for="(warning, idx) in mergePreview.warnings" :key="idx">{{ warning }}</li>
                  </ul>
                </v-alert>
              </div>
            </div>
          </v-card-text>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn variant="text" @click="mergePreviewDialog = false">Cancel</v-btn>
            <v-btn
              color="warning"
              variant="flat"
              :loading="merging"
              :disabled="!mergePreview?.can_proceed"
              @click="executeMerge"
            >
              Confirm Merge
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-container>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import axios from '@/utils/axios'

// State
const activeTab = ref('overview')
const scanning = ref(false)
const scanned = ref(false)
const scanningEntity = ref(false)
const scanningIssues = ref(false)
const merging = ref(false)

const duplicateSummary = ref(null)
const issuesSummary = ref(null)
const entityDuplicates = ref(null)
const selectedEntityType = ref('movements')
const selectedKeepIds = ref({})

const mergePreviewDialog = ref(false)
const mergePreview = ref(null)
const currentMergeGroup = ref(null)
const selectedIssueTypeFilter = ref(null)

const totalDuplicates = ref(0)
const totalIssues = ref(0)

// Computed properties
const filteredIssues = computed(() => {
  if (!issuesSummary.value?.issues) return []
  if (!selectedIssueTypeFilter.value) return issuesSummary.value.issues
  return issuesSummary.value.issues.filter(
    issue => issue.issue_type === selectedIssueTypeFilter.value
  )
})

const entityTypes = [
  { value: 'movements', label: 'Movements' },
  { value: 'wods', label: 'WODs' },
  { value: 'user_workouts', label: 'User Workouts' },
  { value: 'users', label: 'Users' },
  { value: 'workouts', label: 'Workout Templates' },
]

const qualityCheckTypes = [
  {
    value: 'orphaned_fk',
    label: 'Orphaned References',
    description: 'FK references to deleted records',
    icon: 'mdi-link-off',
    severity: 'error'
  },
  {
    value: 'empty_required_field',
    label: 'Empty Required Fields',
    description: 'Records with missing required data',
    icon: 'mdi-text-box-remove',
    severity: 'error'
  },
  {
    value: 'future_date',
    label: 'Future Dates',
    description: 'Workout dates set in the future',
    icon: 'mdi-calendar-alert',
    severity: 'warning'
  },
  {
    value: 'invalid_email',
    label: 'Invalid Emails',
    description: 'Users with malformed email addresses',
    icon: 'mdi-email-alert',
    severity: 'warning'
  },
]

const issueHeaders = [
  { title: 'Severity', key: 'severity', width: '100px' },
  { title: 'Entity', key: 'entity_type', width: '120px' },
  { title: 'Record ID', key: 'record_id', width: '100px' },
  { title: 'Field', key: 'field', width: '100px' },
  { title: 'Issue Type', key: 'issue_type', width: '150px' },
  { title: 'Description', key: 'description' },
]

// Methods
const runFullScan = async () => {
  scanning.value = true
  try {
    const response = await axios.get('/api/admin/data-quality/full-scan')
    duplicateSummary.value = { by_entity: response.data.duplicates }
    issuesSummary.value = response.data.issues
    totalDuplicates.value = response.data.summary.total_duplicate_records
    totalIssues.value = response.data.summary.total_data_quality_issues
    scanned.value = true
  } catch (err) {
    console.error('Failed to run full scan:', err)
    alert(err.response?.data?.error || 'Failed to run scan')
  } finally {
    scanning.value = false
  }
}

const scanSelectedEntity = async () => {
  scanningEntity.value = true
  selectedKeepIds.value = {}
  try {
    const response = await axios.get(`/api/admin/data-quality/duplicates/${selectedEntityType.value}`)
    entityDuplicates.value = response.data
  } catch (err) {
    console.error('Failed to scan entity:', err)
    alert(err.response?.data?.error || 'Failed to scan entity')
  } finally {
    scanningEntity.value = false
  }
}

const scanDataQuality = async () => {
  scanningIssues.value = true
  try {
    const response = await axios.get('/api/admin/data-quality/issues')
    issuesSummary.value = response.data
    totalIssues.value = response.data.total_count
  } catch (err) {
    console.error('Failed to scan data quality:', err)
    alert(err.response?.data?.error || 'Failed to scan data quality')
  } finally {
    scanningIssues.value = false
  }
}

const viewEntityDuplicates = (entity) => {
  selectedEntityType.value = entity
  activeTab.value = 'duplicates'
  // Use cached data from full scan if available
  if (duplicateSummary.value?.by_entity?.[entity]) {
    entityDuplicates.value = duplicateSummary.value.by_entity[entity]
  } else {
    scanSelectedEntity()
  }
}

const viewIssuesByType = (issueType) => {
  selectedIssueTypeFilter.value = issueType
  activeTab.value = 'issues'
}

const previewMerge = async (group) => {
  const keepId = selectedKeepIds.value[group.key]
  if (!keepId) return

  const deleteIds = group.records
    .filter(r => r.id !== keepId)
    .map(r => r.id)

  try {
    const response = await axios.post('/api/admin/data-quality/duplicates/merge/preview', {
      entity_type: selectedEntityType.value,
      keep_id: keepId,
      delete_ids: deleteIds
    })
    mergePreview.value = response.data
    currentMergeGroup.value = group
    mergePreviewDialog.value = true
  } catch (err) {
    console.error('Failed to preview merge:', err)
    alert(err.response?.data?.error || 'Failed to preview merge')
  }
}

const executeMerge = async () => {
  if (!mergePreview.value || !mergePreview.value.can_proceed) return

  merging.value = true
  try {
    await axios.post('/api/admin/data-quality/duplicates/merge/confirm', {
      entity_type: mergePreview.value.entity_type,
      keep_id: mergePreview.value.keep_id,
      delete_ids: mergePreview.value.delete_ids
    })
    mergePreviewDialog.value = false

    // Refresh the scan
    await scanSelectedEntity()

    alert('Merge completed successfully!')
  } catch (err) {
    console.error('Failed to execute merge:', err)
    alert(err.response?.data?.error || 'Failed to execute merge')
  } finally {
    merging.value = false
  }
}

const formatEntityName = (entity) => {
  const names = {
    movements: 'Movements',
    wods: 'WODs',
    user_workouts: 'User Workouts',
    users: 'Users',
    workouts: 'Workout Templates'
  }
  return names[entity] || entity
}

const formatIssueType = (issueType) => {
  const types = {
    orphaned_fk: 'Orphaned FK Reference',
    empty_required_field: 'Empty Required Field',
    future_date: 'Future Date',
    invalid_email: 'Invalid Email'
  }
  return types[issueType] || issueType
}

const formatDateTime = (dateString) => {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getIssueCount = (issueType) => {
  if (!issuesSummary.value?.by_type) return 0
  return issuesSummary.value.by_type[issueType] || 0
}
</script>

<style scoped>
.entity-card-warning {
  border-color: rgb(var(--v-theme-warning)) !important;
  background-color: rgba(var(--v-theme-warning), 0.05) !important;
}

.entity-card-error {
  border-color: rgb(var(--v-theme-error)) !important;
  background-color: rgba(var(--v-theme-error), 0.05) !important;
}

.gap-2 {
  gap: 8px;
}
</style>
