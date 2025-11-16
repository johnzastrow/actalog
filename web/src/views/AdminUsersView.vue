<template>
  <v-container fluid class="pa-4">
    <!-- Header -->
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
        </template>

        <!-- Last Login Column -->
        <template #item.last_login_at="{ item }">
          <span v-if="item.last_login_at" class="text-body-2">
            {{ formatDate(item.last_login_at) }}
          </span>
          <span v-else class="text-body-2 text-medium-emphasis">Never</span>
        </template>

        <!-- Actions Column -->
        <template #item.actions="{ item }">
          <div class="d-flex gap-1">
            <!-- Unlock Button -->
            <v-tooltip text="Unlock Account" location="top">
              <template #activator="{ props }">
                <v-btn
                  v-bind="props"
                  icon
                  size="small"
                  variant="text"
                  :disabled="!isLocked(item)"
                  @click="unlockUser(item)"
                >
                  <v-icon>mdi-lock-open</v-icon>
                </v-btn>
              </template>
            </v-tooltip>

            <!-- Enable/Disable Button -->
            <v-tooltip :text="item.account_disabled ? 'Enable Account' : 'Disable Account'" location="top">
              <template #activator="{ props }">
                <v-btn
                  v-bind="props"
                  icon
                  size="small"
                  variant="text"
                  @click="item.account_disabled ? enableUser(item) : disableUser(item)"
                >
                  <v-icon>{{ item.account_disabled ? 'mdi-account-check' : 'mdi-account-cancel' }}</v-icon>
                </v-btn>
              </template>
            </v-tooltip>

            <!-- Change Role Button -->
            <v-tooltip text="Change Role" location="top">
              <template #activator="{ props }">
                <v-btn
                  v-bind="props"
                  icon
                  size="small"
                  variant="text"
                  @click="openRoleDialog(item)"
                >
                  <v-icon>mdi-shield-account</v-icon>
                </v-btn>
              </template>
            </v-tooltip>
          </div>
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
  </v-container>
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
const selectedUser = ref(null)
const disableReason = ref('')
const newRole = ref('user')

const headers = [
  { title: 'User', value: 'email', sortable: false },
  { title: 'Role', value: 'role', sortable: false },
  { title: 'Status', value: 'status', sortable: false },
  { title: 'Last Login', value: 'last_login_at', sortable: false },
  { title: 'Actions', value: 'actions', sortable: false, align: 'end' }
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
</style>
