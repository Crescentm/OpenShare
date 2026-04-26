import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";

const PublicLayout = () => import("@/layouts/PublicLayout.vue");
const HomeView = () => import("@/views/public/Home.vue");
const UploadView = () => import("@/views/public/UploadView.vue");
const PublicFileDetailView = () =>
  import("@/views/public/PublicFileDetailView.vue");

const AdminLayout = () => import("@/layouts/AdminLayout.vue");
const AdminDashboard = () => import("@/views/admin/AdminDashboard.vue");
const AdminAdminsView = () => import("@/views/admin/AdminAdminsView.vue");
const AdminAuditView = () => import("@/views/admin/AdminAuditView.vue");
const AdminOperationLogsView = () =>
  import("@/views/admin/AdminOperationLogsView.vue");
const AdminAnnouncementsView = () =>
  import("@/views/admin/AdminAnnouncementsView.vue");
const AdminAccountSettingsView = () =>
  import("@/views/admin/AdminAccountSettingsView.vue");

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
        path: "upload",
        name: "public-upload",
        component: UploadView,
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
        component: AdminDashboard,
      },
      {
        path: "admins",
        redirect: "/admin/permissions",
      },
      {
        path: "permissions",
        name: "admin-permissions",
        component: AdminAdminsView,
      },
      {
        path: "audit",
        name: "admin-audit",
        component: AdminAuditView,
      },
      {
        path: "logs",
        name: "admin-logs",
        component: AdminOperationLogsView,
      },
      {
        path: "announcements",
        name: "admin-announcements",
        component: AdminAnnouncementsView,
      },
      {
        path: "account",
        name: "admin-account",
        component: AdminAccountSettingsView,
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
