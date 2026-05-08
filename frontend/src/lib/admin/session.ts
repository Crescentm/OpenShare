import { HttpError, httpClient } from "../http/client";
import { useSessionStore } from "../../stores/session";

export interface AdminMeResponse {
  admin: {
    id: string;
    username: string;
    display_name: string;
    avatar_url: string;
    role: string;
    status: string;
    permissions: string[];
  };
}

interface AdminDashboardStatsResponse {
  pending_audit_count: number;
}

export function applyAdminSession(response: AdminMeResponse) {
  const sessionStore = useSessionStore();
  sessionStore.setAuthenticated(
    true,
    response.admin.display_name || response.admin.username,
    {
      username: response.admin.username,
      adminId: response.admin.id,
      avatarUrl: response.admin.avatar_url,
      role: response.admin.role,
      status: response.admin.status,
      permissions: response.admin.permissions,
    },
  );
}

export async function loadAdminPendingAuditCount() {
  const sessionStore = useSessionStore();
  try {
    const response = await httpClient.get<AdminDashboardStatsResponse>(
      "/admin/dashboard/stats",
    );
    sessionStore.setPendingAuditCount(response.pending_audit_count ?? 0);
    return sessionStore.pendingAuditCount;
  } catch {
    sessionStore.setPendingAuditCount(0);
    return 0;
  }
}

export async function restoreAdminSession() {
  const sessionStore = useSessionStore();
  try {
    const response = await httpClient.get<AdminMeResponse>("/admin/me");
    applyAdminSession(response);
    await loadAdminPendingAuditCount();
    return true;
  } catch {
    sessionStore.reset();
    sessionStore.setPendingAuditCount(0);
    return false;
  }
}

export async function loginAdminSession(username: string, password: string) {
  const response = await httpClient.post<AdminMeResponse>("/admin/session/login", {
    username,
    password,
  });
  applyAdminSession(response);
  await loadAdminPendingAuditCount();
  return response;
}

export async function logoutAdminSession() {
  const sessionStore = useSessionStore();
  await httpClient.post("/admin/session/logout");
  sessionStore.setPendingAuditCount(0);
  sessionStore.reset();
}

export async function hasAdminPermission(permission: string) {
  try {
    const response = await httpClient.get<AdminMeResponse>("/admin/me");
    return response.admin.role === "super_admin" || (response.admin.permissions ?? []).includes(permission);
  } catch (error: unknown) {
    if (error instanceof HttpError && error.status === 401) {
      return false;
    }
    return false;
  }
}
