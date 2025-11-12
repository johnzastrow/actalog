import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from '@/utils/axios'

export const useTemplatesStore = defineStore('templates', () => {
  const templates = ref([])
  const currentTemplate = ref(null)
  const loading = ref(false)
  const error = ref(null)

  // Fetch all templates (standard + user custom)
  async function fetchTemplates() {
    loading.value = true
    error.value = null
    try {
      const response = await axios.get('/api/templates')
      templates.value = response.data.templates || []
      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to fetch templates'
      console.error('Failed to fetch templates:', e)
      return false
    } finally {
      loading.value = false
    }
  }

  // Fetch single template by ID with full movement details
  async function fetchTemplateById(id) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.get(`/api/templates/${id}`)
      currentTemplate.value = response.data
      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to fetch template'
      console.error('Failed to fetch template:', e)
      return false
    } finally {
      loading.value = false
    }
  }

  // Fetch user's custom templates
  async function fetchMyTemplates() {
    loading.value = true
    error.value = null
    try {
      const response = await axios.get('/api/workouts/my-templates')
      templates.value = response.data.templates || []
      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to fetch my templates'
      console.error('Failed to fetch my templates:', e)
      return false
    } finally {
      loading.value = false
    }
  }

  // Create custom template (requires authentication)
  async function createTemplate(templateData) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.post('/api/templates', templateData)
      templates.value.push(response.data)
      return response.data
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to create template'
      console.error('Failed to create template:', e)
      return null
    } finally {
      loading.value = false
    }
  }

  // Update custom template (requires authentication)
  async function updateTemplate(id, templateData) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.put(`/api/templates/${id}`, templateData)
      const index = templates.value.findIndex(t => t.id === id)
      if (index !== -1) {
        templates.value[index] = response.data
      }
      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to update template'
      console.error('Failed to update template:', e)
      return false
    } finally {
      loading.value = false
    }
  }

  // Delete custom template (requires authentication)
  async function deleteTemplate(id) {
    loading.value = true
    error.value = null
    try {
      await axios.delete(`/api/templates/${id}`)
      templates.value = templates.value.filter(t => t.id !== id)
      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to delete template'
      console.error('Failed to delete template:', e)
      return false
    } finally {
      loading.value = false
    }
  }

  // Fetch WODs associated with a template
  async function fetchTemplateWods(templateId) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.get(`/api/templates/${templateId}/wods`)
      return response.data.wods || []
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to fetch template WODs'
      console.error('Failed to fetch template WODs:', e)
      return []
    } finally {
      loading.value = false
    }
  }

  // Add WOD to template
  async function addWodToTemplate(templateId, wodId, orderIndex = 1) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.post(`/api/templates/${templateId}/wods`, {
        wod_id: wodId,
        order_index: orderIndex
      })
      return response.data
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to add WOD to template'
      console.error('Failed to add WOD to template:', e)
      return null
    } finally {
      loading.value = false
    }
  }

  // Remove WOD from template
  async function removeWodFromTemplate(workoutWodId) {
    loading.value = true
    error.value = null
    try {
      await axios.delete(`/api/templates/wods/${workoutWodId}`)
      return true
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to remove WOD from template'
      console.error('Failed to remove WOD from template:', e)
      return false
    } finally {
      loading.value = false
    }
  }

  // Toggle PR tracking for a WOD in a template
  async function toggleWodPR(workoutWodId) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.post(`/api/templates/wods/${workoutWodId}/toggle-pr`)
      return response.data
    } catch (e) {
      error.value = e.response?.data?.message || 'Failed to toggle WOD PR'
      console.error('Failed to toggle WOD PR:', e)
      return null
    } finally {
      loading.value = false
    }
  }

  // Filter templates by type
  function filterByType(type) {
    if (!type) return templates.value
    return templates.value.filter(t => {
      // Type can be inferred from movements or template name
      // This is a simple implementation - enhance based on your needs
      return t.workout_name && t.workout_name.toLowerCase().includes(type.toLowerCase())
    })
  }

  // Get standard (pre-seeded) templates only
  function getStandardTemplates() {
    return templates.value.filter(t => t.is_standard)
  }

  // Get custom (user-created) templates only
  function getCustomTemplates() {
    return templates.value.filter(t => !t.is_standard)
  }

  // Get templates with movements count
  function getTemplatesWithMovementCount() {
    return templates.value.map(t => ({
      ...t,
      movementCount: t.movements ? t.movements.length : 0
    }))
  }

  return {
    templates,
    currentTemplate,
    loading,
    error,
    fetchTemplates,
    fetchTemplateById,
    fetchMyTemplates,
    createTemplate,
    updateTemplate,
    deleteTemplate,
    fetchTemplateWods,
    addWodToTemplate,
    removeWodFromTemplate,
    toggleWodPR,
    filterByType,
    getStandardTemplates,
    getCustomTemplates,
    getTemplatesWithMovementCount
  }
})
