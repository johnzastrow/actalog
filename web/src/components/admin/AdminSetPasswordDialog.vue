<template>
  <v-dialog
    :model-value="modelValue"
    max-width="500"
    persistent
    @update:model-value="(v) => $emit('update:modelValue', v)"
  >
    <v-card>
      <v-card-title class="text-h6">
        Set password for
        <span class="font-weight-regular">{{ targetUser?.email }}</span>
      </v-card-title>
      <v-card-text>
        <v-alert type="info" density="compact" class="mb-3">
          Sets a new password directly. The user will be signed out of all
          current sessions and any account lockout will be cleared.
        </v-alert>
        <v-form @submit.prevent="submit">
          <v-text-field
            v-model="pw.password.value"
            label="New password"
            :type="pw.visible.value ? 'text' : 'password'"
            :append-inner-icon="pw.visible.value ? 'mdi-eye-off' : 'mdi-eye'"
            hint="Min 12 chars; uppercase, lowercase, digit."
            persistent-hint
            autofocus
            required
            @click:append-inner="pw.toggleVisible"
          />
          <v-text-field
            v-model="pw.confirm.value"
            label="Confirm new password"
            :type="pw.visible.value ? 'text' : 'password'"
            :error-messages="pw.errorMessage.value"
            required
          />
          <v-alert
            v-if="errorMessage"
            type="error"
            density="compact"
            class="mt-2"
          >
            {{ errorMessage }}
          </v-alert>
        </v-form>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="cancel">Cancel</v-btn>
        <v-btn
          color="primary"
          :disabled="!pw.isValid.value || submitting"
          :loading="submitting"
          @click="submit"
        >
          Set Password
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import axios from '@/utils/axios'
import { usePasswordInputs } from '@/components/admin/composables/usePasswordInputs.js'

const props = defineProps({
  modelValue: { type: Boolean, required: true },
  targetUser: { type: Object, required: true },
})
const emit = defineEmits(['update:modelValue', 'password-set'])

const pw = usePasswordInputs()
const submitting = ref(false)
const errorMessage = ref('')

async function submit() {
  errorMessage.value = ''
  submitting.value = true
  try {
    await axios.post(`/api/admin/users/${props.targetUser.id}/password`, {
      new_password: pw.password.value,
    })
    emit('password-set', { id: props.targetUser.id })
    closeAndReset()
  } catch (e) {
    if (e?.response?.data?.message) {
      errorMessage.value = e.response.data.message
    } else if (e?.response?.data?.error) {
      errorMessage.value = e.response.data.error
    } else {
      errorMessage.value = 'Could not set password. Check server logs.'
    }
  } finally {
    submitting.value = false
  }
}

function cancel() {
  closeAndReset()
}

function closeAndReset() {
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

defineExpose({ pw, submit, errorMessage })
</script>
