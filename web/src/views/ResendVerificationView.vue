<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="4">
        <v-card elevation="4" rounded="lg">
          <v-card-title class="text-h4 text-center pa-6">
            <div class="d-flex flex-column align-center">
              <div class="text-primary mb-2">ActaLog</div>
              <div class="text-body-2 text-medium-emphasis">
                {{ emailSent ? 'Email Sent' : 'Resend Verification' }}
              </div>
            </div>
          </v-card-title>

          <v-card-text class="pa-6">
            <!-- Success Message -->
            <div v-if="emailSent" class="text-center">
              <v-icon color="success" size="64" class="mb-4">mdi-email-check</v-icon>
              <h3 class="text-h5 mb-3">Verification Email Sent!</h3>
              <p class="text-body-1 mb-4">
                We've sent a new verification email to <strong>{{ email }}</strong>
              </p>
              <p class="text-body-2 text-medium-emphasis mb-6">
                Please check your email and click the verification link to activate your account.
                The link will expire in 24 hours.
              </p>
              <v-btn color="primary" block @click="router.push('/login')">
                Go to Login
              </v-btn>
            </div>

            <!-- Resend Form -->
            <div v-else>
              <p class="text-body-2 text-medium-emphasis mb-6">
                Enter your email address and we'll send you a new verification link.
              </p>

              <v-form @submit.prevent="handleResend">
                <v-text-field
                  v-model="email"
                  label="Email"
                  type="email"
                  prepend-inner-icon="mdi-email"
                  required
                  :error-messages="errors.email"
                />

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
                  class="mt-6"
                  :loading="loading"
                >
                  Resend Verification Email
                </v-btn>

                <div class="text-center mt-4">
                  <router-link to="/login" class="text-decoration-none">
                    Back to Login
                  </router-link>
                </div>
              </v-form>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

const router = useRouter()

const email = ref('')
const loading = ref(false)
const emailSent = ref(false)
const errors = ref({})

const handleResend = async () => {
  errors.value = {}

  if (!email.value) {
    errors.value.email = 'Email is required'
    return
  }

  loading.value = true

  try {
    const response = await axios.post('/api/auth/resend-verification', {
      email: email.value
    })

    if (response.status === 200) {
      emailSent.value = true
    }
  } catch (error) {
    if (error.response?.status === 404) {
      errors.value.general = 'No account found with this email address'
    } else if (error.response?.status === 400) {
      errors.value.general = error.response.data.error || 'Email is already verified or request failed'
    } else {
      errors.value.general = 'Failed to send verification email. Please try again.'
    }
  } finally {
    loading.value = false
  }
}
</script>
