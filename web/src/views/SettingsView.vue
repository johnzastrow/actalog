<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <v-btn icon="mdi-arrow-left" variant="text" @click="$router.back()" />
        <h1 class="text-h4 mb-6 d-inline ml-2">Settings</h1>
      </v-col>

      <v-col cols="12">
        <!-- Profile Editing Section -->
        <v-card elevation="2" rounded="lg" class="mb-4">
          <v-card-title>Profile Information</v-card-title>
          <v-card-text>
            <v-form @submit.prevent="updateProfile">
              <v-text-field
                v-model="profileForm.name"
                label="Name"
                prepend-inner-icon="mdi-account"
                :error-messages="errors.name"
              />

              <v-text-field
                v-model="profileForm.email"
                label="Email"
                type="email"
                prepend-inner-icon="mdi-email"
                :error-messages="errors.email"
                hint="Changing your email will require re-verification"
              />

              <v-text-field
                v-model="profileForm.birthday"
                label="Birthday"
                type="date"
                prepend-inner-icon="mdi-cake-variant"
                :error-messages="errors.birthday"
              />

              <v-alert
                v-if="successMessage"
                type="success"
                class="mt-4"
                closable
                @click:close="successMessage = ''"
              >
                {{ successMessage }}
              </v-alert>

              <v-alert
                v-if="errors.general"
                type="error"
                class="mt-4"
                closable
                @click:close="errors.general = ''"
              >
                {{ errors.general }}
              </v-alert>

              <v-btn
                type="submit"
                color="primary"
                block
                size="large"
                class="mt-4"
                :loading="loading"
              >
                Update Profile
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>

        <v-card elevation="2" rounded="lg" class="mb-4">
          <v-card-title>Preferences</v-card-title>
          <v-card-text>
            <v-switch
              v-model="darkMode"
              label="Dark Mode"
              color="primary"
            />
            <v-switch
              v-model="notifications"
              label="Notifications"
              color="primary"
            />
          </v-card-text>
        </v-card>

        <v-card elevation="2" rounded="lg" class="mb-4">
          <v-card-title>Data</v-card-title>
          <v-card-text>
            <v-list>
              <v-list-item prepend-icon="mdi-download">
                <v-list-item-title>Export Data</v-list-item-title>
              </v-list-item>
              <v-list-item prepend-icon="mdi-upload">
                <v-list-item-title>Import Data</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>

        <v-card elevation="2" rounded="lg">
          <v-card-title>Account</v-card-title>
          <v-card-text>
            <v-list>
              <v-list-item prepend-icon="mdi-delete" class="text-error">
                <v-list-item-title>Delete Account</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()

const darkMode = ref(false)
const notifications = ref(true)

const profileForm = ref({
  name: '',
  email: '',
  birthday: ''
})

const loading = ref(false)
const successMessage = ref('')
const errors = ref({})

// Load current user data
onMounted(() => {
  if (authStore.user) {
    profileForm.value.name = authStore.user.name || ''
    profileForm.value.email = authStore.user.email || ''

    // Format birthday if it exists (from ISO to YYYY-MM-DD)
    if (authStore.user.birthday) {
      const date = new Date(authStore.user.birthday)
      profileForm.value.birthday = date.toISOString().split('T')[0]
    }
  }
})

const updateProfile = async () => {
  errors.value = {}
  successMessage.value = ''

  // Basic validation
  if (!profileForm.value.name) {
    errors.value.name = 'Name is required'
    return
  }

  if (!profileForm.value.email) {
    errors.value.email = 'Email is required'
    return
  }

  loading.value = true

  try {
    const response = await axios.put('/api/users/profile', {
      name: profileForm.value.name,
      email: profileForm.value.email,
      birthday: profileForm.value.birthday || undefined
    })

    if (response.status === 200) {
      // Update the auth store with new user data
      authStore.user = response.data.user
      successMessage.value = 'Profile updated successfully!'

      // If email changed, show additional message
      if (response.data.user.email !== authStore.user.email) {
        successMessage.value += ' Please check your email to verify your new address.'
      }
    }
  } catch (error) {
    if (error.response?.status === 409) {
      errors.value.email = 'Email already in use by another account'
    } else if (error.response?.status === 400) {
      errors.value.general = error.response.data.message || 'Invalid input'
    } else {
      errors.value.general = 'Failed to update profile. Please try again.'
    }
  } finally {
    loading.value = false
  }
}
</script>
