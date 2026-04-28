<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <AdminHeader
        title="Gym Management"
        subtitle="Manage your gyms, boxes, and affiliates"
        :breadcrumbs="[{ title: 'Organizations', to: '/admin/organizations' }]"
      />

      <div class="d-flex justify-end mb-4">
        <v-btn color="primary" @click="openCreateDialog">
          <v-icon start>mdi-plus</v-icon>
          Add Gym
        </v-btn>
      </div>

      <!-- Info Box -->
      <v-alert type="info" variant="tonal" density="compact" class="mb-4">
        <div class="text-body-2">
          Each gym can have multiple <strong>Locations</strong> (physical spaces like Main Floor, Studio A) where classes are held.
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

      <!-- Gyms Table -->
      <v-card elevation="0" rounded="lg">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-domain</v-icon>
          Gyms ({{ total }})
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

          <!-- Members Column -->
          <template #item.member_count="{ item }">
            <v-chip size="small" :color="item.member_count > 0 ? 'primary' : 'default'" variant="tonal">
              {{ item.member_count }}
            </v-chip>
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
            <v-btn icon size="small" variant="text" title="Edit Gym" @click="openEditDialog(item)">
              <v-icon color="warning">mdi-pencil</v-icon>
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

    <!-- Create / Edit Dialog -->
    <v-dialog v-model="dialog" max-width="500px">
      <v-card>
        <v-card-title>{{ editMode ? 'Edit Gym' : 'Add New Gym' }}</v-card-title>
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
          <v-btn color="primary" @click="saveOrganization">{{ editMode ? 'Save' : 'Create' }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400px">
      <v-card>
        <v-card-title>Delete Gym</v-card-title>
        <v-card-text>
          Are you sure you want to delete the gym "{{ organizationToDelete?.name }}"?
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
import AdminHeader from '@/components/AdminHeader.vue'
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
const editingId = ref(null)
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
  { title: 'Members', key: 'member_count', sortable: true, align: 'center' },
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
    console.error('Failed to fetch gyms:', err)
    error.value = err.response?.data?.error || 'Failed to fetch gyms'
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  editMode.value = false
  editingId.value = null
  formData.value = { name: '', description: null }
  dialog.value = true
}

function openEditDialog(org) {
  editMode.value = true
  editingId.value = org.id
  formData.value = { name: org.name, description: org.description || null }
  dialog.value = true
}

function closeDialog() {
  dialog.value = false
  editMode.value = false
  editingId.value = null
  formData.value = { name: '', description: null }
}

async function saveOrganization() {
  if (!formData.value.name) {
    error.value = 'Gym name is required'
    return
  }

  loading.value = true
  error.value = null

  try {
    const payload = {
      name: formData.value.name,
      description: formData.value.description || null
    }
    if (editMode.value) {
      await axios.put(`/api/admin/organizations/${editingId.value}`, payload)
      successMessage.value = 'Gym updated successfully'
    } else {
      await axios.post('/api/admin/organizations', payload)
      successMessage.value = 'Gym created successfully'
    }
    closeDialog()
    await fetchOrganizations()
  } catch (err) {
    console.error('Failed to save gym:', err)
    error.value = err.response?.data?.error || 'Failed to save gym'
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
    successMessage.value = 'Gym deleted successfully'
    deleteDialog.value = false
    organizationToDelete.value = null
    await fetchOrganizations()
  } catch (err) {
    console.error('Failed to delete gym:', err)
    error.value = err.response?.data?.error || 'Failed to delete gym'
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
