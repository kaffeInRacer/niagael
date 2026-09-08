<template>
  <div>
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <div v-for="stat in stats" :key="stat.name" class="bg-white rounded-lg shadow p-6">
        <div class="flex items-center">
          <div class="p-3 rounded-full" :class="stat.bgColor">
            <component :is="stat.icon" class="w-6 h-6" :class="stat.iconColor" />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">{{ stat.name }}</p>
            <p class="text-2xl font-semibold text-gray-900">{{ stat.value }}</p>
          </div>
        </div>
      </div>
    </div>

    <div class="bg-white rounded-lg shadow p-6">
      <h2 class="text-lg font-semibold text-gray-800 mb-4">Quick Actions</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <router-link
          to="/admin/products"
          class="flex items-center p-4 border rounded-lg hover:bg-gray-50 transition-colors"
        >
          <div class="p-2 bg-blue-100 rounded-lg">
            <svg class="w-6 h-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
          </div>
          <div class="ml-4">
            <p class="font-medium text-gray-800">Manage Products</p>
            <p class="text-sm text-gray-500">Add, edit, or delete products</p>
          </div>
        </router-link>

        <router-link
          to="/admin/categories"
          class="flex items-center p-4 border rounded-lg hover:bg-gray-50 transition-colors"
        >
          <div class="p-2 bg-green-100 rounded-lg">
            <svg class="w-6 h-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          </div>
          <div class="ml-4">
            <p class="font-medium text-gray-800">Manage Categories</p>
            <p class="text-sm text-gray-500">Organize product categories</p>
          </div>
        </router-link>

        <router-link
          to="/admin/orders"
          class="flex items-center p-4 border rounded-lg hover:bg-gray-50 transition-colors"
        >
          <div class="p-2 bg-purple-100 rounded-lg">
            <svg class="w-6 h-6 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
            </svg>
          </div>
          <div class="ml-4">
            <p class="font-medium text-gray-800">View Orders</p>
            <p class="text-sm text-gray-500">Track and manage orders</p>
          </div>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { h } from 'vue'

const stats = [
  {
    name: 'Total Products',
    value: '-',
    bgColor: 'bg-blue-100',
    iconColor: 'text-blue-600',
    icon: {
      render() {
        return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor' }, [
          h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4' })
        ])
      }
    }
  },
  {
    name: 'Total Categories',
    value: '-',
    bgColor: 'bg-green-100',
    iconColor: 'text-green-600',
    icon: {
      render() {
        return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor' }, [
          h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10' })
        ])
      }
    }
  },
  {
    name: 'Active Flash Sales',
    value: '-',
    bgColor: 'bg-yellow-100',
    iconColor: 'text-yellow-600',
    icon: {
      render() {
        return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor' }, [
          h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M13 10V3L4 14h7v7l9-11h-7z' })
        ])
      }
    }
  },
  {
    name: 'Total Orders',
    value: '-',
    bgColor: 'bg-purple-100',
    iconColor: 'text-purple-600',
    icon: {
      render() {
        return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor' }, [
          h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01' })
        ])
      }
    }
  }
]
</script>
