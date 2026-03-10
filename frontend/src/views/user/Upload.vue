<script setup lang="ts">
import { ref } from 'vue'

const form = ref({
  title: '',
  description: '',
  tags: [] as string[],
  receiptCode: '',
  file: null as File | null,
})
const uploading = ref(false)
const uploadProgress = ref(0)
const uploadResult = ref<{ success: boolean; receiptCode?: string; message?: string } | null>(null)

const handleFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files && target.files[0]) {
    form.value.file = target.files[0]
  }
}

const handleSubmit = async () => {
  if (!form.value.file) {
    alert('请选择文件')
    return
  }
  
  if (!form.value.title.trim()) {
    alert('请输入标题')
    return
  }

  uploading.value = true
  uploadProgress.value = 0

  // TODO: 调用上传 API
  // 模拟上传
  setTimeout(() => {
    uploading.value = false
    uploadResult.value = {
      success: true,
      receiptCode: form.value.receiptCode || 'AUTO-' + Date.now(),
    }
  }, 1500)
}

const resetForm = () => {
  form.value = {
    title: '',
    description: '',
    tags: [],
    receiptCode: '',
    file: null,
  }
  uploadResult.value = null
}
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold mb-6">上传资料</h1>

    <!-- 上传结果 -->
    <div v-if="uploadResult" class="card mb-6">
      <div v-if="uploadResult.success" class="text-center">
        <div class="text-4xl mb-4">✅</div>
        <h2 class="text-xl font-medium text-green-600 mb-2">上传成功！</h2>
        <p class="text-gray-600 mb-4">您的资料已提交，等待管理员审核</p>
        <div class="bg-gray-100 rounded-lg p-4 mb-4">
          <p class="text-sm text-gray-500 mb-1">回执码</p>
          <p class="text-lg font-mono font-medium">{{ uploadResult.receiptCode }}</p>
          <p class="text-xs text-gray-400 mt-2">请妥善保存此回执码，用于查询审核状态</p>
        </div>
        <button @click="resetForm" class="btn btn-primary">继续上传</button>
      </div>
      <div v-else class="text-center">
        <div class="text-4xl mb-4">❌</div>
        <h2 class="text-xl font-medium text-red-600 mb-2">上传失败</h2>
        <p class="text-gray-600 mb-4">{{ uploadResult.message }}</p>
        <button @click="uploadResult = null" class="btn btn-primary">重试</button>
      </div>
    </div>

    <!-- 上传表单 -->
    <form v-else @submit.prevent="handleSubmit" class="card space-y-6">
      <!-- 文件选择 -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">
          选择文件 <span class="text-red-500">*</span>
        </label>
        <input
          type="file"
          @change="handleFileChange"
          class="w-full"
          :disabled="uploading"
        />
        <p v-if="form.file" class="mt-2 text-sm text-gray-500">
          已选择：{{ form.file.name }} ({{ (form.file.size / 1024 / 1024).toFixed(2) }} MB)
        </p>
      </div>

      <!-- 标题 -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">
          标题 <span class="text-red-500">*</span>
        </label>
        <input
          v-model="form.title"
          type="text"
          class="input"
          placeholder="为资料起一个标题"
          :disabled="uploading"
        />
      </div>

      <!-- 描述 -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">描述</label>
        <textarea
          v-model="form.description"
          class="input"
          rows="3"
          placeholder="简单描述一下这份资料（可选）"
          :disabled="uploading"
        ></textarea>
      </div>

      <!-- 回执码 -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">回执码</label>
        <input
          v-model="form.receiptCode"
          type="text"
          class="input"
          placeholder="自定义回执码（可选，留空则自动生成）"
          :disabled="uploading"
        />
        <p class="mt-1 text-xs text-gray-500">回执码用于后续查询投稿状态</p>
      </div>

      <!-- 上传进度 -->
      <div v-if="uploading">
        <div class="h-2 bg-gray-200 rounded-full overflow-hidden">
          <div
            class="h-full bg-primary-600 transition-all duration-300"
            :style="{ width: uploadProgress + '%' }"
          ></div>
        </div>
        <p class="text-center text-sm text-gray-500 mt-2">上传中... {{ uploadProgress }}%</p>
      </div>

      <!-- 提交按钮 -->
      <button
        type="submit"
        class="btn btn-primary w-full"
        :disabled="uploading"
      >
        {{ uploading ? '上传中...' : '提交上传' }}
      </button>
    </form>
  </div>
</template>
