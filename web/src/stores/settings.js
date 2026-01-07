import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from '@/utils/axios'
import { getBrowserTimezone, TIMEZONE_OPTIONS } from '@/utils/timezone'
import { useFontStore } from '@/stores/font'

export const useSettingsStore = defineStore('settings', () => {
  // State
  const settings = ref(null)
  const loading = ref(false)
  const error = ref(null)

  // Computed
  const timezone = computed(() => {
    return settings.value?.timezone || 'America/New_York'
  })

  const weightUnit = computed(() => {
    return settings.value?.weight_unit || 'lbs'
  })

  const distanceUnit = computed(() => {
    return settings.value?.distance_unit || 'miles'
  })

  const dataExportFormat = computed(() => {
    return settings.value?.data_export_format || 'JSON'
  })

  const theme = computed(() => {
    return settings.value?.theme || 'light'
  })

  const fontFamily = computed(() => {
    return settings.value?.font_family || 'system'
  })

  const adminUserEventNotifications = computed(() => {
    return settings.value?.admin_user_event_notifications ?? true
  })

  // Actions
  async function fetchSettings() {
    loading.value = true
    error.value = null

    try {
      const response = await axios.get('/api/users/settings')
      settings.value = response.data

      // Sync font with font store
      const fontStore = useFontStore()
      fontStore.syncFromSettings(response.data.font_family)

      return settings.value
    } catch (err) {
      console.error('Failed to fetch settings:', err)
      error.value = err.response?.data?.error || 'Failed to load settings'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateSettings(updates) {
    loading.value = true
    error.value = null

    try {
      const response = await axios.put('/api/users/settings', updates)
      settings.value = response.data
      return settings.value
    } catch (err) {
      console.error('Failed to update settings:', err)
      error.value = err.response?.data?.error || 'Failed to update settings'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateTimezone(newTimezone) {
    // Validate timezone
    const isValid = TIMEZONE_OPTIONS.some(opt => opt.value === newTimezone) ||
      newTimezone === getBrowserTimezone()

    if (!isValid) {
      error.value = 'Invalid timezone'
      throw new Error('Invalid timezone')
    }

    return updateSettings({
      ...settings.value,
      timezone: newTimezone
    })
  }

  async function updateWeightUnit(newUnit) {
    if (!['lbs', 'kg'].includes(newUnit)) {
      error.value = 'Invalid weight unit'
      throw new Error('Invalid weight unit')
    }

    return updateSettings({
      ...settings.value,
      weight_unit: newUnit
    })
  }

  async function updateDistanceUnit(newUnit) {
    if (!['miles', 'km'].includes(newUnit)) {
      error.value = 'Invalid distance unit'
      throw new Error('Invalid distance unit')
    }

    return updateSettings({
      ...settings.value,
      distance_unit: newUnit
    })
  }

  async function updateFontFamily(newFont) {
    const validFonts = ['system', 'inter', 'roboto', 'lato', 'firasans', 'lexend', 'opendyslexic', 'atkinson', 'sourceserif', 'jetbrainsmono']
    if (!validFonts.includes(newFont)) {
      error.value = 'Invalid font family'
      throw new Error('Invalid font family')
    }

    // Update font store immediately for instant feedback
    const fontStore = useFontStore()
    fontStore.setFont(newFont)

    return updateSettings({
      ...settings.value,
      font_family: newFont
    })
  }

  async function updateAdminUserEventNotifications(enabled) {
    return updateSettings({
      ...settings.value,
      admin_user_event_notifications: enabled
    })
  }

  function detectBrowserTimezone() {
    return getBrowserTimezone()
  }

  // Initialize - fetch settings when store is created (if user is logged in)
  const token = localStorage.getItem('token')
  if (token) {
    fetchSettings().catch(() => {
      // Silently fail on initial load - settings will be fetched when needed
    })
  }

  return {
    // State
    settings,
    loading,
    error,
    // Computed
    timezone,
    weightUnit,
    distanceUnit,
    dataExportFormat,
    theme,
    fontFamily,
    adminUserEventNotifications,
    // Actions
    fetchSettings,
    updateSettings,
    updateTimezone,
    updateWeightUnit,
    updateDistanceUnit,
    updateFontFamily,
    updateAdminUserEventNotifications,
    detectBrowserTimezone
  }
})
