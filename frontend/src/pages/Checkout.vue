<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Checkout</h1>

    <div v-if="cart.loading && cart.items.length === 0" class="text-center py-16 text-gray-500">Loading cart...</div>
    <div v-else-if="cart.items.length === 0" class="text-center py-16">
      <p class="text-gray-500 font-medium">Your cart is empty</p>
      <router-link
        to="/"
        class="mt-6 inline-block bg-blue-600 text-white px-6 py-2.5 rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
      >
        Continue Shopping
      </router-link>
    </div>
    <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="bg-white rounded-xl shadow-sm p-6">
        <h2 class="text-lg font-semibold mb-4">Shipping Address</h2>
        
        <form @submit.prevent="submitOrder" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">Address</label>
            <textarea
              v-model="address"
              rows="3"
              required
              class="mt-1 block w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter your address"
            ></textarea>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700">City</label>
              <input
                v-model="city"
                type="text"
                required
                class="mt-1 block w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">Postal Code</label>
              <input
                v-model="postalCode"
                type="text"
                required
                class="mt-1 block w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>

          <div>
			<label class="block text-sm font-medium text-gray-700">Province</label>
			<input
			  v-model="province"
			  type="text"
			  required
			  class="mt-1 block w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
			/>
          </div>

		  <div>
			<label class="block text-sm font-medium text-gray-700">Country</label>
			<input
			  v-model="country"
			  type="text"
			  required
			  class="mt-1 block w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
			/>
		  </div>
        </form>
      </div>

      <div class="bg-white rounded-xl shadow-sm p-6">
        <h2 class="text-lg font-semibold mb-4">Order Summary</h2>

        <div class="space-y-4">
          <div v-for="item in cart.selectedItems" :key="item.id" class="flex justify-between">
            <div>
              <p class="font-medium">{{ item.name }}</p>
              <p v-if="item.variantName" class="text-sm text-gray-500">{{ item.variantName }}</p>
              <p class="text-sm text-gray-500">Qty: {{ item.quantity }}</p>
            </div>
            <span>Rp {{ formatPrice(item.lineTotal) }}</span>
          </div>
        </div>

        <div class="mt-4">
          <label class="text-sm font-medium text-gray-700">Voucher Code</label>
          <div class="mt-2 flex gap-2">
            <input
              v-model="promoCode"
              type="text"
              placeholder="Enter promo code"
              class="flex-1 px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              @click="applyPromo"
              :disabled="!promoCode || promoApplying"
              class="px-4 py-2 bg-gray-200 rounded-lg hover:bg-gray-300 disabled:bg-gray-100"
            >
              {{ promoApplying ? 'Checking...' : 'Apply' }}
            </button>
          </div>
          <p v-if="promoError" class="mt-1 text-sm text-red-500">{{ promoError }}</p>
          <p v-if="cart.promoDiscount > 0" class="mt-1 text-sm text-green-600">Voucher {{ cart.promoCode }} applied</p>
        </div>

        <div class="border-t mt-4 pt-4 space-y-2">
          <div class="flex justify-between">
            <span>Subtotal</span>
            <span>Rp {{ formatPrice(cart.selectedSubtotal) }}</span>
          </div>
          <div v-if="cart.promoDiscount > 0" class="flex justify-between text-green-600">
            <span>Voucher ({{ cart.promoCode }})</span>
            <span>- Rp {{ formatPrice(cart.promoDiscount) }}</span>
          </div>
          <div class="flex justify-between font-bold text-lg">
            <span>Total</span>
            <span>Rp {{ formatPrice(cart.selectedTotal) }}</span>
          </div>
        </div>

        <button
          @click="submitOrder"
          :disabled="loading"
          class="mt-6 w-full bg-blue-600 text-white py-3 rounded-xl font-semibold hover:bg-blue-700 active:scale-[0.98] transition-all disabled:bg-gray-200 disabled:text-gray-400 disabled:cursor-not-allowed disabled:active:scale-100"
        >
          {{ loading ? 'Processing...' : 'Place Order' }}
        </button>

        <p v-if="error" class="mt-2 text-sm text-red-500">{{ error }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '../stores/cart'
import { addressApi, orderApi, promoApi } from '../api'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const cart = useCartStore()
const auth = useAuthStore()

const address = ref('')
const city = ref('')
const postalCode = ref('')
const province = ref('')
const country = ref('Indonesia')
const loading = ref(false)
const error = ref(null)
const promoCode = ref('')
const promoApplying = ref(false)
const promoError = ref(null)

const formatPrice = (price) => {
  return new Intl.NumberFormat('id-ID').format(price)
}

const applyPromo = async () => {
  promoError.value = null
  promoApplying.value = true
  try {
    const response = await promoApi.apply(promoCode.value, auth.userId, cart.selectedSubtotal)
    cart.setPromo(promoCode.value, response.data.discount)
  } catch (e) {
    promoError.value = e.response?.data?.error || e.response?.data?.message || 'Failed to apply promo'
  } finally {
    promoApplying.value = false
  }
}

const submitOrder = async () => {
  loading.value = true
  error.value = null
  const idempotencyKey = orderApi.newIdempotencyKey()

  try {
	const addressResponse = await addressApi.create({
	  user_id: auth.userId,
	  street: address.value,
	  city: city.value,
	  province: province.value,
	  postal_code: postalCode.value,
	  country: country.value,
	  is_default: false
	})

    const orderData = {
	  user_id: auth.userId,
	  address_id: addressResponse.data.data.id,
      items: cart.selectedItems.map(item => ({
        product_id: item.productId,
        variant_id: item.variantId,
        quantity: item.quantity
      })),
      promo_code: cart.promoCode || undefined
    }

    const response = await orderApi.create(orderData, idempotencyKey)
    const orderId = response.data.data.id

    try {
      await cart.removeSelectedItems()
    } catch {
    }
    router.push(`/payment/${orderId}`)
  } catch (e) {
	error.value = e.response?.data?.error || e.response?.data?.message || 'Failed to create order'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  cart.loadCart(true).catch(() => {})
})
</script>
