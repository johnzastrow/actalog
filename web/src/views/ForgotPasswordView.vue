<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="4">
        <v-card elevation="4" rounded="lg">
          <v-card-title class="text-h4 text-center pa-6">
            <div class="d-flex flex-column align-center">
              <div class="text-primary mb-2">Reset Password</div>
              <div class="text-body-2 text-medium-emphasis">
                Enter your email to receive a reset link
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

            <v-form @submit.prevent="handleSubmit" :disabled="submitted">
              <v-text-field
                v-model="email"
                label="Email"
                type="email"
                prepend-inner-icon="mdi-email"
                required
                :error-messages="errors.email"
                :disabled="submitted"
              />

              <v-btn
                type="submit"
                color="primary"
                block
                size="large"
                class="mt-6"
                :loading="loading"
                :disabled="submitted"
              >
                Send Reset Link
              </v-btn>

              <div class="text-center mt-4">
                <router-link to="/login" class="text-decoration-none">
                  <v-icon size="small" class="mr-1">mdi-arrow-left</v-icon>
                  Back to Sign In
                </router-link>
              </div>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'
import axios from '@/utils/axios'

const email = ref('')
const loading = ref(false)
const submitted = ref(false)
const errors = ref({})
const successMessage = ref('')
const errorMessage = ref('')

const handleSubmit = async () => {
  // Reset messages and errors
  errors.value = {}
  successMessage.value = ''
  errorMessage.value = ''
  loading.value = true

  try {
    const response = await axios.post('/api/auth/forgot-password', {
      email: email.value
    })

    successMessage.value = response.data.message || 'If your email is registered, you will receive a password reset link shortly'
    submitted.value = true
  } catch (error) {
    if (error.response?.data?.message) {
      errorMessage.value = error.response.data.message
    } else {
      errorMessage.value = 'Failed to send reset link. Please try again.'
    }
  } finally {
    loading.value = false
  }
}
</script>
