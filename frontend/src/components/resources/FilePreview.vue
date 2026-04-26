<script setup lang="ts">
import { computed, ref, watch } from "vue";
import NativeFilePreview from "./previews/NativeFilePreview.vue";
import OfficeFilePreview from "./previews/OfficeFilePreview.vue";
import TextFilePreview from "./previews/TextFilePreview.vue";

interface Props {
  fileId: string;
  fileName: string;
  extension: string;
  mimeType: string;
  size: number;
  previewEnabled?: boolean;
}

type TextPreviewCachePayload = {
  content: string;
  fileName: string;
  mimeType: string;
  size: number;
  timestamp: number;
};

const MAX_TEXT_PREVIEW_SIZE = 512 * 1024;
const MAX_OFFICE_PREVIEW_SIZE = 10 * 1024 * 1024;
const TEXT_CACHE_TTL = 24 * 60 * 60 * 1000;
const officePreviewCache = new Map<string, ArrayBuffer>();

const props = withDefaults(defineProps<Props>(), {
  previewEnabled: true,
});
const emit = defineEmits<{
  previewReady: [];
}>();

const previewLoading = ref(false);
const previewError = ref("");
const previewProgress = ref(0);
const textContent = ref("");
const officeFileContent = ref<ArrayBuffer | null>(null);
let previewRequestToken = 0;

const normalizedExtension = computed(() =>
  (props.extension?.trim().toLowerCase() ?? "").replace(/^\.+/, ""),
);
const normalizedMimeType = computed(
  () => props.mimeType?.trim().toLowerCase() ?? "",
);
const normalizedFileName = computed(
  () => props.fileName?.trim().toLowerCase() ?? "",
);

const textCacheKey = computed(() => `file_preview_text_${props.fileId}`);
const inlineContentURL = computed(() => buildFilePreviewURL("inline"));

const isMarkdownFile = computed(
  () =>
    normalizedExtension.value === "md"
    || normalizedExtension.value === "markdown"
    || normalizedMimeType.value === "text/markdown"
    || normalizedMimeType.value === "text/x-markdown",
);

const supportsImagePreview = computed(
  () =>
    normalizedMimeType.value.startsWith("image/")
    || [
      "png",
      "jpg",
      "jpeg",
      "gif",
      "webp",
      "svg",
      "bmp",
      "ico",
      "avif",
    ].includes(normalizedExtension.value),
);

const supportsVideoPreview = computed(
  () =>
    normalizedMimeType.value.startsWith("video/")
    || ["mp4", "webm", "mov", "m4v", "ogg"].includes(normalizedExtension.value),
);

const supportsAudioPreview = computed(
  () =>
    normalizedMimeType.value.startsWith("audio/")
    || ["mp3", "wav", "flac", "aac", "m4a", "ogg"].includes(normalizedExtension.value),
);

const supportsPdfPreview = computed(
  () =>
    normalizedExtension.value === "pdf"
    || normalizedMimeType.value === "application/pdf",
);

const supportsOfficePreview = computed(
  () =>
    props.previewEnabled
    && (
      ["docx", "xlsx", "xls", "pptx"].includes(normalizedExtension.value)
      || [
        "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "application/vnd.ms-excel",
        "application/vnd.openxmlformats-officedocument.presentationml.presentation",
      ].includes(normalizedMimeType.value)
    ),
);

const supportsTextPreview = computed(() => {
  if (!props.previewEnabled) {
    return false;
  }

  if (isMarkdownFile.value) {
    return true;
  }

  if (normalizedFileName.value.startsWith("readme")) {
    return true;
  }

  const textExtensions = [
    "txt",
    "log",
    "json",
    "csv",
    "yaml",
    "yml",
    "toml",
    "xml",
    "html",
    "css",
    "js",
    "ts",
    "jsx",
    "tsx",
    "sh",
    "py",
    "go",
    "java",
    "c",
    "cpp",
    "h",
    "hpp",
    "rs",
    "sql",
  ];

  const textMimeTypes = [
    "application/json",
    "application/xml",
    "application/javascript",
    "application/x-javascript",
    "application/yaml",
    "application/x-yaml",
    "application/toml",
  ];

  return (
    textExtensions.includes(normalizedExtension.value)
    || normalizedMimeType.value.startsWith("text/")
    || textMimeTypes.includes(normalizedMimeType.value)
  );
});

const canPreviewText = computed(() => props.size <= MAX_TEXT_PREVIEW_SIZE);
const canPreviewOffice = computed(() => props.size <= MAX_OFFICE_PREVIEW_SIZE);

const previewHeight = computed(() => {
  if (typeof window === "undefined") {
    return "560px";
  }

  const targetHeight = window.innerHeight * 0.7;
  return `${Math.max(420, Math.min(760, targetHeight))}px`;
});

const unsupportedReason = computed(() => {
  if (supportsTextPreview.value && !canPreviewText.value) {
    return `文本预览仅支持 ${Math.round(MAX_TEXT_PREVIEW_SIZE / 1024)} KB 以内文件`;
  }

  if (supportsOfficePreview.value && !canPreviewOffice.value) {
    return `Office 预览仅支持 ${(MAX_OFFICE_PREVIEW_SIZE / (1024 * 1024)).toFixed(0)} MB 以内文件`;
  }

  return "此文件类型暂不支持预览";
});

function buildFilePreviewURL(view: "inline" | "text") {
  const query = new URLSearchParams({ view });
  return `/api/public/files/${encodeURIComponent(props.fileId)}/preview?${query.toString()}`;
}

function resetFetchedPreviewState() {
  previewLoading.value = false;
  previewError.value = "";
  previewProgress.value = 0;
  textContent.value = "";
  officeFileContent.value = null;
}

function emitPreviewReady(requestToken: number) {
  if (requestToken !== previewRequestToken) {
    return;
  }

  window.requestAnimationFrame(() => {
    if (requestToken === previewRequestToken) {
      emit("previewReady");
    }
  });
}

function getHttpErrorMessage(status: number): string {
  switch (status) {
    case 403:
      return "没有权限访问这个文件";
    case 404:
      return "文件不存在或已被删除";
    case 410:
      return "文件已被永久删除";
    case 413:
      return "文件过大，无法在线预览";
    case 429:
      return "请求过于频繁，请稍后再试";
    case 500:
      return "服务器内部错误";
    case 502:
      return "网关错误";
    case 503:
      return "服务暂时不可用";
    case 504:
      return "网关超时";
    default:
      return `预览加载失败（HTTP ${status}）`;
  }
}

function readTextCache() {
  try {
    const raw = window.localStorage.getItem(textCacheKey.value);
    if (!raw) {
      return null;
    }

    const parsed = JSON.parse(raw) as TextPreviewCachePayload;
    if (Date.now() - parsed.timestamp > TEXT_CACHE_TTL) {
      window.localStorage.removeItem(textCacheKey.value);
      return null;
    }

    if (
      parsed.fileName !== props.fileName
      || parsed.mimeType !== props.mimeType
      || parsed.size !== props.size
    ) {
      window.localStorage.removeItem(textCacheKey.value);
      return null;
    }

    return parsed.content;
  } catch {
    return null;
  }
}

function writeTextCache(content: string) {
  try {
    const payload: TextPreviewCachePayload = {
      content,
      fileName: props.fileName,
      mimeType: props.mimeType,
      size: props.size,
      timestamp: Date.now(),
    };
    window.localStorage.setItem(textCacheKey.value, JSON.stringify(payload));
  } catch {
    // 忽略缓存失败，避免影响预览主流程。
  }
}

function setPreviewError(message: string, requestToken: number) {
  if (requestToken !== previewRequestToken) {
    return;
  }

  previewError.value = message;
}

async function loadTextContent(requestToken: number) {
  if (!supportsTextPreview.value || !canPreviewText.value) {
    return;
  }

  const cached = readTextCache();
  if (cached !== null) {
    if (requestToken === previewRequestToken) {
      textContent.value = cached;
      emitPreviewReady(requestToken);
    }
    return;
  }

  previewLoading.value = true;
  previewError.value = "";
  previewProgress.value = 0;

  try {
    const response = await fetch(buildFilePreviewURL("text"));
    if (!response.ok) {
      throw new Error(getHttpErrorMessage(response.status));
    }

    let content = "";
    const reader = response.body?.getReader();
    const decoder = new TextDecoder();

    if (reader) {
      while (true) {
        const { done, value } = await reader.read();
        if (done) {
          break;
        }

        content += decoder.decode(value, { stream: true });
        if (requestToken !== previewRequestToken) {
          return;
        }

        previewProgress.value = Math.min(
          90,
          (content.length / Math.max(props.size, 1)) * 100,
        );
      }
      content += decoder.decode();
    } else {
      content = await response.text();
    }

    if (requestToken !== previewRequestToken) {
      return;
    }

    textContent.value = content;
    previewProgress.value = 100;
    writeTextCache(content);
    emitPreviewReady(requestToken);
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : "加载文本内容失败";
    setPreviewError(message, requestToken);
  } finally {
    if (requestToken === previewRequestToken) {
      previewLoading.value = false;
    }
  }
}

async function loadOfficeFileContent(requestToken: number) {
  if (!supportsOfficePreview.value || !canPreviewOffice.value) {
    return;
  }

  const cached = officePreviewCache.get(props.fileId);
  if (cached) {
    if (requestToken === previewRequestToken) {
      officeFileContent.value = cached;
      emitPreviewReady(requestToken);
    }
    return;
  }

  previewLoading.value = true;
  previewError.value = "";
  previewProgress.value = 0;

  try {
    const response = await fetch(buildFilePreviewURL("inline"));
    if (!response.ok) {
      throw new Error(getHttpErrorMessage(response.status));
    }

    const contentLengthHeader = response.headers.get("content-length");
    const total = Number.parseInt(contentLengthHeader || "", 10) || props.size;
    const reader = response.body?.getReader();
    let arrayBuffer: ArrayBuffer;

    if (reader) {
      const chunks: Uint8Array[] = [];
      let loaded = 0;

      while (true) {
        const { done, value } = await reader.read();
        if (done) {
          break;
        }

        chunks.push(value);
        loaded += value.length;
        if (requestToken !== previewRequestToken) {
          return;
        }

        previewProgress.value = Math.min(90, (loaded / Math.max(total, 1)) * 100);
      }

      const blob = new Blob(chunks as BlobPart[]);
      arrayBuffer = await blob.arrayBuffer();
    } else {
      arrayBuffer = await response.arrayBuffer();
    }

    if (requestToken !== previewRequestToken) {
      return;
    }

    officePreviewCache.set(props.fileId, arrayBuffer);
    officeFileContent.value = arrayBuffer;
    previewProgress.value = 100;
    emitPreviewReady(requestToken);
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : "加载 Office 文件失败";
    setPreviewError(message, requestToken);
  } finally {
    if (requestToken === previewRequestToken) {
      previewLoading.value = false;
    }
  }
}

async function loadPreview() {
  previewRequestToken += 1;
  const requestToken = previewRequestToken;
  resetFetchedPreviewState();

  if (!props.previewEnabled) {
    return;
  }

  if (supportsTextPreview.value && canPreviewText.value) {
    await loadTextContent(requestToken);
    return;
  }

  if (supportsOfficePreview.value && canPreviewOffice.value) {
    await loadOfficeFileContent(requestToken);
  }
}

function handleNativePreviewError() {
  const requestToken = ++previewRequestToken;
  setPreviewError("当前文件无法在浏览器中直接预览，请尝试下载后查看", requestToken);
}

watch(
  () => [
    props.fileId,
    props.fileName,
    props.extension,
    props.mimeType,
    props.size,
    props.previewEnabled,
  ],
  () => {
    if (props.fileId) {
      void loadPreview();
    } else {
      resetFetchedPreviewState();
    }
  },
  { immediate: true },
);

</script>

<template>
  <div
    v-if="!previewEnabled"
    class="flex min-h-[220px] items-center justify-center"
  >
    <p class="text-sm text-slate-400">预览已禁用</p>
  </div>

  <NativeFilePreview
    v-else-if="supportsImagePreview || supportsVideoPreview || supportsAudioPreview || supportsPdfPreview"
    :file-name="fileName"
    :inline-content-url="inlineContentURL"
    :preview-error="previewError"
    :preview-height="previewHeight"
    :supports-audio-preview="supportsAudioPreview"
    :supports-image-preview="supportsImagePreview"
    :supports-pdf-preview="supportsPdfPreview"
    :supports-video-preview="supportsVideoPreview"
    @native-error="handleNativePreviewError"
  />

  <div
    v-else-if="(supportsTextPreview && !canPreviewText) || (supportsOfficePreview && !canPreviewOffice)"
    class="flex min-h-[220px] items-center justify-center"
  >
    <div class="space-y-2 text-center">
      <p class="text-sm text-slate-400">{{ unsupportedReason }}</p>
      <p class="text-xs text-slate-500">
        文件大小: {{ (size / (1024 * 1024)).toFixed(2) }} MB
      </p>
    </div>
  </div>

  <OfficeFilePreview
    v-else-if="supportsOfficePreview"
    :normalized-extension="normalizedExtension"
    :office-file-content="officeFileContent"
    :preview-error="previewError"
    :preview-height="previewHeight"
    :preview-loading="previewLoading"
    :preview-progress="previewProgress"
    @retry="loadPreview"
  />

  <TextFilePreview
    v-else-if="supportsTextPreview"
    :is-markdown-file="isMarkdownFile"
    :preview-error="previewError"
    :preview-loading="previewLoading"
    :preview-progress="previewProgress"
    :text-content="textContent"
    @retry="loadPreview"
  />

  <div v-else class="flex min-h-[220px] items-center justify-center">
    <div class="space-y-2 text-center">
      <p class="text-sm text-slate-400">{{ unsupportedReason }}</p>
      <p class="text-xs text-slate-500">
        {{ extension?.toUpperCase() || "UNKNOWN" }}
      </p>
      <p class="text-xs text-slate-400">MIME: {{ mimeType || "未知" }}</p>
      <p class="text-xs text-slate-400">文件名: {{ fileName }}</p>
    </div>
  </div>
</template>
