<template>
  <v-card elevation="0" rounded="lg" class="pa-4" style="background: white">
    <div class="d-flex align-center justify-space-between mb-3">
      <h3 class="text-h6" style="color: #1a1a1a">Workout Frequency</h3>
      <v-btn-toggle v-model="timeRange" mandatory variant="outlined" density="compact">
        <v-btn value="30">30d</v-btn>
        <v-btn value="90">90d</v-btn>
        <v-btn value="180">6m</v-btn>
        <v-btn value="365">1y</v-btn>
      </v-btn-toggle>
    </div>

    <div v-if="loading" class="text-center py-8">
      <v-progress-circular indeterminate color="#00bcd4"></v-progress-circular>
    </div>

    <div v-else-if="chartData.labels.length === 0" class="text-center py-8" style="color: #666">
      <v-icon size="48" color="#ccc">mdi-chart-bar</v-icon>
      <p class="mt-2">No workout data available</p>
    </div>

    <div v-else style="position: relative; height: 250px">
      <Bar :data="chartData" :options="chartOptions" />
    </div>

    <div v-if="!loading && chartData.labels.length > 0" class="mt-4 d-flex justify-space-around text-center">
      <div>
        <div class="text-h6 font-weight-bold" style="color: #00bcd4">
          {{ totalWorkouts }}
        </div>
        <div class="text-caption" style="color: #666">Total Workouts</div>
      </div>
      <div>
        <div class="text-h6 font-weight-bold" style="color: #4caf50">
          {{ averagePerWeek }}
        </div>
        <div class="text-caption" style="color: #666">Avg/Week</div>
      </div>
      <div>
        <div class="text-h6 font-weight-bold" style="color: #ff9800">
          {{ longestStreak }}
        </div>
        <div class="text-caption" style="color: #666">Longest Streak</div>
      </div>
    </div>
  </v-card>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { Bar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend
} from 'chart.js'
import axios from '@/utils/axios'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend)

const loading = ref(false)
const workouts = ref([])
const timeRange = ref('90') // Default to 90 days

const totalWorkouts = computed(() => workouts.value.length)

const averagePerWeek = computed(() => {
  if (workouts.value.length === 0) return 0
  const days = parseInt(timeRange.value)
  const weeks = days / 7
  return (workouts.value.length / weeks).toFixed(1)
})

const longestStreak = computed(() => {
  if (workouts.value.length === 0) return 0

  const dates = workouts.value
    .map(w => new Date(w.workout_date).toDateString())
    .sort((a, b) => new Date(a) - new Date(b))

  let streak = 1
  let maxStreak = 1

  for (let i = 1; i < dates.length; i++) {
    const prevDate = new Date(dates[i - 1])
    const currDate = new Date(dates[i])
    const diffDays = Math.floor((currDate - prevDate) / (1000 * 60 * 60 * 24))

    if (diffDays === 0) {
      // Same day, multiple workouts
      continue
    } else if (diffDays === 1) {
      streak++
      maxStreak = Math.max(maxStreak, streak)
    } else {
      streak = 1
    }
  }

  return maxStreak
})

const chartData = computed(() => {
  if (workouts.value.length === 0) {
    return { labels: [], datasets: [] }
  }

  // Group workouts by week
  const weekCounts = {}
  const days = parseInt(timeRange.value)
  const startDate = new Date()
  startDate.setDate(startDate.getDate() - days)

  workouts.value.forEach(workout => {
    const date = new Date(workout.workout_date)
    if (date >= startDate) {
      const weekStart = getWeekStart(date)
      const weekLabel = formatWeekLabel(weekStart)
      weekCounts[weekLabel] = (weekCounts[weekLabel] || 0) + 1
    }
  })

  const labels = Object.keys(weekCounts).sort()
  const data = labels.map(label => weekCounts[label])

  return {
    labels,
    datasets: [
      {
        label: 'Workouts',
        data,
        backgroundColor: '#00bcd4',
        borderColor: '#00bcd4',
        borderWidth: 0,
        borderRadius: 4
      }
    ]
  }
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      callbacks: {
        label: function (context) {
          return `${context.parsed.y} workout${context.parsed.y === 1 ? '' : 's'}`
        }
      }
    }
  },
  scales: {
    y: {
      beginAtZero: true,
      ticks: {
        stepSize: 1
      }
    }
  }
}

function getWeekStart(date) {
  const d = new Date(date)
  const day = d.getDay()
  const diff = d.getDate() - day + (day === 0 ? -6 : 1) // Adjust to Monday
  return new Date(d.setDate(diff))
}

function formatWeekLabel(date) {
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

async function fetchWorkouts() {
  loading.value = true
  try {
    const response = await axios.get('/api/workouts')
    const workoutsData = Array.isArray(response.data) ? response.data : (response.data.workouts || [])
    const days = parseInt(timeRange.value)
    const startDate = new Date()
    startDate.setDate(startDate.getDate() - days)

    workouts.value = workoutsData.filter(w => new Date(w.workout_date) >= startDate)
  } catch (error) {
    console.error('Failed to fetch workouts:', error)
    workouts.value = []
  } finally {
    loading.value = false
  }
}

watch(timeRange, () => {
  fetchWorkouts()
})

onMounted(() => {
  fetchWorkouts()
})
</script>
