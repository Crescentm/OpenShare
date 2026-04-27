package search

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"openshare/backend/internal/model"
)

type SearchRepository struct {
	db *gorm.DB
}

func NewSearchRepository(db *gorm.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

func (r *SearchRepository) GetDescendantFolderIDs(ctx context.Context, folderID string) ([]string, error) {
	result := []string{folderID}
	queue := []string{folderID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		var childIDs []string
		if err := r.db.WithContext(ctx).
			Model(&model.Folder{}).
			Where("parent_id = ?", current).
			Pluck("id", &childIDs).Error; err != nil {
			return nil, fmt.Errorf("get child folders: %w", err)
		}

		result = append(result, childIDs...)
		queue = append(queue, childIDs...)
	}
	return result, nil
}
