<script setup lang="ts">
interface Props {
  fileName: string;
  inlineContentUrl: string;
  previewHeight: string;
  previewError: string;
  supportsAudioPreview: boolean;
  supportsImagePreview: boolean;
  supportsPdfPreview: boolean;
  supportsVideoPreview: boolean;
}

defineProps<Props>();

defineEmits<{
  nativeError: [];
}>();
</script>

<template>
  <div class="space-y-4">
    <div
      class="overflow-hidden border border-slate-200 bg-slate-50"
    >
      <img
        v-if="supportsImagePreview"
        :src="inlineContentUrl"
        :alt="fileName"
        class="max-h-[70vh] w-full object-contain"
        loading="lazy"
        @error="$emit('nativeError')"
      />
      <video
        v-else-if="supportsVideoPreview"
        :src="inlineContentUrl"
        controls
        preload="metadata"
        class="max-h-[70vh] w-full bg-black"
        @error="$emit('nativeError')"
      ></video>
      <div v-else-if="supportsAudioPreview" class="p-5">
        <audio
          :src="inlineContentUrl"
          controls
          preload="metadata"
          class="w-full"
          @error="$emit('nativeError')"
        ></audio>
      </div>
      <iframe
        v-else-if="supportsPdfPreview"
        :src="inlineContentUrl"
        :title="`${fileName} PDF preview`"
        class="w-full bg-white"
        :style="{ height: previewHeight }"
        @error="$emit('nativeError')"
      />
    </div>

    <div
      v-if="previewError"
      class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700"
    >
      {{ previewError }}
    </div>
  </div>
</template>
