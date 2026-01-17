<template>
  <v-dialog v-model="dialogVisible" max-width="600" scrollable>
    <v-card>
      <!-- Header -->
      <v-card-title
        class="d-flex align-center py-3 px-4"
        :style="{ backgroundColor: headerColor, color: 'white' }"
      >
        <v-icon color="white" class="mr-2">{{ headerIcon }}</v-icon>
        <span class="text-h6 font-weight-bold">{{ entityName }}</span>
        <v-spacer />
        <v-btn icon variant="text" color="white" size="small" @click="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <!-- Loading State -->
      <v-card-text v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="primary" size="48" />
        <p class="text-body-2 mt-3 text-medium-emphasis">Loading details...</p>
      </v-card-text>

      <!-- Error State -->
      <v-card-text v-else-if="error" class="py-4">
        <v-alert type="error" variant="tonal">{{ error }}</v-alert>
      </v-card-text>

      <!-- Content -->
      <v-card-text v-else-if="entityData" class="pa-4">
        <!-- Type Chips -->
        <div class="d-flex flex-wrap gap-2 mb-3">
          <template v-if="type === 'wod'">
            <v-chip size="small" color="secondary" variant="flat">{{ entityData.type }}</v-chip>
            <v-chip size="small" color="primary" variant="flat">{{ entityData.regime }}</v-chip>
            <v-chip size="small" color="success" variant="flat">{{ entityData.score_type }}</v-chip>
          </template>
          <template v-else-if="type === 'movement'">
            <v-chip size="small" :color="getMovementTypeColor(entityData.type)" variant="flat">
              {{ capitalizeFirst(entityData.type) }}
            </v-chip>
          </template>
          <template v-else-if="type === 'template'">
            <v-chip v-if="entityData.movements?.length" size="small" color="info" variant="flat">
              <v-icon start size="x-small">mdi-weight-lifter</v-icon>
              {{ entityData.movements.length }} movement{{ entityData.movements.length > 1 ? 's' : '' }}
            </v-chip>
            <v-chip v-if="entityData.wods?.length" size="small" color="warning" variant="flat">
              <v-icon start size="x-small">mdi-fire</v-icon>
              {{ entityData.wods.length }} WOD{{ entityData.wods.length > 1 ? 's' : '' }}
            </v-chip>
          </template>
          <v-chip v-if="!isStandard" size="x-small" color="primary">Custom</v-chip>
        </div>

        <!-- Description -->
        <div v-if="entityData.description" class="mb-3">
          <p class="text-caption font-weight-bold mb-1 text-medium-emphasis">Description</p>
          <div class="text-body-2">
            <markdown-renderer :content="entityData.description" />
          </div>
        </div>

        <!-- Notes -->
        <div v-if="entityData.notes" class="mb-3">
          <p class="text-caption font-weight-bold mb-1 text-medium-emphasis">Notes</p>
          <div class="text-body-2">
            <markdown-renderer :content="entityData.notes" />
          </div>
        </div>

        <!-- WOD-specific: Source and URL -->
        <template v-if="type === 'wod'">
          <div v-if="entityData.source" class="mb-3">
            <p class="text-caption font-weight-bold mb-1 text-medium-emphasis">Source</p>
            <v-chip size="small" color="info" variant="flat">{{ entityData.source }}</v-chip>
          </div>
          <div v-if="entityData.url" class="mb-3">
            <p class="text-caption font-weight-bold mb-1 text-medium-emphasis">Video</p>
            <v-btn
              :href="entityData.url"
              target="_blank"
              color="primary"
              prepend-icon="mdi-play-circle"
              size="small"
              variant="tonal"
              rounded="lg"
              style="text-transform: none"
            >
              Watch Video
            </v-btn>
          </div>
        </template>

        <!-- Template-specific: Movements and WODs list -->
        <template v-if="type === 'template'">
          <div v-if="entityData.movements?.length" class="mb-3">
            <p class="text-caption font-weight-bold mb-1 text-medium-emphasis">Movements</p>
            <v-list density="compact" bg-color="transparent" class="py-0">
              <v-list-item v-for="m in entityData.movements" :key="m.id" class="px-0 mb-2">
                <template #prepend>
                  <v-icon size="small" color="primary" class="mt-1">mdi-weight-lifter</v-icon>
                </template>
                <div>
                  <v-list-item-title class="text-body-2 font-weight-medium">
                    {{ m.movement?.name || 'Movement' }}
                  </v-list-item-title>
                  <!-- Instructions first -->
                  <div v-if="m.instructions" class="mt-2">
                    <p class="text-caption font-weight-medium mb-1 text-info">
                      <v-icon size="x-small" class="mr-1">mdi-clipboard-text-outline</v-icon>
                      Instructions
                    </p>
                    <div class="text-caption">
                      <markdown-renderer :content="m.instructions" />
                    </div>
                  </div>
                  <!-- Movement specifics -->
                  <div class="d-flex flex-wrap gap-1 mt-1">
                    <v-chip v-if="m.sets" size="x-small" color="primary" variant="tonal">
                      {{ m.sets }} sets
                    </v-chip>
                    <v-chip v-if="m.reps" size="x-small" color="secondary" variant="tonal">
                      {{ m.reps }} reps
                    </v-chip>
                    <v-chip v-if="m.weight" size="x-small" color="warning" variant="tonal">
                      {{ m.weight }} {{ m.weight_unit || 'lbs' }}
                    </v-chip>
                    <v-chip v-if="m.time" size="x-small" color="info" variant="tonal">
                      {{ formatTime(m.time) }}
                    </v-chip>
                    <v-chip v-if="m.distance" size="x-small" color="success" variant="tonal">
                      {{ m.distance }} {{ m.distance_unit || 'm' }}
                    </v-chip>
                  </div>
                  <!-- Notes last -->
                  <p v-if="m.notes" class="text-caption text-medium-emphasis mt-1 mb-0">
                    <strong>Notes:</strong> {{ m.notes }}
                  </p>
                </div>
              </v-list-item>
            </v-list>
          </div>
          <div v-if="entityData.wods?.length" class="mb-3">
            <p class="text-caption font-weight-bold mb-1 text-medium-emphasis">WODs</p>
            <v-list density="compact" bg-color="transparent" class="py-0">
              <v-list-item v-for="w in entityData.wods" :key="w.id" class="px-0 mb-2">
                <template #prepend>
                  <v-icon size="small" color="#ff5722" class="mt-1">mdi-fire</v-icon>
                </template>
                <div>
                  <v-list-item-title class="text-body-2 font-weight-medium">
                    {{ w.wod?.name || 'WOD' }}
                  </v-list-item-title>
                  <!-- Instructions first -->
                  <div v-if="w.instructions" class="mt-2">
                    <p class="text-caption font-weight-medium mb-1 text-info">
                      <v-icon size="x-small" class="mr-1">mdi-clipboard-text-outline</v-icon>
                      Instructions
                    </p>
                    <div class="text-caption">
                      <markdown-renderer :content="w.instructions" />
                    </div>
                  </div>
                  <!-- WOD specifics -->
                  <div class="d-flex flex-wrap gap-1 mt-1">
                    <v-chip v-if="w.wod?.type" size="x-small" color="secondary" variant="tonal">
                      {{ w.wod.type }}
                    </v-chip>
                    <v-chip v-if="w.wod?.regime" size="x-small" color="primary" variant="tonal">
                      {{ w.wod.regime }}
                    </v-chip>
                    <v-chip v-if="w.wod?.score_type" size="x-small" color="success" variant="tonal">
                      {{ w.wod.score_type }}
                    </v-chip>
                  </div>
                  <p v-if="w.wod?.description" class="text-caption text-medium-emphasis mt-1 mb-0" style="white-space: pre-wrap;">
                    {{ w.wod.description }}
                  </p>
                  <!-- Notes last -->
                  <p v-if="w.notes" class="text-caption text-medium-emphasis mt-1 mb-0">
                    <strong>Notes:</strong> {{ w.notes }}
                  </p>
                </div>
              </v-list-item>
            </v-list>
          </div>
        </template>
      </v-card-text>

      <!-- Actions -->
      <v-divider />
      <v-card-actions class="pa-3">
        <v-btn variant="text" @click="close">
          <v-icon start>mdi-arrow-left</v-icon>
          Back
        </v-btn>
        <v-spacer />
        <v-btn
          color="warning"
          variant="flat"
          @click="quickLog"
        >
          <v-icon start>mdi-lightning-bolt</v-icon>
          QuickLog
        </v-btn>
        <v-btn
          v-if="canEdit"
          color="primary"
          variant="flat"
          @click="edit"
        >
          <v-icon start>mdi-pencil</v-icon>
          Edit
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import { useAuthStore } from '@/stores/auth'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  type: {
    type: String,
    required: true,
    validator: (value) => ['wod', 'movement', 'template'].includes(value)
  },
  id: {
    type: [Number, String],
    required: true
  },
  handleQuickLogExternally: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'quicklog', 'edit', 'close'])

const router = useRouter()
const authStore = useAuthStore()

// State
const entityData = ref(null)
const loading = ref(false)
const error = ref('')

// Computed
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const entityName = computed(() => {
  if (!entityData.value) return ''
  return entityData.value.name || 'Unknown'
})

const isStandard = computed(() => {
  if (!entityData.value) return true
  return entityData.value.is_standard !== false
})

const canEdit = computed(() => {
  if (!entityData.value || !authStore.user) return false
  // Admin can edit anything
  if (authStore.user.role === 'admin') return true
  // User can edit their own custom items
  if (!isStandard.value && entityData.value.created_by === authStore.user.id) return true
  return false
})

const headerIcon = computed(() => {
  switch (props.type) {
    case 'wod': return 'mdi-fire'
    case 'movement': return 'mdi-weight-lifter'
    case 'template': return 'mdi-clipboard-text'
    default: return 'mdi-information'
  }
})

const headerColor = computed(() => {
  switch (props.type) {
    case 'wod': return '#ff5722'
    case 'movement': return 'rgb(var(--v-theme-primary))'
    case 'template': return 'rgb(var(--v-theme-primary))'
    default: return 'rgb(var(--v-theme-primary))'
  }
})

// Methods
function getMovementTypeColor(type) {
  const colors = {
    weightlifting: '#2196F3',
    gymnastics: '#9C27B0',
    cardio: '#4CAF50',
    bodyweight: '#FF9800'
  }
  return colors[type] || '#757575'
}

function capitalizeFirst(str) {
  if (!str) return ''
  return str.charAt(0).toUpperCase() + str.slice(1)
}

function formatTime(seconds) {
  if (!seconds) return ''
  if (seconds < 60) return `${seconds}s`
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return secs > 0 ? `${mins}m ${secs}s` : `${mins}m`
}

async function fetchData() {
  if (!props.id) return

  loading.value = true
  error.value = ''
  entityData.value = null

  try {
    let endpoint = ''
    switch (props.type) {
      case 'wod':
        endpoint = `/api/wods/${props.id}`
        break
      case 'movement':
        endpoint = `/api/movements/${props.id}`
        break
      case 'template':
        endpoint = `/api/templates/${props.id}`
        break
    }

    const response = await axios.get(endpoint)
    // Handle different response structures
    entityData.value = response.data.wod || response.data.movement || response.data.template || response.data
  } catch (err) {
    console.error('Failed to fetch entity:', err)
    error.value = 'Failed to load details. Please try again.'
  } finally {
    loading.value = false
  }
}

function close() {
  dialogVisible.value = false
  emit('close')
}

function quickLog() {
  emit('quicklog', { type: props.type, id: props.id, data: entityData.value })

  // If parent handles quicklog externally, just close the dialog
  if (props.handleQuickLogExternally) {
    close()
    return
  }

  // Navigate based on type
  switch (props.type) {
    case 'wod':
      router.push({ path: '/workouts/log', query: { wodId: props.id } })
      break
    case 'movement':
      router.push({ path: '/workouts/log', query: { movementId: props.id } })
      break
    case 'template':
      router.push({ path: '/workouts/log', query: { templateId: props.id } })
      break
  }
  close()
}

function edit() {
  emit('edit', { type: props.type, id: props.id, data: entityData.value })
  // Navigate to edit page
  switch (props.type) {
    case 'wod':
      router.push(`/wods/${props.id}/edit`)
      break
    case 'movement':
      router.push(`/movements/${props.id}/edit`)
      break
    case 'template':
      router.push(`/workouts/templates/${props.id}/edit`)
      break
  }
  close()
}

// Watch for dialog open
watch(() => props.modelValue, (newVal) => {
  if (newVal && props.id) {
    fetchData()
  }
})

// Watch for id change while dialog is open
watch(() => props.id, (newVal) => {
  if (props.modelValue && newVal) {
    fetchData()
  }
})
</script>
