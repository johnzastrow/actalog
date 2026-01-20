<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <div class="d-flex align-center mb-4">
        <v-btn icon class="mr-2" @click="$router.back()">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <div>
          <h1 class="text-h5">Class Scheduling</h1>
          <div class="text-body-2 text-medium-emphasis">Manage templates, sessions, and coaches</div>
        </div>
      </div>

      <!-- Organization Selector -->
      <v-select
        v-model="selectedOrgId"
        :items="organizations"
        item-title="name"
        item-value="id"
        label="Select Gym"
        variant="outlined"
        density="comfortable"
        class="mb-4"
        @update:model-value="onOrgChange"
      />

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

      <!-- Tabs -->
      <v-tabs v-model="activeTab" class="mb-4">
        <v-tab value="locations">Locations</v-tab>
        <v-tab value="templates">Templates</v-tab>
        <v-tab value="sessions">Sessions</v-tab>
        <v-tab value="coaches">Coaches</v-tab>
      </v-tabs>

      <!-- Locations Tab -->
      <v-window v-model="activeTab">
        <v-window-item value="locations">
          <v-card elevation="0" rounded="lg">
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2">mdi-map-marker</v-icon>
              Locations ({{ locations.length }})
              <v-spacer />
              <v-btn color="primary" size="small" @click="openLocationDialog()">
                <v-icon start>mdi-plus</v-icon>
                Add Location
              </v-btn>
            </v-card-title>
            <v-divider />
            <v-list v-if="locations.length > 0">
              <v-list-item v-for="loc in locations" :key="loc.id">
                <v-list-item-title>{{ loc.name }}</v-list-item-title>
                <v-list-item-subtitle>
                  <span v-if="loc.address">{{ loc.address }}</span>
                  <span v-if="loc.capacity"> | Capacity: {{ loc.capacity }}</span>
                </v-list-item-subtitle>
                <template #append>
                  <v-chip :color="loc.is_active ? 'success' : 'grey'" size="x-small" class="mr-2">
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
            <v-card-text v-else>
              <v-alert type="info" variant="tonal">No locations configured.</v-alert>
            </v-card-text>
          </v-card>
        </v-window-item>

        <!-- Templates Tab -->
        <v-window-item value="templates">
          <v-card elevation="0" rounded="lg">
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2">mdi-clipboard-list</v-icon>
              Class Templates ({{ templates.length }})
              <v-spacer />
              <v-btn color="primary" size="small" @click="openTemplateDialog()">
                <v-icon start>mdi-plus</v-icon>
                Add Template
              </v-btn>
            </v-card-title>
            <v-divider />
            <v-list v-if="templates.length > 0">
              <v-list-item v-for="tmpl in templates" :key="tmpl.id">
                <template #prepend>
                  <v-avatar :color="tmpl.color || '#00bcd4'" size="40">
                    <v-icon color="white">mdi-dumbbell</v-icon>
                  </v-avatar>
                </template>
                <v-list-item-title>{{ tmpl.name }}</v-list-item-title>
                <v-list-item-subtitle>
                  {{ tmpl.duration_minutes }} min | Capacity: {{ tmpl.default_capacity }}
                  <span v-if="tmpl.workout_name"> | Workout: {{ tmpl.workout_name }}</span>
                </v-list-item-subtitle>
                <template #append>
                  <v-chip :color="tmpl.is_active ? 'success' : 'grey'" size="x-small" class="mr-2">
                    {{ tmpl.is_active ? 'Active' : 'Inactive' }}
                  </v-chip>
                  <v-btn icon size="small" variant="text" @click="editTemplate(tmpl)">
                    <v-icon>mdi-pencil</v-icon>
                  </v-btn>
                  <v-btn icon size="small" variant="text" @click="deleteTemplate(tmpl)">
                    <v-icon color="error">mdi-delete</v-icon>
                  </v-btn>
                </template>
              </v-list-item>
            </v-list>
            <v-card-text v-else>
              <v-alert type="info" variant="tonal">No class templates configured.</v-alert>
            </v-card-text>
          </v-card>
        </v-window-item>

        <!-- Sessions Tab -->
        <v-window-item value="sessions">
          <v-card elevation="0" rounded="lg">
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2">mdi-calendar</v-icon>
              Sessions
              <v-spacer />
              <v-btn color="primary" size="small" @click="openSessionDialog()">
                <v-icon start>mdi-plus</v-icon>
                Create Session
              </v-btn>
            </v-card-title>
            <v-divider />

            <!-- Date Navigation -->
            <div class="pa-3 d-flex align-center justify-space-between">
              <v-btn icon variant="text" size="small" @click="previousWeek">
                <v-icon>mdi-chevron-left</v-icon>
              </v-btn>
              <span class="text-body-2">{{ formatDateRange(startDate, endDate) }}</span>
              <v-btn icon variant="text" size="small" @click="nextWeek">
                <v-icon>mdi-chevron-right</v-icon>
              </v-btn>
            </div>

            <v-list v-if="sessions.length > 0">
              <v-list-item v-for="sess in sessions" :key="sess.id">
                <v-list-item-title>
                  {{ sess.name }}
                  <v-chip :color="getSessionStatusColor(sess.status)" size="x-small" class="ml-2">
                    {{ sess.status }}
                  </v-chip>
                </v-list-item-title>
                <v-list-item-subtitle>
                  {{ formatDateTime(sess.start_time) }}
                  | {{ sess.reservation_count }}/{{ sess.capacity }} reserved
                  <span v-if="sess.location_name"> | {{ sess.location_name }}</span>
                </v-list-item-subtitle>
                <template #append>
                  <v-btn icon size="small" variant="text" title="View Roster" @click="viewRoster(sess)">
                    <v-icon>mdi-account-multiple</v-icon>
                  </v-btn>
                  <v-btn v-if="sess.status === 'scheduled'" icon size="small" variant="text" @click="editSession(sess)">
                    <v-icon>mdi-pencil</v-icon>
                  </v-btn>
                  <v-btn v-if="sess.status === 'scheduled'" icon size="small" variant="text" @click="cancelSession(sess)">
                    <v-icon color="error">mdi-cancel</v-icon>
                  </v-btn>
                </template>
              </v-list-item>
            </v-list>
            <v-card-text v-else>
              <v-alert type="info" variant="tonal">No sessions scheduled for this period.</v-alert>
            </v-card-text>
          </v-card>
        </v-window-item>

        <!-- Coaches Tab -->
        <v-window-item value="coaches">
          <v-card elevation="0" rounded="lg">
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2">mdi-account-tie</v-icon>
              Coaches ({{ coaches.length }})
              <v-spacer />
              <v-btn color="primary" size="small" @click="openCoachDialog()">
                <v-icon start>mdi-plus</v-icon>
                Assign Coach
              </v-btn>
            </v-card-title>
            <v-divider />
            <v-list v-if="coaches.length > 0">
              <v-list-item v-for="coach in coaches" :key="coach.id">
                <template #prepend>
                  <v-avatar color="primary">
                    <span class="text-caption">{{ getInitials(coach.user_name || coach.user_email) }}</span>
                  </v-avatar>
                </template>
                <v-list-item-title>{{ coach.user_name || 'Unknown' }}</v-list-item-title>
                <v-list-item-subtitle>{{ coach.user_email }}</v-list-item-subtitle>
                <template #append>
                  <v-chip :color="coach.is_active ? 'success' : 'grey'" size="x-small" class="mr-2">
                    {{ coach.is_active ? 'Active' : 'Inactive' }}
                  </v-chip>
                  <v-btn icon size="small" variant="text" @click="removeCoach(coach)">
                    <v-icon color="error">mdi-account-remove</v-icon>
                  </v-btn>
                </template>
              </v-list-item>
            </v-list>
            <v-card-text v-else>
              <v-alert type="info" variant="tonal">No coaches assigned to this organization.</v-alert>
            </v-card-text>
          </v-card>
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

    <!-- Template Dialog -->
    <v-dialog v-model="templateDialog" max-width="500">
      <v-card>
        <v-card-title>{{ editingTemplate ? 'Edit Template' : 'Add Template' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="templateForm.name" label="Name" required />
          <v-textarea v-model="templateForm.description" label="Description" rows="2" />
          <v-text-field v-model.number="templateForm.duration_minutes" label="Duration (minutes)" type="number" />
          <v-text-field v-model.number="templateForm.default_capacity" label="Default Capacity" type="number" />
          <v-text-field v-model="templateForm.color" label="Color (hex)" placeholder="#00bcd4" />
          <v-switch v-if="editingTemplate" v-model="templateForm.is_active" label="Active" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="templateDialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="saveTemplate">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

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
        <v-card-title>Assign Coach</v-card-title>
        <v-card-text>
          <v-text-field v-model.number="coachForm.user_id" label="User ID" type="number" required />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="coachDialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="assignCoach">Assign</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'

const loading = ref(false)
const error = ref(null)
const successMessage = ref(null)

const organizations = ref([])
const selectedOrgId = ref(null)
const activeTab = ref('locations')

const locations = ref([])
const templates = ref([])
const sessions = ref([])
const coaches = ref([])
const roster = ref([])

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
const currentRosterSessionId = ref(null)
const currentRosterSession = ref(null)

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
    error.value = err.response?.data?.error || 'Failed to save location'
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
    error.value = err.response?.data?.error || 'Failed to delete location'
  }
}

// Template functions
function openTemplateDialog(tmpl = null) {
  editingTemplate.value = tmpl
  templateForm.value = tmpl ? { ...tmpl } : { name: '', description: '', duration_minutes: 60, default_capacity: 20, color: '#00bcd4', is_active: true }
  templateDialog.value = true
}

function editTemplate(tmpl) {
  openTemplateDialog(tmpl)
}

async function saveTemplate() {
  loading.value = true
  try {
    if (editingTemplate.value) {
      await axios.put(`/api/admin/scheduling/templates/${editingTemplate.value.id}`, templateForm.value)
    } else {
      await axios.post(`/api/admin/gyms/${selectedOrgId.value}/templates`, templateForm.value)
    }
    successMessage.value = 'Template saved successfully'
    templateDialog.value = false
    fetchTemplates()
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to save template'
  } finally {
    loading.value = false
  }
}

async function deleteTemplate(tmpl) {
  if (!confirm(`Delete template "${tmpl.name}"?`)) return
  try {
    await axios.delete(`/api/admin/scheduling/templates/${tmpl.id}`)
    successMessage.value = 'Template deleted'
    fetchTemplates()
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to delete template'
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
    error.value = err.response?.data?.error || 'Failed to save session'
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
    fetchSessions()
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to cancel session'
  }
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

async function completeSession() {
  if (!confirm('Are you sure you want to complete this session?')) return
  try {
    await axios.post(`/api/admin/sessions/${currentRosterSessionId.value}/complete`)
    successMessage.value = 'Session completed successfully'
    currentRosterSession.value.status = 'completed'
    rosterDialog.value = false
    fetchSessions()
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to complete session'
  }
}

// Coach functions
function openCoachDialog() {
  coachForm.value = { user_id: null }
  coachDialog.value = true
}

async function assignCoach() {
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
    error.value = err.response?.data?.error || 'Failed to assign coach'
  }
}

async function removeCoach(coach) {
  if (!confirm(`Remove coach assignment?`)) return
  try {
    await axios.delete(`/api/admin/gyms/${selectedOrgId.value}/coaches/${coach.id}`)
    successMessage.value = 'Coach removed'
    fetchCoaches()
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to remove coach'
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
.mobile-view-wrapper {
  min-height: 100vh;
  padding-bottom: 70px;
}
</style>
