<template>
  <v-app>

    <!-- Main Content -->
    <v-main style="background: #f5f7fa;  margin-bottom: 70px; overflow-y: auto">
      <v-container fluid class="pa-4">
        <!-- Page Title -->
        <div class="mb-4">
          <h1 style="color: #2c3e50; font-size: 24px; font-weight: 600">
            <v-icon color="#e91e63" size="28" class="mr-2">mdi-database-refresh</v-icon>
            WOD Score Type Cleanup
          </h1>
          <p style="color: #666; font-size: 14px">
            Detect and fix WOD performance records that don't match their defined score_type
          </p>
        </div>

        <!-- Scan Section -->
        <v-card elevation="0" rounded="lg" class="pa-4 mb-4">
          <h3 style="color: #2c3e50; font-size: 18px; font-weight: 600; margin-bottom: 12px">
            Step 1: Scan for Mismatches
          </h3>
          <p style="color: #666; font-size: 14px; margin-bottom: 16px">
            Click the button below to scan all WOD performance records and detect entries that don't match their WOD's defined score_type.
          </p>

          <v-btn
            color="#00bcd4"
            size="large"
            :loading="scanning"
            :disabled="scanning || fixing"
            @click="scanForMismatches"
            prepend-icon="mdi-magnify-scan"
          >
            Scan for Mismatches
          </v-btn>
        </v-card>

        <!-- Results Section -->
        <v-card v-if="scanned" elevation="0" rounded="lg" class="pa-4 mb-4">
          <h3 style="color: #2c3e50; font-size: 18px; font-weight: 600; margin-bottom: 12px">
            Scan Results
          </h3>

          <!-- No Issues Found -->
          <v-alert v-if="mismatches.length === 0" type="success" variant="tonal" class="mb-3">
            <v-icon start>mdi-check-circle</v-icon>
            <strong>All Clear!</strong> No score_type mismatches detected.
          </v-alert>

          <!-- Issues Found -->
          <div v-else>
            <v-alert type="warning" variant="tonal" class="mb-3">
              <v-icon start>mdi-alert</v-icon>
              <strong>{{ mismatches.length }} Mismatch{{ mismatches.length !== 1 ? 'es' : '' }} Found</strong>
            </v-alert>

            <!-- Mismatch List -->
            <div class="mb-4" style="max-height: 400px; overflow-y: auto">
              <v-card
                v-for="(mismatch, index) in mismatches"
                :key="index"
                elevation="1"
                rounded="lg"
                class="pa-3 mb-2 mismatch-card"
                style="border-left: 4px solid #ffc107; cursor: pointer"
                @click="openEditDialog(mismatch)"
              >
                <div class="d-flex align-center mb-2">
                  <v-icon color="teal" size="20" class="mr-2">mdi-alert-circle</v-icon>
                  <strong style="color: #2c3e50">{{ mismatch.wod_name }}</strong>
                  <v-chip size="x-small" color="#666" class="ml-2">ID: {{ mismatch.id }}</v-chip>
                  <v-spacer></v-spacer>
                  <v-icon color="#00bcd4" size="20">mdi-pencil</v-icon>
                </div>

                <div style="font-size: 13px; color: #666">
                  <div class="mb-1">
                    <strong>User:</strong> {{ mismatch.user_email }}
                  </div>
                  <div class="mb-1">
                    <strong>Workout Date:</strong> {{ formatDate(mismatch.workout_date) }}
                  </div>
                  <div class="mb-1">
                    <strong>Expected Score Type:</strong>
                    <v-chip size="x-small" color="#00bcd4" class="ml-1">{{ mismatch.expected_score_type }}</v-chip>
                  </div>
                  <div class="mb-1">
                    <strong>Issue:</strong> {{ mismatch.issue }}
                  </div>
                  <div>
                    <strong>Current Data:</strong> {{ formatCurrentData(mismatch) }}
                  </div>
                </div>
              </v-card>
            </div>

            <!-- Fix Action -->
            <v-divider class="my-4"></v-divider>
            <h3 style="color: #2c3e50; font-size: 18px; font-weight: 600; margin-bottom: 12px">
              Step 2: Fix Mismatches
            </h3>
            <p style="color: #666; font-size: 14px; margin-bottom: 16px">
              This will <strong>delete</strong> all {{ mismatches.length }} mismatched record{{ mismatches.length !== 1 ? 's' : '' }} from the database.
              This action cannot be undone.
            </p>

            <v-btn
              color="#e91e63"
              size="large"
              :loading="fixing"
              :disabled="scanning || fixing"
              @click="confirmFix"
              prepend-icon="mdi-delete-sweep"
            >
              Delete {{ mismatches.length }} Mismatched Record{{ mismatches.length !== 1 ? 's' : '' }}
            </v-btn>
          </div>
        </v-card>

        <!-- Success Message -->
        <v-alert v-if="fixSuccess" type="success" variant="tonal" class="mb-3">
          <v-icon start>mdi-check-circle</v-icon>
          <strong>Success!</strong> {{ fixedCount }} record{{ fixedCount !== 1 ? 's' : '' }} deleted successfully.
        </v-alert>

        <!-- Info Card -->
        <v-card elevation="0" rounded="lg" class="pa-4" style="background: #e3f2fd">
          <h3 style="color: #1976d2; font-size: 16px; font-weight: 600; margin-bottom: 8px">
            <v-icon color="#1976d2" size="20" class="mr-1">mdi-information</v-icon>
            About This Tool
          </h3>
          <p style="color: #1565c0; font-size: 13px; margin-bottom: 8px">
            This tool identifies WOD performance records where the logged data doesn't match the WOD's defined score_type:
          </p>
          <ul style="color: #1565c0; font-size: 13px; margin-left: 20px">
            <li><strong>Time (HH:MM:SS)</strong> WODs must have time_seconds, not rounds/reps/weight</li>
            <li><strong>Rounds+Reps</strong> WODs must have rounds (and optionally reps), not time_seconds/weight</li>
            <li><strong>Max Weight</strong> WODs must have weight, not time_seconds/rounds/reps</li>
          </ul>
          <p style="color: #1565c0; font-size: 13px; margin-top: 8px">
            Invalid records cannot be displayed in the Performance view and should be deleted or manually corrected.
          </p>
        </v-card>
      </v-container>
    </v-main>

    <!-- Edit Dialog -->
    <v-dialog v-model="editDialog" max-width="600">
      <v-card>
        <v-card-title style="background: #00bcd4; color: white">
          <v-icon color="white" class="mr-2">mdi-pencil</v-icon>
          Edit WOD Record
        </v-card-title>
        <v-card-text class="pt-4">
          <div v-if="editingMismatch">
            <!-- Record Info -->
            <div class="mb-4" style="background: #f5f7fa; padding: 12px; border-radius: 8px">
              <div style="font-size: 14px; color: #666">
                <div><strong>WOD:</strong> {{ editingMismatch.wod_name }}</div>
                <div><strong>User:</strong> {{ editingMismatch.user_email }}</div>
                <div><strong>Date:</strong> {{ formatDate(editingMismatch.workout_date) }}</div>
                <div><strong>Expected Score Type:</strong> <v-chip size="x-small" color="#00bcd4">{{ editingMismatch.expected_score_type }}</v-chip></div>
              </div>
            </div>

            <!-- Edit Form Fields -->
            <div v-if="editingMismatch.expected_score_type === 'Time (HH:MM:SS)'">
              <h4 class="mb-2" style="color: #2c3e50">Time-based WOD</h4>
              <v-text-field
                v-model.number="editForm.time_minutes"
                label="Minutes"
                type="number"
                variant="outlined"
                density="compact"
                min="0"
                class="mb-2"
              ></v-text-field>
              <v-text-field
                v-model.number="editForm.time_seconds"
                label="Seconds"
                type="number"
                variant="outlined"
                density="compact"
                min="0"
                max="59"
                class="mb-2"
              ></v-text-field>
            </div>

            <div v-else-if="editingMismatch.expected_score_type === 'Rounds+Reps'">
              <h4 class="mb-2" style="color: #2c3e50">Rounds+Reps WOD</h4>
              <v-text-field
                v-model.number="editForm.rounds"
                label="Rounds"
                type="number"
                variant="outlined"
                density="compact"
                min="0"
                class="mb-2"
              ></v-text-field>
              <v-text-field
                v-model.number="editForm.reps"
                label="Reps"
                type="number"
                variant="outlined"
                density="compact"
                min="0"
                class="mb-2"
              ></v-text-field>
            </div>

            <div v-else-if="editingMismatch.expected_score_type === 'Max Weight'">
              <h4 class="mb-2" style="color: #2c3e50">Max Weight WOD</h4>
              <v-text-field
                v-model.number="editForm.weight"
                label="Weight (lbs)"
                type="number"
                variant="outlined"
                density="compact"
                min="0"
                step="0.5"
                class="mb-2"
              ></v-text-field>
            </div>

            <!-- Notes -->
            <v-textarea
              v-model="editForm.notes"
              label="Notes (optional)"
              variant="outlined"
              density="compact"
              rows="3"
              class="mt-2"
            ></v-textarea>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="editDialog = false" variant="text">Cancel</v-btn>
          <v-btn @click="saveEdit" color="#00bcd4" variant="flat" :loading="saving">Save Changes</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Confirm Dialog -->
    <v-dialog v-model="confirmDialog" max-width="400">
      <v-card>
        <v-card-title style="background: #e91e63; color: white">
          <v-icon color="white" class="mr-2">mdi-alert</v-icon>
          Confirm Deletion
        </v-card-title>
        <v-card-text class="pt-4">
          <p style="font-size: 14px; color: #666">
            Are you sure you want to delete <strong>{{ mismatches.length }} mismatched record{{ mismatches.length !== 1 ? 's' : '' }}</strong>?
          </p>
          <p style="font-size: 14px; color: #e91e63; margin-top: 12px">
            <v-icon color="#e91e63" size="18">mdi-alert-circle</v-icon>
            This action cannot be undone.
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="confirmDialog = false" variant="text">Cancel</v-btn>
          <v-btn @click="fixMismatches" color="#e91e63" variant="flat">Delete Records</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-app>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const router = useRouter()
const authStore = useAuthStore()

const scanning = ref(false)
const fixing = ref(false)
const scanned = ref(false)
const mismatches = ref([])
const confirmDialog = ref(false)
const fixSuccess = ref(false)
const fixedCount = ref(0)
const editDialog = ref(false)
const editingMismatch = ref(null)
const saving = ref(false)
const editForm = ref({
  time_minutes: null,
  time_seconds: null,
  rounds: null,
  reps: null,
  weight: null,
  notes: ''
})

const scanForMismatches = async () => {
  scanning.value = true
  fixSuccess.value = false

  try {
    const response = await axios.get('/api/admin/data-cleanup/wod-mismatches')
    mismatches.value = response.data.mismatches || []
    scanned.value = true
  } catch (err) {
    console.error('Failed to scan for mismatches:', err)
    alert(err.response?.data?.message || 'Failed to scan for mismatches')
  } finally {
    scanning.value = false
  }
}

const confirmFix = () => {
  confirmDialog.value = true
}

const fixMismatches = async () => {
  confirmDialog.value = false
  fixing.value = true
  fixSuccess.value = false

  try {
    const response = await axios.delete('/api/admin/data-cleanup/wod-mismatches')
    fixedCount.value = response.data.deleted_count || 0
    fixSuccess.value = true

    // Refresh scan results
    await scanForMismatches()
  } catch (err) {
    console.error('Failed to fix mismatches:', err)
    alert(err.response?.data?.message || 'Failed to fix mismatches')
  } finally {
    fixing.value = false
  }
}

const formatDate = (dateString) => {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

const formatCurrentData = (mismatch) => {
  const parts = []
  if (mismatch.time_seconds) parts.push(`time_seconds=${mismatch.time_seconds}`)
  if (mismatch.rounds) parts.push(`rounds=${mismatch.rounds}`)
  if (mismatch.reps) parts.push(`reps=${mismatch.reps}`)
  if (mismatch.weight) parts.push(`weight=${mismatch.weight}`)
  return parts.join(', ') || 'No data'
}

const openEditDialog = (mismatch) => {
  editingMismatch.value = mismatch

  // Reset form
  editForm.value = {
    time_minutes: null,
    time_seconds: null,
    rounds: null,
    reps: null,
    weight: null,
    notes: ''
  }

  // Populate form with existing data
  if (mismatch.time_seconds) {
    editForm.value.time_minutes = Math.floor(mismatch.time_seconds / 60)
    editForm.value.time_seconds = mismatch.time_seconds % 60
  }
  if (mismatch.rounds) editForm.value.rounds = mismatch.rounds
  if (mismatch.reps) editForm.value.reps = mismatch.reps
  if (mismatch.weight) editForm.value.weight = mismatch.weight

  editDialog.value = true
}

const saveEdit = async () => {
  saving.value = true

  try {
    // Build update payload based on expected score type
    const payload = {
      time_seconds: null,
      rounds: null,
      reps: null,
      weight: null,
      notes: editForm.value.notes || ''
    }

    if (editingMismatch.value.expected_score_type === 'Time (HH:MM:SS)') {
      const minutes = editForm.value.time_minutes || 0
      const seconds = editForm.value.time_seconds || 0
      payload.time_seconds = (minutes * 60) + seconds
    } else if (editingMismatch.value.expected_score_type === 'Rounds+Reps') {
      payload.rounds = editForm.value.rounds || 0
      payload.reps = editForm.value.reps || 0
    } else if (editingMismatch.value.expected_score_type === 'Max Weight') {
      payload.weight = editForm.value.weight || 0
    }

    // Call backend API to update the record
    await axios.put(`/api/admin/data-cleanup/wod-record/${editingMismatch.value.id}`, payload)

    editDialog.value = false

    // Refresh scan results
    await scanForMismatches()
  } catch (err) {
    console.error('Failed to save edit:', err)
    alert(err.response?.data?.message || 'Failed to save changes')
  } finally {
    saving.value = false
  }
}

const logout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.v-card {
  transition: all 0.2s;
}
</style>
