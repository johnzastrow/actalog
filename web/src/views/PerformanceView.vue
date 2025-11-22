<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">

    <v-container class="px-3 pb-1 pt-0" style="margin-bottom: 70px">
      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable @click:close="error = null" class="mb-3">
        {{ error }}
      </v-alert>

      <!-- Unified Search (Movements + WODs) -->
      <v-card elevation="2" rounded="lg" class="pa-3 mb-3" style="background: white">
        <v-autocomplete
          v-model="selectedItem"
          :items="searchResults"
          item-title="name"
          item-value="id"
          :loading="loadingSearch"
          placeholder="Search for a WOD or Movement..."
          variant="outlined"
          density="comfortable"
          rounded="lg"
          clearable
          auto-select-first
          hide-details
          return-object
          @update:search="handleSearch"
          @update:model-value="handleSelection"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-magnify</v-icon>
          </template>
          <template #item="{ props, item }">
            <v-list-item v-bind="props" density="compact">
              <template #prepend>
                <v-icon
                  :color="item.raw.type === 'movement' ? '#00bcd4' : '#ffc107'"
                  size="small"
                >
                  {{ item.raw.type === 'movement' ? 'mdi-dumbbell' : 'mdi-fire' }}
                </v-icon>
              </template>
              <v-list-item-title class="text-body-2">
                {{ item.raw.name }}
              </v-list-item-title>
              <v-list-item-subtitle class="text-caption">
                {{ item.raw.type === 'movement' ? formatMovementType(item.raw.data?.type) : item.raw.data?.type || 'WOD' }}
              </v-list-item-subtitle>
            </v-list-item>
          </template>
        </v-autocomplete>
      </v-card>

      <!-- Empty State - Prompt to Search -->
      <v-card v-if="!selectedItem" elevation="0" rounded="lg" class="pa-6 mb-3 text-center" style="background: white">
        <v-icon size="64" color="#ccc">mdi-chart-line-variant</v-icon>
        <h2 class="text-h6 font-weight-bold mt-3" style="color: #1a1a1a">Track Your Performance</h2>
        <p class="text-body-2 mt-2" style="color: #666">
          Search for a movement or WOD above to view your progress, PRs, and performance history
        </p>
      </v-card>

      <!-- Performance Content (Movement or WOD) -->
      <div v-if="selectedItem">
        <!-- Quick Log Button -->
        <v-btn
          block
          size="large"
          color="teal"
          rounded="lg"
          elevation="2"
          class="mb-3 font-weight-bold"
          style="text-transform: none"
          @click="quickLog"
        >
          <v-icon start>mdi-lightning-bolt</v-icon>
          Quick Log {{ selectedItem.name }}
        </v-btn>

        <!-- MOVEMENT-SPECIFIC CONTENT -->
        <template v-if="selectedItem.type === 'movement'">
          <!-- Heaviest Lifts (Top 3 Maxes) -->
          <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
            <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">
              <v-icon color="teal" size="small" class="mr-1">mdi-trophy</v-icon>
              Heaviest Lifts
            </h2>

            <div v-if="loadingPerformance" class="text-center py-4">
              <v-progress-circular indeterminate color="#00bcd4" size="32" />
            </div>

            <div v-else-if="heaviestLifts.length === 0" class="text-center py-4">
              <p class="text-caption" style="color: #999">No performance data yet</p>
            </div>

            <div v-else>
              <v-row>
                <v-col v-for="(lift, index) in heaviestLifts" :key="index" cols="4">
                  <div class="text-center">
                    <v-chip
                      :color="index === 0 ? '#ffc107' : index === 1 ? '#9e9e9e' : '#cd7f32'"
                      size="small"
                      class="mb-2"
                      label
                    >
                      #{{ index + 1 }}
                    </v-chip>
                    <div class="font-weight-bold text-h6" style="color: #1a1a1a">
                      {{ lift.weight }}
                    </div>
                    <div class="text-caption" style="color: #666">
                      lbs
                    </div>
                    <div v-if="lift.reps" class="text-caption" style="color: #999">
                      {{ lift.reps }} reps
                    </div>
                  </div>
                </v-col>
              </v-row>
            </div>
          </v-card>

          <!-- Rep Scheme Dropdown Filter -->
          <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
            <v-select
              v-model="selectedRepScheme"
              :items="repSchemes"
              label="Filter by Rep Scheme"
              variant="outlined"
              density="comfortable"
              rounded="lg"
              hide-details
              @update:model-value="filterChart"
            >
              <template #prepend-inner>
                <v-icon color="#00bcd4" size="small">mdi-filter</v-icon>
              </template>
            </v-select>
          </v-card>

          <!-- Performance Chart -->
          <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
            <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Performance Chart</h2>

            <div v-if="loadingPerformance" class="text-center py-4">
              <v-progress-circular indeterminate color="#00bcd4" size="32" />
            </div>

            <div v-else-if="filteredChartData.length === 0" class="text-center py-4">
              <p class="text-caption" style="color: #999">No data for selected filter</p>
            </div>

            <div v-else style="height: 250px; position: relative; width: 100%">
              <canvas ref="chartCanvas" style="width: 100%; height: 100%"></canvas>
            </div>
          </v-card>
        </template>

        <!-- WOD-SPECIFIC CONTENT -->
        <template v-if="selectedItem.type === 'wod'">
          <!-- Best WOD Performances (Top 3) -->
          <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
            <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">
              <v-icon color="teal" size="small" class="mr-1">mdi-trophy</v-icon>
              Best Performances
            </h2>

            <div v-if="loadingPerformance" class="text-center py-4">
              <v-progress-circular indeterminate color="#00bcd4" size="32" />
            </div>

            <div v-else-if="bestWODPerformances.length === 0" class="text-center py-4">
              <p class="text-caption" style="color: #999">No performance data yet</p>
            </div>

            <div v-else>
              <v-row>
                <v-col v-for="(perf, index) in bestWODPerformances" :key="index" cols="4">
                  <div class="text-center">
                    <v-chip
                      :color="index === 0 ? '#ffc107' : index === 1 ? '#9e9e9e' : '#cd7f32'"
                      size="small"
                      class="mb-2"
                      label
                    >
                      #{{ index + 1 }}
                    </v-chip>
                    <div class="font-weight-bold text-body-2" style="color: #1a1a1a">
                      {{ formatWODScore(perf) }}
                    </div>
                    <div class="text-caption" style="color: #666">
                      {{ formatDate(perf.created_at) }}
                    </div>
                  </div>
                </v-col>
              </v-row>
            </div>
          </v-card>

          <!-- WOD Performance Chart -->
          <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
            <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Performance Chart</h2>

            <div v-if="loadingPerformance" class="text-center py-4">
              <v-progress-circular indeterminate color="#00bcd4" size="32" />
            </div>

            <div v-else-if="wodPerformanceData.length === 0" class="text-center py-4">
              <p class="text-caption" style="color: #999">No performance data yet</p>
            </div>

            <div v-else style="height: 250px; position: relative; width: 100%">
              <canvas ref="wodChartCanvas" style="width: 100%; height: 100%"></canvas>
            </div>
          </v-card>
        </template>

        <!-- Performance History (Grouped by Year) -->
        <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
          <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Performance History</h2>

          <div v-if="loadingPerformance" class="text-center py-4">
            <v-progress-circular indeterminate color="#00bcd4" size="32" />
          </div>

          <div v-else-if="Object.keys(groupedHistory).length === 0" class="text-center py-4">
            <v-icon size="48" color="#ccc">mdi-history</v-icon>
            <p class="text-body-2 mt-2" style="color: #666">No history yet</p>
            <p class="text-caption" style="color: #999">
              Start logging workouts with {{ selectedItem.name }} to track your progress
            </p>
          </div>

          <!-- History Grouped by Year -->
          <div v-else>
            <div v-for="(entries, year) in groupedHistory" :key="year" class="mb-4">
              <v-chip size="small" color="#00bcd4" label class="mb-2">
                {{ year }}
              </v-chip>

              <v-card
                v-for="(entry, index) in entries"
                :key="index"
                elevation="0"
                rounded="lg"
                class="mb-2 pa-2"
                style="background: #f5f7fa"
              >
                <div class="d-flex align-center">
                  <!-- PR Trophy Icon -->
                  <v-icon v-if="entry.is_pr" color="teal" size="small" class="mr-2">mdi-trophy</v-icon>

                  <div class="flex-grow-1">
                    <!-- Movement Performance Display -->
                    <div v-if="selectedItem.type === 'movement'" class="font-weight-bold text-body-2" style="color: #1a1a1a">
                      <span v-if="entry.sets && entry.reps && entry.weight">
                        {{ entry.sets }} × {{ entry.reps }} @ {{ entry.weight }} lbs
                      </span>
                      <span v-else-if="entry.weight">
                        {{ entry.weight }} lbs
                        <span v-if="entry.reps"> × {{ entry.reps }}</span>
                      </span>
                      <span v-else-if="entry.time">
                        {{ formatTime(entry.time) }}
                      </span>
                    </div>

                    <!-- WOD Performance Display -->
                    <div v-if="selectedItem.type === 'wod'" class="font-weight-bold text-body-2" style="color: #1a1a1a">
                      {{ formatWODScore(entry) }}
                    </div>

                    <div class="text-caption" style="color: #666">
                      {{ formatDate(entry.created_at) }}
                      <span v-if="entry.notes"> • {{ entry.notes }}</span>
                    </div>
                  </div>

                  <!-- PR Badge -->
                  <v-chip
                    v-if="entry.is_pr"
                    size="x-small"
                    color="teal"
                    class="ml-2"
                  >
                    PR
                  </v-chip>
                </div>
              </v-card>
            </div>
          </div>
        </v-card>
      </div>
    </v-container>

    <!-- Bottom Navigation -->
    <v-bottom-navigation
      v-model="activeTab"
      grow
      style="position: fixed; bottom: 0; background: white"
      elevation="8"
    >
      <v-btn value="dashboard" to="/dashboard">
        <v-icon>mdi-view-dashboard</v-icon>
        <span style="font-size: 10px">Dashboard</span>
      </v-btn>
      <v-btn value="performance" to="/performance">
        <v-icon>mdi-chart-line</v-icon>
        <span style="font-size: 10px">Performance</span>
      </v-btn>
      <v-btn value="log" to="/dashboard?open=quick-log" style="position: relative; bottom: 20px">
        <v-avatar color="teal" size="56" style="box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2)">
          <v-icon color="white" size="32">mdi-plus</v-icon>
        </v-avatar>
      </v-btn>
      <v-btn value="workouts" to="/workouts">
        <v-icon>mdi-dumbbell</v-icon>
        <span style="font-size: 10px">Templates</span>
      </v-btn>
      <v-btn value="profile" to="/profile">
        <v-icon>mdi-account</v-icon>
        <span style="font-size: 10px">Profile</span>
      </v-btn>
    </v-bottom-navigation>
  </v-container>
</template>

<script setup>
import { ref, computed, watch, nextTick, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'
import { Chart, registerables } from 'chart.js'

Chart.register(...registerables)

const router = useRouter()
const activeTab = ref('performance')

// State
const selectedItem = ref(null) // { type: 'movement' | 'wod', id, name, data }
const searchQuery = ref('')
const searchResults = ref([])
const loadingSearch = ref(false)

// Performance data
const performanceData = ref([]) // Raw performance data from API
const loadingPerformance = ref(false)

// Movement-specific
const selectedRepScheme = ref('All Reps')
const repSchemes = ref(['All Reps'])

// Chart instances
const chartCanvas = ref(null)
const wodChartCanvas = ref(null)
let chartInstance = null
let wodChartInstance = null

const error = ref(null)

// Computed: Heaviest Lifts (Top 3 Maxes - Movement only)
const heaviestLifts = computed(() => {
  if (!selectedItem.value || selectedItem.value.type !== 'movement') return []

  // Get all unique weight records, sorted by weight descending
  const weightRecords = performanceData.value
    .filter(p => p.weight)
    .sort((a, b) => b.weight - a.weight)

  // Get top 3 unique weights
  const seen = new Set()
  const top3 = []

  for (const record of weightRecords) {
    const key = `${record.weight}-${record.reps || 0}`
    if (!seen.has(key) && top3.length < 3) {
      seen.add(key)
      top3.push(record)
    }
  }

  return top3
})

// Computed: Best WOD Performances (Top 3)
const bestWODPerformances = computed(() => {
  if (!selectedItem.value || selectedItem.value.type !== 'wod' || performanceData.value.length === 0) return []

  // Get WOD's defined score type
  const wodScoreType = performanceData.value[0]?.wod_score_type || ''

  // Filter data based on WOD's score type
  let validData = []
  if (wodScoreType.includes('Time')) {
    validData = performanceData.value.filter(p => p.time_seconds)
  } else if (wodScoreType.includes('Rounds')) {
    validData = performanceData.value.filter(p => p.rounds !== null || p.reps !== null)
  } else if (wodScoreType.includes('Weight')) {
    validData = performanceData.value.filter(p => p.weight)
  } else {
    validData = performanceData.value
  }

  // Sort by best performance based on score type
  const sorted = [...validData].sort((a, b) => {
    if (wodScoreType.includes('Time')) {
      // Time-based (lower is better)
      return a.time_seconds - b.time_seconds
    } else if (wodScoreType.includes('Rounds')) {
      // Rounds+Reps (higher is better)
      const aTotal = (a.rounds || 0) * 1000 + (a.reps || 0)
      const bTotal = (b.rounds || 0) * 1000 + (b.reps || 0)
      return bTotal - aTotal
    } else if (wodScoreType.includes('Weight')) {
      // Weight (higher is better)
      return b.weight - a.weight
    }
    return 0
  })

  return sorted.slice(0, 3)
})

// Computed: Filtered Chart Data (for movement chart with rep scheme filter)
const filteredChartData = computed(() => {
  if (!selectedItem.value || selectedItem.value.type !== 'movement') return []

  if (selectedRepScheme.value === 'All Reps') {
    return performanceData.value.filter(p => p.weight)
  }

  const targetReps = parseInt(selectedRepScheme.value.split(' ')[0])
  return performanceData.value.filter(p => p.weight && p.reps === targetReps)
})

// Watch performance data and auto-select rep scheme
watch(performanceData, async (newData) => {
  if (!selectedItem.value || selectedItem.value.type !== 'movement') {
    return
  }

  const schemes = new Set(['All Reps'])
  newData.forEach(p => {
    if (p.reps) {
      schemes.add(`${p.reps} reps`)
    }
  })
  repSchemes.value = Array.from(schemes)

  // Auto-select the rep scheme with the heaviest weight
  if (newData.length > 0) {
    const heaviest = newData
      .filter(p => p.weight && p.reps)
      .sort((a, b) => b.weight - a.weight)[0]

    if (heaviest) {
      selectedRepScheme.value = `${heaviest.reps} reps`
    } else {
      selectedRepScheme.value = 'All Reps'
    }
  }
}, { immediate: true })

// Watch filtered chart data and render movement chart when it changes
watch(filteredChartData, async () => {
  if (selectedItem.value?.type === 'movement' && filteredChartData.value.length > 0) {
    await nextTick()
    renderMovementChart()
  }
}, { deep: true })

// Computed: WOD Performance Data for Chart
const wodPerformanceData = computed(() => {
  if (!selectedItem.value || selectedItem.value.type !== 'wod') return []
  return performanceData.value
})

// Computed: History Grouped by Year
const groupedHistory = computed(() => {
  if (!selectedItem.value || performanceData.value.length === 0) return {}

  const grouped = {}

  performanceData.value.forEach(entry => {
    const year = new Date(entry.created_at).getFullYear()
    if (!grouped[year]) {
      grouped[year] = []
    }
    grouped[year].push(entry)
  })

  // Sort years descending
  const sorted = {}
  Object.keys(grouped)
    .sort((a, b) => b - a)
    .forEach(year => {
      // Sort entries within year by date descending
      sorted[year] = grouped[year].sort((a, b) =>
        new Date(b.created_at) - new Date(a.created_at)
      )
    })

  return sorted
})

// Search Handler (debounced unified search)
let searchTimeout = null
function handleSearch(query) {
  searchQuery.value = query
  if (!query || query.length < 2) {
    searchResults.value = []
    return
  }

  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(async () => {
    await performUnifiedSearch(query)
  }, 300)
}

// Unified Search (Movements + WODs)
async function performUnifiedSearch(query) {
  loadingSearch.value = true
  try {
    const response = await axios.get('/api/performance/search', {
      params: { q: query, limit: 20 }
    })

    const results = response.data.results || []
    searchResults.value = results.map(r => ({
      id: `${r.type}-${r.id}`,
      name: r.name,
      type: r.type,
      data: r.data
    }))
  } catch (err) {
    console.error('Failed to search:', err)
    error.value = 'Failed to search movements and WODs'
    searchResults.value = []
  } finally {
    loadingSearch.value = false
  }
}

// Selection Handler
async function handleSelection(item) {
  if (!item) {
    clearSelection()
    return
  }

  selectedItem.value = item
  await fetchPerformanceData()
}

// Clear Selection
function clearSelection() {
  selectedItem.value = null
  performanceData.value = []
  destroyCharts()
}

// Fetch Performance Data (Movement or WOD specific)
async function fetchPerformanceData() {
  if (!selectedItem.value) return

  loadingPerformance.value = true
  error.value = null

  try {
    let response
    if (selectedItem.value.type === 'movement') {
      const movementId = selectedItem.value.data.id
      response = await axios.get(`/api/performance/movements/${movementId}`)
      performanceData.value = response.data.performances || []
    } else if (selectedItem.value.type === 'wod') {
      const wodId = selectedItem.value.data.id
      response = await axios.get(`/api/performance/wods/${wodId}`)
      performanceData.value = response.data.performances || []
    }

    // Note: Chart rendering is handled by the performanceData watcher for movements
    // For WODs, render directly since there's no rep scheme selection
    if (selectedItem.value.type === 'wod') {
      await nextTick()
      renderCharts()
    }
  } catch (err) {
    console.error('Failed to fetch performance data:', err)
    error.value = `Failed to load ${selectedItem.value.type} performance data`
    performanceData.value = []
  } finally {
    loadingPerformance.value = false
  }
}

// Filter Chart (when rep scheme changes)
async function filterChart() {
  await nextTick()
  renderCharts()
}

// Render Charts
function renderCharts() {
  if (selectedItem.value?.type === 'movement') {
    renderMovementChart()
  } else if (selectedItem.value?.type === 'wod') {
    renderWODChart()
  }
}

// Render Movement Chart
function renderMovementChart() {
  if (!chartCanvas.value || filteredChartData.value.length === 0) {
    return
  }

  destroyCharts()

  const data = filteredChartData.value
    .filter(p => p.weight)
    .sort((a, b) => new Date(a.created_at) - new Date(b.created_at))

  if (data.length === 0) {
    return
  }

  const labels = data.map(p => formatDate(p.created_at))
  const weights = data.map(p => p.weight)

  chartInstance = new Chart(chartCanvas.value, {
    type: 'line',
    data: {
      labels,
      datasets: [{
        label: 'Weight (lbs)',
        data: weights,
        borderColor: '#2c3e50',
        backgroundColor: 'rgba(44, 62, 80, 0.1)',
        tension: 0.4,
        pointRadius: 5,
        pointHoverRadius: 7,
        pointBackgroundColor: '#2c3e50',
        pointBorderColor: '#2c3e50',
        fill: true
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        },
        tooltip: {
          callbacks: {
            label: function(context) {
              const entry = data[context.dataIndex]
              let label = `${entry.weight} lbs`
              if (entry.reps) label += ` × ${entry.reps} reps`
              if (entry.is_pr) label += ' 🏆 PR'
              return label
            }
          }
        }
      },
      scales: {
        y: {
          beginAtZero: false,
          ticks: {
            stepSize: 1,
            callback: function(value) {
              return Math.round(value) + ' lbs'
            }
          }
        }
      }
    }
  })
}

// Render WOD Chart
function renderWODChart() {
  if (!wodChartCanvas.value || wodPerformanceData.value.length === 0) return

  destroyCharts()

  // Use the WOD's defined score_type to determine what to plot
  const wodScoreType = wodPerformanceData.value[0]?.wod_score_type || ''

  // Filter data to only include records matching the WOD's score_type
  let filteredData = [...wodPerformanceData.value]
  let values, label, isTimeBased = false

  if (wodScoreType.includes('Time')) {
    // Only include records with time_seconds
    filteredData = filteredData.filter(p => p.time_seconds)
    values = filteredData.map(p => p.time_seconds / 60)
    label = 'Time (minutes)'
    isTimeBased = true
  } else if (wodScoreType.includes('Rounds')) {
    // Only include records with rounds or reps
    filteredData = filteredData.filter(p => p.rounds !== null || p.reps !== null)
    values = filteredData.map(p => (p.rounds || 0) * 100 + (p.reps || 0))
    label = 'Rounds + Reps'
  } else if (wodScoreType.includes('Weight')) {
    // Only include records with weight
    filteredData = filteredData.filter(p => p.weight)
    values = filteredData.map(p => p.weight)
    label = 'Weight (lbs)'
  } else {
    return
  }

  if (filteredData.length === 0) return

  const data = filteredData.sort((a, b) => new Date(a.created_at) - new Date(b.created_at))
  const labels = data.map(p => formatDate(p.created_at))

  chartInstance = new Chart(wodChartCanvas.value, {
    type: 'line',
    data: {
      labels,
      datasets: [{
        label,
        data: values,
        borderColor: '#2c3e50',
        backgroundColor: 'rgba(44, 62, 80, 0.1)',
        tension: 0.4,
        pointRadius: 5,
        pointHoverRadius: 7,
        pointBackgroundColor: '#2c3e50',
        pointBorderColor: '#2c3e50',
        fill: true
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        },
        tooltip: {
          callbacks: {
            label: function(context) {
              const entry = data[context.dataIndex]
              let labelText = formatWODScore(entry)
              if (entry.is_pr) labelText += ' 🏆 PR'
              return labelText
            }
          }
        }
      },
      scales: {
        y: {
          beginAtZero: false,
          reverse: isTimeBased, // For time-based, lower is better
          ticks: {
            stepSize: 1,
            callback: function(value) {
              return Math.round(value)
            }
          }
        }
      }
    }
  })
}

// Destroy Charts
function destroyCharts() {
  if (chartInstance) {
    chartInstance.destroy()
    chartInstance = null
  }
  if (wodChartInstance) {
    wodChartInstance.destroy()
    wodChartInstance = null
  }
}

// Quick Log Navigation
function quickLog() {
  if (!selectedItem.value) return

  // Navigate to log workout page with pre-selected item
  if (selectedItem.value.type === 'movement') {
    router.push({
      path: '/workouts/log',
      query: { movement: selectedItem.value.data.id }
    })
  } else if (selectedItem.value.type === 'wod') {
    router.push({
      path: '/workouts/log',
      query: { wod: selectedItem.value.data.id }
    })
  }
}

// Format WOD Score
function formatWODScore(performance) {
  if (performance.time_seconds) {
    return formatTime(performance.time_seconds)
  } else if (performance.rounds !== null && performance.reps !== null) {
    return `${performance.rounds}+${performance.reps}`
  } else if (performance.rounds !== null) {
    return `${performance.rounds} rounds`
  } else if (performance.reps !== null) {
    return `${performance.reps} reps`
  } else if (performance.weight) {
    return `${performance.weight} lbs`
  } else if (performance.score_value) {
    return performance.score_value
  }
  return 'N/A'
}

// Format Date
function formatDate(dateString) {
  const datePart = dateString.split('T')[0]
  const [year, month, day] = datePart.split('-').map(Number)
  const date = new Date(year, month - 1, day)

  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  const dateOnly = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  const todayOnly = new Date(today.getFullYear(), today.getMonth(), today.getDate())
  const yesterdayOnly = new Date(yesterday.getFullYear(), yesterday.getMonth(), yesterday.getDate())

  if (dateOnly.getTime() === todayOnly.getTime()) {
    return 'Today'
  } else if (dateOnly.getTime() === yesterdayOnly.getTime()) {
    return 'Yesterday'
  } else {
    const options = { month: 'short', day: 'numeric', year: 'numeric' }
    return date.toLocaleDateString('en-US', options)
  }
}

// Format Time
function formatTime(seconds) {
  if (!seconds) return ''

  if (seconds < 60) {
    return `${seconds}s`
  } else if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60)
    const secs = seconds % 60
    return secs > 0 ? `${minutes}:${secs.toString().padStart(2, '0')}` : `${minutes}:00`
  } else {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    return `${hours}h ${minutes}m`
  }
}

// Format Movement Type
function formatMovementType(type) {
  if (!type) return ''
  return type.charAt(0).toUpperCase() + type.slice(1)
}

// Cleanup on unmount
onBeforeUnmount(() => {
  destroyCharts()
})
</script>
