<script setup lang="ts">
withDefaults(
  defineProps<{
    padded?: boolean;
    subtle?: boolean;
  }>(),
  {
    padded: true,
    subtle: false,
  },
);
</script>

<template>
  <section
    class="rounded-2xl border shadow-sm"
    :class="[
      subtle ? 'border-slate-100 bg-slate-50/70 shadow-none' : 'border-slate-200 bg-white shadow-slate-950/[0.03]',
      padded ? 'p-6' : '',
    ]"
  >
    <slot />
  </section>
</template>
