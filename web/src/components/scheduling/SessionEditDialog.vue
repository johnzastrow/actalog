<template>
  <v-dialog v-model="dialogOpen" max-width="500">
    <v-card>
      <v-card-title class="d-flex align-center">
        <v-icon class="mr-2">mdi-calendar-edit</v-icon>
        Edit Session
      </v-card-title>

      <v-card-text v-if="session">
        <!-- Read-only session info -->
        <v-alert type="info" variant="tonal" density="compact" class="mb-4">
          <div class="text-body-2 font-weight-medium">{{ session.name }}</div>
          <div class="text-caption">{{ formatDateTime(session.start_time) }}</div>
        </v-alert>

        <!-- Editable fields -->
        <v-select
          v-model="form.location_id"
          :items="locations"
          item-title="name"
          item-value="id"
          label="Location"
          variant="outlined"
          density="comfortable"
          clearable
          class="mb-3"
        />

        <v-select
          v-model="form.workout_id"
          :items="workouts"
          item-title="name"
          item-value="id"
          label="Workout (override)"
          variant="outlined"
          density="comfortable"
          clearable
          class="mb-3"
        />

        <v-text-field
          v-model.number="form.capacity"
          label="Capacity"
          type="number"
          variant="outlined"
          density="comfortable"
          class="mb-3"
        />

        <v-textarea
          v-model="form.notes"
          label="Notes"
          variant="outlined"
          density="comfortable"
          rows="2"
        />
      </v-card-text>

      <v-card-actions>
        <v-btn
          v-if="session && session.status === 'scheduled'"
          color="error"
          variant="text"
          @click="onCancel"
        >
          Cancel Session
        </v-btn>
        <v-spacer />
        <v-btn @click="close">Close</v-btn>
        <v-btn color="primary" :loading="saving" @click="save">Save</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, watch, computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  session: {
    type: Object,
    default: null
  },
  locations: {
    type: Array,
    default: () => []
  },
  workouts: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue', 'save', 'cancel'])

const saving = ref(false)

const form = ref({
  location_id: null,
  workout_id: null,
  capacity: 20,
  notes: ''
})

const dialogOpen = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

// Update form when session changes
watch(() => props.session, (newSession) => {
  if (newSession) {
    form.value = {
      location_id: newSession.location_id,
      workout_id: newSession.workout_id,
      capacity: newSession.capacity,
      notes: newSession.description || ''
    }
  }
}, { immediate: true })

function formatDateTime(dateTimeStr) {
  if (!dateTimeStr) return ''
  return new Date(dateTimeStr).toLocaleString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  })
}

function close() {
  dialogOpen.value = false
}

async function save() {
  saving.value = true
  try {
    emit('save', {
      ...form.value,
      id: props.session?.id
    })
    close()
  } finally {
    saving.value = false
  }
}

function onCancel() {
  emit('cancel', props.session)
}
</script>
