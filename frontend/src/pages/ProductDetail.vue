<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 gap-8">
      <div class="aspect-square bg-gray-200 rounded-2xl animate-pulse"></div>
      <div class="space-y-4">
        <div class="h-8 bg-gray-200 rounded animate-pulse w-3/4"></div>
        <div class="h-4 bg-gray-200 rounded animate-pulse w-1/4"></div>
        <div class="h-10 bg-gray-200 rounded animate-pulse w-1/2"></div>
        <div class="h-12 bg-gray-200 rounded animate-pulse"></div>
        <div class="h-12 bg-gray-200 rounded animate-pulse"></div>
      </div>
    </div>

    <div v-else-if="error" class="text-center py-16">
      <p class="text-red-500">{{ error }}</p>
      <router-link to="/" class="mt-4 inline-block text-blue-600 hover:underline text-sm font-medium">Back to Products</router-link>
    </div>

    <div v-else-if="product">
      <nav class="mb-6 flex items-center gap-2 text-sm text-gray-400">
        <router-link to="/" class="hover:text-blue-600 transition-colors">Home</router-link>
        <span>/</span>
        <span class="text-gray-600">{{ product.category?.name }}</span>
      </nav>

      <div class="bg-white rounded-2xl shadow-sm p-6 md:p-8">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div>
            <div class="aspect-square bg-gray-100 rounded-2xl overflow-hidden">
              <img v-if="product.images && product.images[0]" :src="productImageUrl(product.images[0])" :alt="product.name" class="w-full h-full object-cover" />
              <div v-else class="w-full h-full flex items-center justify-center text-gray-300">
                <svg class="w-16 h-16" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>
            </div>
          </div>

          <div class="flex flex-col">
            <h1 class="text-2xl md:text-3xl font-bold text-gray-900">{{ product.name }}</h1>
            <p class="text-sm text-gray-400 mt-2 uppercase tracking-wide">{{ product.category?.name }}</p>

            <div class="mt-4 flex items-baseline gap-3">
              <p class="text-3xl font-bold" :class="selectedPrice !== selectedFinalPrice ? 'text-red-500' : 'text-gray-900'">Rp {{ formatPrice(selectedFinalPrice) }}</p>
              <p v-if="selectedPrice !== selectedFinalPrice" class="text-lg text-gray-400 line-through">Rp {{ formatPrice(selectedPrice) }}</p>
              <span v-if="selectedDiscountPercent > 0" class="bg-red-500 text-white text-xs font-bold px-2 py-1 rounded-md">
                -{{ selectedDiscountPercent }}%
              </span>
            </div>
            <p v-if="selectedDiscountPercent > 0" class="mt-1.5">
              <span class="text-[11px] font-semibold text-red-600 bg-red-50 px-2 py-0.5 rounded-full">Flash Sale</span>
            </p>

            <div v-if="variantType !== 'none'" class="mt-6 space-y-4">
              <div v-for="attr in variantAttributes" :key="attr">
                <p class="text-sm font-medium text-gray-700">{{ getLabel(attr) }}</p>
                <div class="mt-2 flex flex-wrap gap-2">
                  <button
                    v-for="option in getOptionsForAttribute(attr)"
                    :key="option.value"
                    @click="!option.disabled && selectOption(attr, option.value)"
                    :disabled="option.disabled"
                    :class="[
                      'px-4 py-2 rounded-lg border text-sm font-medium transition-colors',
                      option.disabled
                        ? 'border-gray-200 bg-gray-100 text-gray-400 cursor-not-allowed'
                        : selectedOptions[attr] === option.value
                          ? 'border-blue-600 bg-blue-50 text-blue-600 ring-1 ring-blue-600'
                          : 'border-gray-300 hover:border-blue-400 text-gray-700'
                    ]"
                  >
                    {{ option.value }}
                  </button>
                </div>
              </div>
            </div>

            <div v-if="selectedStock > 0" class="mt-6">
              <label class="text-sm font-medium text-gray-700 mr-5">Qty</label>
              <div class="mt-2 inline-flex items-center border rounded-lg overflow-hidden">
                <button
                  @click="quantity > 1 && quantity--"
                  class="px-3 py-2 hover:bg-gray-100 transition-colors disabled:opacity-40"
                  :disabled="quantity <= 1"
                >
                  &minus;
                </button>
                <span class="w-12 text-center text-sm font-medium">{{ quantity }}</span>
                <button
                  @click="quantity < selectedStock && quantity++"
                  class="px-3 py-2 hover:bg-gray-100 transition-colors disabled:opacity-40"
                  :disabled="quantity >= selectedStock"
                >
                  +
                </button>
              </div>
              <p class="mt-2 ml-10 text-sm" :class="quantity >= selectedStock ? 'text-orange-500' : 'text-green-600'">
                <span class="inline-flex items-center gap-1.5">
                  <span class="w-2 h-2 rounded-full" :class="quantity >= selectedStock ? 'bg-orange-500' : 'bg-green-500'"></span>
                  {{ quantity >= selectedStock ? 'Mencapai batas qty' : `In Stock (${selectedStock})` }}
                </span>
              </p>
            </div>

            <div class="mt-6">
              <button
                @click="handleAddToCart"
                :disabled="!canAddToCart || addingToCart"
                class="w-full sm:w-auto sm:min-w-[220px] bg-blue-600 text-white px-8 py-3 rounded-xl font-semibold hover:bg-blue-700 active:scale-[0.98] transition-all disabled:bg-gray-300 disabled:text-gray-500 disabled:cursor-not-allowed disabled:active:scale-100"
              >
                {{ addingToCart ? 'Adding...' : addedToCart ? 'Added to Cart ✓' : getButtonText() }}
              </button>
              <p v-if="cartError" class="mt-2 text-sm text-red-500">{{ cartError }}</p>
            </div>

            <div v-if="product.description" class="mt-8 pt-6 border-t border-gray-100">
              <h2 class="text-sm font-semibold text-gray-900 uppercase tracking-wide">Description</h2>
              <p class="mt-2 text-gray-600 leading-relaxed">{{ product.description }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { productApi } from '../api'
import { useCartStore } from '../stores/cart'
import { useAuthStore } from '../stores/auth'
import { productImageUrl } from '../config'

const route = useRoute()
const router = useRouter()
const cart = useCartStore()
const auth = useAuthStore()

const product = ref(null)
const variantData = ref(null)
const selectedOptions = ref({})
const selectedVariant = ref(null)
const quantity = ref(1)
const loading = ref(true)
const error = ref(null)
const addingToCart = ref(false)
const addedToCart = ref(false)
const cartError = ref(null)

const labelMap = {
  color: 'Warna',
  size: 'Ukuran',
  material: 'Bahan',
  storage: 'Kapasitas',
  ram: 'RAM',
  chip: 'Prosesor',
  model: 'Model',
  type: 'Tipe'
}

const formatPrice = (price) => {
  return new Intl.NumberFormat('id-ID').format(price)
}

const variantType = computed(() => {
  if (!variantData.value) return 'none'
  const attrs = variantData.value.attributes
  if (!attrs || attrs.length === 0) return 'none'
  if (attrs.length === 1) return 'single'
  return 'matrix'
})

const variantAttributes = computed(() => {
  return variantData.value?.attributes || []
})

const getOptionsForAttribute = (attr) => {
  if (!variantData.value?.items) return []
  
  const attrIndex = variantAttributes.value.indexOf(attr)
  
  const allOptions = [...new Set(variantData.value.items.map(item => item.options[attr]).filter(Boolean))]
  
  const previousAttrs = variantAttributes.value.slice(0, attrIndex)
  const previousSelections = Object.entries(selectedOptions.value).filter(([key]) => previousAttrs.includes(key))
  
  return allOptions.map(option => {
    const matchingVariants = variantData.value.items.filter(item => {
      const matchesThis = item.options[attr] === option
      const matchesPrevious = previousSelections.every(([key, value]) => item.options[key] === value)
      return matchesThis && matchesPrevious
    })
    
    const hasStock = matchingVariants.some(v => v.stock > 0)
    
    return {
      value: option,
      disabled: !hasStock
    }
  })
}

const getLabel = (attr) => {
  return labelMap[attr] || attr.charAt(0).toUpperCase() + attr.slice(1)
}

const selectOption = (attr, option) => {
  const attrIndex = variantAttributes.value.indexOf(attr)
  const newOptions = { ...selectedOptions.value }
  
  newOptions[attr] = option
  
  variantAttributes.value.slice(attrIndex + 1).forEach(downstreamAttr => {
    delete newOptions[downstreamAttr]
    
    selectedOptions.value = { ...newOptions }
    const downstreamOptions = getOptionsForAttribute(downstreamAttr)
    const firstAvailable = downstreamOptions.find(o => !o.disabled)
    if (firstAvailable) {
      newOptions[downstreamAttr] = firstAvailable.value
    }
  })
  
  selectedOptions.value = newOptions
  quantity.value = 1
  findSelectedVariant()
}

const findSelectedVariant = () => {
  if (!variantData.value?.items) {
    selectedVariant.value = null
    return
  }

  const allOptionsSelected = variantAttributes.value.every(attr => selectedOptions.value[attr])
  if (!allOptionsSelected) {
    selectedVariant.value = null
    return
  }

  selectedVariant.value = variantData.value.items.find(item => {
    return variantAttributes.value.every(attr => item.options[attr] === selectedOptions.value[attr])
  })
}

const selectedPrice = computed(() => {
  if (selectedVariant.value) {
    return selectedVariant.value.price ?? selectedVariant.value.final_price ?? 0
  }
  return product.value?.price ?? 0
})

const selectedFinalPrice = computed(() => {
  return selectedVariant.value?.final_price ?? product.value?.final_price ?? 0
})

const selectedDiscountPercent = computed(() => {
  return selectedVariant.value?.discount_percent ?? product.value?.discount_percent ?? 0
})

const selectedStock = computed(() => {
  if (variantType.value === 'none') return product.value?.stock ?? 0
  return selectedVariant.value?.stock ?? 0
})

const canAddToCart = computed(() => {
  if (variantType.value === 'none') return product.value && product.value.stock > 0
  return selectedVariant.value && selectedStock.value > 0
})

const getButtonText = () => {
  if (variantType.value !== 'none' && !selectedVariant.value) return 'Pilih Varian'
  if (selectedStock.value === 0) return 'Out of Stock'
  return 'Add to Cart'
}

const handleAddToCart = () => {
  if (!auth.isAuthenticated) {
    router.push({ path: '/login', query: { redirect: route.fullPath } })
    return
  }
  
  addToCart()
}

const addToCart = async () => {
  if (canAddToCart.value && !addingToCart.value) {
    addingToCart.value = true
    cartError.value = null
    try {
      await cart.addItem(product.value, selectedVariant.value, quantity.value)
      addedToCart.value = true
      setTimeout(() => {
        addedToCart.value = false
      }, 2000)
    } catch (e) {
      cartError.value = e.response?.data?.error || 'Failed to add item to cart'
    } finally {
      addingToCart.value = false
    }
  }
}

onMounted(async () => {
  try {
    const response = await productApi.getById(route.params.id)
    const data = response.data.data
    product.value = data
    variantData.value = data.variants || null
    
    if (variantData.value?.attributes) {
      variantData.value.attributes.forEach(attr => {
        const options = getOptionsForAttribute(attr)
        const firstAvailable = options.find(o => !o.disabled)
        if (firstAvailable) {
          selectedOptions.value[attr] = firstAvailable.value
        }
      })
      findSelectedVariant()
    }
  } catch (e) {
    error.value = e.message || 'Failed to load product'
  } finally {
    loading.value = false
  }
})
</script>
