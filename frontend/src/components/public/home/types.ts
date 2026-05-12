export interface AnnouncementItem {
  id: string;
  title: string;
  content: string;
  is_pinned: boolean;
  creator: {
    id: string;
    username: string;
    display_name: string;
    avatar_url: string;
    role: string;
  };
  published_at?: string;
  updated_at: string;
}

export function announcementAuthorName(item: AnnouncementItem) {
  return (
    item.creator?.display_name?.trim()
    || item.creator?.username?.trim()
    || "OpenShare"
  );
}

export function announcementAuthorInitial(item: AnnouncementItem) {
  return announcementAuthorName(item).slice(0, 1).toUpperCase() || "A";
}

export function announcementAuthorIsSuperAdmin(item: AnnouncementItem) {
  return item.creator?.role === "super_admin";
}

export interface PublicFolderItem {
  id: string;
  name: string;
  updated_at: string;
  file_count: number;
  download_count: number;
  total_size: number;
}

export interface PublicFileItem {
  id: string;
  name: string;
  description: string;
  extension: string;
  uploaded_at: string;
  download_count: number;
  size: number;
}

export interface HotDownloadItem {
  id: string;
  name: string;
  downloadCount: number;
}

export interface LatestItem {
  id: string;
  name: string;
}

export interface SidebarDetailItem {
  id: string;
  label: string;
  meta?: string;
}

export interface SidebarDetailModalState {
  eyebrow: string;
  title: string;
  description: string;
  items: SidebarDetailItem[];
}

export interface SearchResultItem {
  entity_type: "file" | "folder";
  id: string;
  name: string;
  path?: string;
  path_segments?: string[];
  extension?: string;
  file_kind?: string;
  category?: string;
  course?: string;
  material_type?: string;
  content_status?: string;
  size?: number;
  download_count?: number;
  uploaded_at?: string;
  updated_at?: string;
  snippet?: string;
  highlights?: Record<string, string>;
}

export interface SearchResultResponse {
  items: SearchResultItem[];
  page: number;
  page_size: number;
  total: number;
}

export interface FolderDetailResponse {
  id: string;
  name: string;
  description: string;
  parent_id: string | null;
  file_count: number;
  download_count: number;
  total_size: number;
  updated_at: string;
  breadcrumbs: Array<{
    id: string;
    name: string;
  }>;
}

export interface DirectoryRow {
  id: string;
  kind: "folder" | "file";
  name: string;
  extension: string;
  description: string;
  downloadCount: number;
  fileCount: number;
  sizeText: string;
  updatedAt: string;
  downloadURL: string;
}

export type PublicHomeSortMode = "name" | "download" | "format";
export type PublicHomeSortDirection = "asc" | "desc";
export type PublicHomeViewMode = "cards" | "table";
