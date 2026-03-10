<script setup lang="ts">
import { ref, computed } from 'vue'
import { RouterLink, RouterView, useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const sidebarOpen = ref(true)

const menuItems = [
  { path: '/admin/submissions', name: '审核管理', icon: 'clipboard-check' },
  { path: '/admin/files', name: '资料管理', icon: 'folder' },
  { path: '/admin/tags', name: 'Tag 管理', icon: 'tag' },
  { path: '/admin/reports', name: '举报管理', icon: 'flag' },
  { path: '/admin/announcements', name: '公告管理', icon: 'megaphone' },
  { path: '/admin/admins', name: '管理员管理', icon: 'users' },
  { path: '/admin/settings', name: '系统设置', icon: 'cog' },
  { path: '/admin/logs', name: '操作日志', icon: 'document-text' },
]

const currentPath = computed(() => route.path)

const handleLogout = () => {
  authStore.logout()
  router.push({ name: 'AdminLogin' })
}
</script>

<template>
  <div class="min-h-screen flex bg-gray-100">
    <!-- 侧边栏 -->
    <aside
      :class="[
        'bg-white shadow-lg transition-all duration-300',
        sidebarOpen ? 'w-64' : 'w-16'
      ]"
    >
      <!-- Logo -->
      <div class="h-16 flex items-center justify-center border-b border-gray-200">
        <RouterLink to="/admin" class="flex items-center">
          <span v-if="sidebarOpen" class="text-xl font-bold text-primary-600">OpenShare</span>
          <span v-else class="text-xl font-bold text-primary-600">OS</span>
        </RouterLink>
      </div>

      <!-- 菜单 -->
      <nav class="p-4">
        <ul class="space-y-2">
          <li v-for="item in menuItems" :key="item.path">
            <RouterLink
              :to="item.path"
              :class="[
                'flex items-center px-3 py-2 rounded-lg transition-colors',
                currentPath === item.path
                  ? 'bg-primary-50 text-primary-600'
                  : 'text-gray-600 hover:bg-gray-100'
              ]"
            >
              <span class="w-5 h-5 mr-3">●</span>
              <span v-if="sidebarOpen">{{ item.name }}</span>
            </RouterLink>
          </li>
        </ul>
      </nav>
    </aside>

    <!-- 主内容区 -->
    <div class="flex-1 flex flex-col">
      <!-- 顶部栏 -->
      <header class="h-16 bg-white shadow-sm flex items-center justify-between px-6">
        <button
          @click="sidebarOpen = !sidebarOpen"
          class="p-2 text-gray-600 hover:bg-gray-100 rounded-lg"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>

        <div class="flex items-center space-x-4">
          <RouterLink to="/" class="text-gray-600 hover:text-primary-600">
            返回前台
          </RouterLink>
          <span class="text-gray-400">|</span>
          <span class="text-gray-600">{{ authStore.adminInfo?.username || '管理员' }}</span>
          <button
            @click="handleLogout"
            class="text-gray-600 hover:text-red-600"
          >
            退出登录
          </button>
        </div>
      </header>

      <!-- 页面内容 -->
      <main class="flex-1 p-6 overflow-auto">
        <RouterView />
      </main>
    </div>
  </div>
</template>
