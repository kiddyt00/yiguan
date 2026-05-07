import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/', name: 'Home', component: () => import('../views/Home.vue') },
  { path: '/login', name: 'Login', component: () => import('../views/Login.vue') },
  { path: '/profile', name: 'Profile', component: () => import('../views/Profile.vue'), meta: { auth: true } },
  { path: '/history', name: 'History', component: () => import('../views/History.vue'), meta: { auth: true } },
  { path: '/stream', name: 'StreamDivine', component: () => import('../views/StreamDivine.vue'), meta: { auth: true } },
  { path: '/ads', name: 'AdCenter', component: () => import('../views/AdCenter.vue'), meta: { auth: true } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.auth && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
