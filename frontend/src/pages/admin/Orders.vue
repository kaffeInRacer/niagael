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
          <option value="processing">Processing</option>
          <option value="shipped">Shipped</option>
          <option value="delivered">Delivered</option>
          <option value="cancelled">Cancelled</option>
          <option value="refunded">Refunded</option>
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

    <ModalDialog :show="!!selectedOrder" title="Order Details" aria-id="order-modal-title" size="xl" scroll-body @close="selectedOrder = null">
        <div class="px-6 py-4 space-y-4" v-if="selectedOrder">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-sm text-gray-500">Ref Order</p>
              <p class="font-mono text-sm font-semibold">{{ selectedOrder.order_ref || selectedOrder.id }}</p>
            </div>
            <div>
              <p class="text-sm text-gray-500">Buyer</p>
              <p class="text-sm font-medium">{{ buyerName(selectedOrder.user_id, selectedOrder) }}</p>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-sm text-gray-500">Status</p>
              <div class="flex items-center gap-2">
                <span class="px-2 py-1 text-xs font-medium rounded-full" :class="getStatusClass(selectedOrder.status)">
                  {{ selectedOrder.status }}
                </span>
                <button
                  type="button"
                  @click="selectedOrder.updated_at = null; focusStatus()"
                  class="text-xs text-blue-600 hover:text-blue-800 underline"
                >Change</button>
              </div>
            </div>
            <div>
              <p class="text-sm text-gray-500">Payment</p>
              <p class="text-sm">{{ selectedOrder.snap_token ? 'Midtrans (token tersimpan)' : 'Belum ada token' }}</p>
            </div>
          </div>

          <div class="bg-blue-50 border border-blue-100 rounded-lg p-4" id="status-changer">
            <h4 class="text-sm font-semibold text-gray-700 mb-3">Ubah Status</h4>
            <div class="flex items-center gap-3 flex-wrap">
              <select
                v-model="newStatus"
                class="border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                aria-label="New order status"
              >
                <option value="">Pilih status baru...</option>
                <option v-for="s in allowedTargets" :key="s" :value="s" class="capitalize">{{ s }}</option>
              </select>
              <button
                type="button"
                @click="applyStatus"
                :disabled="!newStatus || updatingStatus"
                class="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {{ updatingStatus ? 'Updating...' : 'Update Status' }}
              </button>
              <p v-if="statusMessage" class="text-sm" :class="statusError ? 'text-red-600' : 'text-green-600'">{{ statusMessage }}</p>
            </div>
            <p class="text-xs text-gray-400 mt-2">Transisi diizinkan backend: pending/paid→processing, processing→shipped/cancelled/refunded, shipped→delivered/refunded.</p>
          </div>

          <div class="border-t pt-4">
            <h4 class="text-sm font-semibold text-gray-700 mb-3">Order Items ({{ selectedItems.length }})</h4>
            <div v-if="itemsLoading" class="flex justify-center py-6" role="status" aria-label="Loading items">
              <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
            </div>
            <div v-else-if="selectedItems.length > 0" class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                  <tr>
                    <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Product</th>
                    <th class="px-3 py-2 text-center text-xs font-medium text-gray-500 uppercase">Qty</th>
                    <th class="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Price</th>
                    <th class="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Final</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200">
                  <tr v-for="item in selectedItems" :key="item.id">
                    <td class="px-3 py-3">
                      <p class="text-sm font-medium text-gray-800">{{ item.product_name }}</p>
                      <p v-if="item.variant_name" class="text-xs text-gray-500">Variant: {{ item.variant_name }}</p>
                      <p v-if="item.flash_sale_name" class="text-xs text-red-600">
                        Flash Sale: {{ item.flash_sale_name }} (-{{ item.flash_sale_discount_percent }}%)
                        <span v-if="item.flash_sale_quantity"> ({{ item.flash_sale_quantity }} unit)</span>
                      </p>
                      <p v-if="item.promo_code" class="text-xs text-green-600">
                        Voucher: {{ item.promo_code }} ({{ item.promo_name }} - {{ formatPrice(item.promo_discount_amount) }})
                      </p>
                    </td>
                    <td class="px-3 py-3 text-center text-sm">{{ item.quantity }}</td>
                    <td class="px-3 py-3 text-right text-sm text-gray-500">{{ formatPrice(item.product_price) }}</td>
                    <td class="px-3 py-3 text-right text-sm font-semibold">{{ formatPrice(item.final_price) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p v-else class="text-sm text-gray-500">No items found for this order.</p>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-sm text-gray-500">Address ID</p>
              <p class="font-mono text-xs">{{ selectedOrder.address_id }}</p>
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
    </ModalDialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api, { userAdminApi } from '../../api'
import AdminDataTable from '../../components/AdminDataTable.vue'
import ModalDialog from '../../components/ModalDialog.vue'
import { createServerSideAjax, escapeHtml } from '../../utils/datatables'
import { formatDate, formatPrice, getStatusClass } from '../../utils/format'

const error = ref(null)
const selectedOrder = ref(null)
const selectedItems = ref([])
const itemsLoading = ref(false)
const newStatus = ref('')
const updatingStatus = ref(false)
const statusMessage = ref(null)
const statusError = ref(false)
const table = ref(null)
const buyerMap = ref({})

const filters = ref({
  status: ''
})

const allowedFrom = {
  processing: ['pending', 'paid'],
  cancelled: ['pending', 'processing'],
  refunded: ['paid', 'processing', 'shipped'],
  shipped: ['processing'],
  delivered: ['shipped']
}

const allowedTargets = computed(() => {
  const status = selectedOrder.value?.status
  if (!status) return []
  return Object.keys(allowedFrom).filter(target => allowedFrom[target].includes(status))
})

const focusStatus = () => {
  document.getElementById('status-changer')?.scrollIntoView({ behavior: 'smooth' })
}

const buyerName = (userId, row) => (row && row.buyer_email) || buyerMap.value[userId] || `${String(userId ?? '').slice(0, 8)}...`

const loadBuyers = async () => {
  try {
    const map = {}
    let page = 1
    for (let i = 0; i < 5; i++) {
      const { data } = await userAdminApi.list({ page, page_size: 100 })
      for (const u of data.data || []) map[u.id] = u.email
      if ((data.data || []).length < (data.page_size || 100)) break
      page++
    }
    buyerMap.value = map
  } catch {
    buyerMap.value = {}
  }
}

const columns = [
  { data: 'order_ref', title: 'Ref Order', orderable: false, render: function(data) {
    return `<span class="font-mono text-sm font-medium">${escapeHtml(String(data ?? '-'))}</span>`
  }},
  { data: 'user_id', title: 'Buyer', orderable: false, render: (data, type, row) => {
    const name = (row && row.buyer_email) || buyerName(data)
    return `<span class="text-sm">${escapeHtml(name)}</span>`
  }},
  { data: 'total_amount', title: 'Total', render: function(data) {
    return `<span class="font-medium">${formatPrice(data)}</span>`
  }},
  { data: 'status', title: 'Status', orderable: false, render: function(data) {
    const cls = getStatusClass(data)
    return `<span class="px-2 py-1 text-xs font-medium rounded-full ${cls}">${escapeHtml(data)}</span>`
  }},
  { data: 'snap_token', title: 'Payment', orderable: false, render: function(data) {
    return data ? '<span class="text-green-600">Has Token</span>' : '<span class="text-gray-400">No Token</span>'
  }},
  { data: 'created_at', title: 'Date', render: function(data) {
    return formatDate(data)
  }},
  { data: null, title: 'Actions', orderable: false, searchable: false, render: function(data, type, row) {
    if (!row) return ''
    return '<div class="flex justify-end"><button type="button" class="text-blue-600 hover:text-blue-900" data-table-action="view">Details</button></div>'
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
  },
  onSuccess: () => { error.value = null }
})

const reloadTable = (resetPaging = false) => {
  error.value = null
  table.value?.reload(resetPaging)
}

const viewOrder = async (order) => {
  selectedOrder.value = order
  selectedItems.value = []
  itemsLoading.value = true
  newStatus.value = ''
  statusMessage.value = null
  try {
    const { data } = await api.get(`/orders/${order.id}`)
    selectedOrder.value = data.data || order
    selectedItems.value = data.items || []
  } catch (e) {
    selectedItems.value = []
    error.value = e.response?.data?.error || 'Failed to load order details'
  } finally {
    itemsLoading.value = false
  }
}

const applyStatus = async () => {
  if (!newStatus.value || !selectedOrder.value) return
  updatingStatus.value = true
  statusMessage.value = null
  statusError.value = false
  try {
    await api.patch(`/orders/${selectedOrder.value.id}/status`, { status: newStatus.value })
    statusMessage.value = `Status changed to "${newStatus.value}"`
    selectedOrder.value = { ...selectedOrder.value, status: newStatus.value }
    newStatus.value = ''
    table.value?.reload()
  } catch (e) {
    statusError.value = true
    statusMessage.value = e.response?.data?.error || 'Failed to update status'
  } finally {
    updatingStatus.value = false
  }
}

const handleTableAction = ({ action, row }) => {
  if (action === 'view') viewOrder(row)
}

onMounted(() => {
  loadBuyers()
})
</script>
