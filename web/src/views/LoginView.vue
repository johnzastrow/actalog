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
                type="password"
                prepend-inner-icon="mdi-lock"
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

const handleLogin = async () => {
  errors.value = {}
  loading.value = true

  const success = await authStore.login(email.value, password.value, rememberMe.value)

  if (success) {
    router.push('/dashboard')
  } else {
    errors.value.email = authStore.error || 'Login failed'
  }

  loading.value = false
}
</script>
