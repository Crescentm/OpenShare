import vue from "@vitejs/plugin-vue";
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";

function manualChunks(id: string) {
  if (!id.includes("node_modules")) {
    return;
  }

  if (id.includes("/node_modules/@vue-office/docx/")) {
    return "office-docx";
  }
  if (id.includes("/node_modules/@vue-office/excel/")) {
    return "office-excel";
  }
  if (id.includes("/node_modules/@vue-office/pptx/")) {
    return "office-pptx";
  }
  if (id.includes("/node_modules/@lucide/vue/")) {
    return "icons-lucide";
  }
  if (
    id.includes("/node_modules/markdown-it/") ||
    id.includes("/node_modules/dompurify/")
  ) {
    return "markdown";
  }
  if (
    id.includes("/node_modules/@vue/") ||
    id.includes("/node_modules/vue/") ||
    id.includes("/node_modules/vue-demi/") ||
    id.includes("/node_modules/vue-router/") ||
    id.includes("/node_modules/pinia/")
  ) {
    return "vue-core";
  }
  return "vendor";
}

export default defineConfig({
  plugins: [vue()],
  build: {
    rollupOptions: {
      output: {
        manualChunks,
      },
    },
  },
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    host: "0.0.0.0",
    port: 5173,
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true,
      },
    },
  },
});
