<template>
  <div class="notification-likes">
    <!-- Like button with count -->
    <v-btn
      :color="hasLiked ? 'primary' : 'grey'"
      variant="text"
      size="small"
      :loading="liking"
      @click="toggleLike"
    >
      <v-icon left>{{ hasLiked ? 'mdi-thumb-up' : 'mdi-thumb-up-outline' }}</v-icon>
      {{ likeCount }}
    </v-btn>

    <!-- Names of who liked (always visible) -->
    <p v-if="likers.length > 0" class="text-caption text-grey mt-1 mb-0">
      {{ likerNames }}
    </p>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from '@/utils/axios'

const props = defineProps({
  notificationId: {
    type: Number,
    required: true
  }
})

const emit = defineEmits(['liked', 'unliked'])

const likeCount = ref(0)
const likers = ref([])
const hasLiked = ref(false)
const liking = ref(false)

// Computed property for comma-separated liker names
const likerNames = computed(() => {
  const names = likers.value.map(liker => liker.user_name).join(', ')
  return names ? `Liked by: ${names}` : ''
})

async function loadLikes() {
  try {
    const { data } = await axios.get(`/api/notifications/${props.notificationId}/likes`)
    likers.value = data.likes || []
    likeCount.value = data.count || 0

    // Check if current user has liked
    const userStr = localStorage.getItem('user')
    if (userStr) {
      const user = JSON.parse(userStr)
      hasLiked.value = likers.value.some(like => like.user_id === user.id)
    }
  } catch (error) {
    console.error('Failed to load likes:', error)
  }
}

async function toggleLike() {
  liking.value = true
  try {
    if (hasLiked.value) {
      await axios.delete(`/api/notifications/${props.notificationId}/like`)
      hasLiked.value = false
      likeCount.value--
      emit('unliked')
    } else {
      await axios.post(`/api/notifications/${props.notificationId}/like`)
      hasLiked.value = true
      likeCount.value++
      emit('liked')
    }
    await loadLikes() // Refresh the list
  } catch (error) {
    console.error('Failed to toggle like:', error)
  } finally {
    liking.value = false
  }
}

onMounted(() => {
  loadLikes()
})
</script>

<style scoped>
.notification-likes {
  margin-top: 8px;
}
</style>
