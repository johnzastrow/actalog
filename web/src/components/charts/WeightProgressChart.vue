<template>
  <v-card elevation="0" rounded="lg" class="pa-4" style="background: white">
    <div class="d-flex align-center justify-space-between mb-3">
      <h3 class="text-h6" style="color: #1a1a1a">Weight Progress</h3>
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
      <v-progress-circular indeterminate color="#00bcd4"></v-progress-circular>
    </div>

    <div v-else-if="!selectedMovementId" class="text-center py-8" style="color: #666">
      <v-icon size="48" color="#ccc">mdi-chart-line</v-icon>
      <p class="mt-2">Select a movement to view progress</p>
    </div>

    <div v-else-if="chartData.labels.length === 0" class="text-center py-8" style="color: #666">
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
        borderColor: '#00bcd4',
        backgroundColor: 'rgba(0, 188, 212, 0.1)',
        fill: true,
        tension: 0.4,
        pointRadius: weights.map((w, idx) => (isPRs[idx] ? 6 : 4)),
        pointBackgroundColor: weights.map((w, idx) => (isPRs[idx] ? '#ffc107' : '#00bcd4')),
        pointBorderColor: weights.map((w, idx) => (isPRs[idx] ? '#ffc107' : '#00bcd4')),
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
