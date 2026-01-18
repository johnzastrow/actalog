<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <div class="d-flex align-center mb-4">
        <v-btn icon class="mr-2" @click="$router.back()">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <div>
          <h1 class="text-h5">My Account</h1>
          <div class="text-body-2 text-medium-emphasis">Credits, Documents & Waitlist</div>
        </div>
      </div>

      <!-- Organization Selector -->
      <v-select
        v-model="selectedOrgId"
        :items="organizations"
        item-title="name"
        item-value="id"
        label="Select Gym"
        variant="outlined"
        density="comfortable"
        class="mb-4"
        @update:model-value="fetchData"
      />

      <!-- Loading State -->
      <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />

      <!-- Error Alert -->
      <v-alert v-if="error" type="error" variant="tonal" closable class="mb-4" @click:close="error = null">
        {{ error }}
      </v-alert>

      <!-- Credits Section -->
      <v-card elevation="0" rounded="lg" class="mb-4" v-if="selectedOrgId">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2" color="primary">mdi-ticket-confirmation</v-icon>
          Class Credits
        </v-card-title>
        <v-card-text>
          <div class="text-h3 text-center py-4">
            {{ availableCredits }}
            <span class="text-body-1 text-medium-emphasis">available</span>
          </div>

          <v-list v-if="creditRecords.length > 0" density="compact">
            <v-list-subheader>Credit History</v-list-subheader>
            <v-list-item v-for="record in creditRecords" :key="record.id">
              <v-list-item-title>
                {{ record.package_name || 'Manual Credit' }}
              </v-list-item-title>
              <v-list-item-subtitle>
                {{ record.credits_total - record.credits_used }} / {{ record.credits_total }} remaining
                <span v-if="record.expires_at" class="ml-2">
                  (Expires: {{ formatDate(record.expires_at) }})
                </span>
              </v-list-item-subtitle>
            </v-list-item>
          </v-list>

          <v-alert v-else type="info" variant="tonal" density="compact">
            No credit packages yet. Contact your gym to purchase credits.
          </v-alert>
        </v-card-text>
      </v-card>

      <!-- Documents Section -->
      <v-card elevation="0" rounded="lg" class="mb-4" v-if="selectedOrgId">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2" color="primary">mdi-file-document</v-icon>
          Required Documents
        </v-card-title>
        <v-card-text>
          <v-list v-if="userDocuments.length > 0" density="compact">
            <v-list-item v-for="doc in userDocuments" :key="doc.id">
              <template v-slot:prepend>
                <v-icon :color="getDocumentStatusColor(doc.status)">
                  {{ getDocumentStatusIcon(doc.status) }}
                </v-icon>
              </template>
              <v-list-item-title>{{ doc.document_name || 'Document' }}</v-list-item-title>
              <v-list-item-subtitle>
                <v-chip :color="getDocumentStatusColor(doc.status)" size="x-small">
                  {{ doc.status }}
                </v-chip>
                <span v-if="doc.expires_at" class="ml-2 text-caption">
                  Expires: {{ formatDate(doc.expires_at) }}
                </span>
              </v-list-item-subtitle>
            </v-list-item>
          </v-list>

          <v-alert v-else type="info" variant="tonal" density="compact">
            No documents required at this time.
          </v-alert>

          <v-alert v-if="pendingDocuments.length > 0" type="warning" variant="tonal" density="compact" class="mt-4">
            You have {{ pendingDocuments.length }} pending document(s) to complete.
          </v-alert>
        </v-card-text>
      </v-card>

      <!-- Waitlist Section -->
      <v-card elevation="0" rounded="lg" class="mb-4">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2" color="primary">mdi-format-list-numbered</v-icon>
          My Waitlist
        </v-card-title>
        <v-card-text>
          <v-list v-if="waitlistEntries.length > 0" density="compact">
            <v-list-item v-for="entry in waitlistEntries" :key="entry.id">
              <v-list-item-title>{{ entry.session_name || 'Class Session' }}</v-list-item-title>
              <v-list-item-subtitle>
                Position: #{{ entry.position }}
                <v-chip :color="getWaitlistStatusColor(entry.status)" size="x-small" class="ml-2">
                  {{ entry.status }}
                </v-chip>
              </v-list-item-subtitle>
              <template v-slot:append>
                <v-btn v-if="entry.status === 'waiting'" color="error" variant="text" size="small" @click="leaveWaitlist(entry)">
                  Leave
                </v-btn>
              </template>
            </v-list-item>
          </v-list>

          <v-alert v-else type="info" variant="tonal" density="compact">
            You're not on any waitlists.
          </v-alert>
        </v-card-text>
      </v-card>

      <!-- No Organization Selected -->
      <v-alert v-if="!selectedOrgId && !loading" type="info" variant="tonal">
        Please select a gym to view your account details.
      </v-alert>
    </v-container>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from '@/utils/axios'

const loading = ref(false)
const error = ref(null)
const organizations = ref([])
const selectedOrgId = ref(null)
const creditRecords = ref([])
const availableCredits = ref(0)
const userDocuments = ref([])
const pendingDocuments = ref([])
const waitlistEntries = ref([])

onMounted(() => {
  fetchOrganizations()
  fetchWaitlistEntries()
})

async function fetchOrganizations() {
  try {
    const response = await axios.get('/api/admin/organizations')
    organizations.value = response.data.organizations || []

    if (organizations.value.length > 0 && !selectedOrgId.value) {
      selectedOrgId.value = organizations.value[0].id
      fetchData()
    }
  } catch (err) {
    console.error('Failed to fetch organizations:', err)
  }
}

async function fetchData() {
  if (!selectedOrgId.value) return

  loading.value = true
  error.value = null

  try {
    await Promise.all([
      fetchCredits(),
      fetchDocuments(),
      fetchPendingDocuments()
    ])
  } catch (err) {
    console.error('Failed to fetch data:', err)
    error.value = err.response?.data?.error || 'Failed to fetch data'
  } finally {
    loading.value = false
  }
}

async function fetchCredits() {
  const response = await axios.get(`/api/gyms/${selectedOrgId.value}/users/me/credits`)
  creditRecords.value = response.data.credit_records || []
  availableCredits.value = response.data.available_credits || 0
}

async function fetchDocuments() {
  const response = await axios.get(`/api/gyms/${selectedOrgId.value}/users/me/documents`)
  userDocuments.value = response.data.user_documents || []
}

async function fetchPendingDocuments() {
  const response = await axios.get(`/api/gyms/${selectedOrgId.value}/users/me/documents/pending`)
  pendingDocuments.value = response.data.pending_documents || []
}

async function fetchWaitlistEntries() {
  try {
    const response = await axios.get('/api/users/me/waitlist')
    waitlistEntries.value = response.data.waitlist_entries || []
  } catch (err) {
    console.error('Failed to fetch waitlist entries:', err)
  }
}

async function leaveWaitlist(entry) {
  if (!confirm('Are you sure you want to leave this waitlist?')) return

  loading.value = true
  try {
    await axios.delete(`/api/sessions/${entry.session_id}/waitlist`)
    await fetchWaitlistEntries()
  } catch (err) {
    console.error('Failed to leave waitlist:', err)
    error.value = err.response?.data?.error || 'Failed to leave waitlist'
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString()
}

function getDocumentStatusColor(status) {
  switch (status) {
    case 'completed': return 'success'
    case 'pending': return 'warning'
    case 'expired': return 'error'
    default: return 'grey'
  }
}

function getDocumentStatusIcon(status) {
  switch (status) {
    case 'completed': return 'mdi-check-circle'
    case 'pending': return 'mdi-clock-outline'
    case 'expired': return 'mdi-alert-circle'
    default: return 'mdi-file-document'
  }
}

function getWaitlistStatusColor(status) {
  switch (status) {
    case 'waiting': return 'primary'
    case 'promoted': return 'success'
    case 'expired': return 'error'
    case 'cancelled': return 'grey'
    default: return 'grey'
  }
}
</script>

<style scoped>
.mobile-view-wrapper {
  max-width: 600px;
  margin: 0 auto;
}
</style>
