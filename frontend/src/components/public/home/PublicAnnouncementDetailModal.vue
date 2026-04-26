<script setup lang="ts">
import { renderSimpleMarkdown } from "../../../lib/markdown";

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
  announcementDetail: AnnouncementItem | null;
}>();

defineEmits<{
  back: [];
  close: [];
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
        v-if="announcementDetail"
        class="fixed inset-0 z-[120] flex items-center justify-center bg-slate-950/30 px-4"
      >
        <div class="modal-card panel w-full max-w-2xl p-6">
          <div
            class="flex items-start justify-between gap-4 border-b border-slate-200 pb-4"
          >
            <div class="min-w-0">
              <p
                class="text-xs font-semibold uppercase tracking-[0.18em] text-blue-600"
              >
                Announcement
              </p>
              <h3 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">
                {{ announcementDetail.title }}
              </h3>
              <div
                class="mt-3 flex flex-wrap items-center gap-3 text-sm text-slate-500"
              >
                <div class="flex items-center gap-2">
                  <div
                    class="flex h-8 w-8 items-center justify-center overflow-hidden rounded-full bg-slate-100 text-xs font-semibold text-slate-600"
                  >
                    <img
                      v-if="announcementDetail.creator?.avatar_url"
                      :src="announcementDetail.creator.avatar_url"
                      alt="发布人头像"
                      class="h-full w-full object-cover"
                    />
                    <span v-else>{{
                      announcementAuthorInitial(announcementDetail)
                    }}</span>
                  </div>
                  <span class="font-medium text-slate-700">
                    {{ announcementAuthorName(announcementDetail) }}
                  </span>
                </div>
                <span
                  v-if="announcementAuthorIsSuperAdmin(announcementDetail)"
                  class="rounded-full bg-[#fff1e4] px-2.5 py-1 text-xs font-semibold text-[#d07a2d]"
                >
                  超级管理员
                </span>
                <span>
                  {{
                    formatDateTime(
                      announcementDetail.published_at
                        || announcementDetail.updated_at,
                    )
                  }}
                </span>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <button type="button" class="btn-secondary" @click="$emit('back')">
                返回
              </button>
              <button type="button" class="btn-secondary" @click="$emit('close')">
                关闭
              </button>
            </div>
          </div>
          <div class="mt-5 rounded-3xl border border-slate-200 bg-white px-5 py-5">
            <div
              class="markdown-content"
              v-html="renderSimpleMarkdown(announcementDetail.content)"
            />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
