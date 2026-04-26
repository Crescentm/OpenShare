import { computed, reactive } from "vue";

type FeedbackTarget = {
  id: string;
  type: "file" | "folder";
  name: string;
};

export function usePublicHomeFeedbackState() {
  const feedbackState = reactive({
    modalOpen: false,
    successOpen: false,
    target: null as FeedbackTarget | null,
    description: "",
    submitting: false,
    message: "",
    error: "",
  });

  const feedbackSubmitDisabled = computed(
    () => feedbackState.submitting || !feedbackState.description.trim(),
  );

  function openFeedbackModal(target: FeedbackTarget) {
    feedbackState.modalOpen = true;
    feedbackState.target = target;
    feedbackState.description = "";
    feedbackState.message = "";
    feedbackState.error = "";
  }

  function closeFeedbackModal() {
    feedbackState.modalOpen = false;
    feedbackState.target = null;
  }

  function closeFeedbackSuccessModal() {
    feedbackState.successOpen = false;
  }

  return {
    closeFeedbackModal,
    closeFeedbackSuccessModal,
    feedbackState,
    feedbackSubmitDisabled,
    openFeedbackModal,
  };
}
