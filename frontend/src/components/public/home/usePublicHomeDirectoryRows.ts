import { computed, ref, watch, type Ref } from "vue";
import {
  FileArchive,
  FileAudio,
  FileCode2,
  FileImage,
  FileSpreadsheet,
  FileText,
  FileType2,
  FileVideo,
} from "lucide-vue-next";
import type {
  DirectoryRow,
  PublicFileItem,
  PublicFolderItem,
  PublicHomeSortDirection,
  PublicHomeSortMode,
  PublicHomeViewMode,
} from "./types";

interface UsePublicHomeDirectoryRowsOptions {
  currentFolderID: Ref<string>;
  files: Ref<PublicFileItem[]>;
  folders: Ref<PublicFolderItem[]>;
  formatDateTime: (value: string) => string;
  formatSize: (size: number) => string;
  searchKeyword: Ref<string>;
  searchRows: Ref<DirectoryRow[]>;
}

export function extractExtension(name: string) {
  const index = name.lastIndexOf(".");
  if (index <= 0 || index === name.length - 1) {
    return "";
  }
  return name.slice(index + 1).toLowerCase();
}

function compareRows(
  left: DirectoryRow,
  right: DirectoryRow,
  mode: PublicHomeSortMode,
  direction: PublicHomeSortDirection,
) {
  let result = 0;

  if (mode === "download") {
    if (left.downloadCount !== right.downloadCount) {
      result = left.downloadCount - right.downloadCount;
    } else {
      result = left.name.localeCompare(right.name, "zh-CN");
    }
  } else if (mode === "format") {
    const leftRank = formatSortRank(left);
    const rightRank = formatSortRank(right);
    if (leftRank !== rightRank) {
      result = leftRank - rightRank;
    } else {
      result = left.name.localeCompare(right.name, "zh-CN");
    }
  } else {
    result = left.name.localeCompare(right.name, "zh-CN");
  }

  return direction === "asc" ? result : -result;
}

function formatSortRank(row: DirectoryRow) {
  if (row.kind === "folder") {
    return 0;
  }

  const extension = row.extension.toLowerCase();
  if (extension === "pdf") {
    return 1;
  }
  if (["doc", "docx", "xls", "xlsx", "ppt", "pptx"].includes(extension)) {
    return 2;
  }
  return 3;
}

export function usePublicHomeDirectoryRows({
  currentFolderID,
  files,
  folders,
  formatDateTime,
  formatSize,
  searchKeyword,
  searchRows,
}: UsePublicHomeDirectoryRowsOptions) {
  const viewMode = ref<PublicHomeViewMode>("table");
  const sortMode = ref<PublicHomeSortMode>("name");
  const sortDirection = ref<PublicHomeSortDirection>("desc");
  const sortMenuOpen = ref(false);
  const viewMenuOpen = ref(false);
  const selectedResourceKeys = ref<string[]>([]);

  const rows = computed<DirectoryRow[]>(() => [
    ...folders.value.map((folder) => ({
      id: folder.id,
      kind: "folder" as const,
      name: folder.name,
      extension: "",
      description: "",
      downloadCount: folder.download_count ?? 0,
      fileCount: folder.file_count ?? 0,
      sizeText: formatSize(folder.total_size ?? 0),
      updatedAt: formatDateTime(folder.updated_at),
      downloadURL: `/api/public/folders/${encodeURIComponent(folder.id)}/download`,
    })),
    ...(currentFolderID.value
      ? files.value.map((file) => ({
          id: file.id,
          kind: "file" as const,
          name: file.name,
          extension: file.extension || extractExtension(file.name),
          description: (file.description ?? "").trim(),
          downloadCount: file.download_count ?? 0,
          fileCount: 0,
          sizeText: formatSize(file.size),
          updatedAt: formatDateTime(file.uploaded_at),
          downloadURL: `/api/public/files/${encodeURIComponent(file.id)}/download`,
        }))
      : []),
  ]);

  const displayedRows = computed<DirectoryRow[]>(() =>
    searchKeyword.value ? searchRows.value : rows.value,
  );

  const sortedRows = computed(() => {
    const sortedFolders = displayedRows.value
      .filter((row) => row.kind === "folder")
      .sort((left, right) =>
        compareRows(left, right, sortMode.value, sortDirection.value),
      );
    const sortedFiles = displayedRows.value
      .filter((row) => row.kind === "file")
      .sort((left, right) =>
        compareRows(left, right, sortMode.value, sortDirection.value),
      );

    return [...sortedFolders, ...sortedFiles];
  });

  const selectedRows = computed(() =>
    sortedRows.value.filter((row) =>
      selectedResourceKeys.value.includes(selectionKey(row)),
    ),
  );

  const hasSelectedRows = computed(() => selectedRows.value.length > 0);
  const allVisibleRowsSelected = computed(
    () =>
      sortedRows.value.length > 0
      && selectedRows.value.length === sortedRows.value.length,
  );

  watch(
    sortedRows,
    (currentRows) => {
      const allowedKeys = new Set(currentRows.map((row) => selectionKey(row)));
      selectedResourceKeys.value = selectedResourceKeys.value.filter((key) =>
        allowedKeys.has(key),
      );
    },
    { immediate: true },
  );

  function restoreDisplayPreferences() {
    const storedViewMode = window.localStorage.getItem("public-home-view-mode");
    if (storedViewMode === "cards" || storedViewMode === "table") {
      viewMode.value = storedViewMode;
    }

    const storedSortMode = window.localStorage.getItem("public-home-sort-mode");
    if (
      storedSortMode === "name"
      || storedSortMode === "download"
      || storedSortMode === "format"
    ) {
      sortMode.value = storedSortMode;
    }

    const storedSortDirection = window.localStorage.getItem(
      "public-home-sort-direction",
    );
    if (storedSortDirection === "asc" || storedSortDirection === "desc") {
      sortDirection.value = storedSortDirection;
    }
  }

  function selectionKey(row: DirectoryRow) {
    return `${row.kind}:${row.id}`;
  }

  function isRowSelected(row: DirectoryRow) {
    return selectedResourceKeys.value.includes(selectionKey(row));
  }

  function toggleRowSelection(row: DirectoryRow) {
    const key = selectionKey(row);
    if (selectedResourceKeys.value.includes(key)) {
      selectedResourceKeys.value = selectedResourceKeys.value.filter(
        (item) => item !== key,
      );
      return;
    }
    selectedResourceKeys.value = [...selectedResourceKeys.value, key];
  }

  function clearSelection() {
    selectedResourceKeys.value = [];
  }

  function toggleSelectAllVisibleRows() {
    if (allVisibleRowsSelected.value) {
      clearSelection();
      return;
    }
    selectedResourceKeys.value = sortedRows.value.map((row) => selectionKey(row));
  }

  function setViewMode(mode: PublicHomeViewMode) {
    viewMode.value = mode;
    viewMenuOpen.value = false;
    window.localStorage.setItem("public-home-view-mode", mode);
  }

  function setSortMode(mode: PublicHomeSortMode) {
    sortMode.value = mode;
    window.localStorage.setItem("public-home-sort-mode", mode);
  }

  function setSortDirection(direction: PublicHomeSortDirection) {
    sortDirection.value = direction;
    sortMenuOpen.value = false;
    window.localStorage.setItem("public-home-sort-direction", direction);
  }

  function sortModeLabel(mode: PublicHomeSortMode) {
    switch (mode) {
      case "download":
        return "下载量排序";
      case "format":
        return "格式排序";
      default:
        return "名称排序";
    }
  }

  function sortDirectionLabel(direction: PublicHomeSortDirection) {
    return direction === "asc" ? "升序" : "降序";
  }

  function viewModeLabel(mode: PublicHomeViewMode) {
    return mode === "cards" ? "卡片" : "表格";
  }

  function fileIconComponent(extension: string) {
    const ext = extension.toLowerCase();
    if (["png", "jpg", "jpeg", "gif", "webp", "svg", "bmp", "ico"].includes(ext)) {
      return FileImage;
    }
    if (["mp4", "mov", "avi", "mkv", "webm"].includes(ext)) {
      return FileVideo;
    }
    if (["mp3", "wav", "flac", "aac", "m4a", "ogg"].includes(ext)) {
      return FileAudio;
    }
    if (["zip", "rar", "7z", "tar", "gz", "bz2", "xz"].includes(ext)) {
      return FileArchive;
    }
    if (["xls", "xlsx", "csv", "numbers"].includes(ext)) {
      return FileSpreadsheet;
    }
    if (
      [
        "js",
        "ts",
        "jsx",
        "tsx",
        "json",
        "html",
        "css",
        "go",
        "py",
        "java",
        "c",
        "cpp",
        "h",
        "hpp",
        "rs",
        "sh",
        "yaml",
        "yml",
        "toml",
        "xml",
      ].includes(ext)
    ) {
      return FileCode2;
    }
    if (["pdf", "doc", "docx", "ppt", "pptx", "txt", "md", "rtf"].includes(ext)) {
      return FileText;
    }
    return FileType2;
  }

  return {
    allVisibleRowsSelected,
    clearSelection,
    fileIconComponent,
    hasSelectedRows,
    isRowSelected,
    restoreDisplayPreferences,
    selectedResourceKeys,
    selectedRows,
    setSortDirection,
    setSortMode,
    setViewMode,
    sortDirection,
    sortDirectionLabel,
    sortMenuOpen,
    sortMode,
    sortModeLabel,
    sortedRows,
    toggleRowSelection,
    toggleSelectAllVisibleRows,
    viewMenuOpen,
    viewMode,
    viewModeLabel,
  };
}
