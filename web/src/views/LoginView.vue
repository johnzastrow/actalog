<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="4">
        <v-card elevation="4" rounded="lg">
          <v-card-title class="text-h4 text-center pa-6">
            <div class="d-flex flex-column align-center">
              <img src="/logo.svg" alt="ActaLog Logo" style="height: 80px; margin-bottom: 16px;" />
              <div class="text-primary mb-2">ActaLog</div>
              <div class="text-body-2 text-medium-emphasis">Sign in to your account</div>
            </div>
          </v-card-title>

          <v-card-text class="pa-6">
            <!-- Account Locked Alert -->
            <v-alert
              v-if="accountLocked"
              type="error"
              variant="tonal"
              class="mb-4"
              closable
              @click:close="accountLocked = false"
            >
              <div class="text-subtitle-2 mb-1">Account Locked</div>
              <div class="text-body-2">
                {{ errorMessage }}
              </div>
              <div class="text-caption mt-2">
                Please try again later or contact support if you need immediate access.
              </div>
            </v-alert>

            <!-- Rate Limited Alert -->
            <v-alert
              v-else-if="rateLimited"
              type="warning"
              variant="tonal"
              class="mb-4"
              closable
              @click:close="rateLimited = false"
            >
              <div class="text-subtitle-2 mb-1">Too Many Attempts</div>
              <div class="text-body-2">
                {{ errorMessage }}
              </div>
            </v-alert>

            <!-- General Error Alert -->
            <v-alert
              v-if="generalError"
              type="error"
              variant="tonal"
              closable
              class="mb-4"
              @click:close="generalError = ''"
            >
              {{ generalError }}
            </v-alert>

            <v-form @submit.prevent="handleLogin">
              <v-text-field
                v-model="email"
                label="Email"
                type="email"
                prepend-inner-icon="mdi-email"
                required
                :error-messages="errors.email"
              />

              <v-text-field
                v-model="password"
                label="Password"
                :type="showPassword ? 'text' : 'password'"
                prepend-inner-icon="mdi-lock"
                :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
                @click:append-inner="showPassword = !showPassword"
                required
                :error-messages="errors.password"
                class="mt-4"
              />

              <v-checkbox
                v-model="rememberMe"
                label="Remember me for 30 days"
                color="primary"
                hide-details
                class="mt-2"
              />

              <v-btn
                type="submit"
                color="primary"
                block
                size="large"
                class="mt-6"
                :loading="loading"
              >
                Sign In
              </v-btn>

              <div class="text-center mt-3">
                <router-link to="/forgot-password" class="text-decoration-none text-body-2">
                  Forgot password?
                </router-link>
              </div>

              <div class="text-center mt-4">
                <router-link to="/register" class="text-decoration-none">
                  Don't have an account? Sign up
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
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const rememberMe = ref(false)
const loading = ref(false)
const errors = ref({})
const generalError = ref('')
const accountLocked = ref(false)
const rateLimited = ref(false)
const errorMessage = ref('')
const showPassword = ref(false)

const handleLogin = async () => {
  errors.value = {}
  generalError.value = ''
  accountLocked.value = false
  rateLimited.value = false
  errorMessage.value = ''
  loading.value = true

  // Client-side validation
  if (!email.value.trim()) {
    errors.value.email = 'Email is required'
    loading.value = false
    return
  }

  if (!password.value) {
    errors.value.password = 'Password is required'
    loading.value = false
    return
  }

  const success = await authStore.login(email.value, password.value, rememberMe.value)

  if (success) {
    router.push('/dashboard')
  } else {
    const error = authStore.error || 'Login failed'
    errorMessage.value = error

    // Check for specific error types
    if (error.toLowerCase().includes('account locked') || error.toLowerCase().includes('too many failed')) {
      accountLocked.value = true
    } else if (error.toLowerCase().includes('too many requests') || error.toLowerCase().includes('rate limit')) {
      rateLimited.value = true
    } else if (error.toLowerCase().includes('invalid email or password') || error.toLowerCase().includes('invalid credentials')) {
      // Show on both fields for security (don't reveal which is wrong)
      generalError.value = 'Invalid email or password. Please check your credentials and try again.'
    } else {
      // Show other errors as general error
      generalError.value = error
    }
  }

  loading.value = false
}
</script>
