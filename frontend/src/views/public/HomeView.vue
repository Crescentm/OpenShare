<script setup lang="ts">
import { onMounted, ref } from "vue";

import { HttpError, httpClient } from "../../lib/http/client";

interface SubmissionLookupResult {
  receipt_code: string;
  title: string;
  status: "pending" | "approved" | "rejected";
  uploaded_at: string;
  download_count: number;
  reject_reason: string;
}

const cachedReceiptCodeKey = "openshare:last_receipt_code";

const receiptCode = ref("");
const record = ref<SubmissionLookupResult | null>(null);
const loading = ref(false);
const errorMessage = ref("");

onMounted(() => {
  const cached = window.localStorage.getItem(cachedReceiptCodeKey);
  if (!cached) {
    return;
  }

  receiptCode.value = cached;
  void lookupSubmission();
});

async function lookupSubmission() {
  const normalized = receiptCode.value.trim();
  if (!normalized) {
    errorMessage.value = "请输入回执码。";
    record.value = null;
    return;
  }

  loading.value = true;
  errorMessage.value = "";

  try {
    const response = await httpClient.get<SubmissionLookupResult>(`/public/submissions/${encodeURIComponent(normalized)}`);
    record.value = response;
    window.localStorage.setItem(cachedReceiptCodeKey, response.receipt_code);
    receiptCode.value = response.receipt_code;
  } catch (error: unknown) {
    record.value = null;
    if (error instanceof HttpError && error.status === 404) {
      errorMessage.value = "未找到对应回执码，请检查输入是否正确。";
    } else if (error instanceof HttpError && error.status === 400) {
      errorMessage.value = "回执码格式无效。";
    } else {
      errorMessage.value = "查询失败，请稍后重试。";
    }
  } finally {
    loading.value = false;
  }
}

function statusLabel(status: SubmissionLookupResult["status"]) {
  switch (status) {
    case "approved":
      return "已通过";
    case "rejected":
      return "已驳回";
    default:
      return "待审核";
  }
}

function statusClass(status: SubmissionLookupResult["status"]) {
  switch (status) {
    case "approved":
      return "bg-emerald-100 text-emerald-800";
    case "rejected":
      return "bg-rose-100 text-rose-800";
    default:
      return "bg-amber-100 text-amber-800";
  }
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
</script>

<template>
  <section class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
    <article class="rounded-[28px] bg-slate-950 px-8 py-10 text-white">
      <p class="text-sm font-semibold uppercase tracking-[0.28em] text-blue-300">Public Portal</p>
      <h2 class="mt-4 max-w-xl text-4xl font-semibold leading-tight">
        资料查询与投稿链路已经开始闭环。
      </h2>
      <p class="mt-4 max-w-2xl text-base text-slate-300">
        你现在可以用回执码直接查询最近一次投稿记录。浏览器会自动缓存最近成功查询的回执码，方便再次打开页面时继续查看。
      </p>
    </article>

    <article class="rounded-[28px] border border-slate-200 bg-white px-6 py-7 shadow-sm">
      <div class="flex items-center justify-between gap-4">
        <div>
          <p class="text-sm font-semibold uppercase tracking-[0.24em] text-blue-700">Receipt Lookup</p>
          <h3 class="mt-2 text-2xl font-semibold text-slate-900">我的上传</h3>
        </div>
        <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600">最近回执码自动缓存</span>
      </div>

      <form class="mt-6 space-y-4" @submit.prevent="lookupSubmission">
        <label class="block">
          <span class="mb-2 block text-sm font-medium text-slate-700">回执码</span>
          <input
            v-model="receiptCode"
            type="text"
            placeholder="例如：A8K2D7Q4M9P1"
            class="w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-blue-500 focus:bg-white"
          />
        </label>

        <button
          type="submit"
          class="inline-flex items-center justify-center rounded-2xl bg-blue-700 px-5 py-3 text-sm font-semibold text-white transition hover:bg-blue-800 disabled:cursor-not-allowed disabled:bg-slate-400"
          :disabled="loading"
        >
          {{ loading ? "查询中..." : "查询投稿记录" }}
        </button>
      </form>

      <p v-if="errorMessage" class="mt-4 rounded-2xl bg-rose-50 px-4 py-3 text-sm text-rose-700">
        {{ errorMessage }}
      </p>

      <div v-if="record" class="mt-5 rounded-[24px] bg-slate-50 p-5">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p class="text-xs uppercase tracking-[0.2em] text-slate-500">投稿标题</p>
            <h4 class="mt-2 text-lg font-semibold text-slate-900">{{ record.title }}</h4>
          </div>
          <span class="rounded-full px-3 py-1 text-xs font-semibold" :class="statusClass(record.status)">
            {{ statusLabel(record.status) }}
          </span>
        </div>

        <dl class="mt-5 grid gap-4 sm:grid-cols-2">
          <div class="rounded-2xl bg-white px-4 py-3">
            <dt class="text-xs uppercase tracking-[0.18em] text-slate-500">回执码</dt>
            <dd class="mt-2 text-sm font-medium text-slate-900">{{ record.receipt_code }}</dd>
          </div>
          <div class="rounded-2xl bg-white px-4 py-3">
            <dt class="text-xs uppercase tracking-[0.18em] text-slate-500">上传时间</dt>
            <dd class="mt-2 text-sm font-medium text-slate-900">{{ formatDate(record.uploaded_at) }}</dd>
          </div>
          <div class="rounded-2xl bg-white px-4 py-3">
            <dt class="text-xs uppercase tracking-[0.18em] text-slate-500">下载量</dt>
            <dd class="mt-2 text-sm font-medium text-slate-900">{{ record.download_count }}</dd>
          </div>
          <div class="rounded-2xl bg-white px-4 py-3">
            <dt class="text-xs uppercase tracking-[0.18em] text-slate-500">驳回原因</dt>
            <dd class="mt-2 text-sm font-medium text-slate-900">
              {{ record.reject_reason || "暂无" }}
            </dd>
          </div>
        </dl>
      </div>
    </article>
  </section>
</template>
