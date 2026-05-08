const zhCN = "zh-CN";

export function formatDateTime(value: string) {
  return new Intl.DateTimeFormat(zhCN, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function formatDateOnly(value: string) {
  return new Intl.DateTimeFormat(zhCN, {
    dateStyle: "medium",
  }).format(new Date(value));
}

export function formatNumericDateTime(value: string) {
  return new Intl.DateTimeFormat(zhCN, {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  }).format(new Date(value));
}

export function formatNumericMinuteDateTime(value: string) {
  return new Intl.DateTimeFormat(zhCN, {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(new Date(value));
}

export function formatOptionalNumericMinuteDateTime(
  value?: string,
  fallback = "未发布",
) {
  return value ? formatNumericMinuteDateTime(value) : fallback;
}

export function formatFileSize(size: number, fractionDigits = 1) {
  if (size < 1024) {
    return `${size} B`;
  }
  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(fractionDigits)} KB`;
  }
  if (size < 1024 * 1024 * 1024) {
    return `${(size / (1024 * 1024)).toFixed(fractionDigits)} MB`;
  }
  return `${(size / (1024 * 1024 * 1024)).toFixed(fractionDigits)} GB`;
}

export function formatPreciseFileSize(size: number) {
  return formatFileSize(size, 2);
}
