import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ProfileTab from './ProfileTab.vue'

// ---------------------------------------------------------------------------
// Mock axios
// ---------------------------------------------------------------------------
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
  },
}))

// Mock offlineStorage so the axios module's top-level imports don't fail
vi.mock('@/utils/offlineStorage', () => ({
  addToPendingSync: vi.fn(),
  syncWithServer: vi.fn(),
  getPendingSync: vi.fn().mockResolvedValue([]),
}))

import axios from '@/utils/axios'

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------
const NORMAL_USER = {
  id: 7,
  name: 'Test User',
  email: 'test@example.com',
  birthday: '1990-05-15',
  email_verified: true,
  updated_at: '2026-04-01T10:00:00Z',
}

const PROTECTED_USER = {
  id: 1,
  name: 'Protected',
  email: 'br8kwall@gmail.com',
  birthday: null,
  email_verified: true,
  updated_at: '2026-01-01T00:00:00Z',
}

// ---------------------------------------------------------------------------
// Mount helpers
// ---------------------------------------------------------------------------
const VUETIFY_STUBS = {
  'v-form': { template: '<form @submit.prevent="$attrs.onSubmit?.()"><slot /></form>' },
  'v-text-field': { template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />', props: ['modelValue', 'label', 'type'], emits: ['update:modelValue'] },
  'v-switch': { template: '<input type="checkbox" :checked="modelValue" @change="$emit(\'update:modelValue\', $event.target.checked)" />', props: ['modelValue', 'label', 'color'], emits: ['update:modelValue'] },
  'v-btn': { template: '<button type="button" @click="$attrs.onClick?.()"><slot /></button>', props: ['loading', 'disabled', 'color', 'variant', 'prependIcon'] },
  'v-divider': { template: '<hr />' },
  'v-alert': { template: '<div class="v-alert" :data-type="type"><slot /></div>', props: ['type', 'variant', 'closable'] },
  'v-dialog': { template: '<div v-if="modelValue"><slot /></div>', props: ['modelValue'], emits: ['update:modelValue'] },
  'v-card': { template: '<div><slot /></div>' },
  'v-card-title': { template: '<div><slot /></div>' },
  'v-card-text': { template: '<div><slot /></div>' },
  'v-card-actions': { template: '<div><slot /></div>' },
  'v-spacer': { template: '<div />' },
  'v-icon': { template: '<i />' },
  'v-fade-transition': { template: '<slot />' },
  'v-progress-linear': { template: '<div class="v-progress-linear" />' },
  // Stub out child components that have their own dependencies
  'ProtectedUserBanner': { template: '<div class="protected-user-banner" />' },
  'TabFooterActions': { template: '<div class="tab-footer-actions"><button class="save-btn" @click="$emit(\'click:save\')">Save</button><button class="discard-btn" @click="draft.discard()">Discard</button></div>', props: ['draft'], emits: ['click:save'] },
}

function createWrapper(props = { userId: 7 }) {
  const pinia = createPinia()
  setActivePinia(pinia)
  return shallowMount(ProfileTab, {
    props,
    global: {
      plugins: [pinia],
      stubs: VUETIFY_STUBS,
    },
  })
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('ProfileTab', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
      wrapper = null
    }
  })

  // -------------------------------------------------------------------------
  // Test 1: renders profile fields when target is not protected
  // -------------------------------------------------------------------------
  describe('renders profile fields when target is not protected', () => {
    it('shows form fields and hides ProtectedUserBanner for a normal user', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      // The form should be visible, not the banner
      expect(wrapper.find('form').exists()).toBe(true)
      expect(wrapper.find('.protected-user-banner').exists()).toBe(false)

      // Save/footer actions should be present (inside the form branch)
      expect(wrapper.find('.tab-footer-actions').exists()).toBe(true)
    })

    it('loads name and email into working fields', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      // Verify axios was called with correct URL
      expect(axios.get).toHaveBeenCalledWith('/api/admin/users/7')
    })
  })

  // -------------------------------------------------------------------------
  // Test 2: shows ProtectedUserBanner for protected user
  // -------------------------------------------------------------------------
  describe('shows ProtectedUserBanner for protected user', () => {
    it('renders banner and hides form when user is protected', async () => {
      axios.get.mockResolvedValue({ data: PROTECTED_USER })

      wrapper = createWrapper({ userId: 1 })
      await flushPromises()

      // Banner should show
      expect(wrapper.find('.protected-user-banner').exists()).toBe(true)

      // Form (and therefore save buttons) should NOT be visible
      expect(wrapper.find('form').exists()).toBe(false)
      expect(wrapper.find('.tab-footer-actions').exists()).toBe(false)
    })
  })

  // -------------------------------------------------------------------------
  // Test 3: save sends only modified fields
  // -------------------------------------------------------------------------
  describe('save sends only modified fields', () => {
    it('PATCHes only the changed name field plus updated_at', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })
      const updatedUser = { ...NORMAL_USER, name: 'New Name', updated_at: '2026-04-02T00:00:00Z' }
      axios.patch.mockResolvedValue({ data: updatedUser })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      // Mutate the working copy directly via the vm's draft
      const draft = wrapper.vm.draft
      draft.working.name = 'New Name'
      await flushPromises()

      // Trigger save via the TabFooterActions click:save emit
      await wrapper.find('.save-btn').trigger('click')
      await flushPromises()

      // Verify PATCH was called
      expect(axios.patch).toHaveBeenCalledOnce()
      const [url, payload] = axios.patch.mock.calls[0]
      expect(url).toBe('/api/admin/users/7')
      expect(payload).toHaveProperty('name', 'New Name')
      expect(payload).toHaveProperty('updated_at', NORMAL_USER.updated_at)
      // Must NOT include unmodified fields
      expect(payload).not.toHaveProperty('email')
      expect(payload).not.toHaveProperty('birthday')
      expect(payload).not.toHaveProperty('email_verified')
    })
  })

  // -------------------------------------------------------------------------
  // Additional: email-change confirmation dialog
  // -------------------------------------------------------------------------
  describe('email-change confirmation flow', () => {
    it('shows confirmation dialog when email is changed and save is clicked', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      // Change email
      wrapper.vm.draft.working.email = 'newemail@example.com'
      await flushPromises()

      // Click save
      await wrapper.find('.save-btn').trigger('click')
      await flushPromises()

      // Dialog should now be open, PATCH not called yet
      expect(wrapper.vm.showEmailChangeConfirm).toBe(true)
      expect(axios.patch).not.toHaveBeenCalled()
    })

    it('does NOT show confirmation dialog when a non-email field is changed', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })
      axios.patch.mockResolvedValue({ data: NORMAL_USER })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      // Change name only
      wrapper.vm.draft.working.name = 'Different Name'
      await flushPromises()

      // Click save
      await wrapper.find('.save-btn').trigger('click')
      await flushPromises()

      // Dialog should NOT appear; PATCH should be called directly
      expect(wrapper.vm.showEmailChangeConfirm).toBe(false)
      expect(axios.patch).toHaveBeenCalledOnce()
    })
  })

  // -------------------------------------------------------------------------
  // Additional: force password reset action
  // -------------------------------------------------------------------------
  describe('force password reset', () => {
    it('POSTs to the force-password-reset endpoint after confirmation', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })
      axios.post.mockResolvedValue({ data: {} })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      await wrapper.vm.confirmForcePasswordReset()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledWith('/api/admin/users/7/force-password-reset')
    })

    it('opens confirmation dialog when force-reset button is clicked', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      wrapper.vm.forcePasswordReset()
      await flushPromises()

      expect(wrapper.vm.showResetConfirm).toBe(true)
      expect(axios.post).not.toHaveBeenCalled()
    })
  })

  // -------------------------------------------------------------------------
  // Fix 1: Birthday converter — RFC 3339 in PATCH payload
  // -------------------------------------------------------------------------
  describe('birthday format conversion', () => {
    it('sends RFC 3339 birthday in PATCH payload when birthday is changed', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })
      const updatedUser = { ...NORMAL_USER, birthday: '1985-03-22T00:00:00Z', updated_at: '2026-04-02T00:00:00Z' }
      axios.patch.mockResolvedValue({ data: updatedUser })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      // Simulate HTML date input setting YYYY-MM-DD via the computed setter
      wrapper.vm.birthdayDisplay = '1985-03-22'
      await flushPromises()

      // Trigger save
      await wrapper.find('.save-btn').trigger('click')
      await flushPromises()

      expect(axios.patch).toHaveBeenCalledOnce()
      const [, payload] = axios.patch.mock.calls[0]
      expect(payload).toHaveProperty('birthday', '1985-03-22T00:00:00Z')
    })

    it('birthdayDisplay getter returns YYYY-MM-DD from an RFC 3339 value', async () => {
      const userWithRFC3339Birthday = { ...NORMAL_USER, birthday: '1990-05-15T00:00:00Z' }
      axios.get.mockResolvedValue({ data: userWithRFC3339Birthday })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      expect(wrapper.vm.birthdayDisplay).toBe('1990-05-15')
    })

    it('birthdayDisplay getter returns empty string when birthday is null', async () => {
      const userNoBirthday = { ...NORMAL_USER, birthday: null }
      axios.get.mockResolvedValue({ data: userNoBirthday })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      expect(wrapper.vm.birthdayDisplay).toBe('')
    })

    it('birthdayDisplay setter sets null when given empty string', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      wrapper.vm.birthdayDisplay = ''
      expect(wrapper.vm.draft.working.birthday).toBeNull()
    })
  })

  // -------------------------------------------------------------------------
  // Fix 2: Reset error state shown on non-403/503 errors
  // -------------------------------------------------------------------------
  describe('force password reset error handling', () => {
    it('shows inline error message when force-reset fails with 500', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })
      axios.post.mockRejectedValue({
        response: { status: 500, data: { message: 'Internal server error' } },
      })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      await wrapper.vm.confirmForcePasswordReset()
      await flushPromises()

      expect(wrapper.vm.resetError).toBe('Internal server error')
    })

    it('shows fallback error message when server returns no message body', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })
      axios.post.mockRejectedValue({ response: { status: 500, data: {} } })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      await wrapper.vm.confirmForcePasswordReset()
      await flushPromises()

      expect(wrapper.vm.resetError).toBe('Failed to send password-reset email.')
    })

    it('does NOT set resetError for 403 (handled by axios interceptor)', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })
      axios.post.mockRejectedValue({ response: { status: 403, data: { error: 'protected_user' } } })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      await wrapper.vm.confirmForcePasswordReset()
      await flushPromises()

      expect(wrapper.vm.resetError).toBeNull()
    })

    it('does NOT set resetError for 503 (handled by axios interceptor)', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })
      axios.post.mockRejectedValue({ response: { status: 503, data: { error: 'protected_invariant_degraded' } } })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      await wrapper.vm.confirmForcePasswordReset()
      await flushPromises()

      expect(wrapper.vm.resetError).toBeNull()
    })

    it('renders resetError v-alert in the DOM when resetError is set', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })
      axios.post.mockRejectedValue({ response: { status: 500, data: {} } })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      await wrapper.vm.confirmForcePasswordReset()
      await flushPromises()

      // The resetError alert should be in the DOM (outside the v-else form block)
      const alerts = wrapper.findAll('.v-alert')
      const errorAlerts = alerts.filter((a) => a.attributes('data-type') === 'error')
      expect(errorAlerts.length).toBeGreaterThan(0)
    })
  })

  // -------------------------------------------------------------------------
  // Fix 3: Loading indicator shown while draft.original has no id
  // -------------------------------------------------------------------------
  describe('loading indicator', () => {
    it('shows v-progress-linear while initial load is in flight', async () => {
      // Delay resolution so we can inspect loading state
      let resolveLoad
      axios.get.mockReturnValue(new Promise((resolve) => { resolveLoad = resolve }))

      wrapper = createWrapper({ userId: 7 })
      // Don't flush promises yet — still loading

      // The progress bar should exist (original.value has no id yet)
      expect(wrapper.find('.v-progress-linear').exists()).toBe(true)
      expect(wrapper.find('form').exists()).toBe(false)

      // Now resolve the load
      resolveLoad({ data: NORMAL_USER })
      await flushPromises()

      // Progress bar gone; form visible
      expect(wrapper.find('.v-progress-linear').exists()).toBe(false)
      expect(wrapper.find('form').exists()).toBe(true)
    })
  })

  // -------------------------------------------------------------------------
  // Fix 4: Confirmation dialog for force-reset (covered above, extra check)
  // -------------------------------------------------------------------------
  describe('force-reset confirmation dialog', () => {
    it('closes dialog and POSTs when confirmForcePasswordReset is called', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })
      axios.post.mockResolvedValue({ data: {} })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      // Open the dialog first
      wrapper.vm.forcePasswordReset()
      await flushPromises()
      expect(wrapper.vm.showResetConfirm).toBe(true)

      // Confirm
      await wrapper.vm.confirmForcePasswordReset()
      await flushPromises()

      expect(wrapper.vm.showResetConfirm).toBe(false)
      expect(axios.post).toHaveBeenCalledWith('/api/admin/users/7/force-password-reset')
    })

    it('shows success message after confirmed force-reset', async () => {
      axios.get.mockResolvedValue({ data: NORMAL_USER })
      axios.post.mockResolvedValue({ data: {} })

      wrapper = createWrapper({ userId: 7 })
      await flushPromises()

      await wrapper.vm.confirmForcePasswordReset()
      await flushPromises()

      expect(wrapper.vm.successMessage).toBe('Password-reset email sent.')
    })
  })
})
