<template>
  <div class="min-h-[calc(100vh-4rem)] flex items-center justify-center px-4 py-12">
    <div class="w-full max-w-md bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
      <div class="text-center">
        <span class="w-12 h-12 mx-auto bg-blue-600 rounded-xl flex items-center justify-center text-white font-bold text-2xl">E</span>
        <h1 class="mt-4 text-2xl font-bold text-gray-900">Create account</h1>
        <p class="mt-1 text-sm text-gray-500">Register as a tenant buyer</p>
      </div>

      <form @submit.prevent="submit" class="mt-8 space-y-5">
        <div>
          <label class="block text-sm font-medium text-gray-700">Email</label>
          <input v-model.trim="form.email" type="email" required autocomplete="email" class="mt-1.5 w-full px-3 py-2.5 border border-gray-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">Password</label>
          <input v-model="form.password" type="password" required minlength="8" autocomplete="new-password" class="mt-1.5 w-full px-3 py-2.5 border border-gray-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
          <p class="mt-1 text-xs text-gray-400">Minimum 8 characters</p>
        </div>

        <p v-if="error" class="text-sm text-red-500">{{ error }}</p>

        <button type="submit" :disabled="auth.loading" class="w-full bg-blue-600 text-white py-3 rounded-xl font-semibold hover:bg-blue-700 disabled:bg-gray-300 disabled:text-gray-500 disabled:cursor-not-allowed">
          {{ auth.loading ? 'Creating...' : 'Create Account' }}
        </button>
      </form>

      <p class="mt-6 text-sm text-center text-gray-500">
        Already registered?
        <router-link to="/login" class="font-medium text-blue-600 hover:underline">Sign In</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const error = ref(null)
const form = ref({ email: '', password: '' })

const submit = async () => {
  error.value = null
  try {
    await auth.register(form.value)
    router.replace('/')
  } catch (e) {
    error.value = e.response?.data?.error || 'Failed to register'
  }
}
</script>
