import { describe, it, expect, vi, afterEach } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import NotFoundView from './NotFoundView.vue'

describe('NotFoundView', () => {
  let wrapper

  const createWrapper = () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = shallowMount(NotFoundView, {
      global: {
        plugins: [pinia],
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-row': { template: '<div><slot /></div>' },
          'v-col': { template: '<div><slot /></div>' },
          'v-icon': { template: '<i></i>' },
          'v-btn': { template: '<button><slot /></button>' }
        }
      }
    })

    return wrapper
  }

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
      wrapper = null
    }
  })

  it('renders 404 page', () => {
    createWrapper()

    expect(wrapper.text()).toContain('404')
    expect(wrapper.text()).toContain('Page Not Found')
  })

  it('shows helpful message', () => {
    createWrapper()

    expect(wrapper.text()).toContain("doesn't exist or has been moved")
  })

  it('has link to dashboard', () => {
    createWrapper()

    expect(wrapper.text()).toContain('Dashboard')
  })
})
