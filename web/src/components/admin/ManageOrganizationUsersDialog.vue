<template>
  <v-dialog v-model="dialog" max-width="700px" persistent>
    <v-card>
      <v-card-title class="d-flex align-center">
        <v-icon class="mr-2">mdi-account-multiple</v-icon>
        Manage Users in {{ organization?.name }}
        <v-spacer />
        <v-btn icon variant="text" @click="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <v-divider />

      <v-card-text class="pa-4">
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

        <!-- Add User Section -->
        <div class="mb-4">
          <h3 class="text-subtitle-1 mb-2">Add User</h3>
          <v-row>
            <v-col cols="8">
              <v-autocomplete
                v-model="selectedUser"
                :items="availableUsers"
                item-title="email"
                item-value="id"
                label="Select user"
                variant="outlined"
                density="compact"
                :disabled="loading"
              >
                <template #item="{ item, props }">
                  <v-list-item v-bind="props">
                    <template #prepend>
                      <v-avatar size="32" color="primary">
                        <span class="text-white">{{ item.raw.email[0].toUpperCase() }}</span>
                      </v-avatar>
                    </template>
                    <v-list-item-title>{{ item.raw.email }}</v-list-item-title>
                    <v-list-item-subtitle>{{ item.raw.name }}</v-list-item-subtitle>
                  </v-list-item>
                </template>
              </v-autocomplete>
            </v-col>
            <v-col cols="4">
              <v-btn
                color="primary"
                block
                :disabled="!selectedUser || loading"
                @click="addUser"
              >
                <v-icon start>mdi-plus</v-icon>
                Add
              </v-btn>
            </v-col>
          </v-row>
        </div>

        <v-divider class="my-4" />

        <!-- Current Users List -->
        <div>
          <h3 class="text-subtitle-1 mb-2">Current Users ({{ organizationUsers.length }})</h3>

          <v-list v-if="organizationUsers.length > 0" class="pa-0">
            <v-list-item
              v-for="user in organizationUsers"
              :key="user.id"
              class="px-0"
            >
              <template #prepend>
                <v-avatar size="40" color="primary">
                  <span class="text-white">{{ user.email[0].toUpperCase() }}</span>
                </v-avatar>
              </template>

              <v-list-item-title>{{ user.email }}</v-list-item-title>
              <v-list-item-subtitle>
                <div class="d-flex align-center gap-2">
                  <span>{{ user.name }}</span>
                  <v-chip size="x-small" :color="user.role === 'admin' ? 'error' : 'default'">
                    {{ user.role }}
                  </v-chip>
                </div>
              </v-list-item-subtitle>

              <template #append>
                <v-btn
                  icon
                  size="small"
                  variant="text"
                  color="error"
                  :disabled="loading"
                  @click="removeUser(user)"
                >
                  <v-icon>mdi-delete</v-icon>
                </v-btn>
              </template>
            </v-list-item>
          </v-list>

          <v-alert v-else type="info" variant="tonal" class="mt-2">
            No users assigned to this organization
          </v-alert>
        </div>
      </v-card-text>

      <v-divider />

      <v-card-actions>
        <v-spacer />
        <v-btn @click="close">Close</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import axios from '@/utils/axios'

const props = defineProps({
  modelValue: Boolean,
  organization: Object
})

const emit = defineEmits(['update:modelValue', 'updated'])

const dialog = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const loading = ref(false)
const error = ref(null)
const successMessage = ref(null)
const organizationUsers = ref([])
const allUsers = ref([])
const selectedUser = ref(null)

const availableUsers = computed(() => {
  // Filter out users already in the organization
  const orgUserIds = organizationUsers.value.map(user => user.id)
  return allUsers.value.filter(user => !orgUserIds.includes(user.id))
})

watch(() => props.organization, (newOrg) => {
  if (newOrg && dialog.value) {
    loadData()
  }
})

watch(dialog, (newValue) => {
  if (newValue && props.organization) {
    loadData()
  }
})

async function loadData() {
  if (!props.organization) return

  loading.value = true
  error.value = null

  try {
    // Load organization's users
    const orgUsersResponse = await axios.get(`/api/admin/organizations/${props.organization.id}/users`)
    organizationUsers.value = orgUsersResponse.data.users || []

    // Load all users (you might want to add pagination or search for large user bases)
    const allUsersResponse = await axios.get('/api/admin/users?limit=1000')
    allUsers.value = allUsersResponse.data.users || []
  } catch (err) {
    console.error('Failed to load users:', err)
    error.value = err.response?.data?.error || 'Failed to load users'
  } finally {
    loading.value = false
  }
}

async function addUser() {
  if (!selectedUser.value || !props.organization) return

  loading.value = true
  error.value = null
  successMessage.value = null

  try {
    await axios.post(`/api/admin/users/${selectedUser.value}/organization`, {
      organization_id: props.organization.id
    })

    successMessage.value = 'User added successfully'
    selectedUser.value = null

    // Reload organization's users
    await loadData()
    emit('updated')
  } catch (err) {
    console.error('Failed to add user:', err)
    error.value = err.response?.data?.error || 'Failed to add user'
  } finally {
    loading.value = false
  }
}

async function removeUser(user) {
  if (!props.organization) return

  loading.value = true
  error.value = null
  successMessage.value = null

  try {
    await axios.delete(`/api/admin/users/${user.id}/organization/${props.organization.id}`)

    successMessage.value = `Removed ${user.email}`

    // Reload organization's users
    await loadData()
    emit('updated')
  } catch (err) {
    console.error('Failed to remove user:', err)
    error.value = err.response?.data?.error || 'Failed to remove user'
  } finally {
    loading.value = false
  }
}

function close() {
  dialog.value = false
  error.value = null
  successMessage.value = null
  selectedUser.value = null
}
</script>
