import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from '@/utils/axios'

export const useWodsStore = defineStore('wods', () => {
  const wods = ref([])
  const currentWod = ref(null)
  const loading = ref(false)
  const error = ref(null)

  // Fetch all WODs (standard + user custom)
  async function fetchWods() {
    loading.value = true
    error.value = null
    try {
      const response = await axios.get('/api/wods')
      wods.value = response.data.wods || []
      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to fetch WODs'
      console.error('Failed to fetch WODs:', e)
      return false
    } finally {
      loading.value = false
    }
  }

  // Fetch single WOD by ID
  async function fetchWodById(id) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.get(`/api/wods/${id}`)
      currentWod.value = response.data
      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to fetch WOD'
      console.error('Failed to fetch WOD:', e)
      return false
    } finally {
      loading.value = false
    }
  }

  // Search WODs by name
  async function searchWods(query) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.get('/api/wods/search', {
        params: { q: query }
      })
      wods.value = response.data.wods || []
      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to search WODs'
      console.error('Failed to search WODs:', e)
      return false
    } finally {
      loading.value = false
    }
  }

  // Create custom WOD (requires authentication)
  async function createWod(wodData) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.post('/api/wods', wodData)
      wods.value.push(response.data)
      return response.data
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to create WOD'
      console.error('Failed to create WOD:', e)
      return null
    } finally {
      loading.value = false
    }
  }

  // Update custom WOD (requires authentication)
  async function updateWod(id, wodData) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.put(`/api/wods/${id}`, wodData)
      const index = wods.value.findIndex(w => w.id === id)
      if (index !== -1) {
        wods.value[index] = response.data
      }
      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to update WOD'
      console.error('Failed to update WOD:', e)
      return false
    } finally {
      loading.value = false
    }
  }

  // Delete custom WOD (requires authentication)
  async function deleteWod(id) {
    loading.value = true
    error.value = null
    try {
      await axios.delete(`/api/wods/${id}`)
      wods.value = wods.value.filter(w => w.id !== id)
      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to delete WOD'
      console.error('Failed to delete WOD:', e)
      return false
    } finally {
      loading.value = false
    }
  }

  // Filter WODs by type
  function filterByType(type) {
    if (!type) return wods.value
    return wods.value.filter(w => w.type === type)
  }

  // Filter WODs by source
  function filterBySource(source) {
    if (!source) return wods.value
    return wods.value.filter(w => w.source === source)
  }

  // Get standard (pre-seeded) WODs only
  function getStandardWods() {
    return wods.value.filter(w => w.is_standard)
  }

  // Get custom (user-created) WODs only
  function getCustomWods() {
    return wods.value.filter(w => !w.is_standard)
  }

  return {
    wods,
    currentWod,
    loading,
    error,
    fetchWods,
    fetchWodById,
    searchWods,
    createWod,
    updateWod,
    deleteWod,
    filterByType,
    filterBySource,
    getStandardWods,
    getCustomWods
  }
})
