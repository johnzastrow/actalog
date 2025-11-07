<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4 mb-6">Performance</h1>
      </v-col>

      <v-col cols="12">
        <v-card elevation="2" rounded="lg">
          <v-card-title>Progress Tracking</v-card-title>
          <v-card-text>
            <v-select
              v-model="selectedMovement"
              label="Select Movement"
              :items="movements"
              variant="outlined"
            />
            <div class="text-center pa-10 text-medium-emphasis">
              Charts and performance metrics will appear here
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-bottom-navigation v-model="activeTab" grow>
      <v-btn value="dashboard" to="/dashboard">
        <v-icon>mdi-view-dashboard</v-icon>
        <span>Dashboard</span>
      </v-btn>
      <v-btn value="performance" to="/performance">
        <v-icon>mdi-chart-line</v-icon>
        <span>Performance</span>
      </v-btn>
      <v-btn value="log" to="/workouts/log" color="gold">
        <v-icon size="large">mdi-plus</v-icon>
      </v-btn>
      <v-btn value="workouts" to="/workouts">
        <v-icon>mdi-dumbbell</v-icon>
        <span>Workouts</span>
      </v-btn>
      <v-btn value="profile" to="/profile">
        <v-icon>mdi-account</v-icon>
        <span>Profile</span>
      </v-btn>
    </v-bottom-navigation>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'

const activeTab = ref('performance')
const selectedMovement = ref(null)
const movements = ref(['Back Squat', 'Deadlift', 'Bench Press', 'Clean', 'Snatch'])
</script>
