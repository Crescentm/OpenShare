<script setup lang="ts">
import { onMounted, ref } from "vue";

import EmptyState from "../../components/ui/EmptyState.vue";
import PageHeader from "../../components/ui/PageHeader.vue";
import SurfaceCard from "../../components/ui/SurfaceCard.vue";
import { httpClient } from "../../lib/http/client";
import { readApiError } from "../../lib/http/helpers";

interface TagItem {
  id: string;
  name: string;
  file_count: number;
  folder_count: number;
  created_at: string;
  updated_at: string;
}

interface TagSubmissionItem {
  id: string;
  proposed_name: string;
  status: string;
  submitter_ip: string;
  created_at: string;
}

const tags = ref<TagItem[]>([]);
const tagsLoading = ref(false);
const tagsError = ref("");

const newTagName = ref("");
const createLoading = ref(false);
const createMessage = ref("");
const createError = ref("");

const editingTagId = ref<string | null>(null);
const editingTagName = ref("");
const editLoading = ref(false);
const editError = ref("");

const mergeSourceId = ref("");
const mergeTargetId = ref("");
const mergeLoading = ref(false);
const mergeMessage = ref("");
const mergeError = ref("");

const pendingSubmissions = ref<TagSubmissionItem[]>([]);
const submissionsLoading = ref(false);

onMounted(async () => {
  await Promise.all([loadTags(), loadPendingSubmissions()]);
});

async function loadTags() {
  tagsLoading.value = true;
  tagsError.value = "";
  try {
    const response = await httpClient.get<{ items: TagItem[] }>("/admin/tags");
    tags.value = response.items ?? [];
  } catch (err: unknown) {
    tagsError.value = readApiError(err, "加载 Tag 列表失败。");
  } finally {
    tagsLoading.value = false;
  }
}

async function createTag() {
  const name = newTagName.value.trim();
  if (!name) {
    createError.value = "请输入 Tag 名称。";
    return;
  }
  createLoading.value = true;
  createError.value = "";
  createMessage.value = "";
  try {
    await httpClient.post("/admin/tags", { name });
    createMessage.value = `Tag "${name}" 已创建。`;
    newTagName.value = "";
    await loadTags();
  } catch (err: unknown) {
    createError.value = readApiError(err, "创建 Tag 失败。");
  } finally {
    createLoading.value = false;
  }
}

function startEdit(tag: TagItem) {
  editingTagId.value = tag.id;
  editingTagName.value = tag.name;
  editError.value = "";
}

function cancelEdit() {
  editingTagId.value = null;
  editingTagName.value = "";
  editError.value = "";
}

async function saveEdit() {
  if (!editingTagId.value || !editingTagName.value.trim()) return;
  editLoading.value = true;
  editError.value = "";
  try {
    await httpClient.request(`/admin/tags/${encodeURIComponent(editingTagId.value)}`, {
      method: "PUT",
      body: { name: editingTagName.value.trim() },
    });
    editingTagId.value = null;
    editingTagName.value = "";
    await loadTags();
  } catch (err: unknown) {
    editError.value = readApiError(err, "更新 Tag 失败。");
  } finally {
    editLoading.value = false;
  }
}

async function deleteTag(tag: TagItem) {
  if (!window.confirm(`确定要删除 Tag "${tag.name}" 吗？关联关系会一并解除。`)) return;
  try {
    await httpClient.request(`/admin/tags/${encodeURIComponent(tag.id)}`, { method: "DELETE" });
    await loadTags();
  } catch (err: unknown) {
    tagsError.value = readApiError(err, "删除 Tag 失败。");
  }
}

async function mergeTags() {
  if (!mergeSourceId.value || !mergeTargetId.value) {
    mergeError.value = "请选择要合并的源 Tag 和目标 Tag。";
    return;
  }
  if (mergeSourceId.value === mergeTargetId.value) {
    mergeError.value = "源 Tag 和目标 Tag 不能相同。";
    return;
  }

  mergeLoading.value = true;
  mergeError.value = "";
  mergeMessage.value = "";
  try {
    await httpClient.post("/admin/tags/merge", {
      source_tag_id: mergeSourceId.value,
      target_tag_id: mergeTargetId.value,
    });
    mergeMessage.value = "Tag 合并成功。";
    mergeSourceId.value = "";
    mergeTargetId.value = "";
    await loadTags();
  } catch (err: unknown) {
    mergeError.value = readApiError(err, "合并 Tag 失败。");
  } finally {
    mergeLoading.value = false;
  }
}

async function loadPendingSubmissions() {
  submissionsLoading.value = true;
  try {
    const response = await httpClient.get<{ items: TagSubmissionItem[] }>("/admin/tag-submissions/pending");
    pendingSubmissions.value = response.items ?? [];
  } catch {
    pendingSubmissions.value = [];
  } finally {
    submissionsLoading.value = false;
  }
}

async function approveSubmission(id: string) {
  try {
    await httpClient.post(`/admin/tag-submissions/${encodeURIComponent(id)}/approve`);
    await Promise.all([loadPendingSubmissions(), loadTags()]);
  } catch (err: unknown) {
    tagsError.value = readApiError(err, "审批失败。");
  }
}

async function rejectSubmission(id: string) {
  const reason = window.prompt("请输入驳回原因：");
  if (reason === null) return;
  try {
    await httpClient.post(`/admin/tag-submissions/${encodeURIComponent(id)}/reject`, { reject_reason: reason });
    await loadPendingSubmissions();
  } catch (err: unknown) {
    tagsError.value = readApiError(err, "驳回失败。");
  }
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
</script>

<template>
  <section class="space-y-8">
    <PageHeader
      eyebrow="Tags"
      title="标签管理"
      description="标签的创建、重命名、合并和用户投稿审批都放在一个工作台中，但分块清晰，避免操作混淆。"
    />

    <section class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
      <div class="space-y-6">
        <SurfaceCard>
          <h2 class="text-lg font-semibold text-slate-900">创建标签</h2>
          <form class="mt-4 flex flex-col gap-3 sm:flex-row" @submit.prevent="createTag">
            <input v-model="newTagName" class="field flex-1" placeholder="输入新 Tag 名称" />
            <button type="submit" class="btn-primary" :disabled="createLoading">
              {{ createLoading ? "创建中…" : "创建" }}
            </button>
          </form>
          <p v-if="createMessage" class="mt-3 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{{ createMessage }}</p>
          <p v-if="createError" class="mt-3 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{{ createError }}</p>
        </SurfaceCard>

        <SurfaceCard>
          <div class="flex items-center justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold text-slate-900">所有标签</h2>
              <p class="mt-1 text-sm text-slate-500">支持直接重命名、删除和查看关联数量。</p>
            </div>
            <span class="rounded-lg bg-slate-100 px-3 py-1 text-sm text-slate-600">{{ tags.length }} 个</span>
          </div>

          <p v-if="tagsLoading" class="mt-4 text-sm text-slate-500">加载中…</p>
          <p v-if="tagsError" class="mt-4 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{{ tagsError }}</p>
          <p v-if="editError" class="mt-4 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{{ editError }}</p>

          <div v-if="!tagsLoading && tags.length > 0" class="mt-5 space-y-3">
            <div v-for="tag in tags" :key="tag.id" class="rounded-xl border border-slate-200 p-4">
              <div v-if="editingTagId === tag.id" class="flex flex-col gap-3 sm:flex-row">
                <input
                  v-model="editingTagName"
                  class="field flex-1"
                  @keyup.enter="saveEdit"
                  @keyup.escape="cancelEdit"
                />
                <button class="btn-primary" :disabled="editLoading" @click="saveEdit">保存</button>
                <button class="btn-secondary" @click="cancelEdit">取消</button>
              </div>

              <div v-else class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-slate-900">{{ tag.name }}</p>
                  <p class="mt-1 text-sm text-slate-500">文件 {{ tag.file_count }} · 文件夹 {{ tag.folder_count }}</p>
                </div>
                <div class="flex gap-2">
                  <button class="btn-secondary" @click="startEdit(tag)">编辑</button>
                  <button class="btn-danger" @click="deleteTag(tag)">删除</button>
                </div>
              </div>
            </div>
          </div>

          <EmptyState v-if="!tagsLoading && tags.length === 0" title="暂无标签" description="创建首个标签后，这里会展示标签及其关联数量。" />
        </SurfaceCard>
      </div>

      <div class="space-y-6">
        <SurfaceCard>
          <h2 class="text-lg font-semibold text-slate-900">合并标签</h2>
          <p class="mt-1 text-sm text-slate-500">将源标签的关联内容迁移到目标标签，完成后源标签会被删除。</p>
          <form class="mt-4 space-y-4" @submit.prevent="mergeTags">
            <select v-model="mergeSourceId" class="field">
              <option value="">请选择源 Tag</option>
              <option v-for="tag in tags" :key="tag.id" :value="tag.id">{{ tag.name }}</option>
            </select>
            <select v-model="mergeTargetId" class="field">
              <option value="">请选择目标 Tag</option>
              <option v-for="tag in tags" :key="tag.id" :value="tag.id">{{ tag.name }}</option>
            </select>
            <button type="submit" class="btn-primary" :disabled="mergeLoading">
              {{ mergeLoading ? "合并中…" : "执行合并" }}
            </button>
          </form>
          <p v-if="mergeMessage" class="mt-3 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{{ mergeMessage }}</p>
          <p v-if="mergeError" class="mt-3 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{{ mergeError }}</p>
        </SurfaceCard>

        <SurfaceCard>
          <div class="flex items-center justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold text-slate-900">用户提交的标签</h2>
              <p class="mt-1 text-sm text-slate-500">审核用户提出的新标签，保持标签体系整洁。</p>
            </div>
            <span class="rounded-lg bg-slate-100 px-3 py-1 text-sm text-slate-600">{{ pendingSubmissions.length }} 条待审</span>
          </div>

          <p v-if="submissionsLoading" class="mt-4 text-sm text-slate-500">加载中…</p>
          <div v-else class="mt-5 space-y-3">
            <div v-for="sub in pendingSubmissions" :key="sub.id" class="rounded-xl border border-slate-200 p-4">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-slate-900">{{ sub.proposed_name }}</p>
                  <p class="mt-1 text-sm text-slate-500">{{ formatDate(sub.created_at) }} · {{ sub.submitter_ip }}</p>
                </div>
                <div class="flex gap-2">
                  <button class="btn-primary" @click="approveSubmission(sub.id)">通过</button>
                  <button class="btn-danger" @click="rejectSubmission(sub.id)">驳回</button>
                </div>
              </div>
            </div>

            <EmptyState v-if="pendingSubmissions.length === 0" title="暂无待审标签" description="新的用户标签投稿会自动进入这里等待审核。" />
          </div>
        </SurfaceCard>
      </div>
    </section>
  </section>
</template>
