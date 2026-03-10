<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Submission } from '@/types'

const receiptCode = ref('')
const savedCodes = ref<string[]>([])
const submissions = ref<Submission[]>([])
const loading = ref(false)
const searched = ref(false)

// 从 localStorage 加载保存的回执码
onMounted(() => {
  const saved = localStorage.getItem('receipt_codes')
  if (saved) {
    savedCodes.value = JSON.parse(saved)
  }
})

const handleSearch = async () => {
  if (!receiptCode.value.trim()) {
    alert('请输入回执码')
    return
  }

  loading.value = true
  searched.value = true

  // TODO: 调用 API 查询
  // 模拟请求
  setTimeout(() => {
    loading.value = false
    submissions.value = []
  }, 500)
}

const saveReceiptCode = () => {
  if (!receiptCode.value.trim()) return
  if (!savedCodes.value.includes(receiptCode.value)) {
    savedCodes.value.push(receiptCode.value)
    localStorage.setItem('receipt_codes', JSON.stringify(savedCodes.value))
  }
}

const useCode = (code: string) => {
  receiptCode.value = code
  handleSearch()
}

const removeCode = (code: string) => {
  savedCodes.value = savedCodes.value.filter(c => c !== code)
  localStorage.setItem('receipt_codes', JSON.stringify(savedCodes.value))
}

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已驳回',
  }
  return map[status] || status
}

const getStatusClass = (status: string) => {
  const map: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-800',
    approved: 'bg-green-100 text-green-800',
    rejected: 'bg-red-100 text-red-800',
  }
  return map[status] || 'bg-gray-100 text-gray-800'
}
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold mb-6">我的上传</h1>

    <!-- 查询表单 -->
    <div class="card mb-6">
      <label class="block text-sm font-medium text-gray-700 mb-2">回执码</label>
      <div class="flex space-x-2">
        <input
          v-model="receiptCode"
          type="text"
          class="input flex-1"
          placeholder="输入回执码查询投稿状态"
          @keyup.enter="handleSearch"
        />
        <button @click="handleSearch" class="btn btn-primary" :disabled="loading">
          {{ loading ? '查询中...' : '查询' }}
        </button>
      </div>
    </div>

    <!-- 已保存的回执码 -->
    <div v-if="savedCodes.length > 0" class="card mb-6">
      <h2 class="text-sm font-medium text-gray-700 mb-3">历史回执码</h2>
      <div class="flex flex-wrap gap-2">
        <div
          v-for="code in savedCodes"
          :key="code"
          class="flex items-center bg-gray-100 rounded-lg px-3 py-1"
        >
          <button @click="useCode(code)" class="text-sm text-gray-700 hover:text-primary-600">
            {{ code }}
          </button>
          <button @click="removeCode(code)" class="ml-2 text-gray-400 hover:text-red-500">
            ×
          </button>
        </div>
      </div>
    </div>

    <!-- 查询结果 -->
    <div v-if="searched">
      <div v-if="loading" class="text-center py-8">
        <div class="text-gray-500">查询中...</div>
      </div>
      
      <div v-else-if="submissions.length === 0" class="card text-center py-8">
        <div class="text-gray-500">未找到相关投稿记录</div>
        <p class="text-sm text-gray-400 mt-2">请检查回执码是否正确</p>
      </div>

      <div v-else class="space-y-4">
        <div v-for="item in submissions" :key="item.id" class="card">
          <div class="flex justify-between items-start mb-2">
            <h3 class="font-medium">{{ item.title }}</h3>
            <span :class="['px-2 py-1 text-xs rounded-full', getStatusClass(item.status)]">
              {{ getStatusText(item.status) }}
            </span>
          </div>
          <p class="text-sm text-gray-500 mb-2">{{ item.filename }}</p>
          <div class="flex justify-between text-sm text-gray-400">
            <span>{{ item.created_at }}</span>
            <span v-if="item.download_count !== undefined">下载 {{ item.download_count }} 次</span>
          </div>
          <p v-if="item.reject_reason" class="mt-2 text-sm text-red-600">
            驳回原因：{{ item.reject_reason }}
          </p>
        </div>

        <!-- 保存回执码按钮 -->
        <button
          v-if="!savedCodes.includes(receiptCode)"
          @click="saveReceiptCode"
          class="btn btn-secondary w-full"
        >
          保存此回执码
        </button>
      </div>
    </div>
  </div>
</template>
