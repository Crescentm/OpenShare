<script setup lang="ts">
import { Clock3, Download, Flag, Folder } from "lucide-vue-next";
import type { DirectoryRow } from "./types";

defineProps<{
  fileIconComponent: (extension: string) => unknown;
  isRowSelected: (row: DirectoryRow) => boolean;
  rows: DirectoryRow[];
}>();

defineEmits<{
  download: [row: DirectoryRow];
  feedback: [row: DirectoryRow];
  open: [row: DirectoryRow];
  toggleSelection: [row: DirectoryRow];
}>();
</script>

<template>
  <div class="grid gap-3 px-4 py-3 md:grid-cols-2 xl:grid-cols-3 sm:px-6 2xl:grid-cols-4">
    <article
      v-for="row in rows"
      :key="`${row.kind}-${row.id}`"
      class="group relative min-w-0 flex min-h-[132px] cursor-pointer flex-col rounded-2xl border border-slate-200 bg-white px-3.5 pt-3 transition hover:border-slate-300 hover:shadow-sm sm:px-4"
      @click="$emit('open', row)"
    >
      <div class="absolute right-4 top-3.5 z-10">
        <input
          :checked="isRowSelected(row)"
          type="checkbox"
          class="h-4.5 w-4.5 rounded-md border-slate-300 text-slate-900 focus:ring-slate-300"
          @click.stop
          @change="$emit('toggleSelection', row)"
        />
      </div>
      <div class="flex items-start gap-3">
        <div
          class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-slate-100 text-slate-500"
        >
          <Folder
            v-if="row.kind === 'folder'"
            class="h-5.5 w-5.5 text-blue-500"
          />
          <component
            v-else
            :is="fileIconComponent(row.extension)"
            class="h-5.5 w-5.5"
          />
        </div>
        <div class="min-w-0 flex-1 pr-8 pt-0.5">
          <h3 class="truncate text-sm font-semibold leading-5 text-slate-900">
            {{ row.name }}
          </h3>
          <div
            class="mt-1 flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-slate-500"
          >
            <template v-if="row.kind === 'file'">
              <span class="inline-flex items-center gap-1.5">
                <Download class="h-3 w-3" />
                {{ row.downloadCount }}
              </span>
              <span>{{ row.sizeText }}</span>
            </template>
            <template v-else>
              <span class="inline-flex items-center gap-1.5">
                <Download class="h-3 w-3" />
                {{ row.downloadCount }}
              </span>
              <span>{{ row.fileCount }} 个文件</span>
              <span>{{ row.sizeText }}</span>
            </template>
            <span class="inline-flex min-w-0 max-w-full items-center gap-1.5">
              <Clock3 class="h-3 w-3" />
              <span class="truncate">{{ row.updatedAt }}</span>
            </span>
          </div>
          <p
            v-if="row.kind === 'file' && row.description"
            class="mt-0.5 line-clamp-1 text-xs leading-4.5 text-slate-500"
          >
            {{ row.description }}
          </p>
        </div>
      </div>

      <div class="mt-auto flex items-center justify-between border-t border-slate-100 py-2">
        <button
          type="button"
          class="inline-flex items-center justify-center rounded-lg border border-slate-200 bg-white p-2 text-slate-700 transition hover:border-slate-300 hover:bg-slate-50 hover:text-slate-900"
          @click.stop="$emit('feedback', row)"
        >
          <Flag class="h-3.5 w-3.5" />
        </button>
        <button
          type="button"
          class="inline-flex items-center justify-center rounded-lg border border-slate-200 bg-white p-2 text-slate-700 transition hover:border-slate-300 hover:bg-slate-50 hover:text-slate-900"
          @click.stop="$emit('download', row)"
        >
          <Download class="h-3.5 w-3.5" />
        </button>
      </div>
    </article>
  </div>
</template>
