import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { cartApi } from '../api'
import { useAuthStore } from './auth'

export const useCartStore = defineStore('cart', () => {
  const auth = useAuthStore()
  const id = ref('')
  const items = ref([])
  const backendSubtotal = ref(0)
  const backendTotalItems = ref(0)
  const promoCode = ref('')
  const promoDiscount = ref(0)
  const selectedItemIds = ref([])
  const loading = ref(false)
  const error = ref(null)
  let loaded = false
  let selectionInitialized = false

  const currentUserId = () => {
    if (!auth.userId) throw new Error('Authentication required')
    return auth.userId
  }

  const totalItems = computed(() => backendTotalItems.value)
  const subtotal = computed(() => backendSubtotal.value)
  const total = computed(() => Math.max(0, subtotal.value - promoDiscount.value))
  const selectedItems = computed(() => items.value.filter(item => selectedItemIds.value.includes(item.id)))
  const selectedTotalItems = computed(() => selectedItems.value.reduce((sum, item) => sum + item.quantity, 0))
  const selectedSubtotal = computed(() => selectedItems.value.reduce((sum, item) => sum + item.lineTotal, 0))
  const selectedTotal = computed(() => Math.max(0, selectedSubtotal.value - promoDiscount.value))
  const allItemsSelected = computed(() => items.value.length > 0 && selectedItemIds.value.length === items.value.length)

  function applyCart(data) {
    id.value = data?.id || ''
    items.value = (data?.items || []).map(item => ({
      id: item.id,
      productId: item.product_id,
      variantId: item.variant_id || '',
      name: item.product_name,
      variantName: item.variant_name || '',
      variantAttributes: item.variant_attributes || {},
      quantity: item.quantity,
      stock: item.available_stock,
      originalPrice: item.original_unit_price,
      price: item.flash_sale_unit_price,
      flashSaleQuantity: item.flash_sale_quantity,
      normalQuantity: item.normal_quantity,
      lineTotal: item.line_total,
      discountPercent: item.flash_sale_discount_percent
    }))
    backendSubtotal.value = data?.subtotal || 0
    backendTotalItems.value = data?.total_items || 0
    const existingIds = new Set(items.value.map(item => item.id))
    selectedItemIds.value = selectedItemIds.value.filter(itemId => existingIds.has(itemId))
  }

  async function loadCart(force = false) {
    if (loaded && !force) return
    loading.value = true
    error.value = null
    try {
      const response = await cartApi.getByUserId(currentUserId())
      applyCart(response.data.data)
      if (!selectionInitialized) {
        selectedItemIds.value = items.value.map(item => item.id)
        selectionInitialized = true
      }
      loaded = true
    } catch (e) {
      error.value = e.response?.data?.error || 'Failed to load cart'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function addItem(product, variant, quantity = 1) {
    loading.value = true
    error.value = null
    try {
      const response = await cartApi.addItem({
        user_id: currentUserId(),
        product_id: product.id,
        variant_id: variant?.id || undefined,
        quantity
      })
      applyCart(response.data.data)
      const addedItem = items.value.find(item => (
        item.productId === product.id && item.variantId === (variant?.id || '')
      ))
      if (addedItem && !selectedItemIds.value.includes(addedItem.id)) {
        selectedItemIds.value = [...selectedItemIds.value, addedItem.id]
      }
      selectionInitialized = true
      loaded = true
    } catch (e) {
      error.value = e.response?.data?.error || 'Failed to add cart item'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function removeItem(itemId) {
    loading.value = true
    error.value = null
    try {
      await cartApi.deleteItem(itemId, currentUserId())
      await loadCart(true)
    } catch (e) {
      error.value = e.response?.data?.error || 'Failed to remove cart item'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function updateQuantity(itemId, quantity) {
    loading.value = true
    error.value = null
    try {
      const response = await cartApi.updateItem(itemId, {
        user_id: currentUserId(),
        quantity
      })
      applyCart(response.data.data)
    } catch (e) {
      error.value = e.response?.data?.error || 'Failed to update cart item'
      throw e
    } finally {
      loading.value = false
    }
  }

  function setPromo(code, discount) {
    promoCode.value = code
    promoDiscount.value = discount
  }

  function clearPromo() {
    promoCode.value = ''
    promoDiscount.value = 0
  }

  function toggleItem(itemId) {
    selectedItemIds.value = selectedItemIds.value.includes(itemId)
      ? selectedItemIds.value.filter(id => id !== itemId)
      : [...selectedItemIds.value, itemId]
    clearPromo()
  }

  function toggleAll() {
    selectedItemIds.value = allItemsSelected.value ? [] : items.value.map(item => item.id)
    clearPromo()
  }

  async function removeSelectedItems() {
    const itemIds = [...selectedItemIds.value]
    await Promise.all(itemIds.map(itemId => cartApi.deleteItem(itemId, currentUserId())))
    selectedItemIds.value = []
    selectionInitialized = true
    await loadCart(true)
    clearPromo()
  }

  async function clearCart() {
    await cartApi.clear(currentUserId())
    applyCart(null)
    selectedItemIds.value = []
    clearPromo()
    loaded = true
  }

  return {
    id,
    items,
    promoCode,
    promoDiscount,
    selectedItemIds,
    totalItems,
    subtotal,
    total,
    selectedItems,
    selectedTotalItems,
    selectedSubtotal,
    selectedTotal,
    allItemsSelected,
    loading,
    error,
    loadCart,
    addItem,
    removeItem,
    updateQuantity,
    setPromo,
    clearPromo,
    toggleItem,
    toggleAll,
    removeSelectedItems,
    clearCart
  }
})
