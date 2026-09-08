<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h2 class="text-2xl font-bold text-gray-800">Orders</h2>
        <p class="text-gray-500">Track and manage customer orders</p>
      </div>
    </div>

    <div class="bg-white rounded-lg shadow p-4 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <select
          v-model="filters.status"
          class="border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          @change="reloadTable(true)"
        >
          <option value="">All Status</option>
          <option value="pending">Pending</option>
          <option value="paid">Paid</option>
          <option value="shipped">Shipped</option>
          <option value="delivered">Delivered</option>
          <option value="cancelled">Cancelled</option>
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
      >
      </AdminDataTable>

    </div>

    <div v-if="selectedOrder" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-2xl mx-4 max-h-[80vh] overflow-y-auto">
        <div class="px-6 py-4 border-b flex justify-between items-center">
          <h3 class="text-lg font-semibold text-gray-800">Order Details</h3>
          <button @click="selectedOrder = null" class="text-gray-400 hover:text-gray-600">
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="px-6 py-4 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-sm text-gray-500">Order ID</p>
              <p class="font-mono text-sm">{{ selectedOrder.id }}</p>
            </div>
            <div>
              <p class="text-sm text-gray-500">Status</p>
              <span class="px-2 py-1 text-xs font-medium rounded-full" :class="getStatusClass(selectedOrder.status)">
                {{ selectedOrder.status }}
              </span>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-sm text-gray-500">User ID</p>
              <p class="font-mono text-sm">{{ selectedOrder.user_id }}</p>
            </div>
            <div>
              <p class="text-sm text-gray-500">Address ID</p>
              <p class="font-mono text-sm">{{ selectedOrder.address_id }}</p>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-sm text-gray-500">Snap Token</p>
              <p class="font-mono text-sm text-break">{{ selectedOrder.snap_token || '-' }}</p>
            </div>
            <div>
              <p class="text-sm text-gray-500">Expire Time</p>
              <p class="text-sm">{{ selectedOrder.expire_time ? formatDate(selectedOrder.expire_time) : '-' }}</p>
            </div>
          </div>
          <div class="border-t pt-4 flex justify-between">
            <p class="font-semibold">Total Amount</p>
            <p class="font-bold text-lg">{{ formatPrice(selectedOrder.total_amount) }}</p>
          </div>
          <div class="text-xs text-gray-400">
            Created: {{ formatDate(selectedOrder.created_at) }}
            <span v-if="selectedOrder.updated_at"> | Updated: {{ formatDate(selectedOrder.updated_at) }}</span>
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
import { createServerSideAjax } from '../../utils/datatables'
import { formatDate, formatPrice, getStatusClass } from '../../utils/format'

const error = ref(null)
const selectedOrder = ref(null)
const table = ref(null)

const filters = ref({
  status: ''
})

const columns = [
  { data: 'id', title: 'Order ID', orderable: false, render: function(data) {
    return `<span class="font-mono text-sm">${data?.slice(0, 8)}...</span>`
  }},
  { data: 'user_id', title: 'User', orderable: false, render: function(data) {
    return `<span class="font-mono text-sm">${data?.slice(0, 12)}...</span>`
  }},
  { data: 'total_amount', title: 'Total', render: function(data) {
    return `<span class="font-medium">${formatPrice(data)}</span>`
  }},
  { data: 'status', title: 'Status', orderable: false, render: function(data) {
    const cls = getStatusClass(data)
    return `<span class="px-2 py-1 text-xs font-medium rounded-full ${cls}">${data}</span>`
  }},
  { data: 'snap_token', title: 'Payment', orderable: false, render: function(data) {
    return data ? '<span class="text-green-600">Has Token</span>' : '<span class="text-gray-400">No Token</span>'
  }},
  { data: 'created_at', title: 'Date', render: function(data) {
    return formatDate(data)
  }},
  { data: null, title: 'Actions', orderable: false, searchable: false, render: function(data, type, row) {
    if (!row) return ''
    return `<div class="flex justify-end"><button class="text-blue-600 hover:text-blue-900 view-btn" data-id="${row.id}">View</button></div>`
  }}
]

const tableOptions = {
  pageLength: 10,
  order: [[5, 'desc']],
  columnDefs: [
    { orderable: false, targets: [0, 1, 3, 4, 6] },
    { searchable: false, targets: [6] }
  ]
}

const tableAjax = createServerSideAjax({
  fetchPage: (params) => api.get('/orders', { params }),
  orderColumns: { 2: 'total_amount', 5: 'created_at' },
  getFilters: () => ({ status: filters.value.status || undefined }),
  onError: (e) => {
    error.value = e.response?.data?.error || 'Failed to load orders'
  }
})

const reloadTable = (resetPaging = false) => {
  error.value = null
  table.value?.reload(resetPaging)
}

const viewOrder = (order) => {
  selectedOrder.value = order
}

onMounted(() => {
  const tableEl = document.querySelector('.dataTable')
  if (tableEl) {
    tableEl.addEventListener('click', (e) => {
      const viewBtn = e.target.closest('.view-btn')
      if (viewBtn) {
        const row = table.value?.getInstance()?.row(viewBtn.closest('tr'))?.data()
        if (row) viewOrder(row)
      }
    })
  }
})
</script>
