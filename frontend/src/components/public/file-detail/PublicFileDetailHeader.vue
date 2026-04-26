<script setup lang="ts">
import { Download, Flag } from "lucide-vue-next";
import type { PublicFileDetailResponse } from "./types";

interface Props {
  canManageResourceDescriptions: boolean;
  detail: PublicFileDetailResponse;
  formattedDate: string;
  formattedSize: string;
}

defineProps<Props>();

defineEmits<{
  download: [];
  edit: [];
  feedback: [];
  goBack: [];
  openDelete: [];
}>();
</script>

<template>
  <div
    class="flex flex-col gap-4 border-b border-slate-200 pb-4 lg:flex-row lg:items-start lg:justify-between"
  >
    <div class="min-w-0 flex-1">
      <button
        type="button"
        class="inline-flex items-center gap-2 text-sm text-slate-500 transition hover:text-slate-800"
        @click="$emit('goBack')"
      >
        返回文件夹
      </button>
      <p class="mt-3 text-xs uppercase tracking-[0.14em] text-slate-400">
        {{ detail.path || "主页根目录" }}
      </p>
      <h1
        class="mt-2 break-words text-xl font-semibold tracking-tight text-slate-900 sm:text-2xl"
      >
        {{ detail.name }}
      </h1>
      <div class="mt-3 flex flex-wrap gap-2">
        <span
          class="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-600"
        >
          {{ formattedSize }}
        </span>
        <span
          class="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-600"
        >
          下载 {{ detail.download_count }}
        </span>
        <span
          class="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-600"
        >
          {{ formattedDate }}
        </span>
        <span
          class="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-600"
        >
          {{ detail.extension || "未知格式" }}
        </span>
      </div>
    </div>
    <div
      class="flex flex-wrap items-center justify-start gap-2 lg:justify-end"
    >
      <button
        v-if="canManageResourceDescriptions"
        type="button"
        class="btn-secondary h-10 px-4 text-sm"
        @click="$emit('edit')"
      >
        编辑
      </button>
      <button
        v-if="canManageResourceDescriptions"
        type="button"
        class="btn-secondary h-10 px-4 text-sm text-rose-600 hover:border-rose-200 hover:bg-rose-50 hover:text-rose-700"
        @click="$emit('openDelete')"
      >
        删除
      </button>
      <button
        type="button"
        class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-500 transition-[transform,background-color,border-color,box-shadow,color] duration-200 hover:-translate-y-0.5 hover:border-slate-300 hover:bg-[#fafafa] hover:text-slate-900 hover:shadow-sm hover:shadow-slate-950/[0.08]"
        aria-label="反馈文件"
        @click="$emit('feedback')"
      >
        <Flag class="h-4 w-4" />
      </button>
      <button
        type="button"
        class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-700 transition-[transform,background-color,border-color,box-shadow,color] duration-200 hover:-translate-y-0.5 hover:border-slate-300 hover:bg-[#fafafa] hover:text-slate-900 hover:shadow-sm hover:shadow-slate-950/[0.08]"
        aria-label="下载文件"
        @click="$emit('download')"
      >
        <Download class="h-4 w-4" />
      </button>
    </div>
  </div>
</template>
