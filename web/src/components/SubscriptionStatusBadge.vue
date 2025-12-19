<template>
  <v-card v-if="subscriptionStore.subscriptionStatus" class="subscription-status-card">
    <v-card-title class="d-flex align-center">
      <v-icon :color="statusColor" class="mr-2">{{ statusIcon }}</v-icon>
      Subscription Status
    </v-card-title>

    <v-card-text>
      <!-- User Subscription -->
      <div v-if="subscriptionStore.userSubscription" class="mb-4">
        <v-chip
          :color="statusColor"
          label
          class="mb-2"
        >
          {{ subscriptionStore.subscriptionTypeLabel }}
        </v-chip>

        <div class="text-body-2 mt-2">
          <div class="d-flex align-center mb-1">
            <strong class="mr-2">Status:</strong>
            <v-chip
              size="small"
              :color="subscriptionStore.status === 'active' ? 'success' : 'error'"
            >
              {{ subscriptionStore.status }}
            </v-chip>
          </div>

          <!-- Expiration Date -->
          <div v-if="subscriptionStore.expiresAt && !subscriptionStore.isPermanentFree" class="mb-1">
            <strong>Expires:</strong> {{ formatDate(subscriptionStore.expiresAt) }}
            <v-chip
              v-if="daysUntilExpiration !== null && daysUntilExpiration <= 7"
              size="small"
              color="warning"
              class="ml-2"
            >
              {{ daysUntilExpiration }} days left
            </v-chip>
          </div>

          <!-- Permanent Free Badge -->
          <div v-if="subscriptionStore.isPermanentFree" class="mb-1">
            <v-chip color="purple" size="small" prepend-icon="mdi-star">
              Never Expires
            </v-chip>
          </div>

          <!-- Start Date -->
          <div v-if="subscriptionStore.userSubscription.start_date" class="mb-1">
            <strong>Started:</strong> {{ formatDate(subscriptionStore.userSubscription.start_date) }}
          </div>

          <!-- Next Billing Date -->
          <div
            v-if="subscriptionStore.userSubscription.next_billing_date"
            class="mb-1"
          >
            <strong>Next Billing:</strong> {{ formatDate(subscriptionStore.userSubscription.next_billing_date) }}
          </div>
        </div>
      </div>

      <!-- Organization Subscriptions -->
      <div v-if="subscriptionStore.orgSubscriptions.length > 0" class="mt-4">
        <v-divider class="mb-3"></v-divider>
        <div class="text-subtitle-2 mb-2">
          <v-icon small class="mr-1">mdi-office-building</v-icon>
          Organization Subscriptions
        </div>
        <div
          v-for="orgSub in subscriptionStore.orgSubscriptions"
          :key="orgSub.id"
          class="mb-2 pa-2 bg-grey-lighten-4 rounded"
        >
          <div class="text-body-2">
            <strong>{{ orgSub.organization_name || `Organization ${orgSub.organization_id}` }}</strong>
          </div>
          <div class="text-caption">
            <v-chip size="x-small" :color="orgSub.status === 'active' ? 'success' : 'error'" class="mr-1">
              {{ orgSub.status }}
            </v-chip>
            {{ orgSub.subscription_type }}
            <span v-if="orgSub.is_permanent_free"> • Permanent Free</span>
            <span v-else-if="orgSub.end_date"> • Expires {{ formatDate(orgSub.end_date) }}</span>
          </div>
        </div>
      </div>

      <!-- Access Summary -->
      <v-alert
        v-if="!subscriptionStore.hasAccess"
        type="warning"
        variant="tonal"
        density="compact"
        class="mt-4"
      >
        <div class="text-body-2">
          <strong>Subscription Expired</strong><br>
          You can view and export data, but cannot create or edit content.
          Please contact your administrator to renew.
        </div>
      </v-alert>

      <v-alert
        v-else-if="subscriptionStore.source === 'organization'"
        type="info"
        variant="tonal"
        density="compact"
        class="mt-4"
      >
        <div class="text-caption">
          Your access is provided through your organization subscription.
        </div>
      </v-alert>
    </v-card-text>

    <!-- Actions -->
    <v-card-actions v-if="!subscriptionStore.hasAccess">
      <v-btn
        color="primary"
        variant="outlined"
        size="small"
        prepend-icon="mdi-email"
        @click="contactAdmin"
      >
        Contact Admin
      </v-btn>
    </v-card-actions>
  </v-card>

  <!-- Loading State -->
  <v-card v-else-if="subscriptionStore.loading" class="subscription-status-card">
    <v-card-text class="text-center pa-4">
      <v-progress-circular indeterminate color="primary"></v-progress-circular>
      <div class="text-body-2 mt-2">Loading subscription status...</div>
    </v-card-text>
  </v-card>

  <!-- Error State -->
  <v-alert v-else-if="subscriptionStore.error" type="error" variant="tonal" density="compact">
    {{ subscriptionStore.error }}
  </v-alert>
</template>

<script setup>
import { computed } from 'vue'
import { useSubscriptionStore } from '@/stores/subscription'

const subscriptionStore = useSubscriptionStore()

const statusColor = computed(() => {
  if (subscriptionStore.isPermanentFree) return 'purple'
  if (subscriptionStore.hasAccess) return 'success'
  return 'error'
})

const statusIcon = computed(() => {
  if (subscriptionStore.isPermanentFree) return 'mdi-star-circle'
  if (subscriptionStore.hasAccess) return 'mdi-check-circle'
  return 'mdi-alert-circle'
})

const daysUntilExpiration = computed(() => {
  if (!subscriptionStore.expiresAt || subscriptionStore.isPermanentFree) return null

  const expiresDate = new Date(subscriptionStore.expiresAt)
  const now = new Date()
  const diffTime = expiresDate - now
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))

  return diffDays > 0 ? diffDays : 0
})

function formatDate(dateString) {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

function contactAdmin() {
  // TODO: Implement contact admin functionality
  // Could open email client or navigate to contact page
  alert('Please contact your administrator to renew your subscription.')
}
</script>

<style scoped>
.subscription-status-card {
  max-width: 600px;
}

.bg-grey-lighten-4 {
  background-color: #f5f5f5;
}
</style>
