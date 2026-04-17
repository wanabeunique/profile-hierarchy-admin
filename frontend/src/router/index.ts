import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/LoginPage.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      name: 'dashboard',
      redirect: '/profiles/type/plus',
      meta: { requiresAuth: true },
    },
    {
      path: '/profiles/type/:type',
      name: 'profile-list',
      component: () => import('@/pages/ProfileListPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/profiles/create',
      name: 'create-profile',
      component: () => import('@/pages/CreateProfilePage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/profiles/:id',
      name: 'profile-detail',
      component: () => import('@/pages/ProfileDetailPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/profiles/:id/orders/create',
      name: 'create-order',
      component: () => import('@/pages/CreateOrderPage.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { path: '/login' }
  }
  if (to.path === '/login' && auth.isAuthenticated) {
    return { path: '/' }
  }
  // Non-admin users can't access profile list pages — redirect to own profile
  if (to.name === 'profile-list' && auth.isAuthenticated && !auth.isAdmin) {
    return { path: `/profiles/${auth.profileId}` }
  }
})

export default router
