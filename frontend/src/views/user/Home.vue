<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'

const announcements = ref<{ id: number; title: string; content: string }[]>([])

onMounted(() => {
  // TODO: 从 API 获取公告
})
</script>

<template>
  <div>
    <!-- Hero 区域 -->
    <section class="bg-gradient-to-r from-primary-600 to-primary-700 text-white py-16">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <h1 class="text-4xl font-bold mb-4">OpenShare</h1>
        <p class="text-xl text-primary-100 mb-8">轻量级资料共享平台，让知识自由流动</p>
        <div class="flex justify-center space-x-4">
          <RouterLink to="/files" class="btn bg-white text-primary-600 hover:bg-primary-50">
            浏览资料
          </RouterLink>
          <RouterLink to="/upload" class="btn bg-primary-500 text-white hover:bg-primary-400 border border-white">
            上传资料
          </RouterLink>
        </div>
      </div>
    </section>

    <!-- 公告区域 -->
    <section v-if="announcements.length > 0" class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
        <h2 class="font-medium text-yellow-800 mb-2">📢 系统公告</h2>
        <ul class="space-y-1">
          <li v-for="item in announcements" :key="item.id" class="text-yellow-700">
            {{ item.title }}
          </li>
        </ul>
      </div>
    </section>

    <!-- 快捷入口 -->
    <section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div class="grid md:grid-cols-3 gap-6">
        <div class="card text-center hover:shadow-md transition-shadow">
          <div class="text-4xl mb-4">📚</div>
          <h3 class="text-lg font-medium mb-2">浏览资料</h3>
          <p class="text-gray-500 text-sm mb-4">探索丰富的学习资源</p>
          <RouterLink to="/files" class="text-primary-600 hover:underline">
            进入资料库 →
          </RouterLink>
        </div>

        <div class="card text-center hover:shadow-md transition-shadow">
          <div class="text-4xl mb-4">📤</div>
          <h3 class="text-lg font-medium mb-2">上传资料</h3>
          <p class="text-gray-500 text-sm mb-4">分享你的学习成果</p>
          <RouterLink to="/upload" class="text-primary-600 hover:underline">
            开始上传 →
          </RouterLink>
        </div>

        <div class="card text-center hover:shadow-md transition-shadow">
          <div class="text-4xl mb-4">🔍</div>
          <h3 class="text-lg font-medium mb-2">我的上传</h3>
          <p class="text-gray-500 text-sm mb-4">查看投稿审核状态</p>
          <RouterLink to="/my-uploads" class="text-primary-600 hover:underline">
            查看记录 →
          </RouterLink>
        </div>
      </div>
    </section>
  </div>
</template>
