<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <div class="d-flex align-center mb-4">
        <v-btn icon @click="$router.back()" class="mr-2">
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
            <v-btn icon size="small" variant="text" @click="viewOrgDetails(item)" title="View Details">
              <v-icon color="primary">mdi-information-outline</v-icon>
            </v-btn>
            <v-btn icon size="small" variant="text" @click="manageUsers(item)" title="Manage Users">
              <v-icon color="info">mdi-account-multiple</v-icon>
            </v-btn>
            <v-btn icon size="small" variant="text" @click="editOrganization(item)" title="Edit">
              <v-icon color="warning">mdi-pencil</v-icon>
            </v-btn>
            <v-btn icon size="small" variant="text" @click="deleteOrganization(item)" title="Delete">
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

    <!-- Create/Edit Dialog -->
    <v-dialog v-model="dialog" max-width="500px">
      <v-card>
        <v-card-title>
          {{ editMode ? 'Edit Organization' : 'Create Organization' }}
        </v-card-title>
        <v-card-text>
          <v-form ref="form">
            <v-text-field
              v-model="formData.name"
              label="Name"
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
          <v-btn color="primary" @click="saveOrganization">
            {{ editMode ? 'Update' : 'Create' }}
          </v-btn>
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
    <ManageOrganizationUsersDialog
      v-model="manageUsersDialog"
      :organization="selectedOrganization"
      @updated="fetchOrganizations"
    />

    <!-- Organization Details Dialog -->
    <v-dialog v-model="detailsDialog" max-width="700">
      <v-card v-if="organizationDetails">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-domain</v-icon>
          Organization Details
        </v-card-title>
        <v-divider />
        <v-card-text>
          <v-list>
            <v-list-item>
              <v-list-item-title class="text-subtitle-2">Name</v-list-item-title>
              <v-list-item-subtitle class="mt-1">{{ organizationDetails.name }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="organizationDetails.description">
              <v-list-item-title class="text-subtitle-2">Description</v-list-item-title>
              <v-list-item-subtitle class="mt-1">{{ organizationDetails.description }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item>
              <v-list-item-title class="text-subtitle-2">Created At</v-list-item-title>
              <v-list-item-subtitle>{{ formatFullDate(organizationDetails.created_at) }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="organizationDetails.updated_at">
              <v-list-item-title class="text-subtitle-2">Last Updated</v-list-item-title>
              <v-list-item-subtitle>{{ formatFullDate(organizationDetails.updated_at) }}</v-list-item-subtitle>
            </v-list-item>
          </v-list>

          <!-- Users Section -->
          <v-divider class="my-4" />
          <div class="px-4">
            <div class="d-flex align-center mb-3">
              <v-icon class="mr-2">mdi-account-multiple</v-icon>
              <span class="text-subtitle-1 font-weight-bold">Users ({{ organizationDetails.users?.length || 0 }})</span>
            </div>

            <div v-if="organizationDetails.users && organizationDetails.users.length > 0">
              <v-list density="compact">
                <v-list-item
                  v-for="user in organizationDetails.users"
                  :key="user.id"
                  class="mb-1"
                  rounded
                  variant="tonal"
                >
                  <template v-slot:prepend>
                    <v-avatar color="primary" size="32">
                      <span class="text-caption">{{ getUserInitials(user) }}</span>
                    </v-avatar>
                  </template>
                  <v-list-item-title>
                    {{ user.name || user.email }}
                    <v-chip v-if="user.role === 'admin'" color="error" size="x-small" class="ml-2">
                      Admin
                    </v-chip>
                  </v-list-item-title>
                  <v-list-item-subtitle>{{ user.email }}</v-list-item-subtitle>
                </v-list-item>
              </v-list>
            </div>

            <v-alert v-else type="info" variant="tonal" density="compact" class="mb-2">
              No users assigned to this organization
            </v-alert>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="detailsDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
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
const editMode = ref(false)
const formData = ref({
  id: null,
  name: '',
  description: null
})

const deleteDialog = ref(false)
const organizationToDelete = ref(null)

const manageUsersDialog = ref(false)
const selectedOrganization = ref(null)

const detailsDialog = ref(false)
const organizationDetails = ref(null)

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
  editMode.value = false
  formData.value = {
    id: null,
    name: '',
    description: null
  }
  dialog.value = true
}

function editOrganization(org) {
  editMode.value = true
  formData.value = {
    id: org.id,
    name: org.name,
    description: org.description
  }
  dialog.value = true
}

function closeDialog() {
  dialog.value = false
  formData.value = {
    id: null,
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
    if (editMode.value) {
      await axios.put(`/api/admin/organizations/${formData.value.id}`, {
        name: formData.value.name,
        description: formData.value.description || null
      })
      successMessage.value = 'Organization updated successfully'
    } else {
      await axios.post('/api/admin/organizations', {
        name: formData.value.name,
        description: formData.value.description || null
      })
      successMessage.value = 'Organization created successfully'
    }

    closeDialog()
    await fetchOrganizations()
  } catch (err) {
    console.error('Failed to save organization:', err)
    error.value = err.response?.data?.error || 'Failed to save organization'
  } finally {
    loading.value = false
  }
}

async function viewOrgDetails(org) {
  loading.value = true
  error.value = null
  try {
    // Fetch organization details and users in parallel
    const [orgResponse, usersResponse] = await Promise.all([
      axios.get(`/api/admin/organizations/${org.id}`),
      axios.get(`/api/admin/organizations/${org.id}/users`)
    ])

    organizationDetails.value = {
      ...orgResponse.data,
      users: usersResponse.data.users || []
    }
    detailsDialog.value = true
  } catch (err) {
    console.error('Failed to load organization details:', err)
    error.value = err.response?.data?.error || 'Failed to load organization details'
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

function formatFullDate(dateString) {
  if (!dateString) return 'Never'
  const date = new Date(dateString)
  return date.toLocaleString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function getUserInitials(user) {
  if (user.name) {
    const names = user.name.split(' ')
    if (names.length >= 2) {
      return (names[0][0] + names[names.length - 1][0]).toUpperCase()
    }
    return names[0][0].toUpperCase()
  }
  return user.email[0].toUpperCase()
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
