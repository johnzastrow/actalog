<template>
  <div
    ref="container"
    class="pull-to-refresh-container"
    @touchstart="onTouchStart"
    @touchmove="onTouchMove"
    @touchend="onTouchEnd"
  >
    <!-- Pull indicator -->
    <div
      class="pull-indicator"
      :class="{ visible: pullDistance > 0, refreshing: isRefreshing }"
      :style="{ transform: `translateY(${Math.min(pullDistance, maxPullDistance)}px)` }"
    >
      <v-progress-circular
        v-if="isRefreshing"
        indeterminate
        size="24"
        width="2"
        color="primary"
      />
      <v-icon
        v-else
        :style="{ transform: `rotate(${pullRotation}deg)` }"
        color="grey"
      >
        mdi-arrow-down
      </v-icon>
      <span class="pull-text">
        {{ isRefreshing ? 'Refreshing...' : (pullDistance >= threshold ? 'Release to refresh' : 'Pull to refresh') }}
      </span>
    </div>

    <!-- Content slot -->
    <div
      class="pull-content"
      :style="{ transform: `translateY(${isRefreshing ? indicatorHeight : Math.min(pullDistance, maxPullDistance)}px)` }"
    >
      <slot />
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  threshold: {
    type: Number,
    default: 60
  },
  maxPullDistance: {
    type: Number,
    default: 100
  },
  indicatorHeight: {
    type: Number,
    default: 50
  }
})

const emit = defineEmits(['refresh'])

const container = ref(null)
const startY = ref(0)
const pullDistance = ref(0)
const isRefreshing = ref(false)
const isPulling = ref(false)

const pullRotation = computed(() => {
  return Math.min(pullDistance.value / props.threshold * 180, 180)
})

function canPull() {
  // Only allow pull when scrolled to top
  return window.scrollY === 0
}

function onTouchStart(e) {
  if (!canPull() || isRefreshing.value) return

  startY.value = e.touches[0].clientY
  isPulling.value = true
}

function onTouchMove(e) {
  if (!isPulling.value || isRefreshing.value) return
  if (!canPull()) {
    pullDistance.value = 0
    return
  }

  const currentY = e.touches[0].clientY
  const diff = currentY - startY.value

  if (diff > 0) {
    // Apply resistance factor for smoother feel
    pullDistance.value = diff * 0.5

    // Prevent default scroll when pulling
    if (pullDistance.value > 10) {
      e.preventDefault()
    }
  }
}

function onTouchEnd() {
  if (!isPulling.value) return
  isPulling.value = false

  if (pullDistance.value >= props.threshold && !isRefreshing.value) {
    // Trigger refresh
    isRefreshing.value = true
    pullDistance.value = 0

    emit('refresh', () => {
      // Callback to stop refreshing
      isRefreshing.value = false
    })
  } else {
    pullDistance.value = 0
  }
}

// Expose method to programmatically stop refresh
defineExpose({
  stopRefresh: () => {
    isRefreshing.value = false
  }
})
</script>

<style scoped>
.pull-to-refresh-container {
  position: relative;
  overflow: hidden;
  min-height: 100%;
}

.pull-indicator {
  position: absolute;
  top: -50px;
  left: 0;
  right: 0;
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.2s ease;
  z-index: 1;
}

.pull-indicator.visible {
  opacity: 1;
}

.pull-indicator.refreshing {
  opacity: 1;
  transform: translateY(50px) !important;
}

.pull-text {
  font-size: 12px;
  color: #666;
}

.pull-content {
  transition: transform 0.2s ease;
  will-change: transform;
}

.pull-indicator .v-icon {
  transition: transform 0.1s ease;
}
</style>
