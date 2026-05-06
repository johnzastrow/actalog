import { ref, computed } from 'vue'

/**
 * usePasswordInputs — shared password+confirm state for admin dialogs.
 *
 * Returns reactive refs and computeds that wire directly into Vuetify
 * v-text-field bindings. Validation policy mirrors the backend (pkg/auth +
 * internal/service/user_service.go validatePassword):
 *   - >=12 characters
 *   - at least one uppercase
 *   - at least one lowercase
 *   - at least one digit
 *
 * The backend remains the source of truth — this is a UX hint that lets the
 * dialog disable the submit button before sending a doomed request.
 */
export function usePasswordInputs() {
  const password = ref('')
  const confirm = ref('')
  const visible = ref(false)

  const errorMessage = computed(() => {
    if (!password.value && !confirm.value) return ''
    if (password.value.length < 12) {
      return 'Password must be at least 12 characters.'
    }
    const hasUpper = /[A-Z]/.test(password.value)
    const hasLower = /[a-z]/.test(password.value)
    const hasDigit = /[0-9]/.test(password.value)
    if (!hasUpper || !hasLower || !hasDigit) {
      return 'Password must include uppercase, lowercase, and a digit.'
    }
    if (password.value !== confirm.value) {
      return 'Passwords do not match.'
    }
    return ''
  })

  const isValid = computed(() => {
    return password.value.length > 0 && errorMessage.value === ''
  })

  function toggleVisible() {
    visible.value = !visible.value
  }

  function reset() {
    password.value = ''
    confirm.value = ''
    visible.value = false
  }

  return { password, confirm, visible, errorMessage, isValid, toggleVisible, reset }
}
