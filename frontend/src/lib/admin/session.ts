import { HttpError, httpClient } from "../http/client";

interface AdminMeResponse {
  admin: {
    role: string;
    permissions: string[];
  };
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
