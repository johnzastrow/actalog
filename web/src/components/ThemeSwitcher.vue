<template>
  <v-menu offset-y>
    <template #activator="{ props }">
      <v-btn
        icon
        variant="text"
        color="white"
        size="small"
        v-bind="props"
      >
        <v-icon>{{ currentThemeDetails.icon }}</v-icon>
        <v-tooltip activator="parent" location="bottom">
          Theme: {{ currentThemeDetails.name }}
        </v-tooltip>
      </v-btn>
    </template>

    <v-card min-width="320" max-width="400">
      <v-card-title class="d-flex align-center">
        <v-icon start>mdi-palette</v-icon>
        Choose Theme
      </v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pa-0">
        <!-- System Preference Toggle -->
        <v-list-item>
          <template #prepend>
            <v-icon>mdi-brightness-auto</v-icon>
          </template>
          <v-list-item-title>Use System Theme</v-list-item-title>
          <template #append>
            <v-switch
              v-model="themeStore.useSystemPreference"
              color="primary"
              hide-details
              density="compact"
              @update:model-value="themeStore.setUseSystemPreference"
            ></v-switch>
          </template>
        </v-list-item>

        <v-divider></v-divider>

        <!-- Theme List -->
        <v-list density="compact" :disabled="themeStore.useSystemPreference">
          <v-list-item
            v-for="theme in themeStore.availableThemes"
            :key="theme.id"
            :active="themeStore.currentTheme === theme.id"
            :class="{ 'v-list-item--active': themeStore.currentTheme === theme.id }"
            @click="selectTheme(theme.id)"
          >
            <template #prepend>
              <v-icon :color="themeStore.currentTheme === theme.id ? 'primary' : ''">
                {{ theme.icon }}
              </v-icon>
            </template>

            <v-list-item-title>{{ theme.name }}</v-list-item-title>
            <v-list-item-subtitle class="text-caption">
              {{ theme.description }}
            </v-list-item-subtitle>

            <template v-if="themeStore.currentTheme === theme.id" #append>
              <v-icon color="primary" size="small">mdi-check-circle</v-icon>
            </template>
          </v-list-item>
        </v-list>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          size="small"
          variant="text"
          @click="closeMenu"
        >
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useThemeStore } from '@/stores/theme'

const themeStore = useThemeStore()
const menuOpen = ref(false)

const currentThemeDetails = computed(() => {
  return themeStore.getCurrentThemeDetails()
})

const selectTheme = (themeId) => {
  themeStore.setTheme(themeId)
}

const closeMenu = () => {
  menuOpen.value = false
}
</script>

<style scoped>
.v-list-item--active {
  background-color: rgb(var(--v-theme-primary), 0.1);
}
</style>
