import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";

import AdminLayout from "@/layouts/AdminLayout.vue";
import AdminAdminsView from "@/views/admin/AdminAdminsView.vue";
import AdminAnnouncementsView from "@/views/admin/AdminAnnouncementsView.vue";
import PublicLayout from "@/layouts/PublicLayout.vue";
import AdminDashboardView from "@/views/admin/AdminDashboardView.vue";
import AdminReportsView from "@/views/admin/AdminReportsView.vue";
import AdminResourcesView from "@/views/admin/AdminResourcesView.vue";
import AdminSettingsView from "@/views/admin/AdminSettingsView.vue";
import AdminTagsView from "@/views/admin/AdminTagsView.vue";
import AdminOperationLogsView from "@/views/admin/AdminOperationLogsView.vue";
import PublicFileDetailView from "@/views/public/PublicFileDetailView.vue";
import HomeView from "@/views/public/HomeView.vue";
import SearchView from "@/views/public/SearchView.vue";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    component: PublicLayout,
    children: [
      {
        path: "",
        name: "public-home",
        component: HomeView,
      },
      {
        path: "search",
        name: "public-search",
        component: SearchView,
      },
      {
        path: "files/:fileID",
        name: "public-file-detail",
        component: PublicFileDetailView,
      },
    ],
  },
  {
    path: "/admin",
    component: AdminLayout,
    children: [
      {
        path: "",
        name: "admin-dashboard",
        component: AdminDashboardView,
      },
      {
        path: "announcements",
        name: "admin-announcements",
        component: AdminAnnouncementsView,
      },
      {
        path: "admins",
        name: "admin-admins",
        component: AdminAdminsView,
      },
      {
        path: "resources",
        name: "admin-resources",
        component: AdminResourcesView,
      },
      {
        path: "reports",
        name: "admin-reports",
        component: AdminReportsView,
      },
      {
        path: "tags",
        name: "admin-tags",
        component: AdminTagsView,
      },
      {
        path: "settings",
        name: "admin-settings",
        component: AdminSettingsView,
      },
      {
        path: "operation-logs",
        name: "admin-operation-logs",
        component: AdminOperationLogsView,
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 };
  },
});

export default router;
