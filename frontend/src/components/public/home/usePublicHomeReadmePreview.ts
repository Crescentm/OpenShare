import { ref, type ComputedRef } from "vue";
import { renderSimpleMarkdown } from "../../../lib/markdown";
import type { PublicFileItem } from "./types";

function pickReadmeFile(entries: PublicFileItem[]) {
  const readmePatterns = [
    (name: string) => name.trim().toLowerCase() === "readme.md",
    (name: string) => name.trim().toLowerCase() === "readme.markdown",
    (name: string) => name.trim().toLowerCase() === "readme.txt",
    (name: string) => name.trim().toLowerCase().startsWith("readme"),
    (name: string) =>
      name.trim().toLowerCase().includes("readme")
      && (
        name.trim().toLowerCase().endsWith(".md")
        || name.trim().toLowerCase().endsWith(".markdown")
        || name.trim().toLowerCase().endsWith(".txt")
      ),
  ];

  for (const pattern of readmePatterns) {
    const file = entries.find((item) => pattern(item.name));
    if (file) {
      return file;
    }
  }

  return null;
}

function getHttpErrorMessage(status: number): string {
  switch (status) {
    case 403:
      return "No permission to access this file";
    case 404:
      return "File not found or has been deleted";
    case 410:
      return "File has been permanently deleted";
    case 429:
      return "Too many requests, please try again later";
    case 500:
      return "Internal server error";
    case 502:
      return "Bad gateway";
    case 503:
      return "Service temporarily unavailable";
    case 504:
      return "Gateway timeout";
    default:
      return `HTTP error ${status}`;
  }
}

export function usePublicHomeReadmePreview(
  currentFolderID: ComputedRef<string>,
  buildPublicFilePreviewURL: (
    fileID: string,
    view: "inline" | "text",
  ) => string,
) {
  const readmePreviewName = ref("");
  const readmePreviewHTML = ref("");
  const readmePreviewLoading = ref(false);
  const readmePreviewError = ref("");
  let readmePreviewRequestID = 0;

  function nextReadmePreviewRequestID() {
    readmePreviewRequestID += 1;
    return readmePreviewRequestID;
  }

  function isCurrentReadmeRequest(requestID: number) {
    return requestID === readmePreviewRequestID;
  }

  function resetReadmePreview() {
    readmePreviewName.value = "";
    readmePreviewHTML.value = "";
    readmePreviewLoading.value = false;
    readmePreviewError.value = "";
  }

  function buildReadmeAssetURL(rawURL: string) {
    if (!currentFolderID.value) {
      return rawURL;
    }

    const query = new URLSearchParams({ path: rawURL });
    return `/api/public/folders/${encodeURIComponent(currentFolderID.value)}/assets?${query.toString()}`;
  }

  async function loadReadmePreview(entries: PublicFileItem[], requestID: number) {
    if (!currentFolderID.value) {
      return;
    }

    const readmeFile = pickReadmeFile(entries);
    if (!readmeFile) {
      return;
    }

    readmePreviewName.value = readmeFile.name;
    readmePreviewLoading.value = true;
    readmePreviewError.value = "";

    try {
      const response = await fetch(buildPublicFilePreviewURL(readmeFile.id, "text"), {
        method: "GET",
        credentials: "include",
        headers: {
          Accept: "text/plain",
        },
      });

      if (!response.ok) {
        throw new Error(getHttpErrorMessage(response.status));
      }

      const content = await response.text();
      if (requestID !== readmePreviewRequestID) {
        return;
      }

      const fileName = readmeFile.name.toLowerCase();
      if (fileName.endsWith(".md") || fileName.endsWith(".markdown")) {
        readmePreviewHTML.value = renderSimpleMarkdown(content, {
          resolveURL: (rawURL) => buildReadmeAssetURL(rawURL),
        });
      } else {
        readmePreviewHTML.value = `<pre class="whitespace-pre-wrap break-words text-sm text-slate-700">${content.replace(/</g, "&lt;").replace(/>/g, "&gt;")}</pre>`;
      }
    } catch (error) {
      if (requestID !== readmePreviewRequestID) {
        return;
      }
      console.error("README preview error:", error);
      readmePreviewError.value = "README 预览加载失败。";
    } finally {
      if (requestID === readmePreviewRequestID) {
        readmePreviewLoading.value = false;
      }
    }
  }

  return {
    isCurrentReadmeRequest,
    loadReadmePreview,
    nextReadmePreviewRequestID,
    readmePreviewError,
    readmePreviewHTML,
    readmePreviewLoading,
    readmePreviewName,
    resetReadmePreview,
  };
}
