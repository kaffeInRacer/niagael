import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { authApi } from '../api'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const initialized = ref(false)
  const loading = ref(false)
  const error = ref(null)

  const userId = computed(() => user.value?.id || '')
  const roles = computed(() => user.value?.role ? [user.value.role] : [])
  
  const isAuthenticated = computed(() => Boolean(user.value && userId.value))
  const isAdmin = computed(() => roles.value.includes('admin'))
  const isStaff = computed(() => roles.value.includes('staff'))
  const canAccessAdmin = computed(() => isAdmin.value || isStaff.value)

  async function initialize() {
    if (initialized.value) return
    
    try {
      const response = await authApi.me()
      user.value = response.data
    } catch {
      user.value = null
    } finally {
      initialized.value = true
    }
  }

  async function login(credentials) {
    loading.value = true
    error.value = null
    try {
      await authApi.login(credentials)
      
      const me = await authApi.me()
      user.value = me.data
      initialized.value = true
      return user.value
    } catch (e) {
      user.value = null
      error.value = e.response?.data?.error || 'Failed to login'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function register(data) {
    loading.value = true
    error.value = null
    try {
      await authApi.register(data)
      return login(data)
    } catch (e) {
      error.value = e.response?.data?.error || 'Failed to register'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch {
    } finally {
      user.value = null
      initialized.value = true
    }
  }

  return {
    user,
    initialized,
    loading,
    error,
    userId,
    roles,
    isAuthenticated,
    isAdmin,
    isStaff,
    canAccessAdmin,
    initialize,
    login,
    register,
    logout
  }
})
