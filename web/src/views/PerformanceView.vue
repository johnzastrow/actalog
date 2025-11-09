<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-toolbar-title class="text-white font-weight-bold">Performance</v-toolbar-title>
    </v-app-bar>

    <v-container class="pa-3" style="margin-top: 56px; margin-bottom: 70px">
      <!-- Progress Tracking -->
      <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
        <h2 class="text-body-1 font-weight-bold mb-2" style="color: #1a1a1a">Progress Tracking</h2>
        <v-autocomplete
          v-model="selectedMovement"
          :items="movements"
          item-title="title"
          item-value="value"
          :loading="loading"
          label="Search for a movement..."
          placeholder="Type to search..."
          variant="outlined"
          density="compact"
          rounded="lg"
          clearable
          auto-select-first
          hide-details
          style="color: #1a1a1a"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4">mdi-magnify</v-icon>
          </template>
          <template #item="{ props, item }">
            <v-list-item v-bind="props">
              <template #prepend>
                <v-icon
                  :color="item.raw.type === 'weightlifting' ? '#00bcd4' : '#666'"
                  size="small"
                >
                  mdi-dumbbell
                </v-icon>
              </template>
              <v-list-item-title class="text-body-2">
                {{ item.raw.title }}
              </v-list-item-title>
              <v-list-item-subtitle class="text-caption">
                {{ item.raw.type }}
              </v-list-item-subtitle>
            </v-list-item>
          </template>
        </v-autocomplete>
      </v-card>

      <!-- Chart Placeholder -->
      <v-card elevation="0" rounded="lg" class="pa-4 text-center" style="background: white">
        <v-icon size="64" color="#ccc">mdi-chart-line</v-icon>
        <p class="text-body-2 mt-2" style="color: #666">
          Charts and performance metrics will appear here
        </p>
        <p class="text-caption" style="color: #999">
          Log more workouts to track your progress
        </p>
      </v-card>
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
      <v-btn value="log" to="/workouts/log" style="position: relative; bottom: 20px">
        <v-avatar color="#ffc107" size="56" style="box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2)">
          <v-icon color="white" size="32">mdi-plus</v-icon>
        </v-avatar>
      </v-btn>
      <v-btn value="workouts" to="/workouts">
        <v-icon>mdi-dumbbell</v-icon>
        <span style="font-size: 10px">Workouts</span>
      </v-btn>
      <v-btn value="profile" to="/profile">
        <v-icon>mdi-account</v-icon>
        <span style="font-size: 10px">Profile</span>
      </v-btn>
    </v-bottom-navigation>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'

const activeTab = ref('performance')
const selectedMovement = ref(null)
const movements = ref([])
const loading = ref(false)

// Fetch available movements
async function fetchMovements() {
  loading.value = true
  try {
    const response = await axios.get('/api/movements')
    movements.value = response.data.movements.map((m) => ({
      value: m.id,
      title: m.name,
      type: m.type,
    }))
  } catch (err) {
    console.error('Failed to fetch movements:', err)
    // Fallback to hardcoded movements
    movements.value = [
      { value: 1, title: 'Back Squat', type: 'weightlifting' },
      { value: 2, title: 'Deadlift', type: 'weightlifting' },
      { value: 3, title: 'Bench Press', type: 'weightlifting' },
    ]
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchMovements()
})
</script>
