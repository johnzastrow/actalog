<template>
  <v-dialog :model-value="modelValue" @update:model-value="$emit('update:modelValue', $event)" max-width="500">
    <v-card v-if="subscription">
      <v-card-title>
        <v-icon class="mr-2" color="success">mdi-cash-check</v-icon>
        Mark Subscription as Paid
      </v-card-title>

      <v-card-text>
        <v-alert type="info" variant="tonal" density="compact" class="mb-4">
          <div class="text-body-2">
            <strong>Current Status:</strong><br>
            Type: {{ subscription.subscription_type }}<br>
            Status: {{ subscription.status }}<br>
            <span v-if="subscription.end_date">
              Current End Date: {{ formatDate(subscription.end_date) }}
            </span>
            <span v-else>No end date set</span>
          </div>
        </v-alert>

        <v-form ref="formRef">
          <v-text-field
            v-model="paymentDate"
            label="Payment Date"
            type="date"
            variant="outlined"
            density="compact"
            :rules="[v => !!v || 'Payment date is required']"
            prepend-inner-icon="mdi-calendar"
            hint="Date when payment was received"
            persistent-hint
            class="mb-3"
          ></v-text-field>

          <!-- Calculate new end date -->
          <v-alert v-if="newEndDate" type="success" variant="tonal" density="compact" class="mb-3">
            <div class="text-body-2">
              <strong>New End Date:</strong> {{ formatDate(newEndDate) }}<br>
              <span class="text-caption">
                Subscription will be extended based on payment date
              </span>
            </div>
          </v-alert>
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
          Cancel
        </v-btn>
        <v-btn
          color="success"
          :loading="loading"
          @click="markAsPaid"
        >
          Confirm Payment
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
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

const emit = defineEmits(['update:modelValue', 'paid'])

const formRef = ref(null)
const loading = ref(false)
const error = ref('')
const paymentDate = ref('')

// Computed - Calculate new end date based on payment date and subscription type
const newEndDate = computed(() => {
  if (!paymentDate.value || !props.subscription) return null

  const payment = new Date(paymentDate.value)
  const newEnd = new Date(payment)

  if (props.subscription.subscription_type === 'monthly') {
    newEnd.setDate(newEnd.getDate() + 30)
  } else if (props.subscription.subscription_type === 'annual') {
    newEnd.setFullYear(newEnd.getFullYear() + 1)
  } else {
    // Free tier - no extension
    return null
  }

  return newEnd.toISOString()
})

// Initialize payment date when dialog opens
watch(() => props.modelValue, (newValue) => {
  if (newValue) {
    // Default to today
    const today = new Date()
    paymentDate.value = today.toISOString().split('T')[0]
    error.value = ''
  }
})

async function markAsPaid() {
  error.value = ''

  // Validate form
  const { valid } = await formRef.value.validate()
  if (!valid) return

  loading.value = true
  try {
    const endpoint = props.type === 'user'
      ? `/api/admin/subscriptions/user/${props.subscription.id}/mark-paid`
      : `/api/admin/subscriptions/organization/${props.subscription.id}/mark-paid`

    await axios.post(endpoint, {
      payment_date: paymentDate.value
    })

    emit('paid')
    emit('update:modelValue', false)
  } catch (err) {
    console.error('Failed to mark subscription as paid:', err)
    error.value = err.response?.data?.error || 'Failed to mark subscription as paid'
  } finally {
    loading.value = false
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
</script>
