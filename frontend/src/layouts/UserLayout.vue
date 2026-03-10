<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'

const router = useRouter()
const searchKeyword = ref('')
const mobileMenuOpen = ref(false)

const handleSearch = () => {
  if (searchKeyword.value.trim()) {
    router.push({ name: 'Search', query: { q: searchKeyword.value.trim() } })
  }
}
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <!-- 顶部导航 -->
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <!-- Logo -->
          <RouterLink to="/" class="flex items-center">
            <span class="text-xl font-bold text-primary-600">OpenShare</span>
          </RouterLink>

          <!-- 搜索框（桌面端） -->
          <div class="hidden md:flex flex-1 max-w-lg mx-8">
            <div class="relative w-full">
              <input
                v-model="searchKeyword"
                type="text"
                placeholder="搜索资料..."
                class="input pr-10"
                @keyup.enter="handleSearch"
              />
              <button
                @click="handleSearch"
                class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-primary-600"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </button>
            </div>
          </div>

          <!-- 导航链接（桌面端） -->
          <nav class="hidden md:flex items-center space-x-4">
            <RouterLink to="/files" class="text-gray-600 hover:text-primary-600">资料库</RouterLink>
            <RouterLink to="/upload" class="btn btn-primary">上传资料</RouterLink>
            <RouterLink to="/my-uploads" class="text-gray-600 hover:text-primary-600">我的上传</RouterLink>
          </nav>

          <!-- 移动端菜单按钮 -->
          <button
            @click="mobileMenuOpen = !mobileMenuOpen"
            class="md:hidden p-2 text-gray-600"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
        </div>

        <!-- 移动端菜单 -->
        <div v-if="mobileMenuOpen" class="md:hidden py-4 border-t border-gray-200">
          <div class="mb-4">
            <input
              v-model="searchKeyword"
              type="text"
              placeholder="搜索资料..."
              class="input"
              @keyup.enter="handleSearch"
            />
          </div>
          <nav class="flex flex-col space-y-2">
            <RouterLink to="/files" class="py-2 text-gray-600">资料库</RouterLink>
            <RouterLink to="/upload" class="py-2 text-primary-600 font-medium">上传资料</RouterLink>
            <RouterLink to="/my-uploads" class="py-2 text-gray-600">我的上传</RouterLink>
          </nav>
        </div>
      </div>
    </header>

    <!-- 主内容区 -->
    <main class="flex-1">
      <RouterView />
    </main>

    <!-- 底部 -->
    <footer class="bg-white border-t border-gray-200 py-6">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <p class="text-center text-gray-500 text-sm">
          OpenShare - 轻量级资料共享平台
        </p>
      </div>
    </footer>
  </div>
</template>
