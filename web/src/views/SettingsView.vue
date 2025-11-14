<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" density="compact" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-btn icon="mdi-arrow-left" color="white" size="small" @click="$router.back()" />
      <v-toolbar-title class="text-white font-weight-bold">Settings</v-toolbar-title>
      <v-spacer />
    </v-app-bar>

    <v-container class="pa-2" style="margin-top: 36px; margin-bottom: 70px">
      <!-- Success/Error Alerts -->
      <v-alert
        v-if="successMessage"
        type="success"
        closable
        @click:close="successMessage = ''"
        class="mb-3"
      >
        {{ successMessage }}
      </v-alert>

      <v-alert
        v-if="errors.general"
        type="error"
        closable
        @click:close="errors.general = ''"
        class="mb-3"
      >
        {{ errors.general }}
      </v-alert>

      <!-- Profile Information Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Profile Information</h2>
        <v-form @submit.prevent="updateProfile">
          <v-text-field
            v-model="profileForm.name"
            label="Name"
            variant="outlined"
            density="compact"
            rounded="lg"
            :error-messages="errors.name"
            class="mb-2"
          >
            <template #prepend-inner>
              <v-icon color="#00bcd4" size="small">mdi-account</v-icon>
            </template>
          </v-text-field>

          <v-text-field
            v-model="profileForm.email"
            label="Email"
            type="email"
            variant="outlined"
            density="compact"
            rounded="lg"
            :error-messages="errors.email"
            hint="Changing your email will require re-verification"
            persistent-hint
            class="mb-2"
          >
            <template #prepend-inner>
              <v-icon color="#00bcd4" size="small">mdi-email</v-icon>
            </template>
          </v-text-field>

          <v-text-field
            v-model="profileForm.birthday"
            label="Birthday"
            type="date"
            variant="outlined"
            density="compact"
            rounded="lg"
            :error-messages="errors.birthday"
            class="mb-3"
          >
            <template #prepend-inner>
              <v-icon color="#00bcd4" size="small">mdi-cake-variant</v-icon>
            </template>
          </v-text-field>

          <v-btn
            type="submit"
            color="#00bcd4"
            block
            rounded="lg"
            :loading="loading"
            style="text-transform: none; font-weight: 600"
          >
            Update Profile
          </v-btn>
        </v-form>
      </v-card>

      <!-- Password Change Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Change Password</h2>
        <v-form @submit.prevent="changePassword">
          <v-text-field
            v-model="passwordForm.currentPassword"
            label="Current Password"
            type="password"
            variant="outlined"
            density="compact"
            rounded="lg"
            :error-messages="passwordErrors.currentPassword"
            class="mb-2"
          >
            <template #prepend-inner>
              <v-icon color="#00bcd4" size="small">mdi-lock</v-icon>
            </template>
          </v-text-field>

          <v-text-field
            v-model="passwordForm.newPassword"
            label="New Password"
            type="password"
            variant="outlined"
            density="compact"
            rounded="lg"
            :error-messages="passwordErrors.newPassword"
            hint="At least 8 characters"
            persistent-hint
            class="mb-2"
          >
            <template #prepend-inner>
              <v-icon color="#00bcd4" size="small">mdi-lock-outline</v-icon>
            </template>
          </v-text-field>

          <v-text-field
            v-model="passwordForm.confirmPassword"
            label="Confirm New Password"
            type="password"
            variant="outlined"
            density="compact"
            rounded="lg"
            :error-messages="passwordErrors.confirmPassword"
            class="mb-3"
          >
            <template #prepend-inner>
              <v-icon color="#00bcd4" size="small">mdi-lock-check</v-icon>
            </template>
          </v-text-field>

          <v-btn
            type="submit"
            color="#00bcd4"
            block
            rounded="lg"
            :loading="passwordLoading"
            style="text-transform: none; font-weight: 600"
          >
            Change Password
          </v-btn>
        </v-form>
      </v-card>

      <!-- Preferences Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Preferences</h2>
        <v-list bg-color="transparent" density="compact">
          <v-list-item>
            <template #prepend>
              <v-icon color="#00bcd4">mdi-theme-light-dark</v-icon>
            </template>
            <v-list-item-title class="font-weight-medium" style="color: #1a1a1a">
              Dark Mode
            </v-list-item-title>
            <template #append>
              <v-switch
                v-model="darkMode"
                color="#00bcd4"
                hide-details
                density="compact"
                @change="saveDarkMode"
              />
            </template>
          </v-list-item>

          <v-list-item>
            <template #prepend>
              <v-icon color="#00bcd4">mdi-bell</v-icon>
            </template>
            <v-list-item-title class="font-weight-medium" style="color: #1a1a1a">
              Notifications
            </v-list-item-title>
            <template #append>
              <v-switch
                v-model="notifications"
                color="#00bcd4"
                hide-details
                density="compact"
                @change="saveNotifications"
              />
            </template>
          </v-list-item>

          <v-list-item>
            <template #prepend>
              <v-icon color="#00bcd4">mdi-weight</v-icon>
            </template>
            <v-list-item-title class="font-weight-medium" style="color: #1a1a1a">
              Weight Unit
            </v-list-item-title>
            <template #append>
              <v-select
                v-model="weightUnit"
                :items="['lbs', 'kg']"
                variant="outlined"
                density="compact"
                hide-details
                style="max-width: 100px"
                @update:model-value="saveWeightUnit"
              />
            </template>
          </v-list-item>
        </v-list>
      </v-card>

      <!-- Data Management Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-2" style="color: #1a1a1a">Data Management</h2>
        <v-list bg-color="transparent" density="compact">
          <v-list-item
            prepend-icon="mdi-download"
            rounded="lg"
            style="cursor: pointer"
            @click="exportData"
          >
            <v-list-item-title class="font-weight-medium" style="color: #1a1a1a">
              Export Data
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption" style="color: #666">
              Download your workout history
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="#ccc" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-upload"
            rounded="lg"
            style="cursor: pointer"
            @click="importData"
          >
            <v-list-item-title class="font-weight-medium" style="color: #1a1a1a">
              Import Data
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption" style="color: #666">
              Restore from backup
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="#ccc" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>
        </v-list>
      </v-card>

      <!-- Danger Zone Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-2" style="color: #e91e63">Danger Zone</h2>
        <v-list bg-color="transparent" density="compact">
          <v-list-item
            prepend-icon="mdi-delete-forever"
            rounded="lg"
            style="cursor: pointer"
            @click="confirmDeleteAccount"
          >
            <v-list-item-title class="font-weight-medium" style="color: #e91e63">
              Delete Account
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption" style="color: #666">
              Permanently delete your account and all data
            </v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </v-card>

      <!-- App Info Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 text-center" style="background: white">
        <div class="text-caption" style="color: #999">
          ActaLog v0.4.0-beta
        </div>
        <div class="text-caption mt-1" style="color: #999">
          © 2025 ActaLog. All rights reserved.
        </div>
      </v-card>
    </v-container>

    <!-- Delete Account Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="text-h6" style="color: #e91e63">Delete Account?</v-card-title>
        <v-card-text>
          <p style="color: #666">
            This action cannot be undone. All your workouts, personal records, and account data will be permanently deleted.
          </p>
          <v-text-field
            v-model="deleteConfirmation"
            label='Type "DELETE" to confirm'
            variant="outlined"
            density="compact"
            class="mt-3"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn
            color="#e91e63"
            variant="flat"
            :disabled="deleteConfirmation !== 'DELETE'"
            :loading="deleteLoading"
            @click="deleteAccount"
          >
            Delete Account
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Bottom Navigation -->
    <v-bottom-navigation
      v-model="activeTab"
      grow
      style="position: fixed; bottom: 0; background: white"
      elevation="8"
    >
      <v-btn value="dashboard" to="/dashboard">
        <v-icon>mdi-view-dashboard</v-icon>
        <span style="font-size: 10px">Dashboard</span>
      </v-btn>
      <v-btn value="performance" to="/performance">
        <v-icon>mdi-chart-line</v-icon>
        <span style="font-size: 10px">Performance</span>
      </v-btn>
      <v-btn value="log" to="/workouts/log" style="position: relative; bottom: 20px">
        <v-avatar color="#ffc107" size="56" style="box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2)">
          <v-icon color="white" size="32">mdi-plus</v-icon>
        </v-avatar>
      </v-btn>
      <v-btn value="workouts" to="/workouts">
        <v-icon>mdi-dumbbell</v-icon>
        <span style="font-size: 10px">Templates</span>
      </v-btn>
      <v-btn value="profile" to="/profile">
        <v-icon>mdi-account</v-icon>
        <span style="font-size: 10px">Profile</span>
      </v-btn>
    </v-bottom-navigation>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const router = useRouter()
const authStore = useAuthStore()
const activeTab = ref('profile')

// State
const darkMode = ref(false)
const notifications = ref(true)
const weightUnit = ref('lbs')

const profileForm = ref({
  name: '',
  email: '',
  birthday: ''
})

const passwordForm = ref({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const loading = ref(false)
const passwordLoading = ref(false)
const deleteLoading = ref(false)
const successMessage = ref('')
const errors = ref({})
const passwordErrors = ref({})
const deleteDialog = ref(false)
const deleteConfirmation = ref('')

// Load current user data and preferences
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

  // Load preferences from localStorage
  darkMode.value = localStorage.getItem('darkMode') === 'true'
  notifications.value = localStorage.getItem('notifications') !== 'false'
  weightUnit.value = localStorage.getItem('weightUnit') || 'lbs'
})

// Update profile
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

// Change password
const changePassword = async () => {
  passwordErrors.value = {}
  successMessage.value = ''

  // Validation
  if (!passwordForm.value.currentPassword) {
    passwordErrors.value.currentPassword = 'Current password is required'
    return
  }

  if (!passwordForm.value.newPassword) {
    passwordErrors.value.newPassword = 'New password is required'
    return
  }

  if (passwordForm.value.newPassword.length < 8) {
    passwordErrors.value.newPassword = 'Password must be at least 8 characters'
    return
  }

  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    passwordErrors.value.confirmPassword = 'Passwords do not match'
    return
  }

  passwordLoading.value = true

  try {
    const response = await axios.put('/api/users/password', {
      current_password: passwordForm.value.currentPassword,
      new_password: passwordForm.value.newPassword
    })

    if (response.status === 200) {
      successMessage.value = 'Password changed successfully!'
      // Clear form
      passwordForm.value = {
        currentPassword: '',
        newPassword: '',
        confirmPassword: ''
      }
    }
  } catch (error) {
    if (error.response?.status === 401) {
      passwordErrors.value.currentPassword = 'Current password is incorrect'
    } else if (error.response?.status === 400) {
      errors.value.general = error.response.data.message || 'Invalid input'
    } else {
      errors.value.general = 'Failed to change password. Please try again.'
    }
  } finally {
    passwordLoading.value = false
  }
}

// Save preferences
const saveDarkMode = () => {
  localStorage.setItem('darkMode', darkMode.value.toString())
  // TODO: Apply theme change
}

const saveNotifications = () => {
  localStorage.setItem('notifications', notifications.value.toString())
}

const saveWeightUnit = () => {
  localStorage.setItem('weightUnit', weightUnit.value)
}

// Data management
const exportData = async () => {
  try {
    const response = await axios.get('/api/export', { responseType: 'blob' })
    const blob = new Blob([response.data], { type: 'application/json' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `actalog-export-${new Date().toISOString().split('T')[0]}.json`
    link.click()
    window.URL.revokeObjectURL(url)
    successMessage.value = 'Data exported successfully!'
  } catch (error) {
    errors.value.general = 'Failed to export data. Please try again.'
  }
}

const importData = () => {
  // TODO: Implement import functionality
  errors.value.general = 'Import functionality coming soon!'
}

// Account deletion
const confirmDeleteAccount = () => {
  deleteDialog.value = true
  deleteConfirmation.value = ''
}

const deleteAccount = async () => {
  deleteLoading.value = true

  try {
    await axios.delete('/api/users/account')
    authStore.logout()
    router.push('/login')
  } catch (error) {
    errors.value.general = 'Failed to delete account. Please try again.'
    deleteDialog.value = false
  } finally {
    deleteLoading.value = false
  }
}
</script>
