<script setup lang="ts">
import { ref } from "vue";
import { Upload } from "lucide-vue-next";

interface UploadEntry {
  relativePath: string;
}

defineProps<{
  breadcrumbs: Array<{ id: string; name: string }>;
  currentReceiptCode: string;
  description: string;
  entries: UploadEntry[];
  error: string;
  message: string;
  open: boolean;
  successMessage: string;
  successOpen: boolean;
  uploadCollecting: boolean;
  uploadDropActive: boolean;
  uploadSubmitting: boolean;
}>();

defineEmits<{
  changeFile: [event: Event];
  clearEntries: [];
  close: [];
  closeSuccess: [];
  dragenter: [];
  dragleave: [event: DragEvent];
  drop: [event: DragEvent];
  submit: [];
  "update:description": [value: string];
  "update:uploadDropActive": [value: boolean];
}>();

const uploadFileInput = ref<HTMLInputElement | null>(null);

function triggerUploadFileSelect() {
  uploadFileInput.value?.click();
}
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
                {{ successMessage }}
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
        class="fixed inset-0 z-[120] overflow-y-auto bg-slate-950/40 backdrop-blur-sm"
      >
        <div class="flex min-h-screen items-start justify-center px-4 py-6">
          <div class="modal-card panel w-full max-w-2xl overflow-hidden">
            <div class="max-h-[calc(100vh-3rem)] overflow-y-auto p-6">
              <div
                class="flex items-start justify-between gap-4 border-b border-slate-200 pb-4"
              >
                <div>
                  <h3 class="text-lg font-semibold text-slate-900">上传资料</h3>
                  <p class="mt-1 text-sm text-slate-500">
                    当前目录下直接上传资料，提交后会进入审核池。
                  </p>
                </div>
                <button type="button" class="btn-secondary" @click="$emit('close')">
                  关闭
                </button>
              </div>

              <form class="mt-5 space-y-4" @submit.prevent="$emit('submit')">
                <div class="panel-muted px-4 py-3 text-sm text-slate-600">
                  <p class="text-xs text-slate-400">目标目录</p>
                  <p class="mt-1 font-medium text-slate-900">
                    {{
                      breadcrumbs.length
                        ? breadcrumbs.map((item) => item.name).join(" / ")
                        : "主页根目录"
                    }}
                  </p>
                </div>

                <label class="space-y-2">
                  <span class="text-sm font-medium text-slate-700">回执码</span>
                  <div class="rounded-xl bg-slate-50 px-4 py-3">
                    <p class="text-sm font-semibold tracking-[0.12em] text-slate-900">
                      {{ currentReceiptCode || "当前会话回执码暂未同步" }}
                    </p>
                  </div>
                </label>

                <label class="space-y-2">
                  <span class="text-sm font-medium text-slate-700">资料简介</span>
                  <textarea
                    :value="description"
                    rows="4"
                    class="field-area"
                    placeholder="可选，简要介绍资料内容和适用场景，支持简单 Markdown 语法"
                    @input="
                      $emit(
                        'update:description',
                        ($event.target as HTMLTextAreaElement).value,
                      )
                    "
                  />
                </label>

                <div class="space-y-2">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium text-slate-700">上传内容</span>
                  </div>

                  <input
                    ref="uploadFileInput"
                    type="file"
                    class="hidden"
                    @change="$emit('changeFile', $event)"
                  />

                  <div
                    class="rounded-[28px] border-2 border-dashed px-6 py-10 text-center transition"
                    :class="
                      uploadDropActive
                        ? 'border-blue-400 bg-blue-50/60'
                        : 'border-slate-200 bg-slate-50/60'
                    "
                    @dragenter.prevent="$emit('dragenter')"
                    @dragover.prevent="$emit('update:uploadDropActive', true)"
                    @dragleave="$emit('dragleave', $event)"
                    @drop="$emit('drop', $event)"
                  >
                    <div
                      class="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-white text-slate-300 shadow-sm"
                    >
                      <Upload class="h-8 w-8" />
                    </div>
                    <p class="mt-5 text-lg text-slate-600">
                      拖拽文件或整个文件夹到这里，或
                      <button
                        type="button"
                        class="font-semibold text-blue-600 transition hover:text-blue-700"
                        @click="triggerUploadFileSelect"
                      >
                        点击选择
                      </button>
                    </p>
                    <p class="mt-2 text-sm text-slate-400">
                      拖拽支持多文件和文件夹。
                    </p>
                    <p v-if="uploadCollecting" class="mt-4 text-sm text-slate-500">
                      正在解析拖拽内容…
                    </p>
                  </div>

                  <div class="panel-muted px-4 py-3 text-sm text-slate-600">
                    <div class="flex flex-wrap items-center justify-between gap-3">
                      <p>
                        已选择
                        <span class="font-semibold text-slate-900">
                          {{ entries.length }}
                        </span>
                        个文件
                      </p>
                      <button
                        v-if="entries.length > 0"
                        type="button"
                        class="text-sm text-slate-500 transition hover:text-slate-900"
                        @click="$emit('clearEntries')"
                      >
                        清空列表
                      </button>
                    </div>
                    <div
                      v-if="entries.length > 0"
                      class="mt-3 max-h-48 space-y-2 overflow-auto pr-1"
                    >
                      <div
                        v-for="entry in entries"
                        :key="entry.relativePath"
                        class="rounded-xl bg-white px-3 py-2 text-sm text-slate-700"
                      >
                        {{ entry.relativePath }}
                      </div>
                    </div>
                    <p v-else class="mt-2 text-sm text-slate-400">
                      当前还没有选择任何文件。
                    </p>
                  </div>
                </div>

                <p
                  v-if="message"
                  class="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700"
                >
                  {{ message }}
                </p>
                <p
                  v-if="error"
                  class="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700"
                >
                  {{ error }}
                </p>

                <div class="flex justify-end gap-3">
                  <button type="button" class="btn-secondary" @click="$emit('close')">
                    取消
                  </button>
                  <button
                    type="submit"
                    class="btn-primary"
                    :disabled="uploadSubmitting || uploadCollecting || entries.length === 0"
                  >
                    {{ uploadSubmitting ? "提交中…" : "提交上传" }}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
