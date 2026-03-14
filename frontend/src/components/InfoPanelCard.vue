<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    title: string;
    items?: string[];
    emptyText?: string;
  }>(),
  {
    items: () => [],
    emptyText: "暂无内容",
  },
);

const hasItems = computed(() => props.items.length > 0);
</script>

<template>
  <section class="panel p-4">
    <header class="flex items-center justify-between gap-3">
      <h2 class="text-sm font-medium tracking-tight text-slate-900 dark:text-slate-100">
        {{ title }}
      </h2>
      <span class="text-xs text-slate-400 dark:text-slate-500">{{ hasItems ? props.items.length : "--" }}</span>
    </header>

    <div class="mt-4">
      <div v-if="hasItems" class="space-y-2">
        <div
          v-for="item in props.items"
          :key="item"
          class="rounded-lg px-2 py-2 text-sm leading-6 text-slate-600 dark:text-slate-300"
        >
          {{ item }}
        </div>
      </div>

      <div v-else class="rounded-lg bg-[#fafafa] px-3 py-3 text-sm text-slate-500 dark:bg-slate-800/70 dark:text-slate-400">
        {{ props.emptyText }}
      </div>
    </div>
  </section>
</template>
