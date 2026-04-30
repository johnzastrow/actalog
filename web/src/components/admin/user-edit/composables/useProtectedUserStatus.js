import { computed } from 'vue'
import { isProtectedEmail } from '@/utils/protectedUsers'

/**
 * Reactive helper: returns a computed boolean given a ref/computed of a User-like object.
 *
 * @param {import('vue').Ref<{email: string} | null | undefined>} userRef — reactive reference to a User
 * @returns {import('vue').ComputedRef<boolean>}
 */
export function useProtectedUserStatus(userRef) {
  return computed(() => {
    const u = userRef.value
    return u ? isProtectedEmail(u.email) : false
  })
}
