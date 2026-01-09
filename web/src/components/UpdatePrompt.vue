<template>
  <v-snackbar
    v-model="pwaStore.showUpdatePrompt"
    :timeout="-1"
    location="bottom"
    color="#2c3657"
    elevation="8"
    class="update-prompt"
  >
    <div class="d-flex align-center">
      <v-icon start size="large" color="primary">mdi-update</v-icon>
      <div>
        <strong style="color: white">Update Available</strong>
        <div class="text-caption" style="color: rgba(255, 255, 255, 0.8)">
          A new version of ActaLog is ready
        </div>
      </div>
    </div>
    <template #actions>
      <v-btn
        variant="text"
        size="small"
        style="color: rgba(255, 255, 255, 0.7)"
        @click="dismissUpdate"
      >
        Later
      </v-btn>
      <v-btn
        variant="flat"
        size="small"
        color="primary"
        :loading="pwaStore.isUpdating"
        @click="applyUpdate"
      >
        Update Now
      </v-btn>
    </template>
  </v-snackbar>
</template>

<script setup>
import { usePwaStore } from '@/stores/pwa'

const pwaStore = usePwaStore()

async function applyUpdate() {
  await pwaStore.applyUpdate()
  // The page will reload after the service worker activates
}

function dismissUpdate() {
  pwaStore.dismissUpdatePrompt()
}
</script>

<style scoped>
.update-prompt {
  margin-bottom: 56px; /* Account for bottom navigation */
}
</style>
