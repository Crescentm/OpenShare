import { http } from '@/utils/http'
import type { FileInfo, PageData, Submission, Tag, Announcement } from '@/types'

// 文件相关 API
export const fileApi = {
  // 获取文件列表
  getList(params?: { page?: number; page_size?: number; folder_id?: string; tag_id?: number }) {
    return http.get<PageData<FileInfo>>('/files', { params })
  },
  
  // 获取文件详情
  getDetail(id: string) {
    return http.get<FileInfo>(`/files/${id}`)
  },
  
  // 下载文件
  download(id: string) {
    return http.download(`/files/${id}/download`)
  },
  
  // 上传文件
  upload(data: FormData, onProgress?: (percent: number) => void) {
    return http.upload<{ receipt_code: string }>('/files/upload', data, onProgress)
  },
}

// 搜索 API
export const searchApi = {
  search(params: { q: string; tags?: number[]; folder_id?: string; page?: number; page_size?: number }) {
    return http.get<PageData<FileInfo>>('/search', { params })
  },
}

// 投稿记录 API
export const submissionApi = {
  // 根据回执码查询
  getByReceiptCode(receiptCode: string) {
    return http.get<Submission[]>('/submissions', { params: { receipt_code: receiptCode } })
  },
}

// Tag API
export const tagApi = {
  getList() {
    return http.get<Tag[]>('/tags')
  },
}

// 公告 API
export const announcementApi = {
  getList() {
    return http.get<Announcement[]>('/announcements')
  },
}

// 举报 API
export const reportApi = {
  submit(data: { target_type: 'file' | 'folder'; target_id: string; reason: string; description?: string }) {
    return http.post('/reports', data)
  },
}
