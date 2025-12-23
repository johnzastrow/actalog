# Progress Charts Design

> **Status:** Planning
> **Last Updated:** 2025-12-23

---

## Overview

This document describes the design for visual progress charts in ActaLog, covering PR calculations, existing components, integration points, and planned enhancements.

---

## PR Calculation System (Existing)

### 1RM Calculation (`pkg/prmath/one_rm.go`)

ActaLog uses a **hybrid formula approach** based on rep count:

| Rep Range | Formula | Description |
|-----------|---------|-------------|
| 1 rep | Actual | Weight lifted is the 1RM |
| 2-10 reps | Epley | `1RM = weight × (1 + reps/30)` |
| 11+ reps | Wathan | `1RM = (100 × weight) / (48.8 + 53.8 × e^(-0.075 × reps))` |

**Additional formulas available** (for comparison display):
- Brzycki: `1RM = weight × (36 / (37 - reps))`
- Lombardi: `1RM = weight × reps^0.10`
- Mayhew: `1RM = (100 × weight) / (52.2 + 41.9 × e^(-0.055 × reps))`
- O'Conner: `1RM = weight × (1 + reps/40)`

**Helper functions:**
- `CompareToBaseline(current1RM, baseline1RM)` → percentage improvement
- `CalculateIntensity(weight, oneRM)` → percentage of 1RM being used

### PR Detection (`internal/service/user_workout_service.go`)

#### Movement PRs (Weight-Based)
```go
func DetectAndFlagMovementPRs(userID int64, movements []*domain.UserWorkoutMovement)
```
- Calculates 1RM for each movement with weight+reps
- Compares against previous best 1RM for that movement
- First time = automatic PR
- Higher 1RM than previous best = PR

#### WOD PRs (Time/Rounds-Based)
```go
func DetectAndFlagWODPRs(userID int64, wods []*domain.UserWorkoutWOD)
```

**Time-based WODs (For Time):**
- Uses `GetBestTimeForWOD(userID, wodID)`
- Faster time = PR

**AMRAP WODs:**
- Uses `GetBestRoundsRepsForWOD(userID, wodID)`
- More rounds = PR
- Same rounds, more reps = PR

---

## Existing Chart Components

### 1. WeightProgressChart.vue
**Location:** `web/src/components/charts/WeightProgressChart.vue`

**Features:**
- Line chart with Chart.js
- Movement selector (filters to Weightlifting type)
- Fetches from `/api/performance/movements/{id}`
- PRs highlighted in gold/warning color (larger points)
- Tooltip shows weight + PR indicator
- Theme-aware colors
- Responsive 250px height

**Data shown:**
- X-axis: Workout dates (MMM DD format)
- Y-axis: Weight (lbs)
- PR points: Highlighted with larger radius and warning color

### 2. WorkoutFrequencyChart.vue
**Location:** `web/src/components/charts/WorkoutFrequencyChart.vue`

**Features:**
- Bar chart with Chart.js
- Time range selector: 30d / 90d / 6m / 1y
- Fetches from `/api/workouts`
- Groups workouts by week
- Summary stats below chart

**Data shown:**
- X-axis: Week start dates (MMM DD format)
- Y-axis: Workout count per week
- Summary: Total Workouts, Avg/Week, Longest Streak

---

## API Endpoints

| Endpoint | Handler | Returns |
|----------|---------|---------|
| `GET /api/performance/search` | UnifiedSearch | Movements + WODs matching query |
| `GET /api/performance/movements/{id}` | GetMovementPerformance | All performances + calculated 1RM + best 1RM |
| `GET /api/performance/wods/{id}` | GetWODPerformance | All WOD performances |
| `GET /api/workouts/personal-records` | GetPersonalRecords | All PRs (movements + WODs) |
| `GET /api/workouts/personal-records/movements` | GetPRMovements | Movement PR summaries |
| `POST /api/workouts/personal-records/toggle` | ToggleMovementPR | Toggle PR flag |

---

## Integration Plan

### Phase 1: Integrate Existing Charts

#### 1.1 Dashboard View (`DashboardView.vue`)
Add WorkoutFrequencyChart below the welcome section:

```
┌─────────────────────────────────────────────┐
│ Welcome back, [User]!                       │
├─────────────────────────────────────────────┤
│ Workout Frequency              30d 90d 6m 1y│
│ ┌─────────────────────────────────────────┐ │
│ │ █   █       █ █ █   █ █     █   █ █     │ │
│ │ █ █ █ █   █ █ █ █ █ █ █ █   █ █ █ █ █   │ │
│ └─────────────────────────────────────────┘ │
│ Total: 24  |  Avg/Week: 4.2  |  Streak: 5   │
├─────────────────────────────────────────────┤
│ Recent Workouts                             │
│ • Dec 22 - Full Body...                     │
│ • Dec 20 - Upper Body...                    │
└─────────────────────────────────────────────┘
```

#### 1.2 Performance View (`PerformanceView.vue`)
Add WeightProgressChart when a movement is selected:

```
┌─────────────────────────────────────────────┐
│ Search for a WOD or Movement...             │
├─────────────────────────────────────────────┤
│ [Quick Log Back Squat]                      │
├─────────────────────────────────────────────┤
│ Weight Progress         Select Movement ▼   │
│ ┌─────────────────────────────────────────┐ │
│ │        *                                │ │
│ │      /   \    *  <- PR (gold)           │ │
│ │    /       \/                           │ │
│ │  /                                      │ │
│ └─────────────────────────────────────────┘ │
├─────────────────────────────────────────────┤
│ Heaviest Lifts: 315 | 305 | 295            │
├─────────────────────────────────────────────┤
│ Performance History (table)                 │
└─────────────────────────────────────────────┘
```

#### 1.3 Profile View (`ProfileView.vue`)
Add WorkoutFrequencyChart in workout summary section:

```
┌─────────────────────────────────────────────┐
│ [Avatar] John Doe                           │
│ Member since Jan 2024                       │
├─────────────────────────────────────────────┤
│ Workout Activity              30d 90d 6m 1y │
│ ┌─────────────────────────────────────────┐ │
│ │ [WorkoutFrequencyChart]                 │ │
│ └─────────────────────────────────────────┘ │
├─────────────────────────────────────────────┤
│ Personal Records Summary                    │
└─────────────────────────────────────────────┘
```

---

### Phase 2: New Charts

#### 2.1 Estimated1RMChart.vue
**Purpose:** Track calculated 1RM progression over time for a movement

**Data source:** `/api/performance/movements/{id}`

**Display:**
- Line chart showing estimated 1RM over time
- Current formula shown (Epley/Wathan/Actual)
- Percentage improvement from first data point
- Goal line (optional user-set target)

**Features:**
- Time range selector (3m/6m/1y/all)
- Compare formulas toggle
- Show actual weights as secondary series

#### 2.2 WODTimeChart.vue
**Purpose:** Track time progression for "For Time" WODs

**Data source:** `/api/performance/wods/{id}`

**Display:**
- Line chart with time (inverted - lower is better)
- PRs highlighted in gold
- Time format: MM:SS or H:MM:SS

**Use case:** Track Fran, Murph, Grace times

#### 2.3 WODAMRAPChart.vue
**Purpose:** Track rounds+reps progression for AMRAP WODs

**Data source:** `/api/performance/wods/{id}`

**Display:**
- Bar or line chart showing total score (rounds × reps_per_round + extra_reps)
- Or stacked bar: rounds + remaining reps
- PRs highlighted

**Use case:** Track Cindy, Mary scores

#### 2.4 VolumeChart.vue
**Purpose:** Track total training volume over time

**Calculation:** `Σ (weight × reps × sets)` per session

**Data source:** `/api/workouts` with movement details

**Display:**
- Area or bar chart by week/month
- Filter by movement type (all, upper, lower, full body)
- Compare periods

#### 2.5 PRDistributionChart.vue
**Purpose:** Show distribution of PRs by movement type

**Data source:** `/api/workouts/personal-records/movements`

**Display:**
- Pie or donut chart
- Segments: Weightlifting, Gymnastics, Cardio, etc.
- Click segment to filter PR list

---

### Phase 3: Enhanced Analytics

#### 3.1 MonthlyComparisonChart.vue
**Purpose:** Compare this month vs last month

**Display:**
- Side-by-side bars
- Metrics: workouts, PRs, total volume, avg frequency
- Trend arrows (up/down/same)

#### 3.2 BodyweightMovementProgressChart.vue
**Purpose:** Track progress on movements without weight (pullups, pushups, etc.)

**Metrics:**
- Max unbroken reps
- Total reps per session
- Time to complete fixed reps

---

## Implementation Priority

### Priority 1 (Quick Win - 1-2 hours)
1. Add `WorkoutFrequencyChart` to `DashboardView.vue`
2. Add `WeightProgressChart` to `PerformanceView.vue`

### Priority 2 (Medium - 4-6 hours)
3. Create `Estimated1RMChart.vue`
4. Create `WODTimeChart.vue`
5. Create `WODAMRAPChart.vue`

### Priority 3 (Enhanced - 4-6 hours)
6. Create `VolumeChart.vue`
7. Create `PRDistributionChart.vue`
8. Create `MonthlyComparisonChart.vue`

---

## Technical Notes

### Chart.js Configuration
All charts use:
- `vue-chartjs` wrapper
- Theme-aware colors via `useTheme()`
- Responsive sizing with `maintainAspectRatio: false`
- Custom tooltips for contextual data

### Color Scheme
| Element | Color | CSS Variable |
|---------|-------|--------------|
| Primary data | Theme primary | `--v-theme-primary` |
| PR highlights | Warning/Gold | `--v-theme-warning` |
| Fill areas | Primary @ 10% opacity | Computed RGBA |
| Grid lines | Surface variant | `--v-theme-surface-variant` |

### Accessibility
- Charts should have aria-labels
- Color alone should not convey meaning (use shapes/labels)
- Provide data tables as alternative

---

## Future Considerations

1. **Export charts as images** - For sharing on social media
2. **Comparison mode** - Compare two time periods or two movements
3. **Goal tracking** - User-set targets with progress indicators
4. **Prediction lines** - Extrapolate trends to predict future PRs
5. **Coach view** - Aggregate charts for multiple athletes
