import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'

// ActaLog custom theme based on design requirements
const actalogTheme = {
  dark: false,
  colors: {
    primary: '#2c3657',      // Primary color from requirements
    secondary: '#597a6a',    // Secondary color
    accent: '#5a4e68',       // Accent color
    error: '#DF3F40',        // Error/Alert color
    warning: '#FFA726',
    info: '#29B6F6',
    success: '#66BB6A',
    background: '#FFFFFF',
    surface: '#FFFFFF',
    'on-primary': '#FFFFFF',
    'on-secondary': '#FFFFFF',
    'on-accent': '#FFFFFF',
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
    primary: '#2c3657',
    secondary: '#597a6a',
    accent: '#5a4e68',
    error: '#DF3F40',
    warning: '#FFA726',
    info: '#29B6F6',
    success: '#66BB6A',
    background: '#1A1A1A',
    surface: '#2A2A2A',
    'on-primary': '#FFFFFF',
    'on-secondary': '#FFFFFF',
    'on-accent': '#FFFFFF',
    'on-background': '#FFFFFF',
    'on-surface': '#FFFFFF',
    gold: '#FFD700',
    border: '#3A3A3A',
  }
}

export default createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'light',
    themes: {
      light: actalogTheme,
      dark: actalogDarkTheme,
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
      rounded: 'lg',
      variant: 'elevated',
    },
    VBtn: {
      elevation: 0,
      rounded: 'lg',
      style: 'text-transform: none;',
    },
    VTextField: {
      variant: 'outlined',
      density: 'comfortable',
      rounded: 'lg',
    },
    VSelect: {
      variant: 'outlined',
      density: 'comfortable',
      rounded: 'lg',
    },
    VTextarea: {
      variant: 'outlined',
      density: 'comfortable',
      rounded: 'lg',
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
