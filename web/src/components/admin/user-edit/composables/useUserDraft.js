import { ref, reactive, computed } from 'vue'

/**
 * Encapsulates the draft+save state machine for a single tab on the
 * Admin User Edit screen.
 *
 * @param {Object} opts
 * @param {() => Promise<{ data: Object }>} opts.fetchOriginal
 *   Fetches the server-side row; must resolve to `{ data: <userObject> }`.
 * @param {(payload: Object) => Promise<{ data: Object }>} opts.saveFields
 *   Sends only modified fields + updated_at; must resolve to `{ data: <userObject> }`.
 * @param {string[]} opts.fields
 *   Field names to track for dirty detection (e.g. ['name', 'email']).
 *
 * @returns {{
 *   original: import('vue').Ref<Object>,
 *   working: Object,
 *   modified: import('vue').ComputedRef<Set<string>>,
 *   isDirty: import('vue').ComputedRef<boolean>,
 *   saving: import('vue').Ref<boolean>,
 *   error: import('vue').Ref<Error|null>,
 *   conflict: import('vue').Ref<boolean>,
 *   fieldErrors: Object  // consumer-managed per-field errors; never written by this composable
 *   load: () => Promise<void>,
 *   save: () => Promise<void>,
 *   discard: () => void,
 * }}
 */
export function useUserDraft({ fetchOriginal, saveFields, fields }) {
  // The last known server state.
  const original = ref({})

  // The in-progress edit state. Components bind directly to these properties.
  const working = reactive({})

  // Async / error state
  const saving = ref(false)
  const error = ref(null)
  const conflict = ref(false)

  // Consumer-managed per-field validation errors (not touched by this composable).
  const fieldErrors = reactive({})

  // Computed set of field names where working differs from original.
  const modified = computed(() => {
    const set = new Set()
    for (const f of fields) {
      if (original.value[f] !== working[f]) set.add(f)
    }
    return set
  })

  // True when at least one tracked field has been changed.
  const isDirty = computed(() => modified.value.size > 0)

  /**
   * Fetch the server row and initialise both original and working copies.
   * Clears error and conflict state.
   */
  async function load() {
    try {
      const res = await fetchOriginal()
      original.value = res.data
      for (const f of fields) {
        working[f] = res.data[f]
      }
      error.value = null
      conflict.value = false
    } catch (e) {
      error.value = e
      throw e
    }
  }

  /**
   * PATCH only the modified fields plus `updated_at` (optimistic-concurrency token).
   * On success, refreshes original from the server response and clears dirty state.
   * On 409, sets `conflict` so the UI can prompt the user to reload.
   * Always re-throws so callers can handle errors.
   */
  async function save() {
    if (!isDirty.value) return
    if (saving.value) return  // concurrency guard — caller must wait for in-flight save
    saving.value = true
    error.value = null
    try {
      const payload = { updated_at: original.value.updated_at }
      for (const f of modified.value) {
        payload[f] = working[f]
      }
      const res = await saveFields(payload)
      // Refresh original and working from server response so they stay in sync.
      original.value = res.data
      for (const f of fields) {
        working[f] = res.data[f]
      }
      conflict.value = false
    } catch (e) {
      if (e?.response?.status === 409) {
        conflict.value = true
      }
      error.value = e
      throw e
    } finally {
      saving.value = false
    }
  }

  /**
   * Reset working copy back to the last known original.
   * Clears conflict and error state.
   */
  function discard() {
    for (const f of fields) {
      working[f] = original.value[f]
    }
    conflict.value = false
    error.value = null
  }

  return {
    original,
    working,
    modified,
    isDirty,
    saving,
    error,
    conflict,
    fieldErrors,
    load,
    save,
    discard,
  }
}
