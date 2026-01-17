# Implementation Plan: January 12, 2026 Features

Based on the screenshots and explanations, this document outlines the implementation plan for 5 new features in ActaLog.

---

## Feature 1: RPE (Rate of Perceived Exertion) Tracking

**Priority:** HIGH
**Complexity:** MEDIUM
**Files to modify:** ~12 files

### Overview
Add RPE (2-10 scale) to workout logging for both movements and WODs.

### Visual Reference
Based on RPE-scale_large.webp, the scale uses color coding:
- **10 (Red):** Max effort - Almost impossible to keep going, completely out of breath
- **9 (Red-Orange):** Very hard - Very difficult to maintain, barely breathe/speak
- **7-8 (Orange):** Vigorous - Borderline uncomfortable, short of breath
- **4-6 (Yellow):** Moderate - Breathing heavily, can hold short conversation
- **2-3 (Green):** Light - Can maintain for hours, easy to breathe and talk

### Implementation Steps

#### 1.1 Database Migration
Create migration: `migrations/XXXXXX_add_rpe_columns.up.sql`
```sql
ALTER TABLE user_workout_movements ADD COLUMN rpe INTEGER;
ALTER TABLE user_workout_wods ADD COLUMN rpe INTEGER;
```

#### 1.2 Domain Layer Changes
**File:** `internal/domain/movement.go`
- Add `RPE *int` field to `UserWorkoutMovement` struct (after `Notes` field)

**File:** `internal/domain/wod.go`
- Add `RPE *int` field to `UserWorkoutWOD` struct (after `Notes` field)

#### 1.3 Repository Layer Changes
**File:** `internal/repository/user_workout_movement_repository.go`
- Update INSERT queries to include `rpe` column
- Update SELECT queries to retrieve `rpe`
- Update UPDATE queries to modify `rpe`

**File:** `internal/repository/user_workout_wod_repository.go`
- Same updates as movement repository

**File:** `internal/repository/database.go`
- Update CREATE TABLE statements for `user_workout_movements` and `user_workout_wods`

#### 1.4 Handler Layer Changes
**File:** `internal/handler/user_workout_handler.go`
- Add `RPE *int` to `MovementPerformance` struct
- Add `RPE *int` to `WODPerformance` struct
- Update request parsing and response building

#### 1.5 Frontend Changes
**File:** `web/src/views/LogWorkoutView.vue`
- **Movement Performance section (lines 137-222):** Add RPE v-select after notes textarea (line 220)
- **WOD Performance section (lines 229-350):** Add RPE v-select after notes textarea (line 348)
- Add `rpe` field to `movementPerformance` and `wodPerformance` arrays in script section

**File:** `web/src/views/PerformanceView.vue`
- Display RPE in Performance History entries (lines 282-325)
- Show RPE badge/chip next to date if recorded

**File:** `web/src/components/rpe-scale.js` (NEW - shared utility)
- Export RPE_SCALE constant for reuse across components

#### 1.6 RPE Scale Reference Data (Color-Coded)
Create static reference in frontend matching the visual reference:
```javascript
const RPE_SCALE = [
  { value: null, text: 'Not recorded', color: 'grey' },
  { value: 2, text: '2 - Very light', description: 'Like a slow walk', color: '#4CAF50' },
  { value: 3, text: '3 - Light', description: 'Comfortable, easy conversation', color: '#8BC34A' },
  { value: 4, text: '4 - Light-Moderate', description: 'Comfortable pace', color: '#CDDC39' },
  { value: 5, text: '5 - Moderate', description: 'Breathing heavier, can talk', color: '#FFEB3B' },
  { value: 6, text: '6 - Moderate-Hard', description: 'More challenging', color: '#FFC107' },
  { value: 7, text: '7 - Vigorous', description: 'Uncomfortable, short sentences', color: '#FF9800' },
  { value: 8, text: '8 - Hard', description: 'Heavy breathing, few words', color: '#FF5722' },
  { value: 9, text: '9 - Very hard', description: 'Can barely speak', color: '#F44336' },
  { value: 10, text: '10 - Max effort', description: 'All-out, cannot continue', color: '#D32F2F' }
]
```

---

## Feature 2: Detail View on Card Click

**Priority:** MEDIUM
**Complexity:** LOW-MEDIUM
**Files to modify:** ~6-8 files

### Overview
Clicking any WOD, Template, or Movement card should display a read-only detail view with Back button, QuickLog link, and Edit button.

### Implementation Steps

#### 2.1 Create Reusable Detail Dialog Component
**New file:** `web/src/components/DetailViewDialog.vue`
- Props: `type` (wod|template|movement), `id`, `show`
- Emits: `close`, `quicklog`, `edit`
- Fetch data based on type and id
- Render markdown description
- Include action buttons: Back, QuickLog, Edit

#### 2.2 Update Card Components
Identify all locations where cards are rendered:
- `web/src/views/LogWorkoutView.vue` - Template/WOD selection cards
- `web/src/views/MovementListView.vue` - Movement list
- `web/src/views/WODListView.vue` - WOD list
- `web/src/views/WorkoutTemplateListView.vue` - Template list
- `web/src/views/DashboardView.vue` - Recent activity cards

For each, add:
- `@click` handler on card body (not on existing action buttons)
- Check if click target is an action button; if not, open detail dialog
- Pass entity type and ID to dialog

#### 2.3 Dialog Content Structure
```vue
<v-dialog v-model="showDialog" max-width="600">
  <v-card>
    <v-card-title>{{ entityName }}</v-card-title>
    <v-card-text>
      <!-- Rendered markdown description -->
      <!-- Key details (type, score type, etc.) -->
    </v-card-text>
    <v-card-actions>
      <v-btn @click="close">Back</v-btn>
      <v-spacer />
      <v-btn color="warning" @click="quicklog">QuickLog</v-btn>
      <v-btn color="primary" @click="edit">Edit</v-btn>
    </v-card-actions>
  </v-card>
</v-dialog>
```

---

## Feature 3: Fix Workout Summary Stats

**Priority:** HIGH
**Complexity:** LOW
**Files to modify:** ~2-3 files

### Overview
Verify and fix workout summary statistics on User Profile/Dashboard screen, particularly "All Time" calculations.

### Investigation Steps

#### 3.1 Review Current Implementation
**File:** `web/src/views/DashboardView.vue`
- Examine how stats are calculated/fetched
- Check if "All Time" uses correct date range (no bounds)

**File:** `internal/handler/user_workout_handler.go`
- Review endpoints that provide stats data
- Check query parameters for date filtering

**File:** `internal/repository/user_workout_repository.go`
- Review `ListByUserAndDateRange()` function
- Check if null date bounds are handled correctly

#### 3.2 Potential Issues to Check
1. Date range filters incorrectly applied to "All Time"
2. Off-by-one errors in month/week calculations
3. Timezone handling issues
4. Cached values not updating

#### 3.3 Fix Implementation
Based on investigation findings:
- Update repository queries if date filtering is wrong
- Update frontend stat calculation if client-side issue
- Add tests for edge cases

---

## Feature 4: Percent of Best Section

**Priority:** MEDIUM
**Complexity:** MEDIUM
**Files to modify:** ~1-2 files

### Overview
Add "Percent of Best" section to Movement Stats screen showing weight percentages based on heaviest lift.

### Current PerformanceView.vue Structure (Movement-specific, lines 78-189)
Current order:
1. Heaviest Lifts (lines 79-119) - v-card, NOT collapsible
2. Best Estimated 1RM (lines 121-153) - v-card, NOT collapsible
3. Rep Scheme Dropdown Filter (lines 155-171)
4. Performance Chart (lines 173-188) - v-card, NOT collapsible

**Required Changes:**
1. Wrap all sections in `v-expansion-panels`
2. Reorder to: Best Estimated 1RM → Percent of Best (NEW) → Heaviest Lifts → Performance Chart
3. Default expanded: Heaviest Lifts, Percent of Best
4. Remove Rep Scheme Dropdown (or move inside Performance Chart panel)

### Implementation Steps

#### 4.1 Frontend Component Changes
**File:** `web/src/views/PerformanceView.vue`

Replace lines 79-189 with collapsible expansion panels:
```vue
<!-- MOVEMENT-SPECIFIC CONTENT -->
<template v-if="selectedItem.type === 'movement'">
  <v-expansion-panels v-model="expandedPanels" multiple class="mb-3">

    <!-- 1. Best Estimated 1RM -->
    <v-expansion-panel value="best1rm">
      <v-expansion-panel-title>
        <v-icon color="warning" size="small" class="mr-2">mdi-arm-flex</v-icon>
        Best Estimated 1RM
      </v-expansion-panel-title>
      <v-expansion-panel-text>
        <!-- existing Best 1RM content from lines 128-152 -->
      </v-expansion-panel-text>
    </v-expansion-panel>

    <!-- 2. Percent of Best (NEW) -->
    <v-expansion-panel v-if="showPercentOfBest" value="percentOfBest">
      <v-expansion-panel-title>
        <v-icon color="primary" size="small" class="mr-2">mdi-percent</v-icon>
        Percent of Best
      </v-expansion-panel-title>
      <v-expansion-panel-text>
        <v-table density="compact">
          <thead>
            <tr>
              <th class="text-left">% of Heaviest</th>
              <th class="text-right">{{ weightUnit }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="pct in percentages" :key="pct.percent">
              <td>{{ pct.percent }}%</td>
              <td class="text-right font-weight-bold">{{ pct.weight }}</td>
            </tr>
          </tbody>
        </v-table>
      </v-expansion-panel-text>
    </v-expansion-panel>

    <!-- 3. Heaviest Lifts -->
    <v-expansion-panel value="heaviestLifts">
      <v-expansion-panel-title>
        <v-icon color="primary" size="small" class="mr-2">mdi-trophy</v-icon>
        Heaviest Lifts
      </v-expansion-panel-title>
      <v-expansion-panel-text>
        <!-- existing Heaviest Lifts content from lines 86-117 -->
      </v-expansion-panel-text>
    </v-expansion-panel>

    <!-- 4. Performance Chart -->
    <v-expansion-panel value="performanceChart">
      <v-expansion-panel-title>
        <v-icon color="primary" size="small" class="mr-2">mdi-chart-line</v-icon>
        Performance Chart
      </v-expansion-panel-title>
      <v-expansion-panel-text>
        <!-- Rep Scheme Filter + Chart from lines 156-187 -->
      </v-expansion-panel-text>
    </v-expansion-panel>

  </v-expansion-panels>
</template>
```

#### 4.2 Add Script Section Variables
```javascript
// Default expanded panels (Heaviest Lifts and Percent of Best)
const expandedPanels = ref(['heaviestLifts', 'percentOfBest'])

// User's weight unit from settings
const weightUnit = computed(() => settingsStore.weightUnit || 'lbs')

// Get heaviest weight from data
const heaviestWeight = computed(() => {
  if (heaviestLifts.value.length === 0) return null
  return heaviestLifts.value[0].weight
})

// Show Percent of Best only for movements with weight data
const showPercentOfBest = computed(() => {
  return heaviestWeight.value && heaviestWeight.value > 0
})

// Calculate percentages based on heaviest lift
const percentages = computed(() => {
  if (!heaviestWeight.value) return []
  // Percentages as specified: 95%, 90%, 85%, 80%, 75%, 70%, 65%, 60%, 55%, 50%, 25%
  const percents = [95, 90, 85, 80, 75, 70, 65, 60, 55, 50, 25]
  return percents.map(pct => ({
    percent: pct,
    weight: Math.round(heaviestWeight.value * (pct / 100))
  }))
})
```

#### 4.3 Visual Design
- Two-column grid layout as shown in image_005.jpg
- Column 1: "% of Heaviest" (left-aligned)
- Column 2: Weight value in user's units (right-aligned, bold)
- Clean, compact table styling

---

## Feature 5: Calendar Date Click Detail View

**Priority:** MEDIUM
**Complexity:** MEDIUM
**Files to modify:** ~3-4 files

### Overview
Clicking a date in the calendar should show detailed view of all workouts logged on that date.

### Implementation Steps

#### 5.1 Update Calendar Component
**File:** `web/src/components/WorkoutCalendar.vue`
- Already emits `daySelected` event with Date object
- No changes needed here

#### 5.2 Create Date Workout Detail Component
**New file:** `web/src/components/DateWorkoutDetail.vue`
- Props: `date` (Date object), `show` (boolean)
- Emits: `close`, `quicklog`, `edit`
- Fetches all workouts for the selected date
- Displays each workout with:
  - Template/WOD name
  - Movements performed
  - Scores/times
  - Back button, QuickLog link, Edit button

#### 5.3 API Endpoint (if not exists)
Check if endpoint exists for fetching workouts by date:
- `GET /api/user-workouts?date=YYYY-MM-DD`

If not, add to:
**File:** `internal/handler/user_workout_handler.go`
- Add date query parameter filtering
- Return workouts with full details (movements, WODs)

**File:** `internal/repository/user_workout_repository.go`
- Add `GetByUserIDAndDate(userID, date)` method if not exists

#### 5.4 Update Views Using Calendar
**File:** `web/src/views/DashboardView.vue`
- Handle `@daySelected` event
- Open DateWorkoutDetail dialog

**File:** `web/src/views/WorkoutCalendarView.vue`
- Handle `@daySelected` event
- Open DateWorkoutDetail dialog

#### 5.5 Dialog Structure
```vue
<v-dialog v-model="showDateDetail" max-width="600">
  <v-card>
    <v-card-title>
      {{ formatDate(selectedDate) }}
    </v-card-title>
    <v-card-text>
      <div v-if="workouts.length === 0">
        No workouts logged on this date.
      </div>
      <v-list v-else>
        <v-list-item v-for="workout in workouts" :key="workout.id">
          <v-list-item-title>{{ workout.name || 'Workout' }}</v-list-item-title>
          <v-list-item-subtitle>
            {{ workout.movements?.length || 0 }} movements,
            {{ workout.wods?.length || 0 }} WODs
          </v-list-item-subtitle>
          <template #append>
            <v-btn icon size="small" @click="editWorkout(workout)">
              <v-icon>mdi-pencil</v-icon>
            </v-btn>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
    <v-card-actions>
      <v-btn @click="close">Back</v-btn>
      <v-spacer />
      <v-btn color="warning" @click="quicklogDate">QuickLog</v-btn>
    </v-card-actions>
  </v-card>
</v-dialog>
```

---

## Implementation Order (Recommended)

1. **Feature 3: Fix Workout Summary Stats** - Quick win, investigate and fix
2. **Feature 1: RPE Tracking** - Foundational data model change, do early
3. **Feature 4: Percent of Best** - Frontend-only, can be done in parallel
4. **Feature 2: Detail View on Card Click** - Reusable component, moderate effort
5. **Feature 5: Calendar Date Click** - Builds on Feature 2's patterns

---

## Testing Checklist

### RPE Feature
- [ ] RPE saves correctly for movements
- [ ] RPE saves correctly for WODs
- [ ] RPE displays in workout history
- [ ] RPE is optional (null allowed)
- [ ] RPE validates range (2-10)

### Detail View Feature
- [ ] WOD detail opens on card click
- [ ] Template detail opens on card click
- [ ] Movement detail opens on card click
- [ ] Back button closes dialog
- [ ] QuickLog navigates correctly
- [ ] Edit navigates correctly
- [ ] Existing card buttons still work

### Stats Fix
- [ ] All Time shows correct total
- [ ] This Month shows correct count
- [ ] This Week shows correct count
- [ ] Streak calculates correctly

### Percent of Best
- [ ] Shows for strength movements
- [ ] Hidden for non-weight movements
- [ ] Calculations are correct
- [ ] Section is collapsible
- [ ] Default expanded state correct
- [ ] Section order correct

### Calendar Date Click
- [ ] Click opens detail view
- [ ] Shows all workouts for date
- [ ] Empty state for no workouts
- [ ] Back button works
- [ ] Edit button navigates correctly

---

## Migration Notes

### Database
- Migrations should be backwards compatible
- New columns should be nullable
- No data migration needed (new fields only)

### API
- New fields should be optional in requests
- Existing clients should continue working
- Version bump: 0.22.0 -> 0.23.0 (feature release)

---

## File Reference Summary

| Feature | Backend Files | Frontend Files |
|---------|---------------|----------------|
| **1. RPE** | `domain/movement.go`, `domain/wod.go`, `repository/user_workout_movement_repository.go`, `repository/user_workout_wod_repository.go`, `repository/database.go`, `handler/user_workout_handler.go` | `views/LogWorkoutView.vue`, `views/PerformanceView.vue`, `components/rpe-scale.js` (new) |
| **2. Detail View** | None | `components/DetailViewDialog.vue` (new), `views/LogWorkoutView.vue`, various list views |
| **3. Stats Fix** | `handler/user_workout_handler.go`, `repository/user_workout_repository.go` | `views/DashboardView.vue` |
| **4. Percent of Best** | None | `views/PerformanceView.vue` |
| **5. Calendar Click** | `handler/user_workout_handler.go` (if date filter needed) | `components/DateWorkoutDetail.vue` (new), `views/DashboardView.vue`, `views/WorkoutCalendarView.vue` |

---

## Quick Start Commands

```bash
# Create RPE migration
make migrate-create name=add_rpe_columns

# Run backend tests
make test

# Run frontend dev server
cd web && npm run dev

# Run frontend tests
cd web && npm run test:run
```

---

## Notes

1. **RPE Percentages Note:** The original spec listed "75%, 60%, 65%" which appears to be a typo. Implementation assumes descending order: 95%, 90%, 85%, 80%, 75%, 70%, 65%, 60%, 55%, 50%, 25%.

2. **Detail View Component:** Consider making a single reusable `DetailViewDialog.vue` component that accepts entity type (wod/template/movement) and ID, rather than three separate components.

3. **Calendar Integration:** The existing `WorkoutCalendar.vue` already emits `daySelected`. Only need to add handler in parent views.

4. **Stats Fix Priority:** This should be investigated first as it may reveal calculation bugs that affect other features.
