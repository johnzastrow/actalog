import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/dashboard'
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('@/views/ForgotPasswordView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/reset-password/:token',
      name: 'reset-password',
      component: () => import('@/views/ResetPasswordView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/verify-email',
      name: 'verify-email',
      component: () => import('@/views/VerifyEmailView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('@/views/DashboardView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workouts',
      name: 'workouts',
      component: () => import('@/views/WorkoutsView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workouts/log',
      name: 'log-workout',
      component: () => import('@/views/LogWorkoutView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workouts/templates/create',
      name: 'create-template',
      component: () => import('@/views/WorkoutTemplateEditView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workouts/templates/:id/edit',
      name: 'edit-template',
      component: () => import('@/views/WorkoutTemplateEditView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workouts/:id',
      name: 'workout-detail',
      component: () => import('@/views/WorkoutDetailView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workouts/calendar',
      name: 'workout-calendar',
      component: () => import('@/views/WorkoutCalendarView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/movements',
      name: 'movements-library',
      component: () => import('@/views/MovementsLibraryView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/movements/create',
      name: 'create-movement',
      component: () => import('@/views/MovementCreateView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/movements/:id/edit',
      name: 'edit-movement',
      component: () => import('@/views/MovementCreateView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/movements/:id',
      name: 'movement-detail',
      component: () => import('@/views/MovementDetailView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/wods',
      name: 'wods-library',
      component: () => import('@/views/WODLibraryView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/wods/create',
      name: 'create-wod',
      component: () => import('@/views/WODCreateView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/wods/:id/edit',
      name: 'edit-wod',
      component: () => import('@/views/WODCreateView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/wods/:id',
      name: 'wod-detail',
      component: () => import('@/views/WODDetailView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/performance',
      name: 'performance',
      component: () => import('@/views/PerformanceView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/prs',
      name: 'pr-history',
      component: () => import('@/views/PRHistoryView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('@/views/ProfileView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/SettingsView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/settings/export',
      name: 'export-data',
      component: () => import('@/views/ExportView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/settings/import',
      name: 'import-data',
      component: () => import('@/views/ImportView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/admin/data-cleanup',
      name: 'admin-data-cleanup',
      component: () => import('@/views/AdminDataCleanupView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/admin/users',
      name: 'admin-users',
      component: () => import('@/views/AdminUsersView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/admin/audit-logs',
      name: 'admin-audit-logs',
      component: () => import('@/views/AdminAuditLogsView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/admin/backups',
      name: 'admin-backups',
      component: () => import('@/views/AdminBackupsView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue')
    }
  ]
})

// Navigation guard for authentication
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
  const requiresAdmin = to.matched.some(record => record.meta.requiresAdmin)

  if (requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (requiresAdmin && authStore.user?.role !== 'admin') {
    // Redirect non-admins trying to access admin routes
    next('/dashboard')
  } else if ((to.name === 'login' || to.name === 'register' || to.name === 'forgot-password' || to.name === 'reset-password' || to.name === 'verify-email') && authStore.isAuthenticated) {
    // If already authenticated, redirect auth flows to dashboard (except verify-email can be accessed)
    if (to.name === 'verify-email') {
      next()
    } else {
      next('/dashboard')
    }
  } else {
    next()
  }
})

export default router
