<template>
  <div class="admin-header mb-4">
    <!-- Breadcrumbs -->
    <v-breadcrumbs :items="breadcrumbItems" class="pa-0 mb-2">
      <template #divider>
        <v-icon size="small">mdi-chevron-right</v-icon>
      </template>
      <template #item="{ item }">
        <v-breadcrumbs-item
          :to="item.to"
          :disabled="item.disabled"
          class="text-body-2"
        >
          {{ item.title }}
        </v-breadcrumbs-item>
      </template>
    </v-breadcrumbs>

    <!-- Title and User Info Row -->
    <div class="d-flex align-center justify-space-between flex-wrap">
      <div>
        <h1 class="admin-title">{{ title }}</h1>
        <p v-if="subtitle" class="admin-subtitle">{{ subtitle }}</p>
      </div>

      <!-- User Info Badge -->
      <div class="user-info-badge">
        <v-chip
          :color="roleColor"
          variant="tonal"
          size="small"
          class="mr-2"
        >
          <v-icon start size="small">mdi-shield-account</v-icon>
          {{ userRole }}
        </v-chip>
        <span class="user-email">{{ userEmail }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'

const props = defineProps({
  title: {
    type: String,
    required: true
  },
  subtitle: {
    type: String,
    default: ''
  },
  breadcrumbs: {
    type: Array,
    default: () => []
  }
})

const authStore = useAuthStore()

const userEmail = computed(() => authStore.user?.email || 'Unknown')
const userRole = computed(() => {
  const role = authStore.user?.role || 'athlete'
  return role.charAt(0).toUpperCase() + role.slice(1)
})

const roleColor = computed(() => {
  const role = authStore.user?.role
  if (role === 'admin') return 'error'
  if (role === 'coach') return 'warning'
  return 'info'
})

const breadcrumbItems = computed(() => {
  const items = [
    { title: 'Admin', to: '/admin', disabled: false }
  ]

  props.breadcrumbs.forEach((crumb, index) => {
    items.push({
      title: crumb.title,
      to: crumb.to,
      disabled: index === props.breadcrumbs.length - 1
    })
  })

  return items
})
</script>

<style scoped>
.admin-header {
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
  padding-bottom: 12px;
}

.admin-title {
  color: #2c3e50;
  font-size: 24px;
  font-weight: 600;
  margin: 0;
}

.admin-subtitle {
  color: rgba(0, 0, 0, 0.6);
  font-size: 14px;
  margin: 4px 0 0 0;
}

.user-info-badge {
  display: flex;
  align-items: center;
  background: rgba(0, 0, 0, 0.03);
  padding: 6px 12px;
  border-radius: 8px;
}

.user-email {
  font-size: 13px;
  color: #546e7a;
}

@media (max-width: 600px) {
  .user-info-badge {
    margin-top: 12px;
    width: 100%;
    justify-content: center;
  }
}
</style>
