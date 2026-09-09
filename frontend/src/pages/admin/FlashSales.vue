<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h2 class="text-2xl font-bold text-gray-800">Flash Sales</h2>
        <p class="text-gray-500">Manage flash sale promotions</p>
      </div>
      <button
        @click="showModal = true; resetForm()"
        class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center"
      >
        <svg class="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Add Flash Sale
      </button>
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

    <ModalDialog :show="!!detailItem" title="Flash Sale Details" aria-id="fs-detail-modal-title" size="lg" scroll-body @close="detailItem = null">
        <div class="px-6 py-4 space-y-4" v-if="detailItem">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500">Flash Sale</p>
              <p class="font-semibold text-gray-800">{{ detailItem.name }}</p>
            </div>
            <span class="px-2 py-1 text-xs font-medium rounded-full" :class="detailItem.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'">
              {{ detailItem.is_active ? 'Active' : 'Inactive' }}
            </span>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-sm text-gray-500">Period</p>
              <p class="text-sm">{{ formatDate(detailItem.start_time) }} - {{ formatDate(detailItem.end_time) }}</p>
            </div>
            <div>
              <p class="text-sm text-gray-500">Total FS Stock (all items)</p>
              <p class="text-sm">{{ detailItem.total_stock }}</p>
            </div>
          </div>
          <div class="border-t pt-4">
            <h4 class="text-sm font-semibold text-gray-700 mb-3">Items on Flash Sale ({{ detailProducts.length }})</h4>
            <div v-if="detailLoading" class="flex justify-center py-6" role="status" aria-label="Loading details">
              <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
            </div>
            <div v-else-if="detailProducts.length > 0" class="space-y-3">
              <div v-for="p in detailProducts" :key="p.id" class="flex items-center gap-4 border rounded-lg p-3">
                <img
                  v-if="p.image"
                  :src="productImageUrl(p)"
                  :alt="p.name"
                  class="w-14 h-14 rounded-lg object-cover flex-shrink-0"
                />
                <div v-else class="w-14 h-14 rounded-lg bg-gray-100 flex items-center justify-center flex-shrink-0">
                  <svg class="w-6 h-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                  </svg>
                </div>
                <div class="flex-1 min-w-0">
                  <p class="font-medium text-gray-800 truncate">{{ p.name }}</p>
                  <p class="text-sm text-gray-500">
                    <template v-if="p.discount_percent > 0">
                      <span class="line-through">{{ formatPrice(p.base_price) }}</span>
                      <span class="ml-2 font-semibold text-red-600">{{ formatPrice(p.final_price) }}</span>
                      <span class="ml-2 px-1.5 py-0.5 bg-red-50 text-red-600 rounded text-xs font-semibold">-{{ p.discount_percent }}%</span>
                    </template>
                    <span v-else class="font-semibold text-gray-800">{{ formatPrice(p.final_price) }}</span>
                  </p>
                  <p class="text-xs text-gray-400 mt-0.5">
                    <span v-if="p.category">Category: {{ p.category }} · </span>
                    <span>Stock: {{ p.stock ?? '-' }}</span>
                    <span v-if="p.max_per_user"> · Max {{ p.max_per_user }}/user</span>
                  </p>
                </div>
                <div v-if="detailItem.variant_id" class="text-right">
                  <p class="text-xs text-gray-400">Variant</p>
                  <p class="text-sm font-medium text-gray-700">{{ p.variant?.name || 'Default' }}</p>
                </div>
              </div>
            </div>
            <p v-else class="text-sm text-gray-500">Product no longer available</p>
          </div>
        </div>
    </ModalDialog>

    <ModalDialog :show="showModal" :title="editingId ? 'Edit Flash Sale' : 'Add Flash Sale'" aria-id="flash-sale-modal-title" size="xl" scroll-body @close="showModal = false">
        <form @submit.prevent="saveFlashSale" class="px-6 py-4 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
            <input v-model="form.name" type="text" required placeholder="Flash Sale name" class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Start Time *</label>
              <input v-model="form.start_time" type="datetime-local" required class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">End Time *</label>
              <input v-model="form.end_time" type="datetime-local" required class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            </div>
          </div>
          <div class="flex items-center gap-2">
            <input v-model="form.is_active" type="checkbox" id="is_active" class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded" />
            <label for="is_active" class="text-sm font-medium text-gray-700">Active</label>
          </div>

          <div>
            <div class="flex justify-between items-center mb-2">
              <label class="block text-sm font-medium text-gray-700">Products *</label>
              <button type="button" @click="addItem" class="text-blue-600 hover:text-blue-800 text-sm font-medium flex items-center">
                <svg class="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Add Product
              </button>
            </div>
            <div class="border rounded-lg overflow-hidden">
              <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                  <tr>
                    <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Product</th>
                    <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Variant</th>
                    <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Discount (%)</th>
                    <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Stock</th>
                    <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Max/User</th>
                    <th class="px-3 py-2 w-10"></th>
                  </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                  <tr v-for="(item, index) in form.items" :key="index">
                    <td class="px-3 py-2">
                      <select v-model="item.product_id" @change="onProductChange(index)" required class="w-full border rounded px-2 py-1 text-sm focus:ring-2 focus:ring-blue-500">
                        <option value="">Select</option>
                        <option v-for="p in productList" :key="p.id" :value="p.id">{{ p.name }}</option>
                      </select>
                    </td>
                    <td class="px-3 py-2">
                      <select v-model="item.variant_id" :disabled="!item.variants?.length" class="w-full border rounded px-2 py-1 text-sm focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100">
                        <option value="">No variant</option>
                        <option v-for="v in (item.variants || [])" :key="v.id" :value="v.id">{{ v.name }}</option>
                      </select>
                    </td>
                    <td class="px-3 py-2">
                      <input v-model.number="item.discount_percent" type="number" required min="1" max="100" step="1" class="w-full border rounded px-2 py-1 text-sm focus:ring-2 focus:ring-blue-500" placeholder="20" />
                    </td>
                    <td class="px-3 py-2">
                      <input v-model.number="item.stock" type="number" required min="1" class="w-full border rounded px-2 py-1 text-sm focus:ring-2 focus:ring-blue-500" placeholder="10" />
                    </td>
                    <td class="px-3 py-2">
                      <input v-model.number="item.max_per_user" type="number" min="1" class="w-full border rounded px-2 py-1 text-sm focus:ring-2 focus:ring-blue-500" placeholder="2" />
                    </td>
                    <td class="px-3 py-2">
                      <button v-if="form.items.length > 1" type="button" @click="removeItem(index)" class="text-red-500 hover:text-red-700" :aria-label="`Remove product ${index + 1}`">
                        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
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
    </ModalDialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../../api'
import { flashSaleApi, adminApi, productApi } from '../../api'
import AdminDataTable from '../../components/AdminDataTable.vue'
import ModalDialog from '../../components/ModalDialog.vue'
import { createServerSideAjax, escapeHtml } from '../../utils/datatables'
import { formatDate, formatPrice, getFlashSaleStatus, getFlashSaleStatusClass, localDateTimeToIso, toLocalDateTimeInput } from '../../utils/format'

const productList = ref([])
const error = ref(null)
const showModal = ref(false)
const saving = ref(false)
const formError = ref(null)
const editingId = ref(null)
const table = ref(null)
const detailItem = ref(null)
const detailLoading = ref(false)
const detailProducts = ref([])

const createEmptyItem = () => ({
  product_id: '',
  variant_id: '',
  variants: [],
  discount_percent: 20,
  stock: 10,
  max_per_user: 2
})

const form = ref({
  name: '',
  start_time: '',
  end_time: '',
  is_active: true,
  items: [createEmptyItem()]
})

const columns = [
  { data: 'name', title: 'Name' },
  { data: 'product_id', title: 'Product', orderable: false, render: function(data, type, row) {
    if (!row) return ''
    const p = productList.value.find(pr => pr.id === data)
    const label = p ? p.name : `<span class="font-mono">${escapeHtml(String(data ?? '').slice(0, 8))}...</span>`
    if (row.variant_id) {
      const v = p?.variants?.find(vr => vr.id === row.variant_id)
      return label + (v ? `<span class="block text-xs text-gray-500">${escapeHtml(v.name)}</span>` : '')
    }
    return label
  }},
  { data: 'discount_percent', title: 'Discount (%)', orderable: false, render: function(data) {
    return `<span class="font-medium">${escapeHtml(data)}%</span>`
  }},
  { data: 'stock', title: 'Stock', orderable: false },
  { data: 'max_per_user', title: 'Max/User', orderable: false },
  { data: null, title: 'Period', render: function(data, type, row) {
    if (!row) return ''
    return `${formatDate(row.start_time)} - ${formatDate(row.end_time)}`
  }},
  { data: null, title: 'Status', orderable: false, render: function(data, type, row) {
    if (!row) return ''
    const status = row.is_active === true ? getFlashSaleStatus(row.start_time, row.end_time) : 'inactive'
    const cls = getFlashSaleStatusClass(status)
    const label = status.charAt(0).toUpperCase() + status.slice(1)
    return `<span class="px-2 py-1 text-xs font-medium rounded-full ${cls}">${label}</span>`
  }},
  { data: null, title: 'Actions', orderable: false, searchable: false, render: function(data, type, row) {
    if (!row) return ''
    return '<div class="flex justify-end gap-3"><button type="button" class="text-blue-600 hover:text-blue-900" data-table-action="details">Details</button><button type="button" class="text-blue-600 hover:text-blue-900" data-table-action="edit">Edit</button><button type="button" class="text-red-600 hover:text-red-900" data-table-action="delete">Delete</button></div>'
  }}
]

const tableOptions = {
  pageLength: 25,
  lengthMenu: [10, 25, 50, 100],
  order: [[0, 'asc']],
  columnDefs: [
    { orderable: false, targets: [1, 2, 3, 4, 6, 7] },
    { searchable: false, targets: [7] }
  ]
}

const tableAjax = createServerSideAjax({
  fetchPage: (params) => api.get('/admin/flash-sales', { params }),
  orderColumns: { 0: 'name', 5: 'start_time' },
  onError: (e) => {
    error.value = e.response?.data?.error || 'Failed to load flash sales'
  },
  onSuccess: () => { error.value = null }
})

const resetForm = () => {
  form.value = {
    name: '',
    start_time: '',
    end_time: '',
    is_active: true,
    items: [createEmptyItem()]
  }
  editingId.value = null
  formError.value = null
}

const fetchProducts = async () => {
  try {
    const res = await api.get('/admin/products', { params: { page_size: 100 } })
    productList.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load products', e)
  }
}

const fetchVariantsForItem = async (index) => {
  const item = form.value.items[index]
  if (!item.product_id) {
    item.variants = []
    item.variant_id = ''
    return
  }
  const productId = item.product_id
  try {
    const res = await api.get('/admin/variants', { params: { product_id: productId } })
    if (form.value.items[index] !== item || item.product_id !== productId) return
    item.variants = res.data.data || []
  } catch (e) {
    if (form.value.items[index] !== item || item.product_id !== productId) return
    item.variants = []
  }
}

const onProductChange = (index) => {
  const item = form.value.items[index]
  item.variant_id = ''
  fetchVariantsForItem(index)
}

const addItem = () => {
  form.value.items.push(createEmptyItem())
}

const removeItem = (index) => {
  form.value.items.splice(index, 1)
}

const editFlashSale = (fs) => {
  editingId.value = fs.id
  form.value = {
    name: fs.name,
    start_time: toLocalDateTimeInput(fs.start_time),
    end_time: toLocalDateTimeInput(fs.end_time),
    is_active: fs.is_active ?? true,
    items: [{
      product_id: fs.product_id || '',
      variant_id: fs.variant_id || '',
      variants: [],
      discount_percent: fs.discount_percent,
      stock: fs.stock,
      max_per_user: fs.max_per_user
    }]
  }
  if (fs.product_id) fetchVariantsForItem(0)
  showModal.value = true
}

const saveFlashSale = async () => {
  saving.value = true
  formError.value = null
  try {
    if (editingId.value) {
      const item = form.value.items[0]
      const payload = {
        name: form.value.name,
        product_id: item.product_id,
        variant_id: item.variant_id || undefined,
        discount_percent: item.discount_percent,
        stock: item.stock,
        max_per_user: item.max_per_user,
        start_time: localDateTimeToIso(form.value.start_time),
        end_time: localDateTimeToIso(form.value.end_time),
        is_active: form.value.is_active
      }
      await api.put(`/admin/flash-sales/${editingId.value}`, payload)
    } else {
      const payload = {
        name: form.value.name,
        start_time: localDateTimeToIso(form.value.start_time),
        end_time: localDateTimeToIso(form.value.end_time),
        is_active: form.value.is_active,
        items: form.value.items.map(item => ({
          product_id: item.product_id,
          variant_id: item.variant_id || undefined,
          discount_percent: item.discount_percent,
          stock: item.stock,
          max_per_user: item.max_per_user
        }))
      }
      await flashSaleApi.createBulk(payload)
    }
    showModal.value = false
    table.value?.reload()
  } catch (e) {
    formError.value = e.response?.data?.error || e.response?.data?.errors || 'Failed to save flash sale'
  } finally {
    saving.value = false
  }
}

const deleteFlashSale = async (id) => {
  if (!confirm('Are you sure you want to delete this flash sale?')) return
  try {
    await api.delete(`/admin/flash-sales/${id}`)
    table.value?.reload()
  } catch (e) {
    alert(e.response?.data?.error || 'Failed to delete flash sale')
  }
}

onMounted(() => {
  fetchProducts()
})

const handleTableAction = ({ action, row }) => {
  if (action === 'details') openDetail(row)
  if (action === 'edit') editFlashSale(row)
  if (action === 'delete') deleteFlashSale(row.id)
}

const buildDetailRow = async (item) => {
  let p = null
  try {
    const { data } = await adminApi.products.getById(item.product_id)
    p = data.data || data
  } catch {
    try {
      const { data } = await productApi.getById(item.product_id)
      p = data.data || data
    } catch { p = null }
  }
  if (!p || !p.id) return null

  let variant = null
  if (item.variant_id) {
    try {
      const vres = await variantApi.list({ product_id: item.product_id })
      variant = (vres.data.data || []).find(v => v.id === item.variant_id) || null
    } catch { variant = null }
  }

  const basePrice = variant ? variant.price : p.price
  const finalPrice = Math.round(basePrice * (1 - item.discount_percent / 100))
  const categoryName = typeof p.category === 'object' ? (p.category?.name || '') : (p.category || '')

  return {
    id: item.id,
    product_id: item.product_id,
    name: p.name,
    category: categoryName,
    image: p.image,
    variant,
    base_price: basePrice,
    final_price: finalPrice,
    discount_percent: item.discount_percent,
    stock: item.stock,
    max_per_user: item.max_per_user,
    start_time: item.start_time,
    end_time: item.end_time
  }
}

const openDetail = async (fs) => {
  detailItem.value = null
  detailProducts.value = []
  detailLoading.value = true
  try {
    const { data: listRes } = await flashSaleApi.adminList({ name: fs.name, page: 1, page_size: 100 })
    const items = listRes.data || []

    detailItem.value = {
      id: fs.id,
      name: fs.name,
      is_active: fs.is_active,
      start_time: fs.start_time,
      end_time: fs.end_time,
      total_stock: items.reduce((sum, i) => sum + Number(i.stock || 0), 0)
    }

    const rows = await Promise.all(items.map(buildDetailRow))
    detailProducts.value = rows.filter(Boolean)
  } catch {
    detailProducts.value = []
    error.value = 'Failed to load flash sale details'
  } finally {
    detailLoading.value = false
  }
}

const productImageUrl = (p) => {
  const base = import.meta.env.VITE_IMG_URL || ''
  return `${base}/${p.id}/${p.image.file_name}`
}
</script>
