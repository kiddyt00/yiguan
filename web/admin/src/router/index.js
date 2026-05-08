import { createRouter, createWebHistory } from 'vue-router'
import { token } from '../stores/auth'
import AdminLayout from '../layout/AdminLayout.vue'
import Login from '../views/Login.vue'
import Dashboard from '../views/Dashboard.vue'
import Users from '../views/Users.vue'

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    { path: '/login', component: Login },
    {
      path: '/',
      component: AdminLayout,
      meta: { requiresAuth: true },
      children: [
        { path: '', component: Dashboard },
        { path: 'users', component: Users },
        { path: 'hexagrams', component: () => import('../views/Hexagrams.vue') },
        { path: 'models', component: () => import('../views/Models.vue') },
        { path: 'ads', component: () => import('../views/Ads.vue') },
      ],
    },
  ],
})

router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !token.value) {
    next('/login')
  } else if (to.path === '/login' && token.value) {
    next('/')
  } else {
    next()
  }
})

export default router
