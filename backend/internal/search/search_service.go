package search

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/meilisearch/meilisearch-go"

	"openshare/backend/internal/config"
	"openshare/backend/internal/searchengine"
)

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

type meilisearchSearcher interface {
	Search(ctx context.Context, query string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error)
}

type meilisearchSearcherFactory func(config.SearchEngineConfig) (meilisearchSearcher, error)

type SearchService struct {
	searchRepo      *SearchRepository
	searchEngineCfg config.SearchEngineConfig
	newSearcher     meilisearchSearcherFactory
}

func NewSearchService(searchRepo *SearchRepository, cfg config.SearchEngineConfig) *SearchService {
	return &SearchService{
		searchRepo:      searchRepo,
		searchEngineCfg: cfg,
		newSearcher:     newMeilisearchSearcher,
	}
}

type SearchInput struct {
	Keyword  string
	FolderID string
	Page     int
	PageSize int
}

type SearchResult struct {
	Items    []SearchResultItem `json:"items"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int64              `json:"total"`
}

type SearchResultItem struct {
	EntityType    string     `json:"entity_type"`
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Extension     string     `json:"extension,omitempty"`
	Size          int64      `json:"size,omitempty"`
	DownloadCount int64      `json:"download_count,omitempty"`
	UploadedAt    *time.Time `json:"uploaded_at,omitempty"`
}

func (s *SearchService) Search(ctx context.Context, input SearchInput) (*SearchResult, error) {
	page, pageSize, err := normalizeSearchPagination(input.Page, input.PageSize, defaultSearchResultWindow())
	if err != nil {
		return nil, err
	}

	query, err := normalizeSearchKeyword(input.Keyword)
	if err != nil {
		return nil, err
	}

	if !s.searchEngineCfg.Enabled {
		return nil, ErrSearchIndexDisabled
	}

	request := &meilisearch.SearchRequest{
		Page:             int64(page),
		HitsPerPage:      int64(pageSize),
		MatchingStrategy: meilisearch.All,
		AttributesToRetrieve: []string{
			"type",
			"resource_id",
			"name",
			"extension",
			"size",
			"download_count",
			"created_at",
		},
	}

	scopeFolderID := strings.TrimSpace(input.FolderID)
	if scopeFolderID != "" {
		scopeFolderIDs, err := s.searchRepo.GetDescendantFolderIDs(ctx, scopeFolderID)
		if err != nil {
			return nil, fmt.Errorf("resolve folder scope: %w", err)
		}
		request.Filter = searchFolderScopeFilter(scopeFolderIDs)
	}

	searcher, err := s.newSearcher(s.searchEngineCfg)
	if err != nil {
		return nil, err
	}

	response, err := searcher.Search(ctx, query, request)
	if err != nil {
		return nil, err
	}

	items, err := searchHitsToResultItems(response.Hits)
	if err != nil {
		return nil, err
	}

	return &SearchResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    searchResponseTotal(response),
	}, nil
}

func newMeilisearchSearcher(cfg config.SearchEngineConfig) (meilisearchSearcher, error) {
	return searchengine.NewMeilisearchClient(cfg)
}

func normalizeSearchKeyword(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if len([]rune(trimmed)) > maxSearchQueryLength {
		return "", ErrSearchQueryTooLong
	}

	query := collapseSearchWhitespace(trimmed)
	if query == "" {
		return "", ErrSearchQueryEmpty
	}
	return query, nil
}

func searchHitsToResultItems(hits meilisearch.Hits) ([]SearchResultItem, error) {
	items := make([]SearchResultItem, 0, len(hits))
	for _, hit := range hits {
		var doc SearchDocument
		if err := hit.DecodeInto(&doc); err != nil {
			return nil, fmt.Errorf("decode search hit: %w", err)
		}
		items = append(items, searchDocumentToResultItem(doc))
	}
	return items, nil
}

func searchDocumentToResultItem(doc SearchDocument) SearchResultItem {
	switch doc.Type {
	case SearchDocumentTypeFile:
		var uploadedAt *time.Time
		if doc.CreatedAt > 0 {
			value := time.Unix(doc.CreatedAt, 0).UTC()
			uploadedAt = &value
		}
		return SearchResultItem{
			EntityType:    SearchDocumentTypeFile,
			ID:            doc.ResourceID,
			Name:          doc.Name,
			Extension:     doc.Extension,
			Size:          doc.Size,
			DownloadCount: doc.DownloadCount,
			UploadedAt:    uploadedAt,
		}
	default:
		return SearchResultItem{
			EntityType: SearchDocumentTypeFolder,
			ID:         doc.ResourceID,
			Name:       doc.Name,
		}
	}
}

func searchResponseTotal(response *meilisearch.SearchResponse) int64 {
	if response.TotalHits > 0 {
		return response.TotalHits
	}
	return response.EstimatedTotalHits
}

func searchFolderScopeFilter(folderIDs []string) string {
	if len(folderIDs) == 0 {
		return ""
	}

	values := make([]string, 0, len(folderIDs))
	for _, folderID := range folderIDs {
		folderID = strings.TrimSpace(folderID)
		if folderID == "" {
			continue
		}
		values = append(values, `"`+escapeMeilisearchFilterString(folderID)+`"`)
	}
	if len(values) == 0 {
		return ""
	}

	list := strings.Join(values, ", ")
	return `(type = "file" AND folder_id IN [` + list + `]) OR (type = "folder" AND resource_id IN [` + list + `])`
}

func escapeMeilisearchFilterString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}

func defaultSearchResultWindow() int {
	return 100
}

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

func collapseSearchWhitespace(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))

	lastWasSpace := true
	for _, r := range value {
		if unicode.IsSpace(r) {
			if !lastWasSpace {
				builder.WriteByte(' ')
			}
			lastWasSpace = true
			continue
		}
		builder.WriteRune(r)
		lastWasSpace = false
	}

	return strings.TrimSpace(builder.String())
}
