<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <!-- Success/Error Alerts -->
      <v-alert
        v-if="successMessage"
        type="success"
        closable
        class="mb-3"
        @click:close="successMessage = ''"
      >
        {{ successMessage }}
      </v-alert>

      <v-alert
        v-if="errors.general"
        type="error"
        closable
        class="mb-3"
        @click:close="errors.general = ''"
      >
        {{ errors.general }}
      </v-alert>

      <!-- Profile Information Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" bg-color="surface">
        <h2 class="text-body-1 font-weight-bold mb-3" >Profile Information</h2>
        <v-form @submit.prevent="updateProfile">
          <v-text-field
            v-model="profileForm.name"
            label="Name"
            
            density="compact"
            rounded="lg"
            :error-messages="errors.name"
            class="mb-2"
          >
            <template #prepend-inner>
              <v-icon color="primary" size="small">mdi-account</v-icon>
            </template>
          </v-text-field>

          <v-text-field
            v-model="profileForm.email"
            label="Email"
            type="email"
            
            density="compact"
            rounded="lg"
            :error-messages="errors.email"
            hint="Changing your email will require re-verification"
            persistent-hint
            class="mb-2"
          >
            <template #prepend-inner>
              <v-icon color="primary" size="small">mdi-email</v-icon>
            </template>
          </v-text-field>

          <v-text-field
            v-model="profileForm.birthday"
            label="Birthday"
            type="date"
            
            density="compact"
            rounded="lg"
            :error-messages="errors.birthday"
            class="mb-3"
          >
            <template #prepend-inner>
              <v-icon color="primary" size="small">mdi-cake-variant</v-icon>
            </template>
          </v-text-field>

          <v-btn
            type="submit"
            color="primary"
            block
            rounded="lg"
            :loading="loading"
            style="text-transform: none; font-weight: 600"
          >
            Update Profile
          </v-btn>
        </v-form>
      </v-card>

      <!-- Subscription Status Card -->
      <subscription-status-badge class="mb-3" />

      <!-- Password Change Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" bg-color="surface">
        <h2 class="text-body-1 font-weight-bold mb-3" >Change Password</h2>
        <v-form @submit.prevent="changePassword">
          <v-text-field
            v-model="passwordForm.currentPassword"
            label="Current Password"
            type="password"
            
            density="compact"
            rounded="lg"
            :error-messages="passwordErrors.currentPassword"
            class="mb-2"
          >
            <template #prepend-inner>
              <v-icon color="primary" size="small">mdi-lock</v-icon>
            </template>
          </v-text-field>

          <v-text-field
            v-model="passwordForm.newPassword"
            label="New Password"
            type="password"
            
            density="compact"
            rounded="lg"
            :error-messages="passwordErrors.newPassword"
            hint="At least 12 characters with uppercase, lowercase, and a number"
            persistent-hint
            class="mb-2"
          >
            <template #prepend-inner>
              <v-icon color="primary" size="small">mdi-lock-outline</v-icon>
            </template>
          </v-text-field>

          <v-text-field
            v-model="passwordForm.confirmPassword"
            label="Confirm New Password"
            type="password"
            
            density="compact"
            rounded="lg"
            :error-messages="passwordErrors.confirmPassword"
            class="mb-3"
          >
            <template #prepend-inner>
              <v-icon color="primary" size="small">mdi-lock-check</v-icon>
            </template>
          </v-text-field>

          <v-btn
            type="submit"
            color="primary"
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
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" bg-color="surface">
        <h2 class="text-body-1 font-weight-bold mb-3" >Preferences</h2>

        <!-- Theme Selection -->
        <div class="mb-3">
          <div class="d-flex align-center mb-2">
            <v-icon color="primary" class="mr-2">mdi-palette</v-icon>
            <span class="font-weight-medium">Theme</span>
          </div>

          <!-- System Preference Toggle -->
          <v-list-item density="compact" class="px-0 mb-1">
            <template #prepend>
              <v-icon size="small">mdi-brightness-auto</v-icon>
            </template>
            <v-list-item-title class="text-body-2">Use System Theme</v-list-item-title>
            <template #append>
              <v-switch
                v-model="themeStore.useSystemPreference"
                color="primary"
                hide-details
                density="compact"
                @update:model-value="themeStore.setUseSystemPreference"
              />
            </template>
          </v-list-item>

          <!-- Theme Grid -->
          <v-row dense class="mt-1">
            <v-col
              v-for="theme in themeStore.availableThemes"
              :key="theme.id"
              cols="6"
            >
              <v-card
                :color="themeStore.currentTheme === theme.id ? 'primary' : 'surface-variant'"
                :variant="themeStore.currentTheme === theme.id ? 'flat' : 'outlined'"
                :disabled="themeStore.useSystemPreference"
                class="pa-2 text-center"
                style="cursor: pointer"
                @click="themeStore.setTheme(theme.id)"
              >
                <v-icon
                  :color="themeStore.currentTheme === theme.id ? 'white' : ''"
                  size="20"
                >
                  {{ theme.icon }}
                </v-icon>
                <div
                  class="text-caption mt-1"
                  :class="themeStore.currentTheme === theme.id ? 'text-white' : ''"
                >
                  {{ theme.name }}
                </div>
              </v-card>
            </v-col>
          </v-row>
        </div>

        <v-divider class="my-3" />

        <!-- Font Selection -->
        <div class="mb-3">
          <div class="d-flex align-center mb-2">
            <v-icon color="primary" class="mr-2">mdi-format-font</v-icon>
            <span class="font-weight-medium">Font</span>
          </div>
          <p class="text-caption text-medium-emphasis mb-2">
            Choose a font that works best for you
          </p>

          <!-- Accessibility fonts badge -->
          <v-chip
            v-if="fontStore.accessibilityFonts.length > 0"
            size="small"
            color="info"
            variant="tonal"
            class="mb-2"
            prepend-icon="mdi-human-greeting-proximity"
          >
            {{ fontStore.accessibilityFonts.length }} accessibility options
          </v-chip>

          <!-- Font Grid -->
          <v-row dense class="mt-1">
            <v-col
              v-for="font in fontStore.availableFonts"
              :key="font.id"
              cols="6"
            >
              <v-card
                :color="fontStore.currentFont === font.id ? 'primary' : 'surface-variant'"
                :variant="fontStore.currentFont === font.id ? 'flat' : 'outlined'"
                class="pa-2 text-center font-card"
                style="cursor: pointer; min-height: 70px"
                @click="changeFontFamily(font.id)"
              >
                <v-icon
                  :color="fontStore.currentFont === font.id ? 'white' : ''"
                  size="18"
                >
                  {{ font.icon }}
                </v-icon>
                <div
                  class="text-caption mt-1"
                  :class="fontStore.currentFont === font.id ? 'text-white' : ''"
                >
                  {{ font.name }}
                </div>
                <v-chip
                  v-if="font.accessibility"
                  size="x-small"
                  :color="fontStore.currentFont === font.id ? 'white' : 'info'"
                  :variant="fontStore.currentFont === font.id ? 'outlined' : 'tonal'"
                  class="mt-1"
                >
                  A11y
                </v-chip>
              </v-card>
            </v-col>
          </v-row>
        </div>

        <v-divider class="mb-3" />

        <v-list bg-color="transparent" density="compact">
          <v-list-item>
            <template #prepend>
              <v-icon color="primary">mdi-bell</v-icon>
            </template>
            <v-list-item-title class="font-weight-medium" >
              Notifications
            </v-list-item-title>
            <template #append>
              <v-switch
                v-model="notifications"
                color="primary"
                hide-details
                density="compact"
                @change="saveNotifications"
              />
            </template>
          </v-list-item>

          <v-list-item>
            <template #prepend>
              <v-icon color="primary">mdi-weight</v-icon>
            </template>
            <v-list-item-title class="font-weight-medium" >
              Weight Unit
            </v-list-item-title>
            <template #append>
              <v-select
                v-model="weightUnit"
                :items="['lbs', 'kg']"
                
                density="compact"
                hide-details
                style="max-width: 100px"
                @update:model-value="saveWeightUnit"
              />
            </template>
          </v-list-item>
        </v-list>

        <v-divider class="my-3" />

        <!-- Timezone Selection -->
        <div>
          <div class="d-flex align-center mb-2">
            <v-icon color="primary" class="mr-2">mdi-earth</v-icon>
            <span class="font-weight-medium">Timezone</span>
          </div>
          <p class="text-caption text-medium-emphasis mb-2">
            Set your timezone to display workout dates correctly
          </p>
          <v-autocomplete
            v-model="timezone"
            :items="timezoneOptions"
            item-title="label"
            item-value="value"
            
            density="compact"
            hide-details
            :loading="timezoneLoading"
            class="mb-2"
            @update:model-value="saveTimezone"
          />
          <v-btn
            variant="text"
            size="small"
            color="primary"
            prepend-icon="mdi-crosshairs-gps"
            @click="detectTimezone"
          >
            Detect browser timezone
          </v-btn>
        </div>
      </v-card>

      <!-- Admin Settings Card (only visible to admins) -->
      <v-card
        v-if="authStore.user?.role === 'admin'"
        elevation="0"
        rounded="lg"
        class="pa-3 mb-3"
        bg-color="surface"
      >
        <h2 class="text-body-1 font-weight-bold mb-3">Admin Settings</h2>
        <v-list bg-color="transparent" density="compact">
          <v-list-item>
            <template #prepend>
              <v-icon color="primary">mdi-email-alert</v-icon>
            </template>
            <v-list-item-title class="font-weight-medium">
              User Event Notifications
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-medium-emphasis">
              Receive email when users register or change
            </v-list-item-subtitle>
            <template #append>
              <v-switch
                v-model="adminUserEventNotifications"
                color="primary"
                hide-details
                density="compact"
                :loading="adminNotifLoading"
                @update:model-value="saveAdminUserEventNotifications"
              />
            </template>
          </v-list-item>
        </v-list>
      </v-card>

      <!-- Leaderboard Settings Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" bg-color="surface">
        <h2 class="text-body-1 font-weight-bold mb-3">Community</h2>
        <v-list bg-color="transparent" density="compact">
          <v-list-item>
            <template #prepend>
              <v-icon color="primary">mdi-podium</v-icon>
            </template>
            <v-list-item-title class="font-weight-medium">
              Show on Community Leaderboard
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-medium-emphasis">
              Let org members see your scores on leaderboards
            </v-list-item-subtitle>
            <template #append>
              <v-switch
                v-model="leaderboardOptIn"
                color="primary"
                hide-details
                density="compact"
                :loading="leaderboardLoading"
                @update:model-value="saveLeaderboardOptIn"
              />
            </template>
          </v-list-item>
        </v-list>
      </v-card>

      <!-- Data Management Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" bg-color="surface">
        <h2 class="text-body-1 font-weight-bold mb-2" >Data Management</h2>
        <v-list bg-color="transparent" density="compact">
          <v-list-item
            prepend-icon="mdi-download"
            rounded="lg"
            style="cursor: pointer"
            @click="exportData"
          >
            <v-list-item-title class="font-weight-medium" >
              Export Data
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-medium-emphasis">
              Download your workout history
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-upload"
            rounded="lg"
            style="cursor: pointer"
            @click="importData"
          >
            <v-list-item-title class="font-weight-medium" >
              Import Data
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-medium-emphasis">
              Restore from backup
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>
        </v-list>
      </v-card>

      <!-- Danger Zone Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" bg-color="surface">
        <h2 class="text-body-1 font-weight-bold mb-2" style="color: rgb(var(--v-theme-error))">Danger Zone</h2>
        <v-list bg-color="transparent" density="compact">
          <v-list-item
            prepend-icon="mdi-delete-forever"
            rounded="lg"
            style="cursor: pointer"
            @click="confirmDeleteAccount"
          >
            <v-list-item-title class="font-weight-medium" style="color: rgb(var(--v-theme-error))">
              Delete Account
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-medium-emphasis">
              Permanently delete your account and all data
            </v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </v-card>

      <!-- App Info Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 text-center" bg-color="surface">
        <div class="text-caption text-disabled">
          ActaLog v0.4.0-beta
        </div>
        <div class="text-caption mt-1 text-disabled">
          © 2025 ActaLog. All rights reserved.
        </div>
      </v-card>
    </v-container>

    <!-- Delete Account Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="text-h6" style="color: rgb(var(--v-theme-error))">Delete Account?</v-card-title>
        <v-card-text>
          <p class="text-medium-emphasis">
            This action cannot be undone. All your workouts, personal records, and account data will be permanently deleted.
          </p>
          <v-text-field
            v-model="deleteConfirmation"
            label='Type "DELETE" to confirm'
            
            density="compact"
            class="mt-3"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn
            color="error"
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

    <!-- Import Data Dialog -->
    <v-dialog v-model="importDialog" max-width="500" persistent>
      <v-card>
        <v-card-title class="d-flex align-center">
          <v-icon start color="primary">mdi-upload</v-icon>
          Import Data
        </v-card-title>

        <v-card-text>
          <!-- Step 1: File Upload -->
          <div v-if="importStep === 'upload'">
            <p class="text-medium-emphasis mb-3">
              Upload a JSON file exported from ActaLog to restore your workout history.
            </p>

            <v-file-input
              v-model="importFile"
              label="Select backup file"
              accept=".json"
              prepend-icon="mdi-file-upload"
              
              density="compact"
              :error-messages="importErrors.file"
              show-size
              @update:model-value="importErrors.file = ''"
            />

            <v-alert
              v-if="importErrors.general"
              type="error"
              density="compact"
              class="mt-2"
              closable
              @click:close="importErrors.general = ''"
            >
              {{ importErrors.general }}
            </v-alert>
          </div>

          <!-- Step 2: Preview -->
          <div v-else-if="importStep === 'preview'">
            <v-alert type="info" density="compact" class="mb-3">
              Review what will be imported
            </v-alert>

            <v-list density="compact" bg-color="surface-variant" rounded="lg">
              <v-list-item>
                <template #prepend>
                  <v-icon color="primary">mdi-dumbbell</v-icon>
                </template>
                <v-list-item-title>Workouts</v-list-item-title>
                <template #append>
                  <v-chip size="small" color="primary">{{ importPreview.total_workouts || 0 }}</v-chip>
                </template>
              </v-list-item>

              <v-list-item v-if="importPreview.duplicates > 0">
                <template #prepend>
                  <v-icon color="warning">mdi-content-duplicate</v-icon>
                </template>
                <v-list-item-title>Potential Duplicates</v-list-item-title>
                <template #append>
                  <v-chip size="small" color="warning">{{ importPreview.duplicates }}</v-chip>
                </template>
              </v-list-item>

              <v-list-item v-if="importPreview.date_range">
                <template #prepend>
                  <v-icon color="secondary">mdi-calendar-range</v-icon>
                </template>
                <v-list-item-title>Date Range</v-list-item-title>
                <v-list-item-subtitle>{{ importPreview.date_range }}</v-list-item-subtitle>
              </v-list-item>
            </v-list>

            <v-checkbox
              v-if="importPreview.duplicates > 0"
              v-model="skipDuplicates"
              label="Skip duplicate workouts"
              density="compact"
              hide-details
              class="mt-3"
            />

            <v-alert
              v-if="importErrors.general"
              type="error"
              density="compact"
              class="mt-3"
              closable
              @click:close="importErrors.general = ''"
            >
              {{ importErrors.general }}
            </v-alert>
          </div>

          <!-- Step 3: Success -->
          <div v-else-if="importStep === 'success'">
            <v-alert type="success" density="compact">
              <div class="font-weight-bold">Import Successful!</div>
              <div class="text-caption mt-1">
                {{ importResult.imported }} workout(s) imported successfully.
                <span v-if="importResult.skipped > 0">
                  {{ importResult.skipped }} duplicate(s) skipped.
                </span>
              </div>
            </v-alert>
          </div>
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <!-- Upload step actions -->
          <template v-if="importStep === 'upload'">
            <v-btn variant="text" @click="closeImportDialog">Cancel</v-btn>
            <v-btn
              color="primary"
              variant="flat"
              :loading="importLoading"
              :disabled="!importFile"
              @click="previewImport"
            >
              Preview
            </v-btn>
          </template>

          <!-- Preview step actions -->
          <template v-else-if="importStep === 'preview'">
            <v-btn variant="text" @click="importStep = 'upload'">Back</v-btn>
            <v-btn
              color="primary"
              variant="flat"
              :loading="importLoading"
              @click="confirmImport"
            >
              Import
            </v-btn>
          </template>

          <!-- Success step actions -->
          <template v-else-if="importStep === 'success'">
            <v-btn color="primary" variant="flat" @click="closeImportDialog">Done</v-btn>
          </template>
        </v-card-actions>
      </v-card>
    </v-dialog>

  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSubscriptionStore } from '@/stores/subscription'
import { useThemeStore } from '@/stores/theme'
import { useFontStore } from '@/stores/font'
import { useSettingsStore } from '@/stores/settings'
import axios from '@/utils/axios'
import { getTimezoneOptions, getBrowserTimezone } from '@/utils/timezone'
import SubscriptionStatusBadge from '@/components/SubscriptionStatusBadge.vue'

const router = useRouter()
const authStore = useAuthStore()
const subscriptionStore = useSubscriptionStore()
const themeStore = useThemeStore()
const fontStore = useFontStore()
const settingsStore = useSettingsStore()
const activeTab = ref('profile')

// State
const notifications = ref(true)
const weightUnit = ref('lbs')
const timezone = ref('America/New_York')
const timezoneOptions = ref(getTimezoneOptions())
const timezoneLoading = ref(false)
const adminUserEventNotifications = ref(true)
const adminNotifLoading = ref(false)
const leaderboardOptIn = ref(false)
const leaderboardLoading = ref(false)

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

// Import state
const importDialog = ref(false)
const importStep = ref('upload') // 'upload', 'preview', 'success'
const importFile = ref(null)
const importLoading = ref(false)
const importErrors = ref({ file: '', general: '' })
const importPreview = ref({})
const importResult = ref({})
const skipDuplicates = ref(true)

// Load current user data and preferences
onMounted(async () => {
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
  notifications.value = localStorage.getItem('notifications') !== 'false'
  weightUnit.value = localStorage.getItem('weightUnit') || 'lbs'

  // Fetch user settings from server
  try {
    await settingsStore.fetchSettings()
    timezone.value = settingsStore.timezone
    weightUnit.value = settingsStore.weightUnit
    adminUserEventNotifications.value = settingsStore.adminUserEventNotifications
    leaderboardOptIn.value = settingsStore.leaderboardOptIn
  } catch (err) {
    console.error('Failed to fetch user settings:', err)
  }

  // Fetch subscription status
  subscriptionStore.fetchStatus().catch(err => {
    console.error('Failed to fetch subscription status:', err)
  })
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

  if (passwordForm.value.newPassword.length < 12) {
    passwordErrors.value.newPassword = 'Password must be at least 12 characters'
    return
  }

  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    passwordErrors.value.confirmPassword = 'Passwords do not match'
    return
  }

  passwordLoading.value = true

  try {
    const response = await axios.put('/api/users/password', {
      old_password: passwordForm.value.currentPassword,
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
const saveNotifications = () => {
  localStorage.setItem('notifications', notifications.value.toString())
}

const saveWeightUnit = async () => {
  localStorage.setItem('weightUnit', weightUnit.value)
  try {
    await settingsStore.updateWeightUnit(weightUnit.value)
  } catch (err) {
    console.error('Failed to save weight unit:', err)
  }
}

const saveTimezone = async () => {
  timezoneLoading.value = true
  try {
    await settingsStore.updateTimezone(timezone.value)
    successMessage.value = 'Timezone updated successfully!'
  } catch {
    errors.value.general = 'Failed to save timezone. Please try again.'
  } finally {
    timezoneLoading.value = false
  }
}

const detectTimezone = () => {
  const browserTz = getBrowserTimezone()
  timezone.value = browserTz
  // Update the options list if browser timezone isn't in the list
  timezoneOptions.value = getTimezoneOptions()
  saveTimezone()
}

const changeFontFamily = async (fontId) => {
  try {
    await settingsStore.updateFontFamily(fontId)
    successMessage.value = 'Font updated!'
  } catch (err) {
    errors.value.general = 'Failed to save font preference.'
    console.error('Failed to save font:', err)
  }
}

// Admin settings
const saveAdminUserEventNotifications = async (enabled) => {
  adminNotifLoading.value = true
  try {
    await settingsStore.updateAdminUserEventNotifications(enabled)
    successMessage.value = enabled
      ? 'You will receive admin notifications'
      : 'Admin notifications disabled'
  } catch {
    errors.value.general = 'Failed to save notification preference. Please try again.'
    // Revert the toggle
    adminUserEventNotifications.value = !enabled
  } finally {
    adminNotifLoading.value = false
  }
}

// Leaderboard opt-in
const saveLeaderboardOptIn = async (enabled) => {
  leaderboardLoading.value = true
  try {
    await settingsStore.updateLeaderboardOptIn(enabled)
    successMessage.value = enabled
      ? 'You will appear on community leaderboards'
      : 'Removed from community leaderboards'
  } catch {
    errors.value.general = 'Failed to save leaderboard preference. Please try again.'
    leaderboardOptIn.value = !enabled
  } finally {
    leaderboardLoading.value = false
  }
}

// Data management
const exportData = async () => {
  try {
    const response = await axios.get('/api/export/user-workouts', { responseType: 'blob' })
    const blob = new Blob([response.data], { type: 'application/json' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `actalog-export-${new Date().toISOString().split('T')[0]}.json`
    link.click()
    window.URL.revokeObjectURL(url)
    successMessage.value = 'Data exported successfully!'
  } catch {
    errors.value.general = 'Failed to export data. Please try again.'
  }
}

const importData = () => {
  // Reset import state and open dialog
  importStep.value = 'upload'
  importFile.value = null
  importErrors.value = { file: '', general: '' }
  importPreview.value = {}
  importResult.value = {}
  skipDuplicates.value = true
  importDialog.value = true
}

const closeImportDialog = () => {
  importDialog.value = false
  importFile.value = null
  importErrors.value = { file: '', general: '' }
}

const previewImport = async () => {
  if (!importFile.value) {
    importErrors.value.file = 'Please select a file'
    return
  }

  importLoading.value = true
  importErrors.value = { file: '', general: '' }

  try {
    const formData = new FormData()
    formData.append('file', importFile.value)

    const response = await axios.post('/api/import/user-workouts/preview', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })

    importPreview.value = response.data
    importStep.value = 'preview'
  } catch (error) {
    importErrors.value.general = error.response?.data?.error || 'Failed to preview import. Please check the file format.'
  } finally {
    importLoading.value = false
  }
}

const confirmImport = async () => {
  importLoading.value = true
  importErrors.value.general = ''

  try {
    const formData = new FormData()
    formData.append('file', importFile.value)
    formData.append('skip_duplicates', skipDuplicates.value.toString())

    const response = await axios.post('/api/import/user-workouts/confirm', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })

    importResult.value = response.data
    importStep.value = 'success'
    successMessage.value = `Successfully imported ${response.data.imported || 0} workout(s)!`
  } catch (error) {
    importErrors.value.general = error.response?.data?.error || 'Import failed. Please try again.'
  } finally {
    importLoading.value = false
  }
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
  } catch {
    errors.value.general = 'Failed to delete account. Please try again.'
    deleteDialog.value = false
  } finally {
    deleteLoading.value = false
  }
}
</script>
