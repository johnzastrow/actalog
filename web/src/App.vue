<template>
  <v-app>
    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<script setup>
import { onMounted } from 'vue'
import { useTheme } from 'vuetify'

const theme = useTheme()

// Check for user's preferred theme
onMounted(() => {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme) {
    theme.global.name.value = savedTheme
  }
})
</script>

<style>
/* Global styles based on requirements */
* {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
    'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

body {
  margin: 0;
  padding: 0;
}

/* Typography from requirements */
h1, h2, h3, h4, h5, h6 {
  font-weight: 500;
}

p, span, div {
  font-weight: 400;
  font-size: 16px;
}

/* Spacing from requirements - 20px outer padding, 16px gutter */
.v-container {
  padding: 20px;
}

.v-row {
  margin: -8px;
}

.v-col {
  padding: 8px;
}
</style>
