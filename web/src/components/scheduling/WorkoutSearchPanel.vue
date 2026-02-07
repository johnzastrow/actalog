<template>
  <div class="workout-search-panel">
    <div class="panel-header mb-3 d-flex align-center justify-space-between">
      <div class="section-title">
        <v-icon size="small" class="mr-2">mdi-dumbbell</v-icon>
        Select Workout
      </div>
      <v-btn
        color="primary"
        variant="tonal"
        size="small"
        prepend-icon="mdi-plus"
        rounded="lg"
        @click="$emit('create')"
      >
        New Workout
      </v-btn>
    </div>

    <!-- Search and Sort Controls -->
    <div class="search-controls mb-3">
      <v-text-field
        v-model="search"
        label="Search workouts"
        prepend-inner-icon="mdi-magnify"
        variant="solo-filled"
        density="compact"
        rounded="lg"
        flat
        clearable
        hide-details
        class="mb-2"
      />
      <v-select
        v-model="sortBy"
        :items="sortOptions"
        item-title="label"
        item-value="value"
        label="Sort by"
        variant="solo-filled"
        density="compact"
        rounded="lg"
        flat
        hide-details
      />
    </div>

    <!-- Workout List -->
    <div class="workout-list">
      <v-radio-group v-model="selectedWorkout" hide-details>
        <template v-if="filteredWorkouts.length > 0">
          <div
            v-for="workout in filteredWorkouts"
            :key="workout.id"
            class="workout-item"
            :class="{ 'selected': selectedWorkout === workout.id }"
            @click="selectWorkout(workout.id)"
          >
            <v-radio :value="workout.id" density="compact" hide-details>
              <template #label>
                <div class="workout-info">
                  <div class="workout-name font-weight-medium">{{ workout.name }}</div>
                  <div v-if="workout.intro_warmup" class="workout-detail text-caption text-medium-emphasis text-truncate">
                    {{ workout.intro_warmup }}
                  </div>
                  <div class="workout-meta text-caption text-medium-emphasis mt-1">
                    <span v-if="workout.movements?.length">{{ workout.movements.length }} movements</span>
                    <span v-if="workout.wods?.length" class="ml-2">{{ workout.wods.length }} WODs</span>
                    <span v-if="workout.updated_at" class="ml-2">
                      Updated {{ formatDate(workout.updated_at) }}
                    </span>
                  </div>
                </div>
              </template>
            </v-radio>
          </div>
        </template>
        <div v-else class="no-results text-center py-4 text-medium-emphasis">
          <v-icon size="32" class="mb-2">mdi-magnify-close</v-icon>
          <div>No workouts match your search</div>
        </div>
      </v-radio-group>
    </div>

    <!-- Selected Workout Preview Button -->
    <div v-if="selectedWorkout" class="mt-3 d-flex justify-center">
      <v-btn
        variant="tonal"
        size="small"
        color="primary"
        @click="$emit('preview', selectedWorkout)"
      >
        <v-icon start size="small">mdi-eye</v-icon>
        Preview Selected Workout
      </v-btn>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  workouts: {
    type: Array,
    default: () => []
  },
  selectedId: {
    type: [Number, null],
    default: null
  }
})

const emit = defineEmits(['select', 'preview', 'create'])

const search = ref('')
const sortBy = ref('newest')
const selectedWorkout = ref(props.selectedId)

const sortOptions = [
  { value: 'name_asc', label: 'Name (A-Z)' },
  { value: 'name_desc', label: 'Name (Z-A)' },
  { value: 'newest', label: 'Newest First' },
  { value: 'oldest', label: 'Oldest First' },
  { value: 'updated', label: 'Recently Updated' }
]

const filteredWorkouts = computed(() => {
  let result = [...props.workouts]

  // Text search across multiple fields
  if (search.value) {
    const searchLower = search.value.toLowerCase()
    result = result.filter(w => {
      // Search in name
      if (w.name?.toLowerCase().includes(searchLower)) return true
      // Search in intro_warmup
      if (w.intro_warmup?.toLowerCase().includes(searchLower)) return true
      // Search in notes
      if (w.notes?.toLowerCase().includes(searchLower)) return true
      // Search in movement names
      if (w.movements?.some(m => m.movement?.name?.toLowerCase().includes(searchLower))) return true
      // Search in WOD names
      if (w.wods?.some(wod => wod.wod?.name?.toLowerCase().includes(searchLower))) return true
      return false
    })
  }

  // Sort
  result.sort((a, b) => {
    switch (sortBy.value) {
      case 'name_asc':
        return (a.name || '').localeCompare(b.name || '')
      case 'name_desc':
        return (b.name || '').localeCompare(a.name || '')
      case 'newest':
        return new Date(b.created_at || 0) - new Date(a.created_at || 0)
      case 'oldest':
        return new Date(a.created_at || 0) - new Date(b.created_at || 0)
      case 'updated':
        return new Date(b.updated_at || 0) - new Date(a.updated_at || 0)
      default:
        return 0
    }
  })

  return result
})

function selectWorkout(id) {
  selectedWorkout.value = id
  emit('select', id)
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

// Watch for external changes to selectedId
watch(() => props.selectedId, (newVal) => {
  selectedWorkout.value = newVal
})
</script>

<style scoped>
.workout-search-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.section-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #546e7a;
  display: flex;
  align-items: center;
}

.workout-list {
  flex: 1;
  overflow-y: auto;
  max-height: 400px;
}

.workout-item {
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 8px;
  background-color: #f8f9fa;
  cursor: pointer;
  transition: all 0.2s;
}

.workout-item:hover {
  background-color: #e9ecef;
}

.workout-item.selected {
  background-color: #e3f2fd;
  border: 1px solid #2196f3;
}

.workout-info {
  margin-left: 8px;
}

.workout-name {
  font-size: 0.95rem;
}

.workout-detail {
  max-width: 250px;
}

.no-results {
  color: #78909c;
}
</style>
