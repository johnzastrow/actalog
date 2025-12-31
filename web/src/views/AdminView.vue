<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
        <!-- Page Title -->
        <div class="mb-4">
          <h1 style="color: #2c3e50; font-size: 24px; font-weight: 600">Administration</h1>
          <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px">Manage system data and configurations</p>
        </div>

        <!-- Admin Tools Grid -->
        <v-row>
          <!-- Data Cleanup Tool -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" @click="navigateTo('/admin/data-cleanup')">
              <div class="d-flex align-center mb-3">
                <v-icon color="primary" size="32" class="mr-3">mdi-database-refresh</v-icon>
                <div>
                  <h3 style="color: #2c3e50; font-size: 18px; font-weight: 600">Data Cleanup</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Fix WOD score_type mismatches</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                Detect and correct WOD performance records that don't match their defined score_type.
              </p>
              <v-chip size="small" color="primary" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- User Management -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/users')">
              <div class="d-flex align-center mb-3">
                <v-icon color="secondary" size="32" class="mr-3">mdi-account-multiple</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">User Management</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Manage user accounts</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                View, edit, and manage user accounts and permissions.
              </p>
              <v-chip size="small" color="secondary" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- Organization Management -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/organizations')">
              <div class="d-flex align-center mb-3">
                <v-icon color="accent" size="32" class="mr-3">mdi-domain</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">Organizations</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Manage gyms and organizations</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                Create and manage organizations, assign users to gyms and groups.
              </p>
              <v-chip size="small" color="accent" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- User Content Management -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/user-content')">
              <div class="d-flex align-center mb-3">
                <v-icon color="error" size="32" class="mr-3">mdi-file-multiple</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">User Content</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Manage user-created content</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                View user-created WODs, movements, and workout templates. Promote them to standard items.
              </p>
              <v-chip size="small" color="error" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- Database Backups -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/backups')">
              <div class="d-flex align-center mb-3">
                <v-icon color="warning" size="32" class="mr-3">mdi-database-export</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">Database Backups</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Create and restore database backups</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                Protect your data with full database backups. Create, download, and restore backups.
              </p>
              <v-chip size="small" color="warning" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- Audit Logs -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/audit-logs')">
              <div class="d-flex align-center mb-3">
                <v-icon color="success" size="32" class="mr-3">mdi-shield-check</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">Audit Logs</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">View security events</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                Monitor login attempts, password changes, and other security-related events.
              </p>
              <v-chip size="small" color="success" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- Data Change Logs -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/data-change-logs')">
              <div class="d-flex align-center mb-3">
                <v-icon color="info" size="32" class="mr-3">mdi-history</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">Data Change Logs</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Track edits and deletions</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                View history of all data changes including before/after values for WODs, movements, and workouts.
              </p>
              <v-chip size="small" color="info" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- Announcements -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/announcements')">
              <div class="d-flex align-center mb-3">
                <v-icon color="warning" size="32" class="mr-3">mdi-bullhorn</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">Announcements</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Send notifications to users</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                Send announcements to all users, specific organizations, or selected users with markdown support.
              </p>
              <v-chip size="small" color="warning" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- Subscriptions -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/subscriptions')">
              <div class="d-flex align-center mb-3">
                <v-icon color="primary" size="32" class="mr-3">mdi-currency-usd</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">Subscriptions</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Manage user and organization subscriptions</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                View and manage all user and organization subscriptions, payments, and billing.
              </p>
              <v-chip size="small" color="primary" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- Email Settings -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/email-settings')">
              <div class="d-flex align-center mb-3">
                <v-icon color="indigo" size="32" class="mr-3">mdi-email-outline</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">Email Settings</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Test email configuration and view logs</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                Test email configuration with detailed SMTP debug output. View history of sent emails.
              </p>
              <v-chip size="small" color="indigo" class="mt-2">
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
      <v-btn to="/admin" style="flex: 1" color="primary">
        <v-icon size="20">mdi-shield-account</v-icon>
        <span style="font-size: 9px">Admin</span>
      </v-btn>
    </v-bottom-navigation>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const navigateTo = (path) => {
  router.push(path)
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
