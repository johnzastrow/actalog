<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <!-- Header -->
      <div class="d-flex align-center mb-4">
        <v-btn icon @click="$router.back()" class="mr-2">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <div>
          <h1 class="text-h5">Send Announcement</h1>
          <div class="text-body-2 text-medium-emphasis">Send notifications to users with markdown support</div>
        </div>
      </div>

      <!-- Success Alert -->
      <v-alert v-if="successMessage" type="success" variant="tonal" closable @click:close="successMessage = null" class="mb-4">
        {{ successMessage }}
      </v-alert>

      <!-- Error Alert -->
      <v-alert v-if="errorMessage" type="error" variant="tonal" closable @click:close="errorMessage = null" class="mb-4">
        {{ errorMessage }}
      </v-alert>

      <!-- Announcement Form -->
      <v-card elevation="0" rounded="lg">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-bullhorn</v-icon>
          Create Announcement
        </v-card-title>

        <v-divider />

        <v-card-text class="pa-4">
          <v-form ref="form" @submit.prevent="sendAnnouncement">
            <!-- Title -->
            <v-text-field
              v-model="announcement.title"
              label="Title"
              placeholder="Enter announcement title"
              variant="outlined"
              density="comfortable"
              :rules="[v => !!v || 'Title is required']"
              class="mb-4"
            />

            <!-- Message with Tabs for Edit/Preview -->
            <v-card variant="outlined" class="mb-4">
              <v-tabs v-model="messageTab" color="primary">
                <v-tab value="edit">
                  <v-icon class="mr-2">mdi-pencil</v-icon>
                  Edit
                </v-tab>
                <v-tab value="preview">
                  <v-icon class="mr-2">mdi-eye</v-icon>
                  Preview
                </v-tab>
              </v-tabs>

              <v-divider />

              <v-window v-model="messageTab">
                <!-- Edit Tab -->
                <v-window-item value="edit">
                  <v-textarea
                    v-model="announcement.message"
                    placeholder="Enter announcement message (supports markdown)"
                    variant="plain"
                    :rules="[v => !!v || 'Message is required']"
                    rows="8"
                    class="pa-4"
                    hide-details
                  />
                  <v-divider />
                  <div class="text-caption text-medium-emphasis pa-2 px-4">
                    Supports markdown: **bold**, *italic*, [links](url), lists, etc.
                  </div>
                </v-window-item>

                <!-- Preview Tab -->
                <v-window-item value="preview">
                  <div class="pa-4" style="min-height: 200px;">
                    <MarkdownRenderer v-if="announcement.message" :content="announcement.message" />
                    <div v-else class="text-medium-emphasis text-center py-8">
                      No content to preview
                    </div>
                  </div>
                </v-window-item>
              </v-window>
            </v-card>

            <!-- Target Type -->
            <v-select
              v-model="announcement.targetType"
              label="Send To"
              :items="targetOptions"
              variant="outlined"
              density="comfortable"
              :rules="[v => !!v || 'Target type is required']"
              class="mb-4"
            />

            <!-- Organization Selector (for organization target type) -->
            <v-autocomplete
              v-if="announcement.targetType === 'organization'"
              v-model="announcement.targetIDs"
              label="Select Organizations"
              :items="organizations"
              item-title="name"
              item-value="id"
              variant="outlined"
              density="comfortable"
              multiple
              chips
              closable-chips
              :loading="loadingOrganizations"
              :rules="[v => (v && v.length > 0) || 'At least one organization is required']"
              class="mb-4"
            />

            <!-- User Selector (for users target type) -->
            <v-autocomplete
              v-if="announcement.targetType === 'users'"
              v-model="announcement.targetIDs"
              label="Select Users"
              :items="users"
              item-title="email"
              item-value="id"
              variant="outlined"
              density="comfortable"
              multiple
              chips
              closable-chips
              :loading="loadingUsers"
              :rules="[v => (v && v.length > 0) || 'At least one user is required']"
              class="mb-4"
            />

            <!-- Organization Context (Optional) -->
            <v-autocomplete
              v-model="announcement.organizationID"
              label="Organization Context (Optional)"
              :items="organizations"
              item-title="name"
              item-value="id"
              variant="outlined"
              density="comfortable"
              clearable
              :loading="loadingOrganizations"
              class="mb-4"
              hint="Associate this announcement with a specific organization"
              persistent-hint
            />

            <!-- Actions -->
            <v-btn
              type="submit"
              color="primary"
              size="large"
              block
              :loading="sending"
              :disabled="sending"
              prepend-icon="mdi-send"
            >
              Send Announcement
            </v-btn>
          </v-form>
        </v-card-text>
      </v-card>
    </v-container>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

const router = useRouter()

// State
const messageTab = ref('edit')
const announcement = ref({
  title: '',
  message: '',
  targetType: 'all',
  targetIDs: [],
  organizationID: null
})

const targetOptions = [
  { title: 'All Users', value: 'all' },
  { title: 'Specific Organizations', value: 'organization' },
  { title: 'Specific Users', value: 'users' }
]

const organizations = ref([])
const users = ref([])
const loadingOrganizations = ref(false)
const loadingUsers = ref(false)
const sending = ref(false)
const successMessage = ref(null)
const errorMessage = ref(null)
const form = ref(null)

// Methods
const loadOrganizations = async () => {
  try {
    loadingOrganizations.value = true
    const response = await axios.get('/api/admin/organizations')
    organizations.value = response.data.organizations || []
  } catch (error) {
    console.error('Failed to load organizations:', error)
    errorMessage.value = 'Failed to load organizations'
  } finally {
    loadingOrganizations.value = false
  }
}

const loadUsers = async () => {
  try {
    loadingUsers.value = true
    const response = await axios.get('/api/admin/users')
    users.value = response.data.users || []
  } catch (error) {
    console.error('Failed to load users:', error)
    errorMessage.value = 'Failed to load users'
  } finally {
    loadingUsers.value = false
  }
}

const sendAnnouncement = async () => {
  // Validate form
  const { valid } = await form.value.validate()
  if (!valid) return

  try {
    sending.value = true
    errorMessage.value = null
    successMessage.value = null

    const payload = {
      title: announcement.value.title,
      message: announcement.value.message,
      target_type: announcement.value.targetType,
      target_ids: announcement.value.targetIDs || [],
      organization_id: announcement.value.organizationID || null
    }

    await axios.post('/api/admin/notifications/announce', payload)

    successMessage.value = 'Announcement sent successfully!'

    // Reset form
    announcement.value = {
      title: '',
      message: '',
      targetType: 'all',
      targetIDs: [],
      organizationID: null
    }
    form.value.reset()
    messageTab.value = 'edit'

  } catch (error) {
    console.error('Failed to send announcement:', error)
    errorMessage.value = error.response?.data?.error || 'Failed to send announcement'
  } finally {
    sending.value = false
  }
}

// Lifecycle
onMounted(() => {
  loadOrganizations()
  loadUsers()
})
</script>

<style scoped>
.mobile-view-wrapper {
  min-height: 100vh;
  background: #f5f7fa;
}
</style>
