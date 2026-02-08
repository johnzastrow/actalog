<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <AdminHeader
        title="Email Settings"
        subtitle="Test email configuration and view email logs"
        :breadcrumbs="[{ title: 'Email Settings', to: '/admin/email-settings' }]"
      />

      <div class="d-flex justify-end mb-4">
        <v-btn
          color="secondary"
          prepend-icon="mdi-email-search"
          :to="{ name: 'admin-email-logs' }"
        >
          View Email Logs
        </v-btn>
      </div>

      <!-- Loading State -->
      <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />

      <!-- Error Alert -->
      <v-alert v-if="error" type="error" variant="tonal" closable class="mb-4" @click:close="error = null">
        {{ error }}
      </v-alert>

      <!-- Success Alert -->
      <v-alert v-if="successMessage" type="success" variant="tonal" closable class="mb-4" @click:close="successMessage = null">
        {{ successMessage }}
      </v-alert>

      <!-- Current Configuration Card -->
      <v-card elevation="0" rounded="lg" class="mb-4">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-cog</v-icon>
          Current Configuration
        </v-card-title>
        <v-divider />
        <v-card-text>
          <v-row>
            <v-col cols="12" md="6">
              <div class="config-item">
                <div class="config-label">SMTP Host</div>
                <div class="config-value">{{ config.smtp_host || 'Not configured' }}</div>
              </div>
            </v-col>
            <v-col cols="12" md="6">
              <div class="config-item">
                <div class="config-label">SMTP Port</div>
                <div class="config-value">{{ config.smtp_port || 'Not configured' }}</div>
              </div>
            </v-col>
            <v-col cols="12" md="6">
              <div class="config-item">
                <div class="config-label">From Address</div>
                <div class="config-value">{{ config.from_name ? `${config.from_name} <${config.from_address}>` : config.from_address || 'Not configured' }}</div>
              </div>
            </v-col>
            <v-col cols="12" md="6">
              <div class="config-item">
                <div class="config-label">Auth User</div>
                <div class="config-value">{{ config.smtp_user || 'Not configured' }}</div>
              </div>
            </v-col>
            <v-col cols="12" md="6">
              <div class="config-item">
                <div class="config-label">TLS Mode</div>
                <div class="config-value">{{ config.tls_mode || 'Auto' }}</div>
              </div>
            </v-col>
            <v-col cols="12" md="6">
              <div class="config-item">
                <div class="config-label">Status</div>
                <v-chip
                  :color="config.enabled ? 'success' : 'error'"
                  size="small"
                  variant="flat"
                >
                  {{ config.enabled ? 'Enabled' : 'Disabled' }}
                </v-chip>
              </div>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- Test Email Card -->
      <v-card elevation="0" rounded="lg" class="mb-4">
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-email-check</v-icon>
          Send Test Email
        </v-card-title>
        <v-divider />
        <v-card-text>
          <v-form @submit.prevent="sendTestEmail">
            <v-text-field
              v-model="testEmail"
              label="Recipient Email"
              type="email"
              placeholder="Enter email address to test"
              prepend-inner-icon="mdi-email"
              
              density="comfortable"
              :disabled="sending || !config.enabled"
              :rules="[v => !!v || 'Email is required', v => /.+@.+\..+/.test(v) || 'Invalid email format']"
              class="mb-4"
            />
            <v-btn
              type="submit"
              color="primary"
              prepend-icon="mdi-send"
              :loading="sending"
              :disabled="sending || !testEmail || !config.enabled"
              block
            >
              Send Test Email
            </v-btn>
          </v-form>
        </v-card-text>
      </v-card>

      <!-- Debug Output Card -->
      <v-card v-if="debugInfo" elevation="0" rounded="lg">
        <v-card-title class="d-flex align-center">
          <v-icon :color="debugInfo.final_error ? 'error' : 'success'" class="mr-2">
            {{ debugInfo.final_error ? 'mdi-close-circle' : 'mdi-check-circle' }}
          </v-icon>
          Debug Output
          <v-spacer />
          <v-chip
            :color="debugInfo.final_error ? 'error' : 'success'"
            size="small"
            variant="flat"
          >
            {{ debugInfo.final_error ? 'FAILED' : 'SUCCESS' }} - {{ debugInfo.total_duration }}
          </v-chip>
        </v-card-title>
        <v-divider />
        <v-card-text>
          <!-- Connection Phase -->
          <div class="debug-phase mb-4">
            <div class="phase-header d-flex align-center">
              <v-icon
                size="small"
                :color="debugInfo.connection_error ? 'error' : 'success'"
                class="mr-2"
              >
                {{ debugInfo.connection_error ? 'mdi-close-circle' : 'mdi-check-circle' }}
              </v-icon>
              <span class="font-weight-bold">Connection Phase</span>
              <v-spacer />
              <span class="text-caption text-medium-emphasis">{{ debugInfo.connection_time }}</span>
            </div>
            <div class="phase-details ml-6">
              <div><span class="detail-label">Host:</span> {{ debugInfo.connection_host }}:{{ debugInfo.connection_port }}</div>
              <div><span class="detail-label">Mode:</span> {{ debugInfo.connection_tls }}</div>
              <div v-if="debugInfo.connection_error" class="text-error">
                <span class="detail-label">Error:</span> {{ debugInfo.connection_error }}
              </div>
            </div>
          </div>

          <!-- Authentication Phase -->
          <div class="debug-phase mb-4">
            <div class="phase-header d-flex align-center">
              <v-icon
                size="small"
                :color="debugInfo.auth_error ? 'error' : 'success'"
                class="mr-2"
              >
                {{ debugInfo.auth_error ? 'mdi-close-circle' : 'mdi-check-circle' }}
              </v-icon>
              <span class="font-weight-bold">Authentication Phase</span>
              <v-spacer />
              <span class="text-caption text-medium-emphasis">{{ debugInfo.auth_time }}</span>
            </div>
            <div class="phase-details ml-6">
              <div><span class="detail-label">Method:</span> {{ debugInfo.auth_method || 'PLAIN' }}</div>
              <div v-if="debugInfo.auth_error" class="text-error">
                <span class="detail-label">Error:</span> {{ debugInfo.auth_error }}
              </div>
            </div>
          </div>

          <!-- Send Phase -->
          <div class="debug-phase mb-4">
            <div class="phase-header d-flex align-center">
              <v-icon
                size="small"
                :color="debugInfo.send_error ? 'error' : 'success'"
                class="mr-2"
              >
                {{ debugInfo.send_error ? 'mdi-close-circle' : 'mdi-check-circle' }}
              </v-icon>
              <span class="font-weight-bold">Send Phase</span>
              <v-spacer />
              <span class="text-caption text-medium-emphasis">{{ debugInfo.send_time }}</span>
            </div>
            <div class="phase-details ml-6">
              <div v-if="debugInfo.send_error" class="text-error">
                <span class="detail-label">Error:</span> {{ debugInfo.send_error }}
              </div>
              <div v-else class="text-success">Message accepted for delivery</div>
            </div>
          </div>

          <!-- SMTP Responses -->
          <div v-if="debugInfo.smtp_responses && debugInfo.smtp_responses.length > 0" class="debug-phase">
            <div class="phase-header d-flex align-center mb-2">
              <v-icon size="small" class="mr-2">mdi-console</v-icon>
              <span class="font-weight-bold">SMTP Responses</span>
            </div>
            <v-sheet
              class="pa-3 rounded"
              color="grey-darken-4"
            >
              <code class="smtp-responses">
                <div v-for="(response, index) in debugInfo.smtp_responses" :key="index" class="smtp-line">
                  {{ response }}
                </div>
              </code>
            </v-sheet>
          </div>

          <!-- Final Error -->
          <v-alert v-if="debugInfo.final_error" type="error" variant="tonal" class="mt-4">
            <strong>Final Error:</strong> {{ debugInfo.final_error }}
          </v-alert>
        </v-card-text>
      </v-card>
    </v-container>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'
import AdminHeader from '@/components/AdminHeader.vue'

// State
const loading = ref(false)
const sending = ref(false)
const error = ref(null)
const successMessage = ref(null)
const config = ref({})
const testEmail = ref('')
const debugInfo = ref(null)

// Fetch email configuration
const fetchConfig = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await axios.get('/api/admin/email/config')
    config.value = response.data
  } catch (err) {
    console.error('Failed to fetch email config:', err)
    error.value = err.response?.data?.error || 'Failed to load email configuration'
  } finally {
    loading.value = false
  }
}

// Send test email
const sendTestEmail = async () => {
  if (!testEmail.value) return

  sending.value = true
  error.value = null
  successMessage.value = null
  debugInfo.value = null

  try {
    const response = await axios.post('/api/admin/email/test', {
      recipient_email: testEmail.value
    })

    debugInfo.value = response.data.debug_info

    if (response.data.success) {
      successMessage.value = response.data.message || 'Test email sent successfully!'
    } else {
      error.value = response.data.message || 'Failed to send test email'
    }
  } catch (err) {
    console.error('Failed to send test email:', err)
    error.value = err.response?.data?.error || 'Failed to send test email'
  } finally {
    sending.value = false
  }
}

// Fetch config on mount
onMounted(() => {
  fetchConfig()
})
</script>

<style scoped>
.config-item {
  margin-bottom: 8px;
}

.config-label {
  font-size: 0.75rem;
  color: rgba(var(--v-theme-on-surface), 0.6);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.config-value {
  font-size: 1rem;
  font-weight: 500;
}

.debug-phase {
  border-left: 2px solid rgba(var(--v-theme-on-surface), 0.12);
  padding-left: 12px;
}

.phase-header {
  margin-bottom: 8px;
}

.phase-details {
  font-size: 0.875rem;
}

.detail-label {
  font-weight: 500;
  margin-right: 4px;
}

.smtp-responses {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.75rem;
  line-height: 1.4;
  color: #4caf50;
  display: block;
}

.smtp-line {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
