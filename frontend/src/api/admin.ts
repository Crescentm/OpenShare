import { http } from '@/utils/http'
import type { AdminInfo, FileInfo, PageData, Submission, Tag, Report, Announcement, OperationLog, LoginResponse } from '@/types'

// 认证 API
export const authApi = {
  login(data: { username: string; password: string }) {
    return http.post<LoginResponse>('/admin/login', data)
  },
  
  // 获取当前登录管理员信息
  getCurrentAdmin() {
    return http.get<AdminInfo>('/admin/me')
  },
  
  // 修改密码
  changePassword(data: { old_password: string; new_password: string }) {
    return http.post('/admin/password', data)
  },
  
  // 刷新 Token
  refreshToken() {
    return http.post<{ token: string; expires_at: number }>('/admin/refresh')
  },
  
  // 退出登录
  logout() {
    return http.post('/admin/logout')
  },
}

// 审核管理 API
export const submissionApi = {
  getList(params?: { page?: number; page_size?: number; status?: string }) {
    return http.get<PageData<Submission>>('/admin/submissions', { params })
  },
  
  approve(id: number) {
    return http.post(`/admin/submissions/${id}/approve`)
  },
  
  reject(id: number, reason: string) {
    return http.post(`/admin/submissions/${id}/reject`, { reason })
  },
}

// 资料管理 API
export const fileApi = {
  getList(params?: { page?: number; page_size?: number; status?: string }) {
    return http.get<PageData<FileInfo>>('/admin/files', { params })
  },
  
  update(id: string, data: { title?: string; description?: string; tags?: number[] }) {
    return http.put(`/admin/files/${id}`, data)
  },
  
  delete(id: string) {
    return http.delete(`/admin/files/${id}`)
  },
  
  offline(id: string) {
    return http.post(`/admin/files/${id}/offline`)
  },
}

// Tag 管理 API
export const tagApi = {
  getList() {
    return http.get<Tag[]>('/admin/tags')
  },
  
  create(data: { name: string }) {
    return http.post<Tag>('/admin/tags', data)
  },
  
  update(id: number, data: { name: string }) {
    return http.put(`/admin/tags/${id}`, data)
  },
  
  delete(id: number) {
    return http.delete(`/admin/tags/${id}`)
  },
}

// 举报管理 API
export const reportApi = {
  getList(params?: { page?: number; page_size?: number; status?: string }) {
    return http.get<PageData<Report>>('/admin/reports', { params })
  },
  
  approve(id: number) {
    return http.post(`/admin/reports/${id}/approve`)
  },
  
  reject(id: number) {
    return http.post(`/admin/reports/${id}/reject`)
  },
}

// 公告管理 API
export const announcementApi = {
  getList(params?: { page?: number; page_size?: number }) {
    return http.get<PageData<Announcement>>('/admin/announcements', { params })
  },
  
  create(data: { title: string; content: string }) {
    return http.post<Announcement>('/admin/announcements', data)
  },
  
  update(id: number, data: { title?: string; content?: string; visible?: boolean }) {
    return http.put(`/admin/announcements/${id}`, data)
  },
  
  delete(id: number) {
    return http.delete(`/admin/announcements/${id}`)
  },
}

// 管理员管理 API
export const adminApi = {
  getList() {
    return http.get<AdminInfo[]>('/admin/admins')
  },
  
  create(data: { username: string; password: string; permissions?: string[] }) {
    return http.post<AdminInfo>('/admin/admins', data)
  },
  
  update(id: number, data: { permissions?: string[] }) {
    return http.put(`/admin/admins/${id}`, data)
  },
  
  delete(id: number) {
    return http.delete(`/admin/admins/${id}`)
  },
}

// 操作日志 API
export const logApi = {
  getList(params?: { page?: number; page_size?: number; action?: string }) {
    return http.get<PageData<OperationLog>>('/admin/logs', { params })
  },
}

// 系统设置 API
export const settingsApi = {
  get() {
    return http.get<Record<string, unknown>>('/admin/settings')
  },
  
  update(data: Record<string, unknown>) {
    return http.put('/admin/settings', data)
  },
}
