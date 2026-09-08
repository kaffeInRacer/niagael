<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h2 class="text-2xl font-bold text-gray-800">Promos</h2>
        <p class="text-gray-500">Manage promotional codes and discounts</p>
      </div>
      <button
        @click="showModal = true; resetForm()"
        class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center"
      >
        <svg class="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Add Promo
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
      >
        <template #promo-actions="props">
          <div class="flex justify-end gap-3">
            <button @click="editPromo(props.rowData)" class="text-blue-600 hover:text-blue-900">Edit</button>
            <button @click="deletePromo(props.rowData.id)" class="text-red-600 hover:text-red-900">Delete</button>
          </div>
        </template>
      </AdminDataTable>
    </div>

    <div v-if="showModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 overflow-y-auto">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-lg mx-4 my-8">
        <div class="px-6 py-4 border-b">
          <h3 class="text-lg font-semibold text-gray-800">{{ editingId ? 'Edit Promo' : 'Add Promo' }}</h3>
        </div>
        <form @submit.prevent="savePromo" class="px-6 py-4 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Code *</label>
              <input v-model="form.code" type="text" required class="w-full border rounded-lg px-3 py-2 font-mono uppercase focus:ring-2 focus:ring-blue-500 focus:border-blue-500" placeholder="e.g. DISCOUNT20" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
              <input v-model="form.name" type="text" required class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea v-model="form.description" rows="2" class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"></textarea>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Discount Type *</label>
              <select v-model="form.discount_type" required class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500">
                <option value="percentage">Percentage (%)</option>
                <option value="fixed">Fixed Amount (Rp)</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Discount Value *</label>
              <input v-model.number="form.discount_value" type="number" required min="0" class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Min Purchase (Rp)</label>
              <input v-model.number="form.min_purchase" type="number" min="0" class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Max Discount (Rp)</label>
              <input v-model.number="form.max_discount" type="number" min="0" class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Quantity</label>
              <input v-model.number="form.quantity" type="number" min="0" class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Max Usage/User</label>
              <input v-model.number="form.max_usage_per_user" type="number" min="0" class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            </div>
          </div>
          <div class="flex items-center">
            <input v-model="form.can_combine_flash_sale" type="checkbox" id="can_combine" class="h-4 w-4 text-blue-600 rounded" />
            <label for="can_combine" class="ml-2 text-sm text-gray-700">Can combine with flash sale</label>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Start Date *</label>
              <input v-model="form.start_date" type="datetime-local" required class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">End Date *</label>
              <input v-model="form.end_date" type="datetime-local" required class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            </div>
          </div>
          <div v-if="formError" class="bg-red-50 border border-red-200 rounded-lg p-3">
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
  </div>
</template>

<script setup>
import { ref } from 'vue'
import api from '../../api'
import AdminDataTable from '../../components/AdminDataTable.vue'
import { createServerSideAjax } from '../../utils/datatables'
import { formatDate, formatPrice, isPromoActive } from '../../utils/format'

const error = ref(null)
const showModal = ref(false)
const saving = ref(false)
const formError = ref(null)
const editingId = ref(null)
const table = ref(null)

const form = ref({
  code: '',
  name: '',
  description: '',
  discount_type: 'percentage',
  discount_value: 10,
  min_purchase: 0,
  max_discount: 0,
  quantity: 100,
  max_usage_per_user: 1,
  can_combine_flash_sale: false,
  start_date: '',
  end_date: ''
})

const columns = [
  { data: 'code', title: 'Code', render: function(data) {
    return `<span class="px-2 py-1 bg-gray-100 rounded font-mono text-sm font-medium">${data}</span>`
  }},
  { data: 'name', title: 'Name' },
  { data: null, title: 'Discount', orderable: false, render: function(data, type, row) {
    return row.discount_type === 'percentage' ? `${row.discount_value}%` : formatPrice(row.discount_value)
  }},
  { data: 'min_purchase', title: 'Min Purchase', orderable: false, render: function(data) {
    return formatPrice(data)
  }},
  { data: null, title: 'Usage', orderable: false, render: function(data, type, row) {
    return `${row.used_count}/${row.quantity}`
  }},
  { data: null, title: 'Period', render: function(data, type, row) {
    return `${formatDate(row.start_date)} - ${formatDate(row.end_date)}`
  }},
  { data: null, title: 'Status', orderable: false, render: function(data, type, row) {
    const active = isPromoActive(row.start_date, row.end_date)
    const cls = active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
    const label = active ? 'Active' : 'Inactive'
    return `<span class="px-2 py-1 text-xs font-medium rounded-full ${cls}">${label}</span>`
  }},
  { data: null, title: 'Actions', orderable: false, searchable: false, render: '#promo-actions' }
]

const tableOptions = {
  pageLength: 25,
  lengthMenu: [10, 25, 50, 100],
  order: [[0, 'asc']],
  columnDefs: [
    { orderable: false, targets: [2, 3, 4, 6, 7] },
    { searchable: false, targets: [7] }
  ]
}

const tableAjax = createServerSideAjax({
  fetchPage: (params) => api.get('/admin/promos', { params }),
  orderColumns: { 0: 'code', 1: 'name', 5: 'start_date' },
  onError: (e) => {
    error.value = e.response?.data?.error || 'Failed to load promos'
  }
})

const resetForm = () => {
  form.value = { code: '', name: '', description: '', discount_type: 'percentage', discount_value: 10, min_purchase: 0, max_discount: 0, quantity: 100, max_usage_per_user: 1, can_combine_flash_sale: false, start_date: '', end_date: '' }
  editingId.value = null
  formError.value = null
}

const editPromo = (promo) => {
  editingId.value = promo.id
  form.value = {
    code: promo.code,
    name: promo.name,
    description: promo.description || '',
    discount_type: promo.discount_type,
    discount_value: promo.discount_value,
    min_purchase: promo.min_purchase,
    max_discount: promo.max_discount || 0,
    quantity: promo.quantity,
    max_usage_per_user: promo.max_usage_per_user,
    can_combine_flash_sale: promo.can_combine_flash_sale,
    start_date: promo.start_date ? new Date(promo.start_date).toISOString().slice(0, 16) : '',
    end_date: promo.end_date ? new Date(promo.end_date).toISOString().slice(0, 16) : ''
  }
  showModal.value = true
}

const savePromo = async () => {
  saving.value = true
  formError.value = null
  try {
    const payload = {
      ...form.value,
      code: form.value.code.toUpperCase(),
      start_date: new Date(form.value.start_date).toISOString(),
      end_date: new Date(form.value.end_date).toISOString()
    }
    if (editingId.value) {
      await api.put(`/admin/promos/${editingId.value}`, payload)
    } else {
      await api.post('/admin/promos', payload)
    }
    showModal.value = false
    table.value?.reload()
  } catch (e) {
    formError.value = e.response?.data?.error || 'Failed to save promo'
  } finally {
    saving.value = false
  }
}

const deletePromo = async (id) => {
  if (!confirm('Are you sure you want to delete this promo?')) return
  try {
    await api.delete(`/admin/promos/${id}`)
    table.value?.reload()
  } catch (e) {
    alert(e.response?.data?.error || 'Failed to delete promo')
  }
}
</script>
