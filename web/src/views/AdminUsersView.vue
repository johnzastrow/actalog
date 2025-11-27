<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <div class="d-flex align-center mb-4">
        <v-btn icon @click="$router.back()" class="mr-2">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <div>
          <h1 class="text-h5">User Management</h1>
          <div class="text-body-2 text-medium-emphasis">Manage user accounts and permissions</div>
        </div>
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

    <!-- Users Table -->
    <v-card elevation="0" rounded="lg">
      <v-card-title class="d-flex align-center">
        <v-icon class="mr-2">mdi-account-multiple</v-icon>
        Users ({{ total }})
        <v-spacer />
        <v-text-field
          v-model="search"
          density="compact"
          label="Search users..."
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
          hide-details
          single-line
          clearable
          style="max-width: 300px;"
        />
      </v-card-title>

      <v-divider />

      <v-data-table
        :headers="headers"
        :items="users"
        :loading="loading"
        :items-per-page="limit"
        hide-default-footer
        class="elevation-0"
      >
        <!-- Email Column -->
        <template #item.email="{ item }">
          <div class="d-flex align-center">
            <v-avatar size="32" color="primary" class="mr-2">
              <span class="text-white">{{ item.email[0].toUpperCase() }}</span>
            </v-avatar>
            <div>
              <div>{{ item.email }}</div>
              <div class="text-caption text-medium-emphasis">{{ item.name }}</div>
            </div>
          </div>
        </template>

        <!-- Role Column -->
        <template #item.role="{ item }">
          <v-chip :color="item.role === 'admin' ? 'error' : 'default'" size="small">
            {{ item.role }}
          </v-chip>
        </template>

        <!-- Status Column -->
        <template #item.status="{ item }">
          <div class="d-flex align-center gap-2">
            <v-chip v-if="item.account_disabled" color="error" size="small">
              <v-icon start size="small">mdi-cancel</v-icon>
              Disabled
            </v-chip>
            <v-chip v-else-if="isLocked(item)" color="warning" size="small">
              <v-icon start size="small">mdi-lock</v-icon>
              Locked
            </v-chip>
            <v-chip v-else color="success" size="small">
              <v-icon start size="small">mdi-check-circle</v-icon>
              Active
            </v-chip>

            <!-- Email Verification Badge -->
            <v-tooltip :text="item.email_verified ? 'Email Verified' : 'Email Not Verified'" location="top">
              <template #activator="{ props }">
                <v-icon
                  v-bind="props"
                  :color="item.email_verified ? 'success' : 'grey-lighten-1'"
                  size="small"
                >
                  {{ item.email_verified ? 'mdi-email-check-outline' : 'mdi-email-outline' }}
                </v-icon>
              </template>
            </v-tooltip>
          </div>
        </template>

        <!-- Last Login Column -->
        <template #item.last_login_at="{ item }">
          <span v-if="item.last_login_at" class="text-body-2">
            {{ formatDate(item.last_login_at) }}
          </span>
          <span v-else class="text-body-2 text-medium-emphasis">Never</span>
        </template>

        <!-- Details Column -->
        <template #item.details="{ item }">
          <v-btn
            icon
            size="small"
            variant="text"
            @click="viewUserDetails(item)"
          >
            <v-icon color="primary">mdi-information-outline</v-icon>
          </v-btn>
        </template>

        <!-- Lock Column -->
        <template #item.lock="{ item }">
          <v-btn
            icon
            size="small"
            variant="text"
            :disabled="!isLocked(item)"
            @click="unlockUser(item)"
          >
            <v-icon :color="isLocked(item) ? 'error' : 'success'">
              {{ isLocked(item) ? 'mdi-lock' : 'mdi-lock-open' }}
            </v-icon>
          </v-btn>
        </template>

        <!-- Enable/Disable Column -->
        <template #item.enable="{ item }">
          <v-btn
            icon
            size="small"
            variant="text"
            @click="item.account_disabled ? enableUser(item) : disableUser(item)"
          >
            <v-icon :color="item.account_disabled ? 'error' : 'success'">
              {{ item.account_disabled ? 'mdi-account-off' : 'mdi-account-check' }}
            </v-icon>
          </v-btn>
        </template>

        <!-- Email Verification Column -->
        <template #item.email_verify="{ item }">
          <v-btn
            icon
            size="small"
            variant="text"
            @click="toggleEmailVerification(item)"
          >
            <v-icon :color="item.email_verified ? 'success' : 'error'">
              {{ item.email_verified ? 'mdi-email-check-outline' : 'mdi-email-remove-outline' }}
            </v-icon>
          </v-btn>
        </template>

        <!-- Change Role Column -->
        <template #item.change_role="{ item }">
          <v-btn
            icon
            size="small"
            variant="text"
            @click="openRoleDialog(item)"
          >
            <v-icon :color="item.role === 'admin' ? 'purple' : 'blue'">
              {{ item.role === 'admin' ? 'mdi-shield-crown' : 'mdi-account' }}
            </v-icon>
          </v-btn>
        </template>

        <!-- Delete Column -->
        <template #item.delete="{ item }">
          <v-btn
            icon
            size="small"
            variant="text"
            @click="openDeleteDialog(item)"
          >
            <v-icon color="error">mdi-delete</v-icon>
          </v-btn>
        </template>
      </v-data-table>

      <!-- Pagination -->
      <v-divider />
      <div class="d-flex align-center justify-space-between pa-4">
        <div class="text-body-2 text-medium-emphasis">
          Showing {{ offset + 1 }} to {{ Math.min(offset + limit, total) }} of {{ total }} users
        </div>
        <div class="d-flex gap-2">
          <v-btn
            variant="outlined"
            size="small"
            :disabled="offset === 0"
            @click="previousPage"
          >
            <v-icon start>mdi-chevron-left</v-icon>
            Previous
          </v-btn>
          <v-btn
            variant="outlined"
            size="small"
            :disabled="offset + limit >= total"
            @click="nextPage"
          >
            Next
            <v-icon end>mdi-chevron-right</v-icon>
          </v-btn>
        </div>
      </div>
    </v-card>

    <!-- Disable User Dialog -->
    <v-dialog v-model="disableDialog" max-width="500">
      <v-card>
        <v-card-title>Disable User Account</v-card-title>
        <v-card-text>
          <div class="mb-4">
            Are you sure you want to disable the account for <strong>{{ selectedUser?.email }}</strong>?
          </div>
          <v-textarea
            v-model="disableReason"
            label="Reason (optional)"
            placeholder="Enter reason for disabling this account..."
            rows="3"
            variant="outlined"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="disableDialog = false">Cancel</v-btn>
          <v-btn color="error" @click="confirmDisable" :loading="actionLoading">
            Disable Account
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Change Role Dialog -->
    <v-dialog v-model="roleDialog" max-width="500">
      <v-card>
        <v-card-title>Change User Role</v-card-title>
        <v-card-text>
          <div class="mb-4">
            Change role for <strong>{{ selectedUser?.email }}</strong>
          </div>
          <v-radio-group v-model="newRole">
            <v-radio label="User" value="user" />
            <v-radio label="Admin" value="admin" />
          </v-radio-group>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="roleDialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="confirmRoleChange" :loading="actionLoading">
            Change Role
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete User Dialog -->
    <v-dialog v-model="deleteDialog" max-width="500">
      <v-card>
        <v-card-title class="text-error">
          <v-icon class="mr-2">mdi-alert</v-icon>
          Delete User Account
        </v-card-title>
        <v-card-text>
          <v-alert type="warning" variant="tonal" class="mb-4">
            <strong>Warning:</strong> This action cannot be undone!
          </v-alert>
          <div class="mb-4">
            Are you sure you want to permanently delete the account for <strong>{{ selectedUser?.email }}</strong>?
          </div>
          <div class="text-body-2 text-medium-emphasis">
            This will remove:
            <ul class="mt-2">
              <li>User profile and settings</li>
              <li>All workout logs and data</li>
              <li>Personal records and performance history</li>
            </ul>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" @click="confirmDelete" :loading="actionLoading">
            Delete Permanently
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- User Details Dialog -->
    <v-dialog v-model="detailsDialog" max-width="700">
      <v-card v-if="userDetails">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-account-details</v-icon>
          User Details
        </v-card-title>
        <v-divider />
        <v-card-text>
          <v-list>
            <v-list-item>
              <v-list-item-title class="text-subtitle-2">Email</v-list-item-title>
              <v-list-item-subtitle class="d-flex align-center mt-1">
                {{ userDetails.email }}
                <v-chip
                  v-if="userDetails.email_verified"
                  color="success"
                  size="x-small"
                  class="ml-2"
                >
                  <v-icon start size="x-small">mdi-check</v-icon>
                  Verified
                </v-chip>
                <v-chip
                  v-else
                  color="warning"
                  size="x-small"
                  class="ml-2"
                >
                  <v-icon start size="x-small">mdi-alert</v-icon>
                  Not Verified
                </v-chip>
              </v-list-item-subtitle>
            </v-list-item>

            <v-list-item>
              <v-list-item-title class="text-subtitle-2">Name</v-list-item-title>
              <v-list-item-subtitle>{{ userDetails.name || 'Not set' }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item>
              <v-list-item-title class="text-subtitle-2">Role</v-list-item-title>
              <v-list-item-subtitle>
                <v-chip :color="userDetails.role === 'admin' ? 'error' : 'default'" size="small" class="mt-1">
                  {{ userDetails.role }}
                </v-chip>
              </v-list-item-subtitle>
            </v-list-item>

            <v-list-item>
              <v-list-item-title class="text-subtitle-2">Account Status</v-list-item-title>
              <v-list-item-subtitle>
                <v-chip
                  v-if="userDetails.account_disabled"
                  color="error"
                  size="small"
                  class="mt-1"
                >
                  <v-icon start size="small">mdi-cancel</v-icon>
                  Disabled
                </v-chip>
                <v-chip
                  v-else
                  color="success"
                  size="small"
                  class="mt-1"
                >
                  <v-icon start size="small">mdi-check-circle</v-icon>
                  Active
                </v-chip>
              </v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="userDetails.account_disabled && userDetails.disable_reason">
              <v-list-item-title class="text-subtitle-2">Disable Reason</v-list-item-title>
              <v-list-item-subtitle class="white-space-pre-wrap mt-1">
                {{ userDetails.disable_reason }}
              </v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="userDetails.disabled_at">
              <v-list-item-title class="text-subtitle-2">Disabled At</v-list-item-title>
              <v-list-item-subtitle>{{ formatFullDate(userDetails.disabled_at) }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item>
              <v-list-item-title class="text-subtitle-2">Created At</v-list-item-title>
              <v-list-item-subtitle>{{ formatFullDate(userDetails.created_at) }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="userDetails.last_login_at">
              <v-list-item-title class="text-subtitle-2">Last Login</v-list-item-title>
              <v-list-item-subtitle>{{ formatFullDate(userDetails.last_login_at) }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item v-if="userDetails.email_verified_at">
              <v-list-item-title class="text-subtitle-2">Email Verified At</v-list-item-title>
              <v-list-item-subtitle>{{ formatFullDate(userDetails.email_verified_at) }}</v-list-item-subtitle>
            </v-list-item>
          </v-list>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="detailsDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    </v-container>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from '@/utils/axios'

const loading = ref(false)
const actionLoading = ref(false)
const error = ref(null)
const successMessage = ref(null)
const users = ref([])
const total = ref(0)
const limit = ref(50)
const offset = ref(0)
const search = ref('')

const disableDialog = ref(false)
const roleDialog = ref(false)
const detailsDialog = ref(false)
const deleteDialog = ref(false)
const selectedUser = ref(null)
const userDetails = ref(null)
const disableReason = ref('')
const newRole = ref('user')

const headers = [
  { title: 'User', value: 'email', sortable: false },
  { title: 'Role', value: 'role', sortable: false },
  { title: 'Status', value: 'status', sortable: false },
  { title: 'Last Login', value: 'last_login_at', sortable: false },
  { title: 'Details', value: 'details', sortable: false, align: 'center', width: '80px' },
  { title: 'Lock', value: 'lock', sortable: false, align: 'center', width: '80px' },
  { title: 'Enable', value: 'enable', sortable: false, align: 'center', width: '80px' },
  { title: 'Email', value: 'email_verify', sortable: false, align: 'center', width: '80px' },
  { title: 'Change Role', value: 'change_role', sortable: false, align: 'center', width: '100px' },
  { title: 'Delete', value: 'delete', sortable: false, align: 'center', width: '80px' }
]

const filteredUsers = computed(() => {
  if (!search.value) return users.value
  const searchLower = search.value.toLowerCase()
  return users.value.filter(user =>
    user.email.toLowerCase().includes(searchLower) ||
    user.name?.toLowerCase().includes(searchLower)
  )
})

function isLocked(user) {
  if (!user.locked_until) return false
  const lockedUntil = new Date(user.locked_until)
  return lockedUntil > new Date()
}

function formatDate(dateString) {
  if (!dateString) return 'Never'
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now - date
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`

  return date.toLocaleDateString()
}

async function loadUsers() {
  loading.value = true
  error.value = null
  try {
    const response = await axios.get(`/api/admin/users?limit=${limit.value}&offset=${offset.value}`)
    users.value = response.data.users || []
    total.value = response.data.total || 0
  } catch (e) {
    console.error('Failed to load users:', e)
    error.value = e.response?.data?.message || 'Failed to load users'
  } finally {
    loading.value = false
  }
}

async function unlockUser(user) {
  actionLoading.value = true
  error.value = null
  try {
    await axios.post(`/api/admin/users/${user.id}/unlock`)
    successMessage.value = `Account unlocked for ${user.email}`
    await loadUsers()
  } catch (e) {
    console.error('Failed to unlock user:', e)
    error.value = e.response?.data?.message || 'Failed to unlock user'
  } finally {
    actionLoading.value = false
  }
}

function disableUser(user) {
  selectedUser.value = user
  disableReason.value = ''
  disableDialog.value = true
}

async function confirmDisable() {
  actionLoading.value = true
  error.value = null
  try {
    await axios.post(`/api/admin/users/${selectedUser.value.id}/disable`, {
      reason: disableReason.value || undefined
    })
    successMessage.value = `Account disabled for ${selectedUser.value.email}`
    disableDialog.value = false
    await loadUsers()
  } catch (e) {
    console.error('Failed to disable user:', e)
    error.value = e.response?.data?.message || 'Failed to disable user'
  } finally {
    actionLoading.value = false
  }
}

async function enableUser(user) {
  actionLoading.value = true
  error.value = null
  try {
    await axios.post(`/api/admin/users/${user.id}/enable`)
    successMessage.value = `Account enabled for ${user.email}`
    await loadUsers()
  } catch (e) {
    console.error('Failed to enable user:', e)
    error.value = e.response?.data?.message || 'Failed to enable user'
  } finally {
    actionLoading.value = false
  }
}

function openRoleDialog(user) {
  selectedUser.value = user
  newRole.value = user.role
  roleDialog.value = true
}

async function confirmRoleChange() {
  actionLoading.value = true
  error.value = null
  try {
    await axios.put(`/api/admin/users/${selectedUser.value.id}/role`, {
      role: newRole.value
    })
    successMessage.value = `Role changed to ${newRole.value} for ${selectedUser.value.email}`
    roleDialog.value = false
    await loadUsers()
  } catch (e) {
    console.error('Failed to change role:', e)
    error.value = e.response?.data?.message || 'Failed to change role'
  } finally {
    actionLoading.value = false
  }
}

function openDeleteDialog(user) {
  selectedUser.value = user
  deleteDialog.value = true
}

async function confirmDelete() {
  actionLoading.value = true
  error.value = null
  try {
    await axios.delete(`/api/admin/users/${selectedUser.value.id}`)
    successMessage.value = `User ${selectedUser.value.email} has been permanently deleted`
    deleteDialog.value = false
    await loadUsers()
  } catch (e) {
    console.error('Failed to delete user:', e)
    error.value = e.response?.data?.message || 'Failed to delete user'
  } finally {
    actionLoading.value = false
  }
}

async function viewUserDetails(user) {
  actionLoading.value = true
  error.value = null
  try {
    const response = await axios.get(`/api/admin/users/${user.id}`)
    userDetails.value = response.data
    detailsDialog.value = true
  } catch (e) {
    console.error('Failed to load user details:', e)
    error.value = e.response?.data?.message || 'Failed to load user details'
  } finally {
    actionLoading.value = false
  }
}

async function toggleEmailVerification(user) {
  actionLoading.value = true
  error.value = null
  try {
    const newStatus = !user.email_verified
    await axios.post(`/api/admin/users/${user.id}/toggle-email-verification`, {
      verified: newStatus
    })
    successMessage.value = `Email ${newStatus ? 'verified' : 'unverified'} for ${user.email}`
    await loadUsers()
  } catch (e) {
    console.error('Failed to toggle email verification:', e)
    error.value = e.response?.data?.message || 'Failed to toggle email verification'
  } finally {
    actionLoading.value = false
  }
}

function formatFullDate(dateString) {
  if (!dateString) return 'Never'
  const date = new Date(dateString)
  return date.toLocaleString()
}

function previousPage() {
  if (offset.value >= limit.value) {
    offset.value -= limit.value
    loadUsers()
  }
}

function nextPage() {
  if (offset.value + limit.value < total.value) {
    offset.value += limit.value
    loadUsers()
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.gap-1 {
  gap: 4px;
}

.gap-2 {
  gap: 8px;
}

.white-space-pre-wrap {
  white-space: pre-wrap;
}
</style>
