import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import AdminUserCreateDialog from './AdminUserCreateDialog.vue'

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
// Stubs that mirror the patterns used in ProfileTab.test.js. v-dialog's content
// is rendered only when modelValue is truthy so the form mounts inline (no
// teleport / overlay that jsdom would struggle with).
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
  'v-select': {
    template:
      '<select :value="modelValue" @change="$emit(\'update:modelValue\', $event.target.value)"><option v-for="i in items" :key="i" :value="i">{{ i }}</option></select>',
    props: ['modelValue', 'label', 'items', 'required'],
    emits: ['update:modelValue'],
  },
  'v-checkbox': {
    template:
      '<input type="checkbox" :checked="modelValue" @change="$emit(\'update:modelValue\', $event.target.checked)" />',
    props: ['modelValue', 'label'],
    emits: ['update:modelValue'],
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

function createWrapper(props = { modelValue: true }) {
  return mount(AdminUserCreateDialog, {
    props,
    global: {
      stubs: VUETIFY_STUBS,
    },
  })
}

// ---------------------------------------------------------------------------
// Test data helpers
// ---------------------------------------------------------------------------
const VALID_PASSWORD = 'ValidPass123Long'

function fillValidForm(wrapper) {
  wrapper.vm.email = 'newuser@example.com'
  wrapper.vm.name = 'New User'
  wrapper.vm.role = 'athlete'
  wrapper.vm.emailVerified = true
  wrapper.vm.pw.password.value = VALID_PASSWORD
  wrapper.vm.pw.confirm.value = VALID_PASSWORD
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe('AdminUserCreateDialog', () => {
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
  it('renders the title "Create User" when open', async () => {
    wrapper = createWrapper({ modelValue: true })
    await flushPromises()

    const title = wrapper.find('.v-card-title')
    expect(title.exists()).toBe(true)
    expect(title.text()).toContain('Create User')
  })

  // -------------------------------------------------------------------------
  it('does not render content when modelValue is false', async () => {
    wrapper = createWrapper({ modelValue: false })
    await flushPromises()

    expect(wrapper.find('.v-card-title').exists()).toBe(false)
  })

  // -------------------------------------------------------------------------
  describe('submit button enable/disable', () => {
    it('disables submit until all fields are valid', async () => {
      wrapper = createWrapper({ modelValue: true })
      await flushPromises()

      // Empty form — submit should be disabled
      const submitBtn = wrapper
        .findAll('button')
        .find((b) => b.text() === 'Create')
      expect(submitBtn.exists()).toBe(true)
      expect(submitBtn.attributes('disabled')).toBeDefined()
    })

    it('enables submit when email, name, and a valid matching password are present', async () => {
      wrapper = createWrapper({ modelValue: true })
      await flushPromises()

      fillValidForm(wrapper)
      await flushPromises()

      expect(wrapper.vm.pw.isValid.value).toBe(true)

      const submitBtn = wrapper
        .findAll('button')
        .find((b) => b.text() === 'Create')
      expect(submitBtn.attributes('disabled')).toBeUndefined()
    })

    it('keeps submit disabled when passwords do not match', async () => {
      wrapper = createWrapper({ modelValue: true })
      await flushPromises()

      wrapper.vm.email = 'a@b.com'
      wrapper.vm.name = 'Someone'
      wrapper.vm.pw.password.value = VALID_PASSWORD
      wrapper.vm.pw.confirm.value = 'DifferentPass1Long'
      await flushPromises()

      const submitBtn = wrapper
        .findAll('button')
        .find((b) => b.text() === 'Create')
      expect(submitBtn.attributes('disabled')).toBeDefined()
    })

    it('keeps submit disabled when password fails complexity', async () => {
      wrapper = createWrapper({ modelValue: true })
      await flushPromises()

      wrapper.vm.email = 'a@b.com'
      wrapper.vm.name = 'Someone'
      wrapper.vm.pw.password.value = 'short'
      wrapper.vm.pw.confirm.value = 'short'
      await flushPromises()

      const submitBtn = wrapper
        .findAll('button')
        .find((b) => b.text() === 'Create')
      expect(submitBtn.attributes('disabled')).toBeDefined()
    })
  })

  // -------------------------------------------------------------------------
  describe('successful submission', () => {
    it('POSTs to /api/admin/users and emits "created" with the new user on 201', async () => {
      const created = {
        id: 42,
        email: 'newuser@example.com',
        name: 'New User',
        role: 'athlete',
        email_verified: true,
      }
      axios.post.mockResolvedValue({ status: 201, data: created })

      wrapper = createWrapper({ modelValue: true })
      await flushPromises()

      fillValidForm(wrapper)
      await flushPromises()

      await wrapper.vm.submit()
      await flushPromises()

      expect(axios.post).toHaveBeenCalledTimes(1)
      const [url, payload] = axios.post.mock.calls[0]
      expect(url).toBe('/api/admin/users')
      expect(payload).toEqual({
        email: 'newuser@example.com',
        password: VALID_PASSWORD,
        name: 'New User',
        role: 'athlete',
        email_verified: true,
      })

      // emits 'created' with response data
      const createdEvents = wrapper.emitted('created')
      expect(createdEvents).toBeTruthy()
      expect(createdEvents[0][0]).toEqual(created)

      // emits update:modelValue=false to close
      const closeEvents = wrapper.emitted('update:modelValue')
      expect(closeEvents).toBeTruthy()
      expect(closeEvents[closeEvents.length - 1][0]).toBe(false)
    })

    it('resets internal state after successful create', async () => {
      axios.post.mockResolvedValue({ status: 201, data: { id: 1 } })

      wrapper = createWrapper({ modelValue: true })
      await flushPromises()

      fillValidForm(wrapper)
      await flushPromises()

      await wrapper.vm.submit()
      await flushPromises()

      expect(wrapper.vm.email).toBe('')
      expect(wrapper.vm.name).toBe('')
      expect(wrapper.vm.role).toBe('athlete')
      expect(wrapper.vm.emailVerified).toBe(true)
      expect(wrapper.vm.pw.password.value).toBe('')
      expect(wrapper.vm.pw.confirm.value).toBe('')
    })

    it('trims whitespace from email and name before posting', async () => {
      axios.post.mockResolvedValue({ status: 201, data: { id: 1 } })

      wrapper = createWrapper({ modelValue: true })
      await flushPromises()

      wrapper.vm.email = '  spaced@example.com  '
      wrapper.vm.name = '  Spaced Name  '
      wrapper.vm.role = 'coach'
      wrapper.vm.pw.password.value = VALID_PASSWORD
      wrapper.vm.pw.confirm.value = VALID_PASSWORD
      await flushPromises()

      await wrapper.vm.submit()
      await flushPromises()

      const [, payload] = axios.post.mock.calls[0]
      expect(payload.email).toBe('spaced@example.com')
      expect(payload.name).toBe('Spaced Name')
      expect(payload.role).toBe('coach')
    })
  })

  // -------------------------------------------------------------------------
  describe('error handling', () => {
    it('shows server error message on 400 invalid_input', async () => {
      axios.post.mockRejectedValue({
        response: {
          status: 400,
          data: { error: 'invalid_input', message: 'Email is required' },
        },
      })

      wrapper = createWrapper({ modelValue: true })
      await flushPromises()

      fillValidForm(wrapper)
      await flushPromises()

      await wrapper.vm.submit()
      await flushPromises()

      expect(wrapper.vm.errorMessage).toBe('Email is required')

      // It should NOT emit 'created' on failure
      expect(wrapper.emitted('created')).toBeFalsy()
    })

    it('shows duplicate-email error on 409', async () => {
      axios.post.mockRejectedValue({
        response: {
          status: 409,
          data: {
            error: 'duplicate_email',
            message: 'A user with that email already exists',
          },
        },
      })

      wrapper = createWrapper({ modelValue: true })
      await flushPromises()

      fillValidForm(wrapper)
      await flushPromises()

      await wrapper.vm.submit()
      await flushPromises()

      expect(wrapper.vm.errorMessage).toBe(
        'A user with that email already exists',
      )
      expect(wrapper.emitted('created')).toBeFalsy()
    })

    it('shows fallback error message when server returns no message', async () => {
      axios.post.mockRejectedValue(new Error('network down'))

      wrapper = createWrapper({ modelValue: true })
      await flushPromises()

      fillValidForm(wrapper)
      await flushPromises()

      await wrapper.vm.submit()
      await flushPromises()

      expect(wrapper.vm.errorMessage).toBe(
        'Could not create user. Check server logs.',
      )
    })

    it('renders the error in a v-alert in the DOM when set', async () => {
      axios.post.mockRejectedValue({
        response: { status: 400, data: { message: 'bad email' } },
      })

      wrapper = createWrapper({ modelValue: true })
      await flushPromises()

      fillValidForm(wrapper)
      await flushPromises()

      await wrapper.vm.submit()
      await flushPromises()

      const alerts = wrapper.findAll('.v-alert')
      const errorAlerts = alerts.filter(
        (a) => a.attributes('data-type') === 'error',
      )
      expect(errorAlerts.length).toBeGreaterThan(0)
      expect(errorAlerts[0].text()).toContain('bad email')
    })

    it('clears errorMessage when dialog is reopened', async () => {
      axios.post.mockRejectedValue({
        response: { status: 400, data: { message: 'bad' } },
      })

      wrapper = createWrapper({ modelValue: true })
      await flushPromises()
      fillValidForm(wrapper)
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
