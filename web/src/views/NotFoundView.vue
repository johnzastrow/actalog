<template>
  <v-container class="fill-height">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="6" class="text-center">
        <v-icon size="120" color="primary">mdi-alert-circle-outline</v-icon>
        <h1 class="text-h2 mt-4 mb-2">404</h1>
        <h2 class="text-h5 text-medium-emphasis mb-6">Page Not Found</h2>
        <p class="text-body-1 mb-8">
          The page you're looking for doesn't exist or has been moved.
        </p>
        <v-btn color="primary" size="large" to="/dashboard">
          <v-icon start>mdi-home</v-icon>
          Go to Dashboard
        </v-btn>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
// No script needed for this simple component
</script>
