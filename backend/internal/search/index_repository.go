package search

import (
	"context"
	"fmt"

	"openshare/backend/internal/model"
	"openshare/backend/internal/resources"
)

func (r *SearchRepository) ListIndexFolders(ctx context.Context) ([]model.Folder, error) {
	var folders []model.Folder
	db := r.db.WithContext(ctx).
		Model(&model.Folder{})
	db = resources.ApplyVisibleManagedFolderFilter(db, "folders.name", "folders.source_path")

	if err := db.Find(&folders).Error; err != nil {
		return nil, fmt.Errorf("list folders for search index: %w", err)
	}
	return folders, nil
}

func (r *SearchRepository) ListIndexFiles(ctx context.Context) ([]model.File, error) {
	var files []model.File
	db := r.db.WithContext(ctx).
		Model(&model.File{}).
		Select("files.*").
		Joins("LEFT JOIN folders ON folders.id = files.folder_id")
	db = resources.ApplyVisibleManagedFileFilter(db, "files.name", "files.folder_id", "folders.source_path")

	if err := db.Find(&files).Error; err != nil {
		return nil, fmt.Errorf("list files for search index: %w", err)
	}
	return files, nil
}
