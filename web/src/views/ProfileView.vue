<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <!-- Profile Card -->
      <v-card elevation="0" rounded class="pa-2 mb-1" bg-color="surface">
        <div class="text-center mb-3">
          <!-- Avatar with Upload -->
          <div style="position: relative; display: inline-block">
            <user-avatar :user="user" :size="100" />
            <v-btn
              icon
              size="small"
              color="primary"
              style="position: absolute; bottom: 0; right: 0"
              :loading="uploadingAvatar"
              @click="openFileDialog"
            >
              <v-icon size="small">mdi-camera</v-icon>
            </v-btn>
          </div>

          <!-- Hidden File Input -->
          <input
            ref="fileInput"
            type="file"
            accept="image/*"
            style="display: none"
            @change="handleFileSelect"
          />

          <h2 class="text-h6 mt-3 font-weight-bold" >
            {{ user?.name || 'User' }}
          </h2>
          <p class="text-body-2 text-medium-emphasis">{{ user?.email || 'email@example.com' }}</p>

          <!-- Delete Avatar Button (only if avatar exists) -->
          <v-btn
            v-if="user?.profile_image"
            size="x-small"
            variant="text"
            color="error"
            class="mt-2"
            :loading="deletingAvatar"
            @click="deleteAvatar"
          >
            <v-icon start size="x-small">mdi-delete</v-icon>
            Remove Avatar
          </v-btn>

          <v-chip
            v-if="user?.role === 'admin'"
            size="small"
            color="error"
            class="mt-2"
          >
            <v-icon start size="x-small">mdi-shield-crown</v-icon>
            Admin
          </v-chip>
        </div>

        <!-- Member Since -->
        <div v-if="user?.created_at" class="text-center text-caption text-disabled">
          Member since {{ formatMemberSince(user.created_at) }}
        </div>
      </v-card>

      <!-- Version Info -->
      <v-card elevation="0" rounded class="pa-2 mb-1" bg-color="surface">
        <div class="d-flex align-center justify-space-between">
          <div>
            <div class="text-caption text-disabled">Version</div>
            <div class="text-body-2 font-weight-bold" >
              {{ appVersion }}
            </div>
          </div>
          <div class="text-right">
            <div class="text-caption text-disabled">Build</div>
            <div class="text-body-2 font-weight-bold" >
              #{{ buildNumber }}
            </div>
          </div>
        </div>
      </v-card>

      <!-- Subscription Status -->
      <v-card elevation="0" rounded class="pa-2 mb-1" bg-color="surface">
        <div class="d-flex align-center justify-space-between mb-2">
          <h2 class="text-body-1 font-weight-bold" >Subscription</h2>
          <v-chip
            v-if="subscriptionStore.isPermanentFree"
            size="small"
            color="purple"
            variant="flat"
          >
            <v-icon start size="x-small">mdi-crown</v-icon>
            Permanent Free
          </v-chip>
          <v-chip
            v-else-if="subscriptionStore.hasAccess"
            size="small"
            :color="getSubscriptionColor(subscriptionStore.subscriptionType)"
            variant="flat"
          >
            <v-icon start size="x-small">mdi-check-circle</v-icon>
            {{ subscriptionStore.subscriptionTypeLabel }}
          </v-chip>
          <v-chip
            v-else
            size="small"
            color="error"
            variant="flat"
          >
            <v-icon start size="x-small">mdi-alert-circle</v-icon>
            Expired
          </v-chip>
        </div>

        <!-- Loading State -->
        <div v-if="subscriptionStore.loading" class="text-center py-2">
          <v-progress-circular indeterminate color="primary" size="24" />
        </div>

        <!-- Subscription Details -->
        <div v-else-if="subscriptionStore.subscriptionStatus">
          <!-- User Subscription -->
          <div v-if="subscriptionStore.userSubscription" class="mb-2">
            <div class="d-flex align-center justify-space-between">
              <div class="text-caption text-disabled">Status</div>
              <div class="text-body-2" :style="{ color: subscriptionStore.hasAccess ? '#4caf50' : '#e91e63' }">
                {{ subscriptionStore.hasAccess ? 'Active' : 'Expired' }}
              </div>
            </div>

            <!-- Expiration Date -->
            <div
              v-if="subscriptionStore.userSubscription.end_date && !subscriptionStore.isPermanentFree"
              class="d-flex align-center justify-space-between mt-1"
            >
              <div class="text-caption text-disabled">
                {{ subscriptionStore.hasAccess ? 'Expires' : 'Expired' }}
              </div>
              <div class="text-body-2" >
                {{ formatDate(subscriptionStore.userSubscription.end_date) }}
              </div>
            </div>

            <!-- Next Billing Date -->
            <div
              v-if="subscriptionStore.userSubscription.next_billing_date && subscriptionStore.hasAccess"
              class="d-flex align-center justify-space-between mt-1"
            >
              <div class="text-caption text-disabled">Next Billing</div>
              <div class="text-body-2" >
                {{ formatDate(subscriptionStore.userSubscription.next_billing_date) }}
              </div>
            </div>
          </div>

          <!-- Organization Subscriptions -->
          <div v-if="subscriptionStore.orgSubscriptions.length > 0" class="mt-2 pt-2" style="border-top: 1px solid #eee">
            <div class="text-caption mb-1 text-disabled">
              Organization Access
            </div>
            <div
              v-for="orgSub in subscriptionStore.orgSubscriptions"
              :key="orgSub.id"
              class="d-flex align-center justify-space-between mb-1"
            >
              <div class="text-body-2" >
                {{ orgSub.organization_name }}
              </div>
              <v-chip
                size="x-small"
                :color="orgSub.status === 'active' ? 'success' : 'error'"
                variant="flat"
              >
                {{ orgSub.status }}
              </v-chip>
            </div>
          </div>

          <!-- Access Source -->
          <div v-if="subscriptionStore.source !== 'none'" class="mt-2">
            <div class="text-caption text-disabled">
              Access via: {{ formatSource(subscriptionStore.source) }}
            </div>
          </div>
        </div>

        <!-- Error State -->
        <div v-else-if="subscriptionStore.error" class="text-center py-2">
          <div class="text-caption" style="color: rgb(var(--v-theme-error))">
            {{ subscriptionStore.error }}
          </div>
        </div>
      </v-card>

      <!-- Stats Summary -->
      <v-card elevation="0" rounded class="pa-2 mb-1" bg-color="surface">
        <div class="d-flex align-center justify-space-between mb-2">
          <h2 class="text-body-1 font-weight-bold" >Workout Summary</h2>
        </div>

        <!-- Time Period Filter -->
        <div class="mb-3">
          <v-chip-group v-model="selectedPeriod" mandatory color="primary" density="compact">
            <v-chip value="week" size="small">This Week</v-chip>
            <v-chip value="month" size="small">This Month</v-chip>
            <v-chip value="year" size="small">This Year</v-chip>
            <v-chip value="all" size="small">All Time</v-chip>
          </v-chip-group>
        </div>

        <!-- Loading State -->
        <div v-if="loadingStats" class="text-center py-4">
          <v-progress-circular indeterminate color="primary" size="32" />
        </div>

        <!-- Stats Grid -->
        <v-row v-else dense>
          <v-col cols="6">
            <v-card elevation="0" rounded class="pa-1 text-center" style="background-color: rgb(var(--v-theme-background))">
              <div class="text-h5 font-weight-bold" style="color: rgb(var(--v-theme-primary))">
                {{ stats.totalWorkouts }}
              </div>
              <div class="text-caption text-medium-emphasis">Total Workouts</div>
            </v-card>
          </v-col>
          <v-col cols="6">
            <v-card elevation="0" rounded class="pa-1 text-center" style="background-color: rgb(var(--v-theme-background))">
              <div class="text-h5 font-weight-bold" style="color: rgb(var(--v-theme-success))">
                {{ stats.currentStreak }}
              </div>
              <div class="text-caption text-medium-emphasis">Day Streak</div>
            </v-card>
          </v-col>
          <v-col cols="6">
            <v-card elevation="0" rounded class="pa-1 text-center" style="background-color: rgb(var(--v-theme-background))">
              <div class="text-h5 font-weight-bold" style="color: #ffc107">
                {{ stats.personalRecords }}
              </div>
              <div class="text-caption text-medium-emphasis">Personal Records</div>
            </v-card>
          </v-col>
          <v-col cols="6">
            <v-card elevation="0" rounded class="pa-1 text-center" style="background-color: rgb(var(--v-theme-background))">
              <div class="text-h5 font-weight-bold" style="color: rgb(var(--v-theme-error))">
                {{ stats.customTemplates }}
              </div>
              <div class="text-caption text-medium-emphasis">Custom Templates</div>
            </v-card>
          </v-col>
        </v-row>
      </v-card>

      <!-- Quick Actions -->
      <v-card elevation="0" rounded class="pa-2 mb-1" bg-color="surface">
        <h2 class="text-body-1 font-weight-bold mb-1" >Quick Actions</h2>
        <v-list bg-color="transparent" density="compact">
          <v-list-item
            prepend-icon="mdi-dumbbell"
            rounded
            style="cursor: pointer"
            @click="$router.push('/workouts')"
          >
            <v-list-item-title class="font-weight-medium" >
              My Templates
            </v-list-item-title>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-fire"
            rounded
            style="cursor: pointer"
            @click="$router.push('/wods')"
          >
            <v-list-item-title class="font-weight-medium" >
              Benchmark WODs
            </v-list-item-title>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-trophy"
            rounded
            style="cursor: pointer"
            @click="$router.push('/prs')"
          >
            <v-list-item-title class="font-weight-medium" >
              Personal Records
            </v-list-item-title>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-divider class="my-2" />

          <v-list-item
            prepend-icon="mdi-download"
            rounded
            style="cursor: pointer"
            @click="$router.push('/settings/export')"
          >
            <v-list-item-title class="font-weight-medium" >
              Export Data
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-disabled">
              Download WODs and Movements as CSV
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-upload"
            rounded
            style="cursor: pointer"
            @click="$router.push('/settings/import')"
          >
            <v-list-item-title class="font-weight-medium" >
              Import Data
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-disabled">
              Upload WODs and Movements from CSV
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>
        </v-list>
      </v-card>

      <!-- Administration (Admin Only) -->
      <v-card v-if="user?.role === 'admin'" elevation="0" rounded class="pa-2 mb-1" bg-color="surface">
        <h2 class="text-body-1 font-weight-bold mb-1" >
          <v-icon color="error" size="small" class="mr-1">mdi-shield-crown</v-icon>
          Administration
        </h2>
        <v-list bg-color="transparent" density="compact">
          <v-list-item
            prepend-icon="mdi-database-refresh"
            rounded
            style="cursor: pointer"
            @click="$router.push('/admin/data-cleanup')"
          >
            <v-list-item-title class="font-weight-medium" >
              Data Cleanup
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-disabled">
              Fix WOD score_type mismatches
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-account-multiple"
            rounded
            style="cursor: pointer"
            @click="$router.push('/admin/users')"
          >
            <v-list-item-title class="font-weight-medium" >
              User Management
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-disabled">
              Manage accounts and permissions
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-file-multiple"
            rounded
            style="cursor: pointer"
            @click="$router.push('/admin/user-content')"
          >
            <v-list-item-title class="font-weight-medium" >
              User Content
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-disabled">
              Manage user-created WODs, movements, workouts
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-history"
            rounded
            style="cursor: pointer"
            @click="$router.push('/admin/audit-logs')"
          >
            <v-list-item-title class="font-weight-medium" >
              Audit Logs
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-disabled">
              View security events and activity
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-database-export"
            rounded
            style="cursor: pointer"
            @click="$router.push('/admin/backups')"
          >
            <v-list-item-title class="font-weight-medium" >
              Database Backups
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-disabled">
              Create and restore backups
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-file-document-edit"
            rounded
            style="cursor: pointer"
            @click="$router.push('/admin/data-change-logs')"
          >
            <v-list-item-title class="font-weight-medium" >
              Data Change Logs
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-disabled">
              Track edits and deletions
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-bullhorn"
            rounded
            style="cursor: pointer"
            @click="$router.push('/admin/announcements')"
          >
            <v-list-item-title class="font-weight-medium" >
              Announcements
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-disabled">
              Send notifications to users
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-domain"
            rounded
            style="cursor: pointer"
            @click="$router.push('/admin/organizations')"
          >
            <v-list-item-title class="font-weight-medium" >
              Organization Management
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-disabled">
              Manage organizations and members
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-credit-card"
            rounded
            style="cursor: pointer"
            @click="$router.push('/admin/subscriptions')"
          >
            <v-list-item-title class="font-weight-medium" >
              Subscription Management
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-disabled">
              Manage user and org subscriptions
            </v-list-item-subtitle>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-chart-bar"
            disabled
            rounded
          >
            <v-list-item-title class="font-weight-medium text-disabled">
              System Reports
            </v-list-item-title>
            <v-list-item-subtitle class="text-caption text-disabled">
              Coming soon
            </v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </v-card>

      <!-- Account Actions -->
      <v-card elevation="0" rounded class="pa-2" bg-color="surface">
        <h2 class="text-body-1 font-weight-bold mb-1" >Account</h2>
        <v-list bg-color="transparent" density="compact">
          <v-list-item
            prepend-icon="mdi-cog"
            rounded
            style="cursor: pointer"
            @click="$router.push('/settings')"
          >
            <v-list-item-title class="font-weight-medium" >
              Settings
            </v-list-item-title>
            <template #append>
              <v-icon color="surface-variant" size="small">mdi-chevron-right</v-icon>
            </template>
          </v-list-item>

          <v-list-item
            prepend-icon="mdi-logout"
            rounded
            style="cursor: pointer"
            @click="handleLogout"
          >
            <v-list-item-title class="font-weight-medium" style="color: rgb(var(--v-theme-error))">
              Logout
            </v-list-item-title>
          </v-list-item>
        </v-list>
      </v-card>
    </v-container>

    <!-- Bottom Navigation -->
    <v-bottom-navigation
      grow
      style="position: fixed; bottom: 0; background-color: rgb(var(--v-theme-surface))"
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
        <v-avatar color="primary" size="56" style="box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2)">
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSubscriptionStore } from '@/stores/subscription'
import axios from '@/utils/axios'
import UserAvatar from '@/components/UserAvatar.vue'
import { getProfileImageUrl } from '@/utils/url'

const router = useRouter()
const authStore = useAuthStore()
const subscriptionStore = useSubscriptionStore()
const activeTab = ref('profile')


// Version info
const appVersion = ref('...')
const buildNumber = ref(0)

const user = computed(() => authStore.user)
const loadingStats = ref(false)
const uploadingAvatar = ref(false)
const deletingAvatar = ref(false)
const fileInput = ref(null)

// Time period filter
const selectedPeriod = ref('month')

// Stats
const stats = ref({
  totalWorkouts: 0,
  currentStreak: 0,
  personalRecords: 0,
  customTemplates: 0
})

// Get date range for selected period
function getDateRange(period) {
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  let startDate = new Date(today)

  switch (period) {
    case 'week':
      // Start of current week (Sunday)
      const dayOfWeek = today.getDay()
      startDate.setDate(today.getDate() - dayOfWeek)
      break
    case 'month':
      // Start of current month
      startDate = new Date(now.getFullYear(), now.getMonth(), 1)
      break
    case 'year':
      // Start of current year
      startDate = new Date(now.getFullYear(), 0, 1)
      break
    case 'all':
      // No filtering
      return null
  }

  return {
    start: startDate.toISOString().split('T')[0],
    end: today.toISOString().split('T')[0]
  }
}

// Filter workouts by date range
function filterWorkoutsByPeriod(workouts, period) {
  if (period === 'all') return workouts

  const range = getDateRange(period)
  if (!range) return workouts

  return workouts.filter(workout => {
    const workoutDate = workout.workout_date
    return workoutDate >= range.start && workoutDate <= range.end
  })
}

// Filter PR movements by date range
function filterPRsByPeriod(prMovements, workouts, period) {
  if (period === 'all') return prMovements
  if (!prMovements || prMovements.length === 0) return []

  const range = getDateRange(period)
  if (!range) return prMovements

  // Filter by last_pr_date
  return prMovements.filter(movement => {
    const lastPRDate = movement.last_pr_date
    if (!lastPRDate) return false
    return lastPRDate >= range.start && lastPRDate <= range.end
  })
}

// Fetch user statistics
async function fetchStats() {
  loadingStats.value = true
  try {
    // Fetch workouts for stats
    const [workoutsRes, prsRes, templatesRes] = await Promise.all([
      axios.get('/api/workouts').catch(() => ({ data: { workouts: [] } })),
      axios.get('/api/pr-movements').catch(() => ({ data: { movements: [] } })),
      axios.get('/api/workouts/my-templates').catch(() => ({ data: { workouts: [] } }))
    ])

    const allWorkouts = workoutsRes.data.workouts || []
    const allPRMovements = prsRes.data.movements || []

    // Filter based on selected period
    const filteredWorkouts = filterWorkoutsByPeriod(allWorkouts, selectedPeriod.value)
    const filteredPRs = filterPRsByPeriod(allPRMovements, allWorkouts, selectedPeriod.value)

    // Calculate total PR count by summing pr_count from each movement
    const totalPRs = filteredPRs.reduce((sum, movement) => {
      return sum + (movement.pr_count || 0)
    }, 0)

    // Calculate stats
    stats.value = {
      totalWorkouts: filteredWorkouts.length,
      currentStreak: calculateStreak(allWorkouts), // Always use all workouts for streak
      personalRecords: totalPRs,
      customTemplates: (templatesRes.data.workouts || []).length // Always show all templates
    }
  } catch (err) {
    console.error('Failed to fetch stats:', err)
  } finally {
    loadingStats.value = false
  }
}

// Calculate current streak
function calculateStreak(workouts) {
  if (workouts.length === 0) return 0

  // Get unique workout dates (in case of multiple workouts per day)
  const uniqueDates = [...new Set(workouts.map(w => w.workout_date))]
    .sort((a, b) => new Date(b) - new Date(a))

  if (uniqueDates.length === 0) return 0

  const today = new Date()
  today.setHours(0, 0, 0, 0)

  const mostRecent = new Date(uniqueDates[0])
  mostRecent.setHours(0, 0, 0, 0)

  // Calculate days since most recent workout
  const daysSinceRecent = Math.floor((today - mostRecent) / (1000 * 60 * 60 * 24))

  // Streak is broken if no workout today or yesterday
  if (daysSinceRecent > 1) return 0

  // Start counting streak from most recent workout
  let streak = 1
  const expectedDate = new Date(mostRecent)

  for (let i = 1; i < uniqueDates.length; i++) {
    expectedDate.setDate(expectedDate.getDate() - 1)
    const workoutDate = new Date(uniqueDates[i])
    workoutDate.setHours(0, 0, 0, 0)

    // Check if this workout is exactly 1 day before the previous
    if (workoutDate.getTime() === expectedDate.getTime()) {
      streak++
    } else {
      // Gap found, streak is broken
      break
    }
  }

  return streak
}

// Format member since date
function formatMemberSince(dateString) {
  const date = new Date(dateString)
  const options = { month: 'long', year: 'numeric' }
  return date.toLocaleDateString('en-US', options)
}

// Format date for subscription display
function formatDate(dateString) {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  const options = { month: 'short', day: 'numeric', year: 'numeric' }
  return date.toLocaleDateString('en-US', options)
}

// Get subscription color based on type
function getSubscriptionColor(type) {
  const colorMap = {
    free: 'grey',
    monthly: 'blue',
    annual: 'green'
  }
  return colorMap[type] || 'grey'
}

// Format access source for display
function formatSource(source) {
  const sourceMap = {
    user: 'Personal Subscription',
    organization: 'Organization Subscription',
    both: 'Personal & Organization',
    none: 'No Access'
  }
  return sourceMap[source] || source
}

const handleLogout = () => {
  authStore.logout()
  subscriptionStore.clear()
  router.push('/login')
}

// Avatar Upload Functions
function openFileDialog() {
  fileInput.value?.click()
}

async function handleFileSelect(event) {
  const file = event.target.files?.[0]
  if (!file) return

  // Validate file type
  if (!file.type.startsWith('image/')) {
    alert('Please select an image file')
    return
  }

  // Validate file size (5MB max)
  if (file.size > 5 * 1024 * 1024) {
    alert('Image must be smaller than 5MB')
    return
  }

  uploadingAvatar.value = true

  try {
    // Create FormData for file upload
    const formData = new FormData()
    formData.append('avatar', file)

    // Upload to backend
    const response = await axios.post('/api/users/avatar', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })

    // Update user in auth store - need to update the ref value property
    const updatedUser = response.data.user
    // Ensure profile_image has full URL for display
    if (updatedUser.profile_image) {
      updatedUser.profile_image = getProfileImageUrl(updatedUser.profile_image)
    }
    authStore.user = updatedUser
    localStorage.setItem('user', JSON.stringify(updatedUser))

    // Clear file input
    if (fileInput.value) {
      fileInput.value.value = ''
    }
  } catch (err) {
    console.error('Failed to upload avatar:', err)
    alert(err.response?.data?.message || 'Failed to upload avatar')
  } finally {
    uploadingAvatar.value = false
  }
}

async function deleteAvatar() {
  if (!confirm('Are you sure you want to remove your avatar?')) return

  deletingAvatar.value = true

  try {
    const response = await axios.delete('/api/users/avatar')

    // Update user in auth store
    const updatedUser = response.data.user
    authStore.user = updatedUser
    localStorage.setItem('user', JSON.stringify(updatedUser))
  } catch (err) {
    console.error('Failed to delete avatar:', err)
    alert(err.response?.data?.message || 'Failed to delete avatar')
  } finally {
    deletingAvatar.value = false
  }
}

// Fetch version info
async function fetchVersionInfo() {
  try {
    const response = await axios.get('/api/version')
    appVersion.value = response.data.fullVersion || response.data.version
    buildNumber.value = response.data.build || 0
    console.log('Version info loaded:', response.data)
  } catch (err) {
    console.error('Failed to fetch version info:', err)
    // Fallback to showing something
    appVersion.value = '0.4.1-beta'
    buildNumber.value = 1
  }
}

// Watch for period changes
watch(selectedPeriod, () => {
  fetchStats()
})

onMounted(() => {
  fetchStats()
  fetchVersionInfo()
  subscriptionStore.fetchStatus().catch(err => {
    console.error('Failed to fetch subscription status:', err)
  })
})
</script>
