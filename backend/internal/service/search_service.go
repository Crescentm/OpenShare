package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"openshare/backend/internal/config"
	"openshare/backend/internal/model"
	"openshare/backend/internal/repository"
	"openshare/backend/internal/search"
)

// ---------------------------------------------------------------------------
// Sentinel errors
// ---------------------------------------------------------------------------

var (
	ErrSearchQueryEmpty   = errors.New("search query is empty")
	ErrSearchQueryTooLong = errors.New("search query exceeds maximum length")
	ErrSearchInvalidInput = errors.New("invalid search parameters")
)

const (
	defaultSearchPage     = 1
	defaultSearchPageSize = 20
	maxSearchPageSize     = 100
	maxSearchQueryLength  = 200
)

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// SearchService implements the public search use-case:
//   - keyword search (FTS5 with LIKE fallback)
//   - folder-scoped search
//   - relevance + download_count ordering
type SearchService struct {
	searchRepo *repository.SearchRepository
	settings   *SystemSettingService
}

func NewSearchService(searchRepo *repository.SearchRepository, settings *SystemSettingService) *SearchService {
	return &SearchService{
		searchRepo: searchRepo,
		settings:   settings,
	}
}

// ---------------------------------------------------------------------------
// Input / Output
// ---------------------------------------------------------------------------

// SearchInput is the external request from the handler layer.
type SearchInput struct {
	Keyword  string // raw user input
	FolderID string // optional folder scope
	Page     int
	PageSize int
}

// SearchResult is the response delivered to the handler layer.
type SearchResult struct {
	Items    []SearchResultItem `json:"items"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int64              `json:"total"`
}

// SearchResultItem represents a single file or folder in the search results.
type SearchResultItem struct {
	EntityType    string     `json:"entity_type"` // "file" | "folder"
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	OriginalName  string     `json:"original_name,omitempty"`
	Extension     string     `json:"extension,omitempty"`
	Size          int64      `json:"size,omitempty"`
	DownloadCount int64      `json:"download_count,omitempty"`
	UploadedAt    *time.Time `json:"uploaded_at,omitempty"`
}

// ---------------------------------------------------------------------------
// Core search
// ---------------------------------------------------------------------------

func (s *SearchService) Search(ctx context.Context, input SearchInput) (*SearchResult, error) {
	policy := s.searchPolicy(ctx)

	// --- 1. Validate & normalise -----------------------------------------
	page, pageSize, err := normalizeSearchPagination(input.Page, input.PageSize, policy.ResultWindow)
	if err != nil {
		return nil, err
	}

	keyword := strings.TrimSpace(input.Keyword)
	if len([]rune(keyword)) > maxSearchQueryLength {
		return nil, ErrSearchQueryTooLong
	}

	// Sanitize keyword for FTS5
	fts5Query, hasFTS := search.SanitizeQuery(keyword)

	if !hasFTS {
		return nil, ErrSearchQueryEmpty
	}

	// --- 2. Resolve folder scope -----------------------------------------
	var scopeFolderIDs []string
	if trimmed := strings.TrimSpace(input.FolderID); trimmed != "" {
		if !policy.EnableFolderScope {
			return nil, ErrSearchInvalidInput
		}
		ids, err := s.searchRepo.GetDescendantFolderIDs(ctx, trimmed)
		if err != nil {
			return nil, fmt.Errorf("resolve folder scope: %w", err)
		}
		scopeFolderIDs = ids
	}

	offset := (page - 1) * pageSize

	// --- 3. FTS5 primary search ------------------------------------------
	rows, total, err := s.searchRepo.Search(ctx, repository.SearchQuery{
		FTS5Query:      fts5Query,
		ScopeFolderIDs: scopeFolderIDs,
		Offset:         offset,
		Limit:          pageSize,
	})
	if err != nil {
		ftsErr := err
		// FTS5 may be unavailable or broken on some SQLite builds.
		// Fall back to LIKE search instead of failing the whole request.
		rows, total, err = s.searchRepo.SearchWithLike(ctx, strings.ToLower(keyword), scopeFolderIDs, offset, pageSize)
		if err != nil {
			return nil, fmt.Errorf("fts5 search: %w; like fallback: %w", ftsErr, err)
		}
	}

	// --- 4. LIKE fallback if FTS5 returned nothing -----------------------
	if total == 0 && hasFTS && policy.EnableFuzzyMatch {
		rows, total, err = s.searchRepo.SearchWithLike(ctx, strings.ToLower(keyword), scopeFolderIDs, offset, pageSize)
		if err != nil {
			return nil, fmt.Errorf("like search fallback: %w", err)
		}
	}

	if total == 0 {
		return &SearchResult{
			Items:    []SearchResultItem{},
			Page:     page,
			PageSize: pageSize,
			Total:    0,
		}, nil
	}

	// --- 5. Hydrate results with metadata --------------------------------
	items, err := s.hydrateResults(ctx, rows)
	if err != nil {
		return nil, fmt.Errorf("hydrate results: %w", err)
	}

	return &SearchResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *SearchService) searchPolicy(ctx context.Context) SearchPolicy {
	if s.settings == nil {
		return defaultSystemPolicy(config.UploadConfig{}).Search
	}

	policy, err := s.settings.GetPolicy(ctx)
	if err != nil || policy == nil {
		return defaultSystemPolicy(config.UploadConfig{}).Search
	}
	return policy.Search
}

// ---------------------------------------------------------------------------
// Result hydration
// ---------------------------------------------------------------------------

func (s *SearchService) hydrateResults(ctx context.Context, rows []repository.SearchResultRow) ([]SearchResultItem, error) {
	var fileIDs, folderIDs []string
	for _, row := range rows {
		switch row.EntityType {
		case "file":
			fileIDs = append(fileIDs, row.EntityID)
		case "folder":
			folderIDs = append(folderIDs, row.EntityID)
		}
	}

	// Load metadata
	fileMap := make(map[string]*fileHydrated)
	if len(fileIDs) > 0 {
		files, err := s.searchRepo.GetFilesByIDs(ctx, fileIDs)
		if err != nil {
			return nil, err
		}
		for i := range files {
			f := files[i]
			fileMap[f.ID] = &fileHydrated{file: &f}
		}
	}

	folderMap := make(map[string]*folderHydrated)
	if len(folderIDs) > 0 {
		folders, err := s.searchRepo.GetFoldersByIDs(ctx, folderIDs)
		if err != nil {
			return nil, err
		}
		for i := range folders {
			f := folders[i]
			folderMap[f.ID] = &folderHydrated{folder: &f}
		}
	}

	// Assemble in original order
	items := make([]SearchResultItem, 0, len(rows))
	for _, row := range rows {
		switch row.EntityType {
		case "file":
			h, ok := fileMap[row.EntityID]
			if !ok {
				continue
			}
			t := h.file.CreatedAt
			items = append(items, SearchResultItem{
				EntityType:    "file",
				ID:            h.file.ID,
				Name:          h.file.Title,
				OriginalName:  h.file.OriginalName,
				Extension:     h.file.Extension,
				Size:          h.file.Size,
				DownloadCount: h.file.DownloadCount,
				UploadedAt:    &t,
			})
		case "folder":
			h, ok := folderMap[row.EntityID]
			if !ok {
				continue
			}
			items = append(items, SearchResultItem{
				EntityType: "folder",
				ID:         h.folder.ID,
				Name:       h.folder.Name,
			})
		}
	}

	return items, nil
}

type fileHydrated struct {
	file *model.File
}

type folderHydrated struct {
	folder *model.Folder
}

// ---------------------------------------------------------------------------
// Index sync public API (called by other services after mutations)
// ---------------------------------------------------------------------------

// IndexFile updates the FTS5 index for a single file.
func (s *SearchService) IndexFile(ctx context.Context, fileID, title, description string) error {
	return s.searchRepo.UpsertFileIndex(ctx, fileID, title, description)
}

// IndexFolder updates the FTS5 index for a single folder.
func (s *SearchService) IndexFolder(ctx context.Context, folderID, name, description string) error {
	return s.searchRepo.UpsertFolderIndex(ctx, folderID, name, description)
}

// RemoveFromIndex removes an entity from the FTS5 index.
func (s *SearchService) RemoveFromIndex(ctx context.Context, entityType, entityID string) error {
	return s.searchRepo.RemoveIndex(ctx, entityType, entityID)
}

// RebuildAllIndexes rebuilds the full FTS5 index from scratch.
func (s *SearchService) RebuildAllIndexes(ctx context.Context) error {
	return s.searchRepo.RebuildAllIndexes(ctx)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func normalizeSearchPagination(page, pageSize, resultWindow int) (int, int, error) {
	if page == 0 {
		page = defaultSearchPage
	}
	if page < 1 {
		return 0, 0, ErrSearchInvalidInput
	}
	if pageSize == 0 {
		pageSize = defaultSearchPageSize
	}
	if pageSize < 1 || pageSize > maxSearchPageSize {
		return 0, 0, ErrSearchInvalidInput
	}
	if resultWindow > 0 && page*pageSize > resultWindow {
		return 0, 0, ErrSearchInvalidInput
	}
	return page, pageSize, nil
}
