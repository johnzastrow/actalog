<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <v-btn icon="mdi-arrow-left" variant="text" @click="$router.back()" />
        <h1 class="text-h4 mb-6 d-inline ml-2">Settings</h1>
      </v-col>

      <v-col cols="12">
        <v-card elevation="2" rounded="lg" class="mb-4">
          <v-card-title>Preferences</v-card-title>
          <v-card-text>
            <v-switch
              v-model="darkMode"
              label="Dark Mode"
              color="primary"
            />
            <v-switch
              v-model="notifications"
              label="Notifications"
              color="primary"
            />
          </v-card-text>
        </v-card>

        <v-card elevation="2" rounded="lg" class="mb-4">
          <v-card-title>Data</v-card-title>
          <v-card-text>
            <v-list>
              <v-list-item prepend-icon="mdi-download">
                <v-list-item-title>Export Data</v-list-item-title>
              </v-list-item>
              <v-list-item prepend-icon="mdi-upload">
                <v-list-item-title>Import Data</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>

        <v-card elevation="2" rounded="lg">
          <v-card-title>Account</v-card-title>
          <v-card-text>
            <v-list>
              <v-list-item prepend-icon="mdi-delete" class="text-error">
                <v-list-item-title>Delete Account</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'

const darkMode = ref(false)
const notifications = ref(true)
</script>
