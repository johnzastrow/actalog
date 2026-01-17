<template>
  <v-card elevation="0" rounded class="pa-3" bg-color="surface">
    <div class="d-flex align-center">
      <v-icon :color="color" size="32" class="mr-3">{{ icon }}</v-icon>
      <div>
        <div class="text-h5 font-weight-bold">
          {{ formattedValue }}
        </div>
        <div class="text-caption text-medium-emphasis">{{ label }}</div>
      </div>
    </div>
    <div v-if="subtitle" class="mt-2 text-caption text-medium-emphasis">
      {{ subtitle }}
    </div>
  </v-card>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  icon: {
    type: String,
    required: true
  },
  color: {
    type: String,
    default: 'primary'
  },
  value: {
    type: [Number, String],
    required: true
  },
  label: {
    type: String,
    required: true
  },
  subtitle: {
    type: String,
    default: null
  },
  format: {
    type: String,
    default: 'number',
    validator: (value) => ['number', 'percent', 'decimal'].includes(value)
  }
})

const formattedValue = computed(() => {
  if (props.value === null || props.value === undefined) {
    return '-'
  }

  switch (props.format) {
    case 'percent':
      return `${Number(props.value).toFixed(1)}%`
    case 'decimal':
      return Number(props.value).toFixed(2)
    default:
      return typeof props.value === 'number'
        ? props.value.toLocaleString()
        : props.value
  }
})
</script>
