import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AdminInfo } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('admin_token'))
  const adminInfo = ref<AdminInfo | null>(null)

  const isAuthenticated = computed(() => !!token.value)
  const isSuperAdmin = computed(() => adminInfo.value?.role === 'super_admin')

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem('admin_token', newToken)
  }

  function setAdminInfo(info: AdminInfo) {
    adminInfo.value = info
  }

  function logout() {
    token.value = null
    adminInfo.value = null
    localStorage.removeItem('admin_token')
  }

  function hasPermission(permission: string): boolean {
    if (!adminInfo.value) return false
    if (adminInfo.value.role === 'super_admin') return true
    return adminInfo.value.permissions?.includes(permission) ?? false
  }

  return {
    token,
    adminInfo,
    isAuthenticated,
    isSuperAdmin,
    setToken,
    setAdminInfo,
    logout,
    hasPermission,
  }
})
