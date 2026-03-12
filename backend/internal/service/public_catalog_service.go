package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"openshare/backend/internal/repository"
)

var (
	ErrInvalidPublicFileQuery = errors.New("invalid public file query")
	ErrFolderNotFound         = errors.New("folder not found")
)

const (
	defaultPublicFilePage     = 1
	defaultPublicFilePageSize = 20
	maxPublicFilePageSize     = 100
)

type PublicCatalogService struct {
	repository *repository.PublicCatalogRepository
}

type PublicFileListInput struct {
	FolderID       string
	FilterByFolder bool // true when the caller explicitly wants to browse within a folder
	Page           int
	PageSize       int
	Sort           string
}

type PublicFileListResult struct {
	Items    []PublicFileItem `json:"items"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Total    int64           `json:"total"`
}

type PublicFileItem struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Tags          []string  `json:"tags"`
	UploadedAt    time.Time `json:"uploaded_at"`
	DownloadCount int64     `json:"download_count"`
	Size          int64     `json:"size"`
}

type PublicFolderItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewPublicCatalogService(repository *repository.PublicCatalogRepository) *PublicCatalogService {
	return &PublicCatalogService{repository: repository}
}

func (s *PublicCatalogService) ListPublicFiles(ctx context.Context, input PublicFileListInput) (*PublicFileListResult, error) {
	normalized, err := normalizePublicFileListInput(input)
	if err != nil {
		return nil, err
	}

	if normalized.FolderID != nil {
		exists, err := s.repository.FolderExists(ctx, *normalized.FolderID)
		if err != nil {
			return nil, fmt.Errorf("validate folder: %w", err)
		}
		if !exists {
			return nil, ErrFolderNotFound
		}
	}

	files, total, err := s.repository.ListPublicFiles(ctx, repository.PublicFileListQuery{
		FolderID:       normalized.FolderID,
		FilterByFolder: normalized.FilterByFolder,
		Offset:         (normalized.Page - 1) * normalized.PageSize,
		Limit:          normalized.PageSize,
		OrderBy:        normalized.OrderBy,
	})
	if err != nil {
		return nil, fmt.Errorf("list public files: %w", err)
	}

	fileIDs := make([]string, 0, len(files))
	for _, file := range files {
		fileIDs = append(fileIDs, file.ID)
	}

	tagRows, err := s.repository.ListTagsByFileIDs(ctx, fileIDs)
	if err != nil {
		return nil, fmt.Errorf("list public file tags: %w", err)
	}

	tagsByFileID := make(map[string][]string, len(files))
	for _, row := range tagRows {
		tagsByFileID[row.FileID] = append(tagsByFileID[row.FileID], row.TagName)
	}

	items := make([]PublicFileItem, 0, len(files))
	for _, file := range files {
		tags := tagsByFileID[file.ID]
		if tags == nil {
			tags = []string{}
		}
		items = append(items, PublicFileItem{
			ID:            file.ID,
			Title:         file.Title,
			Tags:          tags,
			UploadedAt:    file.CreatedAt,
			DownloadCount: file.DownloadCount,
			Size:          file.Size,
		})
	}

	return &PublicFileListResult{
		Items:    items,
		Page:     normalized.Page,
		PageSize: normalized.PageSize,
		Total:    total,
	}, nil
}

func (s *PublicCatalogService) ListPublicFolders(ctx context.Context, parentID string) ([]PublicFolderItem, error) {
	var parentPtr *string
	if trimmed := strings.TrimSpace(parentID); trimmed != "" {
		exists, err := s.repository.FolderExists(ctx, trimmed)
		if err != nil {
			return nil, fmt.Errorf("validate parent folder: %w", err)
		}
		if !exists {
			return nil, ErrFolderNotFound
		}
		parentPtr = &trimmed
	}

	rows, err := s.repository.ListPublicFolders(ctx, parentPtr)
	if err != nil {
		return nil, fmt.Errorf("list public folders: %w", err)
	}

	items := make([]PublicFolderItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, PublicFolderItem{
			ID:   row.ID,
			Name: row.Name,
		})
	}

	return items, nil
}

type normalizedPublicFileListInput struct {
	FolderID       *string
	FilterByFolder bool
	Page           int
	PageSize       int
	OrderBy        []string
}

func normalizePublicFileListInput(input PublicFileListInput) (*normalizedPublicFileListInput, error) {
	page := input.Page
	if page == 0 {
		page = defaultPublicFilePage
	}
	if page < 1 {
		return nil, ErrInvalidPublicFileQuery
	}

	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = defaultPublicFilePageSize
	}
	if pageSize < 1 || pageSize > maxPublicFilePageSize {
		return nil, ErrInvalidPublicFileQuery
	}

	orderBy, err := resolvePublicFileSort(input.Sort)
	if err != nil {
		return nil, err
	}

	var folderID *string
	filterByFolder := input.FilterByFolder
	if trimmed := strings.TrimSpace(input.FolderID); trimmed != "" {
		folderID = &trimmed
		filterByFolder = true
	}

	return &normalizedPublicFileListInput{
		FolderID:       folderID,
		FilterByFolder: filterByFolder,
		Page:           page,
		PageSize:       pageSize,
		OrderBy:        orderBy,
	}, nil
}

func resolvePublicFileSort(sort string) ([]string, error) {
	switch strings.TrimSpace(sort) {
	case "", "created_at_desc":
		return []string{"created_at DESC", "id DESC"}, nil
	case "download_count_desc":
		return []string{"download_count DESC", "created_at DESC", "id DESC"}, nil
	case "title_asc":
		return []string{"title ASC", "created_at DESC", "id DESC"}, nil
	default:
		return nil, ErrInvalidPublicFileQuery
	}
}
