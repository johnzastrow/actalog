<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="4">
        <v-card elevation="4" rounded="lg">
          <v-card-title class="text-h4 text-center pa-6">
            <div class="d-flex flex-column align-center">
              <img src="/logo.svg" alt="ActaLog Logo" style="height: 80px; margin-bottom: 16px;" />
              <div class="text-primary mb-2">ActaLog</div>
              <div class="text-body-2 text-medium-emphasis">
                {{ registrationSuccess ? 'Check Your Email' : 'Create your account' }}
              </div>
            </div>
          </v-card-title>

          <v-card-text class="pa-6">
            <!-- Success Message -->
            <div v-if="registrationSuccess" class="text-center">
              <v-icon color="success" size="64" class="mb-4">mdi-email-check</v-icon>
              <h3 class="text-h5 mb-3">Registration Successful!</h3>
              <p class="text-body-1 mb-4">
                We've sent a verification email to <strong>{{ email }}</strong>
              </p>
              <p class="text-body-2 text-medium-emphasis mb-4">
                Please check your email and click the verification link to activate your account.
                The link will expire in 24 hours.
              </p>

              <!-- Email tip alert -->
              <v-alert
                type="info"
                variant="tonal"
                class="text-left mb-4"
                density="compact"
              >
                <div class="text-subtitle-2 mb-1">
                  <v-icon size="small" class="mr-1">mdi-lightbulb-outline</v-icon>
                  Can't find the email?
                </div>
                <ul class="text-body-2 mb-0 pl-4">
                  <li>Check your <strong>spam</strong> or <strong>junk</strong> folder</li>
                  <li>The email is sent from <strong>ActaLog</strong></li>
                  <li>Add our email to your contacts to ensure delivery</li>
                </ul>
              </v-alert>

              <v-btn color="primary" block @click="router.push('/login')">
                Go to Login
              </v-btn>
              <div class="text-center mt-4">
                <span class="text-body-2">Didn't receive the email? </span>
                <router-link to="/resend-verification" class="text-decoration-none">
                  Resend verification email
                </router-link>
              </div>
            </div>

            <!-- Registration Form -->
            <v-form v-else @submit.prevent="handleRegister">
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

              <v-text-field
                v-model="name"
                label="Name"
                prepend-inner-icon="mdi-account"
                required
                :error-messages="errors.name"
              />

              <v-text-field
                v-model="email"
                label="Email"
                type="email"
                prepend-inner-icon="mdi-email"
                required
                :error-messages="errors.email"
                class="mt-4"
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
                hint="Must be at least 8 characters"
                class="mt-4"
              />

              <v-text-field
                v-model="confirmPassword"
                label="Confirm Password"
                :type="showConfirmPassword ? 'text' : 'password'"
                prepend-inner-icon="mdi-lock-check"
                :append-inner-icon="showConfirmPassword ? 'mdi-eye-off' : 'mdi-eye'"
                @click:append-inner="showConfirmPassword = !showConfirmPassword"
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
                Sign Up
              </v-btn>

              <div class="text-center mt-4">
                <router-link to="/login" class="text-decoration-none">
                  Already have an account? Sign in
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

const name = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const errors = ref({})
const generalError = ref('')
const registrationSuccess = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)

const handleRegister = async () => {
  errors.value = {}
  generalError.value = ''

  // Client-side validation
  if (!name.value.trim()) {
    errors.value.name = 'Name is required'
    return
  }

  if (!email.value.trim()) {
    errors.value.email = 'Email is required'
    return
  }

  if (password.value.length < 8) {
    errors.value.password = 'Password must be at least 8 characters'
    return
  }

  if (password.value !== confirmPassword.value) {
    errors.value.confirmPassword = 'Passwords do not match'
    return
  }

  loading.value = true

  const success = await authStore.register({
    name: name.value,
    email: email.value,
    password: password.value
  })

  if (success) {
    // Show verification message instead of redirecting
    registrationSuccess.value = true
  } else {
    // Parse the error message to show it in the right place
    const errorMsg = authStore.error || 'Registration failed'

    if (errorMsg.toLowerCase().includes('password')) {
      errors.value.password = errorMsg
    } else if (errorMsg.toLowerCase().includes('email')) {
      errors.value.email = errorMsg
    } else if (errorMsg.toLowerCase().includes('name')) {
      errors.value.name = errorMsg
    } else {
      // Show as general error
      generalError.value = errorMsg
    }
  }

  loading.value = false
}
</script>
