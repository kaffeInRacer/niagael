<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Products</h1>
      <p class="text-gray-500 mt-1">Temukan produk terbaik untuk kebutuhanmu</p>
    </div>

    <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
      <div v-for="n in 8" :key="n" class="bg-white rounded-xl overflow-hidden shadow-sm">
        <div class="aspect-square bg-gray-200 animate-pulse"></div>
        <div class="p-4 space-y-2">
          <div class="h-4 bg-gray-200 rounded animate-pulse w-3/4"></div>
          <div class="h-3 bg-gray-200 rounded animate-pulse w-1/3"></div>
          <div class="h-5 bg-gray-200 rounded animate-pulse w-1/2"></div>
        </div>
      </div>
    </div>

    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-500">{{ error }}</p>
      <button @click="fetchProducts" class="mt-4 text-blue-600 hover:underline text-sm font-medium">Try Again</button>
    </div>

    <div v-else-if="products.length === 0" class="text-center py-16">
      <p class="text-gray-500">No products available</p>
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
      <router-link
        v-for="product in products"
        :key="product.id"
        :to="`/product/${product.id}`"
        class="group bg-white rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all duration-200 hover:-translate-y-0.5"
      >
        <div class="relative aspect-square bg-gray-100 overflow-hidden">
          <img
            v-if="product.image"
            :src="productImageUrl(product.image)"
            :alt="product.name"
            class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
          />
          <div v-else class="w-full h-full flex items-center justify-center text-gray-300">
            <svg class="w-12 h-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
          <span
            v-if="product.discount_percent > 0"
            class="absolute top-3 left-3 bg-red-500 text-white text-xs font-bold px-2 py-1 rounded-md shadow-sm"
          >
            -{{ product.discount_percent }}%
          </span>
        </div>

        <div class="p-4">
          <p class="text-[11px] font-medium text-gray-400 uppercase tracking-wide">{{ product.category }}</p>
          <h2 class="mt-1 text-sm font-semibold text-gray-900 line-clamp-2 group-hover:text-blue-600 transition-colors">{{ product.name }}</h2>
          <div class="mt-2 flex items-baseline gap-2">
            <p
              class="text-base font-bold"
              :class="product.discount > 0 ? 'text-red-500' : 'text-gray-900'"
            >
              Rp {{ formatPrice(product.final_price) }}
            </p>
            <p v-if="product.discount > 0" class="text-xs text-gray-400 line-through">
              Rp {{ formatPrice(product.price) }}
            </p>
          </div>
          <p v-if="product.discount > 0" class="mt-1.5">
            <span class="text-[11px] font-semibold text-red-600 bg-red-50 px-2 py-0.5 rounded-full">Flash Sale</span>
          </p>
          <p v-if="product.has_variants" class="mt-1 text-xs text-gray-400">Multiple options available</p>
        </div>
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { productApi } from '../api'
import { productImageUrl } from '../config'

const products = ref([])
const loading = ref(true)
const error = ref(null)

const formatPrice = (price) => {
  return new Intl.NumberFormat('id-ID').format(price)
}

const fetchProducts = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await productApi.list()
    products.value = response.data.data || []
  } catch (e) {
    error.value = e.message || 'Failed to load products'
  } finally {
    loading.value = false
  }
}

onMounted(fetchProducts)
</script>
