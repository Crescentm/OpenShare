<script setup lang="ts">
withDefaults(
  defineProps<{
    label: string;
    selected?: boolean;
  }>(),
  {
    selected: false,
  },
);
</script>

<template>
  <button
    type="button"
    class="inline-flex items-center rounded-md border px-2.5 py-1 text-sm font-medium transition"
    :class="
      selected
        ? 'border-slate-300 bg-slate-100 text-slate-900 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100'
        : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300 hover:bg-slate-100 hover:text-slate-900 dark:border-slate-800 dark:bg-slate-950 dark:text-slate-400 dark:hover:border-slate-700 dark:hover:bg-slate-900 dark:hover:text-slate-100'
    "
  >
    {{ label }}
  </button>
</template>
