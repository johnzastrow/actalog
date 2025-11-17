<template>
  <v-container fluid class="pa-0" style="background-color: #f5f7fa; min-height: 100vh; overflow-y: auto">
    <!-- Header -->
    <v-app-bar color="#2c3e50" elevation="0" density="compact" style="position: fixed; top: 0; z-index: 10; width: 100%">
      <v-toolbar-title class="text-white font-weight-bold">Dashboard</v-toolbar-title>
      <v-spacer />
      <v-btn icon="mdi-bell-outline" color="white" size="small" />
    </v-app-bar>

    <v-container class="px-1 pb-1 pt-0" style="margin-top: 48px; margin-bottom: 70px">
      <!-- Email Verification Alert -->
      <v-alert
        v-if="authStore.user && !authStore.user.email_verified"
        type="warning"
        prominent
        closable
        class="mb-2"
      >
        <v-row align="center">
          <v-col class="grow">
            <div class="text-h6">Email Verification Required</div>
            <div class="text-body-2">
              Please verify your email address to access all features. Check your inbox for the verification link.
            </div>
          </v-col>
          <v-col class="shrink">
            <v-btn
              color="warning"
              variant="elevated"
              @click="$router.push('/resend-verification')"
            >
              Resend Email
            </v-btn>
          </v-col>
        </v-row>
      </v-alert>

      <!-- Stats Cards -->
      <v-row dense class="mb-1">
        <v-col cols="6">
          <v-card elevation="0" rounded class="pa-1" style="background: white">
            <div class="d-flex align-center">
              <v-icon color="#00bcd4" size="28" class="mr-2">mdi-dumbbell</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" style="color: #1a1a1a">
                  {{ totalWorkouts }}
                </div>
                <div class="text-caption" style="color: #666">Total Workouts</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="6">
          <v-card elevation="0" rounded class="pa-1" style="background: white">
            <div class="d-flex align-center">
              <v-icon color="#ffc107" size="28" class="mr-2">mdi-calendar-month</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" style="color: #1a1a1a">
                  {{ monthWorkouts }}
                </div>
                <div class="text-caption" style="color: #666">This Month</div>
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>

      <v-row dense class="mb-1">
        <v-col cols="6">
          <v-card elevation="0" rounded class="pa-1" style="background: white">
            <div class="d-flex align-center">
              <v-icon color="#4caf50" size="28" class="mr-2">mdi-fire</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" style="color: #1a1a1a">
                  {{ currentStreak }}
                </div>
                <div class="text-caption" style="color: #666">Day Streak</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="6">
          <v-card elevation="0" rounded class="pa-1" style="background: white">
            <div class="d-flex align-center">
              <v-icon color="#e91e63" size="28" class="mr-2">mdi-clock-outline</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" style="color: #1a1a1a">
                  {{ avgTimePerWorkout }}m
                </div>
                <div class="text-caption" style="color: #666">Avg Time</div>
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>

      <!-- Quick Actions -->
      <v-row dense class="mb-1">
        <v-col cols="6">
          <v-card
            elevation="0"
            rounded
            class="pa-1 text-center"
            style="background: linear-gradient(135deg, #ffc107 0%, #ffb300 100%); cursor: pointer"
            @click="openQuickLog"
          >
            <v-icon size="28" color="white" class="mb-1">mdi-lightning-bolt</v-icon>
            <div class="text-body-2 font-weight-bold text-white">Quick Log</div>
            <div class="text-caption text-white" style="opacity: 0.9; font-size: 9px">
              Fast entry
            </div>
          </v-card>
        </v-col>
        <v-col cols="6">
          <v-card
            elevation="0"
            rounded
            class="pa-1 text-center"
            style="background: linear-gradient(135deg, #00bcd4 0%, #00acc1 100%); cursor: pointer"
            @click="$router.push('/workouts/log')"
          >
            <v-icon size="28" color="white" class="mb-1">mdi-dumbbell</v-icon>
            <div class="text-body-2 font-weight-bold text-white">Log Workout</div>
            <div class="text-caption text-white" style="opacity: 0.9; font-size: 9px">
              Use template
            </div>
          </v-card>
        </v-col>
      </v-row>

      <!-- Recent Workouts -->
      <div class="mb-1">
        <div class="d-flex align-center justify-space-between mb-1">
          <h3 class="text-h6 font-weight-bold" style="color: #1a1a1a">Last 30 Days</h3>
          <v-btn
            size="small"
            variant="text"
            color="#00bcd4"
            style="text-transform: none"
            @click="$router.push('/workouts')"
          >
            View All
          </v-btn>
        </div>

        <!-- Loading State -->
        <div v-if="loading" class="text-center py-8">
          <v-progress-circular indeterminate color="#00bcd4" size="48" />
          <p class="mt-2 text-caption" style="color: #666">Loading workouts...</p>
        </div>

        <!-- Empty State -->
        <v-card
          v-else-if="!loading && recentWorkouts.length === 0"
          elevation="0"
          rounded
          class="pa-2 text-center"
          style="background: white"
        >
          <v-icon size="48" color="#ccc">mdi-clipboard-text-outline</v-icon>
          <p class="text-body-1 mt-1 mb-0" style="color: #2c3e50">No workouts logged yet</p>
          <p class="text-body-2 mb-0" style="color: #666">
            Start tracking your fitness journey today!
          </p>
        </v-card>

        <!-- Recent Workouts List -->
        <div v-else>
          <v-card
            v-for="workout in recentWorkouts"
            :key="workout.id"
            elevation="0"
            rounded
            class="mb-1 pa-1"
            style="background: white; cursor: pointer"
            @click="toggleWorkoutExpand(workout.id)"
          >
            <div class="d-flex align-center mb-1">
              <v-icon color="#00bcd4" class="mr-2" size="small">mdi-dumbbell</v-icon>
              <div class="flex-grow-1">
                <div class="font-weight-bold text-body-1" style="color: #1a1a1a">
                  {{ workout.workout_name || 'Workout' }}
                </div>
                <div class="text-caption" style="color: #666">
                  {{ formatDate(workout.workout_date) }}
                  <span v-if="workout.total_time"> • {{ formatTime(workout.total_time) }}</span>
                </div>
              </div>
              <v-icon color="#00bcd4" size="small">
                {{ expandedWorkouts.has(workout.id) ? 'mdi-chevron-up' : 'mdi-chevron-down' }}
              </v-icon>
            </div>

            <!-- Collapsed view - showing summary only -->
            <div v-if="!expandedWorkouts.has(workout.id)">
              <!-- Display movement performance (first 2 only) -->
              <div v-if="workout.performance_movements && workout.performance_movements.length > 0" class="ml-7 mt-2">
                <div v-for="(perf, index) in workout.performance_movements.slice(0, 2)" :key="index" class="text-caption mb-1" style="color: #666">
                  <v-icon size="small" color="#00bcd4">mdi-weight-lifter</v-icon>
                  <strong>{{ perf.movement?.name || 'Movement' }}:</strong>
                  <span v-if="perf.weight"> {{ perf.weight }}lbs</span>
                  <span v-if="perf.sets"> {{ perf.sets }}x</span><span v-if="perf.reps">{{ perf.reps }}</span>
                  <span v-if="perf.distance"> {{ perf.distance }}m</span>
                  <span v-if="perf.time_seconds"> {{ formatTime(perf.time_seconds) }}</span>
                  <v-chip v-if="perf.is_pr" color="#ffc107" size="x-small" class="ml-2" variant="flat">
                    <v-icon size="x-small" start>mdi-trophy</v-icon>PR
                  </v-chip>
                </div>
                <div v-if="workout.performance_movements.length > 2" class="text-caption ml-7" style="color: #999">
                  +{{ workout.performance_movements.length - 2 }} more...
                </div>
              </div>

              <!-- Display WOD performance (first 1 only) -->
              <div v-if="workout.performance_wods && workout.performance_wods.length > 0" class="ml-7 mt-2">
                <div v-for="(perf, index) in workout.performance_wods.slice(0, 1)" :key="index" class="text-caption mb-1" style="color: #666">
                  <v-icon size="small" color="#ffc107">mdi-fire</v-icon>
                  <strong>{{ perf.wod?.name || 'WOD' }}:</strong>
                  <span v-if="perf.time_seconds"> {{ formatTime(perf.time_seconds) }}</span>
                  <span v-if="perf.rounds && perf.reps"> {{ perf.rounds }}+{{ perf.reps }}</span>
                  <span v-else-if="perf.rounds"> {{ perf.rounds }} rounds</span>
                  <span v-else-if="perf.reps"> {{ perf.reps }} reps</span>
                  <span v-if="perf.score_value"> ({{ perf.score_value }})</span>
                  <v-chip v-if="perf.is_pr" color="#ffc107" size="x-small" class="ml-2" variant="flat">
                    <v-icon size="x-small" start>mdi-trophy</v-icon>PR
                  </v-chip>
                </div>
                <div v-if="workout.performance_wods.length > 1" class="text-caption ml-7" style="color: #999">
                  +{{ workout.performance_wods.length - 1 }} more...
                </div>
              </div>

              <!-- Notes preview -->
              <div v-if="workout.notes" class="ml-7 mt-2 text-caption" style="color: #666">
                📝 {{ truncateText(workout.notes, 80) }}
              </div>
            </div>

            <!-- Expanded view - showing all details -->
            <div v-else>
              <!-- Display ALL movement performances -->
              <div v-if="workout.performance_movements && workout.performance_movements.length > 0" class="ml-7 mt-2">
                <div class="text-caption font-weight-bold mb-1" style="color: #1a1a1a">Movements:</div>
                <div v-for="(perf, index) in workout.performance_movements" :key="index" class="text-caption mb-1 ml-2" style="color: #666">
                  <v-icon size="small" color="#00bcd4">mdi-weight-lifter</v-icon>
                  <strong>{{ perf.movement?.name || 'Movement' }}:</strong>
                  <span v-if="perf.weight"> {{ perf.weight }}lbs</span>
                  <span v-if="perf.sets"> × {{ perf.sets }} sets</span>
                  <span v-if="perf.reps"> × {{ perf.reps }} reps</span>
                  <span v-if="perf.distance"> × {{ perf.distance }}m</span>
                  <span v-if="perf.time_seconds"> in {{ formatTime(perf.time_seconds) }}</span>
                  <v-chip v-if="perf.is_pr" color="#ffc107" size="x-small" class="ml-2" variant="flat">
                    <v-icon size="x-small" start>mdi-trophy</v-icon>PR
                  </v-chip>
                </div>
              </div>

              <!-- Display ALL WOD performances -->
              <div v-if="workout.performance_wods && workout.performance_wods.length > 0" class="ml-7 mt-2">
                <div class="text-caption font-weight-bold mb-1" style="color: #1a1a1a">WODs:</div>
                <div v-for="(perf, index) in workout.performance_wods" :key="index" class="text-caption mb-1 ml-2" style="color: #666">
                  <v-icon size="small" color="#ffc107">mdi-fire</v-icon>
                  <strong>{{ perf.wod?.name || 'WOD' }}:</strong>
                  <span v-if="perf.time_seconds"> Time: {{ formatTime(perf.time_seconds) }}</span>
                  <span v-if="perf.rounds && perf.reps"> Score: {{ perf.rounds }}+{{ perf.reps }}</span>
                  <span v-else-if="perf.rounds"> Rounds: {{ perf.rounds }}</span>
                  <span v-else-if="perf.reps"> Reps: {{ perf.reps }}</span>
                  <span v-if="perf.score_value"> ({{ perf.score_value }})</span>
                  <v-chip v-if="perf.is_pr" color="#ffc107" size="x-small" class="ml-2" variant="flat">
                    <v-icon size="x-small" start>mdi-trophy</v-icon>PR
                  </v-chip>
                </div>
              </div>

              <!-- Full notes with markdown rendering -->
              <div v-if="workout.notes" class="ml-7 mt-2 text-caption" style="color: #666">
                <div class="font-weight-bold mb-1" style="color: #1a1a1a">Notes:</div>
                <div class="ml-2">
                  <MarkdownRenderer :content="workout.notes" />
                </div>
              </div>

              <!-- Action buttons when expanded -->
              <div class="ml-7 mt-3 d-flex gap-2">
                <v-btn
                  size="small"
                  color="#00bcd4"
                  variant="outlined"
                  @click.stop="viewWorkout(workout.id)"
                >
                  <v-icon start size="small">mdi-open-in-new</v-icon>
                  View Details
                </v-btn>
              </div>
            </div>
          </v-card>
        </div>
      </div>
    </v-container>

    <!-- Quick Log Dialog -->
    <v-dialog v-model="quickLogDialog" max-width="500px">
      <v-card>
        <v-card-title class="text-h6 font-weight-bold" style="background: #ffc107; color: white">
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

            <!-- Performance Type Selector -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" style="color: #1a1a1a">
                Add Performance Data (Optional)
              </label>
              <v-select
                v-model="quickLogData.performanceType"
                :items="['None', 'Movement', 'WOD']"
                variant="outlined"
                density="compact"
                hide-details
              />
            </div>

            <!-- Movement Performance -->
            <div v-if="quickLogData.performanceType === 'Movement'" class="mb-1">
              <label class="text-caption font-weight-bold d-block" style="color: #1a1a1a">
                Select Movement
              </label>
              <v-autocomplete
                v-model="quickLogData.movementId"
                :items="movements"
                item-title="name"
                item-value="id"
                :loading="loadingMovements"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                auto-select-first
                placeholder="Search for a movement..."
              >
                <template #prepend-inner>
                  <v-icon color="#00bcd4" size="small">mdi-magnify</v-icon>
                </template>
              </v-autocomplete>

              <div v-if="quickLogData.movementId" class="mt-3 pa-3" style="background: #f5f5f5; border-radius: 8px">
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
            </div>

            <!-- WOD Performance -->
            <div v-if="quickLogData.performanceType === 'WOD'" class="mb-1">
              <label class="text-caption font-weight-bold d-block" style="color: #1a1a1a">
                Select WOD
              </label>
              <v-autocomplete
                v-model="quickLogData.wodId"
                :items="wods"
                item-title="name"
                item-value="id"
                :loading="loadingWods"
                variant="outlined"
                density="compact"
                hide-details
                clearable
                auto-select-first
                placeholder="Search for a WOD..."
              >
                <template #prepend-inner>
                  <v-icon color="#ffc107" size="small">mdi-magnify</v-icon>
                </template>
              </v-autocomplete>

              <div v-if="quickLogData.wodId" class="mt-3 pa-3" style="background: #f5f5f5; border-radius: 8px">
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
            color="#ffc107"
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
      <v-btn value="log" to="/workouts/log" style="position: relative; bottom: 20px">
        <v-avatar color="#ffc107" size="56" style="box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2)">
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
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'
import { useAuthStore } from '@/stores/auth'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

const router = useRouter()
const authStore = useAuthStore()
const activeTab = ref('dashboard')

const loading = ref(false)
const userWorkouts = ref([])
const expandedWorkouts = ref(new Set())

// Get today's date in YYYY-MM-DD format
function getTodayDate() {
  const today = new Date()
  const year = today.getFullYear()
  const month = String(today.getMonth() + 1).padStart(2, '0')
  const day = String(today.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// Format date for Quick Log name like "Quick Log Thu Nov 13, 2025"
function formatQuickLogName(dateString) {
  const date = new Date(dateString + 'T00:00:00') // Ensure local timezone
  const options = { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' }
  const formatted = date.toLocaleDateString('en-US', options)
  return `Quick Log ${formatted}`
}

// Quick Log state
const quickLogDialog = ref(false)
const quickLogSubmitting = ref(false)
const quickLogData = ref({
  name: formatQuickLogName(getTodayDate()),
  date: getTodayDate(),
  totalTime: null,
  notes: '',
  performanceType: 'None',
  movementId: null,
  wodId: null,
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
    timeHours: null,
    timeMinutes: null,
    timeSecondsInput: null,
    rounds: null,
    reps: null,
    weight: null,
    notes: ''
  }
})

// Lists for movements and WODs
const movements = ref([])
const wods = ref([])
const loadingMovements = ref(false)
const loadingWods = ref(false)

// Computed stats
const totalWorkouts = computed(() => userWorkouts.value.length)

const monthWorkouts = computed(() => {
  const now = new Date()
  const firstDayOfMonth = new Date(now.getFullYear(), now.getMonth(), 1)

  return userWorkouts.value.filter(w => {
    const workoutDate = new Date(w.workout_date)
    return workoutDate >= firstDayOfMonth
  }).length
})

const currentStreak = computed(() => {
  if (userWorkouts.value.length === 0) return 0

  // Get unique workout dates (since users can log multiple workouts per day)
  const uniqueDates = [...new Set(userWorkouts.value.map(w => {
    const d = new Date(w.workout_date)
    d.setHours(0, 0, 0, 0)
    return d.getTime()
  }))].sort((a, b) => b - a) // Sort newest to oldest

  if (uniqueDates.length === 0) return 0

  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const todayTime = today.getTime()

  // Check if most recent workout is today or yesterday (active streak)
  const mostRecent = uniqueDates[0]
  const daysSinceLastWorkout = Math.floor((todayTime - mostRecent) / (1000 * 60 * 60 * 24))

  if (daysSinceLastWorkout > 1) {
    return 0 // Streak broken if last workout was more than 1 day ago
  }

  // Count consecutive days going backwards
  let streak = 0
  let expectedDate = mostRecent

  for (const workoutTime of uniqueDates) {
    if (workoutTime === expectedDate) {
      streak++
      expectedDate -= (1000 * 60 * 60 * 24) // Move back one day
    } else {
      break // Gap in dates, streak ends
    }
  }

  return streak
})

const avgTimePerWorkout = computed(() => {
  const workoutsWithTime = userWorkouts.value.filter(w => w.total_time)
  if (workoutsWithTime.length === 0) return 0

  const totalMinutes = workoutsWithTime.reduce((sum, w) => sum + (w.total_time / 60), 0)
  return Math.round(totalMinutes / workoutsWithTime.length)
})

const recentWorkouts = computed(() => {
  const now = new Date()
  const thirtyDaysAgo = new Date(now.getTime() - (30 * 24 * 60 * 60 * 1000))

  return [...userWorkouts.value]
    .filter(w => {
      const workoutDate = new Date(w.workout_date)
      return workoutDate >= thirtyDaysAgo
    })
    .sort((a, b) => new Date(b.workout_date) - new Date(a.workout_date))
    .slice(0, 30) // Show up to 30 most recent from last 30 days
})

// Computed property to get the selected WOD object
const selectedWOD = computed(() => {
  if (!quickLogData.value.wodId) return null
  return wods.value.find(w => w.id === quickLogData.value.wodId)
})

// Watch for WOD selection changes and auto-set score_type
watch(() => quickLogData.value.wodId, (newWodId) => {
  if (newWodId && selectedWOD.value) {
    // Auto-populate the score type from the WOD definition
    quickLogData.value.wod.scoreType = selectedWOD.value.score_type
  } else {
    // Clear score type when no WOD is selected
    quickLogData.value.wod.scoreType = null
  }
})

// Watch for time input changes and auto-calculate total seconds
watch(
  () => [
    quickLogData.value.wod.timeHours,
    quickLogData.value.wod.timeMinutes,
    quickLogData.value.wod.timeSecondsInput
  ],
  ([hours, minutes, seconds]) => {
    const h = hours || 0
    const m = minutes || 0
    const s = seconds || 0
    quickLogData.value.wod.timeSeconds = (h * 3600) + (m * 60) + s
  }
)

// Fetch user's logged workouts
async function fetchUserWorkouts() {
  loading.value = true
  try {
    const response = await axios.get('/api/workouts')
    userWorkouts.value = response.data.workouts || []
    console.log('Fetched user workouts:', userWorkouts.value.length)
  } catch (err) {
    console.error('Failed to fetch user workouts:', err)
    userWorkouts.value = []
  } finally {
    loading.value = false
  }
}

// Format date for display
function formatDate(dateString) {
  // Parse as local date to avoid timezone conversion issues
  // Extract YYYY-MM-DD from the date string
  const datePart = dateString.split('T')[0]
  const [year, month, day] = datePart.split('-').map(Number)
  const date = new Date(year, month - 1, day) // Month is 0-indexed

  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  // Reset time parts for comparison
  const dateOnly = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  const todayOnly = new Date(today.getFullYear(), today.getMonth(), today.getDate())
  const yesterdayOnly = new Date(yesterday.getFullYear(), yesterday.getMonth(), yesterday.getDate())

  if (dateOnly.getTime() === todayOnly.getTime()) {
    return 'Today'
  } else if (dateOnly.getTime() === yesterdayOnly.getTime()) {
    return 'Yesterday'
  } else {
    const options = { weekday: 'short', month: 'short', day: 'numeric' }
    return date.toLocaleDateString('en-US', options)
  }
}

// Format time (seconds to readable format)
function formatTime(seconds) {
  if (!seconds) return ''

  if (seconds < 60) {
    return `${seconds}s`
  } else if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60)
    return `${minutes}min`
  } else {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    return `${hours}h ${minutes}m`
  }
}

// Truncate text
function truncateText(text, maxLength) {
  if (!text || text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

// View workout details
function viewWorkout(workoutId) {
  console.log('View workout details:', workoutId)
  router.push(`/workouts/${workoutId}`)
}

// Toggle workout card expansion
function toggleWorkoutExpand(workoutId) {
  if (expandedWorkouts.value.has(workoutId)) {
    expandedWorkouts.value.delete(workoutId)
  } else {
    expandedWorkouts.value.add(workoutId)
  }
  // Force reactivity update for Set
  expandedWorkouts.value = new Set(expandedWorkouts.value)
}

// Update Quick Log name when date changes
function updateQuickLogName() {
  if (quickLogData.value.date) {
    quickLogData.value.name = formatQuickLogName(quickLogData.value.date)
  }
}

// Open Quick Log dialog and fetch data
async function openQuickLog() {
  quickLogDialog.value = true

  // Fetch movements and WODs if not already loaded
  if (movements.value.length === 0) {
    loadingMovements.value = true
    try {
      const response = await axios.get('/api/movements')
      movements.value = response.data.movements || []
      console.log('Loaded movements:', movements.value.length)
    } catch (error) {
      console.error('Error fetching movements:', error)
    } finally {
      loadingMovements.value = false
    }
  }

  if (wods.value.length === 0) {
    loadingWods.value = true
    try {
      const response = await axios.get('/api/wods')
      wods.value = response.data.wods || []
      console.log('Loaded WODs:', wods.value.length)
    } catch (error) {
      console.error('Error fetching WODs:', error)
    } finally {
      loadingWods.value = false
    }
  }
}

// Close Quick Log dialog
function closeQuickLog() {
  quickLogDialog.value = false
  // Reset form
  const today = getTodayDate()
  quickLogData.value = {
    name: formatQuickLogName(today),
    date: today,
    totalTime: null,
    notes: '',
    performanceType: 'None',
    movementId: null,
    wodId: null,
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
}

// Submit Quick Log
async function submitQuickLog() {
  if (!quickLogData.value.name || !quickLogData.value.date) {
    return
  }

  quickLogSubmitting.value = true

  try {
    const payload = {
      workout_name: quickLogData.value.name,
      workout_date: quickLogData.value.date,
      total_time: quickLogData.value.totalTime ? quickLogData.value.totalTime * 60 : null, // Convert to seconds
      notes: quickLogData.value.notes || null
    }

    // Add movement performance data if selected
    if (quickLogData.value.performanceType === 'Movement' && quickLogData.value.movementId) {
      payload.movements = [{
        movement_id: quickLogData.value.movementId,
        sets: quickLogData.value.movement.sets || null,
        reps: quickLogData.value.movement.reps || null,
        weight: quickLogData.value.movement.weight || null,
        time: quickLogData.value.movement.time || null,
        distance: quickLogData.value.movement.distance || null,
        notes: quickLogData.value.movement.notes || '',
        order_index: 0
      }]
    }

    // Add WOD performance data if selected
    if (quickLogData.value.performanceType === 'WOD' && quickLogData.value.wodId) {
      payload.wods = [{
        wod_id: quickLogData.value.wodId,
        score_type: quickLogData.value.wod.scoreType || null,
        score_value: quickLogData.value.wod.scoreValue || null,
        time_seconds: quickLogData.value.wod.timeSeconds || null,
        rounds: quickLogData.value.wod.rounds || null,
        reps: quickLogData.value.wod.reps || null,
        notes: quickLogData.value.wod.notes || ''
      }]
    }

    await axios.post('/api/workouts', payload)

    // Close dialog
    closeQuickLog()

    // Refresh workouts list
    await fetchUserWorkouts()
  } catch (err) {
    console.error('Failed to log workout:', err)
    alert(err.response?.data?.message || 'Failed to log workout')
  } finally {
    quickLogSubmitting.value = false
  }
}

// Load data on mount
onMounted(() => {
  fetchUserWorkouts()
})
</script>

<style scoped>
/* Dashboard specific styles */
</style>
