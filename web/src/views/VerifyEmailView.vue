<template>
  <v-container fluid class="pa-0" style="height: 100vh; overflow-y: auto; margin-top: 36px; margin-bottom: 70px; background-color: #f5f7fa;">
    <v-container style="max-width: 600px;" class="py-8">
      <v-card elevation="0" rounded="lg" class="pa-6">
        <!-- Loading State -->
        <div v-if="loading" class="text-center py-8">
          <v-progress-circular
            indeterminate
            color="#00bcd4"
            size="64"
          ></v-progress-circular>
          <p class="mt-4 text-h6">Verifying your email...</p>
        </div>

        <!-- Success State -->
        <div v-else-if="success" class="text-center py-8">
          <v-icon size="80" color="success" class="mb-4">mdi-check-circle</v-icon>
          <h2 class="text-h4 mb-3" style="color: #1a1a1a;">Email Verified!</h2>
          <p class="text-body-1 mb-6" style="color: #666;">
            Your email has been successfully verified. You can now access all features of ActaLog.
          </p>
          <v-btn
            color="teal"
            size="large"
            rounded
            @click="goToLogin"
            class="text-none font-weight-bold"
            style="color: #1a1a1a;"
          >
            Go to Login
          </v-btn>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="text-center py-8">
          <v-icon size="80" color="error" class="mb-4">mdi-alert-circle</v-icon>
          <h2 class="text-h4 mb-3" style="color: #1a1a1a;">Verification Failed</h2>
          <p class="text-body-1 mb-6" style="color: #666;">
            {{ errorMessage }}
          </p>

          <!-- Show resend option if token expired -->
          <div v-if="showResend">
            <v-text-field
              v-model="email"
              label="Email Address"
              type="email"
              variant="outlined"
              prepend-inner-icon="mdi-email"
              class="mb-4"
              :disabled="resending"
            ></v-text-field>
            <v-btn
              color="#00bcd4"
              size="large"
              rounded
              @click="resendVerification"
              :loading="resending"
              :disabled="!email || resending"
              class="text-none font-weight-bold mb-3"
              block
            >
              Resend Verification Email
            </v-btn>
          </div>

          <v-btn
            variant="outlined"
            color="#00bcd4"
            size="large"
            rounded
            @click="goToLogin"
            class="text-none font-weight-bold"
          >
            Back to Login
          </v-btn>
        </div>
      </v-card>
    </v-container>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from '@/utils/axios'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const success = ref(false)
const error = ref(false)
const errorMessage = ref('')
const showResend = ref(false)
const email = ref('')
const resending = ref(false)

onMounted(async () => {
  const token = route.query.token

  if (!token) {
    error.value = true
    errorMessage.value = 'Invalid verification link. Please check your email and try again.'
    loading.value = false
    return
  }

  try {
    // Use relative URL to leverage Vite proxy in dev and work with any domain in production
    const response = await axios.get('/api/auth/verify-email', {
      params: { token }
    })

    if (response.status === 200) {
      success.value = true
    }
  } catch (err) {
    error.value = true

    if (err.response?.data?.message) {
      errorMessage.value = err.response.data.message

      // Show resend option if token expired
      if (errorMessage.value.includes('expired')) {
        showResend.value = true
      }
    } else {
      errorMessage.value = 'An error occurred while verifying your email. Please try again later.'
    }
  } finally {
    loading.value = false
  }
})

const resendVerification = async () => {
  if (!email.value) return

  resending.value = true
  try {
    await axios.post(`${API_URL}/api/auth/resend-verification`, {
      email: email.value
    })

    alert('Verification email sent! Please check your inbox.')
    showResend.value = false
  } catch (err) {
    if (err.response?.data?.message) {
      alert(err.response.data.message)
    } else {
      alert('Failed to resend verification email. Please try again.')
    }
  } finally {
    resending.value = false
  }
}

const goToLogin = () => {
  router.push('/login')
}
</script>

<style scoped>
.v-card {
  background-color: white;
}
</style>
