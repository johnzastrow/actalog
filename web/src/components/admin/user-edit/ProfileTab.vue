<script setup>
import { ref, computed } from 'vue'
import axios from '@/utils/axios'
import { useUserDraft } from './composables/useUserDraft'
import { useProtectedUserStatus } from './composables/useProtectedUserStatus'
import ProtectedUserBanner from './ProtectedUserBanner.vue'
import TabFooterActions from './TabFooterActions.vue'

const props = defineProps({
  userId: { type: Number, required: true },
})

const draft = useUserDraft({
  fetchOriginal: () => axios.get(`/api/admin/users/${props.userId}`),
  saveFields: (fields) => axios.patch(`/api/admin/users/${props.userId}`, fields),
  fields: ['name', 'email', 'birthday', 'email_verified'],
})

const isProtected = useProtectedUserStatus(draft.original)

// Inline success/error banners — the project pattern for admin components.
const successMessage = ref(null)

draft.load()

// ---------------------------------------------------------------------------
// Birthday display ↔ wire format converter
//   display: 'YYYY-MM-DD' (HTML type=date input)
//   wire:    RFC 3339 string (Go *time.Time JSON contract)
// ---------------------------------------------------------------------------
const birthdayDisplay = computed({
  get() {
    const v = draft.working.birthday
    if (!v) return ''
    // Already display format (YYYY-MM-DD)?
    if (typeof v === 'string' && v.length === 10 && v.includes('-')) return v
    // RFC 3339 → take first 10 chars
    return String(v).substring(0, 10)
  },
  set(newVal) {
    if (!newVal) {
      draft.working.birthday = null
      return
    }
    // YYYY-MM-DD → YYYY-MM-DDT00:00:00Z (RFC 3339 with UTC midnight)
    draft.working.birthday = `${newVal}T00:00:00Z`
  },
})

// ---------------------------------------------------------------------------
// Force password reset — separate admin action, outside the draft+save flow.
// ---------------------------------------------------------------------------
const sendingReset = ref(false)
const resetError = ref(null)
const showResetConfirm = ref(false)

function forcePasswordReset() {
  resetError.value = null
  showResetConfirm.value = true
}

async function confirmForcePasswordReset() {
  showResetConfirm.value = false
  sendingReset.value = true
  successMessage.value = null
  resetError.value = null
  try {
    await axios.post(`/api/admin/users/${props.userId}/force-password-reset`)
    successMessage.value = 'Password-reset email sent.'
  } catch (e) {
    // 403 protected_user / 503 degraded are handled by axios interceptor
    // For other errors (500, network, etc.), surface an inline error.
    const status = e.response?.status
    if (status !== 403 && status !== 503) {
      resetError.value = e.response?.data?.message || 'Failed to send password-reset email.'
    }
  } finally {
    sendingReset.value = false
  }
}

// ---------------------------------------------------------------------------
// Email-change confirmation — show dialog before saving if email was changed.
// ---------------------------------------------------------------------------
const showEmailChangeConfirm = ref(false)

async function attemptSave() {
  if (draft.modified.value.has('email')) {
    showEmailChangeConfirm.value = true
    return
  }
  await executeSave()
}

async function confirmEmailChange() {
  showEmailChangeConfirm.value = false
  await executeSave()
}

async function executeSave() {
  try {
    await draft.save()
    successMessage.value = 'Profile saved.'
  } catch {
    // draft.error.value is set by the composable;
    // 403/503 are also handled globally by the axios interceptor.
  }
}
</script>

<template>
  <ProtectedUserBanner v-if="isProtected" />

  <template v-else>
    <!-- Success banner -->
    <v-alert
      v-if="successMessage"
      type="success"
      variant="tonal"
      closable
      class="mb-4"
      @click:close="successMessage = null"
    >
      {{ successMessage }}
    </v-alert>

    <!-- Conflict banner -->
    <v-alert
      v-if="draft.conflict.value"
      type="warning"
      variant="tonal"
      closable
      class="mb-4"
      @click:close="draft.conflict.value = false"
    >
      This record was updated by someone else. Discard your changes and reload to see the latest
      version.
    </v-alert>

    <!-- Load error banner -->
    <v-alert
      v-if="draft.error.value && !draft.conflict.value"
      type="error"
      variant="tonal"
      closable
      class="mb-4"
      @click:close="draft.error.value = null"
    >
      {{ draft.error.value?.message || 'An error occurred.' }}
    </v-alert>

    <!-- Reset error banner -->
    <v-alert
      v-if="resetError"
      type="error"
      variant="tonal"
      closable
      class="mb-4"
      @click:close="resetError = null"
    >
      {{ resetError }}
    </v-alert>

    <!-- Loading indicator — shown while initial fetch is in flight -->
    <v-progress-linear v-if="!draft.original.value?.id" indeterminate color="primary" />

    <v-form v-else @submit.prevent="attemptSave">
      <v-text-field
        v-model="draft.working.name"
        label="Name"
        :error-messages="draft.fieldErrors.name"
      />
      <v-text-field
        v-model="draft.working.email"
        label="Email"
        type="email"
        :error-messages="draft.fieldErrors.email"
      />
      <v-text-field
        v-model="birthdayDisplay"
        label="Birthday"
        type="date"
        :error-messages="draft.fieldErrors.birthday"
      />
      <v-switch
        v-model="draft.working.email_verified"
        label="Email verified"
        color="primary"
      />

      <v-divider class="my-4" />
      <h3 class="text-subtitle-1 mb-2">Account actions</h3>
      <v-btn
        variant="tonal"
        color="warning"
        prepend-icon="mdi-key-alert"
        :loading="sendingReset"
        @click="forcePasswordReset"
      >
        Force password reset
      </v-btn>

      <TabFooterActions :draft="draft" @click:save="attemptSave" />
    </v-form>
  </template>

  <!-- Email change confirmation dialog -->
  <v-dialog v-model="showEmailChangeConfirm" max-width="500">
    <v-card>
      <v-card-title>Confirm email change</v-card-title>
      <v-card-text>
        Changing email will reset email-verified status to false and send a new verification
        message to <strong>{{ draft.working.email }}</strong>. Continue?
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="showEmailChangeConfirm = false">Cancel</v-btn>
        <v-btn color="primary" @click="confirmEmailChange">Continue</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- Force password reset confirmation dialog -->
  <v-dialog v-model="showResetConfirm" max-width="500">
    <v-card>
      <v-card-title>Force password reset?</v-card-title>
      <v-card-text>
        This will send a password-reset email to
        <strong>{{ draft.original.value?.email }}</strong>
        and log them out of all active sessions. The user will need to use the reset link to set a
        new password.
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="showResetConfirm = false">Cancel</v-btn>
        <v-btn color="warning" @click="confirmForcePasswordReset">Continue</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
