<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <div class="d-flex align-center justify-space-between">
        <AdminHeader
          title="System Metrics"
          subtitle="Real-time system statistics and health"
          :breadcrumbs="[{ title: 'Metrics', to: '/admin/metrics' }]"
        />
        <v-btn
          icon
          variant="text"
          :loading="loading"
          @click="fetchMetrics"
        >
          <v-icon>mdi-refresh</v-icon>
        </v-btn>
      </div>

      <!-- Loading State -->
      <div v-if="loading && !metrics" class="text-center py-8">
        <v-progress-circular indeterminate color="primary" size="48"></v-progress-circular>
        <p class="mt-2 text-caption text-medium-emphasis">Loading metrics...</p>
      </div>

      <!-- Error State -->
      <v-alert v-else-if="error" type="error" class="mb-4">
        {{ error }}
      </v-alert>

      <!-- Metrics Content -->
      <template v-else-if="metrics">
        <!-- User Statistics -->
        <div class="mb-4">
          <h2 class="text-h6 mb-3" style="color: rgb(var(--v-theme-on-surface))">
            <v-icon size="20" class="mr-2" color="primary">mdi-account-group</v-icon>
            User Statistics
          </h2>
          <v-row dense>
            <v-col cols="6" md="3">
              <stat-card
                icon="mdi-account-multiple"
                color="primary"
                :value="metrics.user_stats?.total_users || 0"
                label="Total Users"
              />
            </v-col>
            <v-col cols="6" md="3">
              <stat-card
                icon="mdi-account-check"
                color="success"
                :value="metrics.user_stats?.active_this_month || 0"
                label="Active This Month"
              />
            </v-col>
            <v-col cols="6" md="3">
              <stat-card
                icon="mdi-account-plus"
                color="info"
                :value="metrics.user_stats?.new_this_month || 0"
                label="New This Month"
              />
            </v-col>
            <v-col cols="6" md="3">
              <stat-card
                icon="mdi-account-cancel"
                color="error"
                :value="metrics.user_stats?.disabled_users || 0"
                label="Disabled Users"
              />
            </v-col>
          </v-row>
        </div>

        <!-- Workout Statistics -->
        <div class="mb-4">
          <h2 class="text-h6 mb-3" style="color: rgb(var(--v-theme-on-surface))">
            <v-icon size="20" class="mr-2" color="secondary">mdi-dumbbell</v-icon>
            Workout Statistics
          </h2>
          <v-row dense>
            <v-col cols="6" md="4">
              <stat-card
                icon="mdi-dumbbell"
                color="secondary"
                :value="metrics.workout_stats?.total_workouts || 0"
                label="Total Workouts"
              />
            </v-col>
            <v-col cols="6" md="4">
              <stat-card
                icon="mdi-calendar-month"
                color="primary"
                :value="metrics.workout_stats?.workouts_this_month || 0"
                label="This Month"
              />
            </v-col>
            <v-col cols="12" md="4">
              <stat-card
                icon="mdi-chart-timeline-variant"
                color="success"
                :value="metrics.workout_stats?.avg_workouts_per_user || 0"
                label="Avg Per User"
                format="decimal"
              />
            </v-col>
          </v-row>
        </div>

        <!-- Workout Trends Chart -->
        <div class="mb-4">
          <workout-trends-chart
            :trends="metrics.workout_trends"
            :loading="loading"
          />
        </div>

        <!-- Content Statistics -->
        <div class="mb-4">
          <h2 class="text-h6 mb-3" style="color: rgb(var(--v-theme-on-surface))">
            <v-icon size="20" class="mr-2" color="accent">mdi-file-multiple</v-icon>
            Content Statistics
          </h2>
          <v-row dense>
            <v-col cols="6" md="4">
              <stat-card
                icon="mdi-weight-lifter"
                color="teal"
                :value="metrics.content_stats?.total_movements || 0"
                label="Total Movements"
                :subtitle="`${metrics.content_stats?.user_created_movements || 0} user-created`"
              />
            </v-col>
            <v-col cols="6" md="4">
              <stat-card
                icon="mdi-fire"
                color="orange"
                :value="metrics.content_stats?.total_wods || 0"
                label="Total WODs"
                :subtitle="`${metrics.content_stats?.user_created_wods || 0} user-created`"
              />
            </v-col>
            <v-col cols="12" md="4">
              <stat-card
                icon="mdi-clipboard-text"
                color="purple"
                :value="metrics.content_stats?.total_workout_templates || 0"
                label="Workout Templates"
                :subtitle="`${metrics.content_stats?.user_created_templates || 0} user-created`"
              />
            </v-col>
          </v-row>
        </div>

        <!-- Subscription Statistics -->
        <div class="mb-4">
          <h2 class="text-h6 mb-3" style="color: rgb(var(--v-theme-on-surface))">
            <v-icon size="20" class="mr-2" color="warning">mdi-currency-usd</v-icon>
            Subscription Statistics
          </h2>
          <v-row dense>
            <v-col cols="6" md="4">
              <stat-card
                icon="mdi-check-circle"
                color="success"
                :value="metrics.subscription_stats?.active_subscriptions || 0"
                label="Active Subscriptions"
              />
            </v-col>
            <v-col cols="6" md="4">
              <stat-card
                icon="mdi-clock-alert"
                color="warning"
                :value="metrics.subscription_stats?.expiring_soon || 0"
                label="Expiring (7 days)"
              />
            </v-col>
            <v-col cols="12" md="4">
              <stat-card
                icon="mdi-close-circle"
                color="error"
                :value="metrics.subscription_stats?.expired_subscriptions || 0"
                label="Expired"
              />
            </v-col>
          </v-row>
        </div>

        <!-- System Health -->
        <div class="mb-4">
          <h2 class="text-h6 mb-3" style="color: rgb(var(--v-theme-on-surface))">
            <v-icon size="20" class="mr-2" color="success">mdi-server</v-icon>
            System Health
          </h2>
          <v-row dense>
            <v-col cols="6" md="3">
              <stat-card
                icon="mdi-shield-check"
                color="info"
                :value="metrics.system_health?.recent_audit_events_24h || 0"
                label="Audit Events"
              />
            </v-col>
            <v-col cols="6" md="3">
              <stat-card
                icon="mdi-email-check"
                color="success"
                :value="metrics.system_health?.email_success_rate || 0"
                label="Email Success"
                format="percent"
              />
            </v-col>
            <v-col cols="6" md="3">
              <stat-card
                icon="mdi-email"
                color="primary"
                :value="metrics.system_health?.total_emails_sent || 0"
                label="Emails Sent"
              />
            </v-col>
            <v-col cols="6" md="3">
              <stat-card
                icon="mdi-email-alert"
                color="error"
                :value="metrics.system_health?.failed_emails || 0"
                label="Failed Emails"
              />
            </v-col>
          </v-row>
        </div>

        <!-- Last Updated -->
        <div class="text-center text-caption text-medium-emphasis mt-4">
          Last updated: {{ lastUpdated }}
        </div>
      </template>
    </v-container>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'
import AdminHeader from '@/components/AdminHeader.vue'
import StatCard from '@/components/admin/StatCard.vue'
import WorkoutTrendsChart from '@/components/admin/WorkoutTrendsChart.vue'

const loading = ref(false)
const error = ref(null)
const metrics = ref(null)
const lastUpdated = ref('')

async function fetchMetrics() {
  loading.value = true
  error.value = null

  try {
    const response = await axios.get('/api/admin/metrics')
    metrics.value = response.data
    lastUpdated.value = new Date().toLocaleString()
  } catch (err) {
    console.error('Failed to fetch admin metrics:', err)
    error.value = err.response?.data?.error || 'Failed to load metrics. Please try again.'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchMetrics()
})
</script>

<style scoped>
.v-card {
  transition: all 0.2s;
}
</style>
