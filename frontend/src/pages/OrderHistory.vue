<template>
  <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">My Orders</h1>

    <div v-if="loading" class="space-y-4">
      <div v-for="n in 3" :key="n" class="bg-white rounded-xl shadow-sm p-5">
        <div class="flex justify-between">
          <div class="space-y-2">
            <div class="h-4 bg-gray-200 rounded animate-pulse w-40"></div>
            <div class="h-3 bg-gray-200 rounded animate-pulse w-24"></div>
          </div>
          <div class="h-6 bg-gray-200 rounded-full animate-pulse w-20"></div>
        </div>
      </div>
    </div>

    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-500">{{ error }}</p>
      <button @click="fetchOrders" class="mt-4 text-blue-600 hover:underline text-sm font-medium">Try Again</button>
    </div>

    <div v-else-if="orders.length === 0" class="text-center py-16">
      <div class="w-16 h-16 mx-auto bg-gray-100 rounded-full flex items-center justify-center">
        <svg class="w-8 h-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 12h.01M9 16h.01" />
        </svg>
      </div>
      <p class="mt-4 text-gray-500 font-medium">No orders yet</p>
      <p class="mt-1 text-sm text-gray-400">Your orders will appear here after checkout</p>
      <router-link
        to="/"
        class="mt-6 inline-block bg-blue-600 text-white px-6 py-2.5 rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
      >
        Start Shopping
      </router-link>
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="order in orders"
        :key="order.id"
        class="bg-white rounded-xl shadow-sm hover:shadow-md transition-shadow overflow-hidden"
      >
        <div class="p-5">
          <div class="flex justify-between items-start gap-4">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <h2 class="text-sm font-semibold text-gray-900 font-mono">{{ order.order_ref }}</h2>
              </div>
              <p class="text-xs text-gray-400 mt-1">{{ formatDate(order.created_at) }}</p>
            </div>
            <span
              class="shrink-0 px-2.5 py-1 text-xs font-semibold rounded-full capitalize"
              :class="getStatusClass(order.status)"
            >
              {{ order.status }}
            </span>
          </div>

          <div class="mt-4 flex justify-between items-center">
            <div>
              <p class="text-xs text-gray-400">Total</p>
              <p class="text-lg font-bold text-gray-900">{{ formatPrice(order.total_amount) }}</p>
            </div>
            <div class="flex gap-2">
              <button
                v-if="order.status === 'pending'"
                @click="payOrder(order.id)"
                :disabled="payingOrderId === order.id"
                class="px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 active:scale-[0.98] transition-all disabled:bg-gray-300 disabled:text-gray-500 disabled:cursor-not-allowed disabled:active:scale-100"
              >
                {{ payingOrderId === order.id ? 'Processing...' : 'Pay Now' }}
              </button>
              <button
                @click="viewDetail(order.id)"
                class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
              >
                View Detail
              </button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="nextCursor" class="text-center pt-2">
        <button
          @click="loadMore"
          :disabled="loadingMore"
          class="px-6 py-2.5 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors disabled:bg-gray-100 disabled:text-gray-400 disabled:cursor-not-allowed"
        >
          {{ loadingMore ? 'Loading...' : 'Load More' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { orderApi } from '../api'
import { formatDate, formatPrice, getStatusClass } from '../utils/format'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const PAGE_SIZE = 10

const orders = ref([])
const loading = ref(true)
const loadingMore = ref(false)
const error = ref(null)
const nextCursor = ref('')
const payingOrderId = ref(null)

const fetchOrders = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await orderApi.listByUser(auth.userId, { page_size: PAGE_SIZE })
    orders.value = response.data.data || []
    nextCursor.value = response.data.next_cursor || ''
  } catch (e) {
    error.value = e.response?.data?.error || 'Failed to load orders'
  } finally {
    loading.value = false
  }
}

const loadMore = async () => {
  if (!nextCursor.value || loadingMore.value) return
  loadingMore.value = true
  try {
    const response = await orderApi.listByUser(auth.userId, {
      page_size: PAGE_SIZE,
      before: nextCursor.value
    })
    const more = response.data.data || []
    const existing = new Set(orders.value.map(o => o.id))
    orders.value = [...orders.value, ...more.filter(o => !existing.has(o.id))]
    nextCursor.value = response.data.next_cursor || ''
  } catch (e) {
    error.value = e.response?.data?.error || 'Failed to load more orders'
  } finally {
    loadingMore.value = false
  }
}

const payOrder = (orderId) => {
  router.push(`/payment/${orderId}`)
}

const viewDetail = (orderId) => {
  router.push(`/payment/${orderId}`)
}

onMounted(fetchOrders)
</script>
