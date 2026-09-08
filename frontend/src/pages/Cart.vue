<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Shopping Cart</h1>

    <div v-if="cart.loading && cart.items.length === 0" class="text-center py-16">
      <p class="text-gray-500">Loading cart...</p>
    </div>

    <div v-else-if="cart.error && cart.items.length === 0" class="text-center py-16">
      <p class="text-red-500">{{ cart.error }}</p>
      <button @click="reloadCart" class="mt-4 text-blue-600 hover:underline text-sm font-medium">Try Again</button>
    </div>

    <div v-else-if="cart.items.length === 0" class="text-center py-16">
      <div class="w-16 h-16 mx-auto bg-gray-100 rounded-full flex items-center justify-center">
        <svg class="w-8 h-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
        </svg>
      </div>
      <p class="mt-4 text-gray-500 font-medium">Your cart is empty</p>
      <p class="mt-1 text-sm text-gray-400">Add items you like to see them here</p>
      <router-link
        to="/"
        class="mt-6 inline-block bg-blue-600 text-white px-6 py-2.5 rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
      >
        Continue Shopping
      </router-link>
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-2">
        <div class="bg-white rounded-xl shadow-sm overflow-hidden">
          <table class="w-full">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-4 py-3 text-left">
                  <input
                    type="checkbox"
                    :checked="cart.allItemsSelected"
                    @change="cart.toggleAll()"
                    aria-label="Select all cart items"
                    class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Product</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Price</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Quantity</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Total</th>
                <th class="px-6 py-3"></th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr v-for="item in cart.items" :key="item.id">
                <td class="px-4 py-4">
                  <input
                    type="checkbox"
                    :checked="cart.selectedItemIds.includes(item.id)"
                    @change="cart.toggleItem(item.id)"
                    :aria-label="`Select ${item.name}`"
                    class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                </td>
                <td class="px-6 py-4">
                  <div class="flex items-center">
                    <div class="ml-4">
                      <p class="font-medium">{{ item.name }}</p>
                      <p v-if="item.variantName" class="text-sm text-gray-500">{{ item.variantName }}</p>
                    </div>
                  </div>
                </td>
                <td class="px-6 py-4">
                  <p v-if="item.discountPercent > 0" class="text-xs text-gray-400 line-through">Rp {{ formatPrice(item.originalPrice) }}</p>
                  <p>Rp {{ formatPrice(item.price) }}</p>
                  <p v-if="item.discountPercent > 0" class="text-xs text-red-500">Flash Sale -{{ item.discountPercent }}%</p>
                </td>
                <td class="px-6 py-4">
                  <div class="flex items-center gap-2">
                    <button
                      @click="queueQuantityUpdate(item, displayedQuantity(item) - 1)"
                      :disabled="isUpdating(item.id) || displayedQuantity(item) <= 1"
                      class="px-2 py-1 border rounded hover:bg-gray-100"
                    >
                      -
                    </button>
                    <span class="w-8 text-center">{{ displayedQuantity(item) }}</span>
                    <button
                      @click="queueQuantityUpdate(item, displayedQuantity(item) + 1)"
                      :disabled="isUpdating(item.id) || displayedQuantity(item) >= item.stock"
                      class="px-2 py-1 border rounded hover:bg-gray-100"
                    >
                      +
                    </button>
                  </div>
                </td>
                <td class="px-6 py-4 font-medium">Rp {{ formatPrice(item.lineTotal) }}</td>
                <td class="px-6 py-4">
                  <button
                    @click="removeItem(item.id)"
                    :disabled="cart.loading"
                    class="text-red-500 hover:text-red-700"
                  >
                    Remove
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="lg:col-span-1">
        <div class="bg-white rounded-xl shadow-sm p-6 lg:sticky lg:top-24">
          <h2 class="text-lg font-semibold mb-4">Order Summary</h2>
          
          <div class="space-y-3">
            <div class="flex justify-between">
              <span>Subtotal</span>
              <span>Rp {{ formatPrice(cart.selectedSubtotal) }}</span>
            </div>

            <div class="flex justify-between text-sm text-gray-500">
              <span>Selected items</span>
              <span>{{ cart.selectedTotalItems }}</span>
            </div>

            <div class="border-t pt-3 flex justify-between font-bold text-lg">
              <span>Total</span>
              <span>Rp {{ formatPrice(cart.selectedSubtotal) }}</span>
            </div>
          </div>

          <router-link
            to="/checkout"
            :class="[
              'mt-6 block w-full py-3 rounded-xl text-center font-semibold transition-colors',
              cart.selectedItems.length > 0
                ? 'bg-blue-600 text-white hover:bg-blue-700 active:scale-[0.98]'
                : 'bg-gray-100 text-gray-400 pointer-events-none'
            ]"
          >
            Checkout ({{ cart.selectedItems.length }})
          </router-link>
        </div>
      </div>
    </div>
    <p v-if="actionError" class="mt-4 text-sm text-red-500">{{ actionError }}</p>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useCartStore } from '../stores/cart'

const cart = useCartStore()

const actionError = ref(null)
const pendingQuantities = ref({})
const updatingItemIds = ref(new Set())
const quantityTimers = new Map()

const formatPrice = (price) => {
  return new Intl.NumberFormat('id-ID').format(price)
}

const displayedQuantity = (item) => pendingQuantities.value[item.id] ?? item.quantity
const isUpdating = (itemId) => updatingItemIds.value.has(itemId)

const queueQuantityUpdate = (item, quantity) => {
  if (quantity < 1 || quantity > item.stock) return

  pendingQuantities.value[item.id] = quantity
  actionError.value = null

  clearTimeout(quantityTimers.get(item.id))
  quantityTimers.set(item.id, setTimeout(async () => {
    updatingItemIds.value = new Set(updatingItemIds.value).add(item.id)
    try {
      await cart.updateQuantity(item.id, pendingQuantities.value[item.id])
    } catch (e) {
      actionError.value = e.response?.data?.error || 'Failed to update quantity'
    } finally {
      delete pendingQuantities.value[item.id]
      quantityTimers.delete(item.id)
      const nextUpdatingItemIds = new Set(updatingItemIds.value)
      nextUpdatingItemIds.delete(item.id)
      updatingItemIds.value = nextUpdatingItemIds
    }
  }, 600))
}

const removeItem = async (itemId) => {
  actionError.value = null
  clearTimeout(quantityTimers.get(itemId))
  quantityTimers.delete(itemId)
  delete pendingQuantities.value[itemId]
  const nextUpdatingItemIds = new Set(updatingItemIds.value)
  nextUpdatingItemIds.delete(itemId)
  updatingItemIds.value = nextUpdatingItemIds
  try {
    await cart.removeItem(itemId)
  } catch (e) {
    actionError.value = e.response?.data?.error || 'Failed to remove item'
  }
}

const reloadCart = () => cart.loadCart(true).catch(() => {})

onMounted(reloadCart)

onBeforeUnmount(() => {
  quantityTimers.forEach(clearTimeout)
  quantityTimers.clear()
})
</script>
