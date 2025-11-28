<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
        <!-- Page Title -->
        <div class="mb-4">
          <h1 style="color: #2c3e50; font-size: 24px; font-weight: 600">Administration</h1>
          <p style="color: #666; font-size: 14px">Manage system data and configurations</p>
        </div>

        <!-- Admin Tools Grid -->
        <v-row>
          <!-- Data Cleanup Tool -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" @click="navigateTo('/admin/data-cleanup')">
              <div class="d-flex align-center mb-3">
                <v-icon color="#00bcd4" size="32" class="mr-3">mdi-database-refresh</v-icon>
                <div>
                  <h3 style="color: #2c3e50; font-size: 18px; font-weight: 600">Data Cleanup</h3>
                  <p style="color: #666; font-size: 14px; margin: 0">Fix WOD score_type mismatches</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: #666; font-size: 13px">
                Detect and correct WOD performance records that don't match their defined score_type.
              </p>
              <v-chip size="small" color="#00bcd4" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- User Management -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover @click="$router.push('/admin/users')" style="cursor: pointer">
              <div class="d-flex align-center mb-3">
                <v-icon color="#9c27b0" size="32" class="mr-3">mdi-account-multiple</v-icon>
                <div>
                  <h3 style="color: #1a1a1a; font-size: 18px; font-weight: 600">User Management</h3>
                  <p style="color: #666; font-size: 14px; margin: 0">Manage user accounts</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: #666; font-size: 13px">
                View, edit, and manage user accounts and permissions.
              </p>
              <v-chip size="small" color="#9c27b0" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- User Content Management -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover @click="$router.push('/admin/user-content')" style="cursor: pointer">
              <div class="d-flex align-center mb-3">
                <v-icon color="#e91e63" size="32" class="mr-3">mdi-file-multiple</v-icon>
                <div>
                  <h3 style="color: #1a1a1a; font-size: 18px; font-weight: 600">User Content</h3>
                  <p style="color: #666; font-size: 14px; margin: 0">Manage user-created content</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: #666; font-size: 13px">
                View user-created WODs, movements, and workout templates. Promote them to standard items.
              </p>
              <v-chip size="small" color="#e91e63" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- Database Backups -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover @click="$router.push('/admin/backups')" style="cursor: pointer">
              <div class="d-flex align-center mb-3">
                <v-icon color="#ff9800" size="32" class="mr-3">mdi-database-export</v-icon>
                <div>
                  <h3 style="color: #1a1a1a; font-size: 18px; font-weight: 600">Database Backups</h3>
                  <p style="color: #666; font-size: 14px; margin: 0">Create and restore database backups</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: #666; font-size: 13px">
                Protect your data with full database backups. Create, download, and restore backups.
              </p>
              <v-chip size="small" color="#ff9800" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- Audit Logs -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover @click="$router.push('/admin/audit-logs')" style="cursor: pointer">
              <div class="d-flex align-center mb-3">
                <v-icon color="#4caf50" size="32" class="mr-3">mdi-shield-check</v-icon>
                <div>
                  <h3 style="color: #1a1a1a; font-size: 18px; font-weight: 600">Audit Logs</h3>
                  <p style="color: #666; font-size: 14px; margin: 0">View security events</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: #666; font-size: 13px">
                Monitor login attempts, password changes, and other security-related events.
              </p>
              <v-chip size="small" color="#4caf50" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- Data Change Logs -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover @click="$router.push('/admin/data-change-logs')" style="cursor: pointer">
              <div class="d-flex align-center mb-3">
                <v-icon color="#2196f3" size="32" class="mr-3">mdi-history</v-icon>
                <div>
                  <h3 style="color: #1a1a1a; font-size: 18px; font-weight: 600">Data Change Logs</h3>
                  <p style="color: #666; font-size: 14px; margin: 0">Track edits and deletions</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: #666; font-size: 13px">
                View history of all data changes including before/after values for WODs, movements, and workouts.
              </p>
              <v-chip size="small" color="#2196f3" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- Reports (Placeholder) -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" disabled>
              <div class="d-flex align-center mb-3">
                <v-icon color="#999" size="32" class="mr-3">mdi-chart-bar</v-icon>
                <div>
                  <h3 style="color: #999; font-size: 18px; font-weight: 600">Reports & Analytics</h3>
                  <p style="color: #999; font-size: 14px; margin: 0">View system statistics</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: #999; font-size: 13px">
                Generate reports on user activity, workout trends, and system health.
              </p>
              <v-chip size="small" color="#999" class="mt-2">
                Coming Soon
              </v-chip>
            </v-card>
          </v-col>
        </v-row>
      </v-container>

    <!-- Bottom Navigation -->
    <v-bottom-navigation fixed style="z-index: 5; height: 56px; box-shadow: 0px -2px 4px rgba(0,0,0,0.1)">
      <v-btn to="/dashboard" style="flex: 1">
        <v-icon size="20">mdi-view-dashboard</v-icon>
        <span style="font-size: 9px">Home</span>
      </v-btn>
      <v-btn to="/log-workout" style="flex: 1">
        <v-icon size="20">mdi-plus-circle</v-icon>
        <span style="font-size: 9px">Log</span>
      </v-btn>
      <v-btn to="/performance" style="flex: 1">
        <v-icon size="20">mdi-chart-line</v-icon>
        <span style="font-size: 9px">Stats</span>
      </v-btn>
      <v-btn to="/admin" style="flex: 1" color="#00bcd4">
        <v-icon size="20">mdi-shield-account</v-icon>
        <span style="font-size: 9px">Admin</span>
      </v-btn>
    </v-bottom-navigation>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const navigateTo = (path) => {
  router.push(path)
}

const logout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.v-card {
  cursor: pointer;
  transition: all 0.2s;
}

.v-card:not([disabled]):hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1) !important;
}

.v-card[disabled] {
  cursor: not-allowed;
  opacity: 0.6;
}
</style>
