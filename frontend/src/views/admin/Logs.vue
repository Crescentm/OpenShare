<script setup lang="ts">
import { ref } from 'vue'

const logs = ref([])
const loading = ref(false)

// TODO: 获取操作日志
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">操作日志</h1>
    
    <div class="card">
      <div v-if="loading" class="text-center py-8">
        <div class="text-gray-500">加载中...</div>
      </div>
      
      <div v-else-if="logs.length === 0" class="text-center py-8">
        <div class="text-gray-500">暂无日志</div>
      </div>
      
      <div v-else>
        <!-- 日志列表将在后续实现 -->
      </div>
    </div>
  </div>
</template>
