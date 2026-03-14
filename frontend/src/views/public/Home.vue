<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import FileCard from "../../components/FileCard.vue";
import InfoPanelCard from "../../components/InfoPanelCard.vue";
import SearchSection from "../../components/SearchSection.vue";
import { httpClient } from "../../lib/http/client";

interface AnnouncementItem {
  id: string;
  title: string;
}

interface PublicFileItem {
  id: string;
  title: string;
  original_name: string;
  tags: string[];
  uploaded_at: string;
  download_count: number;
  size: number;
}

interface PublicFileListResponse {
  items: PublicFileItem[];
}

const announcements = ref<string[]>([]);
const hotDownloads = ref<string[]>([]);
const latestTitles = ref<string[]>([]);
const latestFiles = ref<PublicFileItem[]>([]);
const featuredFile = ref<PublicFileItem | null>(null);

const searchTags = computed(() => {
  const values = new Set<string>();
  for (const file of latestFiles.value) {
    for (const tag of file.tags ?? []) {
      if (values.size >= 6) break;
      values.add(tag);
    }
    if (values.size >= 6) break;
  }
  return [...values];
});

onMounted(async () => {
  await Promise.all([loadAnnouncements(), loadHotDownloads(), loadLatestFiles()]);
});

async function loadAnnouncements() {
  try {
    const response = await httpClient.get<{ items: AnnouncementItem[] }>("/public/announcements");
    announcements.value = (response.items ?? []).map((item) => item.title);
  } catch {
    announcements.value = [];
  }
}

async function loadHotDownloads() {
  try {
    const response = await httpClient.get<PublicFileListResponse>("/public/files?sort=download_count_desc&page=1&page_size=3");
    hotDownloads.value = (response.items ?? []).map((item) => item.title);
  } catch {
    hotDownloads.value = [];
  }
}

async function loadLatestFiles() {
  try {
    const response = await httpClient.get<PublicFileListResponse>("/public/files?sort=created_at_desc&page=1&page_size=5");
    latestFiles.value = response.items ?? [];
    latestTitles.value = latestFiles.value.map((item) => item.title);
    featuredFile.value = latestFiles.value[0] ?? null;
  } catch {
    latestFiles.value = [];
    latestTitles.value = [];
    featuredFile.value = null;
  }
}

function formatSize(size: number) {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
  }).format(new Date(value));
}

function fileTypeLabel(name: string) {
  const index = name.lastIndexOf(".");
  return index >= 0 ? name.slice(index + 1).toUpperCase() : "FILE";
}
</script>

<template>
  <main class="app-container py-8 lg:py-10">
    <div class="grid gap-6 xl:grid-cols-[248px_minmax(0,1fr)]">
      <aside class="space-y-4 xl:pt-2">
        <InfoPanelCard title="公告栏" :items="announcements" empty-text="暂无公告" />
        <InfoPanelCard title="热门下载" :items="hotDownloads" empty-text="暂无下载数据" />
        <InfoPanelCard title="资料上新" :items="latestTitles" empty-text="暂无最新资料" />
      </aside>

      <section class="min-w-0 space-y-5">
        <SearchSection :tags="searchTags" />

        <FileCard
          v-if="featuredFile"
          :title="featuredFile.title"
          :description="featuredFile.original_name"
          :size="formatSize(featuredFile.size)"
          :updated-at="formatDate(featuredFile.uploaded_at)"
          :source="`下载 ${featuredFile.download_count} 次`"
          :tags="featuredFile.tags"
          badge="最新资料"
          action-text="查看详情"
          :thumbnail-label="fileTypeLabel(featuredFile.original_name)"
          :action-href="`/files/${featuredFile.id}`"
        />
        <div v-else class="panel p-6 text-sm text-slate-500 dark:text-slate-400">
          暂无可展示的资料。
        </div>
      </section>
    </div>
  </main>
</template>
