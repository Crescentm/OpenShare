<script setup lang="ts">
import { ChevronLeft, ChevronRight, LayoutGrid, List, Upload } from "lucide-vue-next";
import type {
  PublicHomeSortDirection,
  PublicHomeSortMode,
  PublicHomeViewMode,
} from "./types";

defineProps<{
  allVisibleRowsSelected: boolean;
  backButtonLabel: string;
  canUploadToCurrentFolder: boolean;
  canUseBackButton: boolean;
  hasRows: boolean;
  sortDirection: PublicHomeSortDirection;
  sortMenuOpen: boolean;
  sortMode: PublicHomeSortMode;
  viewMenuOpen: boolean;
  viewMode: PublicHomeViewMode;
}>();

defineEmits<{
  goUp: [];
  openUpload: [];
  setSortDirection: [direction: PublicHomeSortDirection];
  setSortMenuOpen: [open: boolean];
  setSortMode: [mode: PublicHomeSortMode];
  setViewMenuOpen: [open: boolean];
  setViewMode: [mode: PublicHomeViewMode];
  toggleSelectAll: [];
}>();

function sortModeLabel(mode: PublicHomeSortMode) {
  switch (mode) {
    case "download":
      return "下载量排序";
    case "format":
      return "格式排序";
    default:
      return "名称排序";
  }
}

function sortDirectionLabel(direction: PublicHomeSortDirection) {
  return direction === "asc" ? "升序" : "降序";
}

function viewModeLabel(mode: PublicHomeViewMode) {
  return mode === "cards" ? "卡片" : "表格";
}
</script>

<template>
  <div class="px-4 pb-2 sm:px-6">
    <div class="flex flex-wrap items-center gap-3 border-t border-slate-100 pt-3">
      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-xl border border-slate-200 px-3 py-2 text-sm font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900 disabled:cursor-not-allowed disabled:opacity-45"
        :disabled="!canUseBackButton"
        @click="$emit('goUp')"
      >
        <ChevronLeft class="h-4 w-4" />
        {{ backButtonLabel }}
      </button>

      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-xl border border-slate-200 px-3 py-2 text-sm font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900"
        :disabled="!canUploadToCurrentFolder"
        :class="
          !canUploadToCurrentFolder
            ? 'cursor-not-allowed opacity-45 hover:border-slate-200 hover:text-slate-600'
            : ''
        "
        @click="$emit('openUpload')"
      >
        <Upload class="h-4 w-4" />
        {{ canUploadToCurrentFolder ? "在该目录上传" : "进入目录后上传" }}
      </button>

      <button
        v-if="hasRows"
        type="button"
        class="inline-flex items-center gap-2 rounded-xl border border-slate-200 px-3 py-2 text-sm font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900"
        @click="$emit('toggleSelectAll')"
      >
        {{ allVisibleRowsSelected ? "取消全选" : "全选" }}
      </button>

      <div
        class="flex w-full flex-wrap items-center gap-3 sm:ml-auto sm:w-auto sm:justify-end"
      >
        <div class="relative">
          <button
            type="button"
            class="inline-flex w-full items-center justify-center gap-2 rounded-xl border border-slate-200 px-3 py-2 text-sm font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900 sm:w-auto"
            @click="
              $emit('setSortMenuOpen', !sortMenuOpen);
              $emit('setViewMenuOpen', false);
            "
          >
            {{ sortModeLabel(sortMode) }} · {{ sortDirectionLabel(sortDirection) }}
            <ChevronRight class="h-4 w-4 rotate-90" />
          </button>
          <div
            v-if="sortMenuOpen"
            class="absolute left-0 top-full z-20 mt-2 min-w-[176px] rounded-2xl border border-slate-200 bg-white p-1 shadow-lg"
          >
            <button
              type="button"
              class="block w-full rounded-xl px-3 py-2 text-left text-sm transition"
              :class="
                sortMode === 'download'
                  ? 'bg-slate-100 font-medium text-slate-900'
                  : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
              "
              @click="$emit('setSortMode', 'download')"
            >
              下载量排序
            </button>
            <button
              type="button"
              class="block w-full rounded-xl px-3 py-2 text-left text-sm transition"
              :class="
                sortMode === 'name'
                  ? 'bg-slate-100 font-medium text-slate-900'
                  : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
              "
              @click="$emit('setSortMode', 'name')"
            >
              名称排序
            </button>
            <button
              type="button"
              class="block w-full rounded-xl px-3 py-2 text-left text-sm transition"
              :class="
                sortMode === 'format'
                  ? 'bg-slate-100 font-medium text-slate-900'
                  : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
              "
              @click="$emit('setSortMode', 'format')"
            >
              格式排序
            </button>
            <div class="mx-2 my-1 border-t border-slate-100"></div>
            <button
              type="button"
              class="block w-full rounded-xl px-3 py-2 text-left text-sm transition"
              :class="
                sortDirection === 'desc'
                  ? 'bg-slate-100 font-medium text-slate-900'
                  : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
              "
              @click="$emit('setSortDirection', 'desc')"
            >
              降序
            </button>
            <button
              type="button"
              class="block w-full rounded-xl px-3 py-2 text-left text-sm transition"
              :class="
                sortDirection === 'asc'
                  ? 'bg-slate-100 font-medium text-slate-900'
                  : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
              "
              @click="$emit('setSortDirection', 'asc')"
            >
              升序
            </button>
          </div>
        </div>

        <div class="relative">
          <button
            type="button"
            class="inline-flex w-full items-center justify-center gap-2 rounded-xl border border-slate-200 px-3 py-2 text-sm font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900 sm:w-auto"
            @click="
              $emit('setViewMenuOpen', !viewMenuOpen);
              $emit('setSortMenuOpen', false);
            "
          >
            <LayoutGrid v-if="viewMode === 'cards'" class="h-4 w-4" />
            <List v-else class="h-4 w-4" />
            {{ viewModeLabel(viewMode) }}
            <ChevronRight class="h-4 w-4 rotate-90" />
          </button>
          <div
            v-if="viewMenuOpen"
            class="absolute left-0 top-full z-20 mt-2 min-w-[124px] rounded-2xl border border-slate-200 bg-white p-1 shadow-lg"
          >
            <button
              type="button"
              class="flex w-full items-center gap-2 rounded-xl px-3 py-2 text-left text-sm transition"
              :class="
                viewMode === 'cards'
                  ? 'bg-slate-100 font-medium text-slate-900'
                  : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
              "
              @click="$emit('setViewMode', 'cards')"
            >
              <LayoutGrid class="h-4 w-4" />
              卡片
            </button>
            <button
              type="button"
              class="flex w-full items-center gap-2 rounded-xl px-3 py-2 text-left text-sm transition"
              :class="
                viewMode === 'table'
                  ? 'bg-slate-100 font-medium text-slate-900'
                  : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
              "
              @click="$emit('setViewMode', 'table')"
            >
              <List class="h-4 w-4" />
              表格
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
