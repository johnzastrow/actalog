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
import { ref, watch, computed } from 'vue'

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

const selectedDays = ref(props.modelValue)

watch(() => props.modelValue, (newVal) => {
  selectedDays.value = newVal
})

watch(selectedDays, (newVal) => {
  emit('update:modelValue', newVal)
})
</script>

<style scoped>
.day-of-week-selector {
  display: inline-block;
}

.v-btn-toggle .v-btn {
  text-transform: none;
  font-weight: 500;
}
</style>
