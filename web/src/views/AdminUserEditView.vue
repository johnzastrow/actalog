<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AdminHeader from '@/components/AdminHeader.vue'
import ProfileTab from '@/components/admin/user-edit/ProfileTab.vue'
// Future: AffiliationsTab, SubscriptionsTab, CreditsTab, PreferencesTab, ActivityTab

const route = useRoute()
const router = useRouter()

const userId = computed(() => Number(route.params.id))
const activeTab = ref(route.query.tab || 'profile')

onMounted(() => {
  if (!Number.isFinite(userId.value) || userId.value <= 0) {
    router.replace('/admin/users')
  }
})

watch(activeTab, (t) => {
  router.replace({ query: { ...route.query, tab: t } })
})
</script>

<template>
  <div class="mobile-view-wrapper">
    <v-container fluid class="pa-4">
      <AdminHeader
        title="Edit User"
        :breadcrumbs="[
          { title: 'Users', to: '/admin/users' },
          { title: 'Edit', to: route.fullPath },
        ]"
      />

      <v-card elevation="0" rounded="lg" class="mt-4">
        <v-tabs v-model="activeTab">
          <v-tab value="profile">Profile</v-tab>
          <v-tab value="affiliations" disabled>
            Affiliations <v-chip size="x-small" class="ml-1">v1.3.1</v-chip>
          </v-tab>
          <v-tab value="subscriptions" disabled>
            Subscriptions <v-chip size="x-small" class="ml-1">v1.3.2</v-chip>
          </v-tab>
          <v-tab value="credits" disabled>
            Credits <v-chip size="x-small" class="ml-1">v1.3.2</v-chip>
          </v-tab>
          <v-tab value="preferences" disabled>
            Preferences <v-chip size="x-small" class="ml-1">v1.3.2</v-chip>
          </v-tab>
          <v-tab value="activity" disabled>
            Activity <v-chip size="x-small" class="ml-1">v1.3.2</v-chip>
          </v-tab>
        </v-tabs>
        <v-divider />
        <v-window v-model="activeTab">
          <v-window-item value="profile" class="pa-4">
            <ProfileTab :user-id="userId" />
          </v-window-item>
          <v-window-item value="affiliations" class="pa-4">
            <v-alert type="info" variant="tonal">
              Coming in v1.3.1 — gym memberships and per-gym coach assignments.
            </v-alert>
          </v-window-item>
          <v-window-item value="subscriptions" class="pa-4">
            <v-alert type="info" variant="tonal">
              Coming in v1.3.2 — subscription details and billing actions.
            </v-alert>
          </v-window-item>
          <v-window-item value="credits" class="pa-4">
            <v-alert type="info" variant="tonal">
              Coming in v1.3.2 — class credits and document completion.
            </v-alert>
          </v-window-item>
          <v-window-item value="preferences" class="pa-4">
            <v-alert type="info" variant="tonal">
              Coming in v1.3.2 — theme, font, and notification preferences.
            </v-alert>
          </v-window-item>
          <v-window-item value="activity" class="pa-4">
            <v-alert type="info" variant="tonal">
              Coming in v1.3.2 — workout history and audit-log activity.
            </v-alert>
          </v-window-item>
        </v-window>
      </v-card>
    </v-container>
  </div>
</template>

<style scoped>
.mobile-view-wrapper {
  max-width: 800px;
  margin: 0 auto;
}
</style>
