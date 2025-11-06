import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from '@/utils/axios'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('token') || null)
  const loading = ref(false)
  const error = ref(null)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  // Initialize auth state from localStorage
  function init() {
    const savedUser = localStorage.getItem('user')
    if (savedUser && token.value) {
      try {
        user.value = JSON.parse(savedUser)
        // Set default authorization header
        axios.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
      } catch (e) {
        logout()
      }
    }
  }

  async function login(email, password) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.post('/api/auth/login', { email, password })
      token.value = response.data.token
      user.value = response.data.user

      // Save to localStorage
      localStorage.setItem('token', token.value)
      localStorage.setItem('user', JSON.stringify(user.value))

      // Set default authorization header
      axios.defaults.headers.common['Authorization'] = `Bearer ${token.value}`

      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Login failed'
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
      error.value = e.response?.data?.message || 'Registration failed'
      return false
    } finally {
      loading.value = false
    }
  }

  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    delete axios.defaults.headers.common['Authorization']
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
    loading,
    error,
    isAuthenticated,
    login,
    register,
    logout,
    updateProfile
  }
})
