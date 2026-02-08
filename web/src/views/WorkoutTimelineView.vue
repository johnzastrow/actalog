<template>
  <div class="mobile-view-wrapper">
    <!-- Header -->
    <div style="background: #2c3e50; padding: 16px; position: fixed; top: 56px; left: 0; right: 0; z-index: 5">
      <h2 style="color: white; margin: 0; font-size: 1.25rem">Workout Timeline</h2>
    </div>

    <!-- Content (with padding for fixed header) -->
    <div style="padding: 72px 16px 16px 16px; overflow-y: auto">
      <!-- Filters -->
      <v-card elevation="0" rounded="lg" class="mb-3 pa-3" bg-color="surface">
        <div class="d-flex align-center gap-2">
          <v-select
            v-model="filterType"
            :items="typeOptions"
            label="Type"
            density="compact"
            
            clearable
            hide-details
            class="flex-grow-1"
          ></v-select>
          <v-select
            v-model="filterMovement"
            :items="movementOptions"
            label="Movement"
            density="compact"
            
            clearable
            hide-details
            class="flex-grow-1"
          ></v-select>
        </div>
      </v-card>

      <!-- Loading State -->
      <div v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="primary"></v-progress-circular>
      </div>

      <!-- Empty State -->
      <div v-else-if="filteredWorkouts.length === 0" class="text-center py-8">
        <v-icon size="64" color="surface-variant">mdi-timeline-clock-outline</v-icon>
        <p class="mt-3 text-medium-emphasis">No workouts found</p>
      </div>

      <!-- Timeline -->
      <div v-else>
        <div v-for="(group, date) in groupedWorkouts" :key="date" class="mb-4">
          <!-- Date Header -->
          <div class="d-flex align-center mb-2">
            <div class="text-subtitle-2 font-weight-bold text-medium-emphasis">
              {{ formatDate(date) }}
            </div>
            <v-divider class="ml-3"></v-divider>
          </div>

          <!-- Workout Cards -->
          <div class="timeline-items">
            <v-card
              v-for="workout in group"
              :key="workout.id"
              elevation="0"
              rounded="lg"
              class="mb-3 timeline-card"
              style="background-color: rgb(var(--v-theme-surface)); border-left: 4px solid rgb(var(--v-theme-primary))"
              @click="viewWorkout(workout.id)"
            >
              <v-card-text>
                <div class="d-flex justify-space-between align-center mb-2">
                  <div class="text-subtitle-1 font-weight-medium" >
                    {{ workout.workout_name || workout.template_name || 'Workout' }}
                  </div>
                  <v-chip v-if="workout.workout_type" size="small" color="primary" text-color="white">
                    {{ workout.workout_type }}
                  </v-chip>
                </div>

                <!-- Total Time -->
                <div v-if="workout.total_time" class="d-flex align-center mb-2">
                  <v-icon size="small" color="medium-emphasis" class="mr-1">mdi-clock-outline</v-icon>
                  <span class="text-body-2 text-medium-emphasis">
                    {{ formatTime(workout.total_time) }}
                  </span>
                </div>

                <!-- Movements -->
                <div v-if="workout.movements && workout.movements.length > 0" class="mb-2">
                  <div class="text-caption font-weight-medium mb-1 text-medium-emphasis">
                    Movements ({{ workout.movements.length }})
                  </div>
                  <div class="d-flex flex-wrap gap-1">
                    <v-chip
                      v-for="movement in workout.movements.slice(0, 3)"
                      :key="movement.id"
                      size="small"
                      :color="movement.is_pr ? '#ffc107' : '#e0e0e0'"
                      :text-color="movement.is_pr ? 'white' : '#666'"
                    >
                      <v-icon v-if="movement.is_pr" start size="small">mdi-trophy</v-icon>
                      {{ movement.movement_name }}
                    </v-chip>
                    <v-chip
                      v-if="workout.movements.length > 3"
                      size="small"
                      color="#e0e0e0"
                      text-color="medium-emphasis"
                    >
                      +{{ workout.movements.length - 3 }} more
                    </v-chip>
                  </div>
                </div>

                <!-- WODs -->
                <div v-if="workout.wods && workout.wods.length > 0">
                  <div class="text-caption font-weight-medium mb-1 text-medium-emphasis">
                    WODs
                  </div>
                  <div class="d-flex flex-wrap gap-1">
                    <v-chip
                      v-for="wod in workout.wods"
                      :key="wod.id"
                      size="small"
                      :color="wod.is_pr ? '#ffc107' : '#e0e0e0'"
                      :text-color="wod.is_pr ? 'white' : '#666'"
                    >
                      <v-icon v-if="wod.is_pr" start size="small">mdi-trophy</v-icon>
                      {{ wod.wod_name }}
                    </v-chip>
                  </div>
                </div>

                <!-- Notes Preview -->
                <div v-if="workout.notes" class="mt-2">
                  <div class="text-caption text-disabled">
                    {{ truncateNotes(workout.notes) }}
                  </div>
                </div>
              </v-card-text>
            </v-card>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'

const router = useRouter()
const loading = ref(true)
const workouts = ref([])
const filterType = ref(null)
const filterMovement = ref(null)
const movements = ref([])

const typeOptions = computed(() => {
  const types = new Set(workouts.value.map(w => w.workout_type).filter(Boolean))
  return Array.from(types)
})

const movementOptions = computed(() => {
  return movements.value.map(m => ({
    title: m.name,
    value: m.id
  }))
})

const filteredWorkouts = computed(() => {
  let filtered = workouts.value

  if (filterType.value) {
    filtered = filtered.filter(w => w.workout_type === filterType.value)
  }

  if (filterMovement.value) {
    filtered = filtered.filter(w =>
      w.movements?.some(m => m.movement_id === filterMovement.value)
    )
  }

  return filtered.sort((a, b) => new Date(b.workout_date) - new Date(a.workout_date))
})

const groupedWorkouts = computed(() => {
  const groups = {}
  filteredWorkouts.value.forEach(workout => {
    const date = new Date(workout.workout_date).toISOString().split('T')[0]
    if (!groups[date]) {
      groups[date] = []
    }
    groups[date].push(workout)
  })
  return groups
})

function formatDate(dateStr) {
  const date = new Date(dateStr)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  const dateOnly = date.toISOString().split('T')[0]
  const todayStr = today.toISOString().split('T')[0]
  const yesterdayStr = yesterday.toISOString().split('T')[0]

  if (dateOnly === todayStr) return 'Today'
  if (dateOnly === yesterdayStr) return 'Yesterday'

  return date.toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  })
}

function formatTime(seconds) {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60

  if (hours > 0) {
    return `${hours}h ${minutes}m`
  } else if (minutes > 0) {
    return `${minutes}m ${secs}s`
  } else {
    return `${secs}s`
  }
}

function truncateNotes(notes) {
  if (notes.length > 100) {
    return notes.substring(0, 100) + '...'
  }
  return notes
}

function viewWorkout(id) {
  router.push(`/workouts/${id}`)
}

async function fetchData() {
  loading.value = true
  try {
    const [workoutsRes, movementsRes] = await Promise.all([
      axios.get('/api/workouts'),
      axios.get('/api/movements')
    ])
    workouts.value = Array.isArray(workoutsRes.data) ? workoutsRes.data : (workoutsRes.data.workouts || [])
    movements.value = Array.isArray(movementsRes.data) ? movementsRes.data : (movementsRes.data.movements || [])
  } catch (error) {
    console.error('Failed to fetch data:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.timeline-items {
  padding-left: 8px;
}

.timeline-card {
  cursor: pointer;
  transition: transform 0.2s;
}

.timeline-card:hover {
  transform: translateX(4px);
}

.gap-1 {
  gap: 4px;
}

.gap-2 {
  gap: 8px;
}
</style>
