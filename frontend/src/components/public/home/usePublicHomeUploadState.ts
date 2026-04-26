import { reactive } from "vue";

import type { UploadSelectionEntry } from "../../../lib/uploads/fileDrop";

export function usePublicHomeUploadState() {
  const uploadState = reactive({
    modalOpen: false,
    successOpen: false,
    submitting: false,
    message: "",
    error: "",
    dropActive: false,
    collecting: false,
    form: {
      description: "",
      entries: [] as UploadSelectionEntry[],
    },
  });

  function resetUploadForm() {
    uploadState.form.description = "";
    uploadState.form.entries = [];
  }

  function clearUploadEntries() {
    uploadState.form.entries = [];
  }

  function openUploadModal() {
    uploadState.modalOpen = true;
    uploadState.error = "";
    uploadState.message = "";
    resetUploadForm();
  }

  function closeUploadModal() {
    uploadState.modalOpen = false;
  }

  function closeUploadSuccessModal() {
    uploadState.successOpen = false;
  }

  return {
    clearUploadEntries,
    closeUploadModal,
    closeUploadSuccessModal,
    openUploadModal,
    resetUploadForm,
    uploadState,
  };
}
