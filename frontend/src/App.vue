<template>
  <div class="min-h-screen bg-gray-50 flex flex-col">
    <nav v-if="!isAdmin && !isAuthPage" class="sticky top-0 z-40 bg-white/95 backdrop-blur border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16 items-center">
          <div class="flex items-center gap-8">
            <router-link to="/" class="flex items-center gap-2">
              <span class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold text-lg">E</span>
              <span class="text-xl font-bold text-gray-900">E-Commerce</span>
            </router-link>
            <div class="hidden sm:flex items-center gap-1">
              <router-link
                to="/"
                class="px-3 py-2 rounded-lg text-sm font-medium hover:bg-gray-100"
                :class="route.path === '/' ? 'text-blue-600 bg-blue-50' : 'text-gray-600'"
              >
                Home
              </router-link>
            </div>
          </div>
          <div class="flex items-center gap-1">
            <button
              @click="handleCartClick"
              class="relative p-2 rounded-lg hover:bg-gray-100 transition-colors"
              aria-label="Shopping cart"
            >
              <svg class="h-6 w-6 text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
              <span
                v-if="cartCount > 0 && auth.isAuthenticated"
                class="absolute -top-0.5 -right-0.5 bg-red-500 text-white rounded-full text-[11px] font-semibold min-w-[18px] h-[18px] px-1 flex items-center justify-center ring-2 ring-white"
              >
                {{ cartCount > 99 ? '99+' : cartCount }}
              </span>
            </button>

            <router-link
              v-if="!auth.isAuthenticated"
              to="/login"
              class="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
            >
              Login
            </router-link>

            <div v-else class="relative" ref="profileRef">
              <button
                @click="showProfileMenu = !showProfileMenu"
                class="p-2 rounded-lg hover:bg-gray-100 transition-colors"
                :class="showProfileMenu ? 'bg-gray-100' : ''"
                aria-label="Account menu"
              >
                <svg class="h-6 w-6 text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5.121 17.804A13.937 13.937 0 0112 16c2.5 0 4.847.655 6.879 1.804M15 10a3 3 0 11-6 0 3 3 0 016 0zm6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </button>

              <transition
                enter-active-class="transition duration-100 ease-out"
                enter-from-class="scale-95 opacity-0"
                enter-to-class="scale-100 opacity-100"
                leave-active-class="transition duration-75 ease-in"
                leave-from-class="scale-100 opacity-100"
                leave-to-class="scale-95 opacity-0"
              >
                <div
                  v-if="showProfileMenu"
                  class="absolute right-0 mt-2 w-48 bg-white rounded-xl shadow-lg border border-gray-100 py-1.5 z-50"
                >
                  <div class="px-4 py-2 border-b border-gray-100">
                    <p class="text-sm font-medium text-gray-900">{{ auth.user?.email }}</p>
                    <p class="text-xs text-gray-400 capitalize">{{ auth.user?.role }}</p>
                  </div>
                  <router-link
                    to="/orders"
                    @click="showProfileMenu = false"
                    class="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                    :class="route.path === '/orders' ? 'text-blue-600 bg-blue-50' : ''"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                    </svg>
                    My Orders
                  </router-link>
                  <button
                    @click="showProfileMenu = false"
                    class="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    Settings
                  </button>
                  <div class="border-t border-gray-100 mt-1 pt-1">
                    <button
                      @click="handleLogout"
                      class="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-red-600 hover:bg-red-50 transition-colors"
                    >
                      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                      </svg>
                      Logout
                    </button>
                  </div>
                </div>
              </transition>
            </div>
          </div>
        </div>
      </div>
    </nav>

    <main class="flex-1">
      <router-view />
    </main>

    <footer v-if="!isAdmin && !isAuthPage" class="bg-white border-t border-gray-200 mt-12">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <p class="text-sm text-gray-400 text-center">&copy; 2026 E-Commerce. All rights reserved.</p>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { computed, onMounted, onBeforeUnmount, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCartStore } from './stores/cart'
import { useAuthStore } from './stores/auth'

const route = useRoute()
const router = useRouter()
const cart = useCartStore()
const auth = useAuthStore()

const cartCount = computed(() => cart.items.length)
const isAdmin = computed(() => route.path.startsWith('/admin'))
const isAuthPage = computed(() => ['/login', '/register'].includes(route.path))
const showProfileMenu = ref(false)
const profileRef = ref(null)

const closeProfileMenu = (e) => {
  if (profileRef.value && !profileRef.value.contains(e.target)) {
    showProfileMenu.value = false
  }
}

const handleCartClick = () => {
  if (!auth.isAuthenticated) {
    router.push({ path: '/login', query: { redirect: '/cart' } })
    return
  }
  router.push('/cart')
}

const handleLogout = async () => {
  showProfileMenu.value = false
  await auth.logout()
  router.push('/')
}

onMounted(() => {
  cart.loadCart().catch(() => {})
  document.addEventListener('click', closeProfileMenu)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', closeProfileMenu)
})
</script>
