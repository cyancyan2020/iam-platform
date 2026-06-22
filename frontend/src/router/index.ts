import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      component: () => import('@/views/Layout.vue'),
      meta: { requiresAuth: true },
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue'),
        },
        {
          path: 'system/users',
          name: 'Users',
          component: () => import('@/views/system/Users.vue'),
        },
        {
          path: 'system/roles',
          name: 'Roles',
          component: () => import('@/views/system/Roles.vue'),
        },
        {
          path: 'system/permissions',
          name: 'Permissions',
          component: () => import('@/views/system/Permissions.vue'),
        },
        {
          path: 'system/logs',
          name: 'Logs',
          component: () => import('@/views/system/Logs.vue'),
        },
        {
          path: ':pathMatch(.*)*',
          redirect: '/dashboard',
        },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const userStore = useUserStore()
  if (to.meta.requiresAuth !== false && !userStore.token) {
    next('/login')
  } else if (to.path === '/login' && userStore.token) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
