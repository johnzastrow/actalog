import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import AdminSetPasswordDialog from './AdminSetPasswordDialog.vue'

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------
vi.mock('@/utils/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
  },
}))

// offlineStorage gets imported transitively via @/utils/axios in some configs;
// stub it just in case so module wiring stays clean across runs.
vi.mock('@/utils/offlineStorage', () => ({
  addToPendingSync: vi.fn(),
  syncWithServer: vi.fn(),
  getPendingSync: vi.fn().mockResolvedValue([]),
}))

import axios from '@/utils/axios'

// ---------------------------------------------------------------------------
// Mount helpers
// ---------------------------------------------------------------------------
// Stubs that mirror the patterns used in AdminUserCreateDialog.test.js. The
// v-dialog stub renders content inline only when modelValue is truthy so the
// form mounts directly (no teleport / overlay that jsdom would struggle with).
const VUETIFY_STUBS = {
  'v-dialog': {
    template: '<div v-if="modelValue" class="v-dialog"><slot /></div>',
    props: ['modelValue', 'maxWidth', 'persistent'],
    emits: ['update:modelValue'],
  },
  'v-card': { template: '<div><slot /></div>' },
  'v-card-title': { template: '<div class="v-card-title"><slot /></div>' },
  'v-card-text': { template: '<div><slot /></div>' },
  'v-card-actions': { template: '<div><slot /></div>' },
  'v-spacer': { template: '<div />' },
  'v-form': {
    template: '<form @submit.prevent="$attrs.onSubmit?.()"><slot /></form>',
  },
  'v-text-field': {
    template:
      '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
    props: [
      'modelValue',
      'label',
      'type',
      'errorMessages',
      'hint',
      'persistentHint',
      'appendInnerIcon',
      'autofocus',
      'required',
    ],
    emits: ['update:modelValue', 'click:appendInner'],
  },
  'v-btn': {
    template:
      '<button type="button" :disabled="disabled" @click="$attrs.onClick?.()"><slot /></button>',
    props: ['loading', 'disabled', 'color', 'variant'],
  },
  'v-alert': {
    template: '<div class="v-alert" :data-type="type"><slot /></div>',
    props: ['type', 'density'],
  },
  'v-icon': { template: '<i><slot /></i>' },
}

const TARGET_USER = { id: 42, email: 'target@example.com' }

function createWrapper(props = {}) {
  return mount(AdminSetPasswordDialog, {
    props: {
      modelValue: true,
      targetUser: { ...TARGET_USER },
      ...props,
    },
    global: {
      stubs: VUETIFY_STUBS,
    },
  })
}

// ---------------------------------------------------------------------------
// Test data helpers
// ---------------------------------------------------------------------------
const VALID_PASSWORD = 'ValidPass123Long'

function fillValidPassword(wrapper) {
  wrapper.vm.pw.password.value = VALID_PASSWORD
  wrapper.vm.pw.confirm.value = VALID_PASSWORD
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe('AdminSetPasswordDialog', () => {
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
  it('renders the target email in the title (anti-mistarget)', async () => {
    wrapper = createWrapper()
    await flushPromises()

    const title = wrapper.find('.v-card-title')
    expect(title.exists()).toBe(true)
    expect(title.text()).toContain('Set password')
    expect(title.text()).toContain(TARGET_USER.email)
  })

  // -------------------------------------------------------------------------
  it('does not render content when modelValue is false', async () => {
    wrapper = createWrapper({ modelValue: false })
    await flushPromises()

    expect(wrapper.find('.v-card-title').exists()).toBe(false)
  })

  // -------------------------------------------------------------------------
  describe('submit button enable/disable', () => {
    it('disables submit until passwords are valid and matching', async () => {
      wrapper = createWrapper()
      await flushPromises()

      const submitBtn = wrapper
        .findAll('button')
        .find((b) => b.text() === 'Set Password')
      expect(submitBtn.exists()).toBe(true)
      expect(submitBtn.attributes('disabled')).toBeDefined()
    })

    it('enables submit when passwords are valid and match', async () => {
      wrapper = createWrapper()
      await flushPromises()

      fillValidPassword(wrapper)
      await flushPromises()

      expect(wrapper.vm.pw.isValid.value).toBe(true)

      const submitBtn = wrapper
        .findAll('button')
        .find((b) => b.text() === 'Set Password')
      expect(submitBtn.attributes('disabled')).toBeUndefined()
    })

    it('keeps submit disabled when passwords do not match', async () => {
      wrapper = createWrapper()
      await flushPromises()

      wrapper.vm.pw.password.value = VALID_PASSWORD
      wrapper.vm.pw.confirm.value = 'DifferentPass1Long'
      await flushPromises()

      const submitBtn = wrapper
        .findAll('button')
        .find((b) => b.text() === 'Set Password')
      expect(submitBtn.attributes('disabled')).toBeDefined()
    })

    it('keeps submit disabled when password fails complexity', async () => {
      wrapper = createWrapper()
      await flushPromises()

      wrapper.vm.pw.password.value = 'short'
      wrapper.vm.pw.confirm.value = 'short'
      await flushPromises()

      const submitBtn = wrapper
        .findAll('button')
        .find((b) => b.text() === 'Set Password')
      expect(submitBtn.attributes('disabled')).toBeDefined()
    })
  })

  // -------------------------------------------------------------------------
  describe('successful submission', () => {
    it('POSTs to /api/admin/users/{id}/password with new_password and emits "password-set" on 204', async () => {
      axios.post.mockResolvedValue({ status: 204, data: '' })

      wrapper = createWrapper()
      await flushPromises()

      fillValidPassword(wrapper)
      await flushPromises()

      await wrapper.vm.submit()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledTimes(1)
      const [url, payload] = axios.post.mock.calls[0]
      expect(url).toBe(`/api/admin/users/${TARGET_USER.id}/password`)
      expect(payload).toEqual({ new_password: VALID_PASSWORD })

      // emits 'password-set' with the target id
      const emitted = wrapper.emitted('password-set')
      expect(emitted).toBeTruthy()
      expect(emitted[0][0]).toEqual({ id: TARGET_USER.id })

      // emits update:modelValue=false to close
      const closeEvents = wrapper.emitted('update:modelValue')
      expect(closeEvents).toBeTruthy()
      expect(closeEvents[closeEvents.length - 1][0]).toBe(false)
    })

    it('resets internal password state after successful submit', async () => {
      axios.post.mockResolvedValue({ status: 204, data: '' })

      wrapper = createWrapper()
      await flushPromises()

      fillValidPassword(wrapper)
      await flushPromises()

      await wrapper.vm.submit()
      await flushPromises()

      expect(wrapper.vm.pw.password.value).toBe('')
      expect(wrapper.vm.pw.confirm.value).toBe('')
      expect(wrapper.vm.errorMessage).toBe('')
    })
  })

  // -------------------------------------------------------------------------
  describe('error handling', () => {
    it('shows complexity error from server on 400', async () => {
      axios.post.mockRejectedValue({
        response: {
          status: 400,
          data: {
            error: 'invalid_input',
            message: 'Password must include uppercase, lowercase, and a digit.',
          },
        },
      })

      wrapper = createWrapper()
      await flushPromises()

      fillValidPassword(wrapper)
      await flushPromises()

      await wrapper.vm.submit()
      await flushPromises()

      expect(wrapper.vm.errorMessage).toBe(
        'Password must include uppercase, lowercase, and a digit.',
      )

      // It should NOT emit 'password-set' on failure
      expect(wrapper.emitted('password-set')).toBeFalsy()
    })

    it('shows fallback error message when server returns no message', async () => {
      axios.post.mockRejectedValue(new Error('network down'))

      wrapper = createWrapper()
      await flushPromises()

      fillValidPassword(wrapper)
      await flushPromises()

      await wrapper.vm.submit()
      await flushPromises()

      expect(wrapper.vm.errorMessage).toBe(
        'Could not set password. Check server logs.',
      )
    })

    it('renders the error in a v-alert in the DOM when set', async () => {
      axios.post.mockRejectedValue({
        response: { status: 400, data: { message: 'too short' } },
      })

      wrapper = createWrapper()
      await flushPromises()

      fillValidPassword(wrapper)
      await flushPromises()

      await wrapper.vm.submit()
      await flushPromises()

      const alerts = wrapper.findAll('.v-alert')
      const errorAlerts = alerts.filter(
        (a) => a.attributes('data-type') === 'error',
      )
      expect(errorAlerts.length).toBeGreaterThan(0)
      expect(errorAlerts[0].text()).toContain('too short')
    })

    it('clears errorMessage when dialog is reopened', async () => {
      axios.post.mockRejectedValue({
        response: { status: 400, data: { message: 'bad' } },
      })

      wrapper = createWrapper()
      await flushPromises()
      fillValidPassword(wrapper)
      await wrapper.vm.submit()
      await flushPromises()
      expect(wrapper.vm.errorMessage).toBe('bad')

      // Simulate parent closing then reopening
      await wrapper.setProps({ modelValue: false })
      await flushPromises()
      await wrapper.setProps({ modelValue: true })
      await flushPromises()

      expect(wrapper.vm.errorMessage).toBe('')
    })
  })
})
