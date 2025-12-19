import { defineStore } from 'pinia'
import axios from '@/utils/axios'

export const useSubscriptionStore = defineStore('subscription', {
  state: () => ({
    subscriptionStatus: null,
    loading: false,
    error: null,
    lastFetched: null
  }),

  getters: {
    /**
     * Check if user has write access (active subscription)
     */
    hasAccess: (state) => {
      return state.subscriptionStatus?.has_access || false
    },

    /**
     * Check if user's subscription is expired
     */
    isExpired: (state) => {
      return state.subscriptionStatus?.has_access === false
    },

    /**
     * Get expiration date if available
     */
    expiresAt: (state) => {
      return state.subscriptionStatus?.expires_at || null
    },

    /**
     * Get source of access (user, organization, both, none)
     */
    source: (state) => {
      return state.subscriptionStatus?.source || 'none'
    },

    /**
     * Get user's personal subscription
     */
    userSubscription: (state) => {
      return state.subscriptionStatus?.user_subscription || null
    },

    /**
     * Get organization subscriptions user benefits from
     */
    orgSubscriptions: (state) => {
      return state.subscriptionStatus?.org_subscriptions || []
    },

    /**
     * Check if user has permanent free subscription
     */
    isPermanentFree: (state) => {
      const userSub = state.subscriptionStatus?.user_subscription
      return userSub?.is_permanent_free || false
    },

    /**
     * Get subscription type (free, monthly, annual)
     */
    subscriptionType: (state) => {
      const userSub = state.subscriptionStatus?.user_subscription
      return userSub?.subscription_type || 'free'
    },

    /**
     * Get formatted subscription type for display
     */
    subscriptionTypeLabel: (state) => {
      const userSub = state.subscriptionStatus?.user_subscription
      if (!userSub) return 'No Subscription'

      if (userSub.is_permanent_free) {
        return 'Permanent Free'
      }

      const typeMap = {
        free: 'Free',
        monthly: 'Monthly',
        annual: 'Annual'
      }
      return typeMap[userSub.subscription_type] || 'Unknown'
    },

    /**
     * Get subscription status (active, expired, cancelled)
     */
    status: (state) => {
      const userSub = state.subscriptionStatus?.user_subscription
      return userSub?.status || 'unknown'
    },

    /**
     * Check if subscription data is stale (> 5 minutes old)
     */
    isStale: (state) => {
      if (!state.lastFetched) return true
      const fiveMinutes = 5 * 60 * 1000
      return Date.now() - state.lastFetched > fiveMinutes
    }
  },

  actions: {
    /**
     * Fetch subscription status from API
     */
    async fetchStatus() {
      // Skip if already loading
      if (this.loading) return

      // Use cached data if fresh
      if (!this.isStale && this.subscriptionStatus) {
        return this.subscriptionStatus
      }

      this.loading = true
      this.error = null

      try {
        const response = await axios.get('/api/subscriptions/status')
        this.subscriptionStatus = response.data
        this.lastFetched = Date.now()
        return response.data
      } catch (err) {
        console.error('Failed to fetch subscription status:', err)
        this.error = err.response?.data?.error || 'Failed to load subscription status'

        // On 401, user is not authenticated - set default state
        if (err.response?.status === 401) {
          this.subscriptionStatus = null
        }

        throw err
      } finally {
        this.loading = false
      }
    },

    /**
     * Refresh subscription status (force fetch)
     */
    async refresh() {
      this.lastFetched = null
      return this.fetchStatus()
    },

    /**
     * Clear subscription status (on logout)
     */
    clear() {
      this.subscriptionStatus = null
      this.loading = false
      this.error = null
      this.lastFetched = null
    },

    /**
     * Set expired state (called when receiving 402 response)
     */
    setExpired() {
      if (this.subscriptionStatus) {
        this.subscriptionStatus.has_access = false
        this.subscriptionStatus.source = 'none'
      }
    }
  }
})
