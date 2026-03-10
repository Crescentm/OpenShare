<script setup lang="ts">
import { ref } from 'vue'

const files = ref([])
const loading = ref(false)

// TODO: 获取文件列表
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">资料管理</h1>
    
    <div class="card">
      <div v-if="loading" class="text-center py-8">
        <div class="text-gray-500">加载中...</div>
      </div>
      
      <div v-else-if="files.length === 0" class="text-center py-8">
        <div class="text-gray-500">暂无资料</div>
      </div>
      
      <div v-else>
        <!-- 文件管理列表将在后续实现 -->
      </div>
    </div>
  </div>
</template>
