<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string;
    description?: string;
  }>(),
  {
    description: "",
  },
);
</script>

<template>
  <div class="rounded-2xl border border-dashed border-slate-300 bg-[#fafafa] px-6 py-12 text-center">
    <p class="text-sm font-medium text-slate-900">{{ title }}</p>
    <p v-if="description" class="mt-2 text-sm text-slate-500">{{ description }}</p>
  </div>
</template>
