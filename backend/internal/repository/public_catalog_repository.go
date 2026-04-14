package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"openshare/backend/internal/model"
)

type PublicCatalogRepository struct {
	db *gorm.DB
}

type PublicFolderFileListQuery struct {
	FolderID string
	Offset   int
	Limit    int
	OrderBy  []string
}

type PublicFileFeedQuery struct {
	Limit   int
	OrderBy []string
}

type PublicHotFileFeedQuery struct {
	SinceDay string
	Limit    int
}

type PublicFolderRow struct {
	ID            string
	ParentID      *string
	Name          string
	Description   string
	UpdatedAt     time.Time
	FileCount     int64
	DownloadCount int64
	TotalSize     int64
}

func NewPublicCatalogRepository(db *gorm.DB) *PublicCatalogRepository {
	return &PublicCatalogRepository{db: db}
}

func (r *PublicCatalogRepository) ListPublicFolderFiles(ctx context.Context, query PublicFolderFileListQuery) ([]model.File, int64, error) {
	base := r.db.WithContext(ctx).
		Model(&model.File{}).
		Where("folder_id = ?", query.FolderID)
	base = applyVisibleManagedFileFilter(base, "files.name", "files.folder_id", "")

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count public files: %w", err)
	}

	listQuery := base
	for _, orderBy := range query.OrderBy {
		listQuery = listQuery.Order(orderBy)
	}

	var files []model.File
	if err := listQuery.Offset(query.Offset).Limit(query.Limit).Find(&files).Error; err != nil {
		return nil, 0, fmt.Errorf("list public files: %w", err)
	}

	return files, total, nil
}

func (r *PublicCatalogRepository) ListManagedFileFeed(ctx context.Context, query PublicFileFeedQuery) ([]model.File, error) {
	listQuery := r.db.WithContext(ctx).
		Model(&model.File{}).
		Joins("LEFT JOIN folders ON folders.id = files.folder_id")
	listQuery = applyVisibleManagedFileFilter(listQuery, "files.name", "files.folder_id", "folders.source_path")
	for _, orderBy := range query.OrderBy {
		listQuery = listQuery.Order(orderBy)
	}

	var files []model.File
	if err := listQuery.Limit(query.Limit).Find(&files).Error; err != nil {
		return nil, fmt.Errorf("list managed file feed: %w", err)
	}
	return files, nil
}

func (r *PublicCatalogRepository) ListRecentHotManagedFiles(ctx context.Context, query PublicHotFileFeedQuery) ([]model.File, error) {
	aggregated := r.db.WithContext(ctx).
		Model(&model.FileDailyDownload{}).
		Select("file_id, SUM(downloads) AS hot_downloads").
		Where("day >= ?", query.SinceDay).
		Group("file_id")

	var files []model.File
	if err := r.db.WithContext(ctx).
		Model(&model.File{}).
		Select("files.*").
		Joins("LEFT JOIN folders ON folders.id = files.folder_id").
		Joins("JOIN (?) AS hot ON hot.file_id = files.id", aggregated).
		Scopes(func(db *gorm.DB) *gorm.DB {
			return applyVisibleManagedFileFilter(db, "files.name", "files.folder_id", "folders.source_path")
		}).
		Order("hot.hot_downloads DESC").
		Order("files.created_at DESC").
		Order("files.id DESC").
		Limit(query.Limit).
		Find(&files).Error; err != nil {
		return nil, fmt.Errorf("list recent hot managed files: %w", err)
	}
	return files, nil
}

func (r *PublicCatalogRepository) FolderExists(ctx context.Context, folderID string) (bool, error) {
	query := r.db.WithContext(ctx).
		Model(&model.Folder{}).
		Where("id = ?", folderID)
	query = applyVisibleManagedFolderFilter(query, "folders.name", "folders.source_path")

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check folder existence: %w", err)
	}

	return count > 0, nil
}

func (r *PublicCatalogRepository) ListPublicFolders(ctx context.Context, parentID *string) ([]PublicFolderRow, error) {
	query := r.db.WithContext(ctx).
		Model(&model.Folder{}).
		Select("id, parent_id, name, description, updated_at, file_count, download_count, total_size")
	query = applyVisibleManagedFolderFilter(query, "folders.name", "folders.source_path")

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	var rows []PublicFolderRow
	if err := query.Order("name ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list public folders: %w", err)
	}

	return rows, nil
}

func (r *PublicCatalogRepository) FindPublicFolderByID(ctx context.Context, folderID string) (*model.Folder, error) {
	query := r.db.WithContext(ctx).
		Model(&model.Folder{}).
		Where("id = ?", folderID)
	query = applyVisibleManagedFolderFilter(query, "folders.name", "folders.source_path")

	var folder model.Folder
	err := query.Take(&folder).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("find public folder: %w", err)
	}

	return &folder, nil
}
