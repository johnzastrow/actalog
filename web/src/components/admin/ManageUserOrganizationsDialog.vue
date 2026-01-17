<template>
  <v-dialog v-model="dialog" max-width="600px" persistent>
    <v-card>
      <v-card-title class="d-flex align-center">
        <v-icon class="mr-2">mdi-domain</v-icon>
        Manage Organizations for {{ user?.email }}
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

        <!-- Add Organization Section -->
        <div class="mb-4">
          <h3 class="text-subtitle-1 mb-2">Add to Organization</h3>
          <v-row>
            <v-col cols="8">
              <v-autocomplete
                v-model="selectedOrganization"
                :items="availableOrganizations"
                item-title="name"
                item-value="id"
                label="Select organization"
                
                density="compact"
                :disabled="loading"
              >
                <template #item="{ item, props }">
                  <v-list-item v-bind="props">
                    <template #prepend>
                      <v-icon>mdi-domain</v-icon>
                    </template>
                    <v-list-item-title>{{ item.raw.name }}</v-list-item-title>
                    <v-list-item-subtitle v-if="item.raw.description">
                      {{ item.raw.description }}
                    </v-list-item-subtitle>
                  </v-list-item>
                </template>
              </v-autocomplete>
            </v-col>
            <v-col cols="4">
              <v-btn
                color="primary"
                block
                :disabled="!selectedOrganization || loading"
                @click="addOrganization"
              >
                <v-icon start>mdi-plus</v-icon>
                Add
              </v-btn>
            </v-col>
          </v-row>
        </div>

        <v-divider class="my-4" />

        <!-- Current Organizations List -->
        <div>
          <h3 class="text-subtitle-1 mb-2">Current Organizations ({{ userOrganizations.length }})</h3>

          <v-list v-if="userOrganizations.length > 0" class="pa-0">
            <v-list-item
              v-for="org in userOrganizations"
              :key="org.id"
              class="px-0"
            >
              <template #prepend>
                <v-icon color="primary">mdi-domain</v-icon>
              </template>

              <v-list-item-title>{{ org.name }}</v-list-item-title>
              <v-list-item-subtitle v-if="org.description">
                {{ org.description }}
              </v-list-item-subtitle>

              <template #append>
                <v-btn
                  icon
                  size="small"
                  variant="text"
                  color="error"
                  :disabled="loading"
                  @click="removeOrganization(org)"
                >
                  <v-icon>mdi-delete</v-icon>
                </v-btn>
              </template>
            </v-list-item>
          </v-list>

          <v-alert v-else type="info" variant="tonal" class="mt-2">
            User is not assigned to any organizations
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
  user: Object
})

const emit = defineEmits(['update:modelValue', 'updated'])

const dialog = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const loading = ref(false)
const error = ref(null)
const successMessage = ref(null)
const userOrganizations = ref([])
const allOrganizations = ref([])
const selectedOrganization = ref(null)

const availableOrganizations = computed(() => {
  // Filter out organizations the user is already in
  const userOrgIds = userOrganizations.value.map(org => org.id)
  return allOrganizations.value.filter(org => !userOrgIds.includes(org.id))
})

watch(() => props.user, (newUser) => {
  if (newUser && dialog.value) {
    loadData()
  }
})

watch(dialog, (newValue) => {
  if (newValue && props.user) {
    loadData()
  }
})

async function loadData() {
  if (!props.user) return

  loading.value = true
  error.value = null

  try {
    // Load user's organizations
    const userOrgsResponse = await axios.get(`/api/admin/users/${props.user.id}/organizations`)
    userOrganizations.value = userOrgsResponse.data.organizations || []

    // Load all organizations
    const allOrgsResponse = await axios.get('/api/admin/organizations?limit=1000')
    allOrganizations.value = allOrgsResponse.data.organizations || []
  } catch (err) {
    console.error('Failed to load organizations:', err)
    error.value = err.response?.data?.error || 'Failed to load organizations'
  } finally {
    loading.value = false
  }
}

async function addOrganization() {
  if (!selectedOrganization.value || !props.user) return

  loading.value = true
  error.value = null
  successMessage.value = null

  try {
    await axios.post(`/api/admin/users/${props.user.id}/organization`, {
      organization_id: selectedOrganization.value
    })

    successMessage.value = 'Organization added successfully'
    selectedOrganization.value = null

    // Reload user's organizations
    await loadData()
    emit('updated')
  } catch (err) {
    console.error('Failed to add organization:', err)
    error.value = err.response?.data?.error || 'Failed to add organization'
  } finally {
    loading.value = false
  }
}

async function removeOrganization(org) {
  if (!props.user) return

  loading.value = true
  error.value = null
  successMessage.value = null

  try {
    await axios.delete(`/api/admin/users/${props.user.id}/organization/${org.id}`)

    successMessage.value = `Removed from ${org.name}`

    // Reload user's organizations
    await loadData()
    emit('updated')
  } catch (err) {
    console.error('Failed to remove organization:', err)
    error.value = err.response?.data?.error || 'Failed to remove organization'
  } finally {
    loading.value = false
  }
}

function close() {
  dialog.value = false
  error.value = null
  successMessage.value = null
  selectedOrganization.value = null
}
</script>
