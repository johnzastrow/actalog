<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <AdminHeader
        :title="organization?.name || 'Gym Details'"
        subtitle="Manage gym details and locations"
        :breadcrumbs="[
          { title: 'Organizations', to: '/admin/organizations' },
          { title: organization?.name || 'Details', to: `/admin/organizations/${$route.params.id}` }
        ]"
      />

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

      <!-- Organization Details Card -->
      <v-card elevation="0" rounded="lg" class="mb-4">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-domain</v-icon>
          Gym Details
        </v-card-title>
        <v-divider />
        <v-card-text>
          <v-form ref="orgForm">
            <v-text-field
              v-model="orgFormData.name"
              label="Gym Name"
              required
              density="comfortable"
              variant="outlined"
              class="mb-3"
            />
            <v-textarea
              v-model="orgFormData.description"
              label="Description (optional)"
              density="comfortable"
              variant="outlined"
              rows="3"
            />
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn color="primary" :loading="saving" @click="saveOrganization">
            <v-icon start>mdi-content-save</v-icon>
            Save Changes
          </v-btn>
        </v-card-actions>
      </v-card>

      <!-- Locations Card -->
      <v-card elevation="0" rounded="lg">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-map-marker</v-icon>
          Locations ({{ locations.length }})
          <v-spacer />
          <v-btn color="primary" size="small" @click="openLocationDialog()">
            <v-icon start>mdi-plus</v-icon>
            Add Location
          </v-btn>
        </v-card-title>
        <v-divider />
        <v-card-text class="pa-0">
          <v-alert type="info" variant="tonal" density="compact" class="ma-3">
            <div class="text-body-2">
              Locations are physical spaces within your gym where classes are held (e.g., Main Floor, Studio A, Outdoor Area).
            </div>
          </v-alert>

          <v-list v-if="locations.length > 0" density="compact">
            <v-list-item v-for="loc in locations" :key="loc.id" class="py-2">
              <template #prepend>
                <v-avatar :color="loc.is_active ? 'primary' : 'grey'" size="40">
                  <v-icon color="white">mdi-map-marker</v-icon>
                </v-avatar>
              </template>
              <v-list-item-title class="font-weight-medium">{{ loc.name }}</v-list-item-title>
              <v-list-item-subtitle>
                <span v-if="loc.address" class="mr-2">{{ loc.address }}</span>
                <span v-if="loc.capacity">Capacity: {{ loc.capacity }}</span>
                <span v-if="loc.description" class="text-disabled"> - {{ loc.description }}</span>
              </v-list-item-subtitle>
              <template #append>
                <v-chip :color="loc.is_active ? 'success' : 'grey'" size="x-small" class="mr-2">
                  {{ loc.is_active ? 'Active' : 'Inactive' }}
                </v-chip>
                <v-btn icon size="small" variant="text" title="Edit" @click="editLocation(loc)">
                  <v-icon color="warning">mdi-pencil</v-icon>
                </v-btn>
                <v-btn icon size="small" variant="text" title="Delete" @click="deleteLocation(loc)">
                  <v-icon color="error">mdi-delete</v-icon>
                </v-btn>
              </template>
            </v-list-item>
          </v-list>
          <div v-else class="pa-4 text-center text-medium-emphasis">
            <v-icon size="48" color="grey-lighten-1" class="mb-2">mdi-map-marker-off</v-icon>
            <div>No locations configured yet.</div>
            <div class="text-caption">Add a location to start scheduling classes.</div>
          </div>
        </v-card-text>
      </v-card>

      <!-- Users Section -->
      <v-card elevation="0" rounded="lg" class="mt-4">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-account-multiple</v-icon>
          Users ({{ users.length }})
          <v-spacer />
          <v-btn color="primary" size="small" @click="manageUsersDialog = true">
            <v-icon start>mdi-account-plus</v-icon>
            Manage Users
          </v-btn>
        </v-card-title>
        <v-divider />
        <v-list v-if="users.length > 0" density="compact">
          <v-list-item v-for="user in users" :key="user.id" class="py-2">
            <template #prepend>
              <v-avatar color="primary" size="32">
                <span class="text-caption">{{ getUserInitials(user) }}</span>
              </v-avatar>
            </template>
            <v-list-item-title>
              {{ user.name || user.email }}
              <v-chip v-if="user.role === 'admin'" color="error" size="x-small" class="ml-2">Admin</v-chip>
            </v-list-item-title>
            <v-list-item-subtitle>{{ user.email }}</v-list-item-subtitle>
          </v-list-item>
        </v-list>
        <v-card-text v-else>
          <v-alert type="info" variant="tonal" density="compact">
            No users assigned to this organization.
          </v-alert>
        </v-card-text>
      </v-card>
    </v-container>

    <!-- Location Dialog -->
    <v-dialog v-model="locationDialog" max-width="500">
      <v-card>
        <v-card-title>{{ editingLocation ? 'Edit Location' : 'Add Location' }}</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="locationForm.name"
            label="Location Name"
            required
            variant="outlined"
            density="comfortable"
            placeholder="e.g., Main Floor, Studio A"
            class="mb-3"
          />
          <v-textarea
            v-model="locationForm.description"
            label="Description (optional)"
            variant="outlined"
            density="comfortable"
            rows="2"
            class="mb-3"
          />
          <v-text-field
            v-model="locationForm.address"
            label="Address (optional)"
            variant="outlined"
            density="comfortable"
            placeholder="Only if different from gym address"
            class="mb-3"
          />
          <v-text-field
            v-model.number="locationForm.capacity"
            label="Capacity"
            type="number"
            variant="outlined"
            density="comfortable"
            min="0"
            class="mb-3"
          />
          <v-switch
            v-if="editingLocation"
            v-model="locationForm.is_active"
            label="Active"
            color="success"
            density="comfortable"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="locationDialog = false">Cancel</v-btn>
          <v-btn color="primary" :loading="savingLocation" @click="saveLocation">
            {{ editingLocation ? 'Update' : 'Create' }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Location Confirmation -->
    <v-dialog v-model="deleteLocationDialog" max-width="400">
      <v-card>
        <v-card-title>Confirm Delete</v-card-title>
        <v-card-text>
          Are you sure you want to delete the location "{{ locationToDelete?.name }}"?
          <v-alert v-if="locationToDelete?.has_sessions" type="warning" variant="tonal" density="compact" class="mt-3">
            This location may have class sessions scheduled. Deleting it will affect those sessions.
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="deleteLocationDialog = false">Cancel</v-btn>
          <v-btn color="error" :loading="deletingLocation" @click="confirmDeleteLocation">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Manage Users Dialog -->
    <manage-organization-users-dialog
      v-model="manageUsersDialog"
      :organization="organization"
      @updated="fetchUsers"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from '@/utils/axios'
import AdminHeader from '@/components/AdminHeader.vue'
import ManageOrganizationUsersDialog from '@/components/admin/ManageOrganizationUsersDialog.vue'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const saving = ref(false)
const error = ref(null)
const successMessage = ref(null)

const organization = ref(null)
const orgFormData = ref({ name: '', description: null })
const locations = ref([])
const users = ref([])

// Location dialog state
const locationDialog = ref(false)
const editingLocation = ref(null)
const locationForm = ref({ name: '', description: '', address: '', capacity: 0, is_active: true })
const savingLocation = ref(false)

// Delete location state
const deleteLocationDialog = ref(false)
const locationToDelete = ref(null)
const deletingLocation = ref(false)

// Users dialog
const manageUsersDialog = ref(false)

onMounted(() => {
  const orgId = route.params.id
  if (!orgId) {
    router.push('/admin/organizations')
    return
  }
  fetchOrganization(orgId)
})

async function fetchOrganization(orgId) {
  loading.value = true
  error.value = null
  try {
    const [orgResponse, locationsResponse, usersResponse] = await Promise.all([
      axios.get(`/api/admin/organizations/${orgId}`),
      axios.get(`/api/gyms/${orgId}/locations?include_inactive=true`),
      axios.get(`/api/admin/organizations/${orgId}/users`)
    ])
    organization.value = orgResponse.data
    orgFormData.value = {
      name: organization.value.name,
      description: organization.value.description
    }
    locations.value = locationsResponse.data.locations || []
    users.value = usersResponse.data.users || []
  } catch (err) {
    console.error('Failed to fetch organization:', err)
    error.value = err.response?.data?.error || 'Failed to fetch organization details'
  } finally {
    loading.value = false
  }
}

async function saveOrganization() {
  if (!orgFormData.value.name) {
    error.value = 'Gym name is required'
    return
  }
  saving.value = true
  error.value = null
  try {
    await axios.put(`/api/admin/organizations/${organization.value.id}`, {
      name: orgFormData.value.name,
      description: orgFormData.value.description || null
    })
    organization.value.name = orgFormData.value.name
    organization.value.description = orgFormData.value.description
    successMessage.value = 'Gym updated successfully'
  } catch (err) {
    console.error('Failed to save organization:', err)
    error.value = err.response?.data?.error || 'Failed to save gym'
  } finally {
    saving.value = false
  }
}

async function fetchLocations() {
  try {
    const response = await axios.get(`/api/gyms/${organization.value.id}/locations?include_inactive=true`)
    locations.value = response.data.locations || []
  } catch (err) {
    console.error('Failed to fetch locations:', err)
  }
}

async function fetchUsers() {
  try {
    const response = await axios.get(`/api/admin/organizations/${organization.value.id}/users`)
    users.value = response.data.users || []
  } catch (err) {
    console.error('Failed to fetch users:', err)
  }
}

function openLocationDialog(loc = null) {
  editingLocation.value = loc
  locationForm.value = loc
    ? { ...loc }
    : { name: '', description: '', address: '', capacity: 20, is_active: true }
  locationDialog.value = true
}

function editLocation(loc) {
  openLocationDialog(loc)
}

async function saveLocation() {
  if (!locationForm.value.name) {
    error.value = 'Location name is required'
    return
  }
  savingLocation.value = true
  error.value = null
  try {
    if (editingLocation.value) {
      await axios.put(
        `/api/admin/gyms/${organization.value.id}/locations/${editingLocation.value.id}`,
        locationForm.value
      )
    } else {
      await axios.post(`/api/admin/gyms/${organization.value.id}/locations`, locationForm.value)
    }
    successMessage.value = editingLocation.value ? 'Location updated' : 'Location created'
    locationDialog.value = false
    fetchLocations()
  } catch (err) {
    console.error('Failed to save location:', err)
    error.value = err.response?.data?.error || 'Failed to save location'
  } finally {
    savingLocation.value = false
  }
}

function deleteLocation(loc) {
  locationToDelete.value = loc
  deleteLocationDialog.value = true
}

async function confirmDeleteLocation() {
  deletingLocation.value = true
  error.value = null
  try {
    await axios.delete(`/api/admin/gyms/${organization.value.id}/locations/${locationToDelete.value.id}`)
    successMessage.value = 'Location deleted'
    deleteLocationDialog.value = false
    locationToDelete.value = null
    fetchLocations()
  } catch (err) {
    console.error('Failed to delete location:', err)
    error.value = err.response?.data?.error || 'Failed to delete location'
  } finally {
    deletingLocation.value = false
  }
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
</script>

<style scoped>
.mobile-view-wrapper {
  min-height: 100vh;
  padding-bottom: 70px;
}
</style>
