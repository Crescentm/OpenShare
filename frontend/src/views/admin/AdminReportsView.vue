<script setup lang="ts">
import { onMounted, ref } from "vue";

import { httpClient } from "../../lib/http/client";
import { readApiError } from "../../lib/http/helpers";
import { useSessionStore } from "../../stores/session";

interface PendingReportItem {
  id: string;
  file_id: string | null;
  folder_id: string | null;
  target_name: string;
  target_type: "file" | "folder";
  reason: string;
  reason_label: string;
  description: string;
  reporter_ip: string;
  status: string;
  created_at: string;
}

const reports = ref<PendingReportItem[]>([]);
const loading = ref(false);
const error = ref("");
const actionError = ref("");
const actionMessage = ref("");
const sessionStore = useSessionStore();

onMounted(() => {
  if (sessionStore.hasPermission("review_reports")) {
    void loadReports();
  }
});

async function loadReports() {
  loading.value = true;
  error.value = "";
  try {
    const response = await httpClient.get<{ items: PendingReportItem[] }>("/admin/reports/pending");
    reports.value = response.items ?? [];
  } catch {
    error.value = "加载举报列表失败。";
  } finally {
    loading.value = false;
  }
}

async function approveReport(reportId: string) {
  const reviewReason = window.prompt("确认举报成立，资源将被下架。请输入处理说明（可选）：");
  if (reviewReason === null) return; // cancelled

  actionError.value = "";
  actionMessage.value = "";
  try {
    await httpClient.post(`/admin/reports/${reportId}/approve`, {
      review_reason: reviewReason,
    });
    actionMessage.value = "举报已处理，目标资源已下架。";
    await loadReports();
  } catch (err: unknown) {
    actionError.value = readApiError(err, "操作失败，请重试。");
  }
}

async function rejectReport(reportId: string) {
  const reviewReason = window.prompt("驳回举报，资源保持可见。请输入驳回说明（可选）：");
  if (reviewReason === null) return; // cancelled

  actionError.value = "";
  actionMessage.value = "";
  try {
    await httpClient.post(`/admin/reports/${reportId}/reject`, {
      review_reason: reviewReason,
    });
    actionMessage.value = "举报已驳回，资源保持公开。";
    await loadReports();
  } catch (err: unknown) {
    actionError.value = readApiError(err, "操作失败，请重试。");
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
  <section class="space-y-6">
    <header class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <p class="text-sm font-semibold uppercase tracking-[0.28em] text-blue-300">Reports</p>
        <h2 class="mt-3 text-3xl font-semibold text-white">举报管理</h2>
      </div>
      <button
        class="rounded-2xl border border-slate-700 px-4 py-3 text-sm font-medium text-slate-200 transition hover:bg-slate-800"
        @click="loadReports"
      >
        刷新
      </button>
    </header>

    <p v-if="actionError" class="rounded-2xl bg-rose-950/50 px-4 py-3 text-sm text-rose-200">
      {{ actionError }}
    </p>
    <p v-if="actionMessage" class="rounded-2xl bg-emerald-950/60 px-4 py-3 text-sm text-emerald-200">
      {{ actionMessage }}
    </p>

    <p
      v-if="!sessionStore.hasPermission('review_reports')"
      class="rounded-2xl bg-slate-950/70 px-4 py-3 text-sm text-slate-400"
    >
      当前账号没有举报处理权限。
    </p>

    <p v-else-if="error" class="rounded-2xl bg-rose-950/50 px-4 py-3 text-sm text-rose-200">
      {{ error }}
    </p>
    <p v-else-if="loading" class="text-sm text-slate-400">加载中...</p>

    <div v-else class="space-y-4">
      <article
        v-for="report in reports"
        :key="report.id"
        class="rounded-[22px] border border-slate-800 bg-slate-950/70 p-5"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span
                class="shrink-0 rounded-lg px-2 py-0.5 text-xs font-semibold"
                :class="report.target_type === 'file' ? 'bg-blue-900/50 text-blue-300' : 'bg-amber-900/50 text-amber-300'"
              >
                {{ report.target_type === "file" ? "文件" : "文件夹" }}
              </span>
              <h4 class="truncate text-lg font-semibold text-white">{{ report.target_name }}</h4>
            </div>

            <div class="mt-3 flex flex-wrap gap-3 text-sm text-slate-400">
              <span class="inline-flex items-center gap-1.5 rounded-full bg-rose-900/30 px-3 py-1 text-xs font-medium text-rose-300">
                {{ report.reason_label || report.reason }}
              </span>
              <span>举报时间：{{ formatDate(report.created_at) }}</span>
              <span class="text-slate-600">IP: {{ report.reporter_ip }}</span>
            </div>

            <p v-if="report.description" class="mt-3 rounded-xl bg-slate-900 px-4 py-3 text-sm text-slate-300">
              {{ report.description }}
            </p>
          </div>

          <div class="flex shrink-0 flex-col gap-2">
            <button
              class="rounded-xl bg-rose-500 px-4 py-2 text-sm font-semibold text-white transition hover:bg-rose-400"
              @click="approveReport(report.id)"
            >
              确认下架
            </button>
            <button
              class="rounded-xl border border-slate-700 px-4 py-2 text-sm font-semibold text-slate-300 transition hover:bg-slate-800"
              @click="rejectReport(report.id)"
            >
              驳回
            </button>
          </div>
        </div>
      </article>

      <p
        v-if="reports.length === 0"
        class="rounded-2xl border border-dashed border-slate-700 px-4 py-8 text-center text-sm text-slate-400"
      >
        当前没有待处理的举报。
      </p>
    </div>
  </section>
</template>
