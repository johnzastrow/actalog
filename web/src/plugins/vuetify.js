import { createVuetify } from 'vuetify'
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'
import { aliases, mdi } from 'vuetify/iconsets/mdi'

// ActaLog custom theme based on design requirements
const actalogTheme = {
  dark: false,
  colors: {
    primary: '#2c3657',      // Dark blue for headers
    secondary: '#597a6a',    // Secondary color
    accent: '#5a4e68',       // Accent color
    teal: '#00bcd4',         // Teal for buttons (logo color)
    error: '#DF3F40',        // Error/Alert color
    warning: '#FFA726',
    info: '#29B6F6',
    success: '#66BB6A',
    background: '#FFFFFF',
    surface: '#FFFFFF',
    'on-primary': '#FFFFFF',
    'on-secondary': '#FFFFFF',
    'on-accent': '#FFFFFF',
    'on-teal': '#FFFFFF',
    'on-background': '#1A1A1A',
    'on-surface': '#1A1A1A',
    // Additional colors
    gold: '#FFD700',         // For action button (+ button)
    border: '#E3E6EA',       // Border color from requirements
  }
}

const actalogDarkTheme = {
  dark: true,
  colors: {
    primary: '#2c3657',      // Dark blue for headers
    secondary: '#597a6a',
    accent: '#5a4e68',
    teal: '#00bcd4',         // Teal for buttons (logo color)
    error: '#DF3F40',
    warning: '#FFA726',
    info: '#29B6F6',
    success: '#66BB6A',
    background: '#1A1A1A',
    surface: '#2A2A2A',
    'on-primary': '#FFFFFF',
    'on-secondary': '#FFFFFF',
    'on-accent': '#FFFFFF',
    'on-teal': '#FFFFFF',
    'on-background': '#FFFFFF',
    'on-surface': '#FFFFFF',
    gold: '#FFD700',
    border: '#3A3A3A',
  }
}

// Athletic Brutalist Theme - Bold, high-contrast performance theme
const brutalistTheme = {
  dark: true,
  colors: {
    primary: '#00bcd4',      // Cyan - matches marketing site
    secondary: '#ffc107',    // Gold accent
    accent: '#ff6b35',       // Orange accent for intensity
    teal: '#00bcd4',
    error: '#ff1744',
    warning: '#ffc107',
    info: '#00bcd4',
    success: '#00e676',
    background: '#0a1628',   // Dark navy from marketing site
    surface: '#0f1d2e',
    'on-primary': '#0a1628',
    'on-secondary': '#0a1628',
    'on-accent': '#FFFFFF',
    'on-teal': '#0a1628',
    'on-background': '#FFFFFF',
    'on-surface': '#e0e0e0',
    gold: '#ffc107',
    border: 'rgba(0, 188, 212, 0.2)',
  }
}

// Ocean Theme - Calming blue tones
const oceanTheme = {
  dark: false,
  colors: {
    primary: '#006d77',      // Deep ocean teal
    secondary: '#83c5be',    // Light aqua
    accent: '#edf6f9',       // Foam white
    teal: '#00bcd4',
    error: '#e63946',
    warning: '#f4a261',
    info: '#2a9d8f',
    success: '#06d6a0',
    background: '#fdfcfa',
    surface: '#FFFFFF',
    'on-primary': '#FFFFFF',
    'on-secondary': '#1a1a1a',
    'on-accent': '#1a1a1a',
    'on-teal': '#FFFFFF',
    'on-background': '#1a1a1a',
    'on-surface': '#1a1a1a',
    gold: '#f4a261',
    border: '#e0e0e0',
  }
}

// Sunset Theme - Warm amber and orange hues
const sunsetTheme = {
  dark: false,
  colors: {
    primary: '#d62828',      // Deep red
    secondary: '#f77f00',    // Orange
    accent: '#fcbf49',       // Golden yellow
    teal: '#00bcd4',
    error: '#d62828',
    warning: '#fcbf49',
    info: '#457b9d',
    success: '#06d6a0',
    background: '#fffcf2',   // Warm cream
    surface: '#FFFFFF',
    'on-primary': '#FFFFFF',
    'on-secondary': '#1a1a1a',
    'on-accent': '#1a1a1a',
    'on-teal': '#FFFFFF',
    'on-background': '#1a1a1a',
    'on-surface': '#1a1a1a',
    gold: '#fcbf49',
    border: '#e0e0e0',
  }
}

// Sunrise Theme - Inspired by beach sunrise with purple clouds, golden horizon, coral sky
const sunriseTheme = {
  dark: false,
  colors: {
    primary: '#7B5D7F',      // Muted purple from sky/clouds
    secondary: '#E8842A',    // Golden orange from horizon
    accent: '#D4637A',       // Coral pink from clouds
    teal: '#00bcd4',
    error: '#D4637A',
    warning: '#F5A623',      // Amber gold
    info: '#3D4F6F',         // Deep blue-purple from ocean
    success: '#66BB6A',
    background: '#FFF8F5',   // Warm cream with pink tint
    surface: '#FFFFFF',
    'on-primary': '#FFFFFF',
    'on-secondary': '#FFFFFF',
    'on-accent': '#FFFFFF',
    'on-teal': '#FFFFFF',
    'on-background': '#3D4F6F',
    'on-surface': '#3D4F6F',
    gold: '#F5A623',
    border: '#E8D4CF',       // Warm pink-tinged border
  }
}

export default createVuetify({
  icons: {
    defaultSet: 'mdi',
    aliases,
    sets: {
      mdi,
    },
  },
  theme: {
    defaultTheme: 'light',
    themes: {
      light: actalogTheme,
      dark: actalogDarkTheme,
      brutalist: brutalistTheme,
      ocean: oceanTheme,
      sunset: sunsetTheme,
      sunrise: sunriseTheme,
    },
    variations: {
      colors: ['primary', 'secondary', 'accent'],
      lighten: 2,
      darken: 2,
    },
  },
  defaults: {
    VCard: {
      elevation: 1,
      rounded: 'sm',
      variant: 'elevated',
    },
    VBtn: {
      color: 'primary',
      elevation: 0,
      rounded: 'sm',
      style: 'text-transform: none;',
    },
    VTextField: {
      variant: 'outlined',
      density: 'comfortable',
      rounded: 'sm',
    },
    VSelect: {
      variant: 'outlined',
      density: 'comfortable',
      rounded: 'sm',
    },
    VTextarea: {
      variant: 'outlined',
      density: 'comfortable',
      rounded: 'sm',
    },
  },
  display: {
    mobileBreakpoint: 'sm',
    thresholds: {
      xs: 0,
      sm: 600,
      md: 960,
      lg: 1280,
      xl: 1920,
    },
  },
})
