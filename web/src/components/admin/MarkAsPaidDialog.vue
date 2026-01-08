<template>
  <v-dialog :model-value="modelValue" max-width="500" @update:model-value="$emit('update:modelValue', $event)">
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
            
            density="compact"
            :rules="[v => !!v || 'Payment date is required']"
            prepend-inner-icon="mdi-calendar"
            hint="Date when payment was received"
            persistent-hint
            class="mb-3"
          ></v-text-field>

          <!-- Duration Selection -->
          <v-select
            v-model="durationType"
            label="Payment Duration"
            :items="durationOptions"
            item-title="label"
            item-value="value"
            
            density="compact"
            prepend-inner-icon="mdi-calendar-clock"
            class="mb-3"
          ></v-select>

          <!-- Custom Duration Input -->
          <v-text-field
            v-if="durationType === 'custom'"
            v-model.number="customDays"
            label="Duration (days)"
            type="number"
            
            density="compact"
            :rules="[v => v > 0 || 'Duration must be greater than 0']"
            prepend-inner-icon="mdi-counter"
            hint="Number of days to extend the subscription"
            persistent-hint
            class="mb-3"
          ></v-text-field>

          <!-- Calculate new end date -->
          <v-alert v-if="newEndDate" type="success" variant="tonal" density="compact" class="mb-3">
            <div class="text-body-2">
              <strong>New End Date:</strong> {{ formatDate(newEndDate) }}<br>
              <strong>Duration:</strong> {{ durationDays }} days<br>
              <span class="text-caption">
                Subscription will be extended from payment date
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
const durationType = ref('default')
const customDays = ref(30)

// Duration options
const durationOptions = computed(() => {
  const defaultDays = props.subscription?.subscription_type === 'monthly' ? 30 : 365
  const defaultPeriod = props.subscription?.subscription_type === 'monthly' ? '1 Month' : '1 Year'

  return [
    { label: `Default - ${defaultPeriod} (${defaultDays} days)`, value: 'default' },
    { label: '1 Week (7 days)', value: 7 },
    { label: '2 Weeks (14 days)', value: 14 },
    { label: '1 Month (30 days)', value: 30 },
    { label: '2 Months (60 days)', value: 60 },
    { label: '3 Months (90 days)', value: 90 },
    { label: '6 Months (180 days)', value: 180 },
    { label: '1 Year (365 days)', value: 365 },
    { label: 'Custom...', value: 'custom' }
  ]
})

// Computed - Calculate duration in days
const durationDays = computed(() => {
  if (durationType.value === 'default') {
    return props.subscription?.subscription_type === 'monthly' ? 30 : 365
  } else if (durationType.value === 'custom') {
    return customDays.value || 30
  } else {
    return durationType.value
  }
})

// Computed - Calculate new end date based on payment date and duration
const newEndDate = computed(() => {
  if (!paymentDate.value || !props.subscription) return null

  const payment = new Date(paymentDate.value)
  const newEnd = new Date(payment)
  newEnd.setDate(newEnd.getDate() + durationDays.value)

  return newEnd.toISOString()
})

// Initialize payment date when dialog opens
watch(() => props.modelValue, (newValue) => {
  if (newValue) {
    // Default to today
    const today = new Date()
    paymentDate.value = today.toISOString().split('T')[0]
    durationType.value = 'default'
    customDays.value = 30
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

    const payload = {
      payment_date: paymentDate.value
    }

    // Only send duration_days if not using default
    if (durationType.value !== 'default') {
      payload.duration_days = durationDays.value
    }

    await axios.post(endpoint, payload)

    emit('paid')
    emit('update:modelValue', false)
  } catch (err) {
    console.error('Failed to mark subscription as paid:', err)
    error.value = err.response?.data?.message || err.response?.data?.error || 'Failed to mark subscription as paid'
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
