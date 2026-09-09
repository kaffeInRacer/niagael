<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h2 class="text-2xl font-bold text-gray-800">Products</h2>
        <p class="text-gray-500">Manage your product inventory</p>
      </div>
      <button
        @click="showModal = true; resetForm()"
        class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center"
      >
        <svg class="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Add Product
      </button>
    </div>

    <div class="bg-white rounded-lg shadow p-4 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <select
          v-model="filters.category_id"
          class="border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          @change="reloadTable(true)"
        >
          <option value="">All Categories</option>
          <option v-for="cat in categoryList" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
        </select>
        <select
          v-model="filters.is_active"
          class="border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          @change="reloadTable(true)"
        >
          <option value="">All Status</option>
          <option value="true">Active</option>
          <option value="false">Inactive</option>
        </select>
      </div>
    </div>

    <div v-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
      <p class="text-red-600">{{ error }}</p>
    </div>

    <div class="bg-white rounded-lg shadow overflow-hidden">
      <AdminDataTable
        ref="table"
        :columns="columns"
        :ajax="tableAjax"
        :server-side="true"
        :options="tableOptions"
        @action="handleTableAction"
      >
      </AdminDataTable>

    </div>

    <div v-if="showModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 overflow-y-auto" role="dialog" aria-modal="true" aria-labelledby="product-modal-title">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-lg mx-4 my-8">
        <div class="px-6 py-4 border-b">
          <h3 id="product-modal-title" class="text-lg font-semibold text-gray-800">{{ editingId ? 'Edit Product' : 'Add Product' }}</h3>
        </div>
        <form @submit.prevent="saveProduct" class="px-6 py-4 space-y-4 max-h-[70vh] overflow-y-auto">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
            <input v-model="form.name" type="text" required class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea v-model="form.description" rows="3" class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"></textarea>
          </div>
          <div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Category *</label>
              <select v-model="form.category_id" required class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500">
                <option value="">Select category</option>
                <option v-for="cat in categoryList" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
              </select>
            </div>
          </div>
          <div v-if="editingHasVariants" class="bg-blue-50 border border-blue-200 rounded-lg p-3">
            <p class="text-sm text-blue-700">Price and stock are managed at variant level. Edit variants to update price and stock.</p>
          </div>
          <div v-else class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Price (Rp) *</label>
              <input v-model.number="form.price" type="number" required min="0" class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Stock *</label>
              <input v-model.number="form.stock" type="number" required min="0" class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            </div>
          </div>
          <div class="flex items-center">
            <input v-model="form.is_active" type="checkbox" id="is_active" class="h-4 w-4 text-blue-600 rounded" />
            <label for="is_active" class="ml-2 text-sm text-gray-700">Active</label>
          </div>
          <div v-if="formError" class="bg-red-50 border border-red-200 rounded-lg p-3" role="alert">
            <p class="text-sm text-red-600">{{ formError }}</p>
          </div>
          <div class="flex justify-end gap-3 pt-4 border-t">
            <button type="button" @click="showModal = false" class="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg">Cancel</button>
            <button type="submit" :disabled="saving" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50">
              {{ saving ? 'Saving...' : (editingId ? 'Update' : 'Create') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="showVariantModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 overflow-y-auto" role="dialog" aria-modal="true" aria-labelledby="variant-modal-title">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-4xl mx-4 my-8">
        <div class="px-6 py-4 border-b flex justify-between items-center">
          <div>
            <h3 id="variant-modal-title" class="text-lg font-semibold text-gray-800">Variants</h3>
            <p class="text-sm text-gray-500">{{ selectedProduct?.name }}</p>
          </div>
          <button type="button" @click="showVariantModal = false" class="text-gray-400 hover:text-gray-600" aria-label="Close variants">
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="px-6 py-4">
          <div v-if="!showVariantForm" class="space-y-4">
            <div class="flex justify-between items-center">
              <p class="text-sm text-gray-500">{{ variants.length }} variant(s)</p>
              <button
                @click="openVariantForm(null)"
                class="bg-blue-600 text-white px-3 py-1.5 rounded-lg text-sm hover:bg-blue-700 flex items-center"
              >
                <svg class="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Add Variant
              </button>
            </div>

            <div v-if="variantLoading" class="text-center py-8">
              <p class="text-gray-500">Loading variants...</p>
            </div>

            <div v-else-if="variants.length === 0" class="text-center py-8 bg-gray-50 rounded-lg">
              <p class="text-gray-500">No variants yet. Add your first variant.</p>
            </div>

            <div v-else class="space-y-3">
              <div
                v-for="variant in variants"
                :key="variant.id"
                class="border rounded-lg p-4 hover:bg-gray-50"
              >
                <div class="flex justify-between items-start">
                  <div class="flex-1">
                    <div class="flex items-center gap-2 mb-2">
                      <span class="font-medium text-gray-900">{{ variant.name }}</span>
                      <span
                        class="px-2 py-0.5 text-xs rounded-full"
                        :class="variant.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'"
                      >
                        {{ variant.is_active ? 'Active' : 'Inactive' }}
                      </span>
                    </div>
                    <div class="flex items-center gap-4 text-sm">
                      <span class="text-gray-600">Price: <strong>{{ formatPrice(variant.price) }}</strong></span>
                      <span class="text-gray-600">Stock: <strong>{{ variant.stock }}</strong></span>
                    </div>
                    <div v-if="variant.attributes" class="mt-2 flex flex-wrap gap-2">
                      <span
                        v-for="(value, key) in variant.attributes"
                        :key="key"
                        class="px-2 py-1 bg-gray-100 text-gray-700 text-xs rounded"
                      >
                        {{ key }}: {{ value }}
                      </span>
                    </div>
                  </div>
                  <div class="flex gap-2 ml-4">
                    <button @click="openVariantForm(variant)" class="text-blue-600 hover:text-blue-900 text-sm">Edit</button>
                    <button @click="deleteVariant(variant.id)" class="text-red-600 hover:text-red-900 text-sm">Delete</button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-else>
            <div class="flex items-center gap-2 mb-4">
              <button @click="showVariantForm = false" class="text-gray-500 hover:text-gray-700">
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
                </svg>
              </button>
              <h4 class="font-semibold text-gray-800">{{ editingVariantId ? 'Edit Variant' : 'Add Variant' }}</h4>
            </div>

            <form @submit.prevent="saveVariant" class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
                <input
                  v-model="variantForm.name"
                  type="text"
                  required
                  class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="e.g. HP 32GB 1TB Silver"
                />
              </div>

              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">Price (Rp) *</label>
                  <input
                    v-model.number="variantForm.price"
                    type="number"
                    required
                    min="0"
                    class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">Stock *</label>
                  <input
                    v-model.number="variantForm.stock"
                    type="number"
                    required
                    min="0"
                    class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
              </div>

              <div>
                <div class="flex justify-between items-center mb-2">
                  <label class="block text-sm font-medium text-gray-700">Attributes</label>
                  <button
                    type="button"
                    @click="addAttribute"
                    class="text-blue-600 hover:text-blue-800 text-sm flex items-center"
                  >
                    <svg class="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Add Attribute
                  </button>
                </div>

                <div v-if="attributeList.length === 0" class="text-sm text-gray-500 bg-gray-50 rounded-lg p-3">
                  No attributes added. Click "Add Attribute" to add properties like color, size, etc.
                </div>

                <div v-else class="space-y-2">
                  <div
                    v-for="(attr, index) in attributeList"
                    :key="index"
                    class="flex gap-2"
                  >
                    <input
                      v-model="attr.key"
                      type="text"
                      placeholder="Key (e.g. color)"
                      class="w-1/3 border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <input
                      v-model="attr.value"
                      type="text"
                      placeholder="Value (e.g. Red)"
                      class="flex-1 border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <button
                      type="button"
                       @click="removeAttribute(index)"
                       :aria-label="`Remove attribute ${index + 1}`"
                      class="text-red-500 hover:text-red-700 px-2"
                    >
                      <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                </div>
              </div>

              <div class="flex items-center">
                <input v-model="variantForm.is_active" type="checkbox" id="variant_active" class="h-4 w-4 text-blue-600 rounded" />
                <label for="variant_active" class="ml-2 text-sm text-gray-700">Active</label>
              </div>

              <div v-if="variantFormError" class="bg-red-50 border border-red-200 rounded-lg p-3">
                <p class="text-sm text-red-600">{{ variantFormError }}</p>
              </div>

              <div class="flex justify-end gap-3 pt-4 border-t">
                <button type="button" @click="showVariantForm = false" class="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg">Cancel</button>
                <button type="submit" :disabled="variantSaving" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50">
                  {{ variantSaving ? 'Saving...' : (editingVariantId ? 'Update' : 'Create') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../../api'
import AdminDataTable from '../../components/AdminDataTable.vue'
import { createServerSideAjax, escapeHtml } from '../../utils/datatables'
import { formatPrice } from '../../utils/format'

const categoryList = ref([])
const error = ref(null)
const showModal = ref(false)
const saving = ref(false)
const formError = ref(null)
const editingId = ref(null)
const editingHasVariants = ref(false)
const table = ref(null)

const showVariantModal = ref(false)
const showVariantForm = ref(false)
const selectedProduct = ref(null)
const variants = ref([])
const variantLoading = ref(false)
const editingVariantId = ref(null)
const variantSaving = ref(false)
const variantFormError = ref(null)
const attributeList = ref([])

const filters = ref({
  category_id: '',
  is_active: ''
})

const discountPercent = (row) => {
  if (!row?.price || !row?.discount) return 0
  return Math.round((row.discount / row.price) * 100)
}

const columns = [
  { data: 'name', title: 'Product', render: function(data, type, row) {
    return `<div class="text-sm font-medium text-gray-900">${escapeHtml(data)}</div><div class="text-sm text-gray-500">${escapeHtml(row.slug)}</div>`
  }},
  { data: 'category', title: 'Category', orderable: false, render: function(data) {
    return escapeHtml(data || '-')
  }},
  { data: null, title: 'Variants', orderable: false, searchable: false, render: function(data, type, row) {
    if (!row) return ''
    const cls = row.has_variants ? 'bg-blue-100 text-blue-800 hover:bg-blue-200' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
    const label = row.has_variants ? 'Manage Variants' : 'Add Variants'
    return `<button type="button" class="px-2 py-1 text-xs font-medium rounded-full cursor-pointer transition-colors ${cls}" data-table-action="variants">${label}</button>`
  }},
  { data: null, title: 'Actions', orderable: false, searchable: false, render: function(data, type, row) {
    if (!row) return ''
    return '<div class="flex justify-end gap-3"><button type="button" class="text-gray-600 hover:text-gray-900" data-table-action="view">View</button><button type="button" class="text-blue-600 hover:text-blue-900" data-table-action="edit">Edit</button><button type="button" class="text-red-600 hover:text-red-900" data-table-action="delete">Delete</button></div>'
  }}
]

const tableOptions = {
  pageLength: 10,
  order: [[0, 'desc']],
  columnDefs: [
    { orderable: false, targets: [1, 2, 3] },
    { searchable: false, targets: [2, 3] }
  ]
}

const tableAjax = createServerSideAjax({
  fetchPage: (params) => api.get('/admin/products', { params }),
  orderColumns: { 0: 'name' },
  getFilters: () => ({
    category_id: filters.value.category_id || undefined,
    is_active: filters.value.is_active || undefined
  }),
  onError: (e) => {
    error.value = e.response?.data?.error || 'Failed to load products'
  },
  onSuccess: () => { error.value = null }
})

const reloadTable = (resetPaging = false) => {
  error.value = null
  table.value?.reload(resetPaging)
}

const form = ref({
  name: '',
  description: '',
  category_id: '',
  price: 0,
  stock: 0,
  is_active: true
})

const variantForm = ref({
  name: '',
  price: 0,
  stock: 0,
  is_active: true
})

const resetForm = () => {
  form.value = { name: '', description: '', category_id: '', price: 0, stock: 0, is_active: true }
  editingId.value = null
  editingHasVariants.value = false
  formError.value = null
}

const fetchCategories = async () => {
  try {
    const res = await api.get('/admin/categories', { params: { page_size: 100 } })
    categoryList.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load categories', e)
  }
}

const editProduct = async (product) => {
  const requestId = ++editRequestId
  editingId.value = product.id
  editingHasVariants.value = product.has_variants || false
  showModal.value = true
  saving.value = true
  try {
    const res = await api.get(`/admin/products/${product.id}`)
    if (requestId !== editRequestId || editingId.value !== product.id || !showModal.value) return
    const detail = res.data.data
    form.value = {
      name: detail.name || product.name,
      description: detail.description || '',
      category_id: detail.category_id || '',
      price: detail.price || 0,
      stock: detail.stock || 0,
      is_active: detail.is_active ?? true
    }
  } catch (e) {
    if (requestId !== editRequestId) return
    formError.value = 'Failed to load product details'
  } finally {
    if (requestId === editRequestId) saving.value = false
  }
}

const viewProduct = (product) => {
  window.open(`/product/${product.id}`, '_blank')
}

const saveProduct = async () => {
  saving.value = true
  formError.value = null
  try {
    if (editingId.value) {
      await api.put(`/admin/products/${editingId.value}`, form.value)
    } else {
      await api.post('/admin/products', form.value)
    }
    showModal.value = false
    table.value?.reload()
  } catch (e) {
    formError.value = e.response?.data?.error || e.response?.data?.errors || 'Failed to save product'
  } finally {
    saving.value = false
  }
}

const deleteProduct = async (id) => {
  if (!confirm('Are you sure you want to delete this product?')) return
  try {
    await api.delete(`/admin/products/${id}`)
    table.value?.reload()
  } catch (e) {
    alert(e.response?.data?.error || 'Failed to delete product')
  }
}

const openVariants = async (product) => {
  selectedProduct.value = product
  showVariantModal.value = true
  showVariantForm.value = false
  await fetchVariants(product.id)
}

const fetchVariants = async (productId) => {
  variantLoading.value = true
  try {
    const res = await api.get('/admin/variants', { params: { product_id: productId } })
    if (selectedProduct.value?.id !== productId || !showVariantModal.value) return
    variants.value = res.data.data || []
  } catch (e) {
    if (selectedProduct.value?.id !== productId) return
    console.error('Failed to load variants', e)
    variants.value = []
  } finally {
    if (selectedProduct.value?.id === productId) variantLoading.value = false
  }
}

const openVariantForm = (variant) => {
  if (variant) {
    editingVariantId.value = variant.id
    variantForm.value = {
      name: variant.name,
      price: variant.price,
      stock: variant.stock,
      is_active: variant.is_active
    }
    attributeList.value = Object.entries(variant.attributes || {}).map(([key, value]) => ({ key, value }))
  } else {
    editingVariantId.value = null
    variantForm.value = { name: '', price: 0, stock: 0, is_active: true }
    attributeList.value = []
  }
  variantFormError.value = null
  showVariantForm.value = true
}

const addAttribute = () => {
  attributeList.value.push({ key: '', value: '' })
}

const removeAttribute = (index) => {
  attributeList.value.splice(index, 1)
}

const saveVariant = async () => {
  variantSaving.value = true
  variantFormError.value = null

  const attributes = {}
  for (const attr of attributeList.value) {
    if (attr.key.trim()) {
      attributes[attr.key.trim()] = attr.value
    }
  }

  const payload = {
    ...variantForm.value,
    attributes
  }

  try {
    if (editingVariantId.value) {
      await api.put(`/admin/variants/${editingVariantId.value}`, payload)
    } else {
      await api.post(`/admin/variants?product_id=${selectedProduct.value.id}`, payload)
    }
    showVariantForm.value = false
    await fetchVariants(selectedProduct.value.id)
    table.value?.reload()
  } catch (e) {
    variantFormError.value = e.response?.data?.error || 'Failed to save variant'
  } finally {
    variantSaving.value = false
  }
}

const deleteVariant = async (id) => {
  if (!confirm('Are you sure you want to delete this variant?')) return
  try {
    await api.delete(`/admin/variants/${id}`)
    await fetchVariants(selectedProduct.value.id)
    table.value?.reload()
  } catch (e) {
    alert(e.response?.data?.error || 'Failed to delete variant')
  }
}

onMounted(() => {
  fetchCategories()
})

let editRequestId = 0

const handleTableAction = ({ action, row }) => {
  if (action === 'variants') openVariants(row)
  if (action === 'view') viewProduct(row)
  if (action === 'edit') editProduct(row)
  if (action === 'delete') deleteProduct(row.id)
}
</script>
