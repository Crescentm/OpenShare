<script setup lang="ts">
import { Folder } from "lucide-vue-next";
import type { DirectoryRow } from "./types";

defineProps<{
  fileIconComponent: (extension: string) => unknown;
  isRowSelected: (row: DirectoryRow) => boolean;
  rows: DirectoryRow[];
}>();

defineEmits<{
  open: [row: DirectoryRow];
  toggleSelection: [row: DirectoryRow];
}>();
</script>

<template>
  <div class="px-4 py-5 sm:px-6">
    <table class="data-table table-fixed">
      <thead>
        <tr>
          <th class="w-10"></th>
          <th class="text-left">名称</th>
          <th class="w-[120px] text-right">大小</th>
          <th class="hidden w-[220px] text-right xl:table-cell">修改时间</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="row in rows"
          :key="`${row.kind}-${row.id}`"
          class="cursor-pointer transition hover:bg-slate-50 dark:hover:bg-slate-800/40"
          @click="$emit('open', row)"
        >
          <td @click.stop>
            <input
              :checked="isRowSelected(row)"
              type="checkbox"
              class="h-5 w-5 rounded-lg border-slate-300 text-slate-900 focus:ring-slate-300"
              @change="$emit('toggleSelection', row)"
            />
          </td>
          <td>
            <div
              v-if="row.kind === 'folder'"
              class="flex min-w-0 items-center gap-3 text-left"
            >
              <Folder class="h-5 w-5 shrink-0 text-blue-500" />
              <span
                class="truncate text-slate-900 dark:text-slate-100"
                :title="row.name"
              >
                {{ row.name }}
              </span>
            </div>
            <div v-else class="flex min-w-0 items-center gap-3 text-left">
              <component
                :is="fileIconComponent(row.extension)"
                class="h-5 w-5 shrink-0 text-slate-500"
              />
              <span
                class="truncate text-slate-900 dark:text-slate-100"
                :title="row.name"
              >
                {{ row.name }}
              </span>
            </div>
          </td>
          <td class="w-[120px] whitespace-nowrap text-right tabular-nums">
            {{ row.sizeText }}
          </td>
          <td
            class="hidden w-[220px] whitespace-nowrap text-right tabular-nums xl:table-cell"
          >
            {{ row.updatedAt }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
