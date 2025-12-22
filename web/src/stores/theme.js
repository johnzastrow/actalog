import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { useTheme } from 'vuetify'

export const useThemeStore = defineStore('theme', () => {
  const vuetifyTheme = useTheme()

  // Available themes
  const availableThemes = [
    {
      id: 'light',
      name: 'Light',
      icon: 'mdi-white-balance-sunny',
      description: 'Clean light theme for daytime use'
    },
    {
      id: 'dark',
      name: 'Dark',
      icon: 'mdi-moon-waning-crescent',
      description: 'Easy on the eyes for night training'
    },
    {
      id: 'brutalist',
      name: 'Athletic Brutalist',
      icon: 'mdi-weight-lifter',
      description: 'Bold, high-contrast performance theme'
    },
    {
      id: 'ocean',
      name: 'Ocean',
      icon: 'mdi-waves',
      description: 'Calming blue tones'
    },
    {
      id: 'sunset',
      name: 'Sunset',
      icon: 'mdi-weather-sunset',
      description: 'Warm amber and orange hues'
    }
  ]

  // Current theme - load from localStorage or default to 'light'
  const currentTheme = ref(localStorage.getItem('actalog-theme') || 'light')

  // Apply theme
  const setTheme = (themeId) => {
    if (!availableThemes.find(t => t.id === themeId)) {
      console.warn(`Theme "${themeId}" not found, defaulting to light`)
      themeId = 'light'
    }

    currentTheme.value = themeId
    vuetifyTheme.global.name.value = themeId
    localStorage.setItem('actalog-theme', themeId)
  }

  // Get current theme details
  const getCurrentThemeDetails = () => {
    return availableThemes.find(t => t.id === currentTheme.value) || availableThemes[0]
  }

  // Toggle between light and dark (for quick toggle)
  const toggleLightDark = () => {
    const newTheme = currentTheme.value === 'dark' ? 'light' : 'dark'
    setTheme(newTheme)
  }

  // Initialize theme on store creation
  setTheme(currentTheme.value)

  // Watch for system preference changes (optional)
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)')
  const useSystemPreference = ref(localStorage.getItem('actalog-use-system-theme') === 'true')

  const setUseSystemPreference = (value) => {
    useSystemPreference.value = value
    localStorage.setItem('actalog-use-system-theme', value.toString())

    if (value) {
      setTheme(prefersDark.matches ? 'dark' : 'light')
    }
  }

  // Listen for system theme changes
  prefersDark.addEventListener('change', (e) => {
    if (useSystemPreference.value) {
      setTheme(e.matches ? 'dark' : 'light')
    }
  })

  return {
    currentTheme,
    availableThemes,
    setTheme,
    toggleLightDark,
    getCurrentThemeDetails,
    useSystemPreference,
    setUseSystemPreference
  }
})
