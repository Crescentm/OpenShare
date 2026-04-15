package imports

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"openshare/backend/internal/model"
)

type FolderTreeFolderRow struct {
	ID            string
	ParentID      *string
	Name          string
	SourcePath    *string
	SyncState     string
	SyncError     string
	LastScannedAt *time.Time
}

type FolderTreeFileRow struct {
	ID            string
	FolderID      *string
	Name          string
	Size          int64
	DownloadCount int64
}

type ManagedRootRow struct {
	ID         string
	Name       string
	SourcePath *string
}

type ManagedSubtreeFolderRow struct {
	ID            string
	ParentID      *string
	Name          string
	Description   string
	SourcePath    *string
	FsDirMtimeNs  int64
	LastScannedAt *time.Time
	SyncState     string
	SyncError     string
	CreatedAt     time.Time
}

type ManagedSubtreeFileRow struct {
	ID             string
	FolderID       *string
	Name           string
	Description    string
	Extension      string
	MimeType       string
	Size           int64
	DownloadCount  int64
	FsFileMtimeNs  int64
	LastVerifiedAt *time.Time
	CreatedAt      time.Time
}

func (r *ImportRepository) ListFolders(ctx context.Context) ([]FolderTreeFolderRow, error) {
	var rows []FolderTreeFolderRow
	err := r.db.WithContext(ctx).
		Model(&model.Folder{}).
		Select("id, parent_id, name, source_path, sync_state, sync_error, last_scanned_at").
		Order("name ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}
	return rows, nil
}

func (r *ImportRepository) ListFiles(ctx context.Context) ([]FolderTreeFileRow, error) {
	var rows []FolderTreeFileRow
	err := r.db.WithContext(ctx).
		Model(&model.File{}).
		Select("id, folder_id, name, size, download_count").
		Order("name ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	return rows, nil
}

func (r *ImportRepository) FindFolderByID(ctx context.Context, folderID string) (*model.Folder, error) {
	var folder model.Folder
	err := r.db.WithContext(ctx).Where("id = ?", folderID).Take(&folder).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find folder by id: %w", err)
	}
	return &folder, nil
}

func (r *ImportRepository) ListManagedRoots(ctx context.Context) ([]ManagedRootRow, error) {
	var rows []ManagedRootRow
	if err := r.db.WithContext(ctx).
		Model(&model.Folder{}).
		Select("id, name, source_path").
		Where("parent_id IS NULL").
		Where("source_path IS NOT NULL AND TRIM(source_path) <> ''").
		Order("source_path ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list managed roots: %w", err)
	}
	return rows, nil
}

func (r *ImportRepository) ListManagedSubtreeFolders(ctx context.Context, rootFolderID string) ([]ManagedSubtreeFolderRow, error) {
	query := `
		WITH RECURSIVE folder_tree(id, parent_id, name, description, source_path, fs_dir_mtime_ns, last_scanned_at, sync_state, sync_error, created_at) AS (
			SELECT id, parent_id, name, description, source_path, fs_dir_mtime_ns, last_scanned_at, sync_state, sync_error, created_at
			FROM folders
			WHERE id = ?
			UNION ALL
			SELECT folders.id, folders.parent_id, folders.name, folders.description, folders.source_path, folders.fs_dir_mtime_ns, folders.last_scanned_at, folders.sync_state, folders.sync_error, folders.created_at
			FROM folders
			JOIN folder_tree ON folders.parent_id = folder_tree.id
		)
		SELECT id, parent_id, name, description, source_path, fs_dir_mtime_ns, last_scanned_at, sync_state, sync_error, created_at
		FROM folder_tree
	`

	var rows []ManagedSubtreeFolderRow
	if err := r.db.WithContext(ctx).Raw(query, rootFolderID).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list managed subtree folders: %w", err)
	}
	return rows, nil
}

func (r *ImportRepository) ListManagedSubtreeFiles(ctx context.Context, rootFolderID string) ([]ManagedSubtreeFileRow, error) {
	query := `
		WITH RECURSIVE folder_tree(id) AS (
			SELECT id
			FROM folders
			WHERE id = ?
			UNION ALL
			SELECT folders.id
			FROM folders
			JOIN folder_tree ON folders.parent_id = folder_tree.id
		)
		SELECT
			files.id,
			files.folder_id,
			files.name,
			files.description,
			files.extension,
			files.mime_type,
			files.size,
			files.download_count,
			files.fs_file_mtime_ns,
			files.last_verified_at,
			files.created_at
		FROM files
		JOIN folder_tree ON files.folder_id = folder_tree.id
	`

	var rows []ManagedSubtreeFileRow
	if err := r.db.WithContext(ctx).Raw(query, rootFolderID).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list managed subtree files: %w", err)
	}
	return rows, nil
}
