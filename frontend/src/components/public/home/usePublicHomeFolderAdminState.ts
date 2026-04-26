import { computed, reactive, type Ref } from "vue";

import type { FolderDetailResponse } from "./types";

type DeleteFolderTarget = {
  id: string;
  kind: "folder";
  name: string;
};

export function usePublicHomeFolderAdminState(
  currentFolderDetail: Ref<FolderDetailResponse | null>,
) {
  const folderAdminState = reactive({
    editorOpen: false,
    nameDraft: "",
    descriptionDraft: "",
    saving: false,
    error: "",
    deleteTarget: null as DeleteFolderTarget | null,
    deletePassword: "",
    deleteSubmitting: false,
    deleteError: "",
  });

  const folderEditorDirty = computed(() => {
    if (!currentFolderDetail.value) {
      return false;
    }

    return (
      folderAdminState.nameDraft.trim() !== currentFolderDetail.value.name ||
      folderAdminState.descriptionDraft.trim() !==
        (currentFolderDetail.value.description ?? "")
    );
  });

  function syncFolderDrafts() {
    folderAdminState.nameDraft = currentFolderDetail.value?.name ?? "";
    folderAdminState.descriptionDraft =
      currentFolderDetail.value?.description ?? "";
  }

  function resetDeleteDialog() {
    folderAdminState.deleteTarget = null;
    folderAdminState.deletePassword = "";
    folderAdminState.deleteError = "";
    folderAdminState.deleteSubmitting = false;
  }

  function openFolderDescriptionEditor() {
    syncFolderDrafts();
    folderAdminState.error = "";
    folderAdminState.editorOpen = true;
  }

  function closeFolderDescriptionEditor() {
    folderAdminState.editorOpen = false;
    folderAdminState.saving = false;
    folderAdminState.error = "";
    syncFolderDrafts();
  }

  function openDeleteFolderDialog() {
    if (!currentFolderDetail.value) {
      return;
    }

    folderAdminState.deleteTarget = {
      id: currentFolderDetail.value.id,
      kind: "folder",
      name: currentFolderDetail.value.name,
    };
    folderAdminState.deletePassword = "";
    folderAdminState.deleteError = "";
  }

  return {
    closeFolderDescriptionEditor,
    folderAdminState,
    folderEditorDirty,
    openDeleteFolderDialog,
    openFolderDescriptionEditor,
    resetDeleteDialog,
    syncFolderDrafts,
  };
}
