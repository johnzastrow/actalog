import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// Font definitions with metadata
const AVAILABLE_FONTS = [
  {
    id: 'system',
    name: 'System Default',
    icon: 'mdi-laptop',
    description: "Uses your device's native font",
    cssClass: 'font-system',
    preload: false
  },
  {
    id: 'inter',
    name: 'Inter',
    icon: 'mdi-format-font',
    description: 'Modern, highly readable UI font',
    cssClass: 'font-inter',
    preload: true
  },
  {
    id: 'roboto',
    name: 'Roboto',
    icon: 'mdi-android',
    description: 'Material Design standard',
    cssClass: 'font-roboto',
    preload: true
  },
  {
    id: 'lato',
    name: 'Lato',
    icon: 'mdi-format-text',
    description: 'Warm humanist sans-serif',
    cssClass: 'font-lato',
    preload: true
  },
  {
    id: 'firasans',
    name: 'Fira Sans',
    icon: 'mdi-firefox',
    description: "Mozilla's UI font, slightly condensed",
    cssClass: 'font-firasans',
    preload: true
  },
  {
    id: 'lexend',
    name: 'Lexend',
    icon: 'mdi-book-open-variant',
    description: 'Optimized for reading fluency',
    cssClass: 'font-lexend',
    preload: true
  },
  {
    id: 'opendyslexic',
    name: 'OpenDyslexic',
    icon: 'mdi-human-greeting-proximity',
    description: 'Designed for readers with dyslexia',
    cssClass: 'font-opendyslexic',
    preload: true,
    accessibility: true
  },
  {
    id: 'atkinson',
    name: 'Atkinson Hyperlegible',
    icon: 'mdi-eye',
    description: 'Optimized for low vision (Braille Institute)',
    cssClass: 'font-atkinson',
    preload: true,
    accessibility: true
  },
  {
    id: 'sourceserif',
    name: 'Source Serif Pro',
    icon: 'mdi-format-text-variant',
    description: 'Classic serif typeface',
    cssClass: 'font-sourceserif',
    preload: true
  },
  {
    id: 'jetbrainsmono',
    name: 'JetBrains Mono',
    icon: 'mdi-code-braces',
    description: 'Developer-focused monospace',
    cssClass: 'font-jetbrainsmono',
    preload: true
  }
]

// localStorage key for caching
const STORAGE_KEY = 'actalog-font-family'

export const useFontStore = defineStore('font', () => {
  // State
  const currentFont = ref(localStorage.getItem(STORAGE_KEY) || 'system')
  const fontsLoaded = ref(new Set(['system'])) // Track which fonts have been loaded

  // Computed
  const availableFonts = computed(() => AVAILABLE_FONTS)

  const currentFontDetails = computed(() => {
    return AVAILABLE_FONTS.find(f => f.id === currentFont.value) || AVAILABLE_FONTS[0]
  })

  const accessibilityFonts = computed(() => {
    return AVAILABLE_FONTS.filter(f => f.accessibility)
  })

  // Actions
  function setFont(fontId) {
    const font = AVAILABLE_FONTS.find(f => f.id === fontId)
    if (!font) {
      console.warn(`Font "${fontId}" not found, defaulting to system`)
      fontId = 'system'
    }

    currentFont.value = fontId
    localStorage.setItem(STORAGE_KEY, fontId)

    // Apply font class to document
    applyFontClass(fontId)

    // Mark font as loaded
    if (font && font.preload) {
      fontsLoaded.value.add(fontId)
    }
  }

  function applyFontClass(fontId) {
    const font = AVAILABLE_FONTS.find(f => f.id === fontId)
    if (!font) return

    // Remove all font classes from html element
    AVAILABLE_FONTS.forEach(f => {
      document.documentElement.classList.remove(f.cssClass)
    })

    // Add the selected font class
    document.documentElement.classList.add(font.cssClass)
  }

  // Initialize on store creation
  function initialize() {
    // Apply saved font on startup
    const savedFont = localStorage.getItem(STORAGE_KEY)
    if (savedFont) {
      setFont(savedFont)
    } else {
      applyFontClass('system')
    }
  }

  // Sync with backend settings
  function syncFromSettings(fontFamily) {
    if (fontFamily && fontFamily !== currentFont.value) {
      setFont(fontFamily)
    }
  }

  // Call initialize
  initialize()

  return {
    // State
    currentFont,
    fontsLoaded,
    // Computed
    availableFonts,
    currentFontDetails,
    accessibilityFonts,
    // Actions
    setFont,
    applyFontClass,
    initialize,
    syncFromSettings
  }
})
