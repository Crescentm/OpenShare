<script setup lang="ts">
import { ref } from 'vue'

const tags = ref([])
const loading = ref(false)

// TODO: 获取 Tag 列表
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold">Tag 管理</h1>
      <button class="btn btn-primary">新建 Tag</button>
    </div>
    
    <div class="card">
      <div v-if="loading" class="text-center py-8">
        <div class="text-gray-500">加载中...</div>
      </div>
      
      <div v-else-if="tags.length === 0" class="text-center py-8">
        <div class="text-gray-500">暂无 Tag</div>
      </div>
      
      <div v-else>
        <!-- Tag 列表将在后续实现 -->
      </div>
    </div>
  </div>
</template>
