import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

/**
 * Font definitions with metadata
 *
 * Fonts are lazy-loaded when selected. The cssFile property maps to
 * the dynamic import path in src/assets/fonts/
 */
const AVAILABLE_FONTS = [
  {
    id: 'system',
    name: 'System Default',
    icon: 'mdi-laptop',
    description: "Uses your device's native font",
    cssClass: 'font-system',
    cssFile: null // No CSS file needed for system fonts
  },
  {
    id: 'inter',
    name: 'Inter',
    icon: 'mdi-format-font',
    description: 'Modern, highly readable UI font',
    cssClass: 'font-inter',
    cssFile: 'inter'
  },
  {
    id: 'roboto',
    name: 'Roboto',
    icon: 'mdi-android',
    description: 'Material Design standard',
    cssClass: 'font-roboto',
    cssFile: 'roboto'
  },
  {
    id: 'lato',
    name: 'Lato',
    icon: 'mdi-format-text',
    description: 'Warm humanist sans-serif',
    cssClass: 'font-lato',
    cssFile: 'lato'
  },
  {
    id: 'firasans',
    name: 'Fira Sans',
    icon: 'mdi-firefox',
    description: "Mozilla's UI font, slightly condensed",
    cssClass: 'font-firasans',
    cssFile: 'fira-sans'
  },
  {
    id: 'lexend',
    name: 'Lexend',
    icon: 'mdi-book-open-variant',
    description: 'Optimized for reading fluency',
    cssClass: 'font-lexend',
    cssFile: 'lexend'
  },
  {
    id: 'opendyslexic',
    name: 'OpenDyslexic',
    icon: 'mdi-human-greeting-proximity',
    description: 'Designed for readers with dyslexia',
    cssClass: 'font-opendyslexic',
    cssFile: 'opendyslexic',
    accessibility: true
  },
  {
    id: 'atkinson',
    name: 'Atkinson Hyperlegible',
    icon: 'mdi-eye',
    description: 'Optimized for low vision (Braille Institute)',
    cssClass: 'font-atkinson',
    cssFile: 'atkinson',
    accessibility: true
  },
  {
    id: 'sourceserif',
    name: 'Source Serif Pro',
    icon: 'mdi-format-text-variant',
    description: 'Classic serif typeface',
    cssClass: 'font-sourceserif',
    cssFile: 'source-serif'
  },
  {
    id: 'jetbrainsmono',
    name: 'JetBrains Mono',
    icon: 'mdi-code-braces',
    description: 'Developer-focused monospace',
    cssClass: 'font-jetbrainsmono',
    cssFile: 'jetbrains-mono'
  }
]

// Map of font CSS file imports for lazy loading
const fontImports = {
  'inter': () => import('@/assets/fonts/inter.css'),
  'roboto': () => import('@/assets/fonts/roboto.css'),
  'lato': () => import('@/assets/fonts/lato.css'),
  'fira-sans': () => import('@/assets/fonts/fira-sans.css'),
  'lexend': () => import('@/assets/fonts/lexend.css'),
  'opendyslexic': () => import('@/assets/fonts/opendyslexic.css'),
  'atkinson': () => import('@/assets/fonts/atkinson.css'),
  'source-serif': () => import('@/assets/fonts/source-serif.css'),
  'jetbrains-mono': () => import('@/assets/fonts/jetbrains-mono.css')
}

// localStorage key for caching
const STORAGE_KEY = 'actalog-font-family'

export const useFontStore = defineStore('font', () => {
  // State
  const currentFont = ref(localStorage.getItem(STORAGE_KEY) || 'system')
  const fontsLoaded = ref(new Set(['system'])) // Track which fonts have been loaded
  const fontLoading = ref(false) // Track loading state

  // Computed
  const availableFonts = computed(() => AVAILABLE_FONTS)

  const currentFontDetails = computed(() => {
    return AVAILABLE_FONTS.find(f => f.id === currentFont.value) || AVAILABLE_FONTS[0]
  })

  const accessibilityFonts = computed(() => {
    return AVAILABLE_FONTS.filter(f => f.accessibility)
  })

  /**
   * Load font CSS dynamically
   * @param {string} cssFile - The font CSS file name (without .css extension)
   * @returns {Promise<boolean>} - Whether the font was loaded successfully
   */
  async function loadFontCSS(cssFile) {
    if (!cssFile || fontsLoaded.value.has(cssFile)) {
      return true // Already loaded or no CSS needed
    }

    const importFn = fontImports[cssFile]
    if (!importFn) {
      console.warn(`No import function for font CSS: ${cssFile}`)
      return false
    }

    try {
      fontLoading.value = true
      await importFn()
      fontsLoaded.value.add(cssFile)
      console.log(`Font CSS loaded: ${cssFile}`)
      return true
    } catch (error) {
      console.error(`Failed to load font CSS: ${cssFile}`, error)
      return false
    } finally {
      fontLoading.value = false
    }
  }

  // Actions
  async function setFont(fontId) {
    const font = AVAILABLE_FONTS.find(f => f.id === fontId)
    if (!font) {
      console.warn(`Font "${fontId}" not found, defaulting to system`)
      fontId = 'system'
    }

    // Load font CSS if needed
    if (font && font.cssFile) {
      await loadFontCSS(font.cssFile)
    }

    currentFont.value = fontId
    localStorage.setItem(STORAGE_KEY, fontId)

    // Apply font class to document
    applyFontClass(fontId)
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
  async function initialize() {
    // Apply saved font on startup
    const savedFont = localStorage.getItem(STORAGE_KEY)
    if (savedFont && savedFont !== 'system') {
      await setFont(savedFont)
    } else {
      applyFontClass('system')
    }
  }

  // Sync with backend settings
  async function syncFromSettings(fontFamily) {
    if (fontFamily && fontFamily !== currentFont.value) {
      await setFont(fontFamily)
    }
  }

  // Call initialize (async, but we don't wait for it)
  initialize()

  return {
    // State
    currentFont,
    fontsLoaded,
    fontLoading,
    // Computed
    availableFonts,
    currentFontDetails,
    accessibilityFonts,
    // Actions
    setFont,
    loadFontCSS,
    applyFontClass,
    initialize,
    syncFromSettings
  }
})
