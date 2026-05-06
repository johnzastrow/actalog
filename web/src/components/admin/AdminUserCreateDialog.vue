<template>
  <v-dialog
    :model-value="modelValue"
    max-width="560"
    persistent
    @update:model-value="(v) => $emit('update:modelValue', v)"
  >
    <v-card>
      <v-card-title class="text-h6">Create User</v-card-title>
      <v-card-text>
        <v-form @submit.prevent="submit">
          <v-text-field
            v-model="email"
            label="Email"
            type="email"
            autofocus
            required
          />
          <v-text-field
            v-model="pw.password.value"
            label="Password"
            :type="pw.visible.value ? 'text' : 'password'"
            :append-inner-icon="pw.visible.value ? 'mdi-eye-off' : 'mdi-eye'"
            hint="Min 12 chars; uppercase, lowercase, digit."
            persistent-hint
            required
            @click:append-inner="pw.toggleVisible"
          />
          <v-text-field
            v-model="pw.confirm.value"
            label="Confirm password"
            :type="pw.visible.value ? 'text' : 'password'"
            :error-messages="pw.errorMessage.value"
            required
          />
          <v-text-field
            v-model="name"
            label="Name"
            required
          />
          <v-select
            v-model="role"
            label="Role"
            :items="['athlete', 'coach', 'admin']"
            required
          />
          <v-checkbox
            v-model="emailVerified"
            label="Mark email as verified (admin-vouched)"
          />

          <v-alert v-if="errorMessage" type="error" density="compact" class="mb-2">
            {{ errorMessage }}
          </v-alert>
        </v-form>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="cancel">Cancel</v-btn>
        <v-btn
          color="primary"
          :disabled="!canSubmit || submitting"
          :loading="submitting"
          @click="submit"
        >
          Create
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import axios from '@/utils/axios'
import { usePasswordInputs } from '@/components/admin/composables/usePasswordInputs.js'

const props = defineProps({
  modelValue: { type: Boolean, required: true },
})
const emit = defineEmits(['update:modelValue', 'created'])

const email = ref('')
const name = ref('')
const role = ref('athlete')
const emailVerified = ref(true)
const pw = usePasswordInputs()
const submitting = ref(false)
const errorMessage = ref('')

const canSubmit = computed(() => {
  return (
    email.value.trim().length > 0 &&
    name.value.trim().length > 0 &&
    pw.isValid.value
  )
})

async function submit() {
  errorMessage.value = ''
  submitting.value = true
  try {
    const res = await axios.post('/api/admin/users', {
      email: email.value.trim(),
      password: pw.password.value,
      name: name.value.trim(),
      role: role.value,
      email_verified: emailVerified.value,
    })
    emit('created', res.data)
    closeAndReset()
  } catch (e) {
    if (e?.response?.data?.message) {
      errorMessage.value = e.response.data.message
    } else if (e?.response?.data?.error) {
      errorMessage.value = e.response.data.error
    } else {
      errorMessage.value = 'Could not create user. Check server logs.'
    }
  } finally {
    submitting.value = false
  }
}

function cancel() {
  closeAndReset()
}

function closeAndReset() {
  email.value = ''
  name.value = ''
  role.value = 'athlete'
  emailVerified.value = true
  pw.reset()
  errorMessage.value = ''
  emit('update:modelValue', false)
}

// Clear stale error each time the dialog reopens.
watch(
  () => props.modelValue,
  (v) => {
    if (v) {
      errorMessage.value = ''
    }
  },
)

defineExpose({ email, name, role, emailVerified, pw, submit, errorMessage })
</script>
