<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <h1 class="text-h5 font-weight-bold mb-4">Subscription Management</h1>

      <!-- Tabs for User vs Organization subscriptions -->
      <v-tabs v-model="activeTab" color="primary" class="mb-4">
        <v-tab value="users">User Subscriptions</v-tab>
        <v-tab value="organizations">Organization Subscriptions</v-tab>
        <v-tab value="expiring">Expiring Soon</v-tab>
        <v-tab value="overdue">Overdue</v-tab>
      </v-tabs>

      <v-window v-model="activeTab">
        <!-- User Subscriptions Tab -->
        <v-window-item value="users">
          <v-card elevation="0" rounded="lg">
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2">mdi-account-group</v-icon>
              User Subscriptions
              <v-spacer></v-spacer>
              <v-btn
                color="primary"
                prepend-icon="mdi-plus"
                @click="openCreateUserSubscriptionDialog"
              >
                Create Subscription
              </v-btn>
            </v-card-title>

            <v-card-text>
              <!-- Filters -->
              <v-row class="mb-3">
                <v-col cols="12" md="4">
                  <v-text-field
                    v-model="userFilters.search"
                    label="Search by email"
                    prepend-inner-icon="mdi-magnify"
                    
                    density="compact"
                    clearable
                  ></v-text-field>
                </v-col>
                <v-col cols="12" md="4">
                  <v-select
                    v-model="userFilters.status"
                    label="Status"
                    :items="statusOptions"
                    
                    density="compact"
                    clearable
                  ></v-select>
                </v-col>
                <v-col cols="12" md="4">
                  <v-select
                    v-model="userFilters.type"
                    label="Type"
                    :items="typeOptions"
                    
                    density="compact"
                    clearable
                  ></v-select>
                </v-col>
              </v-row>

              <!-- Data Table -->
              <v-data-table
                :headers="userSubscriptionHeaders"
                :items="filteredUserSubscriptions"
                :loading="loading"
                :items-per-page="10"
                class="elevation-1"
              >
                <template #item.user="{ item }">
                  {{ item.user_email || `User ${item.user_id}` }}
                </template>

                <template #item.subscription_type="{ item }">
                  <v-chip
                    :color="getTypeColor(item.subscription_type)"
                    size="small"
                    label
                  >
                    {{ item.is_permanent_free ? 'Permanent Free' : item.subscription_type }}
                  </v-chip>
                </template>

                <template #item.status="{ item }">
                  <v-chip
                    :color="getStatusColor(item.status)"
                    size="small"
                  >
                    {{ item.status }}
                  </v-chip>
                </template>

                <template #item.end_date="{ item }">
                  {{ item.is_permanent_free ? 'Never' : formatDate(item.end_date) }}
                </template>

                <template #item.permanent="{ item }">
                  <v-tooltip :text="item.is_permanent_free ? 'Remove permanent free status' : 'Set as permanent free'" location="top">
                    <template #activator="{ props }">
                      <v-btn
                        v-bind="props"
                        :icon="item.is_permanent_free ? 'mdi-star' : 'mdi-star-outline'"
                        size="small"
                        variant="text"
                        :color="item.is_permanent_free ? 'orange' : 'grey'"
                        @click="togglePermanent(item, 'user')"
                      ></v-btn>
                    </template>
                  </v-tooltip>
                </template>

                <template #item.view="{ item }">
                  <v-tooltip text="View subscription details" location="top">
                    <template #activator="{ props }">
                      <v-btn
                        v-bind="props"
                        icon="mdi-eye"
                        size="small"
                        variant="text"
                        @click="viewSubscriptionDetails(item, 'user')"
                      ></v-btn>
                    </template>
                  </v-tooltip>
                </template>

                <template #item.mark_paid="{ item }">
                  <v-tooltip text="Mark this subscription as paid and extend end date" location="top">
                    <template #activator="{ props }">
                      <v-btn
                        v-if="item.status === 'active' && !item.is_permanent_free"
                        v-bind="props"
                        icon="mdi-cash"
                        size="small"
                        variant="text"
                        color="success"
                        @click="openMarkAsPaidDialog(item, 'user')"
                      ></v-btn>
                      <span v-else class="text-grey">—</span>
                    </template>
                  </v-tooltip>
                </template>

                <template #item.cancel="{ item }">
                  <v-tooltip text="Cancel this subscription" location="top">
                    <template #activator="{ props }">
                      <v-btn
                        v-if="item.status === 'active'"
                        v-bind="props"
                        icon="mdi-cancel"
                        size="small"
                        variant="text"
                        color="error"
                        @click="openCancelDialog(item, 'user')"
                      ></v-btn>
                      <span v-else class="text-grey">—</span>
                    </template>
                  </v-tooltip>
                </template>
              </v-data-table>
            </v-card-text>
          </v-card>
        </v-window-item>

        <!-- Organization Subscriptions Tab -->
        <v-window-item value="organizations">
          <v-card elevation="0" rounded="lg">
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2">mdi-office-building</v-icon>
              Organization Subscriptions
              <v-spacer></v-spacer>
              <v-btn
                color="primary"
                prepend-icon="mdi-plus"
                @click="openCreateOrgSubscriptionDialog"
              >
                Create Subscription
              </v-btn>
            </v-card-title>

            <v-card-text>
              <!-- Filters (same as user subscriptions) -->
              <v-row class="mb-3">
                <v-col cols="12" md="4">
                  <v-text-field
                    v-model="orgFilters.search"
                    label="Search by organization name"
                    prepend-inner-icon="mdi-magnify"
                    
                    density="compact"
                    clearable
                  ></v-text-field>
                </v-col>
                <v-col cols="12" md="4">
                  <v-select
                    v-model="orgFilters.status"
                    label="Status"
                    :items="statusOptions"
                    
                    density="compact"
                    clearable
                  ></v-select>
                </v-col>
                <v-col cols="12" md="4">
                  <v-select
                    v-model="orgFilters.type"
                    label="Type"
                    :items="typeOptions"
                    
                    density="compact"
                    clearable
                  ></v-select>
                </v-col>
              </v-row>

              <!-- Data Table -->
              <v-data-table
                :headers="orgSubscriptionHeaders"
                :items="filteredOrgSubscriptions"
                :loading="loading"
                :items-per-page="10"
                class="elevation-1"
              >
                <template #item.organization="{ item }">
                  {{ item.organization_name || `Organization ${item.organization_id}` }}
                </template>

                <template #item.subscription_type="{ item }">
                  <v-chip
                    :color="getTypeColor(item.subscription_type)"
                    size="small"
                    label
                  >
                    {{ item.is_permanent_free ? 'Permanent Free' : item.subscription_type }}
                  </v-chip>
                </template>

                <template #item.status="{ item }">
                  <v-chip
                    :color="getStatusColor(item.status)"
                    size="small"
                  >
                    {{ item.status }}
                  </v-chip>
                </template>

                <template #item.end_date="{ item }">
                  {{ item.is_permanent_free ? 'Never' : formatDate(item.end_date) }}
                </template>

                <template #item.permanent="{ item }">
                  <v-tooltip :text="item.is_permanent_free ? 'Remove permanent free status' : 'Set as permanent free'" location="top">
                    <template #activator="{ props }">
                      <v-btn
                        v-bind="props"
                        :icon="item.is_permanent_free ? 'mdi-star' : 'mdi-star-outline'"
                        size="small"
                        variant="text"
                        :color="item.is_permanent_free ? 'orange' : 'grey'"
                        @click="togglePermanent(item, 'organization')"
                      ></v-btn>
                    </template>
                  </v-tooltip>
                </template>

                <template #item.view="{ item }">
                  <v-tooltip text="View subscription details" location="top">
                    <template #activator="{ props }">
                      <v-btn
                        v-bind="props"
                        icon="mdi-eye"
                        size="small"
                        variant="text"
                        @click="viewSubscriptionDetails(item, 'organization')"
                      ></v-btn>
                    </template>
                  </v-tooltip>
                </template>

                <template #item.mark_paid="{ item }">
                  <v-tooltip text="Mark this subscription as paid and extend end date" location="top">
                    <template #activator="{ props }">
                      <v-btn
                        v-if="item.status === 'active' && !item.is_permanent_free"
                        v-bind="props"
                        icon="mdi-cash"
                        size="small"
                        variant="text"
                        color="success"
                        @click="openMarkAsPaidDialog(item, 'organization')"
                      ></v-btn>
                      <span v-else class="text-grey">—</span>
                    </template>
                  </v-tooltip>
                </template>

                <template #item.cancel="{ item }">
                  <v-tooltip text="Cancel this subscription" location="top">
                    <template #activator="{ props }">
                      <v-btn
                        v-if="item.status === 'active'"
                        v-bind="props"
                        icon="mdi-cancel"
                        size="small"
                        variant="text"
                        color="error"
                        @click="openCancelDialog(item, 'organization')"
                      ></v-btn>
                      <span v-else class="text-grey">—</span>
                    </template>
                  </v-tooltip>
                </template>
              </v-data-table>
            </v-card-text>
          </v-card>
        </v-window-item>

        <!-- Expiring Soon Tab -->
        <v-window-item value="expiring">
          <v-card elevation="0" rounded="lg">
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2" color="warning">mdi-clock-alert</v-icon>
              Expiring Soon
              <v-spacer></v-spacer>
              <v-select
                v-model="expiringDays"
                :items="expiringDaysOptions"
                label="Days ahead"
                
                density="compact"
                style="max-width: 150px;"
                @update:model-value="loadExpiringExpired"
              ></v-select>
            </v-card-title>
            <v-card-text>
              <v-progress-linear v-if="loadingExpiring" indeterminate color="warning"></v-progress-linear>

              <!-- User subscriptions expiring -->
              <div v-if="expiringUserSubscriptions.length > 0" class="mb-4">
                <h3 class="text-subtitle-1 mb-2">User Subscriptions ({{ expiringUserSubscriptions.length }})</h3>
                <v-list density="compact">
                  <v-list-item
                    v-for="sub in expiringUserSubscriptions"
                    :key="'user-' + sub.id"
                    class="mb-1"
                  >
                    <template #prepend>
                      <v-icon :color="getDaysLeftColor(calculateDaysLeft(sub.end_date))" size="small">
                        mdi-account
                      </v-icon>
                    </template>
                    <v-list-item-title>
                      {{ sub.user_email || `User ${sub.user_id}` }}
                    </v-list-item-title>
                    <v-list-item-subtitle>
                      {{ sub.subscription_type }} - Expires {{ formatDate(sub.end_date) }}
                      ({{ calculateDaysLeft(sub.end_date) }} days)
                    </v-list-item-subtitle>
                    <template #append>
                      <v-btn
                        size="small"
                        color="primary"
                        variant="tonal"
                        @click="openMarkAsPaidDialog(sub, 'user')"
                      >
                        Mark Paid
                      </v-btn>
                    </template>
                  </v-list-item>
                </v-list>
              </div>

              <!-- Organization subscriptions expiring -->
              <div v-if="expiringOrgSubscriptions.length > 0">
                <h3 class="text-subtitle-1 mb-2">Organization Subscriptions ({{ expiringOrgSubscriptions.length }})</h3>
                <v-list density="compact">
                  <v-list-item
                    v-for="sub in expiringOrgSubscriptions"
                    :key="'org-' + sub.id"
                    class="mb-1"
                  >
                    <template #prepend>
                      <v-icon :color="getDaysLeftColor(calculateDaysLeft(sub.end_date))" size="small">
                        mdi-office-building
                      </v-icon>
                    </template>
                    <v-list-item-title>
                      {{ sub.organization_name || `Organization ${sub.organization_id}` }}
                    </v-list-item-title>
                    <v-list-item-subtitle>
                      {{ sub.subscription_type }} - Expires {{ formatDate(sub.end_date) }}
                      ({{ calculateDaysLeft(sub.end_date) }} days)
                    </v-list-item-subtitle>
                    <template #append>
                      <v-btn
                        size="small"
                        color="primary"
                        variant="tonal"
                        @click="openMarkAsPaidDialog(sub, 'organization')"
                      >
                        Mark Paid
                      </v-btn>
                    </template>
                  </v-list-item>
                </v-list>
              </div>

              <v-alert
                v-if="!loadingExpiring && expiringUserSubscriptions.length === 0 && expiringOrgSubscriptions.length === 0"
                type="success"
                variant="tonal"
              >
                No subscriptions expiring in the next {{ expiringDays }} days
              </v-alert>
            </v-card-text>
          </v-card>
        </v-window-item>

        <!-- Overdue Tab -->
        <v-window-item value="overdue">
          <v-card elevation="0" rounded="lg">
            <v-card-title>
              <v-icon class="mr-2" color="error">mdi-alert</v-icon>
              Expired Subscriptions
            </v-card-title>
            <v-card-text>
              <v-progress-linear v-if="loadingExpired" indeterminate color="error"></v-progress-linear>

              <!-- User subscriptions expired -->
              <div v-if="expiredUserSubscriptions.length > 0" class="mb-4">
                <h3 class="text-subtitle-1 mb-2">User Subscriptions ({{ expiredUserSubscriptions.length }})</h3>
                <v-list density="compact">
                  <v-list-item
                    v-for="sub in expiredUserSubscriptions"
                    :key="'user-' + sub.id"
                    class="mb-1"
                  >
                    <template #prepend>
                      <v-icon color="error" size="small">mdi-account</v-icon>
                    </template>
                    <v-list-item-title>
                      {{ sub.user_email || `User ${sub.user_id}` }}
                    </v-list-item-title>
                    <v-list-item-subtitle>
                      {{ sub.subscription_type }} - Expired {{ formatDate(sub.end_date) }}
                      <v-chip size="x-small" color="grey" class="ml-1">{{ sub.status }}</v-chip>
                    </v-list-item-subtitle>
                    <template #append>
                      <v-btn
                        size="small"
                        color="error"
                        variant="tonal"
                        @click="openMarkAsPaidDialog(sub, 'user')"
                      >
                        Renew
                      </v-btn>
                    </template>
                  </v-list-item>
                </v-list>
              </div>

              <!-- Organization subscriptions expired -->
              <div v-if="expiredOrgSubscriptions.length > 0">
                <h3 class="text-subtitle-1 mb-2">Organization Subscriptions ({{ expiredOrgSubscriptions.length }})</h3>
                <v-list density="compact">
                  <v-list-item
                    v-for="sub in expiredOrgSubscriptions"
                    :key="'org-' + sub.id"
                    class="mb-1"
                  >
                    <template #prepend>
                      <v-icon color="error" size="small">mdi-office-building</v-icon>
                    </template>
                    <v-list-item-title>
                      {{ sub.organization_name || `Organization ${sub.organization_id}` }}
                    </v-list-item-title>
                    <v-list-item-subtitle>
                      {{ sub.subscription_type }} - Expired {{ formatDate(sub.end_date) }}
                      <v-chip size="x-small" color="grey" class="ml-1">{{ sub.status }}</v-chip>
                    </v-list-item-subtitle>
                    <template #append>
                      <v-btn
                        size="small"
                        color="error"
                        variant="tonal"
                        @click="openMarkAsPaidDialog(sub, 'organization')"
                      >
                        Renew
                      </v-btn>
                    </template>
                  </v-list-item>
                </v-list>
              </div>

              <v-alert
                v-if="!loadingExpired && expiredUserSubscriptions.length === 0 && expiredOrgSubscriptions.length === 0"
                type="success"
                variant="tonal"
              >
                No expired subscriptions
              </v-alert>
            </v-card-text>
          </v-card>
        </v-window-item>
      </v-window>

      <!-- Dialogs -->
      <create-subscription-dialog
        v-model="createDialog.show"
        :type="createDialog.type"
        @created="handleSubscriptionCreated"
      />

      <subscription-detail-dialog
        v-model="detailDialog.show"
        :subscription="detailDialog.subscription"
        :type="detailDialog.type"
        @refresh="loadSubscriptions"
      />

      <MarkAsPaidDialog
        v-model="markAsPaidDialog.show"
        :subscription="markAsPaidDialog.subscription"
        :type="markAsPaidDialog.type"
        @paid="handleMarkedAsPaid"
      />

      <cancel-subscription-dialog
        v-model="cancelDialog.show"
        :subscription="cancelDialog.subscription"
        :type="cancelDialog.type"
        @cancelled="handleCancelled"
      />
    </v-container>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import axios from '@/utils/axios'
import CreateSubscriptionDialog from '@/components/admin/CreateSubscriptionDialog.vue'
import SubscriptionDetailDialog from '@/components/admin/SubscriptionDetailDialog.vue'
import MarkAsPaidDialog from '@/components/admin/MarkAsPaidDialog.vue'
import CancelSubscriptionDialog from '@/components/admin/CancelSubscriptionDialog.vue'

const activeTab = ref('users')
const loading = ref(false)
const loadingExpiring = ref(false)
const loadingExpired = ref(false)

// Data
const userSubscriptions = ref([])
const orgSubscriptions = ref([])

// Expiring/Expired data from API
const expiringUserSubscriptions = ref([])
const expiringOrgSubscriptions = ref([])
const expiredUserSubscriptions = ref([])
const expiredOrgSubscriptions = ref([])

// Days filter for expiring subscriptions
const expiringDays = ref(30)
const expiringDaysOptions = [
  { title: '7 days', value: 7 },
  { title: '14 days', value: 14 },
  { title: '30 days', value: 30 },
  { title: '60 days', value: 60 },
  { title: '90 days', value: 90 }
]

// Filters
const userFilters = ref({
  search: '',
  status: null,
  type: null
})

const orgFilters = ref({
  search: '',
  status: null,
  type: null
})

// Filter options
const statusOptions = [
  { title: 'All', value: null },
  { title: 'Active', value: 'active' },
  { title: 'Expired', value: 'expired' },
  { title: 'Cancelled', value: 'cancelled' }
]

const typeOptions = [
  { title: 'All', value: null },
  { title: 'Free', value: 'free' },
  { title: 'Monthly', value: 'monthly' },
  { title: 'Annual', value: 'annual' }
]

// Table headers
const userSubscriptionHeaders = [
  { title: 'User', key: 'user', sortable: true },
  { title: 'Type', key: 'subscription_type', sortable: true },
  { title: 'Status', key: 'status', sortable: true },
  { title: 'Start Date', key: 'start_date', sortable: true },
  { title: 'End Date', key: 'end_date', sortable: true },
  { title: 'Permanent', key: 'permanent', sortable: false, align: 'center', width: '100px' },
  { title: 'View', key: 'view', sortable: false, align: 'center', width: '80px' },
  { title: 'Mark Paid', key: 'mark_paid', sortable: false, align: 'center', width: '100px' },
  { title: 'Cancel', key: 'cancel', sortable: false, align: 'center', width: '80px' }
]

const orgSubscriptionHeaders = [
  { title: 'Organization', key: 'organization', sortable: true },
  { title: 'Type', key: 'subscription_type', sortable: true },
  { title: 'Status', key: 'status', sortable: true },
  { title: 'Start Date', key: 'start_date', sortable: true },
  { title: 'End Date', key: 'end_date', sortable: true },
  { title: 'Permanent', key: 'permanent', sortable: false, align: 'center', width: '100px' },
  { title: 'View', key: 'view', sortable: false, align: 'center', width: '80px' },
  { title: 'Mark Paid', key: 'mark_paid', sortable: false, align: 'center', width: '100px' },
  { title: 'Cancel', key: 'cancel', sortable: false, align: 'center', width: '80px' }
]

// Dialog states
const createDialog = ref({
  show: false,
  type: 'user' // 'user' or 'organization'
})

const detailDialog = ref({
  show: false,
  subscription: null,
  type: 'user'
})

const markAsPaidDialog = ref({
  show: false,
  subscription: null,
  type: 'user'
})

const cancelDialog = ref({
  show: false,
  subscription: null,
  type: 'user'
})

// Computed - Filtered subscriptions
const filteredUserSubscriptions = computed(() => {
  let filtered = userSubscriptions.value

  if (userFilters.value.search) {
    const search = userFilters.value.search.toLowerCase()
    filtered = filtered.filter(sub =>
      (sub.user_email || '').toLowerCase().includes(search)
    )
  }

  if (userFilters.value.status) {
    filtered = filtered.filter(sub => sub.status === userFilters.value.status)
  }

  if (userFilters.value.type) {
    filtered = filtered.filter(sub => sub.subscription_type === userFilters.value.type)
  }

  return filtered
})

const filteredOrgSubscriptions = computed(() => {
  let filtered = orgSubscriptions.value

  if (orgFilters.value.search) {
    const search = orgFilters.value.search.toLowerCase()
    filtered = filtered.filter(sub =>
      (sub.organization_name || '').toLowerCase().includes(search)
    )
  }

  if (orgFilters.value.status) {
    filtered = filtered.filter(sub => sub.status === orgFilters.value.status)
  }

  if (orgFilters.value.type) {
    filtered = filtered.filter(sub => sub.subscription_type === orgFilters.value.type)
  }

  return filtered
})

// Helper function to calculate days left
function calculateDaysLeft(endDate) {
  if (!endDate) return 0
  const now = new Date()
  const end = new Date(endDate)
  return Math.ceil((end - now) / (1000 * 60 * 60 * 24))
}

// Methods
async function loadSubscriptions() {
  loading.value = true
  try {
    // Load user subscriptions
    const userResponse = await axios.get('/api/admin/subscriptions/users')
    userSubscriptions.value = userResponse.data.subscriptions || []

    // Load organization subscriptions
    const orgResponse = await axios.get('/api/admin/subscriptions/organizations')
    orgSubscriptions.value = orgResponse.data.subscriptions || []
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
  } finally {
    loading.value = false
  }
}

async function loadExpiringExpired() {
  // Load expiring subscriptions
  loadingExpiring.value = true
  loadingExpired.value = true

  try {
    const [expiringUsersRes, expiringOrgsRes, expiredUsersRes, expiredOrgsRes] = await Promise.all([
      axios.get(`/api/admin/subscriptions/users/expiring?days=${expiringDays.value}`),
      axios.get(`/api/admin/subscriptions/organizations/expiring?days=${expiringDays.value}`),
      axios.get('/api/admin/subscriptions/users/expired'),
      axios.get('/api/admin/subscriptions/organizations/expired')
    ])

    expiringUserSubscriptions.value = expiringUsersRes.data.subscriptions || []
    expiringOrgSubscriptions.value = expiringOrgsRes.data.subscriptions || []
    expiredUserSubscriptions.value = expiredUsersRes.data.subscriptions || []
    expiredOrgSubscriptions.value = expiredOrgsRes.data.subscriptions || []
  } catch (error) {
    console.error('Failed to load expiring/expired subscriptions:', error)
  } finally {
    loadingExpiring.value = false
    loadingExpired.value = false
  }
}

function openCreateUserSubscriptionDialog() {
  createDialog.value = {
    show: true,
    type: 'user'
  }
}

function openCreateOrgSubscriptionDialog() {
  createDialog.value = {
    show: true,
    type: 'organization'
  }
}

function viewSubscriptionDetails(subscription, type) {
  detailDialog.value = {
    show: true,
    subscription,
    type
  }
}

function openMarkAsPaidDialog(subscription, type) {
  markAsPaidDialog.value = {
    show: true,
    subscription,
    type
  }
}

function openCancelDialog(subscription, type) {
  cancelDialog.value = {
    show: true,
    subscription,
    type
  }
}

function handleSubscriptionCreated() {
  loadSubscriptions()
  loadExpiringExpired()
}

function handleMarkedAsPaid() {
  loadSubscriptions()
  loadExpiringExpired()
}

function handleCancelled() {
  loadSubscriptions()
  loadExpiringExpired()
}

async function togglePermanent(subscription, type) {
  const action = subscription.is_permanent_free ? 'remove permanent free status from' : 'set as permanent free'
  const confirmMessage = `Are you sure you want to ${action} this subscription?`

  if (!confirm(confirmMessage)) {
    return
  }

  loading.value = true
  try {
    const endpoint = type === 'user'
      ? `/api/admin/subscriptions/user/${subscription.id}/set-permanent`
      : `/api/admin/subscriptions/organization/${subscription.id}/set-permanent`

    await axios.post(endpoint, {
      is_permanent: !subscription.is_permanent_free
    })

    // Reload subscriptions to reflect changes
    await loadSubscriptions()
  } catch (error) {
    console.error('Failed to toggle permanent status:', error)
    alert(error.response?.data?.error || 'Failed to toggle permanent status')
  } finally {
    loading.value = false
  }
}

function formatDate(dateString) {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

function getTypeColor(type) {
  const colors = {
    free: 'grey',
    monthly: 'blue',
    annual: 'green'
  }
  return colors[type] || 'grey'
}

function getStatusColor(status) {
  const colors = {
    active: 'success',
    expired: 'error',
    cancelled: 'warning'
  }
  return colors[status] || 'grey'
}

function getDaysLeftColor(days) {
  if (days <= 7) return 'error'
  if (days <= 14) return 'warning'
  return 'info'
}

// Watch for tab changes to load expiring/expired data when needed
watch(activeTab, (newTab) => {
  if (newTab === 'expiring' || newTab === 'overdue') {
    // Only load if we haven't already (or want to refresh)
    if (expiringUserSubscriptions.value.length === 0 && expiredUserSubscriptions.value.length === 0) {
      loadExpiringExpired()
    }
  }
})

onMounted(() => {
  loadSubscriptions()
  // Also load expiring/expired data in case user navigates directly to those tabs
  loadExpiringExpired()
})
</script>

<style scoped>
.mobile-view-wrapper {
  padding-bottom: 70px;
}
</style>
