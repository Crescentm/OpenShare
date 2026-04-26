<script setup lang="ts">
import {
  computed,
  defineAsyncComponent,
  nextTick,
  onMounted,
  ref,
  watch,
} from "vue";
import { useRoute, useRouter } from "vue-router";

import PublicFileDeleteDialog from "../../components/public/file-detail/PublicFileDeleteDialog.vue";
import PublicFileDescriptionEditor from "../../components/public/file-detail/PublicFileDescriptionEditor.vue";
import PublicFileDetailHeader from "../../components/public/file-detail/PublicFileDetailHeader.vue";
import PublicFileFeedbackDialog from "../../components/public/file-detail/PublicFileFeedbackDialog.vue";
import PublicFileMetadataDetails from "../../components/public/file-detail/PublicFileMetadataDetails.vue";
import type {
  FileMetadataRow,
  PublicFileDetailResponse,
} from "../../components/public/file-detail/types";
import SurfaceCard from "../../components/ui/SurfaceCard.vue";
import { HttpError, httpClient } from "../../lib/http/client";
import { readApiError } from "../../lib/http/helpers";
import {
  ensureSessionReceiptCode,
  readStoredReceiptCode,
} from "../../lib/receiptCode";
import { hasAdminPermission } from "../../lib/admin/session";
import { renderSimpleMarkdown } from "../../lib/markdown";

const route = useRoute();
const router = useRouter();
const FilePreview = defineAsyncComponent(
  () => import("../../components/resources/FilePreview.vue"),
);
const detail = ref<PublicFileDetailResponse | null>(null);
const loading = ref(false);
const error = ref("");
const message = ref("");
const saveError = ref("");
const saving = ref(false);
const editFileName = ref("");
const editDescription = ref("");
const descriptionEditorOpen = ref(false);
const canManageResourceDescriptions = ref(false);
const deleteDialogOpen = ref(false);
const deletePassword = ref("");
const deleteSubmitting = ref(false);
const deleteError = ref("");
const feedbackModalOpen = ref(false);
const feedbackSuccessModalOpen = ref(false);
const feedbackDescription = ref("");
const feedbackSubmitting = ref(false);
const feedbackMessage = ref("");
const feedbackError = ref("");
const currentReceiptCode = ref("");
const previewSectionRef = ref<HTMLElement | null>(null);
const fileID = computed(() => String(route.params.fileID ?? ""));
const downloadURL = computed(
  () => `/api/public/files/${encodeURIComponent(fileID.value)}/download`,
);
const descriptionHTML = computed(() =>
  renderSimpleMarkdown(detail.value?.description ?? ""),
);
const hasDescription = computed(() => Boolean(descriptionHTML.value));
const feedbackSubmitDisabled = computed(
  () => feedbackSubmitting.value || !feedbackDescription.value.trim(),
);
const metadataRows = computed<FileMetadataRow[]>(() => {
  if (!detail.value) {
    return [];
  }

  return [
    { label: "文件名", value: detail.value.name },
    { label: "所属文件夹", value: detail.value.path || "主页根目录" },
    { label: "下载量", value: String(detail.value.download_count) },
    { label: "文件大小", value: formatSize(detail.value.size) },
    { label: "更新时间", value: formatDate(detail.value.uploaded_at) },
    { label: "MIME", value: detail.value.mime_type || "未知" },
  ];
});
const editorDirty = computed(() => {
  if (!detail.value) {
    return false;
  }

  return (
    editFileName.value.trim() !== detail.value.name ||
    editDescription.value.trim() !== (detail.value.description ?? "")
  );
});

function centerPreviewSection() {
  previewSectionRef.value?.scrollIntoView({
    block: "center",
  });
}

onMounted(() => {
  void Promise.all([
    loadDetail(),
    loadAdminPermission(),
    syncSessionReceiptCode(),
  ]);
});

watch(fileID, () => {
  void Promise.all([
    loadDetail(),
    loadAdminPermission(),
    syncSessionReceiptCode(),
  ]);
});

watch(
  detail,
  async (currentDetail) => {
    if (!currentDetail) {
      return;
    }

    await nextTick();
    centerPreviewSection();
  },
);

async function loadDetail() {
  loading.value = true;
  error.value = "";
  detail.value = null;
  try {
    detail.value = await httpClient.get<PublicFileDetailResponse>(
      `/public/files/${encodeURIComponent(fileID.value)}`,
    );
    if (detail.value) {
      editFileName.value = detail.value.name;
      editDescription.value = detail.value.description;
    }
  } catch (err: unknown) {
    if (err instanceof HttpError) {
      if (err.status === 404) {
        error.value = "File not found or not public.";
      } else if (err.status === 410) {
        error.value = "File has been permanently deleted.";
      } else if (err.status === 403) {
        error.value = "No permission to access this file.";
      } else {
        error.value = `Failed to load file details (HTTP ${err.status}).`;
      }
    } else {
      error.value = "Failed to load file details.";
    }
  } finally {
    loading.value = false;
  }
}

async function loadAdminPermission() {
  canManageResourceDescriptions.value = await hasAdminPermission(
    "resource_moderation",
  );
}

function openDescriptionEditor() {
  editFileName.value = detail.value?.name ?? "";
  editDescription.value = detail.value?.description ?? "";
  saveError.value = "";
  message.value = "";
  descriptionEditorOpen.value = true;
}

function closeDescriptionEditor() {
  descriptionEditorOpen.value = false;
  saving.value = false;
  saveError.value = "";
  editFileName.value = detail.value?.name ?? "";
  editDescription.value = detail.value?.description ?? "";
}

function openDeleteDialog() {
  deletePassword.value = "";
  deleteError.value = "";
  deleteDialogOpen.value = true;
}

function openFeedbackModal() {
  feedbackDescription.value = "";
  feedbackError.value = "";
  feedbackMessage.value = "";
  feedbackModalOpen.value = true;
  void syncSessionReceiptCode();
}

function closeFeedbackModal() {
  feedbackModalOpen.value = false;
}

function closeFeedbackSuccessModal() {
  feedbackSuccessModalOpen.value = false;
}

function closeDeleteDialog() {
  deleteDialogOpen.value = false;
  deletePassword.value = "";
  deleteError.value = "";
  deleteSubmitting.value = false;
}

async function saveDescription() {
  if (!detail.value || !editorDirty.value) return;
  const normalizedName = editFileName.value.trim();
  if (!normalizedName) {
    saveError.value = "请输入有效的文件名。";
    return;
  }
  saving.value = true;
  saveError.value = "";
  message.value = "";
  try {
    await httpClient.request(
      `/admin/resources/files/${encodeURIComponent(detail.value.id)}`,
      {
        method: "PUT",
        body: {
          name: normalizedName,
          description: editDescription.value.trim(),
        },
      },
    );
    message.value = "文件信息已更新。";
    await loadDetail();
    descriptionEditorOpen.value = false;
  } catch (err: unknown) {
    saveError.value = readApiError(err, "更新文件简介失败。");
  } finally {
    saving.value = false;
  }
}

async function confirmDeleteFile() {
  if (!detail.value) {
    return;
  }
  if (!deletePassword.value.trim()) {
    deleteError.value = "请输入当前管理员密码。";
    return;
  }

  deleteSubmitting.value = true;
  deleteError.value = "";
  try {
    await httpClient.request(
      `/admin/resources/files/${encodeURIComponent(detail.value.id)}`,
      {
        method: "DELETE",
        body: { password: deletePassword.value },
      },
    );
    closeDeleteDialog();
    goBack();
  } catch (err: unknown) {
    deleteError.value = readApiError(err, "删除文件失败。");
  } finally {
    deleteSubmitting.value = false;
  }
}

async function submitFeedback() {
  if (!detail.value || !feedbackDescription.value.trim()) {
    return;
  }

  feedbackSubmitting.value = true;
  feedbackMessage.value = "";
  feedbackError.value = "";
  try {
    const response = await httpClient.post<{ receipt_code: string }>(
      "/public/feedback",
      {
        file_id: detail.value.id,
        folder_id: "",
        description: feedbackDescription.value.trim(),
      },
    );
    feedbackMessage.value = `反馈已提交，请保存回执码 ${response.receipt_code}。`;
    currentReceiptCode.value = response.receipt_code;
    window.sessionStorage.setItem(
      "openshare_receipt_code",
      response.receipt_code,
    );
    closeFeedbackModal();
    feedbackSuccessModalOpen.value = true;
  } catch (err: unknown) {
    if (err instanceof HttpError && err.status === 400) {
      feedbackError.value = "请填写问题说明。";
    } else if (err instanceof HttpError && err.status === 404) {
      feedbackError.value = "目标不存在或已删除。";
    } else {
      feedbackError.value = "提交反馈失败。";
    }
  } finally {
    feedbackSubmitting.value = false;
  }
}

async function syncSessionReceiptCode() {
  try {
    currentReceiptCode.value = await ensureSessionReceiptCode();
  } catch {
    currentReceiptCode.value = readStoredReceiptCode();
  }
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function formatSize(size: number) {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function goBack() {
  const folderID = detail.value?.folder_id?.trim() ?? "";
  if (folderID) {
    void router.push({ name: "public-home", query: { folder: folderID } });
    return;
  }
  void router.push({ name: "public-home" });
}

function downloadFile() {
  const link = document.createElement("a");
  link.href = downloadURL.value;
  link.rel = "noopener";
  document.body.appendChild(link);
  link.click();
  link.remove();

  if (detail.value) {
    detail.value = {
      ...detail.value,
      download_count: detail.value.download_count + 1,
    };
  }
}
</script>

<template>
  <section class="app-container py-5 sm:py-6 lg:py-8">
    <div class="mx-auto w-full max-w-6xl space-y-5">
      <SurfaceCard>
        <p v-if="loading" class="text-sm text-slate-500">加载中…</p>

        <div v-else-if="error" class="space-y-4">
          <p
            class="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700"
          >
            {{ error }}
          </p>
          <div class="flex flex-col gap-3 sm:flex-row">
            <button
              type="button"
              class="btn-secondary w-full sm:w-auto"
              @click="goBack"
            >
              返回上一页
            </button>
            <button
              type="button"
              class="btn-primary w-full sm:w-auto"
              @click="$router.push({ name: 'public-home' })"
            >
              返回首页
            </button>
          </div>
        </div>

        <template v-else-if="detail">
          <p
            v-if="message"
            class="mb-4 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700"
          >
            {{ message }}
          </p>

          <section class="space-y-4">
            <PublicFileDetailHeader
              :can-manage-resource-descriptions="canManageResourceDescriptions"
              :detail="detail"
              :formatted-date="formatDate(detail.uploaded_at)"
              :formatted-size="formatSize(detail.size)"
              @download="downloadFile"
              @edit="openDescriptionEditor"
              @feedback="openFeedbackModal"
              @go-back="goBack"
              @open-delete="openDeleteDialog"
            />

            <div ref="previewSectionRef">
              <FilePreview
                :file-id="detail.id"
                :file-name="detail.name"
                :extension="detail.extension"
                :mime-type="detail.mime_type"
                :size="detail.size"
                @preview-ready="centerPreviewSection"
              />
            </div>

            <div
              v-if="hasDescription"
              class="rounded-2xl border border-slate-200 bg-white px-4 py-4 sm:px-5"
            >
              <div class="markdown-content" v-html="descriptionHTML" />
            </div>

            <PublicFileMetadataDetails :metadata-rows="metadataRows" />
          </section>
        </template>
      </SurfaceCard>
    </div>

    <PublicFileDeleteDialog
      v-if="detail"
      :delete-error="deleteError"
      :delete-password="deletePassword"
      :delete-submitting="deleteSubmitting"
      :file-name="detail.name"
      :open="deleteDialogOpen"
      @close="closeDeleteDialog"
      @confirm="confirmDeleteFile"
      @update:delete-password="deletePassword = $event"
    />

    <PublicFileFeedbackDialog
      v-if="detail"
      :current-receipt-code="currentReceiptCode"
      :feedback-description="feedbackDescription"
      :feedback-error="feedbackError"
      :feedback-message="feedbackMessage"
      :feedback-submit-disabled="feedbackSubmitDisabled"
      :feedback-submitting="feedbackSubmitting"
      :file-name="detail.name"
      :open="feedbackModalOpen"
      :success-open="feedbackSuccessModalOpen"
      @close="closeFeedbackModal"
      @close-success="closeFeedbackSuccessModal"
      @submit="submitFeedback"
      @update:feedback-description="feedbackDescription = $event"
    />

    <PublicFileDescriptionEditor
      :can-manage-resource-descriptions="canManageResourceDescriptions"
      :description="editDescription"
      :file-name="editFileName"
      :open="descriptionEditorOpen"
      :save-error="saveError"
      :saving="saving"
      :submit-disabled="!editorDirty"
      @close="closeDescriptionEditor"
      @save="saveDescription"
      @update:description="editDescription = $event"
      @update:file-name="editFileName = $event"
    />
  </section>
</template>
