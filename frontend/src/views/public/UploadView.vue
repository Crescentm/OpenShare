<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import EmptyState from "../../components/ui/EmptyState.vue";
import PageHeader from "../../components/ui/PageHeader.vue";
import SurfaceCard from "../../components/ui/SurfaceCard.vue";
import { httpClient } from "../../lib/http/client";
import { readApiError } from "../../lib/http/helpers";

interface PublicFolderItem {
  id: string;
  name: string;
}

interface PublicFolderListResponse {
  items: PublicFolderItem[];
}

interface UploadResponse {
  receipt_code: string;
  status: string;
  title: string;
  uploaded_at: string;
}

interface SubmissionLookupResponse {
  receipt_code: string;
  items: Array<{
    title: string;
    status: string;
    uploaded_at: string;
    download_count: number;
    reject_reason?: string;
  }>;
}

const folders = ref<{ id: string; name: string }[]>([]);
const foldersLoading = ref(false);

const uploadForm = ref({
  folderID: "",
  description: "",
  tags: "",
  receiptCode: "",
  file: null as File | null,
});

const uploadLoading = ref(false);
const uploadError = ref("");
const uploadMessage = ref("");
const uploadResult = ref<UploadResponse | null>(null);

const receiptCode = ref("");
const lookupLoading = ref(false);
const lookupError = ref("");
const lookupResult = ref<SubmissionLookupResponse | null>(null);

const flattenedFolders = computed(() => folders.value);

onMounted(async () => {
  receiptCode.value = localStorage.getItem("openshare_receipt_code") ?? "";
  await loadFolders();

  if (receiptCode.value.trim()) {
    await lookupReceipt();
  }
});

async function loadFolders() {
  foldersLoading.value = true;
  try {
    const result: { id: string; name: string }[] = [];

    async function loadLevel(parentId: string | null, prefix: string) {
      let url = "/public/folders";
      if (parentId) {
        url += `?parent_id=${encodeURIComponent(parentId)}`;
      }
      const response = await httpClient.get<PublicFolderListResponse>(url);
      for (const item of response.items ?? []) {
        const displayName = prefix ? `${prefix} / ${item.name}` : item.name;
        result.push({ id: item.id, name: displayName });
        await loadLevel(item.id, displayName);
      }
    }

    await loadLevel(null, "");
    folders.value = result;
  } catch {
    folders.value = [];
  } finally {
    foldersLoading.value = false;
  }
}

function onFileChange(event: Event) {
  const target = event.target as HTMLInputElement;
  uploadForm.value.file = target.files?.[0] ?? null;
}

async function submitUpload() {
  if (!uploadForm.value.file) {
    uploadError.value = "请选择要上传的文件。";
    return;
  }

  uploadLoading.value = true;
  uploadError.value = "";
  uploadMessage.value = "";
  uploadResult.value = null;

  try {
    const formData = new FormData();
    formData.set("file", uploadForm.value.file);
    formData.set("folder_id", uploadForm.value.folderID);
    formData.set("description", uploadForm.value.description.trim());

    const tags = uploadForm.value.tags
      .split(",")
      .map((entry) => entry.trim())
      .filter(Boolean);
    for (const tag of tags) {
      formData.append("tags", tag);
    }

    if (uploadForm.value.receiptCode.trim()) {
      formData.set("receipt_code", uploadForm.value.receiptCode.trim());
    }

    const response = await httpClient.post<UploadResponse>("/public/submissions", formData);
    uploadResult.value = response;
    uploadMessage.value = `资料《${response.title}》已提交，当前状态为 ${response.status}。`;
    receiptCode.value = response.receipt_code;
    localStorage.setItem("openshare_receipt_code", response.receipt_code);
    uploadForm.value.description = "";
    uploadForm.value.tags = "";
    uploadForm.value.file = null;
    await lookupReceipt();
  } catch (error: unknown) {
    uploadError.value = readApiError(error, "提交上传失败，请稍后重试。");
  } finally {
    uploadLoading.value = false;
  }
}

async function lookupReceipt() {
  const code = receiptCode.value.trim();
  if (!code) {
    lookupError.value = "请输入回执码。";
    lookupResult.value = null;
    return;
  }

  lookupLoading.value = true;
  lookupError.value = "";
  try {
    const response = await httpClient.get<SubmissionLookupResponse>(`/public/submissions/${encodeURIComponent(code)}`);
    lookupResult.value = response;
    localStorage.setItem("openshare_receipt_code", response.receipt_code);
  } catch (error: unknown) {
    lookupResult.value = null;
    lookupError.value = readApiError(error, "查询投稿记录失败。");
  } finally {
    lookupLoading.value = false;
  }
}

function clearReceipt() {
  receiptCode.value = "";
  lookupResult.value = null;
  lookupError.value = "";
  localStorage.removeItem("openshare_receipt_code");
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function statusLabel(status: string) {
  const labels: Record<string, string> = {
    pending: "待审核",
    approved: "已通过",
    rejected: "已驳回",
  };
  return labels[status] ?? status;
}
</script>

<template>
  <div class="app-container py-8 sm:py-10">
    <section class="grid gap-6 xl:grid-cols-[minmax(0,1.1fr)_minmax(360px,0.9fr)]">
      <SurfaceCard>
        <PageHeader
          eyebrow="Upload"
          title="上传资料"
          description="上传资料后会进入审核池。标题默认取文件名，Tag 和描述可选填写。"
        />

        <form class="mt-6 space-y-4" @submit.prevent="submitUpload">
          <div class="grid gap-4 md:grid-cols-2">
            <div class="space-y-2">
              <label class="text-sm font-medium text-slate-700 dark:text-slate-300">目标目录</label>
              <select v-model="uploadForm.folderID" class="field">
                <option value="">请选择目录</option>
                <option v-for="folder in flattenedFolders" :key="folder.id" :value="folder.id">
                  {{ folder.name }}
                </option>
              </select>
              <p v-if="foldersLoading" class="text-xs text-slate-400">目录加载中…</p>
            </div>

            <div class="space-y-2">
              <label class="text-sm font-medium text-slate-700 dark:text-slate-300">回执码</label>
              <input v-model="uploadForm.receiptCode" class="field" placeholder="可选，自定义或留空自动生成" />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-slate-700 dark:text-slate-300">描述</label>
            <textarea
              v-model="uploadForm.description"
              rows="4"
              class="field-area"
              placeholder="可选，简要说明资料内容和适用场景"
            />
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-slate-700 dark:text-slate-300">标签</label>
            <input v-model="uploadForm.tags" class="field" placeholder="多个标签用逗号分隔，例如：计算机网络, 期末复习" />
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-slate-700 dark:text-slate-300">文件</label>
            <input type="file" class="field flex items-center py-2.5" @change="onFileChange" />
          </div>

          <button type="submit" class="btn-primary" :disabled="uploadLoading">
            {{ uploadLoading ? "提交中…" : "提交上传" }}
          </button>
        </form>

        <p v-if="uploadMessage" class="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950/40 dark:text-emerald-300">
          {{ uploadMessage }}
        </p>
        <p v-if="uploadError" class="mt-4 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700 dark:border-rose-900 dark:bg-rose-950/40 dark:text-rose-300">
          {{ uploadError }}
        </p>
      </SurfaceCard>

      <SurfaceCard>
        <PageHeader
          eyebrow="Receipt"
          title="查询投稿记录"
          description="输入回执码查看历史投稿记录，浏览器会自动记住最近一次使用的回执码。"
        />

        <div class="mt-6 flex gap-3">
          <input v-model="receiptCode" class="field flex-1" placeholder="输入回执码" @keydown.enter.prevent="lookupReceipt" />
          <button class="btn-secondary" @click="lookupReceipt">查询</button>
        </div>

        <div class="mt-4 flex gap-3">
          <button class="text-sm text-slate-500 transition hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100" @click="clearReceipt">
            清除本地回执码
          </button>
        </div>

        <p v-if="lookupError" class="mt-4 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700 dark:border-rose-900 dark:bg-rose-950/40 dark:text-rose-300">
          {{ lookupError }}
        </p>
        <p v-else-if="lookupLoading" class="mt-4 text-sm text-slate-500 dark:text-slate-400">正在查询…</p>

        <div v-else-if="lookupResult" class="mt-6 space-y-3">
          <article
            v-for="item in lookupResult.items"
            :key="`${item.title}-${item.uploaded_at}`"
            class="rounded-xl border border-slate-200 p-4 dark:border-slate-800"
          >
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h3 class="text-sm font-semibold text-slate-900 dark:text-slate-100">{{ item.title }}</h3>
                <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ formatDate(item.uploaded_at) }}</p>
              </div>
              <span class="rounded-md bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-700 dark:bg-slate-800 dark:text-slate-200">
                {{ statusLabel(item.status) }}
              </span>
            </div>
            <div class="mt-3 flex flex-wrap gap-4 text-sm text-slate-500 dark:text-slate-400">
              <span>下载 {{ item.download_count }}</span>
              <span v-if="item.reject_reason">驳回原因：{{ item.reject_reason }}</span>
            </div>
          </article>
        </div>

        <div v-else class="mt-6">
          <EmptyState title="输入回执码后查看投稿记录" description="会显示文件标题、审核状态、上传时间和历史下载量。" />
        </div>
      </SurfaceCard>
    </section>
  </div>
</template>
