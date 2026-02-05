<template>
  <div class="day-of-week-selector">
    <div class="text-caption text-medium-emphasis mb-2" v-if="label">{{ label }}</div>
    <v-btn-toggle
      v-model="selectedDays"
      :multiple="multiple"
      :mandatory="!multiple"
      color="primary"
      variant="outlined"
      divided
    >
      <v-btn
        v-for="(day, index) in dayLabels"
        :key="index"
        :value="index"
        size="small"
        min-width="40"
      >
        {{ day }}
      </v-btn>
    </v-btn-toggle>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'

const props = defineProps({
  modelValue: {
    type: [Number, Array],
    default: 0
  },
  multiple: {
    type: Boolean,
    default: false
  },
  label: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue'])

// Day labels starting from Sunday (0)
const dayLabels = ['S', 'M', 'T', 'W', 'T', 'F', 'S']

// Initialize based on multiple mode
function normalizeValue(val, isMultiple) {
  if (isMultiple) {
    // Multiple mode: ensure it's an array
    if (Array.isArray(val)) return [...val]
    if (typeof val === 'number') return [val]
    return [1] // Default to Monday
  } else {
    // Single mode: ensure it's a number
    if (typeof val === 'number') return val
    if (Array.isArray(val) && val.length > 0) return val[0]
    return 1 // Default to Monday
  }
}

// Use a ref with watchers to handle the v-btn-toggle properly
const selectedDays = ref(normalizeValue(props.modelValue, props.multiple))

// Flag to prevent circular updates
let isUpdatingFromProp = false

// Helper to check if arrays are equal
function arraysEqual(a, b) {
  if (!Array.isArray(a) || !Array.isArray(b)) return a === b
  if (a.length !== b.length) return false
  const sortedA = [...a].sort()
  const sortedB = [...b].sort()
  return sortedA.every((val, i) => val === sortedB[i])
}

// Watch for external changes to modelValue
watch(() => props.modelValue, (newVal) => {
  isUpdatingFromProp = true
  const normalized = normalizeValue(newVal, props.multiple)
  // Only update if values actually changed to prevent unnecessary watch triggers
  if (!arraysEqual(selectedDays.value, normalized)) {
    selectedDays.value = normalized
  }
  // Use nextTick to ensure the flag stays true until after any queued watches execute
  nextTick(() => {
    isUpdatingFromProp = false
  })
}, { deep: true })

// Watch for internal changes and emit to parent
// IMPORTANT: Spread to create new array so parent's watchEffect detects the change
watch(selectedDays, (newVal) => {
  if (!isUpdatingFromProp) {
    emit('update:modelValue', Array.isArray(newVal) ? [...newVal] : newVal)
  }
}, { deep: true })
</script>

<style scoped>
.day-of-week-selector {
  display: inline-block;
}

.v-btn-toggle .v-btn {
  text-transform: none;
  font-weight: 500;
}

/* Selected day buttons - dark blue with light text */
.v-btn-toggle .v-btn.v-btn--active {
  background-color: #1a365d !important;
  color: #ffffff !important;
}

.v-btn-toggle .v-btn.v-btn--active:hover {
  background-color: #2c5282 !important;
}
</style>
