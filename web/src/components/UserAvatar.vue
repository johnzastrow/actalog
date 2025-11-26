<template>
  <v-avatar :size="size" :color="avatarColor">
    <v-img
      v-if="user?.profile_image && !imageLoadError"
      :src="user.profile_image"
      :alt="user.name || 'User avatar'"
      cover
      @error="handleImageError"
    />
    <span
      v-else
      class="font-weight-bold"
      :style="{ fontSize: fontSize, color: 'white' }"
    >
      {{ initials }}
    </span>
  </v-avatar>
</template>

<script setup>
import { computed, ref, watch } from 'vue'

const props = defineProps({
  user: {
    type: Object,
    default: null
  },
  size: {
    type: [Number, String],
    default: 80
  }
})

// Track image load errors
const imageLoadError = ref(false)

// Reset error state when profile image changes
watch(
  () => props.user?.profile_image,
  () => {
    imageLoadError.value = false
  }
)

// Handle image load error
function handleImageError() {
  console.warn('Failed to load profile image, falling back to initials')
  imageLoadError.value = true
}

// Generate initials from user name
const initials = computed(() => {
  if (!props.user?.name) return 'U'

  const nameParts = props.user.name.trim().split(/\s+/)

  if (nameParts.length === 1) {
    // Single name - use first letter
    return nameParts[0].charAt(0).toUpperCase()
  } else {
    // Multiple names - use first letter of first and last name
    const first = nameParts[0].charAt(0).toUpperCase()
    const last = nameParts[nameParts.length - 1].charAt(0).toUpperCase()
    return first + last
  }
})

// Generate consistent color based on user name
const avatarColor = computed(() => {
  if (!props.user?.name) return '#00bcd4'

  // Hash the name to get a consistent color
  let hash = 0
  const name = props.user.name.toLowerCase()
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }

  // Use predefined color palette for consistent, pleasant colors
  const colors = [
    '#00bcd4', // cyan
    '#4caf50', // green
    '#ff9800', // orange
    '#e91e63', // pink
    '#9c27b0', // purple
    '#2196f3', // blue
    '#f44336', // red
    '#009688', // teal
    '#ff5722', // deep orange
    '#673ab7'  // deep purple
  ]

  return colors[Math.abs(hash) % colors.length]
})

// Calculate font size based on avatar size
const fontSize = computed(() => {
  const size = typeof props.size === 'number' ? props.size : parseInt(props.size)
  return `${Math.floor(size * 0.4)}px`
})
</script>
