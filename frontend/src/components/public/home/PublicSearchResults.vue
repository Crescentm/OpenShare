<script setup lang="ts">
import {
  Archive,
  BookOpen,
  Download,
  File,
  FileImage,
  FileText,
  Folder,
  Layers,
  SearchX,
} from "lucide-vue-next";
import type { SearchResultItem } from "./types";

export type SearchFilterKey =
  | "all"
  | "file"
  | "folder"
  | "pdf"
  | "office"
  | "image"
  | "archive"
  | "exam"
  | "courseware"
  | "homework"
  | "note"
  | "experiment"
  | "textbook"
  | "review";

type HighlightPart = {
  text: string;
  marked: boolean;
};

defineProps<{
  activeFilter: SearchFilterKey;
  filters: Array<{ key: SearchFilterKey; label: string }>;
  formatDate: (value?: string) => string;
  formatSize: (value?: number) => string;
  items: SearchResultItem[];
  keyword: string;
  loading: boolean;
  total: number;
}>();

defineEmits<{
  clear: [];
  download: [item: SearchResultItem];
  feedback: [item: SearchResultItem];
  open: [item: SearchResultItem];
  updateFilter: [filter: SearchFilterKey];
}>();

function iconFor(item: SearchResultItem) {
  if (item.entity_type === "folder") {
    return Folder;
  }
  switch (item.file_kind) {
    case "pdf":
    case "text":
    case "markdown":
      return FileText;
    case "office":
      return BookOpen;
    case "image":
      return FileImage;
    case "archive":
      return Archive;
    default:
      return File;
  }
}

function typeLabel(item: SearchResultItem) {
  if (item.entity_type === "folder") {
    return "目录";
  }
  if (item.extension) {
    return item.extension.toUpperCase();
  }
  return "文件";
}

function resultPath(item: SearchResultItem) {
  if (item.path) {
    return item.path;
  }
  if (item.path_segments?.length) {
    return item.path_segments.join(" / ");
  }
  return "";
}

function resultSnippet(item: SearchResultItem) {
  return (
    item.snippet
    || item.highlights?.content_text
    || item.highlights?.description
    || item.highlights?.readme
    || item.highlights?.path
    || ""
  );
}

function highlightText(value: string): HighlightPart[] {
  const result: HighlightPart[] = [];
  let cursor = 0;

  while (cursor < value.length) {
    const start = value.indexOf("[[", cursor);
    if (start < 0) {
      result.push({ text: value.slice(cursor), marked: false });
      break;
    }
    if (start > cursor) {
      result.push({ text: value.slice(cursor, start), marked: false });
    }

    const end = value.indexOf("]]", start + 2);
    if (end < 0) {
      result.push({ text: value.slice(start), marked: false });
      break;
    }

    result.push({ text: value.slice(start + 2, end), marked: true });
    cursor = end + 2;
  }

  return result.filter((part) => part.text.length > 0);
}
</script>

<template>
  <section class="px-4 py-5 sm:px-6">
    <div class="space-y-4">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
        <div class="min-w-0">
          <p class="text-xs font-semibold uppercase tracking-[0.16em] text-blue-600">
            Search Results
          </p>
          <h2 class="mt-1 truncate text-lg font-semibold text-slate-950">
            {{ keyword }}
          </h2>
          <p class="mt-1 text-sm text-slate-500">共 {{ total }} 条结果</p>
        </div>

        <button
          type="button"
          class="btn-secondary w-full sm:w-auto"
          @click="$emit('clear')"
        >
          清除搜索
        </button>
      </div>

      <div class="flex gap-2 overflow-x-auto pb-1">
        <button
          v-for="filter in filters"
          :key="filter.key"
          type="button"
          class="inline-flex h-9 shrink-0 items-center rounded-full border px-3 text-sm transition"
          :class="
            activeFilter === filter.key
              ? 'border-slate-900 bg-slate-900 text-white'
              : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300 hover:bg-slate-50 hover:text-slate-900'
          "
          @click="$emit('updateFilter', filter.key)"
        >
          {{ filter.label }}
        </button>
      </div>

      <div v-if="loading" class="grid gap-3">
        <div
          v-for="index in 6"
          :key="index"
          class="animate-pulse rounded-xl border border-slate-200 bg-white px-4 py-4"
        >
          <div class="h-5 w-3/5 rounded bg-slate-100"></div>
          <div class="mt-3 h-4 w-4/5 rounded bg-slate-100"></div>
          <div class="mt-4 flex gap-2">
            <div class="h-6 w-16 rounded-full bg-slate-100"></div>
            <div class="h-6 w-20 rounded-full bg-slate-100"></div>
            <div class="h-6 w-14 rounded-full bg-slate-100"></div>
          </div>
        </div>
      </div>

      <div
        v-else-if="items.length === 0"
        class="rounded-xl border border-slate-200 bg-white px-5 py-10 text-center"
      >
        <SearchX class="mx-auto h-8 w-8 text-slate-300" />
        <p class="mt-3 text-sm font-medium text-slate-800">没有找到匹配结果</p>
        <p class="mt-1 text-sm text-slate-500">
          可以换课程全称、简称，或尝试“试卷”“课件”“答案”等资料类型词。
        </p>
      </div>

      <div v-else class="grid gap-3">
        <article
          v-for="item in items"
          :key="`${item.entity_type}-${item.id}`"
          class="group rounded-xl border border-slate-200 bg-white px-4 py-4 transition hover:border-slate-300 hover:shadow-sm"
        >
          <button
            type="button"
            class="flex w-full min-w-0 items-start gap-3 text-left"
            @click="$emit('open', item)"
          >
            <div
              class="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-500"
            >
              <component :is="iconFor(item)" class="h-5 w-5" />
            </div>

            <div class="min-w-0 flex-1">
              <div class="flex min-w-0 flex-wrap items-center gap-2">
                <h3 class="min-w-0 max-w-full truncate text-sm font-semibold text-slate-950">
                  <template
                    v-for="(part, index) in highlightText(item.highlights?.name || item.name)"
                    :key="index"
                  >
                    <mark
                      v-if="part.marked"
                      class="rounded bg-amber-100 px-0.5 text-slate-950"
                    >
                      {{ part.text }}
                    </mark>
                    <template v-else>{{ part.text }}</template>
                  </template>
                </h3>
                <span class="rounded-full bg-slate-100 px-2 py-0.5 text-[11px] font-medium text-slate-600">
                  {{ typeLabel(item) }}
                </span>
              </div>

              <p
                v-if="resultPath(item)"
                class="mt-1 line-clamp-1 text-xs text-slate-500"
              >
                <template
                  v-for="(part, index) in highlightText(item.highlights?.path || resultPath(item))"
                  :key="index"
                >
                  <mark
                    v-if="part.marked"
                    class="rounded bg-amber-100 px-0.5 text-slate-800"
                  >
                    {{ part.text }}
                  </mark>
                  <template v-else>{{ part.text }}</template>
                </template>
              </p>

              <p
                v-if="resultSnippet(item)"
                class="mt-2 line-clamp-2 text-sm leading-6 text-slate-600"
              >
                <template
                  v-for="(part, index) in highlightText(resultSnippet(item))"
                  :key="index"
                >
                  <mark
                    v-if="part.marked"
                    class="rounded bg-amber-100 px-0.5 text-slate-900"
                  >
                    {{ part.text }}
                  </mark>
                  <template v-else>{{ part.text }}</template>
                </template>
              </p>

              <div class="mt-3 flex flex-wrap items-center gap-2 text-xs text-slate-500">
                <span
                  v-if="item.category"
                  class="inline-flex items-center gap-1 rounded-full border border-slate-200 px-2 py-1"
                >
                  <Layers class="h-3 w-3" />
                  {{ item.category }}
                </span>
                <span
                  v-if="item.course"
                  class="rounded-full border border-slate-200 px-2 py-1"
                >
                  {{ item.course }}
                </span>
                <span
                  v-if="item.material_type"
                  class="rounded-full border border-slate-200 px-2 py-1"
                >
                  {{ item.material_type }}
                </span>
                <span>{{ formatSize(item.size) }}</span>
                <span>{{ item.download_count || 0 }} 次下载</span>
                <span>{{ formatDate(item.updated_at || item.uploaded_at) }}</span>
              </div>
            </div>
          </button>

          <div class="mt-3 flex items-center justify-end gap-2 border-t border-slate-100 pt-3">
            <button
              type="button"
              class="btn-secondary h-9 px-3 text-xs"
              @click="$emit('feedback', item)"
            >
              反馈
            </button>
            <button
              type="button"
              class="btn-secondary h-9 px-3 text-xs"
              @click="$emit('download', item)"
            >
              <Download class="h-3.5 w-3.5" />
              下载
            </button>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>
