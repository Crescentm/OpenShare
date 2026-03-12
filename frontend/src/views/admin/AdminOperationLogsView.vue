<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import { httpClient } from "../../lib/http/client";
import { readApiError } from "../../lib/http/helpers";

interface OperationLogItem {
  id: string;
  admin_id?: string | null;
  admin_name: string;
  action: string;
  target_type: string;
  target_id: string;
  detail: string;
  ip: string;
  created_at: string;
}

const items = ref<OperationLogItem[]>([]);
const loading = ref(false);
const error = ref("");
const actionFilter = ref("");
const targetTypeFilter = ref("");
const page = ref(1);
const pageSize = 20;
const total = ref(0);

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)));

onMounted(() => {
  void loadItems();
});

async function loadItems() {
  loading.value = true;
  error.value = "";
  try {
    const params = new URLSearchParams({
      page: String(page.value),
      page_size: String(pageSize),
    });
    if (actionFilter.value.trim()) params.set("action", actionFilter.value.trim());
    if (targetTypeFilter.value.trim()) params.set("target_type", targetTypeFilter.value.trim());

    const response = await httpClient.get<{ items: OperationLogItem[]; total: number }>(`/admin/operation-logs?${params.toString()}`);
    items.value = response.items ?? [];
    total.value = response.total ?? 0;
  } catch (err: unknown) {
    error.value = readApiError(err, "加载审计日志失败。");
  } finally {
    loading.value = false;
  }
}

function goToPage(next: number) {
  if (next < 1 || next > totalPages.value) return;
  page.value = next;
  void loadItems();
}

function applyFilters() {
  page.value = 1;
  void loadItems();
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
</script>

<template>
  <section class="space-y-8">
    <header>
      <p class="text-sm font-semibold uppercase tracking-[0.22em] text-blue-300">Audit Logs</p>
      <h2 class="mt-2 text-3xl font-semibold text-white">审计日志</h2>
      <p class="mt-2 text-sm text-slate-400">当前所有管理员均可查看操作留痕。</p>
    </header>

    <article class="rounded-[28px] border border-slate-800 bg-slate-950/70 p-6">
      <div class="grid gap-4 lg:grid-cols-[1fr_220px_auto]">
        <input
          v-model="actionFilter"
          placeholder="按 action 过滤"
          class="rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-sm text-white outline-none focus:border-blue-400"
        />
        <input
          v-model="targetTypeFilter"
          placeholder="按 target_type 过滤"
          class="rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-sm text-white outline-none focus:border-blue-400"
        />
        <button class="rounded-2xl bg-blue-500 px-5 py-3 text-sm font-semibold text-slate-950" @click="applyFilters">
          查询
        </button>
      </div>

      <p v-if="error" class="mt-4 rounded-2xl bg-rose-950/60 px-4 py-3 text-sm text-rose-200">{{ error }}</p>
      <p v-else-if="loading" class="mt-4 text-sm text-slate-400">加载中...</p>

      <div v-else class="mt-6 space-y-4">
        <article
          v-for="item in items"
          :key="item.id"
          class="rounded-[22px] border border-slate-800 bg-slate-900 p-5"
        >
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div class="space-y-2">
              <div class="flex flex-wrap items-center gap-2">
                <span class="rounded-full bg-blue-500 px-3 py-1 text-xs font-semibold text-slate-950">{{ item.action }}</span>
                <span class="rounded-full border border-slate-700 px-3 py-1 text-xs text-slate-300">{{ item.target_type }} / {{ item.target_id || "-" }}</span>
              </div>
              <p class="text-sm text-slate-300">操作人：{{ item.admin_name || "guest/system" }}</p>
              <p class="text-sm text-slate-400">详情：{{ item.detail || "-" }}</p>
            </div>
            <div class="text-right text-sm text-slate-500">
              <p>{{ formatDate(item.created_at) }}</p>
              <p class="mt-1">IP: {{ item.ip || "-" }}</p>
            </div>
          </div>
        </article>

        <p v-if="items.length === 0" class="rounded-2xl bg-slate-900 px-4 py-6 text-sm text-slate-400">
          当前没有匹配的审计日志。
        </p>

        <div v-if="totalPages > 1" class="flex items-center justify-center gap-3 pt-2">
          <button class="rounded-xl border border-slate-700 px-4 py-2 text-sm text-slate-200 disabled:text-slate-600" :disabled="page <= 1" @click="goToPage(page - 1)">
            上一页
          </button>
          <span class="text-sm text-slate-400">{{ page }} / {{ totalPages }}</span>
          <button class="rounded-xl border border-slate-700 px-4 py-2 text-sm text-slate-200 disabled:text-slate-600" :disabled="page >= totalPages" @click="goToPage(page + 1)">
            下一页
          </button>
        </div>
      </div>
    </article>
  </section>
</template>
