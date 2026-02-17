<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
        <!-- Admin Header with User Info -->
        <AdminHeader
          title="Administration"
          subtitle="Manage system data and configurations"
          :breadcrumbs="[]"
        />

        <!-- Admin Tools Grid -->
        <v-row>
          <!-- Data Cleanup Tool -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" @click="navigateTo('/admin/data-cleanup')">
              <div class="d-flex align-center mb-3">
                <v-icon color="primary" size="32" class="mr-3">mdi-database-refresh</v-icon>
                <div>
                  <h3 class="text-h6 font-weight-bold">Data Cleanup</h3>
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

          <!-- Data Quality Checks -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/data-quality')">
              <div class="d-flex align-center mb-3">
                <v-icon color="teal" size="32" class="mr-3">mdi-database-check</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">Data Quality Checks</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Duplicate detection and data issues</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                Scan for duplicate records and data quality issues. Merge duplicates safely with FK updates.
              </p>
              <v-chip size="small" color="teal" class="mt-2">
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

          <!-- Gym Management -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/organizations')">
              <div class="d-flex align-center mb-3">
                <v-icon color="accent" size="32" class="mr-3">mdi-domain</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">Gyms</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Manage gyms, boxes, and affiliates</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                Create and manage gyms, assign users, and configure locations.
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
                Send announcements to all users, specific gyms, or selected users with markdown support.
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
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Manage user and gym subscriptions</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                View and manage all user and gym subscriptions, payments, and billing.
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

          <!-- System Metrics -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/metrics')">
              <div class="d-flex align-center mb-3">
                <v-icon color="teal" size="32" class="mr-3">mdi-chart-bar</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">System Metrics</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">View real-time statistics</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                Dashboard with user activity, workout trends, content stats, and system health.
              </p>
              <v-chip size="small" color="teal" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>

          <!-- User Import/Export -->
          <v-col cols="12" md="6">
            <v-card elevation="0" rounded="lg" class="pa-4" hover style="cursor: pointer" @click="$router.push('/admin/users/import-export')">
              <div class="d-flex align-center mb-3">
                <v-icon color="deep-purple" size="32" class="mr-3">mdi-account-arrow-left-outline</v-icon>
                <div>
                  <h3 style="color: rgb(var(--v-theme-on-surface)); font-size: 18px; font-weight: 600">User Import/Export</h3>
                  <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 14px; margin: 0">Bulk user management</p>
                </div>
              </div>
              <v-divider class="mb-3"></v-divider>
              <p style="color: rgb(var(--v-theme-on-surface), 0.6); font-size: 13px">
                Import users from CSV, export user lists, and send batch password reset emails.
              </p>
              <v-chip size="small" color="deep-purple" class="mt-2">
                <v-icon start size="small">mdi-chevron-right</v-icon>
                Open Tool
              </v-chip>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import AdminHeader from '@/components/AdminHeader.vue'

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
