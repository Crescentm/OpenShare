<script setup lang="ts">
import { ref } from 'vue'

const submissions = ref([])
const loading = ref(false)

// TODO: 获取待审核列表
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">审核管理</h1>
    
    <div class="card">
      <div v-if="loading" class="text-center py-8">
        <div class="text-gray-500">加载中...</div>
      </div>
      
      <div v-else-if="submissions.length === 0" class="text-center py-8">
        <div class="text-gray-500">暂无待审核资料</div>
      </div>
      
      <div v-else>
        <!-- 审核列表将在后续实现 -->
      </div>
    </div>
  </div>
</template>
