package resources

import (
	"context"
	"errors"
	"time"

	"openshare/backend/internal/storage"
)

var (
	ErrManagedFileNotFound   = errors.New("managed file not found")
	ErrManagedFileConflict   = errors.New("managed file conflict")
	ErrManagedFolderNotFound = errors.New("managed folder not found")
	ErrManagedFolderConflict = errors.New("managed folder conflict")
	ErrInvalidResourceEdit   = errors.New("invalid resource edit")
	ErrInvalidResourceQuery  = errors.New("invalid resource query")
)

const (
	defaultManagedFilePage     = 1
	defaultManagedFilePageSize = 20
	maxManagedFilePageSize     = 100
)

type ResourceManagementService struct {
	repo    *ResourceManagementRepository
	storage *storage.Service
	nowFunc func() time.Time
}

type ManagedFileItem struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Extension     string    `json:"extension"`
	Size          int64     `json:"size"`
	DownloadCount int64     `json:"download_count"`
	FolderName    string    `json:"folder_name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ManagedFileListResult struct {
	Items    []ManagedFileItem `json:"items"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
}

type ListManagedFilesInput struct {
	Query    string
	Page     int
	PageSize int
}

type UpdateManagedFileInput struct {
	Name        string
	Description string
	OperatorID  string
	OperatorIP  string
}

type UpdateManagedFolderDescriptionInput struct {
	Name        string
	Description string
	OperatorID  string
	OperatorIP  string
}

func NewResourceManagementService(repo *ResourceManagementRepository, storageService *storage.Service) *ResourceManagementService {
	return &ResourceManagementService{
		repo:    repo,
		storage: storageService,
		nowFunc: func() time.Time { return time.Now().UTC() },
	}
}

func (s *ResourceManagementService) ListFiles(ctx context.Context, input ListManagedFilesInput) (*ManagedFileListResult, error) {
	page, pageSize, err := normalizeManagedFileListPagination(input.Page, input.PageSize)
	if err != nil {
		return nil, err
	}

	rows, total, err := s.repo.ListFiles(ctx, ManagedFileListQuery{
		Query:  input.Query,
		Offset: (page - 1) * pageSize,
		Limit:  pageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]ManagedFileItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, ManagedFileItem{
			ID:            row.ID,
			Name:          row.Name,
			Description:   row.Description,
			Extension:     row.Extension,
			Size:          row.Size,
			DownloadCount: row.DownloadCount,
			FolderName:    row.FolderName,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		})
	}
	return &ManagedFileListResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func normalizeManagedFileListPagination(page int, pageSize int) (int, int, error) {
	if page == 0 {
		page = defaultManagedFilePage
	}
	if pageSize == 0 {
		pageSize = defaultManagedFilePageSize
	}
	if page < 1 || pageSize < 1 || pageSize > maxManagedFilePageSize {
		return 0, 0, ErrInvalidResourceQuery
	}
	return page, pageSize, nil
}
