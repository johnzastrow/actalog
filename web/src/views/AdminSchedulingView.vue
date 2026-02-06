<template>
  <div class="scheduling-view">
    <!-- Page Header -->
    <div class="page-header pa-4 d-flex align-center">
      <v-btn icon variant="text" color="white" class="mr-2" @click="$router.back()">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <div class="flex-grow-1">
        <div class="text-h6 font-weight-bold text-white">Class Scheduling</div>
        <div class="text-caption text-white" style="opacity: 0.8;">Manage templates, sessions, and coaches</div>
      </div>
      <v-icon color="white" size="large">mdi-calendar-clock</v-icon>
    </div>

    <v-container fluid class="pa-4 bg-grey-lighten-4">
      <!-- Organization Selector -->
      <div class="form-section mb-4">
        <v-select
          v-model="selectedOrgId"
          :items="organizations"
          item-title="name"
          item-value="id"
          label="Select Gym"
          variant="solo-filled"
          density="comfortable"
          rounded="lg"
          flat
          hide-details
          @update:model-value="onOrgChange"
        />
      </div>

      <!-- No Organizations Alert -->
      <v-alert v-if="!loading && organizations.length === 0" type="warning" variant="tonal" class="mb-4" rounded="lg">
        No gyms configured. Create a gym in the Organizations section first.
      </v-alert>

      <!-- Loading State -->
      <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />

      <!-- Error Alert -->
      <v-alert v-if="error" type="error" variant="tonal" closable class="mb-4" rounded="lg" @click:close="error = null">
        {{ error }}
      </v-alert>

      <!-- Success Alert -->
      <v-alert v-if="successMessage" type="success" variant="tonal" closable class="mb-4" rounded="lg" @click:close="successMessage = null">
        {{ successMessage }}
      </v-alert>

      <!-- Tabs -->
      <div class="form-section mb-4 pa-1">
        <v-tabs v-model="activeTab" bg-color="transparent" slider-color="primary">
          <v-tab value="locations" class="text-none" rounded="lg">
            <v-icon start size="small">mdi-map-marker</v-icon>
            Locations
          </v-tab>
          <v-tab value="templates" class="text-none" rounded="lg">
            <v-icon start size="small">mdi-clipboard-list</v-icon>
            Classes
          </v-tab>
          <v-tab value="sessions" class="text-none" rounded="lg">
            <v-icon start size="small">mdi-calendar</v-icon>
            Sessions
          </v-tab>
          <v-tab value="coaches" class="text-none" rounded="lg">
            <v-icon start size="small">mdi-account-tie</v-icon>
            Coaches
          </v-tab>
        </v-tabs>
      </div>

      <!-- Locations Tab -->
      <v-window v-model="activeTab">
        <v-window-item value="locations">
          <div class="form-section">
            <div class="section-header d-flex align-center mb-3">
              <div class="section-title">
                <v-icon size="small" class="mr-2">mdi-map-marker</v-icon>
                Locations ({{ locations.length }})
              </div>
              <v-spacer />
              <v-btn color="primary" size="small" rounded="lg" :disabled="!selectedOrgId" @click="openLocationDialog()">
                <v-icon start>mdi-plus</v-icon>
                Add Location
              </v-btn>
            </div>
            <v-list v-if="locations.length > 0" bg-color="transparent" class="pa-0">
              <v-list-item v-for="loc in locations" :key="loc.id" class="list-item-card mb-2" rounded="lg">
                <v-list-item-title class="font-weight-medium">{{ loc.name }}</v-list-item-title>
                <v-list-item-subtitle>
                  <span v-if="loc.address">{{ loc.address }}</span>
                  <span v-if="loc.capacity"> | Capacity: {{ loc.capacity }}</span>
                </v-list-item-subtitle>
                <template #append>
                  <v-chip :color="loc.is_active ? 'success' : 'grey'" size="x-small" class="mr-2" variant="flat" rounded="lg">
                    {{ loc.is_active ? 'Active' : 'Inactive' }}
                  </v-chip>
                  <v-btn icon size="small" variant="text" @click="editLocation(loc)">
                    <v-icon>mdi-pencil</v-icon>
                  </v-btn>
                  <v-btn icon size="small" variant="text" @click="deleteLocation(loc)">
                    <v-icon color="error">mdi-delete</v-icon>
                  </v-btn>
                </template>
              </v-list-item>
            </v-list>
            <v-alert v-else type="info" variant="tonal" rounded="lg">No locations configured.</v-alert>
          </div>
        </v-window-item>

        <!-- Classes Tab -->
        <v-window-item value="templates">
          <div class="form-section">
            <div class="section-header d-flex align-center mb-3">
              <div class="section-title">
                <v-icon size="small" class="mr-2">mdi-clipboard-list</v-icon>
                Classes ({{ filteredTemplates.length }}<span v-if="filteredTemplates.length !== templates.length"> of {{ templates.length }}</span>)
              </div>
              <v-spacer />
              <v-btn color="primary" size="small" rounded="lg" :disabled="!selectedOrgId" @click="openTemplateDialog()">
                <v-icon start>mdi-plus</v-icon>
                Add Class
              </v-btn>
            </div>

            <!-- Filter Section -->
            <div class="filter-section mb-3">
              <v-row dense>
                <v-col cols="12" sm="4">
                  <v-text-field
                    v-model="classFilter.search"
                    label="Search classes"
                    prepend-inner-icon="mdi-magnify"
                    variant="solo-filled"
                    density="compact"
                    rounded="lg"
                    flat
                    clearable
                    hide-details
                  />
                </v-col>
                <v-col cols="6" sm="4">
                  <v-select
                    v-model="classFilter.locationId"
                    :items="[{ id: null, name: 'All Locations' }, ...locations]"
                    item-title="name"
                    item-value="id"
                    label="Location"
                    variant="solo-filled"
                    density="compact"
                    rounded="lg"
                    flat
                    hide-details
                  />
                </v-col>
                <v-col cols="6" sm="4">
                  <v-select
                    v-model="classFilter.coachId"
                    :items="[{ user_id: null, user_name: 'All Coaches' }, ...coaches.map(c => ({ user_id: c.user_id, user_name: c.user_name || c.user_email }))]"
                    item-title="user_name"
                    item-value="user_id"
                    label="Coach"
                    variant="solo-filled"
                    density="compact"
                    rounded="lg"
                    flat
                    hide-details
                  />
                </v-col>
              </v-row>
            </div>

            <v-list v-if="filteredTemplates.length > 0" lines="three" bg-color="transparent" class="pa-0">
              <v-list-item v-for="tmpl in filteredTemplates" :key="tmpl.id" class="list-item-card mb-2" rounded="lg" @click="editTemplate(tmpl)">
                <template #prepend>
                  <v-avatar :color="tmpl.color || '#00bcd4'" size="40">
                    <v-icon color="white">mdi-dumbbell</v-icon>
                  </v-avatar>
                </template>
                <v-list-item-title class="font-weight-medium">{{ tmpl.name }}</v-list-item-title>
                <v-list-item-subtitle>
                  <div>
                    {{ tmpl.duration_minutes }} min | Capacity: {{ tmpl.default_capacity }}
                    <span v-if="tmpl.default_location_name"> | <v-icon size="x-small">mdi-map-marker</v-icon> {{ tmpl.default_location_name }}</span>
                  </div>
                  <div v-if="tmpl.default_coaches && tmpl.default_coaches.length" class="mt-1">
                    <v-icon size="x-small">mdi-account-tie</v-icon>
                    {{ tmpl.default_coaches.map(c => c.user_name || c.user_email).join(', ') }}
                  </div>
                  <div v-if="tmpl.schedule_slots && tmpl.schedule_slots.length" class="mt-1">
                    <v-icon size="x-small">mdi-calendar-clock</v-icon>
                    {{ getScheduleSummary(tmpl.schedule_slots) }}
                  </div>
                </v-list-item-subtitle>
                <template #append>
                  <div class="d-flex flex-column align-end">
                    <v-chip :color="tmpl.is_active ? 'success' : 'grey'" size="x-small" class="mb-1" variant="flat" rounded="lg">
                      {{ tmpl.is_active ? 'Active' : 'Inactive' }}
                    </v-chip>
                    <div>
                      <v-btn icon size="small" variant="text" @click.stop="editTemplate(tmpl)">
                        <v-icon>mdi-pencil</v-icon>
                      </v-btn>
                      <v-btn icon size="small" variant="text" @click.stop="deleteTemplate(tmpl)">
                        <v-icon color="error">mdi-delete</v-icon>
                      </v-btn>
                    </div>
                  </div>
                </template>
              </v-list-item>
            </v-list>
            <v-alert v-else-if="templates.length > 0" type="info" variant="tonal" rounded="lg">No classes match the current filters.</v-alert>
            <v-alert v-else type="info" variant="tonal" rounded="lg">No class templates configured.</v-alert>
          </div>
        </v-window-item>

        <!-- Sessions Tab -->
        <v-window-item value="sessions">
          <div class="form-section">
            <SessionsGrid
              ref="sessionsGrid"
              :gym-id="selectedOrgId"
              :locations="locations"
              :coaches="coaches"
              :templates="templates"
              @create="openSessionDialog()"
              @view-roster="viewRoster"
              @cancel="cancelSession"
              @complete="completeSession"
              @updated="onSessionUpdated"
            />
          </div>
        </v-window-item>

        <!-- Coaches Tab -->
        <v-window-item value="coaches">
          <div class="form-section">
            <div class="section-header d-flex align-center mb-3">
              <div class="section-title">
                <v-icon size="small" class="mr-2">mdi-account-tie</v-icon>
                Coaches per Gym ({{ coaches.length }})
              </div>
              <v-spacer />
              <v-btn color="primary" size="small" rounded="lg" :disabled="!selectedOrgId" @click="openCoachDialog()">
                <v-icon start>mdi-plus</v-icon>
                Assign Coach
              </v-btn>
            </div>
            <v-list v-if="coaches.length > 0" bg-color="transparent" class="pa-0">
              <v-list-item v-for="coach in coaches" :key="coach.id" class="list-item-card mb-2" rounded="lg">
                <template #prepend>
                  <v-avatar color="primary">
                    <span class="text-caption">{{ getInitials(coach.user_name || coach.user_email) }}</span>
                  </v-avatar>
                </template>
                <v-list-item-title class="font-weight-medium">{{ coach.user_name || 'Unknown' }}</v-list-item-title>
                <v-list-item-subtitle>{{ coach.user_email }}</v-list-item-subtitle>
                <template #append>
                  <v-chip :color="coach.is_active ? 'success' : 'grey'" size="x-small" class="mr-2" variant="flat" rounded="lg">
                    {{ coach.is_active ? 'Active' : 'Inactive' }}
                  </v-chip>
                  <v-btn icon size="small" variant="text" @click="removeCoach(coach)">
                    <v-icon color="error">mdi-account-remove</v-icon>
                  </v-btn>
                </template>
              </v-list-item>
            </v-list>
            <v-alert v-else type="info" variant="tonal" rounded="lg">No coaches assigned to this organization.</v-alert>
          </div>
        </v-window-item>
      </v-window>
    </v-container>

    <!-- Location Dialog -->
    <v-dialog v-model="locationDialog" max-width="500">
      <v-card>
        <v-card-title>{{ editingLocation ? 'Edit Location' : 'Add Location' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="locationForm.name" label="Name" required />
          <v-textarea v-model="locationForm.description" label="Description" rows="2" />
          <v-text-field v-model="locationForm.address" label="Address" />
          <v-text-field v-model.number="locationForm.capacity" label="Capacity" type="number" />
          <v-switch v-if="editingLocation" v-model="locationForm.is_active" label="Active" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="locationDialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="saveLocation">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Template Dialog (Enhanced) -->
    <TemplateEditDialog
      v-model="templateDialog"
      :template="editingTemplate"
      :gym-id="selectedOrgId"
      :locations="locations"
      :workouts="workouts"
      :coaches="coaches"
      @saved="onTemplateSaved"
    />

    <!-- Session Dialog -->
    <v-dialog v-model="sessionDialog" max-width="500">
      <v-card>
        <v-card-title>{{ editingSession ? 'Edit Session' : 'Create Session' }}</v-card-title>
        <v-card-text>
          <v-select
            v-model="sessionForm.template_id"
            :items="templates"
            item-title="name"
            item-value="id"
            label="From Template (optional)"
            clearable
          />
          <v-text-field v-model="sessionForm.name" label="Name" required />
          <v-text-field v-model="sessionForm.start_time" label="Start Time" type="datetime-local" required />
          <v-text-field v-model="sessionForm.end_time" label="End Time" type="datetime-local" />
          <v-text-field v-model.number="sessionForm.capacity" label="Capacity" type="number" />
          <v-select
            v-model="sessionForm.location_id"
            :items="locations"
            item-title="name"
            item-value="id"
            label="Location"
            clearable
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="sessionDialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="saveSession">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Roster Dialog -->
    <v-dialog v-model="rosterDialog" max-width="700">
      <v-card>
        <v-card-title class="d-flex align-center">
          <v-icon icon="mdi-clipboard-list" class="mr-2"></v-icon>
          Session Roster
          <v-spacer></v-spacer>
          <v-chip v-if="currentRosterSession" :color="getSessionStatusColor(currentRosterSession.status)" size="small">
            {{ currentRosterSession.status }}
          </v-chip>
        </v-card-title>
        <v-card-text>
          <v-list v-if="roster.length > 0">
            <v-list-item v-for="res in roster" :key="res.id">
              <template #prepend>
                <v-avatar :color="getReservationStatusColor(res.status)">
                  <v-icon color="white">{{ getReservationStatusIcon(res.status) }}</v-icon>
                </v-avatar>
              </template>
              <v-list-item-title>{{ res.user_name || res.user_email }}</v-list-item-title>
              <v-list-item-subtitle>{{ res.user_email }}</v-list-item-subtitle>
              <template #append>
                <v-chip :color="getReservationStatusColor(res.status)" size="small" class="mr-2">
                  {{ res.status }}
                </v-chip>
                <v-btn v-if="res.status === 'reserved'" color="success" size="small" class="mr-1" @click="checkInReservation(res)">
                  Check In
                </v-btn>
                <v-btn v-if="res.status === 'reserved' || res.status === 'checked_in'" color="error" size="small" variant="tonal" @click="markNoShow(res)">
                  No Show
                </v-btn>
              </template>
            </v-list-item>
          </v-list>
          <v-alert v-else type="info" variant="tonal">No reservations for this session.</v-alert>
        </v-card-text>
        <v-card-actions>
          <v-btn
            v-if="currentRosterSession && (currentRosterSession.status === 'scheduled' || currentRosterSession.status === 'in_progress')"
            color="success"
            variant="flat"
            @click="completeSession"
          >
            <v-icon icon="mdi-check-circle" class="mr-1"></v-icon>
            Complete Session
          </v-btn>
          <v-spacer />
          <v-btn @click="rosterDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Coach Assignment Dialog -->
    <v-dialog v-model="coachDialog" max-width="400">
      <v-card>
        <v-card-title>Assign Coach to Gym</v-card-title>
        <v-card-text>
          <v-autocomplete
            v-model="coachForm.user_id"
            v-model:search="userSearch"
            :items="availableUsers"
            :loading="searchingUsers"
            item-title="display_name"
            item-value="id"
            label="Search user by name or email"
            placeholder="Start typing to search..."
            no-data-text="No users found"
            clearable
            @update:search="onUserSearch"
          >
            <template #item="{ props, item }">
              <v-list-item v-bind="props">
                <template #prepend>
                  <v-avatar color="primary" size="32">
                    <span class="text-caption">{{ getInitials(item.raw.name || item.raw.email) }}</span>
                  </v-avatar>
                </template>
                <v-list-item-subtitle>{{ item.raw.email }}</v-list-item-subtitle>
              </v-list-item>
            </template>
          </v-autocomplete>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="coachDialog = false">Cancel</v-btn>
          <v-btn color="primary" :disabled="!coachForm.user_id" @click="assignCoach">Assign</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Snackbar for notifications -->
    <v-snackbar
      v-model="snackbar"
      :color="snackbarColor"
      :timeout="4000"
      location="bottom"
      :style="{ marginBottom: '80px' }"
    >
      {{ snackbarText }}
      <template #actions>
        <v-btn variant="text" @click="snackbar = false">Close</v-btn>
      </template>
    </v-snackbar>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from '@/utils/axios'
import TemplateEditDialog from '@/components/scheduling/TemplateEditDialog.vue'
import SessionsGrid from '@/components/scheduling/SessionsGrid.vue'

const loading = ref(false)
const error = ref(null)
const successMessage = ref(null)

// Snackbar state
const snackbar = ref(false)
const snackbarText = ref('')
const snackbarColor = ref('success')

function showSnackbar(message, color = 'success') {
  snackbarText.value = message
  snackbarColor.value = color
  snackbar.value = true
}

const organizations = ref([])
const selectedOrgId = ref(null)
const activeTab = ref('locations')

const locations = ref([])
const templates = ref([])
const sessions = ref([])
const coaches = ref([])
const roster = ref([])
const workouts = ref([])

// Class filter state
const classFilter = ref({
  search: '',
  locationId: null,
  coachId: null
})

// Filtered templates computed
const filteredTemplates = computed(() => {
  let result = templates.value

  // Text search
  if (classFilter.value.search) {
    const search = classFilter.value.search.toLowerCase()
    result = result.filter(t =>
      t.name.toLowerCase().includes(search) ||
      (t.description && t.description.toLowerCase().includes(search)) ||
      (t.default_location_name && t.default_location_name.toLowerCase().includes(search))
    )
  }

  // Location filter
  if (classFilter.value.locationId) {
    result = result.filter(t => t.default_location_id === classFilter.value.locationId)
  }

  // Coach filter
  if (classFilter.value.coachId) {
    result = result.filter(t =>
      t.default_coaches && t.default_coaches.some(c => c.user_id === classFilter.value.coachId)
    )
  }

  return result
})

// Generate schedule summary text
function getScheduleSummary(slots) {
  if (!slots || slots.length === 0) return 'No schedule'

  const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  const summaries = []

  for (const slot of slots) {
    // Get days from either days_of_week array or single day_of_week
    const days = slot.days_of_week && slot.days_of_week.length > 0
      ? slot.days_of_week
      : [slot.day_of_week]

    const dayStr = days.map(d => dayNames[d]).join('/')
    const time = slot.start_time ? slot.start_time.substring(0, 5) : '??:??'
    summaries.push(`${dayStr} @ ${time}`)
  }

  return summaries.join(', ')
}

const startDate = ref(new Date())
const endDate = ref(new Date())

// Dialogs
const locationDialog = ref(false)
const templateDialog = ref(false)
const sessionDialog = ref(false)
const rosterDialog = ref(false)
const coachDialog = ref(false)

// Form states
const editingLocation = ref(null)
const locationForm = ref({ name: '', description: '', address: '', capacity: 0, is_active: true })

const editingTemplate = ref(null)
const templateForm = ref({ name: '', description: '', duration_minutes: 60, default_capacity: 20, color: '#00bcd4', is_active: true })

const editingSession = ref(null)
const sessionForm = ref({ name: '', template_id: null, start_time: '', end_time: '', capacity: 20, location_id: null })

const coachForm = ref({ user_id: null })
const availableUsers = ref([])
const searchingUsers = ref(false)
const userSearch = ref('')
const currentRosterSessionId = ref(null)
const currentRosterSession = ref(null)
const sessionsGrid = ref(null)

// Initialize
onMounted(() => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  startDate.value = today
  const end = new Date(today)
  end.setDate(end.getDate() + 6)
  endDate.value = end
  fetchOrganizations()
})

async function fetchOrganizations() {
  try {
    const response = await axios.get('/api/admin/organizations')
    organizations.value = response.data.organizations || []
    if (organizations.value.length > 0) {
      selectedOrgId.value = organizations.value[0].id
      onOrgChange()
    }
  } catch (err) {
    error.value = 'Failed to fetch organizations'
  }
}

function onOrgChange() {
  fetchLocations()
  fetchTemplates()
  fetchSessions()
  fetchCoaches()
  fetchWorkouts()
}

async function fetchLocations() {
  if (!selectedOrgId.value) return
  try {
    const response = await axios.get(`/api/gyms/${selectedOrgId.value}/locations?include_inactive=true`)
    locations.value = response.data.locations || []
  } catch (err) {
    console.error('Failed to fetch locations:', err)
  }
}

async function fetchTemplates() {
  if (!selectedOrgId.value) return
  try {
    const response = await axios.get(`/api/gyms/${selectedOrgId.value}/templates?include_inactive=true`)
    templates.value = response.data.templates || []
  } catch (err) {
    console.error('Failed to fetch templates:', err)
  }
}

async function fetchSessions() {
  if (!selectedOrgId.value) return
  try {
    const start = startDate.value.toISOString().split('T')[0]
    const end = endDate.value.toISOString().split('T')[0]
    const response = await axios.get(`/api/gyms/${selectedOrgId.value}/sessions?start_date=${start}&end_date=${end}`)
    sessions.value = response.data.sessions || []
  } catch (err) {
    console.error('Failed to fetch sessions:', err)
  }
}

async function fetchCoaches() {
  if (!selectedOrgId.value) return
  try {
    const response = await axios.get(`/api/admin/gyms/${selectedOrgId.value}/coaches?include_inactive=true`)
    coaches.value = response.data.coaches || []
  } catch (err) {
    console.error('Failed to fetch coaches:', err)
  }
}

async function fetchWorkouts() {
  try {
    const response = await axios.get('/api/templates')
    workouts.value = response.data.templates || []
  } catch (err) {
    console.error('Failed to fetch workouts:', err)
  }
}

// Location functions
function openLocationDialog(loc = null) {
  editingLocation.value = loc
  locationForm.value = loc ? { ...loc } : { name: '', description: '', address: '', capacity: 0, is_active: true }
  locationDialog.value = true
}

function editLocation(loc) {
  openLocationDialog(loc)
}

async function saveLocation() {
  if (!selectedOrgId.value) {
    error.value = 'Please select a gym first'
    return
  }
  if (!locationForm.value.name?.trim()) {
    error.value = 'Location name is required'
    return
  }
  loading.value = true
  try {
    if (editingLocation.value) {
      await axios.put(`/api/admin/gyms/${selectedOrgId.value}/locations/${editingLocation.value.id}`, locationForm.value)
    } else {
      await axios.post(`/api/admin/gyms/${selectedOrgId.value}/locations`, locationForm.value)
    }
    successMessage.value = 'Location saved successfully'
    locationDialog.value = false
    fetchLocations()
  } catch (err) {
    error.value = err.response?.data?.message || err.response?.data?.error || 'Failed to save location'
  } finally {
    loading.value = false
  }
}

async function deleteLocation(loc) {
  if (!confirm(`Delete location "${loc.name}"?`)) return
  try {
    await axios.delete(`/api/admin/gyms/${selectedOrgId.value}/locations/${loc.id}`)
    successMessage.value = 'Location deleted'
    fetchLocations()
  } catch (err) {
    error.value = err.response?.data?.message || err.response?.data?.error || 'Failed to delete location'
  }
}

// Template functions
function openTemplateDialog(tmpl = null) {
  editingTemplate.value = tmpl
  templateDialog.value = true
}

function editTemplate(tmpl) {
  openTemplateDialog(tmpl)
}

function onTemplateSaved() {
  console.log('[AdminSchedulingView] onTemplateSaved called')
  successMessage.value = 'Class saved successfully'
  fetchTemplates()
}

async function deleteTemplate(tmpl) {
  if (!confirm(`Delete class "${tmpl.name}"?`)) return
  try {
    await axios.delete(`/api/admin/scheduling/templates/${tmpl.id}`)
    successMessage.value = 'Class deleted'
    fetchTemplates()
  } catch (err) {
    error.value = err.response?.data?.message || err.response?.data?.error || 'Failed to delete class'
  }
}

// Session functions
function openSessionDialog(sess = null) {
  editingSession.value = sess
  if (sess) {
    sessionForm.value = {
      name: sess.name,
      template_id: sess.template_id,
      start_time: new Date(sess.start_time).toISOString().slice(0, 16),
      end_time: new Date(sess.end_time).toISOString().slice(0, 16),
      capacity: sess.capacity,
      location_id: sess.location_id
    }
  } else {
    sessionForm.value = { name: '', template_id: null, start_time: '', end_time: '', capacity: 20, location_id: null }
  }
  sessionDialog.value = true
}

function editSession(sess) {
  openSessionDialog(sess)
}

async function saveSession() {
  if (!selectedOrgId.value) {
    error.value = 'Please select a gym first'
    return
  }
  if (!sessionForm.value.name) {
    error.value = 'Session name is required'
    return
  }
  if (!sessionForm.value.start_time) {
    error.value = 'Start time is required'
    return
  }
  loading.value = true
  try {
    const data = {
      ...sessionForm.value,
      start_time: new Date(sessionForm.value.start_time).toISOString(),
      end_time: sessionForm.value.end_time ? new Date(sessionForm.value.end_time).toISOString() : null
    }
    if (editingSession.value) {
      await axios.put(`/api/admin/gyms/${selectedOrgId.value}/sessions/${editingSession.value.id}`, data)
    } else {
      await axios.post(`/api/admin/gyms/${selectedOrgId.value}/sessions`, data)
    }
    successMessage.value = 'Session saved successfully'
    sessionDialog.value = false
    fetchSessions()
  } catch (err) {
    error.value = err.response?.data?.message || err.response?.data?.error || 'Failed to save session'
  } finally {
    loading.value = false
  }
}

async function cancelSession(sess) {
  const reason = prompt('Enter cancellation reason:')
  if (reason === null) return
  try {
    await axios.post(`/api/admin/gyms/${selectedOrgId.value}/sessions/${sess.id}/cancel`, { reason })
    successMessage.value = 'Session cancelled'
    if (sessionsGrid.value) sessionsGrid.value.refresh()
  } catch (err) {
    error.value = err.response?.data?.message || err.response?.data?.error || 'Failed to cancel session'
  }
}

async function completeSession(sess) {
  if (!confirm(`Mark "${sess.name}" as completed?`)) return
  try {
    await axios.post(`/api/admin/sessions/${sess.id}/complete`)
    successMessage.value = 'Session completed'
    if (sessionsGrid.value) sessionsGrid.value.refresh()
  } catch (err) {
    error.value = err.response?.data?.message || err.response?.data?.error || 'Failed to complete session'
  }
}

function onSessionUpdated() {
  successMessage.value = 'Session updated'
}

async function viewRoster(sess) {
  currentRosterSessionId.value = sess.id
  currentRosterSession.value = sess
  try {
    const response = await axios.get(`/api/admin/sessions/${sess.id}/roster`)
    roster.value = response.data.roster || []
    rosterDialog.value = true
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to fetch roster'
  }
}

async function checkInReservation(res) {
  try {
    await axios.post(`/api/admin/sessions/${currentRosterSessionId.value}/check-in/${res.id}`)
    successMessage.value = 'Checked in successfully'
    viewRoster(currentRosterSession.value)
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to check in'
  }
}

async function markNoShow(res) {
  try {
    await axios.post(`/api/admin/sessions/${currentRosterSessionId.value}/no-show/${res.id}`)
    successMessage.value = 'Marked as no-show'
    viewRoster(currentRosterSession.value)
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to mark no-show'
  }
}

// Coach functions
function openCoachDialog() {
  coachForm.value = { user_id: null }
  availableUsers.value = []
  userSearch.value = ''
  coachDialog.value = true
}

async function onUserSearch(search) {
  if (!search || search.length < 2) {
    availableUsers.value = []
    return
  }
  searchingUsers.value = true
  try {
    const response = await axios.get(`/api/admin/user-management/filter?search=${encodeURIComponent(search)}&limit=20`)
    const users = response.data.users || []
    // Filter out users already assigned as coaches
    const existingCoachIds = coaches.value.map(c => c.user_id)
    availableUsers.value = users
      .filter(u => !existingCoachIds.includes(u.id))
      .map(u => ({
        ...u,
        display_name: u.name ? `${u.name} (${u.email})` : u.email
      }))
  } catch (err) {
    console.error('Failed to search users:', err)
    availableUsers.value = []
  } finally {
    searchingUsers.value = false
  }
}

async function assignCoach() {
  if (!selectedOrgId.value) {
    error.value = 'Please select a gym first'
    return
  }
  if (!coachForm.value.user_id) {
    error.value = 'User ID is required'
    return
  }
  try {
    await axios.post(`/api/admin/gyms/${selectedOrgId.value}/coaches`, coachForm.value)
    successMessage.value = 'Coach assigned successfully'
    coachDialog.value = false
    fetchCoaches()
  } catch (err) {
    error.value = err.response?.data?.message || err.response?.data?.error || 'Failed to assign coach'
  }
}

async function removeCoach(coach) {
  if (!confirm(`Remove coach assignment?`)) return
  try {
    await axios.delete(`/api/admin/gyms/${selectedOrgId.value}/coaches/${coach.id}`)
    successMessage.value = 'Coach removed'
    fetchCoaches()
  } catch (err) {
    error.value = err.response?.data?.message || err.response?.data?.error || 'Failed to remove coach'
  }
}

// Utility functions
function previousWeek() {
  const newStart = new Date(startDate.value)
  newStart.setDate(newStart.getDate() - 7)
  startDate.value = newStart
  const newEnd = new Date(endDate.value)
  newEnd.setDate(newEnd.getDate() - 7)
  endDate.value = newEnd
  fetchSessions()
}

function nextWeek() {
  const newStart = new Date(startDate.value)
  newStart.setDate(newStart.getDate() + 7)
  startDate.value = newStart
  const newEnd = new Date(endDate.value)
  newEnd.setDate(newEnd.getDate() + 7)
  endDate.value = newEnd
  fetchSessions()
}

function formatDateRange(start, end) {
  const startStr = start.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  const endStr = end.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  return `${startStr} - ${endStr}`
}

function formatDateTime(dateTimeStr) {
  return new Date(dateTimeStr).toLocaleString('en-US', {
    month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit'
  })
}

function getSessionStatusColor(status) {
  const colors = { scheduled: 'primary', in_progress: 'warning', completed: 'success', cancelled: 'error' }
  return colors[status] || 'grey'
}

function getReservationStatusColor(status) {
  const colors = { reserved: 'primary', checked_in: 'success', attended: 'success', cancelled: 'grey', no_show: 'error' }
  return colors[status] || 'grey'
}

function getReservationStatusIcon(status) {
  const icons = { reserved: 'mdi-calendar-clock', checked_in: 'mdi-check', attended: 'mdi-check-all', cancelled: 'mdi-cancel', no_show: 'mdi-account-off' }
  return icons[status] || 'mdi-help'
}

function getInitials(name) {
  if (!name) return '?'
  const parts = name.split(' ')
  return parts.length > 1 ? (parts[0][0] + parts[parts.length - 1][0]).toUpperCase() : name[0].toUpperCase()
}
</script>

<style scoped>
.scheduling-view {
  min-height: 100vh;
  padding-bottom: 70px;
  background-color: #f5f5f5;
}

/* Page Header - matches dialog header */
.page-header {
  background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
  color: white;
}

/* Form Sections - white cards with subtle shadow */
.form-section {
  background-color: white;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

/* Section Title */
.section-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #546e7a;
  display: flex;
  align-items: center;
}

/* List item cards */
.list-item-card {
  background-color: #f8f9fa;
  transition: background-color 0.2s;
}

.list-item-card:hover {
  background-color: #e9ecef;
}

/* Filter section background */
.filter-section {
  background-color: #f8f9fa;
  border-radius: 8px;
  padding: 12px;
}
</style>
