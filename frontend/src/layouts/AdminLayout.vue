<template>
  <div class="min-h-screen bg-gray-100 flex flex-col md:flex-row">
    <div class="w-full md:w-64 md:min-h-screen bg-gray-900 text-white flex flex-col flex-shrink-0">
      <div class="p-4 border-b border-gray-800">
        <router-link to="/admin" class="text-xl font-bold text-white">
          Admin Panel
        </router-link>
      </div>

      <nav class="flex gap-1 overflow-x-auto p-3 md:flex-1 md:block md:space-y-1 md:overflow-visible md:p-4">
        <router-link
          v-for="item in menuItems"
          :key="item.path"
          :to="item.path"
          class="flex flex-shrink-0 items-center px-3 py-2 rounded-lg text-sm font-medium whitespace-nowrap transition-colors"
          :class="isActive(item.path) ? 'bg-gray-800 text-white' : 'text-gray-300 hover:bg-gray-800 hover:text-white'"
        >
          <component :is="item.icon" class="w-5 h-5 mr-3" />
          {{ item.name }}
        </router-link>
      </nav>

      <div class="hidden md:block p-4 border-t border-gray-800">
        <router-link
          to="/"
          class="flex items-center px-3 py-2 rounded-lg text-sm font-medium text-gray-300 hover:bg-gray-800 hover:text-white transition-colors"
        >
          <svg class="w-5 h-5 mr-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          Back to Store
        </router-link>
      </div>
    </div>

    <div class="min-w-0 flex-1 flex flex-col">
      <header class="bg-white shadow-sm h-16 flex items-center px-4 md:px-6">
        <h1 class="text-lg font-semibold text-gray-800">{{ pageTitle }}</h1>
      </header>

      <main class="min-w-0 flex-1 p-4 md:p-6 overflow-x-hidden overflow-y-auto">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { computed, h } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const menuItems = [
  {
    name: 'Dashboard',
    path: '/admin',
    icon: {
      render() {
        return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor' }, [
          h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' })
        ])
      }
    }
  },
  {
    name: 'Categories',
    path: '/admin/categories',
    icon: {
      render() {
        return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor' }, [
          h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10' })
        ])
      }
    }
  },
  {
    name: 'Products',
    path: '/admin/products',
    icon: {
      render() {
        return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor' }, [
          h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4' })
        ])
      }
    }
  },
  {
    name: 'Flash Sales',
    path: '/admin/flash-sales',
    icon: {
      render() {
        return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor' }, [
          h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M13 10V3L4 14h7v7l9-11h-7z' })
        ])
      }
    }
  },
  {
    name: 'Promos',
    path: '/admin/promos',
    icon: {
      render() {
        return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor' }, [
          h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z' })
        ])
      }
    }
  },
  {
    name: 'Orders',
    path: '/admin/orders',
    icon: {
      render() {
        return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor' }, [
          h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01' })
        ])
      }
    }
  },
  {
    name: 'Roles & Permissions',
    path: '/admin/roles',
    icon: {
      render() {
        return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor' }, [
          h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' })
        ])
      }
    }
  }
]

const pageTitle = computed(() => {
  const item = menuItems.find(i => i.path === route.path)
  return item ? item.name : 'Admin'
})

const isActive = (path) => {
  if (path === '/admin') {
    return route.path === '/admin'
  }
  return route.path.startsWith(path)
}
</script>
