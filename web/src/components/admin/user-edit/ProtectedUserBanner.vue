<template>
  <v-alert
    type="info"
    prominent
    class="my-4"
    variant="tonal"
  >
    <template #prepend>
      <v-icon size="36">mdi-shield-account</v-icon>
    </template>
    <strong>This is a system-reserved account.</strong>
    <div class="mt-2">
      All write actions are disabled by design. You can still view profile,
      affiliations, audit logs, and activity, but no edits are possible.
    </div>
    <div class="mt-3">
      <v-btn
        variant="text"
        size="small"
        href="/docs/protected-users"
        target="_blank"
        prepend-icon="mdi-book-open-variant"
      >
        Read more about protected users
      </v-btn>
    </div>
  </v-alert>
</template>

<script setup>
// Purely presentational — no props, no logic, no external state.
</script>
