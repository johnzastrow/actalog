<template>
  <pull-to-refresh @refresh="handleRefresh">
    <div class="mobile-view-wrapper">
      <v-container class="pa-3">
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
          <v-card elevation="0" rounded class="pa-1" bg-color="surface">
            <div class="d-flex align-center">
              <v-icon color="primary" size="28" class="mr-2">mdi-dumbbell</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" >
                  {{ totalWorkouts }}
                </div>
                <div class="text-caption text-medium-emphasis">Total Workouts</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="6">
          <v-card elevation="0" rounded class="pa-1" bg-color="surface">
            <div class="d-flex align-center">
              <v-icon color="primary" size="28" class="mr-2">mdi-calendar-month</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" >
                  {{ monthWorkouts }}
                </div>
                <div class="text-caption text-medium-emphasis">This Month</div>
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>

      <v-row dense class="mb-1">
        <v-col cols="6">
          <v-card elevation="0" rounded class="pa-1" bg-color="surface">
            <div class="d-flex align-center">
              <v-icon color="success" size="28" class="mr-2">mdi-fire</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" >
                  {{ currentStreak }}
                </div>
                <div class="text-caption text-medium-emphasis">Day Streak</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="6">
          <v-card elevation="0" rounded class="pa-1" bg-color="surface">
            <div class="d-flex align-center">
              <v-icon color="error" size="28" class="mr-2">mdi-clock-outline</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" >
                  {{ avgTimePerWorkout }}m
                </div>
                <div class="text-caption text-medium-emphasis">Avg Time</div>
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>

      <!-- Additional Stats -->
      <v-row dense class="mb-1">
        <v-col cols="6">
          <v-card elevation="0" rounded class="pa-1" bg-color="surface">
            <div class="d-flex align-center">
              <v-icon color="secondary" size="28" class="mr-2">mdi-chart-timeline-variant</v-icon>
              <div>
                <div class="text-h5 font-weight-bold" >
                  {{ avgWorkoutsPerWeek }}
                </div>
                <div class="text-caption text-medium-emphasis">Avg Wrk/Week ({{ new Date().getFullYear() }})</div>
              </div>
            </div>
          </v-card>
        </v-col>
        <v-col cols="6">
          <v-card elevation="0" rounded class="pa-2" bg-color="surface">
            <div class="text-caption mb-1 text-medium-emphasis" style="text-align: center">This Week</div>
            <div class="d-flex justify-space-around">
              <div
                v-for="(day, index) in weekDays"
                :key="index"
                class="text-center"
                style="flex: 1; cursor: pointer"
                @click="handleWeekDayClick(day)"
              >
                <v-avatar
                  size="24"
                  :color="day.hasWorkout ? 'primary' : 'transparent'"
                  :style="day.hasWorkout ? {} : { border: '2px solid rgb(var(--v-theme-on-surface), 0.3)' }"
                  class="d-flex align-center justify-center week-day-avatar"
                >
                  <span
                    class="text-caption font-weight-bold d-flex align-center justify-center"
                    :style="[
                      { width: '100%', height: '100%', lineHeight: '1' },
                      day.hasWorkout ? { color: 'white' } : { color: 'rgb(var(--v-theme-on-surface), 0.6)' }
                    ]"
                  >
                    {{ day.letter }}
                  </span>
                </v-avatar>
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>

      <!-- Active Users This Month Card -->
      <v-row v-if="activeUsersStats.length > 0" dense class="mb-1">
        <v-col cols="12">
          <v-card elevation="0" rounded class="pa-2" bg-color="surface">
            <div class="text-caption mb-2 font-weight-bold text-medium-emphasis">
              <v-icon size="small" color="primary" class="mr-1">mdi-account-group</v-icon>
              Active This Month
            </div>
            <div v-for="user in activeUsersStats" :key="user.id" class="mb-2">
              <div class="d-flex align-center justify-space-between">
                <div class="d-flex align-center">
                  <v-avatar size="32" :color="user.is_current ? 'teal' : 'secondary'" class="mr-2">
                    <span class="text-white text-body-2">{{ user.name[0].toUpperCase() }}</span>
                  </v-avatar>
                  <div>
                    <div
                      class="text-body-2"
                      :class="{ 'font-weight-bold': user.is_current }"
                      
                    >
                      {{ user.name }}{{ user.is_current ? ' (You)' : '' }}
                    </div>
                  </div>
                </div>
                <v-chip size="small" :color="user.is_current ? 'teal' : 'secondary'" variant="flat">
                  <span class="text-white font-weight-bold">{{ user.workout_count }}</span>
                </v-chip>
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>

      <!-- Quick Actions -->
      <v-row dense class="mb-2">
        <v-col cols="6">
          <v-btn
            block
            size="large"
            class="quick-action-btn"
            @click="openQuickLog"
          >
            <v-icon start size="24">mdi-lightning-bolt</v-icon>
            <div class="d-flex flex-column align-start">
              <span class="text-body-2 font-weight-bold">Quick Log</span>
              <span class="text-caption" style="opacity: 0.8; font-size: 9px; line-height: 1">Fast entry</span>
            </div>
          </v-btn>
        </v-col>
        <v-col cols="6">
          <v-btn
            block
            size="large"
            class="quick-action-btn"
            @click="$router.push('/workouts/calendar')"
          >
            <v-icon start size="24">mdi-calendar-month</v-icon>
            <div class="d-flex flex-column align-start">
              <span class="text-body-2 font-weight-bold">Calendar</span>
              <span class="text-caption" style="opacity: 0.8; font-size: 9px; line-height: 1">View by date</span>
            </div>
          </v-btn>
        </v-col>
      </v-row>

      <!-- Recent Workouts -->
      <div class="mb-1">
        <div
          class="d-flex align-center justify-space-between mb-1"
          style="cursor: pointer"
          @click="showRecentWorkouts = !showRecentWorkouts"
        >
          <div class="d-flex align-center">
            <h3 class="text-body-1 font-weight-bold">Last 30 Days</h3>
            <v-icon size="small" class="ml-1" color="medium-emphasis">
              {{ showRecentWorkouts ? 'mdi-chevron-up' : 'mdi-chevron-down' }}
            </v-icon>
          </div>
        </div>

        <div v-if="showRecentWorkouts">
          <!-- Loading State -->
          <div v-if="loading" class="text-center py-8">
            <v-progress-circular indeterminate color="primary" size="48" />
            <p class="mt-2 text-caption text-medium-emphasis">Loading workouts...</p>
          </div>

          <!-- Empty State -->
          <v-card
            v-else-if="!loading && recentWorkouts.length === 0"
            elevation="0"
            rounded
            class="pa-2 text-center"
            bg-color="surface"
          >
            <v-icon size="48" color="surface-variant">mdi-clipboard-text-outline</v-icon>
            <p class="text-body-2 mt-1 mb-0 text-high-emphasis">No workouts logged yet</p>
            <p class="text-caption mb-0 text-medium-emphasis">
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
            style="background-color: rgb(var(--v-theme-surface)); cursor: pointer"
            @click="toggleWorkoutExpand(workout.id)"
          >
            <div class="d-flex align-center mb-1">
              <v-icon color="primary" class="mr-2" size="x-small">mdi-dumbbell</v-icon>
              <div class="flex-grow-1">
                <div class="font-weight-bold text-body-2" >
                  {{ workout.workout_name || 'Workout' }}
                </div>
                <div class="text-medium-emphasis" style="font-size: 11px">
                  {{ formatDate(workout.workout_date) }}
                  <span v-if="workout.total_time"> • {{ formatTime(workout.total_time) }}</span>
                </div>
              </div>
              <v-icon color="primary" size="x-small">
                {{ expandedWorkouts.has(workout.id) ? 'mdi-chevron-up' : 'mdi-chevron-down' }}
              </v-icon>
            </div>

            <!-- Collapsed view - showing summary only -->
            <div v-if="!expandedWorkouts.has(workout.id)">
              <!-- Display movement performance (first 2 only) -->
              <div v-if="workout.performance_movements && workout.performance_movements.length > 0" class="ml-7 mt-2">
                <div v-for="(perf, index) in workout.performance_movements.slice(0, 2)" :key="index" class="text-caption mb-1 text-medium-emphasis">
                  <v-icon size="small" color="primary">mdi-weight-lifter</v-icon>
                  <strong>{{ perf.movement?.name || 'Movement' }}:</strong>
                  <span v-if="perf.weight"> {{ perf.weight }}lbs</span>
                  <span v-if="perf.sets"> {{ perf.sets }}x</span><span v-if="perf.reps">{{ perf.reps }}</span>
                  <span v-if="perf.distance"> {{ perf.distance }}m</span>
                  <span v-if="perf.time_seconds"> {{ formatTime(perf.time_seconds) }}</span>
                  <v-chip v-if="perf.is_pr" color="primary" size="x-small" class="ml-2" variant="flat">
                    <v-icon size="x-small" start>mdi-trophy</v-icon>PR
                  </v-chip>
                </div>
                <div v-if="workout.performance_movements.length > 2" class="text-caption ml-7 text-disabled">
                  +{{ workout.performance_movements.length - 2 }} more...
                </div>
              </div>

              <!-- Display WOD performance (first 1 only) -->
              <div v-if="workout.performance_wods && workout.performance_wods.length > 0" class="ml-7 mt-2">
                <div v-for="(perf, index) in workout.performance_wods.slice(0, 1)" :key="index" class="text-caption mb-1 text-medium-emphasis">
                  <v-icon size="small" color="primary">mdi-fire</v-icon>
                  <strong>{{ perf.wod?.name || 'WOD' }}:</strong>
                  <span v-if="perf.time_seconds"> {{ formatTime(perf.time_seconds) }}</span>
                  <span v-if="perf.rounds && perf.reps"> {{ perf.rounds }}+{{ perf.reps }}</span>
                  <span v-else-if="perf.rounds"> {{ perf.rounds }} rounds</span>
                  <span v-else-if="perf.reps"> {{ perf.reps }} reps</span>
                  <span v-if="perf.score_value"> ({{ perf.score_value }})</span>
                  <v-chip v-if="perf.is_pr" color="primary" size="x-small" class="ml-2" variant="flat">
                    <v-icon size="x-small" start>mdi-trophy</v-icon>PR
                  </v-chip>
                </div>
                <div v-if="workout.performance_wods.length > 1" class="text-caption ml-7 text-disabled">
                  +{{ workout.performance_wods.length - 1 }} more...
                </div>
              </div>

              <!-- Notes preview -->
              <div v-if="workout.notes" class="ml-7 mt-2 text-caption text-medium-emphasis">
                📝 {{ truncateText(workout.notes, 80) }}
              </div>
            </div>

            <!-- Expanded view - showing all details -->
            <div v-else>
              <!-- Display ALL movement performances -->
              <div v-if="workout.performance_movements && workout.performance_movements.length > 0" class="ml-7 mt-2">
                <div class="text-caption font-weight-bold mb-1" >Movements:</div>
                <div v-for="(perf, index) in workout.performance_movements" :key="index" class="text-caption mb-1 ml-2 text-medium-emphasis">
                  <v-icon size="small" color="primary">mdi-weight-lifter</v-icon>
                  <strong>{{ perf.movement?.name || 'Movement' }}:</strong>
                  <span v-if="perf.weight"> {{ perf.weight }}lbs</span>
                  <span v-if="perf.sets"> × {{ perf.sets }} sets</span>
                  <span v-if="perf.reps"> × {{ perf.reps }} reps</span>
                  <span v-if="perf.distance"> × {{ perf.distance }}m</span>
                  <span v-if="perf.time_seconds"> in {{ formatTime(perf.time_seconds) }}</span>
                  <v-chip v-if="perf.is_pr" color="primary" size="x-small" class="ml-2" variant="flat">
                    <v-icon size="x-small" start>mdi-trophy</v-icon>PR
                  </v-chip>
                </div>
              </div>

              <!-- Display ALL WOD performances -->
              <div v-if="workout.performance_wods && workout.performance_wods.length > 0" class="ml-7 mt-2">
                <div class="text-caption font-weight-bold mb-1" >WODs:</div>
                <div v-for="(perf, index) in workout.performance_wods" :key="index" class="text-caption mb-1 ml-2 text-medium-emphasis">
                  <v-icon size="small" color="primary">mdi-fire</v-icon>
                  <strong>{{ perf.wod?.name || 'WOD' }}:</strong>
                  <span v-if="perf.time_seconds"> Time: {{ formatTime(perf.time_seconds) }}</span>
                  <span v-if="perf.rounds && perf.reps"> Score: {{ perf.rounds }}+{{ perf.reps }}</span>
                  <span v-else-if="perf.rounds"> Rounds: {{ perf.rounds }}</span>
                  <span v-else-if="perf.reps"> Reps: {{ perf.reps }}</span>
                  <span v-if="perf.score_value"> ({{ perf.score_value }})</span>
                  <v-chip v-if="perf.is_pr" color="primary" size="x-small" class="ml-2" variant="flat">
                    <v-icon size="x-small" start>mdi-trophy</v-icon>PR
                  </v-chip>
                </div>
              </div>

              <!-- Full notes with markdown rendering -->
              <div v-if="workout.notes" class="ml-7 mt-2 text-caption text-medium-emphasis">
                <div class="font-weight-bold mb-1" >Notes:</div>
                <div class="ml-2">
                  <markdown-renderer :content="workout.notes" />
                </div>
              </div>

              <!-- Action buttons when expanded -->
              <div class="ml-7 mt-3 d-flex gap-2">
                <v-btn
                  size="small"
                  color="primary"
                  
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
      </div>
    </v-container>

    <!-- Quick Log Dialog -->
    <v-dialog v-model="quickLogDialog" max-width="500px">
      <v-card>
        <v-card-title class="text-h6 font-weight-bold" bg-color="primary">
          <v-icon color="white" class="mr-2">mdi-lightning-bolt</v-icon>
          Quick Log Workout
        </v-card-title>

        <v-card-text class="pa-2">
          <v-form ref="quickLogForm" @submit.prevent="submitQuickLog">
            <!-- Date -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" >
                Date *
              </label>
              <v-text-field
                v-model="quickLogData.date"
                type="date"
                
                density="compact"
                hide-details
                required
                @update:model-value="updateQuickLogName"
              />
            </div>

            <!-- Workout Name -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" >
                Workout Name *
              </label>
              <v-text-field
                v-model="quickLogData.name"
                
                density="compact"
                placeholder="e.g., Morning Run, Upper Body, etc."
                hide-details
                required
              />
            </div>

            <!-- Total Time -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" >
                Total Time (minutes)
              </label>
              <v-text-field
                v-model.number="quickLogData.totalTime"
                type="number"
                
                density="compact"
                placeholder="e.g., 30"
                hide-details
                min="0"
              />
            </div>

            <!-- Quick Entry -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" >
                Quick Entry
              </label>
              <v-textarea
                v-model="quickLogData.notes"
                
                density="compact"
                rows="3"
                placeholder="Enter your workout text here, along with any notes. Or pick a WOD or Movement below."
                hide-details
              />
            </div>

            <!-- Unified Search for Movement/WOD -->
            <div class="mb-1">
              <label class="text-caption font-weight-bold d-block" >
                Add Performance Data (Optional)
              </label>
              <v-autocomplete
                v-model="quickLogData.selectedItem"
                :items="unifiedSearchItems"
                item-title="displayName"
                return-object
                :loading="loadingMovements || loadingWods"
                
                density="compact"
                hide-details
                clearable
                auto-select-first
                placeholder="Search for movement or WOD..."
              >
                <template #prepend-inner>
                  <v-icon color="primary" size="small">mdi-magnify</v-icon>
                </template>
                <template #item="{ props, item }">
                  <v-list-item v-bind="props">
                    <template #prepend>
                      <v-icon
                        :color="item.raw.type === 'movement' ? 'teal' : (item.raw.type === 'wod' ? 'teal' : 'secondary')"
                        size="small"
                      >
                        {{ item.raw.type === 'movement' ? 'mdi-weight-lifter' : (item.raw.type === 'wod' ? 'mdi-fire' : 'mdi-clipboard-text') }}
                      </v-icon>
                    </template>
                    <template #append>
                      <v-chip
                        :color="item.raw.type === 'movement' ? 'teal' : (item.raw.type === 'wod' ? 'teal' : 'secondary')"
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

              <!-- Movement Performance Form -->
              <div v-if="quickLogData.selectedItem && quickLogData.selectedItem.type === 'movement'" class="mt-3 pa-3" bg-color="surface-variant" style="; border-radius: 8px">
                <div class="mb-2">
                  <label class="text-caption">Sets</label>
                  <v-text-field
                    v-model.number="quickLogData.movement.sets"
                    type="number"
                    
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
                    
                    density="compact"
                    rows="2"
                    hide-details
                  />
                </div>
              </div>

              <!-- WOD Performance Form -->
              <div v-if="quickLogData.selectedItem && quickLogData.selectedItem.type === 'wod'" class="mt-3 pa-3" bg-color="surface-variant" style="; border-radius: 8px">
                <div class="mb-2">
                  <label class="text-caption">Score Type (from WOD)</label>
                  <v-text-field
                    v-model="quickLogData.wod.scoreType"
                    
                    density="compact"
                    hide-details
                    readonly
                    bg-color="surface"
                  />
                </div>
                <!-- Time-based WOD fields -->
                <div v-if="quickLogData.wod.scoreType === 'Time (HH:MM:SS)'">
                  <label class="text-caption d-block mb-1">Time (HH:MM:SS) *</label>
                  <div class="d-flex gap-2 mb-2">
                    <v-text-field
                      v-model.number="quickLogData.wod.timeHours"
                      type="number"
                      
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
                    
                    density="compact"
                    rows="2"
                    hide-details
                    placeholder="How did it feel?"
                  />
                </div>
              </div>

              <!-- Template Info -->
              <div v-if="quickLogData.selectedItem && quickLogData.selectedItem.type === 'template'" class="mt-3 pa-3 border-warning" bg-color="warning-lighten-4" style="border-radius: 8px">
                <div class="d-flex align-center mb-2">
                  <v-icon color="warning" size="small" class="mr-2">mdi-information</v-icon>
                  <span class="text-caption font-weight-bold text-warning">Template Selected</span>
                </div>
                <p class="text-caption mb-1 text-medium-emphasis">
                  Clicking "Log Workout" will take you to the full workout logging page with the "{{ quickLogData.selectedItem.name }}" template pre-loaded.
                </p>
                <p class="text-caption mb-0 text-warning font-weight-medium">
                  ⚠️ Only the date will be preserved. Notes, workout name, and total time entered here will not be carried over.
                </p>
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
            color="primary"
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

    <!-- Date Workout Detail Dialog -->
    <DateWorkoutDetailDialog
      v-model="showDateDialog"
      :date="selectedWeekDate"
      :preloaded-workouts="userWorkouts"
    />

    </div>
  </pull-to-refresh>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/axios'
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'
import { formatDateInTimezone, getTodayInTimezone } from '@/utils/timezone'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import PullToRefresh from '@/components/PullToRefresh.vue'
import DateWorkoutDetailDialog from '@/components/DateWorkoutDetailDialog.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()

const loading = ref(false)
const userWorkouts = ref([])
const expandedWorkouts = ref(new Set())
const showRecentWorkouts = ref(false) // Collapsed by default
const activeUsersStats = ref([])

// Date detail dialog state
const showDateDialog = ref(false)
const selectedWeekDate = ref(null)

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
const workoutTemplates = ref([])
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

const avgWorkoutsPerWeek = computed(() => {
  const now = new Date()
  const currentYear = now.getFullYear()
  const startOfYear = new Date(currentYear, 0, 1) // January 1st of current year

  // Filter workouts from this year
  const thisYearWorkouts = userWorkouts.value.filter(w => {
    const workoutDate = new Date(w.workout_date)
    return workoutDate.getFullYear() === currentYear
  })

  if (thisYearWorkouts.length === 0) return 0

  // Calculate weeks elapsed in the year
  const daysSinceStart = Math.floor((now - startOfYear) / (1000 * 60 * 60 * 24))
  const weeksElapsed = Math.max(1, daysSinceStart / 7)

  return (thisYearWorkouts.length / weeksElapsed).toFixed(1)
})

const weekDays = computed(() => {
  const now = new Date()
  const dayOfWeek = now.getDay() // 0 = Sunday, 1 = Monday, etc.

  // Adjust to get Monday as start of week
  const mondayOffset = dayOfWeek === 0 ? -6 : 1 - dayOfWeek
  const monday = new Date(now)
  monday.setDate(now.getDate() + mondayOffset)
  monday.setHours(0, 0, 0, 0)

  const days = ['M', 'T', 'W', 'T', 'F', 'S', 'S']

  return days.map((letter, index) => {
    const date = new Date(monday)
    date.setDate(monday.getDate() + index)

    // Format as YYYY-MM-DD for comparison
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const dateStr = `${year}-${month}-${day}`

    // Check if there's a workout on this day
    const hasWorkout = userWorkouts.value.some(w => {
      const workoutDateStr = w.workout_date.split('T')[0]
      return workoutDateStr === dateStr
    })

    return {
      letter,
      date: dateStr,
      hasWorkout
    }
  })
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

// Unified search items combining movements and WODs
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
      displayName: wod.name,
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

  return items.sort((a, b) => a.name.localeCompare(b.name))
})

// Watch for selectedItem changes and auto-set score_type for WODs
watch(() => quickLogData.value.selectedItem, (newItem) => {
  if (newItem && newItem.type === 'wod') {
    // Auto-populate the score type from the WOD definition
    quickLogData.value.wod.scoreType = newItem.data.score_type
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
    // Request all workouts by setting a high limit
    const response = await axios.get('/api/workouts?limit=10000')
    userWorkouts.value = response.data.workouts || []
    console.log('Fetched user workouts:', userWorkouts.value.length)
  } catch (err) {
    console.error('Failed to fetch user workouts:', err)
    userWorkouts.value = []
  } finally {
    loading.value = false
  }
}

async function fetchActiveUsersStats() {
  try {
    const response = await axios.get('/api/stats/active-users-this-month')
    activeUsersStats.value = response.data.users || []
    console.log('Fetched active users stats:', activeUsersStats.value.length)
  } catch (err) {
    console.error('Failed to fetch active users stats:', err)
    activeUsersStats.value = []
  }
}

// Handle pull-to-refresh
async function handleRefresh(done) {
  try {
    await Promise.all([
      fetchUserWorkouts(),
      fetchActiveUsersStats()
    ])
  } finally {
    // Call done callback to stop the refresh animation
    if (done) done()
  }
}

// Format date for display using user's timezone
function formatDate(dateString) {
  const tz = settingsStore.timezone
  const datePart = dateString.split('T')[0]

  // Get today and yesterday in user's timezone
  const todayStr = getTodayInTimezone(tz)
  const todayDate = new Date(todayStr)
  const yesterdayDate = new Date(todayDate)
  yesterdayDate.setDate(yesterdayDate.getDate() - 1)
  const yesterdayStr = yesterdayDate.toISOString().split('T')[0]

  if (datePart === todayStr) {
    return 'Today'
  } else if (datePart === yesterdayStr) {
    return 'Yesterday'
  } else {
    return formatDateInTimezone(dateString, tz, 'EEE, MMM d')
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

// Handle clicking on a week day to show workout details
function handleWeekDayClick(day) {
  // Parse the date string (YYYY-MM-DD) into a Date object
  const [year, month, dayNum] = day.date.split('-').map(Number)
  selectedWeekDate.value = new Date(year, month - 1, dayNum)
  showDateDialog.value = true
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
      const [standardRes, customRes] = await Promise.all([
        axios.get('/api/wods/standard'),
        axios.get('/api/wods/my-wods')
      ])
      const standard = Array.isArray(standardRes.data.wods) ? standardRes.data.wods : []
      const custom = Array.isArray(customRes.data.wods) ? customRes.data.wods : []
      wods.value = [...standard, ...custom]
      console.log('Loaded WODs:', wods.value.length, '(standard:', standard.length, ', custom:', custom.length, ')')
    } catch (error) {
      console.error('Error fetching WODs:', error)
    } finally {
      loadingWods.value = false
    }
  }

  // Fetch workout templates if not already loaded
  if (workoutTemplates.value.length === 0) {
    try {
      const [standardRes, customRes] = await Promise.all([
        axios.get('/api/workouts/standard'),
        axios.get('/api/workouts/my-templates')
      ])

      const standard = Array.isArray(standardRes.data.workouts) ? standardRes.data.workouts : []
      const custom = Array.isArray(customRes.data.workouts) ? customRes.data.workouts : []

      workoutTemplates.value = [...standard, ...custom]
      console.log('Loaded workout templates:', workoutTemplates.value.length)
    } catch (error) {
      console.error('Error fetching workout templates:', error)
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
      timeHours: null,
      timeMinutes: null,
      timeSecondsInput: null,
      rounds: null,
      reps: null,
      weight: null,
      notes: ''
    }
  }
}

// Submit Quick Log
async function submitQuickLog() {
  if (!quickLogData.value.name || !quickLogData.value.date) {
    return
  }

  // If template is selected, navigate to log workout page with template pre-selected
  if (quickLogData.value.selectedItem?.type === 'template') {
    const entityId = quickLogData.value.selectedItem?.entityId
    if (!entityId) {
      console.error('Template selected but entityId is missing:', quickLogData.value.selectedItem)
      alert('Error: Template data is invalid. Please try selecting the template again.')
      return
    }
    closeQuickLog()
    router.push({
      path: '/workouts/log',
      query: {
        template: entityId,
        date: quickLogData.value.date
      }
    })
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
    if (quickLogData.value.selectedItem?.type === 'movement') {
      const entityId = quickLogData.value.selectedItem?.entityId
      if (entityId) {
        payload.movements = [{
          movement_id: entityId,
          sets: quickLogData.value.movement.sets || null,
          reps: quickLogData.value.movement.reps || null,
          weight: quickLogData.value.movement.weight || null,
          time: quickLogData.value.movement.time || null,
          distance: quickLogData.value.movement.distance || null,
          notes: quickLogData.value.movement.notes || '',
          order_index: 0
        }]
      }
    }

    // Add WOD performance data if selected
    if (quickLogData.value.selectedItem?.type === 'wod') {
      const entityId = quickLogData.value.selectedItem?.entityId
      if (entityId) {
        payload.wods = [{
          wod_id: entityId,
          score_type: quickLogData.value.wod.scoreType || null,
          score_value: quickLogData.value.wod.scoreValue || null,
          time_seconds: quickLogData.value.wod.timeSeconds || null,
          rounds: quickLogData.value.wod.rounds || null,
          reps: quickLogData.value.wod.reps || null,
          notes: quickLogData.value.wod.notes || ''
        }]
      }
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

// Watch for route query parameter to auto-open Quick Log
watch(() => route.query.open, (value) => {
  if (value === 'quick-log' && !quickLogDialog.value) {
    openQuickLog()
    // Remove query parameter from URL without navigation
    router.replace({ query: {} })
  }
}, { immediate: true })

// Load data on mount
onMounted(() => {
  fetchUserWorkouts()
  fetchActiveUsersStats()
})
</script>

<style scoped>
/* Dashboard specific styles */

/* Quick action buttons - distinct background that works in all themes */
.quick-action-btn {
  background-color: rgb(var(--v-theme-primary)) !important;
  color: white !important;
  height: 56px !important;
  text-transform: none !important;
  letter-spacing: normal !important;
  border-radius: 12px !important;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15) !important;
  transition: all 0.2s ease !important;
}

.quick-action-btn:hover {
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2) !important;
  transform: translateY(-1px);
}

.quick-action-btn:active {
  transform: translateY(0);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.15) !important;
}

.quick-action-btn .v-icon {
  color: white !important;
}

/* Week day avatar hover effect */
.week-day-avatar {
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.week-day-avatar:hover {
  transform: scale(1.15);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}
</style>
