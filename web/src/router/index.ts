import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
    },
    {
      path: '/',
      name: 'Dashboard',
      component: () => import('../views/Dashboard.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/app/:id',
      name: 'AppDetail',
      component: () => import('../views/AppDetail.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/upload',
      name: 'Upload',
      component: () => import('../views/Upload.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/d/:id',
      name: 'Download',
      component: () => import('../views/Download.vue'),
    },
  ],
})

router.beforeEach((to, _from, next) => {
  if (to.meta.requiresAuth && !localStorage.getItem('token')) {
    next('/login')
  } else {
    next()
  }
})

export default router
