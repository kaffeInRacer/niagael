import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/',
    name: 'Products',
    component: () => import('../pages/Products.vue')
  },
  {
    path: '/product/:id',
    name: 'ProductDetail',
    component: () => import('../pages/ProductDetail.vue')
  },
  {
    path: '/cart',
    name: 'Cart',
    component: () => import('../pages/Cart.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/orders',
    name: 'OrderHistory',
    component: () => import('../pages/OrderHistory.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/checkout',
    name: 'Checkout',
    component: () => import('../pages/Checkout.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/payment/:orderId',
    name: 'Payment',
    component: () => import('../pages/Payment.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../pages/Login.vue'),
    meta: { guestOnly: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../pages/Register.vue'),
    meta: { guestOnly: true }
  },
  {
    path: '/admin',
    component: () => import('../layouts/AdminLayout.vue'),
    meta: { requiresAuth: true, roles: ['staff', 'admin'] },
    children: [
      {
        path: '',
        name: 'AdminDashboard',
        component: () => import('../pages/admin/Dashboard.vue')
      },
      {
        path: 'categories',
        name: 'AdminCategories',
        component: () => import('../pages/admin/Categories.vue')
      },
      {
        path: 'products',
        name: 'AdminProducts',
        component: () => import('../pages/admin/Products.vue')
      },
      {
        path: 'flash-sales',
        name: 'AdminFlashSales',
        component: () => import('../pages/admin/FlashSales.vue')
      },
      {
        path: 'promos',
        name: 'AdminPromos',
        component: () => import('../pages/admin/Promos.vue')
      },
      {
        path: 'orders',
        name: 'AdminOrders',
        component: () => import('../pages/admin/Orders.vue')
      },
      {
        path: 'roles',
        name: 'AdminRoles',
        component: () => import('../pages/admin/Roles.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.initialize()

  if (to.meta.guestOnly && auth.isAuthenticated) {
    return auth.canAccessAdmin ? '/admin' : '/'
  }

  if (to.matched.some(record => record.meta.requiresAuth) && !auth.isAuthenticated) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  const roles = to.matched.flatMap(record => record.meta.roles || [])
  if (roles.length > 0 && !roles.some(role => auth.roles.includes(role))) {
    return '/'
  }
})

export default router
