<script setup lang="ts">
import { onMounted, ref } from "vue";
import { CalendarClock, Download, Files } from "lucide-vue-next";

import AdminSuperadminControls from "../../components/AdminSuperadminControls.vue";
import StatCard from "../../components/StatCard.vue";
import { httpClient } from "../../lib/http/client";
import { useSessionStore } from "../../stores/session";

interface MetricItem {
  title: string;
  value: string | number;
  hint: string;
  icon: typeof Files;
}

interface DashboardStatsResponse {
  total_files: number;
  total_downloads: number;
  recent_files: number;
  recent_downloads: number;
}

const sessionStore = useSessionStore();
const loading = ref(true);
const metrics = ref<MetricItem[]>([
  {
    title: "总资料数",
    value: "--",
    hint: "",
    icon: Files,
  },
  {
    title: "总下载数",
    value: "--",
    hint: "",
    icon: Download,
  },
  {
    title: "近7天新增资料数",
    value: "--",
    hint: "",
    icon: CalendarClock,
  },
  {
    title: "近7天下载数",
    value: "--",
    hint: "",
    icon: Download,
  },
]);

onMounted(async () => {
  await loadMetrics();
});

async function loadMetrics() {
  loading.value = true;
  await loadDashboardStats();
  loading.value = false;
}

async function loadDashboardStats() {
  try {
    const response = await httpClient.get<DashboardStatsResponse>("/admin/dashboard/stats");
    setMetric("总资料数", response.total_files);
    setMetric("总下载数", response.total_downloads);
    setMetric("近7天新增资料数", response.recent_files);
    setMetric("近7天下载数", response.recent_downloads);
  } catch {
    setMetric("总资料数", "--");
    setMetric("总下载数", "--");
    setMetric("近7天新增资料数", "--");
    setMetric("近7天下载数", "--");
  }
}

function setMetric(title: string, value: string | number) {
  const metric = metrics.value.find((item) => item.title === title);
  if (!metric) return;
  metric.value = value;
  metric.hint = "";
}
</script>

<template>
  <section class="space-y-6">
    <header class="space-y-2">
      <div class="space-y-2">
        <p class="text-xs font-semibold uppercase tracking-[0.12em] text-blue-600">CONSOLE</p>
        <h1 class="text-2xl font-semibold tracking-tight text-slate-900 dark:text-slate-100">控制台</h1>
      </div>
    </header>

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <StatCard
        v-for="metric in metrics"
        :key="metric.title"
        :title="metric.title"
        :value="metric.value"
        :hint="metric.hint"
      >
        <template #icon>
          <component :is="metric.icon" class="h-4 w-4" />
        </template>
      </StatCard>
    </section>

    <p v-if="loading" class="text-sm text-slate-500 dark:text-slate-400">正在刷新控制台数据…</p>

    <AdminSuperadminControls v-if="sessionStore.isSuperAdmin" />
  </section>
</template>
