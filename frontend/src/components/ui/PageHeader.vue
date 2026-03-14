<script setup lang="ts">
withDefaults(
  defineProps<{
    eyebrow?: string;
    title: string;
    description?: string;
  }>(),
  {
    eyebrow: "",
    description: "",
  },
);
</script>

<template>
  <header class="space-y-2">
    <p v-if="eyebrow" class="text-xs font-semibold uppercase tracking-[0.18em] text-blue-600">
      {{ eyebrow }}
    </p>
    <div class="space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight text-slate-900 sm:text-3xl">
        {{ title }}
      </h1>
      <p v-if="description" class="max-w-3xl text-sm leading-6 text-slate-500">
        {{ description }}
      </p>
    </div>
  </header>
</template>
