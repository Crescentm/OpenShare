<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";
import { LayoutDashboard, Inbox, Megaphone, ScrollText, Shield, UserRound } from "@lucide/vue";

import AdminSidebar, { type AdminSidebarItem } from "../components/admin/AdminSidebar.vue";
import {
  loadAdminPendingAuditCount,
  loginAdminSession,
  logoutAdminSession,
  restoreAdminSession,
} from "../lib/admin/session";
import { httpClient } from "../lib/http/client";
import { readApiError } from "../lib/http/helpers";
import { useSessionStore } from "../stores/session";

const sessionStore = useSessionStore();
const route = useRoute();

const username = ref("");
const password = ref("");
const loading = ref(true);
const loginLoading = ref(false);
const loginError = ref("");

const navItems = computed<AdminSidebarItem[]>(() => [
  { to: "/admin", label: "控制台", icon: LayoutDashboard },
  { to: "/admin/audit", label: "审核", icon: Inbox, hasAlert: sessionStore.pendingAuditCount > 0 },
  ...(sessionStore.hasPermission("announcements") ? [{ to: "/admin/announcements", label: "公告", icon: Megaphone }] : []),
  { to: "/admin/logs", label: "操作记录", icon: ScrollText },
  { to: "/admin/permissions", label: "权限管理", icon: Shield },
]);

onMounted(async () => {
  window.addEventListener("admin-pending-audit-refresh", handlePendingAuditRefresh);
  await restoreSession();
});

onBeforeUnmount(() => {
  window.removeEventListener("admin-pending-audit-refresh", handlePendingAuditRefresh);
});

async function restoreSession() {
  loading.value = true;
  try {
    if (await restoreAdminSession()) {
      await trackVisit();
    }
  } catch {
    sessionStore.reset();
  } finally {
    loading.value = false;
  }
}

async function login() {
  loginLoading.value = true;
  loginError.value = "";

  try {
    await loginAdminSession(username.value, password.value);
    await trackVisit();
    password.value = "";
  } catch (error: unknown) {
    loginError.value = readApiError(error, "登录失败，请重试。");
  } finally {
    loginLoading.value = false;
  }
}

async function logout() {
  await logoutAdminSession();
}

function handlePendingAuditRefresh() {
  void loadAdminPendingAuditCount();
}

async function trackVisit() {
  try {
    await httpClient.request("/visits", {
      method: "POST",
      body: {
        scope: "admin",
        path: route.path,
      },
    });
  } catch {
    // Ignore analytics failures.
  }
}
</script>

<template>
  <div class="app-shell">
    <div v-if="loading" class="flex min-h-screen items-center justify-center">
      <p class="text-sm text-slate-500">正在加载管理后台…</p>
    </div>

    <div v-else-if="!sessionStore.authenticated" class="app-container flex min-h-screen items-center justify-center py-10 sm:py-16">
      <section class="panel w-full max-w-[420px] p-5 sm:p-8">
        <div class="space-y-2">
          <p class="text-sm font-semibold text-slate-600 dark:text-slate-400">OpenShare Admin</p>
          <h2 class="text-3xl font-semibold tracking-tight text-slate-900 dark:text-slate-100">管理员登录</h2>
        </div>

        <form class="mt-8 space-y-4" @submit.prevent="login">
            <div class="space-y-2">
              <label class="text-sm font-medium text-slate-700 dark:text-slate-300">标示ID</label>
              <input v-model="username" class="field" placeholder="请输入标示ID" />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-medium text-slate-700 dark:text-slate-300">密码</label>
              <input v-model="password" type="password" class="field" placeholder="请输入密码" />
            </div>

            <button type="submit" class="btn-primary h-11 w-full" :disabled="loginLoading">
              {{ loginLoading ? "登录中…" : "登录后台" }}
            </button>

            <RouterLink
              to="/"
              class="flex h-11 w-full items-center justify-center rounded-xl border border-slate-200 bg-white text-sm font-semibold text-slate-900 transition hover:border-slate-300 hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:border-slate-600 dark:hover:bg-slate-800"
            >
              返回前台
            </RouterLink>
          </form>

        <p v-if="loginError" class="mt-4 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
            {{ loginError }}
        </p>

      </section>
    </div>

    <div v-else class="flex min-h-screen flex-col bg-[#fafafa] dark:bg-slate-950 lg:flex-row">
      <aside class="z-20 w-full lg:fixed lg:inset-y-0 lg:left-0 lg:w-[240px]">
        <AdminSidebar
          :current-path="route.path"
          :items="navItems"
          :title="sessionStore.displayName"
          subtitle=""
          :avatar-url="sessionStore.avatarUrl"
          :avatar-fallback="sessionStore.displayName.slice(0, 1).toUpperCase() || 'A'"
          @logout="logout"
        >
          <template #footer-actions>
            <RouterLink
              to="/admin/account"
              class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition"
              :class="route.path.startsWith('/admin/account') ? 'bg-slate-200/70 text-slate-900 dark:bg-slate-800 dark:text-slate-100' : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-400 dark:hover:bg-slate-900 dark:hover:text-slate-100'"
            >
              <UserRound class="h-4 w-4 shrink-0" />
              <span>账号设置</span>
            </RouterLink>
          </template>
        </AdminSidebar>
      </aside>

      <div class="min-w-0 flex-1 lg:pl-[240px]">
        <main class="mx-auto w-full max-w-[1240px] px-4 py-5 sm:px-6 sm:py-8 lg:px-8">
          <RouterView />
        </main>
      </div>
    </div>
  </div>
</template>
