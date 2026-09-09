<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h2 class="text-2xl font-bold text-gray-800">Roles & Permissions</h2>
        <p class="text-gray-500">Configure RBAC policies for each role across all services</p>
      </div>
    </div>

    <div v-if="loading" class="flex justify-center items-center py-12" role="status" aria-label="Loading permissions">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
      <p class="text-red-600" role="alert">{{ error }}</p>
    </div>

    <template v-else>
      <div class="bg-white rounded-lg shadow mb-6">
        <div class="border-b border-gray-200">
          <nav class="flex -mb-px">
            <button
              v-for="role in roles"
              :key="role"
              @click="selectedRole = role"
              class="px-6 py-3 text-sm font-medium border-b-2 transition-colors"
              :class="selectedRole === role ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'"
              :aria-pressed="selectedRole === role"
            >
              {{ role.charAt(0).toUpperCase() + role.slice(1) }}
            </button>
          </nav>
        </div>

        <div class="p-6">
          <div v-if="mutationError" class="bg-red-50 border border-red-200 rounded-lg p-3 mb-4" role="alert">{{ mutationError }}</div>
          <div v-for="(resources, serviceName) in services" :key="serviceName" class="mb-8 last:mb-0">
            <h3 class="text-lg font-semibold text-gray-800 mb-4 flex items-center">
              <span class="px-2 py-1 bg-gray-100 rounded text-sm font-mono mr-2">{{ serviceNames[serviceName] || serviceName }}</span>
            </h3>
            
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                  <tr>
                    <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Resource</th>
                    <th class="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Create</th>
                    <th class="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Read</th>
                    <th class="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Update</th>
                    <th class="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Delete</th>
                    <th class="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Apply</th>
                  </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                  <tr v-for="resource in resources" :key="resource.name" class="hover:bg-gray-50">
                    <td class="px-4 py-3 whitespace-nowrap">
                      <span class="font-mono text-sm text-gray-800">{{ resource.name }}</span>
                    </td>
                    <td v-for="action in actions" :key="action" class="px-4 py-3 text-center">
                       <label class="relative inline-flex items-center cursor-pointer" v-if="isActionAvailable(resource, action)">
                        <input
                          type="checkbox"
                          :checked="hasPermission(selectedRole, serviceName, resource.name, action)"
                          :disabled="isPending(selectedRole, serviceName, resource.name, action)"
                          :aria-label="`${selectedRole} ${action} ${resource.name}`"
                          @change="togglePermission(selectedRole, serviceName, resource.name, action, $event.target.checked)"
                          class="sr-only peer"
                        />
                        <div class="w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-blue-600"></div>
                      </label>
                      <span v-else class="text-gray-300">-</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-lg shadow p-6">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">Permission Summary</h3>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div v-for="role in roles" :key="role" class="border rounded-lg p-4">
            <h4 class="font-medium text-gray-800 mb-2">{{ role.charAt(0).toUpperCase() + role.slice(1) }}</h4>
            <div class="text-sm text-gray-600">
              <p>Total Permissions: <span class="font-semibold">{{ getPermissionCount(role) }}</span></p>
              <p>Services: <span class="font-semibold">{{ getRoleServiceCount(role) }}</span></p>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { rbacApi } from '../../api'

const loading = ref(true)
const error = ref(null)
const mutationError = ref(null)
const policies = ref([])
const resources = ref({})
const pending = ref(new Set())
const selectedRole = ref('admin')

const roles = ['admin', 'staff', 'tenant']
const actions = ['create', 'read', 'update', 'delete', 'apply']

const fallbackCapabilities = {
  product: {
    categories: ['create', 'read', 'update', 'delete'],
    products: ['create', 'read', 'update', 'delete'],
    variants: ['create', 'read', 'update', 'delete'],
    'product-images': ['create', 'read', 'delete']
  },
  'dynamic-pricing': {
    'flash-sales': ['create', 'read', 'update', 'delete'],
    promos: ['create', 'read', 'update', 'delete', 'apply']
  },
  order: {
    orders: ['create', 'read', 'update'],
    carts: ['create', 'read', 'update', 'delete'],
    addresses: ['create', 'read', 'update', 'delete'],
    payments: ['create']
  },
  auth: {
    users: ['read', 'update'],
    policies: ['create', 'read', 'delete']
  }
}

const serviceNames = {
  'product': 'Product Service',
  'dynamic-pricing': 'Dynamic Pricing',
  'order': 'Order Service',
  'auth': 'Auth Service'
}

const services = computed(() => {
  const result = structuredClone(fallbackCapabilities)
  for (const [serviceName, serviceResources] of Object.entries(resources.value || {})) {
    result[serviceName] ||= {}
    if (Array.isArray(serviceResources)) {
      for (const resource of serviceResources) {
        if (typeof resource === 'string') {
          result[serviceName][resource] ||= []
        } else if (resource?.name && Array.isArray(resource.actions)) {
          result[serviceName][resource.name] = resource.actions
        }
      }
    } else {
      for (const [resource, resourceActions] of Object.entries(serviceResources || {})) {
        if (Array.isArray(resourceActions)) result[serviceName][resource] = resourceActions
      }
    }
  }
  return Object.fromEntries(Object.entries(result).map(([service, entries]) => [
    service,
    Object.entries(entries)
      .map(([name, resourceActions]) => ({ name, actions: resourceActions }))
      .sort((a, b) => a.name.localeCompare(b.name))
  ]))
})

const loadData = async () => {
  loading.value = true
  error.value = null
  
  try {
    const [policiesRes, resourcesRes] = await Promise.all([
      rbacApi.listPolicies(),
      rbacApi.getResources()
    ])
    
    policies.value = policiesRes.data.data || []
    resources.value = resourcesRes.data.data || {}
  } catch (e) {
    error.value = e.response?.data?.error || 'Failed to load RBAC data'
    console.error('Failed to load RBAC data:', e)
  } finally {
    loading.value = false
  }
}

const hasPermission = (role, service, resource, action) => {
  return policies.value.some(p => 
    p.role === role && 
    p.service === service && 
    p.resource === resource && 
    p.action === action
  )
}

const isActionAvailable = (resource, action) => resource.actions.includes(action)

const permissionKey = (role, service, resource, action) => `${role}:${service}:${resource}:${action}`
const isPending = (role, service, resource, action) => pending.value.has(permissionKey(role, service, resource, action))

const togglePermission = async (role, service, resource, action, checked) => {
  const key = permissionKey(role, service, resource, action)
  if (pending.value.has(key)) return
  mutationError.value = null
  pending.value = new Set(pending.value).add(key)
  const policy = { service, role, resource, action }
  policies.value = checked
    ? [...policies.value, policy]
    : policies.value.filter(p => permissionKey(p.role, p.service, p.resource, p.action) !== key)
  try {
    if (checked) {
      await rbacApi.addPolicy(policy)
    } else {
      await rbacApi.deletePolicy(policy)
    }
  } catch (e) {
    if (checked) {
      policies.value = policies.value.filter(p => permissionKey(p.role, p.service, p.resource, p.action) !== key)
    } else if (!hasPermission(role, service, resource, action)) {
      policies.value = [...policies.value, policy]
    }
    mutationError.value = e.response?.data?.error || 'Failed to update permission'
  } finally {
    const next = new Set(pending.value)
    next.delete(key)
    pending.value = next
  }
}

const getPermissionCount = (role) => {
  return policies.value.filter(p => p.role === role).length
}

const getRoleServiceCount = (role) => {
  const services = new Set(policies.value.filter(p => p.role === role).map(p => p.service))
  return services.size
}

onMounted(loadData)
</script>
