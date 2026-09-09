<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h2 class="text-2xl font-bold text-gray-800">Categories</h2>
        <p class="text-gray-500">Manage your product categories</p>
      </div>
      <button
        @click="showModal = true; resetForm()"
        class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center"
      >
        <svg class="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Add Category
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

    <ModalDialog :show="showModal" :title="editingId ? 'Edit Category' : 'Add Category'" aria-id="category-modal-title" @close="showModal = false">
        <form @submit.prevent="saveCategory" class="px-6 py-4 space-y-4">
          <div>
            <label for="category-name" class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
            <input
              v-model="form.name"
              id="category-name"
              type="text"
              required
              class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="Category name"
            />
          </div>
          <div>
            <label for="category-description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea
              v-model="form.description"
              id="category-description"
              rows="3"
              class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="Optional description"
            ></textarea>
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
import { ref } from 'vue'
import api from '../../api'
import AdminDataTable from '../../components/AdminDataTable.vue'
import ModalDialog from '../../components/ModalDialog.vue'
import { createServerSideAjax, escapeHtml } from '../../utils/datatables'
import { formatDate } from '../../utils/format'

const error = ref(null)
const showModal = ref(false)
const saving = ref(false)
const formError = ref(null)
const editingId = ref(null)
const table = ref(null)

const form = ref({
  name: '',
  description: ''
})

const columns = [
  { data: 'name', title: 'Name', render: function(data, type, row) {
    let html = `<div class="text-sm font-medium text-gray-900">${escapeHtml(data)}</div>`
    if (row.description) {
      html += `<div class="text-sm text-gray-500">${escapeHtml(row.description)}</div>`
    }
    return html
  }},
  { data: 'slug', title: 'Slug', render: escapeHtml },
  { data: 'product_count', title: 'Products', orderable: false, render: function(data) {
    return `<span class="px-2 py-1 bg-gray-100 rounded-full">${escapeHtml(data ?? 0)}</span>`
  }},
  { data: 'is_active', title: 'Status', orderable: false, render: function(data) {
    const cls = data ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
    const label = data ? 'Active' : 'Inactive'
    return `<span class="px-2 py-1 text-xs font-medium rounded-full ${cls}">${label}</span>`
  }},
  { data: 'created_at', title: 'Created', render: function(data) {
    return formatDate(data)
  }},
  { data: null, title: 'Actions', orderable: false, searchable: false, render: function(data, type, row) {
    if (!row) return ''
    return '<div class="flex justify-end gap-3"><button type="button" class="text-blue-600 hover:text-blue-900" data-table-action="edit">Edit</button><button type="button" class="text-red-600 hover:text-red-900" data-table-action="delete">Delete</button></div>'
  }}
]

const tableOptions = {
  pageLength: 25,
  lengthMenu: [10, 25, 50, 100],
  order: [[0, 'asc']],
  columnDefs: [
    { orderable: false, targets: [2, 3, 5] },
    { searchable: false, targets: [5] }
  ]
}

const tableAjax = createServerSideAjax({
  fetchPage: (params) => api.get('/admin/categories', {
    params: { ...params, include_product_count: true }
  }),
  orderColumns: { 0: 'name', 1: 'slug', 4: 'created_at' },
  defaultOrderBy: 'name',
  defaultOrderDir: 'asc',
  onError: (e) => {
    error.value = e.response?.data?.error || 'Failed to load categories'
  },
  onSuccess: () => { error.value = null }
})

const resetForm = () => {
  form.value = { name: '', description: '' }
  editingId.value = null
  formError.value = null
}

const editCategory = (category) => {
  editingId.value = category.id
  form.value = {
    name: category.name,
    description: category.description || ''
  }
  showModal.value = true
}

const saveCategory = async () => {
  saving.value = true
  formError.value = null
  try {
    if (editingId.value) {
      await api.put(`/admin/categories/${editingId.value}`, form.value)
    } else {
      await api.post('/admin/categories', form.value)
    }
    showModal.value = false
    table.value?.reload()
  } catch (e) {
    formError.value = e.response?.data?.error || e.response?.data?.errors || 'Failed to save category'
  } finally {
    saving.value = false
  }
}

const deleteCategory = async (id) => {
  if (!confirm('Are you sure you want to delete this category?')) return
  try {
    await api.delete(`/admin/categories/${id}`)
    table.value?.reload()
  } catch (e) {
    alert(e.response?.data?.error || 'Failed to delete category')
  }
}

const handleTableAction = ({ action, row }) => {
  if (action === 'edit') editCategory(row)
  if (action === 'delete') deleteCategory(row.id)
}
</script>
