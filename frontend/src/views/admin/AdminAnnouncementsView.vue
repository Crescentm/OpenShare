<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";

import EmptyState from "../../components/ui/EmptyState.vue";
import PageHeader from "../../components/ui/PageHeader.vue";
import SurfaceCard from "../../components/ui/SurfaceCard.vue";
import { httpClient } from "../../lib/http/client";
import { readApiError } from "../../lib/http/helpers";
import { useSessionStore } from "../../stores/session";

type AnnouncementStatus = "draft" | "published" | "hidden";

interface AnnouncementItem {
  id: string;
  title: string;
  content: string;
  status: AnnouncementStatus;
  created_by_id: string;
  published_at?: string | null;
  created_at: string;
  updated_at: string;
}

const sessionStore = useSessionStore();
const items = ref<AnnouncementItem[]>([]);
const loading = ref(false);
const error = ref("");
const message = ref("");
const saving = ref(false);
const editingId = ref("");
const form = reactive({
  title: "",
  content: "",
  status: "draft" as AnnouncementStatus,
});

onMounted(() => {
  if (sessionStore.hasPermission("manage_announcements")) {
    void loadItems();
  }
});

async function loadItems() {
  loading.value = true;
  error.value = "";
  message.value = "";
  try {
    const response = await httpClient.get<{ items: AnnouncementItem[] }>("/admin/announcements");
    items.value = response.items ?? [];
  } catch {
    error.value = "加载公告失败。";
  } finally {
    loading.value = false;
  }
}

async function saveAnnouncement() {
  saving.value = true;
  error.value = "";
  message.value = "";
  try {
    const isEditing = editingId.value !== "";
    if (editingId.value) {
      await httpClient.request(`/admin/announcements/${editingId.value}`, {
        method: "PUT",
        body: form,
      });
    } else {
      await httpClient.post("/admin/announcements", form);
    }
    resetForm();
    message.value = isEditing ? "公告已更新。" : "公告已创建。";
    await loadItems();
  } catch (err: unknown) {
    error.value = readApiError(err, "保存公告失败。");
  } finally {
    saving.value = false;
  }
}

async function removeAnnouncement(id: string) {
  if (!window.confirm("确认删除这条公告吗？")) return;
  error.value = "";
  message.value = "";
  try {
    await httpClient.request(`/admin/announcements/${id}`, { method: "DELETE" });
    message.value = "公告已删除。";
    await loadItems();
  } catch (err: unknown) {
    error.value = readApiError(err, "删除公告失败。");
  }
}

function editAnnouncement(item: AnnouncementItem) {
  editingId.value = item.id;
  form.title = item.title;
  form.content = item.content;
  form.status = item.status;
}

function resetForm() {
  editingId.value = "";
  form.title = "";
  form.content = "";
  form.status = "draft";
}

function formatDate(value?: string | null) {
  if (!value) return "未发布";
  return new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function canDeleteAnnouncement(item: AnnouncementItem) {
  if (sessionStore.isSuperAdmin) {
    return true;
  }
  return item.created_by_id === sessionStore.adminId;
}
</script>

<template>
  <section class="space-y-8">
    <PageHeader
      eyebrow="Announcements"
      title="公告管理"
      description="首页公告应该是少量、明确、可控的内容，因此后台编辑区和列表区保持简单清晰。"
    />

    <SurfaceCard v-if="!sessionStore.hasPermission('manage_announcements')">
      <p class="text-sm text-slate-500">当前账号没有公告管理权限。</p>
    </SurfaceCard>

    <template v-else>
      <section class="grid gap-6 xl:grid-cols-[minmax(0,420px)_minmax(0,1fr)]">
        <SurfaceCard>
          <div class="flex items-center justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold text-slate-900">{{ editingId ? "编辑公告" : "新建公告" }}</h2>
              <p class="mt-1 text-sm text-slate-500">只有已发布公告会展示在前台首页。</p>
            </div>
            <button class="btn-secondary" @click="resetForm">重置</button>
          </div>

          <form class="mt-6 space-y-4" @submit.prevent="saveAnnouncement">
            <input v-model="form.title" class="field" placeholder="公告标题" />
            <textarea v-model="form.content" rows="6" class="field-area" placeholder="公告内容" />
            <select v-model="form.status" class="field">
              <option value="draft">草稿</option>
              <option value="published">发布</option>
              <option value="hidden">隐藏</option>
            </select>
            <button type="submit" class="btn-primary" :disabled="saving">
              {{ saving ? "保存中…" : editingId ? "保存修改" : "创建公告" }}
            </button>
          </form>
        </SurfaceCard>

        <SurfaceCard>
          <div class="flex items-center justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold text-slate-900">公告列表</h2>
              <p class="mt-1 text-sm text-slate-500">按状态与发布时间管理对外展示内容。</p>
            </div>
            <button class="btn-secondary" @click="loadItems">刷新</button>
          </div>

          <p v-if="loading" class="mt-4 text-sm text-slate-500">加载中…</p>
          <div v-else class="mt-6 space-y-4">
            <article v-for="item in items" :key="item.id" class="rounded-xl border border-slate-200 p-5">
              <div class="flex flex-wrap items-start justify-between gap-4">
                <div class="min-w-0 flex-1">
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="text-base font-semibold text-slate-900">{{ item.title }}</h3>
                    <span class="rounded-lg bg-slate-100 px-2.5 py-1 text-xs text-slate-600">{{ item.status }}</span>
                  </div>
                  <p class="mt-2 text-sm leading-6 text-slate-500">{{ item.content }}</p>
                  <p class="mt-3 text-xs text-slate-400">发布时间：{{ formatDate(item.published_at) }}</p>
                </div>
                <div class="flex gap-2">
                  <button class="btn-secondary" @click="editAnnouncement(item)">编辑</button>
                  <button v-if="canDeleteAnnouncement(item)" class="btn-danger" @click="removeAnnouncement(item.id)">删除</button>
                </div>
              </div>
            </article>

            <EmptyState v-if="items.length === 0" title="还没有公告" description="创建第一条公告后，已发布内容会出现在首页公告区。" />
          </div>
        </SurfaceCard>
      </section>

      <p v-if="message" class="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{{ message }}</p>
      <p v-if="error" class="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{{ error }}</p>
    </template>
  </section>
</template>
