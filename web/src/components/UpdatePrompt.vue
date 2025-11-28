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
      <v-icon start size="large" color="#00bcd4">mdi-update</v-icon>
      <div>
        <strong style="color: white">Update Available</strong>
        <div class="text-caption" style="color: rgba(255, 255, 255, 0.8)">
          A new version of ActaLog is ready
        </div>
      </div>
    </div>
    <template v-slot:actions>
      <v-btn
        variant="text"
        size="small"
        @click="dismissUpdate"
        style="color: rgba(255, 255, 255, 0.7)"
      >
        Later
      </v-btn>
      <v-btn
        variant="flat"
        size="small"
        color="#00bcd4"
        @click="applyUpdate"
        :loading="isUpdating"
      >
        Update Now
      </v-btn>
    </template>
  </v-snackbar>
</template>

<script setup>
import { ref } from 'vue'
import { usePwaStore } from '@/stores/pwa'

const pwaStore = usePwaStore()
const isUpdating = ref(false)

async function applyUpdate() {
  isUpdating.value = true
  await pwaStore.applyUpdate()
  // The page will reload, so we don't need to reset isUpdating
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
