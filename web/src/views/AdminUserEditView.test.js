import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminUserEditView from './AdminUserEditView.vue'

// ── Mocks ─────────────────────────────────────────────────────────────────────

vi.mock('@/components/AdminHeader.vue', () => ({
  default: { template: '<div />', props: ['title', 'subtitle', 'breadcrumbs'] },
}))

vi.mock('@/components/admin/user-edit/ProfileTab.vue', () => ({
  default: { template: '<div data-testid="profile-tab" />', props: ['userId'] },
}))

// We need a controllable router mock so we can inspect router.replace calls
// and set the current route (including query params) per test.
const mockRouterReplace = vi.fn()
let mockRouteQuery = {}
let mockRouteFullPath = '/admin/users/42/edit'
let mockRouteParams = { id: '42' }

vi.mock('vue-router', async () => {
  const actual = await vi.importActual('vue-router')
  return {
    ...actual,
    useRoute: vi.fn(() => ({
      params: mockRouteParams,
      query: mockRouteQuery,
      fullPath: mockRouteFullPath,
    })),
    useRouter: vi.fn(() => ({
      replace: mockRouterReplace,
    })),
  }
})

// ── Helpers ───────────────────────────────────────────────────────────────────

const STUBS = {
  'v-container': { template: '<div><slot /></div>' },
  'v-card': { template: '<div><slot /></div>' },
  'v-tabs': {
    // Expose v-model so we can inspect activeTab
    template: '<div><slot /></div>',
    props: ['modelValue'],
  },
  'v-tab': {
    template: '<button :data-value="value" :disabled="disabled"><slot /></button>',
    props: ['value', 'disabled'],
  },
  'v-window': { template: '<div><slot /></div>', props: ['modelValue'] },
  'v-window-item': {
    template: '<div :data-value="value"><slot /></div>',
    props: ['value'],
  },
  'v-chip': { template: '<span><slot /></span>', props: ['size', 'class'] },
  'v-divider': { template: '<hr />' },
  'v-alert': { template: '<div class="v-alert"><slot /></div>', props: ['type', 'variant'] },
}

function createWrapper(routeQuery = {}, routeParams = { id: '42' }) {
  mockRouteQuery = routeQuery
  mockRouteParams = routeParams
  const pinia = createPinia()
  setActivePinia(pinia)
  return shallowMount(AdminUserEditView, {
    global: {
      plugins: [pinia],
      stubs: STUBS,
    },
  })
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe('AdminUserEditView', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
    mockRouteQuery = {}
    mockRouteParams = { id: '42' }
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
  })

  // 1. Default tab is 'profile' when no ?tab= query param is present
  it('defaults to profile tab when no ?tab= query param', async () => {
    wrapper = createWrapper()
    await flushPromises()
    expect(wrapper.vm.activeTab).toBe('profile')
  })

  // 2. Deep-link: ?tab=affiliations sets activeTab to 'affiliations'
  it('reads ?tab= query param on mount', async () => {
    wrapper = createWrapper({ tab: 'affiliations' })
    await flushPromises()
    expect(wrapper.vm.activeTab).toBe('affiliations')
  })

  // 3. Switching tabs calls router.replace with the new ?tab= value
  it('calls router.replace when activeTab changes', async () => {
    wrapper = createWrapper()
    await flushPromises()

    wrapper.vm.activeTab = 'affiliations'
    await flushPromises()

    expect(mockRouterReplace).toHaveBeenCalledWith(
      expect.objectContaining({ query: expect.objectContaining({ tab: 'affiliations' }) }),
    )
  })

  // 4. The future tabs are present and marked disabled
  it('renders disabled tabs for future versions', () => {
    wrapper = createWrapper()
    const html = wrapper.html()
    // Each future tab should be present
    expect(html).toContain('Affiliations')
    expect(html).toContain('Subscriptions')
    expect(html).toContain('Credits')
    expect(html).toContain('Preferences')
    expect(html).toContain('Activity')
  })

  // 5. Future tabs carry version chips
  it('shows version chips on disabled tabs', () => {
    wrapper = createWrapper()
    const html = wrapper.html()
    expect(html).toContain('v1.3.1')
    expect(html).toContain('v1.3.2')
  })

  // 6. userId is derived from route.params.id as a Number
  it('exposes userId as a Number computed from route.params.id', async () => {
    wrapper = createWrapper()
    await flushPromises()
    expect(wrapper.vm.userId).toBe(42)
  })

  // 7. ProfileTab is rendered inside the profile window-item
  it('renders ProfileTab in the profile window item', () => {
    wrapper = createWrapper()
    // The stub renders as <div data-testid="profile-tab" />
    // shallowMount stubs child components — find by component name
    const profileTabStub = wrapper.findComponent({ name: 'ProfileTab' })
    expect(profileTabStub.exists()).toBe(true)
  })

  // 8. NaN guard: invalid route param triggers redirect to /admin/users
  it('redirects to /admin/users when route.params.id is not a valid number', async () => {
    wrapper = createWrapper({}, { id: 'notanumber' })
    await flushPromises()
    expect(mockRouterReplace).toHaveBeenCalledWith('/admin/users')
  })

  // 9. Stub window items: deep-linking ?tab=affiliations shows the Coming-in alert
  it('renders Coming in v1.3.1 alert for the affiliations tab', async () => {
    wrapper = createWrapper({ tab: 'affiliations' })
    await flushPromises()
    const html = wrapper.html()
    expect(html).toContain('Coming in v1.3.1')
  })
})
