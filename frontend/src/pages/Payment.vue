<template>
  <div class="max-w-3xl mx-auto px-4 py-8">
    <div v-if="loading" class="text-center py-16">
      <p class="text-gray-500">Loading...</p>
    </div>

    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-500">{{ error }}</p>
      <router-link to="/orders" class="mt-4 inline-block text-blue-600 hover:underline text-sm font-medium">Back to My Orders</router-link>
    </div>

    <div v-else-if="order">
      <div class="bg-white rounded-2xl shadow-sm p-6 md:p-8">
        <ol class="flex items-center w-full" :class="isCancelled ? 'opacity-50' : ''">
          <li
            v-for="(step, index) in steps"
            :key="step.key"
            class="flex items-center"
            :class="index < steps.length - 1 ? 'w-full' : ''"
          >
            <div class="flex flex-col items-center shrink-0">
              <div
                class="w-9 h-9 rounded-full flex items-center justify-center text-sm font-semibold border-2 transition-colors"
                :class="stepClass(index)"
              >
                <svg v-if="stepIcon(index) === 'check'" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
                <svg v-else-if="stepIcon(index) === 'x'" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
                <svg v-else-if="stepIcon(index) === 'refresh'" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
                <span v-else>{{ index + 1 }}</span>
              </div>
              <span
                class="mt-1.5 text-[11px] font-medium"
                :class="index <= currentStep ? 'text-gray-900' : 'text-gray-400'"
              >
                {{ step.label }}
              </span>
            </div>
              <div
                v-if="index < steps.length - 1"
                class="flex-1 h-0.5 mx-3 mb-5 rounded"
                :class="lineClass(index)"
              ></div>
          </li>
        </ol>

        <div v-if="isCancelled" class="mt-6 rounded-xl p-4 text-center border" :class="order.status === 'refunded' ? 'bg-orange-50 border-orange-200' : 'bg-red-50 border-red-200'">
          <p class="text-sm font-semibold" :class="order.status === 'refunded' ? 'text-orange-600' : 'text-red-600'">
            {{ order.status === 'refunded' ? 'This order has been refunded' : 'This order has been cancelled' }}
          </p>
        </div>
      </div>

      <div class="bg-white rounded-2xl shadow-sm p-6 md:p-8 mt-6">
        <div class="flex justify-between items-start gap-4">
          <div class="min-w-0">
            <h1 class="text-xs font-semibold text-gray-400 uppercase tracking-wide">Ref Order</h1>
            <p class="text-lg font-bold text-gray-900 font-mono mt-1">{{ order.order_ref || order.id }}</p>
            <p class="text-xs text-gray-400 mt-1">{{ formatDate(order.created_at) }}</p>
          </div>
          <span
            class="shrink-0 px-2.5 py-1 text-xs font-semibold rounded-full capitalize"
            :class="getStatusClass(order.status)"
          >
            {{ order.status }}
          </span>
        </div>

        <div class="border-t mt-5 pt-5">
          <h2 class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-4">Items</h2>

          <div class="space-y-4">
            <div v-for="item in items" :key="item.id" class="flex justify-between gap-4">
              <div class="min-w-0">
                <p class="text-sm font-medium text-gray-900">{{ item.product_name }}</p>
                <p v-if="item.variant_name" class="text-xs text-gray-500 mt-0.5">{{ item.variant_name }}</p>
                <div v-else-if="item.variant_id" class="text-xs text-gray-400 mt-0.5">
                  <span class="font-mono">{{ item.variant_id }}</span>
                </div>
                <p v-if="item.variant_attributes && Object.keys(item.variant_attributes).length" class="text-xs text-gray-400 mt-0.5">
                  {{ formatAttributes(item.variant_attributes) }}
                </p>
                <p class="text-xs text-gray-500 mt-0.5">
                  {{ item.quantity }} &times; {{ formatPrice(item.product_price) }}
                </p>

                <div v-if="item.flash_sale_id" class="mt-1 flex flex-wrap items-center gap-1.5">
                  <span class="text-[11px] font-semibold text-red-600 bg-red-50 px-2 py-0.5 rounded-full">
                    Flash Sale {{ item.flash_sale_discount_percent }}%
                  </span>
                  <span v-if="item.flash_sale_discount_price" class="text-[11px] text-gray-400 line-through">
                    {{ formatPrice(item.flash_sale_original_price) }}
                  </span>
                  <span v-if="item.flash_sale_quantity" class="text-[11px] text-gray-500">
                    ({{ item.flash_sale_quantity }} at sale price)
                  </span>
                </div>
              </div>

              <div class="text-right shrink-0">
                <p v-if="item.flash_sale_discount_price" class="text-xs text-gray-400 line-through">
                  {{ formatPrice(item.product_price * item.quantity) }}
                </p>
                <p class="text-sm font-bold text-gray-900">{{ formatPrice(item.final_price) }}</p>
              </div>
            </div>
          </div>
        </div>

        <div class="border-t mt-5 pt-4 space-y-2">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500">Subtotal</span>
            <span class="text-gray-700">{{ formatPrice(subtotal) }}</span>
          </div>
          <div v-if="flashDiscount > 0" class="flex justify-between text-sm">
            <span class="text-gray-500">Flash Sale Discount</span>
            <span class="text-green-600">-{{ formatPrice(flashDiscount) }}</span>
          </div>
          <div v-if="promoDiscount > 0" class="flex justify-between text-sm">
            <span class="text-gray-500">
              Voucher
              <span class="font-semibold text-green-700">{{ items[0]?.promo_code }}</span>
            </span>
            <span class="text-green-600">-{{ formatPrice(promoDiscount) }}</span>
          </div>
          <div class="flex justify-between text-base font-bold text-gray-900 pt-2 border-t border-gray-100">
            <span>Total</span>
            <span>{{ formatPrice(order.total_amount) }}</span>
          </div>
        </div>

        <div v-if="order.status === 'pending'" class="border-t mt-5 pt-5">
          <div v-if="!paymentUrl">
            <button
              @click="createPayment"
              :disabled="paymentLoading"
              class="w-full bg-green-600 text-white py-3 rounded-xl font-semibold hover:bg-green-700 active:scale-[0.98] transition-all disabled:bg-gray-300 disabled:text-gray-500 disabled:cursor-not-allowed disabled:active:scale-100"
            >
              {{ paymentLoading ? 'Processing...' : 'Pay Now' }}
            </button>
            <p v-if="paymentError" class="mt-2 text-sm text-red-500">{{ paymentError }}</p>
          </div>

          <div v-else>
            <p class="text-green-600 text-sm mb-4 font-medium">Payment page is ready</p>
            <a
              :href="paymentUrl"
              target="_blank"
              class="block w-full bg-blue-600 text-white py-3 rounded-xl text-center font-semibold hover:bg-blue-700 active:scale-[0.98] transition-all"
            >
              Open Payment Page
            </a>
          </div>
        </div>

        <router-link
          to="/orders"
          class="mt-6 block w-full text-center text-blue-600 hover:underline text-sm"
        >
          Back to My Orders
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { orderApi, paymentApi } from '../api'
import { getStatusClass } from '../utils/format'

const route = useRoute()

const order = ref(null)
const items = ref([])
const loading = ref(true)
const error = ref(null)

const paymentUrl = ref('')
const paymentLoading = ref(false)
const paymentError = ref(null)

const steps = [
  { key: 'placed', label: 'Order Placed' },
  { key: 'paid', label: 'Payment' },
  { key: 'processing', label: 'Processing' },
  { key: 'shipped', label: 'Shipping' },
  { key: 'delivered', label: 'Delivered' }
]

const statusStepMap = {
  pending: 0,
  paid: 1,
  processing: 2,
  shipped: 3,
  delivered: 4
}

const currentStep = computed(() => statusStepMap[order.value?.status] ?? 0)
const isCancelled = computed(() => ['cancelled', 'refunded'].includes(order.value?.status))

const effectiveStep = computed(() => {
  const status = order.value?.status
  if (status === 'refunded') return 2
  if (status === 'cancelled') return order.value?.snap_token ? 2 : 1
  return currentStep.value
})

const stepClass = (index) => {
  if (order.value?.status === 'delivered') return 'bg-green-500 border-green-500 text-white'
  if (isCancelled.value) {
    if (index < effectiveStep.value) return 'bg-green-500 border-green-500 text-white'
    if (index === effectiveStep.value) {
      return order.value?.status === 'refunded'
        ? 'bg-orange-500 border-orange-500 text-white'
        : 'bg-red-500 border-red-500 text-white'
    }
    return 'bg-white border-gray-200 text-gray-400'
  }
  if (index < effectiveStep.value) return 'bg-green-500 border-green-500 text-white'
  if (index === effectiveStep.value) return 'bg-blue-600 border-blue-600 text-white'
  return 'bg-white border-gray-200 text-gray-400'
}

const stepIcon = (index) => {
  const status = order.value?.status
  if (status === 'delivered') return 'check'
  if (isCancelled.value) {
    if (index < effectiveStep.value) return 'check'
    if (index === effectiveStep.value) return status === 'refunded' ? 'refresh' : 'x'
    return 'number'
  }
  return index < effectiveStep.value ? 'check' : 'number'
}

const lineClass = (index) => {
  if (order.value?.status === 'delivered') return 'bg-green-500'
  return index < effectiveStep.value ? 'bg-green-500' : 'bg-gray-200'
}

const formatPrice = (price) => {
  return new Intl.NumberFormat('id-ID').format(price)
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('id-ID', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatAttributes = (attrs) => {
  return Object.values(attrs).join(' • ')
}

const subtotal = computed(() => {
  return items.value.reduce((sum, item) => sum + item.product_price * item.quantity, 0)
})

const flashDiscount = computed(() => {
  return items.value.reduce((sum, item) => sum + (item.product_price * item.quantity - item.final_price), 0)
})

const promoDiscount = computed(() => {
  return items.value.find(item => item.promo_discount_amount)?.promo_discount_amount || 0
})

const createPayment = async () => {
  paymentLoading.value = true
  paymentError.value = null

  try {
    const response = await paymentApi.create(order.value.id, {
      method: 'bank_transfer'
    })
    paymentUrl.value = response.data.redirect_url
  } catch (e) {
    paymentError.value = e.response?.data?.error || e.response?.data?.message || 'Failed to create payment'
  } finally {
    paymentLoading.value = false
  }
}

onMounted(async () => {
  try {
    const response = await orderApi.getById(route.params.orderId)
    order.value = response.data.data
    items.value = response.data.items || []
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to load order'
  } finally {
    loading.value = false
  }
})
</script>
