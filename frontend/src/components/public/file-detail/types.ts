export interface PublicFileDetailResponse {
  id: string;
  name: string;
  extension: string;
  folder_id: string;
  path: string;
  description: string;
  mime_type: string;
  size: number;
  uploaded_at: string;
  download_count: number;
}

export interface FileMetadataRow {
  label: string;
  value: string;
}
