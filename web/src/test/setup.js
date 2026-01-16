import { config } from '@vue/test-utils'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { beforeAll, afterEach, vi } from 'vitest'

// Create Vuetify instance for tests
const vuetify = createVuetify({
  components,
  directives,
})

// Configure Vue Test Utils globally
config.global.plugins = [vuetify]

// Mock ResizeObserver (not available in jsdom)
beforeAll(() => {
  global.ResizeObserver = class ResizeObserver {
    constructor(callback) {
      this.callback = callback
    }
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

// Clean up after each test
afterEach(() => {
  vi.clearAllMocks()
})

// Mock window.matchMedia (not available in jsdom)
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Mock IntersectionObserver (not available in jsdom)
global.IntersectionObserver = class IntersectionObserver {
  constructor(callback) {
    this.callback = callback
  }
  observe() {}
  unobserve() {}
  disconnect() {}
}

// Mock scrollTo (not available in jsdom)
window.scrollTo = vi.fn()

// Mock visualViewport (not available in jsdom, needed by Vuetify VDialog)
if (!global.visualViewport) {
  global.visualViewport = {
    width: 1024,
    height: 768,
    offsetLeft: 0,
    offsetTop: 0,
    pageLeft: 0,
    pageTop: 0,
    scale: 1,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  }
}
