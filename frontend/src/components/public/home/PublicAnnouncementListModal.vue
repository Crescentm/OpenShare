<script setup lang="ts">
interface AnnouncementItem {
  id: string;
  title: string;
  content: string;
  is_pinned: boolean;
  creator: {
    id: string;
    username: string;
    display_name: string;
    avatar_url: string;
    role: string;
  };
  published_at?: string;
  updated_at: string;
}

defineProps<{
  announcements: AnnouncementItem[];
  open: boolean;
}>();

defineEmits<{
  close: [];
  openDetail: [item: { id: string; label: string }];
}>();

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function announcementAuthorName(item: AnnouncementItem) {
  return (
    item.creator?.display_name?.trim()
    || item.creator?.username?.trim()
    || "OpenShare"
  );
}

function announcementAuthorInitial(item: AnnouncementItem) {
  return announcementAuthorName(item).slice(0, 1).toUpperCase() || "A";
}

function announcementAuthorIsSuperAdmin(item: AnnouncementItem) {
  return item.creator?.role === "super_admin";
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-shell">
      <div
        v-if="open"
        class="fixed inset-0 z-[120] flex items-center justify-center bg-slate-950/30 px-4"
      >
        <div class="modal-card panel w-full max-w-3xl p-6">
          <div
            class="flex items-start justify-between gap-4 border-b border-slate-200 pb-4"
          >
            <div class="min-w-0">
              <p
                class="text-xs font-semibold uppercase tracking-[0.18em] text-blue-600"
              >
                Announcements
              </p>
              <h3 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">
                全部公告
              </h3>
            </div>
            <button type="button" class="btn-secondary" @click="$emit('close')">
              关闭
            </button>
          </div>
          <div class="mt-5 max-h-[70vh] space-y-3 overflow-auto pr-1">
            <button
              v-for="item in announcements"
              :key="item.id"
              type="button"
              class="flex w-full items-start justify-between gap-4 rounded-2xl border border-slate-200 bg-white px-4 py-4 text-left transition hover:border-blue-200 hover:bg-blue-50/40"
              @click="$emit('openDetail', { id: item.id, label: item.title })"
            >
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <span
                    v-if="item.is_pinned"
                    class="rounded-md bg-[#dcecff] px-2 py-0.5 text-xs font-semibold text-[#4f8ff7]"
                  >
                    置顶
                  </span>
                  <p class="text-base font-semibold text-slate-900">
                    {{ item.title }}
                  </p>
                </div>
                <div class="mt-3 flex flex-wrap items-center gap-2">
                  <div
                    class="flex h-8 w-8 items-center justify-center overflow-hidden rounded-full bg-slate-100 text-xs font-semibold text-slate-600"
                  >
                    <img
                      v-if="item.creator?.avatar_url"
                      :src="item.creator.avatar_url"
                      alt="发布人头像"
                      class="h-full w-full object-cover"
                    />
                    <span v-else>{{ announcementAuthorInitial(item) }}</span>
                  </div>
                  <span class="text-sm font-medium text-slate-700">
                    {{ announcementAuthorName(item) }}
                  </span>
                  <span
                    v-if="announcementAuthorIsSuperAdmin(item)"
                    class="rounded-full bg-[#fff1e4] px-2.5 py-1 text-xs font-semibold text-[#d07a2d]"
                  >
                    超级管理员
                  </span>
                </div>
                <p class="mt-2 line-clamp-2 text-sm text-slate-500">
                  {{ item.content }}
                </p>
              </div>
              <span class="shrink-0 text-sm text-slate-400">
                {{ formatDateTime(item.published_at || item.updated_at) }}
              </span>
            </button>
            <p
              v-if="announcements.length === 0"
              class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-6 text-center text-sm text-slate-500"
            >
              暂无公告
            </p>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
