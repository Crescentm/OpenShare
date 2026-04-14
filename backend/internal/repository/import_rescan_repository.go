package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"openshare/backend/internal/model"
	"openshare/backend/pkg/identity"
)

type ManagedFolderUpdate struct {
	ID             string
	ParentID       *string
	Name           string
	SourcePath     string
	FsDirMtimeNs   int64
	LastScannedAt  *time.Time
	SyncState      string
	SyncError      string
	TouchUpdatedAt bool
}

type ManagedFileUpdate struct {
	ID             string
	FolderID       *string
	Name           string
	Description    string
	Extension      string
	MimeType       string
	Size           int64
	FsFileMtimeNs  int64
	LastVerifiedAt *time.Time
	TouchUpdatedAt bool
}

type RescanSyncInput struct {
	RootFolderID     string
	OperatorID       string
	OperatorIP       string
	Detail           string
	Now              time.Time
	AddedFolders     []*model.Folder
	UpdatedFolders   []ManagedFolderUpdate
	DeletedFolderIDs []string
	AddedFiles       []*model.File
	UpdatedFiles     []ManagedFileUpdate
	DeletedFileIDs   []string
}

func (r *ImportRepository) ApplyRescanSync(ctx context.Context, input RescanSyncInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(input.DeletedFileIDs) > 0 || len(input.DeletedFolderIDs) > 0 {
			if err := detachDeletedResourcesTx(tx, input.DeletedFileIDs, input.DeletedFolderIDs); err != nil {
				return err
			}
			if len(input.DeletedFileIDs) > 0 {
				if err := tx.Where("id IN ?", input.DeletedFileIDs).Delete(&model.File{}).Error; err != nil {
					return fmt.Errorf("delete rescanned files: %w", err)
				}
			}
			if len(input.DeletedFolderIDs) > 0 {
				if err := tx.Where("id IN ?", input.DeletedFolderIDs).Delete(&model.Folder{}).Error; err != nil {
					return fmt.Errorf("delete rescanned folders: %w", err)
				}
			}
		}

		for _, update := range input.UpdatedFolders {
			values := map[string]any{
				"parent_id":       update.ParentID,
				"name":            update.Name,
				"source_path":     update.SourcePath,
				"fs_dir_mtime_ns": update.FsDirMtimeNs,
				"last_scanned_at": update.LastScannedAt,
				"sync_state":      update.SyncState,
				"sync_error":      update.SyncError,
			}
			if update.TouchUpdatedAt {
				values["updated_at"] = input.Now
			}
			if err := tx.Model(&model.Folder{}).
				Where("id = ?", update.ID).
				Updates(values).Error; err != nil {
				return fmt.Errorf("update rescanned folder %s: %w", update.ID, err)
			}
		}

		for _, folder := range input.AddedFolders {
			if err := tx.Create(folder).Error; err != nil {
				return fmt.Errorf("create rescanned folder %s: %w", folder.ID, err)
			}
		}

		for _, update := range input.UpdatedFiles {
			values := map[string]any{
				"folder_id":        update.FolderID,
				"name":             update.Name,
				"description":      update.Description,
				"extension":        update.Extension,
				"mime_type":        update.MimeType,
				"size":             update.Size,
				"fs_file_mtime_ns": update.FsFileMtimeNs,
				"last_verified_at": update.LastVerifiedAt,
			}
			if update.TouchUpdatedAt {
				values["updated_at"] = input.Now
			}
			if err := tx.Model(&model.File{}).
				Where("id = ?", update.ID).
				Updates(values).Error; err != nil {
				return fmt.Errorf("update rescanned file %s: %w", update.ID, err)
			}
		}

		for _, file := range input.AddedFiles {
			if err := tx.Create(file).Error; err != nil {
				return fmt.Errorf("create rescanned file %s: %w", file.ID, err)
			}
		}

		if err := model.RebuildFolderStatsTx(tx); err != nil {
			return fmt.Errorf("rebuild folder stats after rescan: %w", err)
		}
		if err := model.RebuildDashboardStatsTx(tx); err != nil {
			return fmt.Errorf("rebuild dashboard stats after rescan: %w", err)
		}

		if input.OperatorID == "" {
			return nil
		}

		logID, err := identity.NewID()
		if err != nil {
			return fmt.Errorf("generate rescan operation log id: %w", err)
		}
		return createOperationLogTx(tx, logID, input.OperatorID, "managed_directory_rescanned", "folder", input.RootFolderID, input.Detail, input.OperatorIP, input.Now)
	})
}

func (r *ImportRepository) UpdateFolderSyncState(ctx context.Context, folderID string, state, syncError string, now time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.Folder{}).
		Where("id = ?", folderID).
		Updates(map[string]any{
			"sync_state": state,
			"sync_error": syncError,
			"updated_at": now,
		}).Error
}
