<template>
  <v-dialog :model-value="modelValue" max-width="500" @update:model-value="$emit('update:modelValue', $event)">
    <v-card v-if="subscription">
      <v-card-title>
        <v-icon class="mr-2" color="error">mdi-cancel</v-icon>
        Cancel Subscription
      </v-card-title>

      <v-card-text>
        <v-alert type="warning" variant="tonal" density="compact" class="mb-4">
          <div class="text-body-2">
            Are you sure you want to cancel this subscription?<br>
            <strong>{{ getSubscriptionName() }}</strong><br>
            Type: {{ subscription.subscription_type }}<br>
            Status: {{ subscription.status }}
          </div>
        </v-alert>

        <v-form ref="formRef">
          <v-textarea
            v-model="cancellationReason"
            label="Cancellation Reason"
            variant="outlined"
            density="compact"
            rows="3"
            :rules="[v => !!v || 'Cancellation reason is required']"
            hint="Please provide a reason for cancelling this subscription"
            persistent-hint
            prepend-inner-icon="mdi-text"
          ></v-textarea>
        </v-form>

        <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mt-3">
          {{ error }}
        </v-alert>
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          variant="text"
          @click="$emit('update:modelValue', false)"
        >
          Keep Subscription
        </v-btn>
        <v-btn
          color="error"
          :loading="loading"
          @click="cancelSubscription"
        >
          Cancel Subscription
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import axios from '@/utils/axios'

const props = defineProps({
  modelValue: Boolean,
  subscription: Object,
  type: {
    type: String,
    required: true,
    validator: (value) => ['user', 'organization'].includes(value)
  }
})

const emit = defineEmits(['update:modelValue', 'cancelled'])

const formRef = ref(null)
const loading = ref(false)
const error = ref('')
const cancellationReason = ref('')

// Reset form when dialog opens
watch(() => props.modelValue, (newValue) => {
  if (newValue) {
    cancellationReason.value = ''
    error.value = ''
  }
})

function getSubscriptionName() {
  if (!props.subscription) return ''

  if (props.type === 'user') {
    return props.subscription.user_email || `User ${props.subscription.user_id}`
  } else {
    return props.subscription.organization_name || `Organization ${props.subscription.organization_id}`
  }
}

async function cancelSubscription() {
  error.value = ''

  // Validate form
  const { valid } = await formRef.value.validate()
  if (!valid) return

  loading.value = true
  try {
    const endpoint = props.type === 'user'
      ? `/api/admin/subscriptions/user/${props.subscription.id}/cancel`
      : `/api/admin/subscriptions/organization/${props.subscription.id}/cancel`

    await axios.post(endpoint, {
      reason: cancellationReason.value
    })

    emit('cancelled')
    emit('update:modelValue', false)
  } catch (err) {
    console.error('Failed to cancel subscription:', err)
    error.value = err.response?.data?.error || 'Failed to cancel subscription'
  } finally {
    loading.value = false
  }
}
</script>
