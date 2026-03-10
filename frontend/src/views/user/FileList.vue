<script setup lang="ts">
import { ref } from 'vue'

const files = ref([])
const loading = ref(false)

// TODO: 获取文件列表
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold mb-6">资料库</h1>
    
    <div v-if="loading" class="text-center py-12">
      <div class="text-gray-500">加载中...</div>
    </div>
    
    <div v-else-if="files.length === 0" class="text-center py-12">
      <div class="text-gray-500">暂无资料</div>
    </div>
    
    <div v-else class="grid gap-4">
      <!-- 文件列表将在后续实现 -->
    </div>
  </div>
</template>
