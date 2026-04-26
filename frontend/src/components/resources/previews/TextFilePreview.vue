<script setup lang="ts">
import { renderSimpleMarkdown } from "../../../lib/markdown";

interface Props {
  isMarkdownFile: boolean;
  previewError: string;
  previewLoading: boolean;
  previewProgress: number;
  textContent: string;
}

defineProps<Props>();

defineEmits<{
  retry: [];
}>();
</script>

<template>
  <div class="space-y-4">
    <div v-if="previewLoading" class="mb-3 flex items-center justify-end gap-2">
      <div
        class="h-4 w-4 animate-spin rounded-full border-2 border-blue-600 border-t-transparent"
      ></div>
      <span class="text-xs text-slate-500">
        {{ Math.round(previewProgress) }}%
      </span>
    </div>

    <div
      v-if="previewLoading && !textContent"
      class="flex min-h-[220px] items-center justify-center"
    >
      <div class="space-y-2 text-center">
        <div
          class="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent"
        ></div>
        <p class="text-sm text-slate-500">正在加载预览内容...</p>
        <p class="text-xs text-slate-400">
          {{ Math.round(previewProgress) }}% 完成
        </p>
      </div>
    </div>

    <div
      v-else-if="previewError"
      class="flex min-h-[220px] items-center justify-center"
    >
      <div class="space-y-2 text-center">
        <p class="text-sm text-rose-600">{{ previewError }}</p>
        <button
          type="button"
          class="btn-secondary text-xs"
          @click="$emit('retry')"
        >
          重试
        </button>
      </div>
    </div>

    <div
      v-else-if="isMarkdownFile && textContent"
      class="rounded-3xl border border-slate-200 p-4 sm:p-5"
    >
      <div
        class="markdown-content"
        v-html="renderSimpleMarkdown(textContent)"
      ></div>
    </div>

    <div
      v-else-if="textContent"
      class="rounded-3xl border border-slate-200 p-4 sm:p-5"
    >
      <pre
        class="max-h-96 overflow-y-auto whitespace-pre-wrap break-words text-sm text-slate-700"
      >{{ textContent }}</pre>
    </div>

    <div v-else class="flex min-h-[220px] items-center justify-center">
      <p class="text-sm text-slate-400">暂无内容可预览</p>
    </div>
  </div>
</template>
