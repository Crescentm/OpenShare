<script setup lang="ts">
import { Download, Flag } from "lucide-vue-next";
import type { FolderDetailResponse } from "./types";

defineProps<{
  canManageResourceDescriptions: boolean;
  currentFolderDescriptionHtml: string;
  currentFolderDetail: FolderDetailResponse | null;
  currentFolderStats: Array<{ label: string; value: string }>;
  readmePreviewError: string;
  readmePreviewHtml: string;
  readmePreviewLoading: boolean;
  readmePreviewName: string;
}>();

defineEmits<{
  deleteFolder: [];
  downloadFolder: [];
  editFolder: [];
  feedbackFolder: [];
}>();
</script>

<template>
  <div
    v-if="currentFolderDetail"
    class="border-t border-slate-200 px-4 py-5 sm:px-6"
  >
    <section>
      <div
        class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
      >
        <div class="min-w-0 flex-1 space-y-3">
          <p class="text-xs font-semibold uppercase tracking-[0.18em] text-blue-600">
            Folder Info
          </p>
          <div class="flex flex-wrap items-center gap-x-8 gap-y-3 text-sm text-slate-500">
            <div
              v-for="item in currentFolderStats"
              :key="item.label"
              class="inline-flex items-center gap-2"
            >
              <span>{{ item.label }}</span>
              <span class="font-medium text-slate-900">{{ item.value }}</span>
            </div>
          </div>
        </div>
        <div class="flex flex-wrap items-start gap-3">
          <button
            v-if="canManageResourceDescriptions"
            type="button"
            class="btn-secondary"
            @click="$emit('editFolder')"
          >
            编辑
          </button>
          <button
            v-if="canManageResourceDescriptions"
            type="button"
            class="btn-secondary text-rose-600 hover:border-rose-200 hover:bg-rose-50 hover:text-rose-700"
            @click="$emit('deleteFolder')"
          >
            删除
          </button>
          <button
            type="button"
            class="inline-flex h-11 w-11 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-500 transition-[transform,background-color,border-color,box-shadow,color] duration-200 hover:-translate-y-0.5 hover:border-slate-300 hover:bg-[#fafafa] hover:text-slate-900 hover:shadow-sm hover:shadow-slate-950/[0.08]"
            aria-label="反馈文件夹"
            @click="$emit('feedbackFolder')"
          >
            <Flag class="h-4 w-4" />
          </button>
          <button
            type="button"
            class="inline-flex h-11 w-11 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-700 transition-[transform,background-color,border-color,box-shadow,color] duration-200 hover:-translate-y-0.5 hover:border-slate-300 hover:bg-[#fafafa] hover:text-slate-900 hover:shadow-sm hover:shadow-slate-950/[0.08]"
            aria-label="下载文件夹"
            @click="$emit('downloadFolder')"
          >
            <Download class="h-4 w-4" />
          </button>
        </div>
      </div>

      <div class="mt-4 rounded-3xl border border-slate-200 bg-white px-4 py-4 sm:px-5 sm:py-5">
        <div
          v-if="currentFolderDescriptionHtml"
          class="markdown-content"
          v-html="currentFolderDescriptionHtml"
        />
        <p v-else class="text-sm text-slate-400">该文件夹暂无简介orz</p>
      </div>

      <div
        v-if="(readmePreviewLoading || readmePreviewError || readmePreviewHtml) && currentFolderDetail.description"
        class="mt-4 rounded-3xl border border-slate-200 bg-white px-4 py-4 sm:px-5 sm:py-5"
      >
        <div class="mb-3 flex items-center justify-between gap-2">
          <p class="text-xs font-semibold uppercase tracking-[0.18em] text-blue-600">
            README Preview
          </p>
          <p class="text-xs text-slate-400">
            {{ readmePreviewName || "README.md" }}
          </p>
        </div>
        <p v-if="readmePreviewLoading" class="text-sm text-slate-500">
          README 加载中…
        </p>
        <p v-else-if="readmePreviewError" class="text-sm text-rose-600">
          {{ readmePreviewError }}
        </p>
        <div
          v-else
          class="markdown-content"
          v-html="readmePreviewHtml"
        />
      </div>
    </section>
  </div>
</template>
