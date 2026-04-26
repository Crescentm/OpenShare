<script setup lang="ts">
import type { FileMetadataRow } from "./types";

defineProps<{
  metadataRows: FileMetadataRow[];
}>();
</script>

<template>
  <details class="group rounded-2xl border border-slate-200 bg-slate-50/70">
    <summary
      class="cursor-pointer list-none px-4 py-3 text-sm font-medium text-slate-700 marker:hidden"
    >
      更多信息
    </summary>
    <div
      class="grid gap-3 border-t border-slate-200 px-4 py-4 sm:grid-cols-2 xl:grid-cols-3"
    >
      <div
        v-for="item in metadataRows"
        :key="item.label"
        class="min-w-0 rounded-xl border border-slate-200 bg-white px-3 py-3"
      >
        <p class="text-[11px] uppercase tracking-[0.12em] text-slate-400">
          {{ item.label }}
        </p>
        <p
          class="mt-1 break-words text-sm font-medium text-slate-800"
          :title="item.value"
        >
          {{ item.value }}
        </p>
      </div>
    </div>
  </details>
</template>
