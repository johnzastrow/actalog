<template>
  <v-dialog :model-value="modelValue" max-width="600" @update:model-value="$emit('update:modelValue', $event)">
    <v-card>
      <v-card-title>
        <v-icon class="mr-2">mdi-plus-circle</v-icon>
        Create {{ type === 'user' ? 'User' : 'Organization' }} Subscription
      </v-card-title>

      <v-card-text>
        <v-form ref="formRef" @submit.prevent="createSubscription">
          <!-- User/Organization Selection -->
          <v-autocomplete
            v-if="type === 'user'"
            v-model="form.user_id"
            label="Select User"
            :items="users"
            item-title="email"
            item-value="id"
            variant="outlined"
            density="compact"
            :loading="loadingUsers"
            :rules="[v => !!v || 'User is required']"
            prepend-inner-icon="mdi-account"
            class="mb-3"
          >
            <template #item="{ props, item }">
              <v-list-item v-bind="props">
                <v-list-item-title>{{ item.raw.email }}</v-list-item-title>
                <v-list-item-subtitle>{{ item.raw.name }}</v-list-item-subtitle>
              </v-list-item>
            </template>
          </v-autocomplete>

          <v-autocomplete
            v-else
            v-model="form.organization_id"
            label="Select Organization"
            :items="organizations"
            item-title="name"
            item-value="id"
            variant="outlined"
            density="compact"
            :loading="loadingOrgs"
            :rules="[v => !!v || 'Organization is required']"
            prepend-inner-icon="mdi-office-building"
            class="mb-3"
          ></v-autocomplete>

          <!-- Subscription Type -->
          <v-select
            v-model="form.subscription_type"
            label="Subscription Type"
            :items="subscriptionTypes"
            variant="outlined"
            density="compact"
            :rules="[v => !!v || 'Subscription type is required']"
            prepend-inner-icon="mdi-tag"
            class="mb-3"
          ></v-select>

          <!-- Permanent Free Checkbox -->
          <v-checkbox
            v-model="form.is_permanent_free"
            label="Permanent Free (Never Expires)"
            hint="This subscription will never expire and never require payment"
            persistent-hint
            class="mb-3"
          ></v-checkbox>

          <!-- Notes -->
          <v-textarea
            v-model="form.notes"
            label="Admin Notes (Optional)"
            variant="outlined"
            density="compact"
            rows="3"
            hint="Internal notes about this subscription"
            persistent-hint
            prepend-inner-icon="mdi-note-text"
          ></v-textarea>
        </v-form>

        <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mt-3">
          {{ error }}
        </v-alert>
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          variant="text"
          @click="$emit('update:modelValue', false)"
        >
          Cancel
        </v-btn>
        <v-btn
          color="primary"
          :loading="loading"
          @click="createSubscription"
        >
          Create Subscription
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import axios from '@/utils/axios'

const props = defineProps({
  modelValue: Boolean,
  type: {
    type: String,
    required: true,
    validator: (value) => ['user', 'organization'].includes(value)
  }
})

const emit = defineEmits(['update:modelValue', 'created'])

const formRef = ref(null)
const loading = ref(false)
const loadingUsers = ref(false)
const loadingOrgs = ref(false)
const error = ref('')

const users = ref([])
const organizations = ref([])

const form = ref({
  user_id: null,
  organization_id: null,
  subscription_type: 'free',
  is_permanent_free: false,
  notes: ''
})

const subscriptionTypes = [
  { title: 'Free', value: 'free' },
  { title: 'Monthly', value: 'monthly' },
  { title: 'Annual', value: 'annual' }
]

// Load users when dialog opens (for user subscriptions)
watch(() => props.modelValue, (newValue) => {
  if (newValue) {
    if (props.type === 'user') {
      loadUsers()
    } else {
      loadOrganizations()
    }
    resetForm()
  }
})

async function loadUsers() {
  loadingUsers.value = true
  try {
    const response = await axios.get('/api/admin/users')
    users.value = response.data.users || []
  } catch (err) {
    console.error('Failed to load users:', err)
    error.value = 'Failed to load users'
  } finally {
    loadingUsers.value = false
  }
}

async function loadOrganizations() {
  loadingOrgs.value = true
  try {
    const response = await axios.get('/api/admin/organizations')
    organizations.value = response.data.organizations || []
  } catch (err) {
    console.error('Failed to load organizations:', err)
    error.value = 'Failed to load organizations'
  } finally {
    loadingOrgs.value = false
  }
}

async function createSubscription() {
  error.value = ''

  // Validate form
  const { valid } = await formRef.value.validate()
  if (!valid) return

  loading.value = true
  try {
    const endpoint = props.type === 'user'
      ? '/api/admin/subscriptions/user'
      : '/api/admin/subscriptions/organization'

    const payload = props.type === 'user'
      ? {
          user_id: form.value.user_id,
          subscription_type: form.value.subscription_type,
          is_permanent_free: form.value.is_permanent_free,
          notes: form.value.notes
        }
      : {
          organization_id: form.value.organization_id,
          subscription_type: form.value.subscription_type,
          is_permanent_free: form.value.is_permanent_free,
          notes: form.value.notes
        }

    await axios.post(endpoint, payload)

    emit('created')
    emit('update:modelValue', false)
  } catch (err) {
    console.error('Failed to create subscription:', err)
    error.value = err.response?.data?.error || 'Failed to create subscription'
  } finally {
    loading.value = false
  }
}

function resetForm() {
  form.value = {
    user_id: null,
    organization_id: null,
    subscription_type: 'free',
    is_permanent_free: false,
    notes: ''
  }
  error.value = ''
  if (formRef.value) {
    formRef.value.resetValidation()
  }
}
</script>
