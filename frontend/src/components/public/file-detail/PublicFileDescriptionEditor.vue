<script setup lang="ts">
defineProps<{
  canManageResourceDescriptions: boolean;
  description: string;
  fileName: string;
  open: boolean;
  saveError: string;
  saving: boolean;
  submitDisabled: boolean;
}>();

defineEmits<{
  close: [];
  save: [];
  "update:description": [value: string];
  "update:fileName": [value: string];
}>();
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-shell">
      <div
        v-if="open"
        class="fixed inset-0 z-[120] bg-slate-950/40 backdrop-blur-sm"
      >
        <div class="flex min-h-screen items-center justify-center px-4 py-6">
          <div class="modal-card panel w-full max-w-3xl overflow-hidden p-6">
            <div class="border-b border-slate-200 pb-4">
              <div>
                <h3 class="text-lg font-semibold text-slate-900">编辑文件信息</h3>
              </div>
            </div>

            <div class="mt-5 space-y-4">
              <label class="space-y-2">
                <span class="text-sm font-medium text-slate-700">文件名</span>
                <input
                  :value="fileName"
                  class="field"
                  :disabled="!canManageResourceDescriptions"
                  placeholder="输入完整文件名，例如 example.xlsx"
                  @input="
                    $emit(
                      'update:fileName',
                      ($event.target as HTMLInputElement).value,
                    )
                  "
                />
              </label>

              <textarea
                :value="description"
                rows="10"
                class="field-area"
                placeholder="输入文件简介，简介支持简单 Markdown。"
                @input="
                  $emit(
                    'update:description',
                    ($event.target as HTMLTextAreaElement).value,
                  )
                "
              />

              <p
                v-if="saveError"
                class="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700"
              >
                {{ saveError }}
              </p>

              <div class="flex justify-end gap-3">
                <button type="button" class="btn-secondary" @click="$emit('close')">
                  取消
                </button>
                <button
                  type="button"
                  class="btn-primary"
                  :disabled="saving || submitDisabled"
                  @click="$emit('save')"
                >
                  {{ saving ? "保存中…" : "保存更改" }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
