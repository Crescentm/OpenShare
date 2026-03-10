<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/admin'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const form = ref({
  username: '',
  password: '',
})
const loading = ref(false)
const error = ref('')

const handleLogin = async () => {
  if (!form.value.username || !form.value.password) {
    error.value = '请输入用户名和密码'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const { data } = await authApi.login(form.value)
    
    if (data.data) {
      authStore.setToken(data.data.token, data.data.expires_at)
      authStore.setAdminInfo(data.data.admin)
      
      const redirect = route.query.redirect as string
      router.push(redirect || '/admin')
    }
  } catch (err: unknown) {
    if (err instanceof Error) {
      error.value = err.message || '登录失败，请检查用户名和密码'
    } else {
      error.value = '登录失败，请检查用户名和密码'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100">
    <div class="max-w-md w-full">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-primary-600">OpenShare</h1>
        <p class="text-gray-500 mt-2">管理后台登录</p>
      </div>

      <form @submit.prevent="handleLogin" class="card">
        <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
          {{ error }}
        </div>

        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-2">用户名</label>
          <input
            v-model="form.username"
            type="text"
            class="input"
            placeholder="请输入用户名"
            :disabled="loading"
          />
        </div>

        <div class="mb-6">
          <label class="block text-sm font-medium text-gray-700 mb-2">密码</label>
          <input
            v-model="form.password"
            type="password"
            class="input"
            placeholder="请输入密码"
            :disabled="loading"
          />
        </div>

        <button
          type="submit"
          class="btn btn-primary w-full"
          :disabled="loading"
        >
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>

      <p class="text-center text-sm text-gray-500 mt-4">
        <RouterLink to="/" class="text-primary-600 hover:underline">返回首页</RouterLink>
      </p>
    </div>
  </div>
</template>
