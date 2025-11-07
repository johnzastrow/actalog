<template>
  <v-container>
    <!-- Header -->
    <v-row>
      <v-col cols="12">
        <v-app-bar color="primary" density="compact" flat>
          <v-btn icon="mdi-arrow-left" @click="$router.back()" />
          <v-toolbar-title>Log Workout</v-toolbar-title>
          <v-btn icon="mdi-dots-vertical" />
        </v-app-bar>
      </v-col>
    </v-row>

    <v-row class="mt-2">
      <v-col cols="12">
        <v-card elevation="2" rounded="lg">
          <v-card-text class="pa-6">
            <v-form>
              <!-- Workout Date -->
              <v-text-field
                v-model="workoutDate"
                label="Workout Date"
                type="date"
                prepend-inner-icon="mdi-calendar"
                variant="outlined"
              />

              <!-- Workout Type -->
              <div class="my-4">
                <label class="text-body-2 text-medium-emphasis mb-2 d-block">
                  Workout Type
                </label>
                <v-btn-toggle
                  v-model="workoutType"
                  color="primary"
                  variant="outlined"
                  divided
                  mandatory
                >
                  <v-btn value="named" prepend-icon="mdi-format-list-bulleted">
                    Named WOD
                  </v-btn>
                  <v-btn value="custom" prepend-icon="mdi-pencil">
                    Custom
                  </v-btn>
                </v-btn-toggle>
              </div>

              <!-- Select Movement -->
              <v-select
                v-model="selectedMovement"
                label="Select Movement"
                :items="movements"
                prepend-inner-icon="mdi-dumbbell"
                variant="outlined"
                class="mt-4"
              />

              <v-btn
                variant="outlined"
                color="accent"
                block
                class="mt-2"
                prepend-icon="mdi-plus"
              >
                Add Custom Movement
              </v-btn>

              <!-- Movement Details -->
              <v-row class="mt-4">
                <v-col cols="6">
                  <v-text-field
                    v-model="weight"
                    label="Weight (lbs)"
                    type="number"
                    prepend-inner-icon="mdi-weight"
                    variant="outlined"
                  />
                </v-col>
                <v-col cols="6">
                  <v-text-field
                    v-model="sets"
                    label="Sets"
                    type="number"
                    prepend-inner-icon="mdi-format-list-numbered"
                    variant="outlined"
                  />
                </v-col>
              </v-row>

              <v-row>
                <v-col cols="6">
                  <v-text-field
                    v-model="reps"
                    label="Reps"
                    type="number"
                    prepend-inner-icon="mdi-numeric"
                    variant="outlined"
                  />
                </v-col>
                <v-col cols="6">
                  <v-btn-toggle
                    v-model="movementType"
                    color="accent"
                    variant="outlined"
                    divided
                    mandatory
                  >
                    <v-btn value="rx">Rx</v-btn>
                    <v-btn value="scaled">Scaled</v-btn>
                  </v-btn-toggle>
                </v-col>
              </v-row>

              <!-- Notes -->
              <v-textarea
                v-model="notes"
                label="Notes (Optional)"
                placeholder="How did it feel? Any observations..."
                variant="outlined"
                rows="3"
                class="mt-4"
              />

              <!-- Action Buttons -->
              <v-row class="mt-4">
                <v-col cols="12">
                  <v-btn color="primary" size="large" block>
                    <v-icon start>mdi-content-save</v-icon>
                    Save Workout
                  </v-btn>
                </v-col>
                <v-col cols="6">
                  <v-btn variant="outlined" size="large" block>
                    <v-icon start>mdi-refresh</v-icon>
                    Reset
                  </v-btn>
                </v-col>
                <v-col cols="6">
                  <v-btn variant="outlined" size="large" block @click="$router.back()">
                    <v-icon start>mdi-close</v-icon>
                    Cancel
                  </v-btn>
                </v-col>
              </v-row>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'

const workoutDate = ref(new Date().toISOString().split('T')[0])
const workoutType = ref('named')
const selectedMovement = ref(null)
const movements = ref(['Back Squat', 'Deadlift', 'Bench Press', 'Clean', 'Snatch'])
const weight = ref(null)
const sets = ref(null)
const reps = ref(null)
const movementType = ref('rx')
const notes = ref('')
</script>
