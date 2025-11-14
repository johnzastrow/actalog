import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from '@/utils/axios'
import { getProfileImageUrl } from '@/utils/url'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('token') || null)
  const refreshToken = ref(localStorage.getItem('refreshToken') || null)
  const loading = ref(false)
  const error = ref(null)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  // Initialize auth state from localStorage
  function init() {
    const savedUser = localStorage.getItem('user')
    if (savedUser && token.value) {
      try {
        const parsedUser = JSON.parse(savedUser)
        // Ensure profile_image has full URL for display
        if (parsedUser.profile_image) {
          parsedUser.profile_image = getProfileImageUrl(parsedUser.profile_image)
        }
        user.value = parsedUser
        // Set default authorization header
        axios.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
      } catch (e) {
        logout()
      }
    }
  }

  async function login(email, password, rememberMe = false) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.post('/api/auth/login', {
        email,
        password,
        remember_me: rememberMe
      })
      token.value = response.data.token
      user.value = response.data.user

      // Save to localStorage
      localStorage.setItem('token', token.value)
      localStorage.setItem('user', JSON.stringify(user.value))

      // Save refresh token if provided (when remember_me is true)
      if (response.data.refresh_token) {
        refreshToken.value = response.data.refresh_token
        localStorage.setItem('refreshToken', refreshToken.value)
      }

      // Set default authorization header
      axios.defaults.headers.common['Authorization'] = `Bearer ${token.value}`

      return true
    } catch (e) {
      console.error('Login error:', e)
      if (e.response) {
        // Server responded with error status
        error.value = e.response.data?.message || `Server error: ${e.response.status}`
      } else if (e.request) {
        // Request was made but no response received
        error.value = 'Cannot connect to server. Please check if the server is running.'
      } else {
        // Something else happened
        error.value = e.message || 'Login failed'
      }
      return false
    } finally {
      loading.value = false
    }
  }

  async function register(userData) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.post('/api/auth/register', userData)
      token.value = response.data.token
      user.value = response.data.user

      localStorage.setItem('token', token.value)
      localStorage.setItem('user', JSON.stringify(user.value))

      axios.defaults.headers.common['Authorization'] = `Bearer ${token.value}`

      return true
    } catch (e) {
      console.error('Registration error:', e)
      if (e.response) {
        // Server responded with error status
        error.value = e.response.data?.message || `Server error: ${e.response.status}`
      } else if (e.request) {
        // Request was made but no response received
        error.value = 'Cannot connect to server. Please check if the server is running.'
      } else {
        // Something else happened
        error.value = e.message || 'Registration failed'
      }
      return false
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    // Revoke refresh token if it exists
    if (refreshToken.value) {
      try {
        await axios.post('/api/auth/revoke', {
          refresh_token: refreshToken.value
        })
      } catch (e) {
        // Ignore errors during revocation
        console.error('Failed to revoke refresh token:', e)
      }
    }

    user.value = null
    token.value = null
    refreshToken.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    localStorage.removeItem('refreshToken')
    delete axios.defaults.headers.common['Authorization']
  }

  async function refreshAccessToken() {
    if (!refreshToken.value) {
      return false
    }

    try {
      const response = await axios.post('/api/auth/refresh', {
        refresh_token: refreshToken.value
      })

      token.value = response.data.token
      user.value = response.data.user

      // Update localStorage
      localStorage.setItem('token', token.value)
      localStorage.setItem('user', JSON.stringify(user.value))

      // Set default authorization header
      axios.defaults.headers.common['Authorization'] = `Bearer ${token.value}`

      return true
    } catch (e) {
      // Refresh token is invalid or expired, logout
      logout()
      return false
    }
  }

  async function updateProfile(updates) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.put('/api/users/profile', updates)
      user.value = response.data.user
      localStorage.setItem('user', JSON.stringify(user.value))
      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Profile update failed'
      return false
    } finally {
      loading.value = false
    }
  }

  // Initialize on store creation
  init()

  return {
    user,
    token,
    refreshToken,
    loading,
    error,
    isAuthenticated,
    login,
    register,
    logout,
    refreshAccessToken,
    updateProfile
  }
})
