import { HttpError } from "./client";

export function readApiError(error: unknown, fallback = "请求失败，请稍后重试。") {
  if (!(error instanceof HttpError) || typeof error.payload !== "object" || error.payload === null) {
    return fallback;
  }

  const payload = error.payload as Record<string, unknown>;
  if (typeof payload.error === "string" && payload.error.trim() !== "") {
    return payload.error;
  }

  return fallback;
}

export async function downloadBlobResponse(response: Response, filename: string) {
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  window.URL.revokeObjectURL(url);
}
