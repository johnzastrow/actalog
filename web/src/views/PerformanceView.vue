<template>
  <div class="mobile-view-wrapper">
    <v-container class="pa-3">
      <!-- Error Alert -->
      <v-alert v-if="error" type="error" closable @click:close="error = null" class="mb-3">
        {{ error }}
      </v-alert>

      <!-- Unified Search (Movements + WODs) -->
      <v-card elevation="2" class="pa-3 mb-3" style="background: white">
        <v-autocomplete
          v-model="selectedItem"
          :items="searchResults"
          item-title="name"
          item-value="id"
          :loading="loadingSearch"
          placeholder="Search for a WOD or Movement..."
          variant="outlined"
          density="comfortable"
          clearable
          auto-select-first
          hide-details
          return-object
          @update:search="handleSearch"
          @update:model-value="handleSelection"
        >
          <template #prepend-inner>
            <v-icon color="#00bcd4" size="small">mdi-magnify</v-icon>
          </template>
          <template #item="{ props, item }">
            <v-list-item v-bind="props" density="compact">
              <template #prepend>
                <v-icon
                  :color="item.raw.type === 'movement' ? '#00bcd4' : 'teal'"
                  size="small"
                >
                  {{ item.raw.type === 'movement' ? 'mdi-dumbbell' : 'mdi-fire' }}
                </v-icon>
              </template>
              <v-list-item-title class="text-body-2">
                {{ item.raw.name }}
              </v-list-item-title>
              <v-list-item-subtitle class="text-caption">
                {{ item.raw.type === 'movement' ? formatMovementType(item.raw.data?.type) : item.raw.data?.type || 'WOD' }}
              </v-list-item-subtitle>
            </v-list-item>
          </template>
        </v-autocomplete>
      </v-card>

      <!-- Empty State - Prompt to Search -->
      <v-card v-if="!selectedItem" elevation="0" rounded="lg" class="pa-6 mb-3 text-center" style="background: white">
        <v-icon size="64" color="#ccc">mdi-chart-line-variant</v-icon>
        <h2 class="text-h6 font-weight-bold mt-3" style="color: #1a1a1a">Track Your Performance</h2>
        <p class="text-body-2 mt-2" style="color: #666">
          Search for a movement or WOD above to view your progress, PRs, and performance history
        </p>
      </v-card>

      <!-- Performance Content (Movement or WOD) -->
      <div v-if="selectedItem">
        <!-- Quick Log Button -->
        <v-btn
          block
          size="large"
          color="teal"
          rounded="lg"
          elevation="2"
          class="mb-3 font-weight-bold"
          style="text-transform: none"
          @click="quickLog"
        >
          <v-icon start>mdi-lightning-bolt</v-icon>
          Quick Log {{ selectedItem.name }}
        </v-btn>

        <!-- MOVEMENT-SPECIFIC CONTENT -->
        <template v-if="selectedItem.type === 'movement'">
          <!-- Heaviest Lifts (Top 3 Maxes) -->
          <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
            <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">
              <v-icon color="teal" size="small" class="mr-1">mdi-trophy</v-icon>
              Heaviest Lifts
            </h2>

            <div v-if="loadingPerformance" class="text-center py-4">
              <v-progress-circular indeterminate color="#00bcd4" size="32" />
            </div>

            <div v-else-if="heaviestLifts.length === 0" class="text-center py-4">
              <p class="text-caption" style="color: #999">No performance data yet</p>
            </div>

            <div v-else>
              <v-row>
                <v-col v-for="(lift, index) in heaviestLifts" :key="index" cols="4">
                  <div class="text-center">
                    <v-chip
                      :color="index === 0 ? '#ffc107' : index === 1 ? '#9e9e9e' : '#cd7f32'"
                      size="small"
                      class="mb-2"
                      label
                    >
                      #{{ index + 1 }}
                    </v-chip>
                    <div class="font-weight-bold text-h6" style="color: #1a1a1a">
                      {{ lift.weight }}
                    </div>
                    <div class="text-caption" style="color: #666">
                      lbs
                    </div>
                    <div v-if="lift.reps" class="text-caption" style="color: #999">
                      {{ lift.reps }} reps
                    </div>
                  </div>
                </v-col>
              </v-row>
            </div>
          </v-card>

          <!-- Best Estimated 1RM -->
          <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
            <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">
              <v-icon color="#ffc107" size="small" class="mr-1">mdi-arm-flex</v-icon>
              Best Estimated 1RM
            </h2>

            <div v-if="loadingPerformance" class="text-center py-4">
              <v-progress-circular indeterminate color="#00bcd4" size="32" />
            </div>

            <div v-else-if="!best1RM" class="text-center py-4">
              <p class="text-caption" style="color: #999">No weight/reps data available</p>
            </div>

            <div v-else class="text-center">
              <div class="font-weight-bold text-h4" style="color: #ffc107">
                {{ Math.round(best1RM) }}
              </div>
              <div class="text-caption mb-2" style="color: #666">
                lbs (estimated)
              </div>
              <v-chip
                v-if="bestFormula"
                size="x-small"
                color="grey-lighten-2"
                label
                class="text-caption"
              >
                {{ bestFormula }}
              </v-chip>
            </div>
          </v-card>

          <!-- Rep Scheme Dropdown Filter -->
          <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
            <v-select
              v-model="selectedRepScheme"
              :items="repSchemes"
              label="Filter by Rep Scheme"
              variant="outlined"
              density="comfortable"
              rounded="lg"
              hide-details
              @update:model-value="filterChart"
            >
              <template #prepend-inner>
                <v-icon color="#00bcd4" size="small">mdi-filter</v-icon>
              </template>
            </v-select>
          </v-card>

          <!-- Performance Chart -->
          <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
            <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Performance Chart</h2>

            <div v-if="loadingPerformance" class="text-center py-4">
              <v-progress-circular indeterminate color="#00bcd4" size="32" />
            </div>

            <div v-else-if="filteredChartData.length === 0" class="text-center py-4">
              <p class="text-caption" style="color: #999">No data for selected filter</p>
            </div>

            <div v-else style="height: 250px; position: relative; width: 100%">
              <canvas ref="chartCanvas" style="width: 100%; height: 100%"></canvas>
            </div>
          </v-card>
        </template>

        <!-- WOD-SPECIFIC CONTENT -->
        <template v-if="selectedItem.type === 'wod'">
          <!-- Best WOD Performances (Top 3) -->
          <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
            <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">
              <v-icon color="teal" size="small" class="mr-1">mdi-trophy</v-icon>
              Best Performances
            </h2>

            <div v-if="loadingPerformance" class="text-center py-4">
              <v-progress-circular indeterminate color="#00bcd4" size="32" />
            </div>

            <div v-else-if="bestWODPerformances.length === 0" class="text-center py-4">
              <p class="text-caption" style="color: #999">No performance data yet</p>
            </div>

            <div v-else>
              <v-row>
                <v-col v-for="(perf, index) in bestWODPerformances" :key="index" cols="4">
                  <div class="text-center">
                    <v-chip
                      :color="index === 0 ? '#ffc107' : index === 1 ? '#9e9e9e' : '#cd7f32'"
                      size="small"
                      class="mb-2"
                      label
                    >
                      #{{ index + 1 }}
                    </v-chip>
                    <div class="font-weight-bold text-body-2" style="color: #1a1a1a">
                      {{ formatWODScore(perf) }}
                    </div>
                    <div class="text-caption" style="color: #666">
                      {{ formatDate(perf.workout_date) }}
                    </div>
                  </div>
                </v-col>
              </v-row>
            </div>
          </v-card>

          <!-- WOD Performance Chart -->
          <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
            <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Performance Chart</h2>

            <div v-if="loadingPerformance" class="text-center py-4">
              <v-progress-circular indeterminate color="#00bcd4" size="32" />
            </div>

            <div v-else-if="wodPerformanceData.length === 0" class="text-center py-4">
              <p class="text-caption" style="color: #999">No performance data yet</p>
            </div>

            <div v-else style="height: 250px; position: relative; width: 100%">
              <canvas ref="wodChartCanvas" style="width: 100%; height: 100%"></canvas>
            </div>
          </v-card>
        </template>

        <!-- Performance History (Grouped by Year) -->
        <v-card elevation="0" rounded="lg" class="pa-3 mb-3" style="background: white">
          <h2 class="text-body-1 font-weight-bold mb-3" style="color: #1a1a1a">Performance History</h2>

          <div v-if="loadingPerformance" class="text-center py-4">
            <v-progress-circular indeterminate color="#00bcd4" size="32" />
          </div>

          <div v-else-if="Object.keys(groupedHistory).length === 0" class="text-center py-4">
            <v-icon size="48" color="#ccc">mdi-history</v-icon>
            <p class="text-body-2 mt-2" style="color: #666">No history yet</p>
            <p class="text-caption" style="color: #999">
              Start logging workouts with {{ selectedItem.name }} to track your progress
            </p>
          </div>

          <!-- History Grouped by Year -->
          <div v-else>
            <div v-for="(entries, year) in groupedHistory" :key="year" class="mb-4">
              <v-chip size="small" color="#00bcd4" label class="mb-2">
                {{ year }}
              </v-chip>

              <v-card
                v-for="(entry, index) in entries"
                :key="index"
                elevation="0"
                rounded="lg"
                class="mb-2 pa-2"
                style="background: #f5f7fa; cursor: pointer; transition: all 0.2s ease"
                hover
                @click="viewWorkout(entry.user_workout_id)"
                @mouseenter="$event.currentTarget.style.background = '#e8f4f8'"
                @mouseleave="$event.currentTarget.style.background = '#f5f7fa'"
              >
                <div class="d-flex align-center">
                  <!-- PR Trophy Icon -->
                  <v-icon v-if="entry.is_pr" color="teal" size="small" class="mr-2">mdi-trophy</v-icon>

                  <div class="flex-grow-1">
                    <!-- Movement Performance Display -->
                    <div v-if="selectedItem.type === 'movement'" class="font-weight-bold text-body-2" style="color: #1a1a1a">
                      <span v-if="entry.sets && entry.reps && entry.weight">
                        {{ entry.sets }} × {{ entry.reps }} @ {{ entry.weight }} lbs
                      </span>
                      <span v-else-if="entry.weight">
                        {{ entry.weight }} lbs
                        <span v-if="entry.reps"> × {{ entry.reps }}</span>
                      </span>
                      <span v-else-if="entry.time">
                        {{ formatTime(entry.time) }}
                      </span>
                    </div>

                    <!-- WOD Performance Display -->
                    <div v-if="selectedItem.type === 'wod'" class="font-weight-bold text-body-2" style="color: #1a1a1a">
                      {{ formatWODScore(entry) }}
                    </div>

                    <div class="text-caption" style="color: #666">
                      {{ formatDate(entry.workout_date) }}
                      <span v-if="entry.notes"> • {{ entry.notes }}</span>
                      <span v-if="entry.calculated_1rm" class="ml-2" style="color: #ffc107">
                        • Est. 1RM: {{ Math.round(entry.calculated_1rm) }} lbs
                      </span>
                    </div>
                  </div>

                  <!-- PR Badge -->
                  <v-chip
                    v-if="entry.is_pr"
                    size="x-small"
                    color="teal"
                    class="ml-2"
                  >
                    PR
                  </v-chip>
                </div>
              </v-card>
            </div>
          </div>
        </v-card>
      </div>
    </v-container>
    <!-- Quick Log Dialog -->
    <v-dialog v-model="quickLogDialog" max-width="500px">
      <v-card>
        <v-card-title class="text-h6 font-weight-bold" style="background: #00bcd4; color: white">
          <v-icon color="white" class="mr-2">mdi-lightning-bolt</v-icon>
          Quick Log Workout
        </v-card-title>

        <v-card-text class="pa-2">
          <v-form ref="quickLogForm" @submit.prevent="submitQuickLog">
            <!-- Date -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" style="color: #1a1a1a">
                Date *
              </label>
              <v-text-field
                v-model="quickLogData.date"
                type="date"
                variant="outlined"
                density="compact"
                hide-details
                required
                @update:model-value="updateQuickLogName"
              />
            </div>

            <!-- Workout Name -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" style="color: #1a1a1a">
                Workout Name *
              </label>
              <v-text-field
                v-model="quickLogData.name"
                variant="outlined"
                density="compact"
                placeholder="e.g., Morning Run, Upper Body, etc."
                hide-details
                required
              />
            </div>

            <!-- Total Time -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" style="color: #1a1a1a">
                Total Time (minutes)
              </label>
              <v-text-field
                v-model.number="quickLogData.totalTime"
                type="number"
                variant="outlined"
                density="compact"
                placeholder="e.g., 30"
                hide-details
                min="0"
              />
            </div>

            <!-- Notes -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" style="color: #1a1a1a">
                Notes
              </label>
              <v-textarea
                v-model="quickLogData.notes"
                variant="outlined"
                density="compact"
                rows="3"
                placeholder="How did it feel? Any highlights?"
                hide-details
              />
            </div>

            <!-- Unified Search for Movement/WOD -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" style="color: #1a1a1a">
                Add Performance Data (Optional)
              </label>
              <v-autocomplete
                v-model="quickLogData.selectedItem"
                :items="unifiedSearchItems"
                item-title="displayName"
                item-value="id"
                return-object
                :loading="loadingMovements || loadingWods"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                auto-select-first
                placeholder="Search for movement or WOD..."
              >
                <template #prepend-inner>
                  <v-icon color="#00bcd4" size="small">mdi-magnify</v-icon>
                </template>
                <template #item="{ props, item }">
                  <v-list-item v-bind="props">
                    <template #prepend>
                      <v-icon
                        :color="item.raw.type === 'movement' ? '#00bcd4' : (item.raw.type === 'wod' ? 'teal' : '#9c27b0')"
                        size="small"
                      >
                        {{ item.raw.type === 'movement' ? 'mdi-weight-lifter' : (item.raw.type === 'wod' ? 'mdi-fire' : 'mdi-clipboard-text') }}
                      </v-icon>
                    </template>
                    <template #append>
                      <v-chip
                        :color="item.raw.type === 'movement' ? '#00bcd4' : (item.raw.type === 'wod' ? 'teal' : '#9c27b0')"
                        size="x-small"
                        variant="flat"
                        class="text-uppercase"
                      >
                        {{ item.raw.type }}
                      </v-chip>
                    </template>
                  </v-list-item>
                </template>
              </v-autocomplete>
            </div>

            <!-- Template Info Display -->
            <div v-if="quickLogData.selectedItem && quickLogData.selectedItem.type === 'template'" class="mt-3 pa-3" style="background: #f3e5f5; border-radius: 8px; border: 1px solid #9c27b0">
              <div class="d-flex align-center mb-2">
                <v-icon color="#9c27b0" size="small" class="mr-2">mdi-clipboard-text</v-icon>
                <span class="text-caption font-weight-bold" style="color: #9c27b0">Workout Template Selected</span>
              </div>
              <p class="text-caption mb-0" style="color: #666">
                This will create a workout based on the "{{ quickLogData.selectedItem.name }}" template.
                Template details will be included automatically.
              </p>
            </div>

              <!-- Movement Performance Form -->
              <div v-if="quickLogData.selectedItem && quickLogData.selectedItem.type === 'movement'" class="mt-3 pa-3" style="background: #f5f5f5; border-radius: 8px">
                <div class="mb-2">
                  <label class="text-caption">Sets</label>
                  <v-text-field
                    v-model.number="quickLogData.movement.sets"
                    type="number"
                    variant="outlined"
                    density="compact"
                    hide-details
                    min="0"
                  />
                </div>
                <div class="mb-2">
                  <label class="text-caption">Reps</label>
                  <v-text-field
                    v-model.number="quickLogData.movement.reps"
                    type="number"
                    variant="outlined"
                    density="compact"
                    hide-details
                    min="0"
                  />
                </div>
                <div class="mb-2">
                  <label class="text-caption">Weight (lbs)</label>
                  <v-text-field
                    v-model.number="quickLogData.movement.weight"
                    type="number"
                    variant="outlined"
                    density="compact"
                    hide-details
                    min="0"
                    step="0.1"
                  />
                </div>
                <div class="mb-2">
                  <label class="text-caption">Time (seconds)</label>
                  <v-text-field
                    v-model.number="quickLogData.movement.time"
                    type="number"
                    variant="outlined"
                    density="compact"
                    hide-details
                    min="0"
                  />
                </div>
                <div class="mb-2">
                  <label class="text-caption">Distance (meters)</label>
                  <v-text-field
                    v-model.number="quickLogData.movement.distance"
                    type="number"
                    variant="outlined"
                    density="compact"
                    hide-details
                    min="0"
                    step="0.1"
                  />
                </div>
                <div>
                  <label class="text-caption">Notes</label>
                  <v-textarea
                    v-model="quickLogData.movement.notes"
                    variant="outlined"
                    density="compact"
                    rows="2"
                    hide-details
                  />
                </div>
              </div>

              <!-- WOD Performance Form -->
              <div v-if="quickLogData.selectedItem && quickLogData.selectedItem.type === 'wod'" class="mt-3 pa-3" style="background: #f5f5f5; border-radius: 8px">
                <div class="mb-2">
                  <label class="text-caption">Score Type (from WOD)</label>
                  <v-text-field
                    v-model="quickLogData.wod.scoreType"
                    variant="outlined"
                    density="compact"
                    hide-details
                    readonly
                    bg-color="#e0e0e0"
                  />
                </div>
                <!-- Time-based WOD fields -->
                <div v-if="quickLogData.wod.scoreType === 'Time (HH:MM:SS)'">
                  <label class="text-caption d-block mb-1">Time (HH:MM:SS) *</label>
                  <div class="d-flex gap-2 mb-2">
                    <v-text-field
                      v-model.number="quickLogData.wod.timeHours"
                      type="number"
                      variant="outlined"
                      density="compact"
                      hide-details
                      min="0"
                      max="23"
                      placeholder="HH"
                      style="flex: 1"
                    />
                    <span class="align-self-center">:</span>
                    <v-text-field
                      v-model.number="quickLogData.wod.timeMinutes"
                      type="number"
                      variant="outlined"
                      density="compact"
                      hide-details
                      min="0"
                      max="59"
                      placeholder="MM"
                      style="flex: 1"
                    />
                    <span class="align-self-center">:</span>
                    <v-text-field
                      v-model.number="quickLogData.wod.timeSecondsInput"
                      type="number"
                      variant="outlined"
                      density="compact"
                      hide-details
                      min="0"
                      max="59"
                      placeholder="SS"
                      style="flex: 1"
                    />
                  </div>
                </div>

                <!-- Rounds+Reps WOD fields -->
                <template v-if="quickLogData.wod.scoreType === 'Rounds+Reps'">
                  <div class="mb-2">
                    <label class="text-caption">Rounds *</label>
                    <v-text-field
                      v-model.number="quickLogData.wod.rounds"
                      type="number"
                      variant="outlined"
                      density="compact"
                      hide-details
                      min="0"
                      placeholder="e.g., 10"
                    />
                  </div>
                  <div class="mb-2">
                    <label class="text-caption">Reps (optional)</label>
                    <v-text-field
                      v-model.number="quickLogData.wod.reps"
                      type="number"
                      variant="outlined"
                      density="compact"
                      hide-details
                      min="0"
                      placeholder="e.g., 15"
                    />
                  </div>
                </template>

                <!-- Max Weight WOD field (note: weight field is missing in quickLogData.wod, needs to be added) -->
                <div v-if="quickLogData.wod.scoreType === 'Max Weight'" class="mb-2">
                  <label class="text-caption">Weight (lbs) *</label>
                  <v-text-field
                    v-model.number="quickLogData.wod.weight"
                    type="number"
                    variant="outlined"
                    density="compact"
                    hide-details
                    min="0"
                    step="0.5"
                    placeholder="e.g., 225"
                  />
                </div>

                <!-- Notes field (always shown) -->
                <div>
                  <label class="text-caption">Notes</label>
                  <v-textarea
                    v-model="quickLogData.wod.notes"
                    variant="outlined"
                    density="compact"
                    rows="2"
                    hide-details
                    placeholder="How did it feel?"
                  />
                </div>
              </div>
          </v-form>
        </v-card-text>

        <v-card-actions class="pa-2 pt-0">
          <v-btn
            variant="text"
            @click="closeQuickLog"
          >
            Cancel
          </v-btn>
          <v-spacer />
          <v-btn
            color="teal"
            variant="elevated"
            :loading="quickLogSubmitting"
            :disabled="!quickLogData.name || !quickLogData.date"
            @click="submitQuickLog"
          >
            <v-icon start>mdi-check</v-icon>
            Log Workout
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Bottom Navigation -->
    <v-bottom-navigation
      v-model="activeTab"
      grow
      style="position: fixed; bottom: 0; background: white"
      elevation="8"
    >
      <v-btn value="dashboard" to="/dashboard">
        <v-icon>mdi-view-dashboard</v-icon>
        <span style="font-size: 10px">Dashboard</span>
      </v-btn>
      <v-btn value="performance" to="/performance">
        <v-icon>mdi-chart-line</v-icon>
        <span style="font-size: 10px">Performance</span>
      </v-btn>
      <v-btn value="log" to="/dashboard?open=quick-log" style="position: relative; bottom: 20px">
        <v-avatar color="teal" size="56" style="box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2)">
          <v-icon color="white" size="32">mdi-plus</v-icon>
        </v-avatar>
      </v-btn>
      <v-btn value="workouts" to="/workouts">
        <v-icon>mdi-dumbbell</v-icon>
        <span style="font-size: 10px">Templates</span>
      </v-btn>
      <v-btn value="profile" to="/profile">
        <v-icon>mdi-account</v-icon>
        <span style="font-size: 10px">Profile</span>
      </v-btn>
    </v-bottom-navigation>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onBeforeUnmount, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/axios'
import { Chart, registerables } from 'chart.js'

Chart.register(...registerables)

const router = useRouter()
const route = useRoute()
const activeTab = ref('performance')

// State
const selectedItem = ref(null) // { type: 'movement' | 'wod', id, name, data }
const searchQuery = ref('')
const searchResults = ref([])
const loadingSearch = ref(false)

// Performance data
const performanceData = ref([]) // Raw performance data from API
const loadingPerformance = ref(false)
const best1RM = ref(null) // Best calculated 1RM for movements
const bestFormula = ref(null) // Formula used for best 1RM

// Movement-specific
const selectedRepScheme = ref('All')
const repSchemes = ref(['All'])

// Chart instances
const chartCanvas = ref(null)
const wodChartCanvas = ref(null)
let chartInstance = null
let wodChartInstance = null

const error = ref(null)

// Quick Log state
const quickLogDialog = ref(false)
const quickLogSubmitting = ref(false)
const quickLogData = ref({
  name: '',
  date: '',
  totalTime: null,
  notes: '',
  selectedItem: null,
  movement: {
    sets: null,
    reps: null,
    weight: null,
    time: null,
    distance: null,
    notes: ''
  },
  wod: {
    scoreType: null,
    scoreValue: null,
    timeSeconds: null,
    rounds: null,
    reps: null,
    notes: ''
  }
})
const movements = ref([])
const wods = ref([])
const workoutTemplates = ref([])
const loadingMovements = ref(false)
const loadingWods = ref(false)

// Computed: Heaviest Lifts (Top 3 Maxes - Movement only)
const heaviestLifts = computed(() => {
  if (!selectedItem.value || selectedItem.value.type !== 'movement') return []

  // Get all unique weight records, sorted by weight descending
  const weightRecords = performanceData.value
    .filter(p => p.weight)
    .sort((a, b) => b.weight - a.weight)

  // Get top 3 unique weights
  const seen = new Set()
  const top3 = []

  for (const record of weightRecords) {
    const key = `${record.weight}-${record.reps || 0}`
    if (!seen.has(key) && top3.length < 3) {
      seen.add(key)
      top3.push(record)
    }
  }

  return top3
})

// Computed: Best WOD Performances (Top 3)
const bestWODPerformances = computed(() => {
  if (!selectedItem.value || selectedItem.value.type !== 'wod' || performanceData.value.length === 0) return []

  // Get WOD's defined score type
  const wodScoreType = performanceData.value[0]?.wod_score_type || ''

  // Filter data based on WOD's score type
  let validData = []
  if (wodScoreType.includes('Time')) {
    validData = performanceData.value.filter(p => p.time_seconds)
  } else if (wodScoreType.includes('Rounds')) {
    validData = performanceData.value.filter(p => p.rounds !== null || p.reps !== null)
  } else if (wodScoreType.includes('Weight')) {
    validData = performanceData.value.filter(p => p.weight)
  } else {
    validData = performanceData.value
  }

  // Sort by best performance based on score type
  const sorted = [...validData].sort((a, b) => {
    if (wodScoreType.includes('Time')) {
      // Time-based (lower is better)
      return a.time_seconds - b.time_seconds
    } else if (wodScoreType.includes('Rounds')) {
      // Rounds+Reps (higher is better)
      const aTotal = (a.rounds || 0) * 1000 + (a.reps || 0)
      const bTotal = (b.rounds || 0) * 1000 + (b.reps || 0)
      return bTotal - aTotal
    } else if (wodScoreType.includes('Weight')) {
      // Weight (higher is better)
      return b.weight - a.weight
    }
    return 0
  })

  return sorted.slice(0, 3)
})

// Computed: Filtered Chart Data (for movement chart with rep scheme filter)
const filteredChartData = computed(() => {
  if (!selectedItem.value || selectedItem.value.type !== 'movement') return []

  if (selectedRepScheme.value === 'All') {
    return performanceData.value.filter(p => p.weight)
  }

  const targetReps = parseInt(selectedRepScheme.value.split(' ')[0])
  return performanceData.value.filter(p => p.weight && p.reps === targetReps)
})

// Watch performance data and auto-select rep scheme
watch(performanceData, async (newData) => {
  if (!selectedItem.value || selectedItem.value.type !== 'movement') {
    return
  }

  const schemes = new Set(['All'])
  newData.forEach(p => {
    if (p.reps) {
      schemes.add(`${p.reps} reps`)
    }
  })
  repSchemes.value = Array.from(schemes)

  // Auto-select the rep scheme with the heaviest weight
  if (newData.length > 0) {
    const heaviest = newData
      .filter(p => p.weight && p.reps)
      .sort((a, b) => b.weight - a.weight)[0]

    if (heaviest) {
      selectedRepScheme.value = `${heaviest.reps} reps`
    } else {
      selectedRepScheme.value = 'All'
    }
  }
}, { immediate: true })

// Watch filtered chart data and render movement chart when it changes
watch(filteredChartData, async () => {
  if (selectedItem.value?.type === 'movement' && filteredChartData.value.length > 0) {
    await nextTick()
    renderMovementChart()
  }
}, { deep: true })

// Watch Quick Log selected item to populate WOD score type
watch(() => quickLogData.value.selectedItem, (newItem) => {
  if (newItem?.type === 'wod' && newItem.data?.score_type) {
    quickLogData.value.wod.scoreType = newItem.data.score_type
  }
})

// Computed: WOD Performance Data for Chart
const wodPerformanceData = computed(() => {
  if (!selectedItem.value || selectedItem.value.type !== 'wod') return []
  return performanceData.value
})

// Computed: History Grouped by Year
const groupedHistory = computed(() => {
  if (!selectedItem.value || performanceData.value.length === 0) return {}

  const grouped = {}

  performanceData.value.forEach(entry => {
    const year = new Date(entry.workout_date).getFullYear()
    if (!grouped[year]) {
      grouped[year] = []
    }
    grouped[year].push(entry)
  })

  // Sort years descending
  const sorted = {}
  Object.keys(grouped)
    .sort((a, b) => b - a)
    .forEach(year => {
      // Sort entries within year by date descending
      sorted[year] = grouped[year].sort((a, b) =>
        new Date(b.workout_date) - new Date(a.workout_date)
      )
    })

  return sorted
})

// Search Handler (debounced unified search)
let searchTimeout = null
function handleSearch(query) {
  searchQuery.value = query
  if (!query || query.length < 2) {
    searchResults.value = []
    return
  }

  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(async () => {
    await performUnifiedSearch(query)
  }, 300)
}

// Unified Search (Movements + WODs)
async function performUnifiedSearch(query) {
  loadingSearch.value = true
  try {
    const response = await axios.get('/api/performance/search', {
      params: { q: query, limit: 20 }
    })

    const results = response.data.results || []
    searchResults.value = results.map(r => ({
      id: `${r.type}-${r.id}`,
      name: r.name,
      type: r.type,
      data: r.data
    }))
  } catch (err) {
    console.error('Failed to search:', err)
    error.value = 'Failed to search movements and WODs'
    searchResults.value = []
  } finally {
    loadingSearch.value = false
  }
}

// Selection Handler
async function handleSelection(item) {
  if (!item) {
    clearSelection()
    return
  }

  selectedItem.value = item

  // Update URL query params to preserve state
  router.replace({
    query: {
      type: item.type,
      id: item.data.id,
      name: item.name
    }
  })

  await fetchPerformanceData()
}

// Clear Selection
function clearSelection() {
  selectedItem.value = null
  performanceData.value = []
  best1RM.value = null
  bestFormula.value = null
  destroyCharts()

  // Clear URL query params
  router.replace({ query: {} })
}

// Fetch Performance Data (Movement or WOD specific)
async function fetchPerformanceData() {
  if (!selectedItem.value) return

  loadingPerformance.value = true
  error.value = null

  try {
    let response
    if (selectedItem.value.type === 'movement') {
      const movementId = selectedItem.value.data.id
      response = await axios.get(`/api/performance/movements/${movementId}`)
      performanceData.value = response.data.performances || []
      best1RM.value = response.data.best_1rm || null
      bestFormula.value = response.data.best_formula || null
    } else if (selectedItem.value.type === 'wod') {
      const wodId = selectedItem.value.data.id
      response = await axios.get(`/api/performance/wods/${wodId}`)
      performanceData.value = response.data.performances || []
    }

    // Stop loading spinner BEFORE rendering charts so canvas element is in DOM
    loadingPerformance.value = false

    // Note: Chart rendering is handled by the performanceData watcher for movements
    // For WODs, render directly since there's no rep scheme selection
    if (selectedItem.value.type === 'wod') {
      await nextTick()
      renderCharts()
    }
  } catch (err) {
    console.error('Failed to fetch performance data:', err)
    error.value = `Failed to load ${selectedItem.value.type} performance data`
    performanceData.value = []
    loadingPerformance.value = false
  }
}

// Filter Chart (when rep scheme changes)
async function filterChart() {
  await nextTick()
  renderCharts()
}

// Render Charts
function renderCharts() {
  if (selectedItem.value?.type === 'movement') {
    renderMovementChart()
  } else if (selectedItem.value?.type === 'wod') {
    renderWODChart()
  }
}

// Render Movement Chart
function renderMovementChart() {
  if (!chartCanvas.value || filteredChartData.value.length === 0) {
    return
  }

  destroyCharts()

  const data = filteredChartData.value
    .filter(p => p.weight)
    .sort((a, b) => {
      // Primary sort: workout_date (ascending - oldest to newest)
      const dateCompare = new Date(a.workout_date) - new Date(b.workout_date)
      if (dateCompare !== 0) return dateCompare

      // Secondary sort: created_at or id (ascending - earliest to latest)
      if (a.created_at && b.created_at) {
        return new Date(a.created_at) - new Date(b.created_at)
      }
      return a.id - b.id
    })

  if (data.length === 0) {
    return
  }

  const labels = data.map(p => formatChartDate(p.workout_date))
  const weights = data.map(p => p.weight)
  const estimated1RMs = data.map(p => p.calculated_1rm || null)

  // Build datasets array - always include weight, only add 1RM if we have data
  const datasets = [{
    label: 'Weight (lbs)',
    data: weights,
    borderColor: '#2c3e50',
    backgroundColor: 'rgba(44, 62, 80, 0.1)',
    tension: 0.4,
    pointRadius: 5,
    pointHoverRadius: 7,
    pointBackgroundColor: '#2c3e50',
    pointBorderColor: '#2c3e50',
    fill: true
  }]

  // Add 1RM dataset if we have any calculated 1RM values
  if (estimated1RMs.some(rm => rm !== null)) {
    datasets.push({
      label: 'Estimated 1RM (lbs)',
      data: estimated1RMs,
      borderColor: '#ffc107',
      backgroundColor: 'rgba(255, 193, 7, 0.1)',
      tension: 0.4,
      pointRadius: 4,
      pointHoverRadius: 6,
      pointBackgroundColor: '#ffc107',
      pointBorderColor: '#ffc107',
      fill: false,
      borderDash: [5, 5] // Dashed line for estimated values
    })
  }

  chartInstance = new Chart(chartCanvas.value, {
    type: 'line',
    data: {
      labels,
      datasets
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: estimated1RMs.some(rm => rm !== null), // Show legend only if 1RM data exists
          position: 'bottom',
          labels: {
            boxWidth: 12,
            padding: 10,
            font: {
              size: 11
            }
          }
        },
        tooltip: {
          callbacks: {
            title: function(context) {
              const entry = data[context[0].dataIndex]
              return formatChartDate(entry.workout_date)
            },
            label: function(context) {
              const entry = data[context.dataIndex]
              const datasetLabel = context.dataset.label

              if (datasetLabel.includes('1RM')) {
                // For 1RM dataset
                if (entry.calculated_1rm) {
                  let label = `Est. 1RM: ${Math.round(entry.calculated_1rm)} lbs`
                  if (entry.formula) label += ` (${entry.formula})`
                  return label
                }
                return null
              } else {
                // For weight dataset
                let label = `${entry.weight} lbs`
                if (entry.reps) label += ` × ${entry.reps} reps`
                if (entry.is_pr) label += ' 🏆 PR'
                return label
              }
            },
            filter: function(tooltipItem) {
              // Don't show tooltip for null values
              return tooltipItem.raw !== null
            }
          }
        }
      },
      scales: {
        y: {
          beginAtZero: false,
          title: {
            display: true,
            text: 'Weight (lbs)',
            font: {
              size: 12,
              weight: 'bold'
            },
            color: '#1a1a1a'
          },
          ticks: {
            stepSize: 1,
            callback: function(value) {
              return Math.round(value) + ' lbs'
            }
          }
        }
      }
    }
  })
}

// Render WOD Chart
function renderWODChart() {
  if (!wodChartCanvas.value || wodPerformanceData.value.length === 0) {
    return
  }

  destroyCharts()

  // Use the WOD's defined score_type to determine what to plot
  const wodScoreType = wodPerformanceData.value[0]?.wod_score_type || ''

  // Filter data to only include records matching the WOD's score_type
  let filteredData = [...wodPerformanceData.value]
  let values, label, isTimeBased = false

  if (wodScoreType.includes('Time')) {
    // Only include records with time_seconds
    filteredData = filteredData.filter(p => p.time_seconds)
    values = filteredData.map(p => p.time_seconds / 60)
    label = 'Time (minutes)'
    isTimeBased = true
  } else if (wodScoreType.includes('Rounds')) {
    // Only include records with rounds or reps
    filteredData = filteredData.filter(p => p.rounds !== null || p.reps !== null)
    values = filteredData.map(p => (p.rounds || 0) * 100 + (p.reps || 0))
    label = 'Rounds + Reps'
  } else if (wodScoreType.includes('Weight')) {
    // Only include records with weight
    filteredData = filteredData.filter(p => p.weight)
    values = filteredData.map(p => p.weight)
    label = 'Weight (lbs)'
  } else {
    return
  }

  if (filteredData.length === 0) {
    return
  }

  const data = filteredData.sort((a, b) => {
    // Primary sort: workout_date (ascending - oldest to newest)
    const dateCompare = new Date(a.workout_date) - new Date(b.workout_date)
    if (dateCompare !== 0) return dateCompare

    // Secondary sort: created_at or id (ascending - earliest to latest)
    if (a.created_at && b.created_at) {
      return new Date(a.created_at) - new Date(b.created_at)
    }
    return a.id - b.id
  })
  const labels = data.map(p => formatChartDate(p.workout_date))

  chartInstance = new Chart(wodChartCanvas.value, {
    type: 'line',
    data: {
      labels,
      datasets: [{
        label,
        data: values,
        borderColor: '#2c3e50',
        backgroundColor: 'rgba(44, 62, 80, 0.1)',
        tension: 0.4,
        pointRadius: 5,
        pointHoverRadius: 7,
        pointBackgroundColor: '#2c3e50',
        pointBorderColor: '#2c3e50',
        fill: true
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        },
        tooltip: {
          callbacks: {
            title: function(context) {
              const entry = data[context[0].dataIndex]
              return formatChartDate(entry.workout_date)
            },
            label: function(context) {
              const entry = data[context.dataIndex]
              let labelText = formatWODScore(entry)
              if (entry.is_pr) labelText += ' 🏆 PR'
              return labelText
            }
          }
        }
      },
      scales: {
        y: {
          beginAtZero: false,
          reverse: isTimeBased, // For time-based, lower is better
          title: {
            display: true,
            text: label,
            font: {
              size: 12,
              weight: 'bold'
            },
            color: '#1a1a1a'
          },
          ticks: {
            stepSize: 1,
            callback: function(value) {
              if (isTimeBased) {
                return Math.round(value) + ' min'
              } else if (wodScoreType.includes('Weight')) {
                return Math.round(value) + ' lbs'
              } else {
                return Math.round(value)
              }
            }
          }
        }
      }
    }
  })
}

// Destroy Charts
function destroyCharts() {
  if (chartInstance) {
    chartInstance.destroy()
    chartInstance = null
  }
  if (wodChartInstance) {
    wodChartInstance.destroy()
    wodChartInstance = null
  }
}

// Helper functions for Quick Log
function getTodayDate() {
  const today = new Date()
  return today.toISOString().split('T')[0]
}

function formatQuickLogName(date) {
  if (!date) return 'Workout'
  const d = new Date(date + 'T00:00:00')
  const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
  return `${days[d.getDay()]} Workout`
}

function updateQuickLogName() {
  quickLogData.value.name = formatQuickLogName(quickLogData.value.date)
}

// Unified search items for Quick Log autocomplete
const unifiedSearchItems = computed(() => {
  const items = []

  // Add movements
  movements.value.forEach(movement => {
    items.push({
      id: `movement-${movement.id}`,
      type: 'movement',
      entityId: movement.id,
      name: movement.name,
      displayName: movement.name,
      data: movement
    })
  })

  // Add WODs
  wods.value.forEach(wod => {
    items.push({
      id: `wod-${wod.id}`,
      type: 'wod',
      entityId: wod.id,
      name: wod.name,
      displayName: `${wod.name} (WOD)`,
      data: wod
    })
  })

  // Add workout templates
  workoutTemplates.value.forEach(template => {
    items.push({
      id: `template-${template.id}`,
      type: 'template',
      entityId: template.id,
      name: template.name,
      displayName: template.name,
      data: template
    })
  })

  return items
})

// Quick Log - Opens dialog with pre-selected item
async function quickLog() {
  if (!selectedItem.value) return

  // Reset quick log data
  quickLogData.value = {
    name: formatQuickLogName(getTodayDate()),
    date: getTodayDate(),
    totalTime: null,
    notes: '',
    selectedItem: null,
    movement: {
      sets: null,
      reps: null,
      weight: null,
      time: null,
      distance: null,
      notes: ''
    },
    wod: {
      scoreType: null,
      scoreValue: null,
      timeSeconds: null,
      rounds: null,
      reps: null,
      notes: ''
    }
  }

  // Pre-populate with current item from Performance screen
  if (selectedItem.value.type === 'movement') {
    quickLogData.value.selectedItem = {
      id: `movement-${selectedItem.value.data.id}`,
      type: 'movement',
      entityId: selectedItem.value.data.id,
      name: selectedItem.value.data.name,
      displayName: selectedItem.value.data.name,
      data: selectedItem.value.data
    }
  } else if (selectedItem.value.type === 'wod') {
    quickLogData.value.selectedItem = {
      id: `wod-${selectedItem.value.data.id}`,
      type: 'wod',
      entityId: selectedItem.value.data.id,
      name: selectedItem.value.data.name,
      displayName: selectedItem.value.data.name,
      data: selectedItem.value.data
    }
  }

  // Load movements and WODs for autocomplete
  if (movements.value.length === 0) {
    loadingMovements.value = true
    try {
      const response = await axios.get('/api/movements')
      movements.value = response.data.movements || []
    } catch (err) {
      console.error('Error fetching movements:', err)
    } finally {
      loadingMovements.value = false
    }
  }

  if (wods.value.length === 0) {
    loadingWods.value = true
    try {
      const [standardRes, customRes] = await Promise.all([
        axios.get('/api/wods/standard'),
        axios.get('/api/wods/my-wods')
      ])
      const standard = Array.isArray(standardRes.data.wods) ? standardRes.data.wods : []
      const custom = Array.isArray(customRes.data.wods) ? customRes.data.wods : []
      wods.value = [...standard, ...custom]
    } catch (err) {
      console.error('Error fetching WODs:', err)
    } finally {
      loadingWods.value = false
    }
  }

  // Load workout templates for autocomplete
  if (workoutTemplates.value.length === 0) {
    try {
      const [standardRes, customRes] = await Promise.all([
        axios.get('/api/workouts/standard'),
        axios.get('/api/workouts/my-templates')
      ])
      const standard = Array.isArray(standardRes.data.workouts) ? standardRes.data.workouts : []
      const custom = Array.isArray(customRes.data.workouts) ? customRes.data.workouts : []
      workoutTemplates.value = [...standard, ...custom]
    } catch (err) {
      console.error('Error fetching workout templates:', err)
    }
  }

  // Open dialog
  quickLogDialog.value = true
}

// Close Quick Log dialog
function closeQuickLog() {
  quickLogDialog.value = false
}

// View workout details
function viewWorkout(workoutId) {
  router.push(`/workouts/${workoutId}`)
}

// Submit Quick Log
async function submitQuickLog() {
  // If template is selected, navigate to log workout page with template pre-selected
  if (quickLogData.value.selectedItem && quickLogData.value.selectedItem.type === 'template') {
    closeQuickLog()
    router.push({
      path: '/workouts/log',
      query: {
        template: quickLogData.value.selectedItem.entityId,
        date: quickLogData.value.date
      }
    })
    return
  }

  quickLogSubmitting.value = true

  try {
    const payload = {
      workout_date: quickLogData.value.date,
      workout_name: quickLogData.value.name,
      total_time: quickLogData.value.totalTime ? quickLogData.value.totalTime * 60 : null,
      notes: quickLogData.value.notes || null,
      movements: [],
      wods: []
    }

    // Add movement or WOD performance
    if (quickLogData.value.selectedItem) {
      if (quickLogData.value.selectedItem.type === 'movement') {
        const m = quickLogData.value.movement
        if (m.sets || m.reps || m.weight || m.time || m.distance) {
          payload.movements.push({
            movement_id: quickLogData.value.selectedItem.entityId,
            sets: m.sets || null,
            reps: m.reps || null,
            weight: m.weight || null,
            time: m.time ? m.time * 60 : null,
            distance: m.distance || null,
            notes: m.notes || null
          })
        }
      } else if (quickLogData.value.selectedItem.type === 'wod') {
        const w = quickLogData.value.wod
        if (w.timeSeconds || w.rounds || w.reps || w.scoreValue) {
          payload.wods.push({
            wod_id: quickLogData.value.selectedItem.entityId,
            time_seconds: w.timeSeconds || null,
            rounds: w.rounds || null,
            reps: w.reps || null,
            score_value: w.scoreValue || null,
            notes: w.notes || null
          })
        }
      }
    }

    await axios.post('/api/workouts', payload)

    // Close dialog
    quickLogDialog.value = false

    // Refresh performance data
    await fetchPerformanceData()
  } catch (err) {
    console.error('Failed to log workout:', err)
    alert(err.response?.data?.message || 'Failed to log workout')
  } finally {
    quickLogSubmitting.value = false
  }
}

// Format WOD Score
function formatWODScore(performance) {
  if (performance.time_seconds) {
    return formatTime(performance.time_seconds)
  } else if (performance.rounds !== null && performance.reps !== null) {
    return `${performance.rounds}+${performance.reps}`
  } else if (performance.rounds !== null) {
    return `${performance.rounds} rounds`
  } else if (performance.reps !== null) {
    return `${performance.reps} reps`
  } else if (performance.weight) {
    return `${performance.weight} lbs`
  } else if (performance.score_value) {
    return performance.score_value
  }
  return 'N/A'
}

// Format Date (for display with Today/Yesterday)
function formatDate(dateString) {
  const datePart = dateString.split('T')[0]
  const [year, month, day] = datePart.split('-').map(Number)
  const date = new Date(year, month - 1, day)

  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  const dateOnly = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  const todayOnly = new Date(today.getFullYear(), today.getMonth(), today.getDate())
  const yesterdayOnly = new Date(yesterday.getFullYear(), yesterday.getMonth(), yesterday.getDate())

  if (dateOnly.getTime() === todayOnly.getTime()) {
    return 'Today'
  } else if (dateOnly.getTime() === yesterdayOnly.getTime()) {
    return 'Yesterday'
  } else {
    const options = { month: 'short', day: 'numeric', year: 'numeric' }
    return date.toLocaleDateString('en-US', options)
  }
}

// Format Date for charts (always show actual date with year)
function formatChartDate(dateString) {
  const datePart = dateString.split('T')[0]
  const [year, month, day] = datePart.split('-').map(Number)
  const date = new Date(year, month - 1, day)
  const options = { month: 'short', day: 'numeric', year: 'numeric' }
  return date.toLocaleDateString('en-US', options)
}

// Format Time
function formatTime(seconds) {
  if (!seconds) return ''

  if (seconds < 60) {
    return `${seconds}s`
  } else if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60)
    const secs = seconds % 60
    return secs > 0 ? `${minutes}:${secs.toString().padStart(2, '0')}` : `${minutes}:00`
  } else {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    return `${hours}h ${minutes}m`
  }
}

// Format Movement Type
function formatMovementType(type) {
  if (!type) return ''
  return type.charAt(0).toUpperCase() + type.slice(1)
}

// Restore selection from URL query params on mount
onMounted(async () => {
  const { type, id, name } = route.query

  if (type && id && name) {
    // Reconstruct the selected item from query params
    selectedItem.value = {
      id: `${type}-${id}`,  // Composite ID matching search results format
      type,
      name,
      data: {
        id: parseInt(id),
        name  // Include name for Quick Log functionality
      }
    }

    // Fetch the performance data
    await fetchPerformanceData()
  }
})

// Cleanup on unmount
onBeforeUnmount(() => {
  destroyCharts()
})
</script>
