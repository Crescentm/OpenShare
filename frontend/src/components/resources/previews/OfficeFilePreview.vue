<script setup lang="ts">
import { defineAsyncComponent } from "vue";

interface Props {
  normalizedExtension: string;
  officeFileContent: ArrayBuffer | null;
  previewError: string;
  previewHeight: string;
  previewLoading: boolean;
  previewProgress: number;
}

defineProps<Props>();

defineEmits<{
  retry: [];
}>();

const VueOfficeDocx = defineAsyncComponent(async () => {
  await import("@vue-office/docx/lib/index.css");
  return import("@vue-office/docx");
});

const VueOfficeExcel = defineAsyncComponent(async () => {
  await import("@vue-office/excel/lib/index.css");
  return import("@vue-office/excel");
});

const VueOfficePptx = defineAsyncComponent(() => import("@vue-office/pptx"));
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
      v-if="previewLoading && !officeFileContent"
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
      v-else-if="officeFileContent"
      class="overflow-hidden rounded-3xl border border-slate-200"
    >
      <Suspense>
        <template #default>
          <VueOfficeDocx
            v-if="normalizedExtension === 'docx'"
            :src="officeFileContent"
            :style="{ height: previewHeight, width: '100%' }"
          />
          <VueOfficePptx
            v-else-if="normalizedExtension === 'pptx'"
            :src="officeFileContent"
            :style="{ height: previewHeight, width: '100%' }"
          />
          <VueOfficeExcel
            v-else
            :src="officeFileContent"
            :style="{ height: previewHeight, width: '100%' }"
          />
        </template>
        <template #fallback>
          <div class="flex min-h-[220px] items-center justify-center">
            <div class="space-y-2 text-center">
              <div
                class="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent"
              ></div>
              <p class="text-sm text-slate-500">正在初始化预览引擎...</p>
            </div>
          </div>
        </template>
      </Suspense>
    </div>

    <div v-else class="flex min-h-[220px] items-center justify-center">
      <p class="text-sm text-slate-400">暂无内容可预览</p>
    </div>
  </div>
</template>
