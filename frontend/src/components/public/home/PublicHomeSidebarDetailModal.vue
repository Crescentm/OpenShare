<script setup lang="ts">
interface SidebarDetailItem {
  id: string;
  label: string;
  meta?: string;
}

interface SidebarDetailModalState {
  eyebrow: string;
  title: string;
  description: string;
  items: SidebarDetailItem[];
}

defineProps<{
  modal: SidebarDetailModalState | null;
}>();

defineEmits<{
  close: [];
  selectItem: [item: SidebarDetailItem];
}>();
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-shell">
      <div
        v-if="modal"
        class="fixed inset-0 z-[120] flex items-center justify-center bg-slate-950/30 px-4"
      >
        <div class="modal-card panel w-full max-w-3xl p-6">
          <div
            class="flex items-start justify-between gap-4 border-b border-slate-200 pb-4"
          >
            <div>
              <p
                class="text-xs font-semibold uppercase tracking-[0.18em] text-blue-600"
              >
                {{ modal.eyebrow }}
              </p>
              <h3 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">
                {{ modal.title }}
              </h3>
              <p class="mt-2 text-sm text-slate-500">
                {{ modal.description }}
              </p>
            </div>
            <button type="button" class="btn-secondary" @click="$emit('close')">
              关闭
            </button>
          </div>
          <div class="mt-5 max-h-[70vh] overflow-y-auto pr-1">
            <div
              v-if="modal.items.length === 0"
              class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500"
            >
              暂无数据
            </div>
            <div v-else class="space-y-3">
              <button
                v-for="(item, index) in modal.items"
                :key="item.id"
                type="button"
                class="flex w-full items-center gap-4 rounded-2xl border border-slate-200 px-4 py-3 text-left transition hover:border-slate-300 hover:bg-slate-50"
                @click="$emit('selectItem', item)"
              >
                <span
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-slate-100 text-sm font-semibold text-slate-600"
                >
                  {{ index + 1 }}
                </span>
                <div class="min-w-0 flex-1">
                  <p class="truncate text-sm font-medium text-slate-900">
                    {{ item.label }}
                  </p>
                </div>
                <span v-if="item.meta" class="shrink-0 text-sm text-slate-500">
                  {{ item.meta }}
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
