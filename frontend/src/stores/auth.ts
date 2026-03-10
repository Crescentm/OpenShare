import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AdminInfo } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('admin_token'))
  const expiresAt = ref<number | null>(
    localStorage.getItem('admin_token_expires') 
      ? Number(localStorage.getItem('admin_token_expires')) 
      : null
  )
  const adminInfo = ref<AdminInfo | null>(null)

  const isAuthenticated = computed(() => {
    if (!token.value) return false
    // 检查 token 是否过期
    if (expiresAt.value && Date.now() / 1000 > expiresAt.value) {
      logout()
      return false
    }
    return true
  })
  const isSuperAdmin = computed(() => adminInfo.value?.role === 'super_admin')

  function setToken(newToken: string, expires?: number) {
    token.value = newToken
    localStorage.setItem('admin_token', newToken)
    if (expires) {
      expiresAt.value = expires
      localStorage.setItem('admin_token_expires', String(expires))
    }
  }

  function setAdminInfo(info: AdminInfo) {
    adminInfo.value = info
  }

  function logout() {
    token.value = null
    expiresAt.value = null
    adminInfo.value = null
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_token_expires')
  }

  function hasPermission(permission: string): boolean {
    if (!adminInfo.value) return false
    if (adminInfo.value.role === 'super_admin') return true
    return adminInfo.value.permissions?.includes(permission) ?? false
  }

  return {
    token,
    expiresAt,
    adminInfo,
    isAuthenticated,
    isSuperAdmin,
    setToken,
    setAdminInfo,
    logout,
    hasPermission,
  }
})
