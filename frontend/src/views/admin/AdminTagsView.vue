<script setup lang="ts">
import { onMounted, ref } from "vue";

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

// --- Tag list ---
const tags = ref<TagItem[]>([]);
const tagsLoading = ref(false);
const tagsError = ref("");

// --- Create tag ---
const newTagName = ref("");
const createLoading = ref(false);
const createMessage = ref("");
const createError = ref("");

// --- Edit tag ---
const editingTagId = ref<string | null>(null);
const editingTagName = ref("");
const editLoading = ref(false);
const editError = ref("");

// --- Merge tags ---
const mergeSourceId = ref("");
const mergeTargetId = ref("");
const mergeLoading = ref(false);
const mergeMessage = ref("");
const mergeError = ref("");

// --- Tag submissions ---
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
    await httpClient.post(`/admin/tag-submissions/${encodeURIComponent(id)}/reject`, {
      reject_reason: reason,
    });
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
    <header>
      <p class="text-xs font-semibold uppercase tracking-[0.22em] text-blue-300">Tag Management</p>
      <h2 class="mt-2 text-3xl font-semibold text-white">Tag 管理</h2>
    </header>

    <!-- Create Tag -->
    <article class="rounded-[28px] border border-slate-800 bg-slate-900 p-6">
      <h3 class="text-lg font-semibold text-white">创建 Tag</h3>
      <form class="mt-4 flex gap-3" @submit.prevent="createTag">
        <input
          v-model="newTagName"
          placeholder="输入新 Tag 名称"
          class="min-w-0 flex-1 rounded-2xl border border-slate-700 bg-slate-950 px-4 py-3 text-sm text-white outline-none focus:border-blue-400"
        />
        <button
          type="submit"
          class="rounded-2xl bg-blue-500 px-5 py-3 text-sm font-semibold text-slate-950 transition hover:bg-blue-400 disabled:bg-slate-600"
          :disabled="createLoading"
        >
          {{ createLoading ? "创建中..." : "创建" }}
        </button>
      </form>
      <p v-if="createMessage" class="mt-3 rounded-xl bg-emerald-950/60 px-4 py-2 text-sm text-emerald-200">{{ createMessage }}</p>
      <p v-if="createError" class="mt-3 rounded-xl bg-rose-950/60 px-4 py-2 text-sm text-rose-200">{{ createError }}</p>
    </article>

    <!-- Tag List -->
    <article class="rounded-[28px] border border-slate-800 bg-slate-900 p-6">
      <div class="flex items-center justify-between gap-4">
        <h3 class="text-lg font-semibold text-white">所有 Tag</h3>
        <span class="rounded-full bg-slate-800 px-3 py-1 text-xs font-medium text-slate-300">{{ tags.length }} 个</span>
      </div>

      <p v-if="tagsLoading" class="mt-4 text-sm text-slate-400">加载中...</p>
      <p v-if="tagsError" class="mt-4 rounded-xl bg-rose-950/60 px-4 py-2 text-sm text-rose-200">{{ tagsError }}</p>
      <p v-if="editError" class="mt-4 rounded-xl bg-rose-950/60 px-4 py-2 text-sm text-rose-200">{{ editError }}</p>

      <div v-if="!tagsLoading && tags.length > 0" class="mt-4 space-y-2">
        <div
          v-for="tag in tags"
          :key="tag.id"
          class="flex flex-wrap items-center justify-between gap-3 rounded-2xl bg-slate-950 px-4 py-3"
        >
          <div v-if="editingTagId === tag.id" class="flex min-w-0 flex-1 items-center gap-2">
            <input
              v-model="editingTagName"
              class="min-w-0 flex-1 rounded-xl border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-white outline-none focus:border-blue-400"
              @keyup.enter="saveEdit"
              @keyup.escape="cancelEdit"
            />
            <button class="rounded-xl bg-blue-500 px-3 py-2 text-xs font-semibold text-slate-950" :disabled="editLoading" @click="saveEdit">
              保存
            </button>
            <button class="rounded-xl border border-slate-700 px-3 py-2 text-xs text-slate-300" @click="cancelEdit">
              取消
            </button>
          </div>

          <template v-else>
            <div class="min-w-0 flex-1">
              <span class="text-sm font-medium text-white">{{ tag.name }}</span>
              <span class="ml-3 text-xs text-slate-500">
                文件 {{ tag.file_count }} · 文件夹 {{ tag.folder_count }}
              </span>
            </div>
            <div class="flex gap-2">
              <button
                class="rounded-xl border border-slate-700 px-3 py-1.5 text-xs text-slate-300 transition hover:bg-slate-800"
                @click="startEdit(tag)"
              >
                编辑
              </button>
              <button
                class="rounded-xl border border-rose-800 px-3 py-1.5 text-xs text-rose-300 transition hover:bg-rose-950"
                @click="deleteTag(tag)"
              >
                删除
              </button>
            </div>
          </template>
        </div>
      </div>

      <p v-if="!tagsLoading && tags.length === 0" class="mt-4 text-sm text-slate-500">暂无 Tag。</p>
    </article>

    <!-- Merge Tags -->
    <article class="rounded-[28px] border border-slate-800 bg-slate-900 p-6">
      <h3 class="text-lg font-semibold text-white">合并 Tag</h3>
      <p class="mt-2 text-sm text-slate-400">将源 Tag 的所有关联合并到目标 Tag，源 Tag 将被删除。</p>

      <form class="mt-4 flex flex-wrap items-end gap-3" @submit.prevent="mergeTags">
        <label class="block min-w-[200px] flex-1">
          <span class="mb-2 block text-xs uppercase tracking-[0.18em] text-slate-500">源 Tag（将被删除）</span>
          <select
            v-model="mergeSourceId"
            class="w-full rounded-2xl border border-slate-700 bg-slate-950 px-4 py-3 text-sm text-white outline-none focus:border-blue-400"
          >
            <option value="">请选择</option>
            <option v-for="tag in tags" :key="tag.id" :value="tag.id">{{ tag.name }}</option>
          </select>
        </label>

        <span class="pb-3 text-slate-500">→</span>

        <label class="block min-w-[200px] flex-1">
          <span class="mb-2 block text-xs uppercase tracking-[0.18em] text-slate-500">目标 Tag（保留）</span>
          <select
            v-model="mergeTargetId"
            class="w-full rounded-2xl border border-slate-700 bg-slate-950 px-4 py-3 text-sm text-white outline-none focus:border-blue-400"
          >
            <option value="">请选择</option>
            <option v-for="tag in tags" :key="tag.id" :value="tag.id">{{ tag.name }}</option>
          </select>
        </label>

        <button
          type="submit"
          class="rounded-2xl bg-amber-500 px-5 py-3 text-sm font-semibold text-slate-950 transition hover:bg-amber-400 disabled:bg-slate-600"
          :disabled="mergeLoading"
        >
          {{ mergeLoading ? "合并中..." : "执行合并" }}
        </button>
      </form>

      <p v-if="mergeMessage" class="mt-3 rounded-xl bg-emerald-950/60 px-4 py-2 text-sm text-emerald-200">{{ mergeMessage }}</p>
      <p v-if="mergeError" class="mt-3 rounded-xl bg-rose-950/60 px-4 py-2 text-sm text-rose-200">{{ mergeError }}</p>
    </article>

    <!-- Pending Tag Submissions -->
    <article class="rounded-[28px] border border-slate-800 bg-slate-900 p-6">
      <div class="flex items-center justify-between gap-4">
        <h3 class="text-lg font-semibold text-white">用户提交的 Tag</h3>
        <span class="rounded-full bg-amber-900 px-3 py-1 text-xs font-medium text-amber-200">
          {{ pendingSubmissions.length }} 条待审
        </span>
      </div>

      <p v-if="submissionsLoading" class="mt-4 text-sm text-slate-400">加载中...</p>

      <div v-if="!submissionsLoading && pendingSubmissions.length > 0" class="mt-4 space-y-2">
        <div
          v-for="sub in pendingSubmissions"
          :key="sub.id"
          class="flex flex-wrap items-center justify-between gap-3 rounded-2xl bg-slate-950 px-4 py-3"
        >
          <div>
            <span class="text-sm font-medium text-white">{{ sub.proposed_name }}</span>
            <span class="ml-3 text-xs text-slate-500">{{ formatDate(sub.created_at) }}</span>
          </div>
          <div class="flex gap-2">
            <button
              class="rounded-xl bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-emerald-500"
              @click="approveSubmission(sub.id)"
            >
              通过
            </button>
            <button
              class="rounded-xl bg-rose-600 px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-rose-500"
              @click="rejectSubmission(sub.id)"
            >
              驳回
            </button>
          </div>
        </div>
      </div>

      <p v-if="!submissionsLoading && pendingSubmissions.length === 0" class="mt-4 text-sm text-slate-500">暂无待审 Tag 提交。</p>
    </article>
  </section>
</template>
