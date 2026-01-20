<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <div class="d-flex align-center mb-4">
        <v-btn icon class="mr-2" @click="$router.back()">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <div>
          <h1 class="text-h5">Organization Management</h1>
          <div class="text-body-2 text-medium-emphasis">Manage gyms and organizations</div>
        </div>
        <v-spacer />
        <v-btn color="primary" @click="openCreateDialog">
          <v-icon start>mdi-plus</v-icon>
          Create Organization
        </v-btn>
      </div>

      <!-- Info Box: Organization = Gym -->
      <v-alert type="info" variant="tonal" density="compact" class="mb-4">
        <div class="text-body-2">
          <strong>Organization = Gym.</strong>
          Each organization represents a gym, box, or affiliate. Use "Locations" within an organization to define physical spaces (e.g., Main Floor, Studio A) where classes are held.
        </div>
      </v-alert>

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

      <!-- Organizations Table -->
      <v-card elevation="0" rounded="lg">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-domain</v-icon>
          Organizations ({{ total }})
        </v-card-title>

        <v-divider />

        <v-data-table
          :headers="headers"
          :items="organizations"
          :loading="loading"
          :items-per-page="limit"
          hide-default-footer
          class="elevation-0"
        >
          <!-- Name Column -->
          <template #item.name="{ item }">
            <div class="font-weight-bold">{{ item.name }}</div>
            <div v-if="item.description" class="text-caption text-medium-emphasis">
              {{ item.description }}
            </div>
          </template>

          <!-- Created Date Column -->
          <template #item.created_at="{ item }">
            <span class="text-body-2">{{ formatDate(item.created_at) }}</span>
          </template>

          <!-- Actions Column -->
          <template #item.actions="{ item }">
            <v-btn icon size="small" variant="text" title="View Details & Locations" @click="$router.push(`/admin/organizations/${item.id}`)">
              <v-icon color="primary">mdi-cog</v-icon>
            </v-btn>
            <v-btn icon size="small" variant="text" title="Manage Users" @click="manageUsers(item)">
              <v-icon color="info">mdi-account-multiple</v-icon>
            </v-btn>
            <v-btn icon size="small" variant="text" title="Delete" @click="deleteOrganization(item)">
              <v-icon color="error">mdi-delete</v-icon>
            </v-btn>
          </template>
        </v-data-table>

        <!-- Pagination -->
        <v-divider />
        <v-card-actions>
          <v-spacer />
          <v-btn :disabled="offset === 0" @click="previousPage">Previous</v-btn>
          <v-btn :disabled="offset + limit >= total" @click="nextPage">Next</v-btn>
        </v-card-actions>
      </v-card>
    </v-container>

    <!-- Create Dialog -->
    <v-dialog v-model="dialog" max-width="500px">
      <v-card>
        <v-card-title>Create Organization (Gym)</v-card-title>
        <v-card-text>
          <v-form ref="form">
            <v-text-field
              v-model="formData.name"
              label="Gym Name"
              required
              variant="outlined"
              density="comfortable"
            />
            <v-textarea
              v-model="formData.description"
              label="Description (optional)"
              variant="outlined"
              density="comfortable"
              rows="3"
            />
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="closeDialog">Cancel</v-btn>
          <v-btn color="primary" @click="saveOrganization">Create</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400px">
      <v-card>
        <v-card-title>Confirm Delete</v-card-title>
        <v-card-text>
          Are you sure you want to delete "{{ organizationToDelete?.name }}"?
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" @click="confirmDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Manage Users Dialog -->
    <manage-organization-users-dialog
      v-model="manageUsersDialog"
      :organization="selectedOrganization"
      @updated="fetchOrganizations"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'
import ManageOrganizationUsersDialog from '@/components/admin/ManageOrganizationUsersDialog.vue'

const loading = ref(false)
const error = ref(null)
const successMessage = ref(null)
const organizations = ref([])
const total = ref(0)
const limit = ref(50)
const offset = ref(0)

const dialog = ref(false)
const formData = ref({
  name: '',
  description: null
})

const deleteDialog = ref(false)
const organizationToDelete = ref(null)

const manageUsersDialog = ref(false)
const selectedOrganization = ref(null)

const headers = [
  { title: 'Name', key: 'name', sortable: true },
  { title: 'Created', key: 'created_at', sortable: true },
  { title: 'Actions', key: 'actions', sortable: false, align: 'end' }
]

async function fetchOrganizations() {
  loading.value = true
  error.value = null

  try {
    const response = await axios.get(`/api/admin/organizations?limit=${limit.value}&offset=${offset.value}`)
    organizations.value = response.data.organizations || []
    total.value = response.data.total || 0
  } catch (err) {
    console.error('Failed to fetch organizations:', err)
    error.value = err.response?.data?.error || 'Failed to fetch organizations'
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  formData.value = {
    name: '',
    description: null
  }
  dialog.value = true
}

function closeDialog() {
  dialog.value = false
  formData.value = {
    name: '',
    description: null
  }
}

async function saveOrganization() {
  if (!formData.value.name) {
    error.value = 'Organization name is required'
    return
  }

  loading.value = true
  error.value = null

  try {
    await axios.post('/api/admin/organizations', {
      name: formData.value.name,
      description: formData.value.description || null
    })
    successMessage.value = 'Organization created successfully'
    closeDialog()
    await fetchOrganizations()
  } catch (err) {
    console.error('Failed to save organization:', err)
    error.value = err.response?.data?.error || 'Failed to save organization'
  } finally {
    loading.value = false
  }
}

function manageUsers(org) {
  selectedOrganization.value = org
  manageUsersDialog.value = true
}

function deleteOrganization(org) {
  organizationToDelete.value = org
  deleteDialog.value = true
}

async function confirmDelete() {
  if (!organizationToDelete.value) return

  loading.value = true
  error.value = null

  try {
    await axios.delete(`/api/admin/organizations/${organizationToDelete.value.id}`)
    successMessage.value = 'Organization deleted successfully'
    deleteDialog.value = false
    organizationToDelete.value = null
    await fetchOrganizations()
  } catch (err) {
    console.error('Failed to delete organization:', err)
    error.value = err.response?.data?.error || 'Failed to delete organization'
  } finally {
    loading.value = false
  }
}

function formatDate(dateString) {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

function previousPage() {
  if (offset.value >= limit.value) {
    offset.value -= limit.value
    fetchOrganizations()
  }
}

function nextPage() {
  if (offset.value + limit.value < total.value) {
    offset.value += limit.value
    fetchOrganizations()
  }
}

onMounted(() => {
  fetchOrganizations()
})
</script>

<style scoped>
.mobile-view-wrapper {
  min-height: 100vh;
  padding-bottom: 70px;
}
</style>
