<script setup lang="ts">
import { ref } from 'vue'

const settings = ref({
  allowGuestUpload: true,
  requireReview: true,
})
const loading = ref(false)

// TODO: 获取系统设置
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">系统设置</h1>
    
    <div class="card">
      <div v-if="loading" class="text-center py-8">
        <div class="text-gray-500">加载中...</div>
      </div>
      
      <div v-else class="space-y-6">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="font-medium">允许游客上传</h3>
            <p class="text-sm text-gray-500">是否允许未登录用户上传资料</p>
          </div>
          <label class="relative inline-flex items-center cursor-pointer">
            <input type="checkbox" v-model="settings.allowGuestUpload" class="sr-only peer">
            <div class="w-11 h-6 bg-gray-200 peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
          </label>
        </div>

        <div class="flex items-center justify-between">
          <div>
            <h3 class="font-medium">上传需要审核</h3>
            <p class="text-sm text-gray-500">上传的资料是否需要管理员审核后才能公开</p>
          </div>
          <label class="relative inline-flex items-center cursor-pointer">
            <input type="checkbox" v-model="settings.requireReview" class="sr-only peer">
            <div class="w-11 h-6 bg-gray-200 peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
          </label>
        </div>

        <div class="pt-4 border-t border-gray-200">
          <button class="btn btn-primary">保存设置</button>
        </div>
      </div>
    </div>
  </div>
</template>
