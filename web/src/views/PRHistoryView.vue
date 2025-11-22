<template>
  <v-container class="px-1 pb-1 pt-0" style="margin-top: 5px; margin-bottom: 70px">
    <!-- Loading State -->
    <v-progress-circular v-if="loading" indeterminate color="#00bcd4" class="mx-auto d-block mt-8" />

    <div v-else>
      <!-- Search Box -->
      <v-card elevation="0" rounded class="pa-1 mb-2" style="background: white">
        <v-text-field
          v-model="searchQuery"
          placeholder="Search for a personal record..."
          variant="plain"
          density="compact"
          hide-details
          rounded
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-magnify</v-icon>
          </template>
        </v-text-field>
      </v-card>

      <!-- PR Movements -->
      <div v-for="(group, movementName) in groupedMovementPRs" :key="movementName">
        <v-card elevation="0" rounded class="pa-2 mb-2" style="background: white">
          <!-- Movement Header -->
          <div class="d-flex align-center mb-1">
            <v-icon color="teal" class="mr-2">mdi-trophy</v-icon>
            <h3 class="text-body-1 font-weight-bold" style="color: #2c3e50">{{ movementName }}</h3>
          </div>

          <!-- PR Entries for this Movement -->
          <div v-for="pr in group" :key="pr.id" class="mb-1 pa-1" style="border-left: 3px solid #ffc107; padding-left: 8px">
            <div class="d-flex justify-space-between align-center">
              <div>
                <span class="text-caption" style="color: #666">{{ formatDate(pr.workout_date) }}</span>
              </div>
              <div class="text-body-2 font-weight-medium" style="color: #2c3e50">
                {{ formatPerformance(pr) }}
              </div>
            </div>
          </div>
        </v-card>
      </div>

      <!-- PR WODs -->
      <div v-for="(group, wodName) in groupedWODPRs" :key="wodName">
        <v-card elevation="0" rounded class="pa-2 mb-2" style="background: white">
          <!-- WOD Header -->
          <div class="d-flex align-center mb-1">
            <v-icon color="teal" class="mr-2">mdi-trophy</v-icon>
            <h3 class="text-body-1 font-weight-bold" style="color: #2c3e50">{{ wodName }}</h3>
          </div>

          <!-- PR Entries for this WOD -->
          <div v-for="pr in group" :key="pr.id" class="mb-1 pa-1" style="border-left: 3px solid #ffc107; padding-left: 8px">
            <div class="d-flex justify-space-between align-center">
              <div>
                <span class="text-caption" style="color: #666">{{ formatDate(pr.workout_date) }}</span>
              </div>
              <div class="text-body-2 font-weight-medium" style="color: #2c3e50">
                {{ formatWODPerformance(pr) }}
              </div>
            </div>
          </div>
        </v-card>
      </div>

      <!-- Empty State -->
      <v-card
        v-if="!loading && filteredMovementPRs.length === 0 && filteredWODPRs.length === 0"
        elevation="0"
        rounded
        class="pa-2 text-center mt-4"
        style="background: white"
      >
        <v-icon size="48" color="#ccc">mdi-trophy-outline</v-icon>
        <p class="text-body-1 mt-1 mb-0" style="color: #2c3e50">No Personal Records Yet</p>
        <p class="text-body-2 mb-0" style="color: #666">
          Keep logging workouts to set new PRs!
        </p>
      </v-card>
    </div>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from '@/utils/axios'

const loading = ref(true)
const searchQuery = ref('')
const movementPRs = ref([])
const wodPRs = ref([])

// Fetch personal records
const fetchPersonalRecords = async () => {
  try {
    loading.value = true
    const response = await axios.get('/api/workouts/personal-records')
    movementPRs.value = response.data.pr_movements || []
    wodPRs.value = response.data.pr_wods || []
  } catch (error) {
    console.error('Error fetching personal records:', error)
  } finally {
    loading.value = false
  }
}

// Filter PRs based on search query
const filteredMovementPRs = computed(() => {
  if (!searchQuery.value) return movementPRs.value
  const query = searchQuery.value.toLowerCase()
  return movementPRs.value.filter(pr =>
    pr.movement_name?.toLowerCase().includes(query)
  )
})

const filteredWODPRs = computed(() => {
  if (!searchQuery.value) return wodPRs.value
  const query = searchQuery.value.toLowerCase()
  return wodPRs.value.filter(pr =>
    pr.wod_name?.toLowerCase().includes(query)
  )
})

// Group PRs by movement name
const groupedMovementPRs = computed(() => {
  const groups = {}
  filteredMovementPRs.value.forEach(pr => {
    const name = pr.movement_name || 'Unknown'
    if (!groups[name]) {
      groups[name] = []
    }
    groups[name].push(pr)
  })
  return groups
})

// Group PRs by WOD name
const groupedWODPRs = computed(() => {
  const groups = {}
  filteredWODPRs.value.forEach(pr => {
    const name = pr.wod_name || 'Unknown'
    if (!groups[name]) {
      groups[name] = []
    }
    groups[name].push(pr)
  })
  return groups
})

// Format date (e.g., "Fri, Oct 10, 2024")
const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' })
}

// Format movement performance (e.g., "2 x 1 @ 175 lbs")
const formatPerformance = (pr) => {
  const parts = []
  if (pr.sets) parts.push(`${pr.sets} x`)
  if (pr.reps) parts.push(`${pr.reps}`)
  if (pr.weight) parts.push(`@ ${pr.weight} lbs`)
  if (pr.time_seconds) {
    const minutes = Math.floor(pr.time_seconds / 60)
    const seconds = pr.time_seconds % 60
    parts.push(`${minutes}:${seconds.toString().padStart(2, '0')}`)
  }
  return parts.join(' ') || 'Completed'
}

// Format WOD performance (e.g., "12:34", "10+15", "225.5")
const formatWODPerformance = (pr) => {
  if (pr.time_seconds) {
    const minutes = Math.floor(pr.time_seconds / 60)
    const seconds = pr.time_seconds % 60
    return `${minutes}:${seconds.toString().padStart(2, '0')}`
  }
  if (pr.rounds !== null && pr.rounds !== undefined) {
    if (pr.reps) {
      return `${pr.rounds} rounds + ${pr.reps} reps`
    }
    return `${pr.rounds} rounds`
  }
  if (pr.weight) {
    return `${pr.weight} lbs`
  }
  return pr.score_value || 'Completed'
}

onMounted(() => {
  fetchPersonalRecords()
})
</script>
