<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const admins = ref([])
const loading = ref(false)

// TODO: 获取管理员列表
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold">管理员管理</h1>
      <button v-if="authStore.isSuperAdmin" class="btn btn-primary">创建管理员</button>
    </div>
    
    <div class="card">
      <div v-if="loading" class="text-center py-8">
        <div class="text-gray-500">加载中...</div>
      </div>
      
      <div v-else-if="admins.length === 0" class="text-center py-8">
        <div class="text-gray-500">暂无管理员</div>
      </div>
      
      <div v-else>
        <!-- 管理员列表将在后续实现 -->
      </div>
    </div>
  </div>
</template>
