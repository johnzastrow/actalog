<template>
  <v-container fluid class="pa-0" style="height: 100vh; overflow-y: auto; margin-top: 56px; margin-bottom: 70px; background-color: #f5f7fa;">
    <v-container style="max-width: 600px;" class="py-8">
      <v-card elevation="0" rounded="lg" class="pa-6">
        <!-- Loading State -->
        <div v-if="loading" class="text-center py-8">
          <v-progress-circular
            indeterminate
            color="#00bcd4"
            size="64"
            class="mb-4"
          />
          <h3 class="text-h5">Verifying your email...</h3>
          <p class="text-body-2 text-medium-emphasis mt-2">
            Please wait while we verify your email address.
          </p>
        </div>

        <!-- Success State -->
        <div v-else-if="success" class="text-center py-4">
          <v-icon color="success" size="80" class="mb-4">mdi-check-circle</v-icon>
          <h2 class="text-h4 mb-3" style="color: #2c3e50;">Email Verified!</h2>
          <p class="text-body-1 mb-4">
            Your email address has been successfully verified.
          </p>
          <p class="text-body-2 text-medium-emphasis mb-6">
            You can now access all features of ActaLog.
          </p>
          <v-btn
            color="primary"
            size="large"
            block
            @click="router.push('/dashboard')"
          >
            Go to Dashboard
          </v-btn>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="text-center py-4">
          <v-icon color="error" size="80" class="mb-4">mdi-alert-circle</v-icon>
          <h2 class="text-h4 mb-3" style="color: #2c3e50;">Verification Failed</h2>
          <v-alert type="error" class="mb-4 text-left">
            {{ errorMessage }}
          </v-alert>
          <div class="text-body-2 text-medium-emphasis mb-6">
            <p v-if="errorType === 'expired'">
              Your verification link has expired. Please request a new verification email.
            </p>
            <p v-else-if="errorType === 'invalid'">
              This verification link is invalid or has already been used.
            </p>
            <p v-else>
              There was a problem verifying your email address.
            </p>
          </div>
          <v-btn
            color="primary"
            variant="outlined"
            size="large"
            block
            @click="router.push('/resend-verification')"
            class="mb-3"
          >
            Resend Verification Email
          </v-btn>
          <v-btn
            color="primary"
            variant="text"
            size="large"
            block
            @click="router.push('/login')"
          >
            Go to Login
          </v-btn>
        </div>
      </v-card>
    </v-container>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const loading = ref(true)
const success = ref(false)
const error = ref(false)
const errorMessage = ref('')
const errorType = ref('')

async function verifyEmail() {
  const token = route.query.token

  if (!token) {
    error.value = true
    errorMessage.value = 'No verification token provided'
    errorType.value = 'invalid'
    loading.value = false
    return
  }

  try {
    const response = await axios.get(`/api/auth/verify-email?token=${token}`)

    if (response.status === 200) {
      success.value = true

      // Update user in auth store if logged in
      if (authStore.user) {
        authStore.user.email_verified = true
        localStorage.setItem('user', JSON.stringify(authStore.user))
      }
    }
  } catch (err) {
    error.value = true

    if (err.response?.status === 400) {
      if (err.response.data.error?.includes('expired')) {
        errorType.value = 'expired'
        errorMessage.value = 'Your verification link has expired'
      } else if (err.response.data.error?.includes('invalid') || err.response.data.error?.includes('not found')) {
        errorType.value = 'invalid'
        errorMessage.value = 'Invalid verification link'
      } else {
        errorType.value = 'unknown'
        errorMessage.value = err.response.data.error || 'Verification failed'
      }
    } else {
      errorType.value = 'unknown'
      errorMessage.value = 'An error occurred during verification. Please try again.'
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  verifyEmail()
})
</script>

<style scoped>
.v-container {
  background-color: #f5f7fa;
}

.v-card {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
</style>
