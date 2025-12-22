<template>
  <v-card elevation="0" rounded="lg" class="pa-4" bg-color="surface">
    <div class="d-flex align-center justify-space-between mb-3">
      <h3 class="text-h6" style="color: rgb(var(--v-theme-on-surface))">Weight Progress</h3>
      <v-menu>
        <template #activator="{ props }">
          <v-btn
            v-bind="props"
            variant="text"
            size="small"
            append-icon="mdi-chevron-down"
          >
            {{ selectedMovementName || 'Select Movement' }}
          </v-btn>
        </template>
        <v-list>
          <v-list-item
            v-for="movement in movements"
            :key="movement.id"
            @click="selectMovement(movement)"
          >
            <v-list-item-title>{{ movement.name }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </div>

    <div v-if="loading" class="text-center py-8">
      <v-progress-circular indeterminate color="primary"></v-progress-circular>
    </div>

    <div v-else-if="!selectedMovementId" class="text-center py-8 text-medium-emphasis">
      <v-icon size="48" color="#ccc">mdi-chart-line</v-icon>
      <p class="mt-2">Select a movement to view progress</p>
    </div>

    <div v-else-if="chartData.labels.length === 0" class="text-center py-8 text-medium-emphasis">
      <v-icon size="48" color="#ccc">mdi-chart-line-variant</v-icon>
      <p class="mt-2">No weight data for this movement</p>
    </div>

    <div v-else style="position: relative; height: 250px">
      <Line :data="chartData" :options="chartOptions" />
    </div>
  </v-card>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useTheme } from 'vuetify'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import axios from '@/utils/axios'

const theme = useTheme()

// Get theme colors as hex values for Chart.js
const primaryColor = computed(() => {
  const colors = theme.current.value.colors
  return colors.primary || '#00bcd4'
})

const warningColor = computed(() => {
  const colors = theme.current.value.colors
  return colors.warning || '#ffc107'
})

const primaryColorWithAlpha = computed(() => {
  // Convert hex to rgba with 0.1 opacity for fill
  const hex = primaryColor.value.replace('#', '')
  const r = parseInt(hex.substring(0, 2), 16)
  const g = parseInt(hex.substring(2, 4), 16)
  const b = parseInt(hex.substring(4, 6), 16)
  return `rgba(${r}, ${g}, ${b}, 0.1)`
})

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const props = defineProps({
  movementId: {
    type: Number,
    default: null
  }
})

const loading = ref(false)
const movements = ref([])
const selectedMovementId = ref(props.movementId)
const selectedMovementName = ref('')
const performanceData = ref([])

const chartData = computed(() => {
  if (performanceData.value.length === 0) {
    return {
      labels: [],
      datasets: []
    }
  }

  const labels = performanceData.value.map(p => {
    const date = new Date(p.workout_date)
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  })

  const weights = performanceData.value.map(p => p.weight || 0)
  const isPRs = performanceData.value.map(p => p.is_pr)

  return {
    labels,
    datasets: [
      {
        label: 'Weight (lbs)',
        data: weights,
        borderColor: primaryColor.value,
        backgroundColor: primaryColorWithAlpha.value,
        fill: true,
        tension: 0.4,
        pointRadius: weights.map((w, idx) => (isPRs[idx] ? 6 : 4)),
        pointBackgroundColor: weights.map((w, idx) => (isPRs[idx] ? warningColor.value : primaryColor.value)),
        pointBorderColor: weights.map((w, idx) => (isPRs[idx] ? warningColor.value : primaryColor.value)),
        pointBorderWidth: 2
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
          const isPR = performanceData.value[context.dataIndex]?.is_pr
          const label = `${context.parsed.y} lbs${isPR ? ' (PR)' : ''}`
          return label
        }
      }
    }
  },
  scales: {
    y: {
      beginAtZero: false,
      ticks: {
        callback: function (value) {
          return value + ' lbs'
        }
      }
    }
  }
}

async function fetchMovements() {
  try {
    const response = await axios.get('/api/movements')
    const movementsData = Array.isArray(response.data) ? response.data : (response.data.movements || [])
    movements.value = movementsData.filter(m => m.type === 'Weightlifting')
  } catch (error) {
    console.error('Failed to fetch movements:', error)
  }
}

async function fetchPerformanceData(movementId) {
  if (!movementId) return

  loading.value = true
  try {
    const response = await axios.get(`/api/performance/movements/${movementId}`)
    performanceData.value = response.data.reverse() // Chronological order
  } catch (error) {
    console.error('Failed to fetch performance data:', error)
    performanceData.value = []
  } finally {
    loading.value = false
  }
}

function selectMovement(movement) {
  selectedMovementId.value = movement.id
  selectedMovementName.value = movement.name
  fetchPerformanceData(movement.id)
}

watch(() => props.movementId, (newId) => {
  if (newId) {
    selectedMovementId.value = newId
    const movement = movements.value.find(m => m.id === newId)
    if (movement) {
      selectedMovementName.value = movement.name
    }
    fetchPerformanceData(newId)
  }
})

onMounted(async () => {
  await fetchMovements()
  if (props.movementId) {
    const movement = movements.value.find(m => m.id === props.movementId)
    if (movement) {
      selectedMovementName.value = movement.name
    }
    fetchPerformanceData(props.movementId)
  }
})
</script>
