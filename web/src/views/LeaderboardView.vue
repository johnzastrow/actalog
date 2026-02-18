<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable class="mb-3" @click:close="error = null">
        {{ error }}
      </v-alert>

      <!-- Tab Selection -->
      <v-card elevation="2" class="pa-3 mb-3" bg-color="surface">
        <v-tabs v-model="activeTab" density="compact" color="primary" grow>
          <v-tab value="movements">
            <v-icon start size="small">mdi-dumbbell</v-icon>
            Movements
          </v-tab>
          <v-tab value="wods">
            <v-icon start size="small">mdi-fire</v-icon>
            WODs
          </v-tab>
        </v-tabs>

        <!-- Search -->
        <v-autocomplete
          v-model="selectedItem"
          :items="searchResults"
          item-title="name"
          item-value="id"
          :loading="loadingSearch"
          :placeholder="activeTab === 'movements' ? 'Search movements...' : 'Search WODs...'"
          density="comfortable"
          clearable
          auto-select-first
          hide-details
          return-object
          class="mt-3"
          @update:search="handleSearch"
          @update:model-value="handleSelection"
        >
          <template #prepend-inner>
            <v-icon color="primary" size="small">mdi-magnify</v-icon>
          </template>
          <template #item="{ props, item }">
            <v-list-item v-bind="props" density="compact">
              <template #prepend>
                <v-icon color="primary" size="small">
                  {{ activeTab === 'movements' ? 'mdi-dumbbell' : 'mdi-fire' }}
                </v-icon>
              </template>
              <v-list-item-title class="text-body-2">{{ item.raw.name }}</v-list-item-title>
            </v-list-item>
          </template>
        </v-autocomplete>

        <!-- Period Toggle -->
        <v-btn-toggle
          v-model="period"
          mandatory
          density="compact"
          color="primary"
          class="mt-3"
          divided
          rounded="lg"
        >
          <v-btn value="all_time" size="small">All Time</v-btn>
          <v-btn value="this_year" size="small">This Year</v-btn>
          <v-btn value="this_month" size="small">This Month</v-btn>
        </v-btn-toggle>
      </v-card>

      <!-- Empty State -->
      <v-card v-if="!selectedItem" elevation="0" rounded="lg" class="pa-6 mb-3 text-center" bg-color="surface">
        <v-icon size="64" color="surface-variant">mdi-podium</v-icon>
        <h2 class="text-h6 font-weight-bold mt-3">Community Leaderboard</h2>
        <p class="text-body-2 mt-2 text-medium-emphasis">
          Search for a {{ activeTab === 'movements' ? 'movement' : 'WOD' }} to see how you rank against your gym mates
        </p>
        <p class="text-caption mt-1 text-medium-emphasis">
          Only members who opt in via Settings appear on leaderboards
        </p>
      </v-card>

      <!-- Loading -->
      <v-card v-else-if="loading" elevation="0" rounded="lg" class="pa-6 text-center" bg-color="surface">
        <v-progress-circular indeterminate color="primary" />
        <p class="text-body-2 mt-3 text-medium-emphasis">Loading leaderboard...</p>
      </v-card>

      <!-- Results -->
      <v-card v-else-if="leaderboard" elevation="0" rounded="lg" class="pa-3" bg-color="surface">
        <div class="d-flex align-center mb-3">
          <v-icon color="primary" class="mr-2">
            {{ activeTab === 'movements' ? 'mdi-dumbbell' : 'mdi-fire' }}
          </v-icon>
          <div>
            <h2 class="text-body-1 font-weight-bold">{{ leaderboard.item_name }}</h2>
            <p class="text-caption text-medium-emphasis">
              {{ leaderboard.total }} participant{{ leaderboard.total !== 1 ? 's' : '' }}
            </p>
          </div>
        </div>

        <!-- No entries -->
        <div v-if="!leaderboard.entries || leaderboard.entries.length === 0" class="text-center pa-4">
          <v-icon size="48" color="surface-variant">mdi-account-group-outline</v-icon>
          <p class="text-body-2 mt-2 text-medium-emphasis">
            No leaderboard entries yet. Members need to opt in and log workouts.
          </p>
        </div>

        <!-- Rankings -->
        <v-list v-else bg-color="transparent" density="compact">
          <v-list-item
            v-for="entry in leaderboard.entries"
            :key="entry.user_id"
            :class="{ 'bg-primary-lighten-5': entry.is_current_user }"
            rounded="lg"
            class="mb-1"
          >
            <template #prepend>
              <div class="rank-badge mr-3" :class="getRankClass(entry.rank)">
                {{ entry.rank }}
              </div>
            </template>
            <v-list-item-title class="text-body-2" :class="{ 'font-weight-bold': entry.is_current_user }">
              {{ entry.user_name }}
              <v-chip v-if="entry.is_current_user" size="x-small" color="primary" class="ml-1">You</v-chip>
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption">
              {{ formatDate(entry.workout_date) }}
            </v-list-item-subtitle>
            <template #append>
              <span class="text-body-2 font-weight-bold">{{ entry.score_display }}</span>
            </template>
          </v-list-item>
        </v-list>
      </v-card>
    </v-container>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import axios from '@/utils/axios'

const error = ref(null)
const activeTab = ref('movements')
const selectedItem = ref(null)
const searchResults = ref([])
const loadingSearch = ref(false)
const loading = ref(false)
const period = ref('all_time')
const leaderboard = ref(null)

let searchTimeout = null

const handleSearch = (query) => {
  if (!query || query.length < 2) {
    searchResults.value = []
    return
  }

  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(async () => {
    loadingSearch.value = true
    try {
      const endpoint = activeTab.value === 'movements'
        ? `/api/movements/search?q=${encodeURIComponent(query)}&limit=20`
        : `/api/wods/search?q=${encodeURIComponent(query)}&limit=20`
      const response = await axios.get(endpoint)
      searchResults.value = response.data || []
    } catch {
      searchResults.value = []
    } finally {
      loadingSearch.value = false
    }
  }, 300)
}

const handleSelection = (item) => {
  if (item) {
    fetchLeaderboard()
  } else {
    leaderboard.value = null
  }
}

const fetchLeaderboard = async () => {
  if (!selectedItem.value) return

  loading.value = true
  error.value = null

  try {
    const type = activeTab.value === 'movements' ? 'movements' : 'wods'
    const id = selectedItem.value.id
    const response = await axios.get(`/api/leaderboards/${type}/${id}`, {
      params: { period: period.value, limit: 50, offset: 0 }
    })
    leaderboard.value = response.data
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to load leaderboard'
    leaderboard.value = null
  } finally {
    loading.value = false
  }
}

const getRankClass = (rank) => {
  if (rank === 1) return 'rank-gold'
  if (rank === 2) return 'rank-silver'
  if (rank === 3) return 'rank-bronze'
  return ''
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

// Re-fetch when period changes
watch(period, () => {
  if (selectedItem.value) {
    fetchLeaderboard()
  }
})

// Clear selection when tab changes
watch(activeTab, () => {
  selectedItem.value = null
  searchResults.value = []
  leaderboard.value = null
})
</script>

<style scoped>
.rank-badge {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: bold;
  background-color: rgba(0, 0, 0, 0.08);
}

.rank-gold {
  background-color: #ffd700;
  color: #333;
}

.rank-silver {
  background-color: #c0c0c0;
  color: #333;
}

.rank-bronze {
  background-color: #cd7f32;
  color: #fff;
}

.bg-primary-lighten-5 {
  background-color: rgba(0, 188, 212, 0.08) !important;
}
</style>
