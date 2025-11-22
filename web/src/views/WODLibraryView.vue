<template>
  <div style="background-color: #f5f7fa; min-height: 100vh">
    <v-container class="pa-2" style="margin-bottom: 70px; overflow-y: auto; max-height: calc(100vh - 70px)">
      <!-- Search and Filters Card -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <v-text-field
          v-model="searchQuery"
          label="Search WODs"
          placeholder="Search by name..."
          variant="outlined"
          density="compact"
          rounded="lg"
          clearable
          hide-details
          class="mb-2"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-magnify</v-icon>
          </template>
        </v-text-field>

        <v-chip-group
          v-model="selectedType"
          selected-class="text-white"
          color="#00bcd4"
          class="mt-2"
          mandatory
        >
          <v-chip value="all" size="small">All</v-chip>
          <v-chip value="Girl" size="small">Girl</v-chip>
          <v-chip value="Hero" size="small">Hero</v-chip>
          <v-chip value="Benchmark" size="small">Benchmark</v-chip>
          <v-chip value="Games" size="small">Games</v-chip>
        </v-chip-group>
      </v-card>

      <!-- Error Alert -->
      <v-alert
        v-if="wodsStore.error"
        type="error"
        closable
        @click:close="wodsStore.error = ''"
        class="mb-3"
      >
        {{ wodsStore.error }}
      </v-alert>

      <!-- Loading State -->
      <div v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="#00bcd4" size="48" />
        <p class="text-body-2 mt-3" style="color: #666">Loading WODs...</p>
      </div>

      <!-- Empty State -->
      <v-card
        v-else-if="filteredWODs.length === 0"
        elevation="0"
        rounded="lg"
        class="pa-6 text-center"
        style="background: white"
      >
        <v-icon size="64" color="#ccc">mdi-fire</v-icon>
        <p class="text-h6 mt-3" style="color: #666">No WODs found</p>
        <p class="text-body-2" style="color: #999">
          {{ searchQuery ? 'Try adjusting your search' : 'Create your first WOD to get started' }}
        </p>
        <v-btn
          v-if="!searchQuery"
          color="#00bcd4"
          size="large"
          rounded="lg"
          class="mt-4"
          prepend-icon="mdi-plus"
          @click="createNewWOD"
          style="text-transform: none"
        >
          Create WOD
        </v-btn>
      </v-card>

      <!-- WODs List -->
      <div v-else>
        <v-card
          v-for="wod in filteredWODs"
          :key="wod.id"
          elevation="0"
          rounded="lg"
          class="pa-3 mb-2"
          style="background: white; border: 1px solid #e0e0e0"
          :ripple="true"
          @click="handleWODClick(wod)"
        >
          <div class="d-flex align-center">
            <v-icon color="#ff5722" class="mr-3" size="32">mdi-fire</v-icon>
            <div style="flex: 1">
              <div class="d-flex align-center mb-1">
                <span class="text-body-1 font-weight-bold" style="color: #1a1a1a">
                  {{ wod.name }}
                </span>
                <v-chip
                  v-if="!wod.is_standard"
                  size="x-small"
                  color="teal"
                  class="ml-2"
                >
                  Custom
                </v-chip>
              </div>
              <div class="d-flex flex-wrap gap-1 mb-1">
                <v-chip size="x-small" color="#9c27b0" variant="outlined">
                  {{ wod.type }}
                </v-chip>
                <v-chip size="x-small" color="#00bcd4" variant="outlined">
                  {{ wod.regime }}
                </v-chip>
                <v-chip size="x-small" color="#4caf50" variant="outlined">
                  {{ wod.score_type }}
                </v-chip>
              </div>
              <p class="text-caption mb-0" style="color: #666; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;">
                {{ wod.description }}
              </p>
            </div>
            <v-icon v-if="selectionMode" color="#00bcd4">mdi-chevron-right</v-icon>
            <v-btn
              v-else-if="!wod.is_standard"
              icon="mdi-pencil"
              size="small"
              variant="text"
              color="#00bcd4"
              @click.stop="editWOD(wod.id)"
            />
          </div>
        </v-card>
      </div>
    </v-container>

    <!-- FAB for Create (when not in selection mode) -->
    <v-btn
      v-if="!selectionMode && !loading"
      icon="mdi-plus"
      size="x-large"
      color="teal"
      elevation="8"
      style="position: fixed; bottom: 80px; right: 16px; z-index: 5"
      @click="createNewWOD"
    />

    <!-- Bottom Navigation (only when not in selection mode) -->
    <v-bottom-navigation
      v-if="!selectionMode"
      v-model="activeNav"
      grow
      style="position: fixed; bottom: 0; width: 100%; z-index: 5; background: white"
      elevation="8"
    >
      <v-btn value="dashboard" @click="$router.push('/dashboard')">
        <v-icon>mdi-view-dashboard</v-icon>
        <span>Dashboard</span>
      </v-btn>

      <v-btn value="workouts" @click="$router.push('/workouts')">
        <v-icon>mdi-clipboard-text</v-icon>
        <span>Workouts</span>
      </v-btn>

      <v-btn value="performance" @click="$router.push('/performance')">
        <v-icon>mdi-chart-line</v-icon>
        <span>Performance</span>
      </v-btn>

      <v-btn value="profile" @click="$router.push('/profile')">
        <v-icon>mdi-account</v-icon>
        <span>Profile</span>
      </v-btn>
    </v-bottom-navigation>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useWodsStore } from '@/stores/wods'

const router = useRouter()
const route = useRoute()
const wodsStore = useWodsStore()

// State
const searchQuery = ref('')
const selectedType = ref('all')
const activeNav = ref('wods')

// Check if in selection mode
const selectionMode = computed(() => route.query.select === 'true')

// Computed properties from store
const loading = computed(() => wodsStore.loading)
const error = computed(() => wodsStore.error)

// Filtered WODs
const filteredWODs = computed(() => {
  let filtered = wodsStore.wods

  // Filter by type
  if (selectedType.value !== 'all') {
    filtered = wodsStore.filterByType(selectedType.value)
  }

  // Filter by search query
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(w =>
      w.name.toLowerCase().includes(query) ||
      (w.description && w.description.toLowerCase().includes(query))
    )
  }

  return filtered
})

// Load WODs
async function fetchWODs() {
  await wodsStore.fetchWods()
}

// Handle WOD click
function handleWODClick(wod) {
  if (selectionMode.value) {
    // In selection mode, return with selected WOD
    router.push({
      path: route.query.returnPath || '/workouts/templates/create',
      query: { selectedWOD: wod.id }
    })
  } else {
    // In browse mode, navigate to WOD detail
    router.push(`/wods/${wod.id}`)
  }
}

// Create new WOD
function createNewWOD() {
  if (selectionMode.value) {
    router.push({
      path: '/wods/create',
      query: { returnPath: route.query.returnPath, select: 'true' }
    })
  } else {
    router.push('/wods/create')
  }
}

// Edit WOD
function editWOD(id) {
  router.push(`/wods/${id}/edit`)
}

// Handle back navigation
function handleBack() {
  if (selectionMode.value && route.query.returnPath) {
    router.push(route.query.returnPath)
  } else {
    router.back()
  }
}

// Initialize
onMounted(() => {
  fetchWODs()
})
</script>
