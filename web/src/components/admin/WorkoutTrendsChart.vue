<template>
  <v-card elevation="0" rounded="lg" class="pa-4" bg-color="surface">
    <div class="d-flex align-center justify-space-between mb-3">
      <h3 class="text-h6" style="color: rgb(var(--v-theme-on-surface))">Workout Trends (30 Days)</h3>
    </div>

    <div v-if="loading" class="text-center py-8">
      <v-progress-circular indeterminate color="primary"></v-progress-circular>
    </div>

    <div v-else-if="!trends || trends.length === 0" class="text-center py-8 text-medium-emphasis">
      <v-icon size="48" color="#ccc">mdi-chart-bar</v-icon>
      <p class="mt-2">No workout data available</p>
    </div>

    <div v-else style="position: relative; height: 250px">
      <bar :data="chartData" :options="chartOptions" />
    </div>

    <div v-if="!loading && trends && trends.length > 0" class="mt-4 d-flex justify-space-around text-center">
      <div>
        <div class="text-h6 font-weight-bold" style="color: rgb(var(--v-theme-primary))">
          {{ totalWorkouts }}
        </div>
        <div class="text-caption text-medium-emphasis">Total (30d)</div>
      </div>
      <div>
        <div class="text-h6 font-weight-bold" style="color: rgb(var(--v-theme-success))">
          {{ averagePerDay }}
        </div>
        <div class="text-caption text-medium-emphasis">Avg/Day</div>
      </div>
      <div>
        <div class="text-h6 font-weight-bold" style="color: rgb(var(--v-theme-warning))">
          {{ peakDay }}
        </div>
        <div class="text-caption text-medium-emphasis">Peak Day</div>
      </div>
    </div>
  </v-card>
</template>

<script setup>
import { computed } from 'vue'
import { useTheme } from 'vuetify'
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

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend)

const props = defineProps({
  trends: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const theme = useTheme()

const primaryColor = computed(() => {
  const colors = theme.current.value.colors
  return colors.primary || '#00bcd4'
})

const totalWorkouts = computed(() => {
  if (!props.trends) return 0
  return props.trends.reduce((sum, t) => sum + (t.count || 0), 0)
})

const averagePerDay = computed(() => {
  if (!props.trends || props.trends.length === 0) return '0.0'
  return (totalWorkouts.value / 30).toFixed(1)
})

const peakDay = computed(() => {
  if (!props.trends || props.trends.length === 0) return 0
  return Math.max(...props.trends.map(t => t.count || 0))
})

const chartData = computed(() => {
  if (!props.trends || props.trends.length === 0) {
    return { labels: [], datasets: [] }
  }

  // Format the period dates for display
  const labels = props.trends.map(t => {
    const date = new Date(t.period)
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  })

  const data = props.trends.map(t => t.count || 0)

  return {
    labels,
    datasets: [
      {
        label: 'Workouts',
        data,
        backgroundColor: primaryColor.value,
        borderColor: primaryColor.value,
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
    x: {
      ticks: {
        maxRotation: 45,
        minRotation: 45,
        maxTicksLimit: 10
      }
    },
    y: {
      beginAtZero: true,
      ticks: {
        stepSize: 1
      }
    }
  }
}
</script>
