<template>
  <v-dialog :model-value="modelValue" max-width="700" @update:model-value="$emit('update:modelValue', $event)">
    <v-card v-if="subscription">
      <v-card-title>
        <v-icon class="mr-2">mdi-information</v-icon>
        Subscription Details
      </v-card-title>

      <v-card-text>
        <!-- Basic Information -->
        <v-card  class="mb-3">
          <v-card-subtitle class="font-weight-bold">Basic Information</v-card-subtitle>
          <v-card-text>
            <v-list density="compact">
              <v-list-item>
                <v-list-item-title>{{ type === 'user' ? 'User' : 'Organization' }}</v-list-item-title>
                <v-list-item-subtitle>
                  {{ getSubscriptionName() }}
                </v-list-item-subtitle>
              </v-list-item>

              <v-list-item>
                <v-list-item-title>Subscription Type</v-list-item-title>
                <v-list-item-subtitle>
                  <v-chip
                    :color="getTypeColor(subscription.subscription_type)"
                    size="small"
                    label
                  >
                    {{ subscription.is_permanent_free ? 'Permanent Free' : subscription.subscription_type }}
                  </v-chip>
                </v-list-item-subtitle>
              </v-list-item>

              <v-list-item>
                <v-list-item-title>Status</v-list-item-title>
                <v-list-item-subtitle>
                  <v-chip
                    :color="getStatusColor(subscription.status)"
                    size="small"
                  >
                    {{ subscription.status }}
                  </v-chip>
                </v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>

        <!-- Dates -->
        <v-card  class="mb-3">
          <v-card-subtitle class="font-weight-bold">Subscription Dates</v-card-subtitle>
          <v-card-text>
            <v-list density="compact">
              <v-list-item>
                <v-list-item-title>Start Date</v-list-item-title>
                <v-list-item-subtitle>
                  {{ formatDate(subscription.start_date) }}
                </v-list-item-subtitle>
              </v-list-item>

              <v-list-item>
                <v-list-item-title>End Date</v-list-item-title>
                <v-list-item-subtitle>
                  {{ subscription.is_permanent_free ? 'Never (Permanent Free)' : formatDate(subscription.end_date) }}
                </v-list-item-subtitle>
              </v-list-item>

              <v-list-item v-if="subscription.last_payment_date">
                <v-list-item-title>Last Payment</v-list-item-title>
                <v-list-item-subtitle>
                  {{ formatDate(subscription.last_payment_date) }}
                </v-list-item-subtitle>
              </v-list-item>

              <v-list-item v-if="subscription.next_billing_date">
                <v-list-item-title>Next Billing</v-list-item-title>
                <v-list-item-subtitle>
                  {{ formatDate(subscription.next_billing_date) }}
                </v-list-item-subtitle>
              </v-list-item>

              <v-list-item v-if="subscription.cancelled_at">
                <v-list-item-title>Cancelled At</v-list-item-title>
                <v-list-item-subtitle>
                  {{ formatDate(subscription.cancelled_at) }}
                </v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>

        <!-- Cancellation Info (if cancelled) -->
        <v-card v-if="subscription.cancelled_at"  color="warning" class="mb-3">
          <v-card-subtitle class="font-weight-bold">Cancellation Information</v-card-subtitle>
          <v-card-text>
            <v-list density="compact">
              <v-list-item v-if="subscription.cancelled_reason">
                <v-list-item-title>Reason</v-list-item-title>
                <v-list-item-subtitle>
                  {{ subscription.cancelled_reason }}
                </v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>

        <!-- Notes -->
        <v-card v-if="subscription.notes"  class="mb-3">
          <v-card-subtitle class="font-weight-bold">Admin Notes</v-card-subtitle>
          <v-card-text>
            {{ subscription.notes }}
          </v-card-text>
        </v-card>

        <!-- Metadata -->
        <v-card >
          <v-card-subtitle class="font-weight-bold">Metadata</v-card-subtitle>
          <v-card-text>
            <v-list density="compact">
              <v-list-item>
                <v-list-item-title>Created At</v-list-item-title>
                <v-list-item-subtitle>
                  {{ formatDateTime(subscription.created_at) }}
                </v-list-item-subtitle>
              </v-list-item>

              <v-list-item>
                <v-list-item-title>Last Updated</v-list-item-title>
                <v-list-item-subtitle>
                  {{ formatDateTime(subscription.updated_at) }}
                </v-list-item-subtitle>
              </v-list-item>

              <v-list-item v-if="subscription.created_by_user_id">
                <v-list-item-title>Created By</v-list-item-title>
                <v-list-item-subtitle>
                  Admin User ID: {{ subscription.created_by_user_id }}
                </v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          variant="text"
          @click="$emit('update:modelValue', false)"
        >
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
const props = defineProps({
  modelValue: Boolean,
  subscription: Object,
  type: {
    type: String,
    required: true,
    validator: (value) => ['user', 'organization'].includes(value)
  }
})

defineEmits(['update:modelValue', 'refresh'])

function getSubscriptionName() {
  if (!props.subscription) return ''

  if (props.type === 'user') {
    return props.subscription.user_email || `User ${props.subscription.user_id}`
  } else {
    return props.subscription.organization_name || `Organization ${props.subscription.organization_id}`
  }
}

function formatDate(dateString) {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

function formatDateTime(dateString) {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function getTypeColor(type) {
  const colors = {
    free: 'grey',
    monthly: 'blue',
    annual: 'green'
  }
  return colors[type] || 'grey'
}

function getStatusColor(status) {
  const colors = {
    active: 'success',
    expired: 'error',
    cancelled: 'warning'
  }
  return colors[status] || 'grey'
}
</script>
