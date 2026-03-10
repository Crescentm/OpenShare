<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const _fileId = route.params.id as string // eslint-disable-line @typescript-eslint/no-unused-vars
const file = ref(null)
const loading = ref(true)

// TODO: 获取文件详情
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div v-if="loading" class="text-center py-12">
      <div class="text-gray-500">加载中...</div>
    </div>
    
    <div v-else-if="!file" class="text-center py-12">
      <div class="text-gray-500">资料不存在</div>
    </div>
    
    <div v-else class="card">
      <!-- 文件详情将在后续实现 -->
    </div>
  </div>
</template>
