<template>
  <ModalDialog :show="show" title="Select Product" aria-id="product-picker-modal" size="lg" scroll-body @close="$emit('close')">
    <div class="space-y-4">
      <div class="relative">
        <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <svg class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search products by name..."
          class="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
          @input="onSearchInput"
        />
      </div>

      <div v-if="loading" class="flex justify-center py-8">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>

      <div v-else-if="products.length > 0" class="space-y-2 max-h-96 overflow-y-auto">
        <button
          v-for="product in products"
          :key="product.id"
          type="button"
          @click="selectProduct(product)"
          class="w-full flex items-center gap-3 p-3 border border-gray-200 rounded-lg hover:bg-blue-50 hover:border-blue-300 transition-colors text-left"
        >
          <div class="flex-shrink-0">
            <img
              v-if="product.image"
              :src="getProductImageUrl(product)"
              :alt="product.name"
              class="w-12 h-12 rounded-lg object-cover"
            />
            <div v-else class="w-12 h-12 rounded-lg bg-gray-100 flex items-center justify-center">
              <svg class="w-6 h-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
              </svg>
            </div>
          </div>

          <div class="flex-1 min-w-0">
            <p class="font-medium text-gray-800 truncate">{{ product.name }}</p>
            <p class="text-sm text-gray-500">
              <span v-if="product.category">{{ product.category }}</span>
              <span v-if="product.has_variants" class="ml-2 px-1.5 py-0.5 bg-purple-100 text-purple-700 rounded text-xs">Has Variants</span>
            </p>
          </div>

          <div v-if="product.is_flash_sale" class="flex-shrink-0">
            <span class="px-2 py-1 bg-red-100 text-red-700 rounded text-xs font-medium">On Flash Sale</span>
          </div>
        </button>
      </div>

      <div v-else class="text-center py-8">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
        </svg>
        <p class="mt-2 text-sm text-gray-500">No products found</p>
      </div>

      <div v-if="totalCount > pageSize" class="flex items-center justify-between pt-4 border-t">
        <p class="text-sm text-gray-500">
          Showing {{ (currentPage - 1) * pageSize + 1 }} to {{ Math.min(currentPage * pageSize, totalCount) }} of {{ totalCount }} products
        </p>
        <div class="flex gap-2">
          <button
            type="button"
            :disabled="currentPage <= 1"
            @click="prevPage"
            class="px-3 py-1 text-sm border rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
          >
            Previous
          </button>
          <button
            type="button"
            :disabled="currentPage >= totalPages"
            @click="nextPage"
            class="px-3 py-1 text-sm border rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
          >
            Next
          </button>
        </div>
      </div>
    </div>
  </ModalDialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { adminApi } from '../api'
import ModalDialog from './ModalDialog.vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  excludeIds: { type: Array, default: () => [] }
})

const emit = defineEmits(['close', 'select'])

const searchQuery = ref('')
const products = ref([])
const loading = ref(false)
const currentPage = ref(1)
const totalCount = ref(0)
const pageSize = 20

const totalPages = computed(() => Math.ceil(totalCount.value / pageSize))

let searchTimeout = null

const onSearchInput = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchProducts()
  }, 300)
}

const fetchProducts = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      page_size: pageSize,
      order_by: 'name',
      order_dir: 'asc'
    }
    if (searchQuery.value.trim()) {
      params.search = searchQuery.value.trim()
    }
    const res = await adminApi.products.list(params)
    const allProducts = res.data.data || []
    products.value = allProducts.filter(p => !props.excludeIds.includes(p.id))
    totalCount.value = res.data.count || 0
  } catch (e) {
    console.error('Failed to fetch products', e)
    products.value = []
    totalCount.value = 0
  } finally {
    loading.value = false
  }
}

const getProductImageUrl = (product) => {
  const base = import.meta.env.VITE_IMG_URL || ''
  return `${base}/${product.id}/${product.image.file_name}`
}

const selectProduct = (product) => {
  emit('select', product)
  emit('close')
}

const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    fetchProducts()
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    fetchProducts()
  }
}

watch(() => props.show, (newVal) => {
  if (newVal) {
    searchQuery.value = ''
    currentPage.value = 1
    fetchProducts()
  }
})
</script>
