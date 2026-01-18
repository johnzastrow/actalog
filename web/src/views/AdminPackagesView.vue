<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <div class="d-flex align-center mb-4">
        <v-btn icon class="mr-2" @click="$router.back()">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <div>
          <h1 class="text-h5">Manage Packages & Documents</h1>
          <div class="text-body-2 text-medium-emphasis">Credit packages and required documents</div>
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
        @update:model-value="fetchData"
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
        <v-tab value="packages">Packages</v-tab>
        <v-tab value="documents">Documents</v-tab>
        <v-tab value="credits">User Credits</v-tab>
      </v-tabs>

      <!-- Packages Tab -->
      <v-window v-model="activeTab">
        <v-window-item value="packages">
          <v-card elevation="0" rounded="lg" v-if="selectedOrgId">
            <v-card-title class="d-flex align-center justify-space-between">
              <span>Class Packages</span>
              <v-btn color="primary" size="small" @click="openPackageDialog()">
                <v-icon start>mdi-plus</v-icon>
                Add Package
              </v-btn>
            </v-card-title>
            <v-card-text>
              <v-list v-if="packages.length > 0" density="compact">
                <v-list-item v-for="pkg in packages" :key="pkg.id">
                  <v-list-item-title>{{ pkg.name }}</v-list-item-title>
                  <v-list-item-subtitle>
                    {{ pkg.credits }} credits - ${{ (pkg.price_cents / 100).toFixed(2) }}
                    <span v-if="pkg.validity_days"> ({{ pkg.validity_days }} days)</span>
                    <v-chip v-if="!pkg.is_active" color="grey" size="x-small" class="ml-2">Inactive</v-chip>
                  </v-list-item-subtitle>
                  <template v-slot:append>
                    <v-btn icon variant="text" size="small" @click="openPackageDialog(pkg)">
                      <v-icon>mdi-pencil</v-icon>
                    </v-btn>
                    <v-btn icon variant="text" size="small" color="error" @click="deletePackage(pkg)">
                      <v-icon>mdi-delete</v-icon>
                    </v-btn>
                  </template>
                </v-list-item>
              </v-list>
              <v-alert v-else type="info" variant="tonal" density="compact">
                No packages configured yet.
              </v-alert>
            </v-card-text>
          </v-card>
        </v-window-item>

        <!-- Documents Tab -->
        <v-window-item value="documents">
          <v-card elevation="0" rounded="lg" v-if="selectedOrgId">
            <v-card-title class="d-flex align-center justify-space-between">
              <span>Required Documents</span>
              <v-btn color="primary" size="small" @click="openDocumentDialog()">
                <v-icon start>mdi-plus</v-icon>
                Add Document
              </v-btn>
            </v-card-title>
            <v-card-text>
              <v-list v-if="documents.length > 0" density="compact">
                <v-list-item v-for="doc in documents" :key="doc.id">
                  <v-list-item-title>{{ doc.name }}</v-list-item-title>
                  <v-list-item-subtitle>
                    {{ doc.document_type }}
                    <v-chip v-if="doc.is_required" color="warning" size="x-small" class="ml-2">Required</v-chip>
                    <v-chip v-if="!doc.is_active" color="grey" size="x-small" class="ml-2">Inactive</v-chip>
                    <span v-if="doc.expires_after_days" class="ml-2">(Expires after {{ doc.expires_after_days }} days)</span>
                  </v-list-item-subtitle>
                  <template v-slot:append>
                    <v-btn icon variant="text" size="small" @click="openDocumentDialog(doc)">
                      <v-icon>mdi-pencil</v-icon>
                    </v-btn>
                    <v-btn icon variant="text" size="small" color="error" @click="deleteDocument(doc)">
                      <v-icon>mdi-delete</v-icon>
                    </v-btn>
                  </template>
                </v-list-item>
              </v-list>
              <v-alert v-else type="info" variant="tonal" density="compact">
                No documents configured yet.
              </v-alert>
            </v-card-text>
          </v-card>
        </v-window-item>

        <!-- Credits Tab -->
        <v-window-item value="credits">
          <v-card elevation="0" rounded="lg" v-if="selectedOrgId">
            <v-card-title class="d-flex align-center justify-space-between">
              <span>Add Credits to User</span>
            </v-card-title>
            <v-card-text>
              <!-- User Selection -->
              <v-autocomplete
                v-model="selectedUserId"
                :items="users"
                item-title="display_name"
                item-value="id"
                label="Select User"
                variant="outlined"
                density="comfortable"
                class="mb-4"
                :loading="loadingUsers"
                @update:search="searchUsers"
              >
                <template v-slot:item="{ props, item }">
                  <v-list-item v-bind="props">
                    <v-list-item-subtitle>{{ item.raw.email }}</v-list-item-subtitle>
                  </v-list-item>
                </template>
              </v-autocomplete>

              <!-- Package Selection -->
              <v-select
                v-model="selectedPackageId"
                :items="packages.filter(p => p.is_active)"
                item-title="name"
                item-value="id"
                label="Select Package (optional)"
                variant="outlined"
                density="comfortable"
                class="mb-4"
                clearable
                @update:model-value="onPackageSelected"
              >
                <template v-slot:item="{ props, item }">
                  <v-list-item v-bind="props">
                    <v-list-item-subtitle>{{ item.raw.credits }} credits - ${{ (item.raw.price_cents / 100).toFixed(2) }}</v-list-item-subtitle>
                  </v-list-item>
                </template>
              </v-select>

              <!-- Manual Credits -->
              <v-text-field
                v-model.number="creditsToAdd"
                label="Credits to Add"
                type="number"
                variant="outlined"
                density="comfortable"
                class="mb-4"
              />

              <!-- Notes -->
              <v-textarea
                v-model="creditNotes"
                label="Notes (optional)"
                variant="outlined"
                density="comfortable"
                rows="2"
                class="mb-4"
              />

              <v-btn color="primary" :disabled="!selectedUserId || creditsToAdd <= 0" @click="purchaseCredits">
                <v-icon start>mdi-plus</v-icon>
                Add Credits
              </v-btn>
            </v-card-text>
          </v-card>
        </v-window-item>
      </v-window>

      <!-- No Organization Selected -->
      <v-alert v-if="!selectedOrgId && !loading" type="info" variant="tonal">
        Please select a gym to manage packages and documents.
      </v-alert>
    </v-container>

    <!-- Package Dialog -->
    <v-dialog v-model="packageDialog" max-width="500">
      <v-card>
        <v-card-title>{{ editingPackage?.id ? 'Edit' : 'Add' }} Package</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="packageForm.name"
            label="Name"
            variant="outlined"
            density="comfortable"
            class="mb-3"
          />
          <v-textarea
            v-model="packageForm.description"
            label="Description"
            variant="outlined"
            density="comfortable"
            rows="2"
            class="mb-3"
          />
          <v-text-field
            v-model.number="packageForm.credits"
            label="Number of Credits"
            type="number"
            variant="outlined"
            density="comfortable"
            class="mb-3"
          />
          <v-text-field
            v-model.number="packageForm.price"
            label="Price ($)"
            type="number"
            step="0.01"
            variant="outlined"
            density="comfortable"
            class="mb-3"
          />
          <v-text-field
            v-model.number="packageForm.validity_days"
            label="Validity (days, leave empty for never expires)"
            type="number"
            variant="outlined"
            density="comfortable"
            class="mb-3"
          />
          <v-switch
            v-model="packageForm.is_active"
            label="Active"
            color="primary"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="packageDialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="savePackage">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Document Dialog -->
    <v-dialog v-model="documentDialog" max-width="500">
      <v-card>
        <v-card-title>{{ editingDocument?.id ? 'Edit' : 'Add' }} Document</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="documentForm.name"
            label="Name"
            variant="outlined"
            density="comfortable"
            class="mb-3"
          />
          <v-textarea
            v-model="documentForm.description"
            label="Description"
            variant="outlined"
            density="comfortable"
            rows="2"
            class="mb-3"
          />
          <v-select
            v-model="documentForm.document_type"
            :items="documentTypes"
            label="Document Type"
            variant="outlined"
            density="comfortable"
            class="mb-3"
          />
          <v-text-field
            v-model="documentForm.url"
            label="Document URL (optional)"
            variant="outlined"
            density="comfortable"
            class="mb-3"
          />
          <v-text-field
            v-model.number="documentForm.expires_after_days"
            label="Expires After (days, leave empty for never)"
            type="number"
            variant="outlined"
            density="comfortable"
            class="mb-3"
          />
          <v-switch
            v-model="documentForm.is_required"
            label="Required for new members"
            color="warning"
            class="mb-2"
          />
          <v-switch
            v-model="documentForm.is_active"
            label="Active"
            color="primary"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="documentDialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="saveDocument">Save</v-btn>
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
const activeTab = ref('packages')

// Packages
const packages = ref([])
const packageDialog = ref(false)
const editingPackage = ref(null)
const packageForm = ref({
  name: '',
  description: '',
  credits: 10,
  price: 0,
  validity_days: null,
  is_active: true
})

// Documents
const documents = ref([])
const documentDialog = ref(false)
const editingDocument = ref(null)
const documentForm = ref({
  name: '',
  description: '',
  document_type: 'waiver',
  url: '',
  expires_after_days: null,
  is_required: false,
  is_active: true
})

const documentTypes = [
  { title: 'Waiver', value: 'waiver' },
  { title: 'Liability', value: 'liability' },
  { title: 'Health Form', value: 'health_form' },
  { title: 'Emergency Contact', value: 'emergency_contact' },
  { title: 'Other', value: 'other' }
]

// Credits
const users = ref([])
const loadingUsers = ref(false)
const selectedUserId = ref(null)
const selectedPackageId = ref(null)
const creditsToAdd = ref(0)
const creditNotes = ref('')
let searchTimeout = null

onMounted(() => {
  fetchOrganizations()
})

async function fetchOrganizations() {
  try {
    const response = await axios.get('/api/admin/organizations')
    organizations.value = response.data.organizations || []

    if (organizations.value.length > 0 && !selectedOrgId.value) {
      selectedOrgId.value = organizations.value[0].id
      fetchData()
    }
  } catch (err) {
    console.error('Failed to fetch organizations:', err)
  }
}

async function fetchData() {
  if (!selectedOrgId.value) return

  loading.value = true
  error.value = null

  try {
    await Promise.all([fetchPackages(), fetchDocuments()])
  } catch (err) {
    console.error('Failed to fetch data:', err)
    error.value = err.response?.data?.error || 'Failed to fetch data'
  } finally {
    loading.value = false
  }
}

async function fetchPackages() {
  const response = await axios.get(`/api/gyms/${selectedOrgId.value}/packages?include_inactive=true`)
  packages.value = response.data.packages || []
}

async function fetchDocuments() {
  const response = await axios.get(`/api/gyms/${selectedOrgId.value}/documents?include_inactive=true`)
  documents.value = response.data.documents || []
}

function openPackageDialog(pkg = null) {
  editingPackage.value = pkg
  if (pkg) {
    packageForm.value = {
      name: pkg.name,
      description: pkg.description || '',
      credits: pkg.credits,
      price: pkg.price_cents / 100,
      validity_days: pkg.validity_days,
      is_active: pkg.is_active
    }
  } else {
    packageForm.value = {
      name: '',
      description: '',
      credits: 10,
      price: 0,
      validity_days: null,
      is_active: true
    }
  }
  packageDialog.value = true
}

async function savePackage() {
  loading.value = true
  error.value = null

  const data = {
    name: packageForm.value.name,
    description: packageForm.value.description || null,
    credits: packageForm.value.credits,
    price_cents: Math.round(packageForm.value.price * 100),
    validity_days: packageForm.value.validity_days || null,
    is_active: packageForm.value.is_active
  }

  try {
    if (editingPackage.value?.id) {
      await axios.put(`/api/admin/gyms/${selectedOrgId.value}/packages/${editingPackage.value.id}`, data)
      successMessage.value = 'Package updated successfully'
    } else {
      await axios.post(`/api/admin/gyms/${selectedOrgId.value}/packages`, data)
      successMessage.value = 'Package created successfully'
    }
    packageDialog.value = false
    await fetchPackages()
  } catch (err) {
    console.error('Failed to save package:', err)
    error.value = err.response?.data?.error || 'Failed to save package'
  } finally {
    loading.value = false
  }
}

async function deletePackage(pkg) {
  if (!confirm(`Are you sure you want to delete "${pkg.name}"?`)) return

  loading.value = true
  try {
    await axios.delete(`/api/admin/gyms/${selectedOrgId.value}/packages/${pkg.id}`)
    successMessage.value = 'Package deleted successfully'
    await fetchPackages()
  } catch (err) {
    console.error('Failed to delete package:', err)
    error.value = err.response?.data?.error || 'Failed to delete package'
  } finally {
    loading.value = false
  }
}

function openDocumentDialog(doc = null) {
  editingDocument.value = doc
  if (doc) {
    documentForm.value = {
      name: doc.name,
      description: doc.description || '',
      document_type: doc.document_type,
      url: doc.url || '',
      expires_after_days: doc.expires_after_days,
      is_required: doc.is_required,
      is_active: doc.is_active
    }
  } else {
    documentForm.value = {
      name: '',
      description: '',
      document_type: 'waiver',
      url: '',
      expires_after_days: null,
      is_required: false,
      is_active: true
    }
  }
  documentDialog.value = true
}

async function saveDocument() {
  loading.value = true
  error.value = null

  const data = {
    name: documentForm.value.name,
    description: documentForm.value.description || null,
    document_type: documentForm.value.document_type,
    url: documentForm.value.url || null,
    expires_after_days: documentForm.value.expires_after_days || null,
    is_required: documentForm.value.is_required,
    is_active: documentForm.value.is_active
  }

  try {
    if (editingDocument.value?.id) {
      await axios.put(`/api/admin/gyms/${selectedOrgId.value}/documents/${editingDocument.value.id}`, data)
      successMessage.value = 'Document updated successfully'
    } else {
      await axios.post(`/api/admin/gyms/${selectedOrgId.value}/documents`, data)
      successMessage.value = 'Document created successfully'
    }
    documentDialog.value = false
    await fetchDocuments()
  } catch (err) {
    console.error('Failed to save document:', err)
    error.value = err.response?.data?.error || 'Failed to save document'
  } finally {
    loading.value = false
  }
}

async function deleteDocument(doc) {
  if (!confirm(`Are you sure you want to delete "${doc.name}"?`)) return

  loading.value = true
  try {
    await axios.delete(`/api/admin/gyms/${selectedOrgId.value}/documents/${doc.id}`)
    successMessage.value = 'Document deleted successfully'
    await fetchDocuments()
  } catch (err) {
    console.error('Failed to delete document:', err)
    error.value = err.response?.data?.error || 'Failed to delete document'
  } finally {
    loading.value = false
  }
}

// Credits functions
async function searchUsers(query) {
  if (!query || query.length < 2) {
    users.value = []
    return
  }

  // Debounce the search
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(async () => {
    loadingUsers.value = true
    try {
      const response = await axios.get(`/api/admin/users?search=${encodeURIComponent(query)}&limit=20`)
      users.value = (response.data.users || []).map(u => ({
        ...u,
        display_name: `${u.name || u.email} (${u.email})`
      }))
    } catch (err) {
      console.error('Failed to search users:', err)
    } finally {
      loadingUsers.value = false
    }
  }, 300)
}

function onPackageSelected(packageId) {
  if (packageId) {
    const pkg = packages.value.find(p => p.id === packageId)
    if (pkg) {
      creditsToAdd.value = pkg.credits
    }
  }
}

async function purchaseCredits() {
  if (!selectedUserId.value || creditsToAdd.value <= 0) return

  loading.value = true
  error.value = null

  try {
    await axios.post(`/api/admin/gyms/${selectedOrgId.value}/users/${selectedUserId.value}/credits`, {
      package_id: selectedPackageId.value || null,
      credits: creditsToAdd.value,
      notes: creditNotes.value || ''
    })
    successMessage.value = `Successfully added ${creditsToAdd.value} credits`

    // Reset form
    selectedUserId.value = null
    selectedPackageId.value = null
    creditsToAdd.value = 0
    creditNotes.value = ''
    users.value = []
  } catch (err) {
    console.error('Failed to purchase credits:', err)
    error.value = err.response?.data?.error || 'Failed to add credits'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.mobile-view-wrapper {
  max-width: 800px;
  margin: 0 auto;
}
</style>
