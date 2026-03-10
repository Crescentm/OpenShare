<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const keyword = ref(route.query.q as string || '')
const results = ref([])
const loading = ref(false)

watch(() => route.query.q, (newVal) => {
  keyword.value = newVal as string || ''
  // TODO: 执行搜索
})

// TODO: 搜索功能实现
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold mb-2">搜索结果</h1>
    <p class="text-gray-500 mb-6">关键词：{{ keyword }}</p>
    
    <div v-if="loading" class="text-center py-12">
      <div class="text-gray-500">搜索中...</div>
    </div>
    
    <div v-else-if="results.length === 0" class="text-center py-12">
      <div class="text-gray-500">未找到相关资料</div>
    </div>
    
    <div v-else class="grid gap-4">
      <!-- 搜索结果将在后续实现 -->
    </div>
  </div>
</template>
