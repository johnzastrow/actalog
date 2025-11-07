<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="4">
        <v-card elevation="4" rounded="lg">
          <v-card-title class="text-h4 text-center pa-6">
            <div class="d-flex flex-column align-center">
              <div class="text-primary mb-2">ActaLog</div>
              <div class="text-body-2 text-medium-emphasis">Create your account</div>
            </div>
          </v-card-title>

          <v-card-text class="pa-6">
            <v-form @submit.prevent="handleRegister">
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
                type="password"
                prepend-inner-icon="mdi-lock"
                required
                :error-messages="errors.password"
                class="mt-4"
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

const handleRegister = async () => {
  errors.value = {}

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
    router.push('/dashboard')
  } else {
    errors.value.email = authStore.error || 'Registration failed'
  }

  loading.value = false
}
</script>
