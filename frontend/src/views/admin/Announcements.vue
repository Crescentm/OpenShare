<script setup lang="ts">
import { ref } from 'vue'

const announcements = ref([])
const loading = ref(false)

// TODO: 获取公告列表
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold">公告管理</h1>
      <button class="btn btn-primary">发布公告</button>
    </div>
    
    <div class="card">
      <div v-if="loading" class="text-center py-8">
        <div class="text-gray-500">加载中...</div>
      </div>
      
      <div v-else-if="announcements.length === 0" class="text-center py-8">
        <div class="text-gray-500">暂无公告</div>
      </div>
      
      <div v-else>
        <!-- 公告列表将在后续实现 -->
      </div>
    </div>
  </div>
</template>
