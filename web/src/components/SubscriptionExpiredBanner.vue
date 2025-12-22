<template>
  <v-banner
    v-if="showBanner"
    color="warning"
    icon="mdi-alert"
    sticky
    lines="two"
    class="subscription-expired-banner"
  >
    <v-banner-text>
      <strong>Subscription Expired</strong><br>
      Your subscription has expired. You can view and export data, but cannot create or edit content.
    </v-banner-text>

    <template #actions>
      <v-btn
        variant="text"
        @click="dismiss"
      >
        Dismiss
      </v-btn>
      <v-btn
        color="primary"
        variant="elevated"
        @click="contactAdmin"
      >
        Renew Subscription
      </v-btn>
    </template>
  </v-banner>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useSubscriptionStore } from '@/stores/subscription'

const subscriptionStore = useSubscriptionStore()

// Track if user has dismissed the banner (session-based)
const dismissed = ref(false)

const showBanner = computed(() => {
  return subscriptionStore.isExpired && !dismissed.value
})

// Re-show banner if subscription changes to expired
watch(() => subscriptionStore.isExpired, (isExpired) => {
  if (isExpired) {
    dismissed.value = false
  }
})

function dismiss() {
  dismissed.value = true
}

function contactAdmin() {
  // TODO: Implement contact admin functionality
  // Could navigate to settings, open email, or show contact form
  alert('Please contact your administrator to renew your subscription.')
}
</script>

<style scoped>
.subscription-expired-banner {
  position: sticky;
  top: 56px; /* Below fixed header */
  z-index: 999;
}
</style>
