<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  ChevronRight,
  Home,
} from "lucide-vue-next";

import InfoPanelCard from "../../components/shared/InfoPanelCard.vue";
import PublicAnnouncementDetailModal from "../../components/public/home/PublicAnnouncementDetailModal.vue";
import PublicAnnouncementListModal from "../../components/public/home/PublicAnnouncementListModal.vue";
import PublicDirectoryCards from "../../components/public/home/PublicDirectoryCards.vue";
import PublicDirectoryTable from "../../components/public/home/PublicDirectoryTable.vue";
import PublicDirectoryToolbar from "../../components/public/home/PublicDirectoryToolbar.vue";
import PublicDeleteResourceDialog from "../../components/public/home/PublicDeleteResourceDialog.vue";
import PublicFolderInfoPanel from "../../components/public/home/PublicFolderInfoPanel.vue";
import PublicFolderDescriptionEditor from "../../components/public/home/PublicFolderDescriptionEditor.vue";
import PublicHomeFeedbackDialog from "../../components/public/home/PublicHomeFeedbackDialog.vue";
import PublicHomeSidebarDetailModal from "../../components/public/home/PublicHomeSidebarDetailModal.vue";
import PublicUploadDialog from "../../components/public/home/PublicUploadDialog.vue";
import type {
  DirectoryRow,
  FolderDetailResponse,
  PublicFileItem,
  PublicFolderItem,
  SearchResultResponse,
} from "../../components/public/home/types";
import {
  extractExtension,
  usePublicHomeDirectoryRows,
} from "../../components/public/home/usePublicHomeDirectoryRows";
import { usePublicHomeFeedbackState } from "../../components/public/home/usePublicHomeFeedbackState";
import { usePublicHomeFolderAdminState } from "../../components/public/home/usePublicHomeFolderAdminState";
import { usePublicHomeReadmePreview } from "../../components/public/home/usePublicHomeReadmePreview";
import { usePublicHomeSidebar } from "../../components/public/home/usePublicHomeSidebar";
import { usePublicHomeUploadState } from "../../components/public/home/usePublicHomeUploadState";
import SearchSection from "../../components/resources/SearchSection.vue";
import { HttpError, httpClient } from "../../lib/http/client";
import { readApiError } from "../../lib/http/helpers";
import {
  ensureSessionReceiptCode,
  readStoredReceiptCode,
} from "../../lib/receiptCode";
import { hasAdminPermission } from "../../lib/admin/session";
import { renderSimpleMarkdown } from "../../lib/markdown";
import {
  collectDroppedEntries,
  normalizeFiles,
} from "../../lib/uploads/fileDrop";

const route = useRoute();
const router = useRouter();

const transientWarning = ref("");
const transientWarningTimer = ref<number | null>(null);
const downloadTimestamps = ref<number[]>([]);
const transientWarningLeaving = ref(false);
const currentReceiptCode = ref("");

const loading = ref(false);
const error = ref("");
const actionMessage = ref("");
const actionError = ref("");
const batchDownloadSubmitting = ref(false);
const folders = ref<PublicFolderItem[]>([]);
const files = ref<PublicFileItem[]>([]);
const searchInput = ref("");
const searchKeyword = ref("");
const searchLoading = ref(false);
const searchError = ref("");
const searchRows = ref<DirectoryRow[]>([]);
const breadcrumbs = ref<Array<{ id: string; name: string }>>([]);
const currentFolderDetail = ref<FolderDetailResponse | null>(null);
const canManageResourceDescriptions = ref(false);
const currentFolderID = computed(() => {
  const raw = route.query.folder;
  return typeof raw === "string" && raw.trim() ? raw.trim() : "";
});
const canUploadToCurrentFolder = computed(
  () => currentFolderID.value.length > 0,
);
const rootViewLocked = computed(() => route.query.root === "1");

function buildPublicFileDownloadURL(fileID: string) {
  return `/api/public/files/${encodeURIComponent(fileID)}/download`;
}

function buildPublicFilePreviewURL(fileID: string, view: "inline" | "text") {
  const query = new URLSearchParams({ view });
  return `/api/public/files/${encodeURIComponent(fileID)}/preview?${query.toString()}`;
}

const {
  announcementDetail,
  announcementListOpen,
  announcements,
  closeAnnouncementDetail,
  closeAnnouncementList,
  closeSidebarDetailModal,
  hotDownloads,
  latestTitles,
  loadAnnouncements,
  loadHotDownloads,
  loadLatestTitles,
  openAnnouncementDetail,
  openAnnouncementList,
  openHotDownloadsModal,
  openLatestItemsModal,
  openSidebarDetailItem,
  recentAnnouncements,
  returnToAnnouncementList,
  sidebarDetailModal,
} = usePublicHomeSidebar(openFile, syncBodyScrollLock);

const {
  clearUploadEntries,
  closeUploadModal: closeUploadDialogState,
  closeUploadSuccessModal: closeUploadSuccessState,
  openUploadModal,
  resetUploadForm,
  uploadState,
} = usePublicHomeUploadState();

const {
  closeFeedbackModal: closeFeedbackDialogState,
  closeFeedbackSuccessModal: closeFeedbackSuccessState,
  feedbackState,
  feedbackSubmitDisabled,
  openFeedbackModal: openFeedbackDialogState,
} = usePublicHomeFeedbackState();

const {
  isCurrentReadmeRequest,
  loadReadmePreview,
  nextReadmePreviewRequestID,
  readmePreviewError,
  readmePreviewHTML,
  readmePreviewLoading,
  readmePreviewName,
  resetReadmePreview,
} = usePublicHomeReadmePreview(currentFolderID, buildPublicFilePreviewURL);

const {
  allVisibleRowsSelected,
  clearSelection,
  fileIconComponent,
  hasSelectedRows,
  isRowSelected,
  restoreDisplayPreferences,
  selectedRows,
  setSortDirection,
  setSortMode,
  setViewMode,
  sortDirection,
  sortMenuOpen,
  sortMode,
  sortedRows,
  toggleRowSelection,
  toggleSelectAllVisibleRows,
  viewMenuOpen,
  viewMode,
} = usePublicHomeDirectoryRows({
  currentFolderID,
  files,
  folders,
  formatDateTime,
  formatSize,
  searchKeyword,
  searchRows,
});

const {
  closeFolderDescriptionEditor: closeFolderDescriptionEditorState,
  folderAdminState,
  folderEditorDirty,
  openDeleteFolderDialog: openDeleteFolderDialogState,
  openFolderDescriptionEditor: openFolderDescriptionEditorState,
  resetDeleteDialog,
  syncFolderDrafts,
} = usePublicHomeFolderAdminState(currentFolderDetail);

const currentFolderDescriptionHTML = computed(() => {
  const desc = currentFolderDetail.value?.description;
  if (desc) {
    return renderSimpleMarkdown(desc);
  }
  if (readmePreviewHTML.value) {
    return readmePreviewHTML.value;
  }
  return "";
});
const currentFolderStats = computed(() => {
  if (!currentFolderDetail.value) {
    return [];
  }

  return [
    { label: "文件夹名", value: currentFolderDetail.value.name },
    {
      label: "下载量",
      value: String(currentFolderDetail.value.download_count ?? 0),
    },
    {
      label: "文件数",
      value: `${currentFolderDetail.value.file_count ?? 0} 个文件`,
    },
    {
      label: "文件夹大小",
      value: formatSize(currentFolderDetail.value.total_size ?? 0),
    },
    {
      label: "更新时间",
      value: formatDateTime(currentFolderDetail.value.updated_at),
    },
  ];
});
const canGoUp = computed(() => currentFolderID.value.length > 0);
const backButtonLabel = computed(() =>
  searchKeyword.value ? "返回所在目录" : "返回上一级",
);
const canUseBackButton = computed(
  () => searchKeyword.value.length > 0 || canGoUp.value,
);

function downloadResource(row: DirectoryRow) {
  actionMessage.value = "";
  actionError.value = "";
  if (!allowDownloadRequest()) {
    showTransientWarning("下载请求过于频繁，请稍后再试。");
    return;
  }

  const link = document.createElement("a");
  link.href = row.downloadURL;
  link.rel = "noopener";
  document.body.appendChild(link);
  link.click();
  link.remove();

  applyDownloadCountUpdate(row);
  void loadHotDownloads();
}

async function downloadSelectedResources() {
  if (!hasSelectedRows.value || batchDownloadSubmitting.value) {
    return;
  }

  actionMessage.value = "";
  actionError.value = "";
  if (!allowDownloadRequest()) {
    showTransientWarning("下载请求过于频繁，请稍后再试。");
    return;
  }

  const fileIDs = selectedRows.value
    .filter((row) => row.kind === "file")
    .map((row) => row.id);
  const folderIDs = selectedRows.value
    .filter((row) => row.kind === "folder")
    .map((row) => row.id);

  batchDownloadSubmitting.value = true;
  try {
    const response = await fetch("/api/public/resources/batch-download", {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/zip",
      },
      body: JSON.stringify({
        file_ids: fileIDs,
        folder_ids: folderIDs,
      }),
    });

    if (!response.ok) {
      throw new Error("batch download failed");
    }

    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "openshare-selection.zip";
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);

    for (const row of selectedRows.value) {
      applyDownloadCountUpdate(row);
    }
    await loadHotDownloads();
    clearSelection();
  } catch (err: unknown) {
    actionError.value = readApiError(err, "批量下载失败。");
  } finally {
    batchDownloadSubmitting.value = false;
  }
}

function syncBodyScrollLock() {
  const shouldLock = Boolean(
    announcementDetail.value ||
    announcementListOpen.value ||
    sidebarDetailModal.value ||
    uploadState.modalOpen ||
    uploadState.successOpen ||
    feedbackState.modalOpen ||
    feedbackState.successOpen ||
    folderAdminState.editorOpen ||
    folderAdminState.deleteTarget,
  );
  document.body.style.overflow = shouldLock ? "hidden" : "";
}

onMounted(async () => {
  restoreDisplayPreferences();
  currentReceiptCode.value = await syncSessionReceiptCode();
  await Promise.all([
    loadAnnouncements(),
    loadHotDownloads(),
    loadLatestTitles(),
    loadDirectory(),
    loadAdminPermission(),
  ]);
});

onBeforeUnmount(() => {
  if (transientWarningTimer.value !== null) {
    window.clearTimeout(transientWarningTimer.value);
  }
  document.body.style.overflow = "";
});

watch(currentFolderID, () => {
  clearSearchState();
  void loadDirectory();
});

async function loadDirectory() {
  const requestID = nextReadmePreviewRequestID();
  loading.value = true;
  error.value = "";
  actionMessage.value = "";
  actionError.value = "";
  resetReadmePreview();
  try {
    const directoryParams = new URLSearchParams();
    if (currentFolderID.value) {
      directoryParams.set("parent_id", currentFolderID.value);
    }

    const requests: Array<Promise<unknown>> = [
      httpClient.get<{ items: PublicFolderItem[] }>(
        `/public/folders${directoryParams.toString() ? `?${directoryParams.toString()}` : ""}`,
      ),
    ];

    if (currentFolderID.value) {
      const folderParams = new URLSearchParams({
        page: "1",
        page_size: "100",
        sort: "name_asc",
      });
      requests.push(
        httpClient.get<{ items: PublicFileItem[] }>(
          `/public/folders/${encodeURIComponent(currentFolderID.value)}/files?${folderParams.toString()}`,
        ),
      );
    }

    if (currentFolderID.value) {
      requests.push(
        httpClient.get<FolderDetailResponse>(
          `/public/folders/${encodeURIComponent(currentFolderID.value)}`,
        ),
      );
    }

    const [folderResponse, fileResponse, folderDetail] =
      await Promise.all(requests);
    if (!isCurrentReadmeRequest(requestID)) {
      return;
    }
    folders.value =
      (folderResponse as { items: PublicFolderItem[] }).items ?? [];
    files.value = currentFolderID.value
      ? ((fileResponse as { items: PublicFileItem[] } | undefined)?.items ?? [])
      : [];

    if (
      !currentFolderID.value &&
      !rootViewLocked.value &&
      folders.value.length === 1
    ) {
      void router.replace({
        name: "public-home",
        query: { folder: folders.value[0].id },
      });
      return;
    }

    if (folderDetail) {
      const detail = folderDetail as FolderDetailResponse;
      currentFolderDetail.value = detail;
      syncFolderDrafts();
      breadcrumbs.value = detail.breadcrumbs ?? [];
    } else {
      currentFolderDetail.value = null;
      syncFolderDrafts();
      breadcrumbs.value = [];
    }

    await loadReadmePreview(files.value, requestID);
  } catch (err: unknown) {
    folders.value = [];
    files.value = [];
    breadcrumbs.value = [];
    currentFolderDetail.value = null;
    syncFolderDrafts();
    resetReadmePreview();
    if (err instanceof HttpError && err.status === 404) {
      error.value = "目录不存在或未公开。";
    } else {
      error.value = "加载目录失败。";
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

function openRoot() {
  clearSearchState();
  void router.push({ name: "public-home", query: { root: "1" } });
}

function goUpOneLevel() {
  if (searchKeyword.value) {
    clearSearchState();
    return;
  }
  if (!currentFolderID.value) {
    return;
  }
  clearSearchState();
  const parent = breadcrumbs.value.at(-2);
  if (parent) {
    void router.push({ name: "public-home", query: { folder: parent.id } });
    return;
  }
  openRoot();
}

function openFolder(folderID: string) {
  clearSearchState();
  void router.push({ name: "public-home", query: { folder: folderID } });
}

function openFile(fileID: string) {
  if (searchKeyword.value) {
    clearSearchState();
  }
  void router.push({ name: "public-file-detail", params: { fileID } });
}

function downloadCurrentFolder() {
  if (!currentFolderDetail.value) {
    return;
  }
  downloadResource({
    id: currentFolderDetail.value.id,
    kind: "folder",
    name: currentFolderDetail.value.name,
    extension: "",
    description: "",
    downloadCount: currentFolderDetail.value.download_count ?? 0,
    fileCount: currentFolderDetail.value.file_count ?? 0,
    sizeText: formatSize(currentFolderDetail.value.total_size ?? 0),
    updatedAt: formatDateTime(currentFolderDetail.value.updated_at),
    downloadURL: `/api/public/folders/${encodeURIComponent(currentFolderDetail.value.id)}/download`,
  });
}

function openDeleteFolderDialog() {
  openDeleteFolderDialogState();
}

function closeDeleteResourceDialog() {
  resetDeleteDialog();
}

async function confirmDeleteResource() {
  if (!folderAdminState.deleteTarget) {
    return;
  }
  if (!folderAdminState.deletePassword.trim()) {
    folderAdminState.deleteError = "请输入当前管理员密码。";
    return;
  }

  folderAdminState.deleteSubmitting = true;
  folderAdminState.deleteError = "";
  try {
    await httpClient.request(
      `/admin/resources/folders/${encodeURIComponent(folderAdminState.deleteTarget.id)}`,
      {
        method: "DELETE",
        body: { password: folderAdminState.deletePassword },
      },
    );
    const parentID = currentFolderDetail.value?.parent_id ?? "";
    closeDeleteResourceDialog();
    actionMessage.value = `文件夹 ${currentFolderDetail.value?.name ?? ""} 已删除。`;
    clearSearchState();
    if (parentID) {
      await router.push({ name: "public-home", query: { folder: parentID } });
    } else {
      await router.push({ name: "public-home", query: { root: "1" } });
    }
  } catch (err: unknown) {
    folderAdminState.deleteError = readApiError(err, "删除文件夹失败。");
  } finally {
    folderAdminState.deleteSubmitting = false;
  }
}

async function runSearch(keyword: string) {
  const normalizedKeyword = keyword.trim();
  if (!normalizedKeyword) {
    clearSearchState();
    return;
  }

  searchInput.value = normalizedKeyword;
  searchKeyword.value = normalizedKeyword;
  searchLoading.value = true;
  searchError.value = "";
  try {
    const query = new URLSearchParams({
      q: normalizedKeyword,
      page: "1",
      page_size: "50",
    });
    if (currentFolderID.value) {
      query.set("folder_id", currentFolderID.value);
    }
    const response = await httpClient.get<SearchResultResponse>(
      `/public/search?${query.toString()}`,
    );
    searchRows.value = response.items.map((item) => ({
      id: item.id,
      kind: item.entity_type,
      name: item.name,
      extension:
        item.entity_type === "file"
          ? item.extension || extractExtension(item.name)
          : "",
      description: "",
      downloadCount: item.download_count ?? 0,
      fileCount: 0,
      sizeText: item.entity_type === "file" ? formatSize(item.size ?? 0) : "-",
      updatedAt: item.uploaded_at ? formatDateTime(item.uploaded_at) : "-",
      downloadURL:
        item.entity_type === "file"
          ? buildPublicFileDownloadURL(item.id)
          : `/api/public/folders/${encodeURIComponent(item.id)}/download`,
    }));
  } catch (err: unknown) {
    searchRows.value = [];
    searchError.value = readApiError(err, "搜索失败。");
  } finally {
    searchLoading.value = false;
  }
}

function clearSearchState() {
  searchInput.value = "";
  searchKeyword.value = "";
  searchLoading.value = false;
  searchError.value = "";
  searchRows.value = [];
  clearSelection();
}

function openUpload() {
  if (!canUploadToCurrentFolder.value) {
    showTransientWarning("请先进入一个目录后再上传。");
    return;
  }
  openUploadModal();
  void syncSessionReceiptCode();
  syncBodyScrollLock();
}

function closeUploadModal() {
  closeUploadDialogState();
  syncBodyScrollLock();
}

function closeUploadSuccessModal() {
  closeUploadSuccessState();
  syncBodyScrollLock();
}

function onUploadFileChange(event: Event) {
  const target = event.target as HTMLInputElement;
  uploadState.form.entries = normalizeFiles(
    Array.from(target.files ?? []).slice(0, 1),
  );
  if (
    uploadState.form.entries.length === 0 &&
    (target.files?.length ?? 0) > 0
  ) {
    uploadState.error = "已自动忽略 .DS_Store，请重新选择可上传文件。";
  }
}

function onUploadDragEnter() {
  uploadState.dropActive = true;
}

function onUploadDragLeave(event: DragEvent) {
  const currentTarget = event.currentTarget as HTMLElement | null;
  if (
    currentTarget &&
    event.relatedTarget instanceof Node &&
    currentTarget.contains(event.relatedTarget)
  ) {
    return;
  }
  uploadState.dropActive = false;
}

async function onUploadDrop(event: DragEvent) {
  event.preventDefault();
  uploadState.dropActive = false;
  uploadState.collecting = true;
  uploadState.error = "";
  try {
    const entries = await collectDroppedEntries(event);
    uploadState.form.entries = entries;
    if (entries.length === 0 && (event.dataTransfer?.files.length ?? 0) > 0) {
      uploadState.error = "检测到的内容仅包含 .DS_Store，已自动忽略。";
    }
  } catch {
    uploadState.error = "解析拖拽内容失败，请重试。";
  } finally {
    uploadState.collecting = false;
  }
}

async function submitUpload() {
  if (uploadState.form.entries.length === 0) {
    uploadState.error = "请选择文件，或直接拖入多文件/文件夹。";
    return;
  }

  uploadState.submitting = true;
  uploadState.error = "";
  uploadState.message = "";
  try {
    const formData = new FormData();
    formData.set("folder_id", currentFolderID.value);
    formData.set("description", uploadState.form.description.trim());
    formData.set(
      "manifest",
      JSON.stringify(
        uploadState.form.entries.map((entry) => ({
          relative_path: entry.relativePath,
        })),
      ),
    );
    uploadState.form.entries.forEach((entry) => {
      formData.append("files", entry.file, entry.file.name);
    });
    const response = await httpClient.post<{
      receipt_code: string;
      item_count: number;
      status: string;
    }>("/public/submissions", formData);
    uploadState.message =
      response.status === "approved"
        ? `已上传 ${response.item_count} 个文件，请保存回执码 ${response.receipt_code}。`
        : `已提交 ${response.item_count} 个文件进入审核，请保存回执码 ${response.receipt_code}。`;
    window.sessionStorage.setItem(
      "openshare_receipt_code",
      response.receipt_code,
    );
    currentReceiptCode.value = response.receipt_code;
    resetUploadForm();
    clearUploadEntries();
    if (response.status === "approved") {
      await loadDirectory();
    }
    closeUploadModal();
    uploadState.successOpen = true;
    syncBodyScrollLock();
  } catch (err) {
    if (err instanceof HttpError && err.status === 400) {
      uploadState.error = "上传参数无效。";
    } else if (err instanceof HttpError && err.status === 409) {
      uploadState.error = "提交上传失败，请检查名称或者联系管理员";
    } else {
      uploadState.error = "提交上传失败。";
    }
  } finally {
    uploadState.submitting = false;
  }
}

function applyDownloadCountUpdate(row: DirectoryRow) {
  if (row.kind === "file") {
    files.value = files.value.map((item) => {
      if (item.id !== row.id) {
        return item;
      }
      return {
        ...item,
        download_count: item.download_count + 1,
      };
    });
    return;
  }

  folders.value = folders.value.map((item) => {
    if (item.id !== row.id) {
      return item;
    }
    return {
      ...item,
      download_count: item.download_count + Math.max(1, item.file_count),
    };
  });
}

function allowDownloadRequest() {
  const now = Date.now();
  const windowMs = 10_000;
  const limit = 10;
  downloadTimestamps.value = downloadTimestamps.value.filter(
    (timestamp) => now - timestamp < windowMs,
  );
  if (downloadTimestamps.value.length >= limit) {
    return false;
  }
  downloadTimestamps.value.push(now);
  return true;
}

function showTransientWarning(message: string) {
  transientWarning.value = message;
  transientWarningLeaving.value = false;
  if (transientWarningTimer.value !== null) {
    window.clearTimeout(transientWarningTimer.value);
  }
  transientWarningTimer.value = window.setTimeout(() => {
    transientWarningLeaving.value = true;
    transientWarningTimer.value = window.setTimeout(() => {
      transientWarning.value = "";
      transientWarningLeaving.value = false;
      transientWarningTimer.value = null;
    }, 1200);
  }, 400);
}

function openFeedbackModal(target: {
  id: string;
  type: "file" | "folder";
  name: string;
}) {
  openFeedbackDialogState(target);
  void syncSessionReceiptCode();
  syncBodyScrollLock();
}

function closeFeedbackModal() {
  closeFeedbackDialogState();
  syncBodyScrollLock();
}

function closeFeedbackSuccessModal() {
  closeFeedbackSuccessState();
  syncBodyScrollLock();
}

function openFolderDescriptionEditor() {
  openFolderDescriptionEditorState();
  syncBodyScrollLock();
}

function closeFolderDescriptionEditor() {
  closeFolderDescriptionEditorState();
  syncBodyScrollLock();
}

async function saveFolderDescription() {
  if (!currentFolderDetail.value || !folderEditorDirty.value) {
    return;
  }

  folderAdminState.saving = true;
  folderAdminState.error = "";
  try {
    await httpClient.request(
      `/admin/resources/folders/${encodeURIComponent(currentFolderDetail.value.id)}`,
      {
        method: "PUT",
        body: {
          name: folderAdminState.nameDraft.trim(),
          description: folderAdminState.descriptionDraft.trim(),
        },
      },
    );
    currentFolderDetail.value = {
      ...currentFolderDetail.value,
      name: folderAdminState.nameDraft.trim(),
      description: folderAdminState.descriptionDraft.trim(),
    };
    breadcrumbs.value = breadcrumbs.value.map((item, index) =>
      index === breadcrumbs.value.length - 1
        ? { ...item, name: folderAdminState.nameDraft.trim() }
        : item,
    );
    folderAdminState.editorOpen = false;
    syncBodyScrollLock();
  } catch (err: unknown) {
    folderAdminState.error = readApiError(err, "更新文件夹简介失败。");
  } finally {
    folderAdminState.saving = false;
  }
}

async function submitFeedback() {
  if (!feedbackState.target) {
    return;
  }
  if (!feedbackState.description.trim()) {
    feedbackState.error = "请填写问题说明。";
    return;
  }

  feedbackState.submitting = true;
  feedbackState.message = "";
  feedbackState.error = "";
  try {
    const response = await httpClient.post<{ receipt_code: string }>(
      "/public/feedback",
      {
        file_id:
          feedbackState.target.type === "file" ? feedbackState.target.id : "",
        folder_id:
          feedbackState.target.type === "folder" ? feedbackState.target.id : "",
        description: feedbackState.description.trim(),
      },
    );
    feedbackState.message = `反馈已提交，请保存回执码 ${response.receipt_code}。`;
    window.sessionStorage.setItem(
      "openshare_receipt_code",
      response.receipt_code,
    );
    currentReceiptCode.value = response.receipt_code;
    closeFeedbackModal();
    feedbackState.successOpen = true;
    syncBodyScrollLock();
  } catch (err: unknown) {
    if (err instanceof HttpError && err.status === 400) {
      feedbackState.error = "请填写问题说明。";
    } else if (err instanceof HttpError && err.status === 404) {
      feedbackState.error = "目标不存在或已删除。";
    } else {
      feedbackState.error = "提交反馈失败。";
    }
  } finally {
    feedbackState.submitting = false;
  }
}

function formatSize(size: number) {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(2)} KB`;
  if (size < 1024 * 1024 * 1024)
    return `${(size / (1024 * 1024)).toFixed(2)} MB`;
  return `${(size / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  }).format(new Date(value));
}

async function syncSessionReceiptCode() {
  try {
    const receiptCode = await ensureSessionReceiptCode();
    currentReceiptCode.value = receiptCode || readStoredReceiptCode();
    return currentReceiptCode.value;
  } catch {
    currentReceiptCode.value = readStoredReceiptCode();
    return currentReceiptCode.value;
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="transientWarning"
      class="fixed inset-0 z-[130] flex items-center justify-center px-4"
    >
      <div
        class="rounded-2xl border border-rose-200 bg-white px-4 py-3 text-sm text-rose-700 shadow-lg shadow-rose-100/70"
        :class="
          transientWarningLeaving
            ? 'animate-[warning-fade-out_1.2s_ease_forwards]'
            : 'animate-[warning-fade-in_0.18s_ease-out_forwards]'
        "
      >
        {{ transientWarning }}
      </div>
    </div>
  </Teleport>

  <main class="app-container py-6 sm:py-8 lg:py-10">
    <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_248px]">
      <section class="order-1 min-w-0">
        <div class="panel overflow-hidden">
          <div
            class="border-b border-slate-200 px-4 py-3 sm:px-6 dark:border-slate-800"
          >
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div class="min-w-0 max-w-full overflow-x-auto">
                <div
                  class="flex min-w-max items-center gap-2 text-sm text-slate-500 dark:text-slate-400"
                >
                  <button
                    type="button"
                    class="inline-flex items-center gap-2 rounded-full px-2 py-1 transition hover:bg-slate-100 hover:text-slate-900"
                    @click="openRoot"
                  >
                    <Home class="h-4 w-4" />
                    <span>主页</span>
                  </button>
                  <template v-for="item in breadcrumbs" :key="item.id">
                    <ChevronRight class="h-4 w-4 text-slate-300" />
                    <button
                      type="button"
                      class="rounded-full px-2 py-1 transition hover:bg-slate-100 hover:text-slate-900"
                      @click="openFolder(item.id)"
                    >
                      {{ item.name }}
                    </button>
                  </template>
                </div>
              </div>
            </div>
          </div>

          <div>
            <SearchSection
              v-model="searchInput"
              embedded
              :loading="searchLoading"
              @search="runSearch"
              @clear="clearSearchState"
            />
          </div>

          <p
            v-if="searchError"
            class="mx-5 mt-3 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700 sm:mx-6"
          >
            {{ searchError }}
          </p>
          <div
            v-else-if="searchKeyword"
            class="mx-5 mt-3 rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-600 sm:mx-6"
          >
            当前搜索：<span class="font-medium text-slate-900">{{
              searchKeyword
            }}</span>
            <span class="ml-2">共 {{ searchRows.length }} 条结果</span>
          </div>

          <PublicDirectoryToolbar
            :all-visible-rows-selected="allVisibleRowsSelected"
            :back-button-label="backButtonLabel"
            :can-upload-to-current-folder="canUploadToCurrentFolder"
            :can-use-back-button="canUseBackButton"
            :has-rows="sortedRows.length > 0"
            :sort-direction="sortDirection"
            :sort-menu-open="sortMenuOpen"
            :sort-mode="sortMode"
            :view-menu-open="viewMenuOpen"
            :view-mode="viewMode"
            @go-up="goUpOneLevel"
            @open-upload="openUpload"
            @set-sort-direction="setSortDirection"
            @set-sort-menu-open="sortMenuOpen = $event"
            @set-sort-mode="setSortMode"
            @set-view-menu-open="viewMenuOpen = $event"
            @set-view-mode="setViewMode"
            @toggle-select-all="toggleSelectAllVisibleRows"
          />

          <p
            v-if="actionMessage"
            class="mx-4 mt-5 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700 sm:mx-6"
          >
            {{ actionMessage }}
          </p>
          <p
            v-if="actionError"
            class="mx-4 mt-5 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700 sm:mx-6"
          >
            {{ actionError }}
          </p>

          <div v-if="loading" class="px-4 py-8 text-sm text-slate-500 sm:px-6">
            加载中…
          </div>
          <div
            v-else-if="error"
            class="px-4 py-8 text-sm text-rose-600 sm:px-6"
          >
            {{ error }}
          </div>
          <div
            v-else-if="sortedRows.length === 0"
            class="px-4 py-8 text-sm text-slate-500 sm:px-6"
          >
            {{ searchKeyword ? "没有找到匹配结果。" : "当前目录为空。" }}
          </div>
          <PublicDirectoryCards
            v-else-if="viewMode === 'cards'"
            :file-icon-component="fileIconComponent"
            :is-row-selected="isRowSelected"
            :rows="sortedRows"
            @download="downloadResource"
            @feedback="
              openFeedbackModal({
                id: $event.id,
                type: $event.kind,
                name: $event.name,
              })
            "
            @open="$event.kind === 'folder' ? openFolder($event.id) : openFile($event.id)"
            @toggle-selection="toggleRowSelection"
          />
          <PublicDirectoryTable
            v-else
            :file-icon-component="fileIconComponent"
            :is-row-selected="isRowSelected"
            :rows="sortedRows"
            @open="$event.kind === 'folder' ? openFolder($event.id) : openFile($event.id)"
            @toggle-selection="toggleRowSelection"
          />

          <PublicFolderInfoPanel
            :can-manage-resource-descriptions="canManageResourceDescriptions"
            :current-folder-description-html="currentFolderDescriptionHTML"
            :current-folder-detail="currentFolderDetail"
            :current-folder-stats="currentFolderStats"
            :readme-preview-error="readmePreviewError"
            :readme-preview-html="readmePreviewHTML"
            :readme-preview-loading="readmePreviewLoading"
            :readme-preview-name="readmePreviewName"
            @delete-folder="openDeleteFolderDialog"
            @download-folder="downloadCurrentFolder"
            @edit-folder="openFolderDescriptionEditor"
            @feedback-folder="
              currentFolderDetail &&
                openFeedbackModal({
                  id: currentFolderDetail.id,
                  type: 'folder',
                  name: currentFolderDetail.name,
                })
            "
          />
        </div>
      </section>

      <aside class="order-2 min-w-0 space-y-4">
        <InfoPanelCard
          title="公告栏"
          :items="recentAnnouncements"
          clickable
          action-label="详情"
          empty-text="暂无公告"
          @select="openAnnouncementDetail"
          @action="openAnnouncementList"
        />
        <InfoPanelCard
          title="热门下载"
          :items="hotDownloads"
          clickable
          action-label="详情"
          empty-text="暂无下载数据"
          @select="openSidebarDetailItem"
          @action="openHotDownloadsModal"
        />
        <InfoPanelCard
          title="资料上新"
          :items="latestTitles"
          clickable
          action-label="详情"
          empty-text="暂无最新资料"
          @select="openSidebarDetailItem"
          @action="openLatestItemsModal"
        />
      </aside>
    </div>
  </main>

  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="translate-y-6 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="translate-y-4 opacity-0"
    >
      <div
        v-if="hasSelectedRows"
        class="pointer-events-none fixed inset-x-0 bottom-6 z-[130] flex justify-center px-4"
      >
        <div
          class="pointer-events-auto flex w-full max-w-3xl flex-col gap-3 rounded-3xl border border-slate-200 bg-white px-4 py-4 shadow-[0_0_0_1px_rgba(15,23,42,0.06),0_22px_60px_-18px_rgba(15,23,42,0.34)] sm:flex-row sm:items-center sm:justify-between sm:px-6"
        >
          <p class="text-sm text-slate-600">
            已选
            <span class="font-semibold text-slate-900">{{
              selectedRows.length
            }}</span>
            项
          </p>
          <div
            class="flex w-full flex-col gap-3 sm:w-auto sm:flex-row sm:items-center"
          >
            <button
              type="button"
              class="btn-secondary w-full sm:w-auto"
              @click="clearSelection"
            >
              取消选择
            </button>
            <button
              type="button"
              class="inline-flex h-11 w-full items-center justify-center rounded-xl border border-slate-200 bg-white px-5 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50 hover:text-slate-900 disabled:cursor-not-allowed disabled:opacity-60 sm:w-auto"
              :disabled="batchDownloadSubmitting"
              @click="downloadSelectedResources"
            >
              {{ batchDownloadSubmitting ? "打包中…" : "批量下载" }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>

  <PublicHomeSidebarDetailModal
    :modal="sidebarDetailModal"
    @close="closeSidebarDetailModal"
    @select-item="openSidebarDetailItem"
  />

  <PublicAnnouncementListModal
    :announcements="announcements"
    :open="announcementListOpen"
    @close="closeAnnouncementList"
    @open-detail="openAnnouncementDetail"
  />

  <PublicAnnouncementDetailModal
    :announcement-detail="announcementDetail"
    @back="returnToAnnouncementList"
    @close="closeAnnouncementDetail"
  />

  <PublicDeleteResourceDialog
    :delete-error="folderAdminState.deleteError"
    :delete-password="folderAdminState.deletePassword"
    :delete-submitting="folderAdminState.deleteSubmitting"
    :open="Boolean(folderAdminState.deleteTarget)"
    :resource-name="folderAdminState.deleteTarget?.name || ''"
    @close="closeDeleteResourceDialog"
    @confirm="confirmDeleteResource"
    @update:delete-password="folderAdminState.deletePassword = $event"
  />

  <PublicUploadDialog
    :breadcrumbs="breadcrumbs"
    :current-receipt-code="currentReceiptCode"
    :description="uploadState.form.description"
    :entries="uploadState.form.entries"
    :error="uploadState.error"
    :message="uploadState.message"
    :open="uploadState.modalOpen"
    :success-message="uploadState.message"
    :success-open="uploadState.successOpen"
    :upload-collecting="uploadState.collecting"
    :upload-drop-active="uploadState.dropActive"
    :upload-submitting="uploadState.submitting"
    @change-file="onUploadFileChange"
    @clear-entries="clearUploadEntries"
    @close="closeUploadModal"
    @close-success="closeUploadSuccessModal"
    @dragenter="onUploadDragEnter"
    @dragleave="onUploadDragLeave"
    @drop="onUploadDrop"
    @submit="submitUpload"
    @update:description="uploadState.form.description = $event"
    @update:upload-drop-active="uploadState.dropActive = $event"
  />

  <PublicHomeFeedbackDialog
    :current-receipt-code="currentReceiptCode"
    :feedback-description="feedbackState.description"
    :feedback-error="feedbackState.error"
    :feedback-message="feedbackState.message"
    :feedback-submit-disabled="feedbackSubmitDisabled"
    :feedback-submitting="feedbackState.submitting"
    :open="feedbackState.modalOpen"
    :success-open="feedbackState.successOpen"
    :target-name="feedbackState.target?.name || ''"
    @close="closeFeedbackModal"
    @close-success="closeFeedbackSuccessModal"
    @submit="submitFeedback"
    @update:feedback-description="feedbackState.description = $event"
  />

  <PublicFolderDescriptionEditor
    :can-manage-resource-descriptions="canManageResourceDescriptions"
    :description="folderAdminState.descriptionDraft"
    :error="folderAdminState.error"
    :folder-name="folderAdminState.nameDraft"
    :open="folderAdminState.editorOpen"
    :saving="folderAdminState.saving"
    :submit-disabled="!folderEditorDirty"
    @close="closeFolderDescriptionEditor"
    @save="saveFolderDescription"
    @update:description="folderAdminState.descriptionDraft = $event"
    @update:folder-name="folderAdminState.nameDraft = $event"
  />
</template>

<style scoped>
@keyframes warning-fade-in {
  0% {
    opacity: 0;
    transform: translateY(8px) scale(0.98);
  }

  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes warning-fade-out {
  0% {
    opacity: 1;
    transform: translateY(0);
  }

  100% {
    opacity: 0;
    transform: translateY(-6px);
  }
}
</style>
