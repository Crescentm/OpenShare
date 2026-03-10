// API 响应结构
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

// 分页响应
export interface PageData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

// 管理员信息
export interface AdminInfo {
  id: number
  username: string
  role: 'admin' | 'super_admin'
  permissions?: string[]
  created_at: string
}

// 文件信息
export interface FileInfo {
  id: string
  title: string
  description?: string
  filename: string
  size: number
  mime_type: string
  tags: Tag[]
  download_count: number
  status: 'active' | 'offline' | 'deleted'
  created_at: string
  updated_at: string
}

// 文件夹信息
export interface FolderInfo {
  id: string
  name: string
  parent_id?: string
  tags: Tag[]
  file_count: number
  created_at: string
}

// Tag 信息
export interface Tag {
  id: number
  name: string
  file_count?: number
}

// 投稿记录
export interface Submission {
  id: number
  title: string
  description?: string
  filename: string
  receipt_code: string
  status: 'pending' | 'approved' | 'rejected'
  reject_reason?: string
  download_count?: number
  created_at: string
}

// 举报记录
export interface Report {
  id: number
  target_type: 'file' | 'folder'
  target_id: string
  reason: string
  description?: string
  status: 'pending' | 'approved' | 'rejected'
  created_at: string
}

// 公告
export interface Announcement {
  id: number
  title: string
  content: string
  visible: boolean
  created_at: string
  updated_at: string
}

// 操作日志
export interface OperationLog {
  id: number
  admin_id?: number
  admin_username?: string
  action: string
  target_type: string
  target_id: string
  ip: string
  result: string
  created_at: string
}
