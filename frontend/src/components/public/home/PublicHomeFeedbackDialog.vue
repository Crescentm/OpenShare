<script setup lang="ts">
defineProps<{
  currentReceiptCode: string;
  feedbackDescription: string;
  feedbackError: string;
  feedbackMessage: string;
  feedbackSubmitDisabled: boolean;
  feedbackSubmitting: boolean;
  open: boolean;
  successOpen: boolean;
  targetName: string;
}>();

defineEmits<{
  close: [];
  closeSuccess: [];
  submit: [];
  "update:feedbackDescription": [value: string];
}>();
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-shell">
      <div
        v-if="successOpen"
        class="fixed inset-0 z-[120] bg-slate-950/40 backdrop-blur-sm"
      >
        <div class="flex min-h-screen items-center justify-center px-4 py-6">
          <div class="modal-card w-full max-w-md rounded-2xl bg-white p-6 shadow-xl">
            <div class="space-y-3">
              <h3 class="text-lg font-semibold text-slate-900">提交成功</h3>
              <p class="text-sm leading-6 text-slate-600">
                {{ feedbackMessage }}
              </p>
            </div>
            <div class="mt-6 flex justify-end">
              <button type="button" class="btn-primary" @click="$emit('closeSuccess')">
                知道了
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>

  <Teleport to="body">
    <Transition name="modal-shell">
      <div
        v-if="open"
        class="fixed inset-0 z-[120] bg-slate-950/40 backdrop-blur-sm"
      >
        <div class="flex min-h-screen items-center justify-center px-4 py-6">
          <div class="modal-card panel w-full max-w-2xl overflow-hidden p-6">
            <div
              class="flex items-start justify-between gap-4 border-b border-slate-200 pb-5"
            >
              <div class="space-y-1">
                <h3 class="text-lg font-semibold text-slate-900">反馈中心</h3>
                <p class="text-sm text-slate-500">
                  填写问题说明后提交，我们会尽快处理。
                </p>
              </div>
              <button type="button" class="btn-secondary" @click="$emit('close')">
                关闭
              </button>
            </div>

            <div class="mt-6 space-y-5">
              <div
                v-if="targetName"
                class="rounded-2xl border border-slate-200 bg-[#fafafafa] px-4 py-3"
              >
                <p
                  class="text-xs font-semibold uppercase tracking-[0.12em] text-slate-400"
                >
                  当前对象
                </p>
                <p class="mt-1 text-sm leading-6 text-slate-700">
                  {{ targetName }}
                </p>
              </div>

              <label class="space-y-2">
                <span class="text-sm font-medium text-slate-700">回执码</span>
                <div
                  class="rounded-2xl border border-slate-200 bg-[#fafafafa] px-4 py-3"
                >
                  <p class="text-sm font-semibold tracking-[0.12em] text-slate-900">
                    {{ currentReceiptCode || "当前会话回执码暂未同步" }}
                  </p>
                </div>
              </label>

              <label class="space-y-2">
                <span class="text-sm font-medium text-slate-700">问题说明</span>
                <textarea
                  :value="feedbackDescription"
                  rows="5"
                  class="field-area"
                  placeholder="信息不当/侵权/内容错误……描述您遇到的问题，我们会尽快改进！"
                  @input="
                    $emit(
                      'update:feedbackDescription',
                      ($event.target as HTMLTextAreaElement).value,
                    )
                  "
                />
              </label>

              <p
                v-if="feedbackMessage"
                class="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700"
              >
                {{ feedbackMessage }}
              </p>
              <p
                v-if="feedbackError"
                class="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700"
              >
                {{ feedbackError }}
              </p>

              <div class="flex justify-end gap-3 pt-1">
                <button type="button" class="btn-secondary" @click="$emit('close')">
                  取消
                </button>
                <button
                  type="button"
                  class="btn-primary"
                  :disabled="feedbackSubmitDisabled"
                  @click="$emit('submit')"
                >
                  {{ feedbackSubmitting ? "提交中…" : "提交反馈" }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
