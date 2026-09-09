<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h2 class="text-2xl font-bold text-gray-800">Users</h2>
        <p class="text-gray-500">Manage user accounts, roles and access status</p>
      </div>
      <button
        @click="openCreate"
        class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center"
      >
        <svg class="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Add User
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

    <div v-if="showModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" role="dialog" aria-modal="true" aria-labelledby="user-modal-title">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-md mx-4">
        <div class="px-6 py-4 border-b">
          <h3 id="user-modal-title" class="text-lg font-semibold text-gray-800">{{ editingId ? 'Edit User' : 'Add User' }}</h3>
        </div>
        <form @submit.prevent="saveUser" class="px-6 py-4 space-y-4">
          <div>
            <label for="user-email" class="block text-sm font-medium text-gray-700 mb-1">Email *</label>
            <input
              v-model="form.email"
              id="user-email"
              type="email"
              required
              class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="user@example.com"
            />
          </div>
          <div v-if="!editingId">
            <label for="user-password" class="block text-sm font-medium text-gray-700 mb-1">Password *</label>
            <input
              v-model="form.password"
              id="user-password"
              type="password"
              required
              minlength="8"
              class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="Min. 8 characters"
            />
          </div>
          <div>
            <label for="user-role" class="block text-sm font-medium text-gray-700 mb-1">Role *</label>
            <select
              v-model="form.role"
              id="user-role"
              required
              class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option v-for="role in roles" :key="role" :value="role">{{ role.charAt(0).toUpperCase() + role.slice(1) }}</option>
            </select>
          </div>
          <div v-if="editingId">
            <label class="flex items-center gap-2 text-sm text-gray-700">
              <input
                v-model="form.isActive"
                type="checkbox"
                class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              Active (unchecking will revoke all sessions)
            </label>
          </div>
          <div v-if="formError" class="bg-red-50 border border-red-200 rounded-lg p-3">
            <p class="text-red-600 text-sm">{{ formError }}</p>
          </div>
          <div class="flex justify-end gap-2 pt-2">
            <button type="button" @click="showModal = false" class="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
            <button
              type="submit"
              :disabled="saving"
              class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
            >
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import AdminDataTable from '../../components/AdminDataTable.vue'
import { createServerSideAjax, escapeHtml } from '../../utils/datatables'
import { formatDate } from '../../utils/format'
import { userAdminApi } from '../../api'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const roles = ['tenant', 'staff', 'admin']

const table = ref(null)
const error = ref(null)
const showModal = ref(false)
const saving = ref(false)
const formError = ref(null)
const editingId = ref(null)
const editingOriginal = ref(null)
const form = ref({ email: '', password: '', role: 'tenant', isActive: true })

const columns = [
  { data: 'email', title: 'Email', render: (data, type, row) => {
    if (type !== 'display') return data
    const badge = row.id === auth.userId ? '<span class="ml-2 px-2 py-0.5 bg-blue-100 text-blue-700 rounded text-xs font-medium">You</span>' : ''
    return `<span class="font-medium text-gray-800">${escapeHtml(data)}</span>${badge}`
  }},
  { data: 'role', title: 'Role', orderable: false, render: (data, type) => {
    if (type !== 'display') return data
    return `<span class="px-2 py-1 bg-gray-100 rounded text-xs font-medium text-gray-700">${escapeHtml(String(data).charAt(0).toUpperCase() + String(data).slice(1))}</span>`
  }},
  { data: 'is_active', title: 'Status', orderable: false, render: (data, type) => {
    if (type !== 'display') return data
    return data
      ? '<span class="px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-700">Active</span>'
      : '<span class="px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-700">Inactive</span>'
  }},
  { data: 'created_at', title: 'Created', render: (data, type) => type === 'display' ? formatDate(data) : data },
  { data: null, title: 'Actions', orderable: false, searchable: false, defaultContent: '', render: (data, type, row) => {
    if (type !== 'display' || !row) return ''
    if (row.id === auth.userId) return '<div class="flex justify-end"><span class="text-xs text-gray-400">Current account</span></div>'
    return `
      <div class="flex items-center gap-2 justify-end">
        <button type="button" data-table-action="edit" class="text-sm font-medium text-blue-600 hover:text-blue-900">Edit</button>
        <button type="button" data-table-action="status" class="text-sm font-medium ${row.is_active ? 'text-red-600 hover:text-red-900' : 'text-green-600 hover:text-green-900'}">${row.is_active ? 'Deactivate' : 'Activate'}</button>
        <button type="button" data-table-action="delete" class="text-sm font-medium text-red-600 hover:text-red-900">Delete</button>
      </div>`
  }}
]

const tableOptions = {
  pageLength: 10,
  order: [[3, 'desc']],
  searching: false,
  ordering: false,
  columnDefs: [{ orderable: false, searchable: false, targets: [0, 1, 2, 4] }]
}

const tableAjax = createServerSideAjax({
  fetchPage: (params) => userAdminApi.list({ params: { ...params, page: params.page + 1 } }),
  orderColumns: {},
  defaultOrderBy: 'created_at',
  defaultOrderDir: 'desc',
  onError: (e) => {
    error.value = e.response?.data?.error || 'Failed to load users'
  },
  onSuccess: () => { error.value = null }
})

const openCreate = () => {
  editingId.value = null
  editingOriginal.value = null
  form.value = { email: '', password: '', role: 'tenant', isActive: true }
  formError.value = null
  showModal.value = true
}

const openEdit = (user) => {
  editingId.value = user.id
  editingOriginal.value = { email: user.email, role: user.role, isActive: user.is_active }
  form.value = { email: user.email, password: '', role: user.role, isActive: user.is_active }
  formError.value = null
  showModal.value = true
}

const saveUser = async () => {
  saving.value = true
  formError.value = null
  try {
    if (!editingId.value) {
      await userAdminApi.create({
        email: form.value.email,
        password: form.value.password,
        role: form.value.role
      })
    } else {
      const orig = editingOriginal.value
      if (form.value.email !== orig.email) {
        await userAdminApi.updateEmail(editingId.value, form.value.email)
      }
      if (form.value.role !== orig.role) {
        await userAdminApi.updateRole(editingId.value, form.value.role)
      }
      if (form.value.isActive !== orig.isActive) {
        await userAdminApi.updateStatus(editingId.value, form.value.isActive)
      }
    }
    showModal.value = false
    table.value?.reload()
  } catch (e) {
    formError.value = e.response?.data?.error || 'Failed to save user'
  } finally {
    saving.value = false
  }
}

const deleteUser = async (user) => {
  if (!confirm(`Are you sure you want to delete ${user.email}? This cannot be undone.`)) return
  try {
    await userAdminApi.delete(user.id)
    table.value?.reload()
  } catch (e) {
    error.value = e.response?.data?.error || 'Failed to delete user'
  }
}

const changeStatus = async (user) => {
  const verb = user.is_active ? 'deactivate' : 'activate'
  if (!confirm(`Are you sure you want to ${verb} ${user.email}?`)) return
  try {
    await userAdminApi.updateStatus(user.id, !user.is_active)
    table.value?.reload()
  } catch (e) {
    error.value = e.response?.data?.error || `Failed to ${verb} user`
  }
}

const handleTableAction = ({ action, row }) => {
  if (action === 'edit') openEdit(row)
  if (action === 'status') changeStatus(row)
  if (action === 'delete') deleteUser(row)
}
</script>
