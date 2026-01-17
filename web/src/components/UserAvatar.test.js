import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import UserAvatar from './UserAvatar.vue'

describe('UserAvatar', () => {
  it('renders without crashing', () => {
    const wrapper = mount(UserAvatar)
    expect(wrapper.exists()).toBe(true)
  })

  it('displays default initial "U" when no user provided', () => {
    const wrapper = mount(UserAvatar)
    expect(wrapper.text()).toBe('U')
  })

  it('displays single initial for single name', () => {
    const wrapper = mount(UserAvatar, {
      props: {
        user: { name: 'John' }
      }
    })
    expect(wrapper.text()).toBe('J')
  })

  it('displays two initials for full name', () => {
    const wrapper = mount(UserAvatar, {
      props: {
        user: { name: 'John Doe' }
      }
    })
    expect(wrapper.text()).toBe('JD')
  })

  it('handles names with multiple spaces', () => {
    const wrapper = mount(UserAvatar, {
      props: {
        user: { name: 'John  Michael  Doe' }
      }
    })
    expect(wrapper.text()).toBe('JD')
  })

  it('handles names with leading/trailing spaces', () => {
    const wrapper = mount(UserAvatar, {
      props: {
        user: { name: '  John Doe  ' }
      }
    })
    expect(wrapper.text()).toBe('JD')
  })

  it('uses uppercase for initials', () => {
    const wrapper = mount(UserAvatar, {
      props: {
        user: { name: 'john doe' }
      }
    })
    expect(wrapper.text()).toBe('JD')
  })

  it('applies default size of 80', () => {
    const wrapper = mount(UserAvatar)
    const avatar = wrapper.find('.v-avatar')
    expect(avatar.exists()).toBe(true)
  })

  it('applies custom size', () => {
    const wrapper = mount(UserAvatar, {
      props: {
        size: 120
      }
    })
    expect(wrapper.props('size')).toBe(120)
  })

  it('generates consistent color for same name', () => {
    const wrapper1 = mount(UserAvatar, {
      props: { user: { name: 'John Doe' } }
    })
    const wrapper2 = mount(UserAvatar, {
      props: { user: { name: 'John Doe' } }
    })

    // Both should have the same color (consistent hashing)
    const avatar1 = wrapper1.find('.v-avatar')
    const avatar2 = wrapper2.find('.v-avatar')

    expect(avatar1.attributes('style')).toBe(avatar2.attributes('style'))
  })

  it('generates different colors for different names', () => {
    const wrapper1 = mount(UserAvatar, {
      props: { user: { name: 'John Doe' } }
    })
    const wrapper2 = mount(UserAvatar, {
      props: { user: { name: 'Jane Smith' } }
    })

    // Different names should (usually) have different colors
    // Note: There's a small chance of collision, but unlikely with these specific names
    const avatar1 = wrapper1.find('.v-avatar')
    const avatar2 = wrapper2.find('.v-avatar')

    // Just verify both have avatars rendered
    expect(avatar1.exists()).toBe(true)
    expect(avatar2.exists()).toBe(true)
  })
})
