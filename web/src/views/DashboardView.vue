<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4 mb-6">Dashboard</h1>
      </v-col>

      <!-- Quick Action -->
      <v-col cols="12">
        <v-card elevation="2" rounded="lg">
          <v-card-text class="pa-6">
            <h2 class="text-h6 mb-4">Quick Actions</h2>
            <v-btn
              color="primary"
              size="large"
              prepend-icon="mdi-plus"
              to="/workouts/log"
            >
              Log Today's Workout
            </v-btn>
          </v-card-text>
        </v-card>
      </v-col>

      <!-- Recent Activity -->
      <v-col cols="12" md="8">
        <v-card elevation="2" rounded="lg">
          <v-card-title>Recent Workouts</v-card-title>
          <v-card-text>
            <v-list>
              <v-list-item
                v-for="n in 5"
                :key="n"
                subtitle="No workouts logged yet"
              >
                <template #prepend>
                  <v-icon>mdi-dumbbell</v-icon>
                </template>
                <v-list-item-title>Workout Placeholder {{ n }}</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>
      </v-col>

      <!-- Stats Summary -->
      <v-col cols="12" md="4">
        <v-card elevation="2" rounded="lg" class="mb-4">
          <v-card-text class="text-center pa-6">
            <div class="text-h3 text-primary">0</div>
            <div class="text-body-1 text-medium-emphasis">Total Workouts</div>
          </v-card-text>
        </v-card>

        <v-card elevation="2" rounded="lg">
          <v-card-text class="text-center pa-6">
            <div class="text-h3 text-secondary">0</div>
            <div class="text-body-1 text-medium-emphasis">This Month</div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Bottom Navigation -->
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

const activeTab = ref('dashboard')
</script>
