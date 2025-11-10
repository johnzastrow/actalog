<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="4">
        <v-card elevation="4" rounded="lg">
          <v-card-title class="text-h4 text-center pa-6">
            <div class="d-flex flex-column align-center">
              <div class="text-primary mb-2">Set New Password</div>
              <div class="text-body-2 text-medium-emphasis">
                Enter your new password below
              </div>
            </div>
          </v-card-title>

          <v-card-text class="pa-6">
            <v-alert
              v-if="successMessage"
              type="success"
              variant="tonal"
              class="mb-4"
            >
              {{ successMessage }}
            </v-alert>

            <v-alert
              v-if="errorMessage"
              type="error"
              variant="tonal"
              class="mb-4"
            >
              {{ errorMessage }}
            </v-alert>

            <v-form v-if="!resetSuccess" @submit.prevent="handleSubmit">
              <v-text-field
                v-model="newPassword"
                label="New Password"
                type="password"
                prepend-inner-icon="mdi-lock"
                required
                :error-messages="errors.password"
                hint="Minimum 8 characters"
                persistent-hint
              />

              <v-text-field
                v-model="confirmPassword"
                label="Confirm Password"
                type="password"
                prepend-inner-icon="mdi-lock-check"
                required
                :error-messages="errors.confirmPassword"
                class="mt-4"
              />

              <v-btn
                type="submit"
                color="primary"
                block
                size="large"
                class="mt-6"
                :loading="loading"
              >
                Reset Password
              </v-btn>
            </v-form>

            <div v-if="resetSuccess" class="text-center">
              <v-btn
                to="/login"
                color="primary"
                size="large"
                prepend-icon="mdi-login"
              >
                Go to Sign In
              </v-btn>
            </div>

            <div v-if="!resetSuccess" class="text-center mt-4">
              <router-link to="/login" class="text-decoration-none">
                <v-icon size="small" class="mr-1">mdi-arrow-left</v-icon>
                Back to Sign In
              </router-link>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'

const route = useRoute()

const newPassword = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const resetSuccess = ref(false)
const errors = ref({})
const successMessage = ref('')
const errorMessage = ref('')
const token = ref('')

onMounted(() => {
  // Get token from URL parameter
  token.value = route.params.token || ''

  if (!token.value) {
    errorMessage.value = 'Invalid reset link. Please request a new password reset.'
  }
})

const handleSubmit = async () => {
  // Reset messages and errors
  errors.value = {}
  successMessage.value = ''
  errorMessage.value = ''

  // Validate passwords match
  if (newPassword.value !== confirmPassword.value) {
    errors.value.confirmPassword = 'Passwords do not match'
    return
  }

  // Validate password length
  if (newPassword.value.length < 8) {
    errors.value.password = 'Password must be at least 8 characters long'
    return
  }

  loading.value = true

  try {
    const response = await axios.post('/api/auth/reset-password', {
      token: token.value,
      new_password: newPassword.value
    })

    successMessage.value = response.data.message || 'Password has been reset successfully. You can now sign in with your new password.'
    resetSuccess.value = true
  } catch (error) {
    if (error.response?.data?.message) {
      errorMessage.value = error.response.data.message
    } else {
      errorMessage.value = 'Failed to reset password. Please try again or request a new reset link.'
    }
  } finally {
    loading.value = false
  }
}
</script>
