package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"openshare/backend/internal/model"
	"openshare/backend/internal/pagination"
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
	repository *PublicCatalogRepository
}

type PublicFolderFileListInput struct {
	FolderID string
	Page     int
	PageSize int
	Sort     string
}

type PublicFolderFileListResult struct {
	Items    []PublicFileItem `json:"items"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
}

type PublicFileFeedResult struct {
	Items []PublicFileItem `json:"items"`
}

type PublicFileItem struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Extension     string    `json:"extension"`
	UploadedAt    time.Time `json:"uploaded_at"`
	DownloadCount int64     `json:"download_count"`
	Size          int64     `json:"size"`
}

type PublicFolderItem struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	UpdatedAt     time.Time `json:"updated_at"`
	FileCount     int64     `json:"file_count"`
	DownloadCount int64     `json:"download_count"`
	TotalSize     int64     `json:"total_size"`
}

type PublicFolderBreadcrumbItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PublicFolderDetail struct {
	ID            string                       `json:"id"`
	Name          string                       `json:"name"`
	Description   string                       `json:"description"`
	ParentID      *string                      `json:"parent_id"`
	Breadcrumbs   []PublicFolderBreadcrumbItem `json:"breadcrumbs"`
	FileCount     int64                        `json:"file_count"`
	DownloadCount int64                        `json:"download_count"`
	TotalSize     int64                        `json:"total_size"`
	UpdatedAt     time.Time                    `json:"updated_at"`
}

func NewPublicCatalogService(repository *PublicCatalogRepository) *PublicCatalogService {
	return &PublicCatalogService{repository: repository}
}

func (s *PublicCatalogService) ListPublicFolderFiles(ctx context.Context, input PublicFolderFileListInput) (*PublicFolderFileListResult, error) {
	normalized, err := normalizePublicFolderFileListInput(input)
	if err != nil {
		return nil, err
	}

	exists, err := s.repository.FolderExists(ctx, normalized.FolderID)
	if err != nil {
		return nil, fmt.Errorf("validate folder: %w", err)
	}
	if !exists {
		return nil, ErrFolderNotFound
	}

	files, total, err := s.repository.ListPublicFolderFiles(ctx, PublicFolderFileListQuery{
		FolderID: normalized.FolderID,
		Offset:   (normalized.Page - 1) * normalized.PageSize,
		Limit:    normalized.PageSize,
		OrderBy:  normalized.OrderBy,
	})
	if err != nil {
		return nil, fmt.Errorf("list public folder files: %w", err)
	}

	return &PublicFolderFileListResult{
		Items:    mapPublicFileItems(files),
		Page:     normalized.Page,
		PageSize: normalized.PageSize,
		Total:    total,
	}, nil
}

func (s *PublicCatalogService) ListHotFiles(ctx context.Context, limit int) (*PublicFileFeedResult, error) {
	normalizedLimit := limit
	if normalizedLimit <= 0 {
		normalizedLimit = 20
	}
	if normalizedLimit > maxPublicFilePageSize {
		normalizedLimit = maxPublicFilePageSize
	}

	files, err := s.repository.ListRecentHotManagedFiles(ctx, PublicHotFileFeedQuery{
		SinceDay: time.Now().UTC().AddDate(0, 0, -6).Format("2006-01-02"),
		Limit:    normalizedLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("list recent hot managed files: %w", err)
	}

	return &PublicFileFeedResult{
		Items: mapPublicFileItems(files),
	}, nil
}

func (s *PublicCatalogService) ListLatestFiles(ctx context.Context, limit int) (*PublicFileFeedResult, error) {
	return s.listManagedFileFeed(ctx, limit, []string{"files.created_at DESC", "files.id DESC"})
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
			ID:            row.ID,
			Name:          row.Name,
			Description:   row.Description,
			UpdatedAt:     row.UpdatedAt,
			FileCount:     row.FileCount,
			DownloadCount: row.DownloadCount,
			TotalSize:     row.TotalSize,
		})
	}

	return items, nil
}

func (s *PublicCatalogService) GetPublicFolderDetail(ctx context.Context, folderID string) (*PublicFolderDetail, error) {
	trimmed := strings.TrimSpace(folderID)
	if trimmed == "" {
		return nil, ErrFolderNotFound
	}

	ancestors, err := s.repository.ListPublicFolderAncestors(ctx, trimmed)
	if err != nil {
		return nil, fmt.Errorf("list public folder ancestors: %w", err)
	}
	if len(ancestors) == 0 || ancestors[0].ID != trimmed {
		return nil, ErrFolderNotFound
	}

	seen := make(map[string]struct{}, len(ancestors))
	for index, folder := range ancestors {
		folderID := strings.TrimSpace(folder.ID)
		if folderID == "" {
			return nil, ErrFolderNotFound
		}
		if _, ok := seen[folderID]; ok {
			return nil, ErrFolderNotFound
		}
		seen[folderID] = struct{}{}

		parentID := strings.TrimSpace(optionalFolderID(folder.ParentID))
		if parentID == "" {
			break
		}
		if index+1 >= len(ancestors) || ancestors[index+1].ID != parentID {
			return nil, ErrFolderNotFound
		}
	}

	breadcrumbs := make([]PublicFolderBreadcrumbItem, 0, len(ancestors))
	for _, folder := range ancestors {
		breadcrumbs = append(breadcrumbs, PublicFolderBreadcrumbItem{
			ID:   folder.ID,
			Name: folder.Name,
		})
	}
	for i, j := 0, len(breadcrumbs)-1; i < j; i, j = i+1, j-1 {
		breadcrumbs[i], breadcrumbs[j] = breadcrumbs[j], breadcrumbs[i]
	}

	current := ancestors[0]
	return &PublicFolderDetail{
		ID:            current.ID,
		Name:          current.Name,
		Description:   current.Description,
		ParentID:      current.ParentID,
		Breadcrumbs:   breadcrumbs,
		FileCount:     current.FileCount,
		DownloadCount: current.DownloadCount,
		TotalSize:     current.TotalSize,
		UpdatedAt:     current.UpdatedAt,
	}, nil
}

func optionalFolderID(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

type normalizedPublicFolderFileListInput struct {
	FolderID string
	Page     int
	PageSize int
	OrderBy  []string
}

func normalizePublicFolderFileListInput(input PublicFolderFileListInput) (*normalizedPublicFolderFileListInput, error) {
	folderID := strings.TrimSpace(input.FolderID)
	if folderID == "" {
		return nil, ErrInvalidPublicFileQuery
	}

	page, ok := pagination.Normalize(input.Page, input.PageSize, defaultPublicFilePage, defaultPublicFilePageSize, maxPublicFilePageSize)
	if !ok {
		return nil, ErrInvalidPublicFileQuery
	}

	orderBy, err := resolvePublicFileSort(input.Sort)
	if err != nil {
		return nil, err
	}

	return &normalizedPublicFolderFileListInput{
		FolderID: folderID,
		Page:     page.Page,
		PageSize: page.PageSize,
		OrderBy:  orderBy,
	}, nil
}

func resolvePublicFileSort(sort string) ([]string, error) {
	switch strings.TrimSpace(sort) {
	case "", "created_at_desc":
		return []string{"created_at DESC", "id DESC"}, nil
	case "download_count_desc":
		return []string{"download_count DESC", "created_at DESC", "id DESC"}, nil
	case "name_asc":
		return []string{"name ASC", "created_at DESC", "id DESC"}, nil
	default:
		return nil, ErrInvalidPublicFileQuery
	}
}

func (s *PublicCatalogService) listManagedFileFeed(ctx context.Context, limit int, orderBy []string) (*PublicFileFeedResult, error) {
	normalizedLimit := limit
	if normalizedLimit <= 0 {
		normalizedLimit = 20
	}
	if normalizedLimit > maxPublicFilePageSize {
		normalizedLimit = maxPublicFilePageSize
	}

	files, err := s.repository.ListManagedFileFeed(ctx, PublicFileFeedQuery{
		Limit:   normalizedLimit,
		OrderBy: orderBy,
	})
	if err != nil {
		return nil, fmt.Errorf("list managed file feed: %w", err)
	}

	return &PublicFileFeedResult{
		Items: mapPublicFileItems(files),
	}, nil
}

func mapPublicFileItems(files []model.File) []PublicFileItem {
	items := make([]PublicFileItem, 0, len(files))
	for _, file := range files {
		items = append(items, PublicFileItem{
			ID:            file.ID,
			Name:          file.Name,
			Description:   file.Description,
			Extension:     file.Extension,
			UploadedAt:    file.CreatedAt,
			DownloadCount: file.DownloadCount,
			Size:          file.Size,
		})
	}
	return items
}
