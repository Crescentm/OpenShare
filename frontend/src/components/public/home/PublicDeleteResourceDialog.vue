<script setup lang="ts">
defineProps<{
  deleteError: string;
  deletePassword: string;
  deleteSubmitting: boolean;
  open: boolean;
  resourceName: string;
}>();

defineEmits<{
  close: [];
  confirm: [];
  "update:deletePassword": [value: string];
}>();
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-shell">
      <div
        v-if="open"
        class="fixed inset-0 z-[120] flex items-center justify-center bg-slate-950/30 px-4"
      >
        <div class="modal-card w-full max-w-md rounded-2xl bg-white p-6 shadow-xl">
          <div>
            <h3 class="text-lg font-semibold text-slate-900">确认删除文件夹</h3>
            <p class="mt-2 text-sm leading-6 text-slate-500">
              删除后会清除该文件夹及其子目录、文件，无法恢复。确认删除
              <span class="font-medium text-slate-900">{{ resourceName }}</span>
              吗？
            </p>
          </div>
          <div class="mt-6 space-y-4">
            <input
              :value="deletePassword"
              type="password"
              class="field"
              placeholder="输入当前管理员密码确认删除"
              @input="
                $emit(
                  'update:deletePassword',
                  ($event.target as HTMLInputElement).value,
                )
              "
            />
            <p
              v-if="deleteError"
              class="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700"
            >
              {{ deleteError }}
            </p>
            <div class="flex justify-end gap-3">
              <button type="button" class="btn-secondary" @click="$emit('close')">
                取消
              </button>
              <button
                type="button"
                class="inline-flex h-11 items-center rounded-xl bg-rose-600 px-5 text-sm font-medium text-white transition hover:bg-rose-700"
                :disabled="deleteSubmitting"
                @click="$emit('confirm')"
              >
                {{ deleteSubmitting ? "删除中…" : "确认删除" }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
