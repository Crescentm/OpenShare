package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/meilisearch/meilisearch-go"

	"openshare/backend/internal/config"
	"openshare/backend/internal/pagination"
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
	searcherMu      sync.Mutex
	searcher        meilisearchSearcher
}

func NewSearchService(searchRepo *SearchRepository, cfg config.SearchEngineConfig) *SearchService {
	return &SearchService{
		searchRepo:      searchRepo,
		searchEngineCfg: cfg,
		newSearcher:     newMeilisearchSearcher,
	}
}

type SearchInput struct {
	Keyword       string
	FolderID      string
	Type          string
	FileKind      string
	Extension     string
	Category      string
	Course        string
	MaterialType  string
	ContentStatus string
	Page          int
	PageSize      int
}

type SearchResult struct {
	Items    []SearchResultItem `json:"items"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int64              `json:"total"`
}

type SearchResultItem struct {
	EntityType    string            `json:"entity_type"`
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Path          string            `json:"path,omitempty"`
	PathSegments  []string          `json:"path_segments,omitempty"`
	Extension     string            `json:"extension,omitempty"`
	FileKind      string            `json:"file_kind,omitempty"`
	Category      string            `json:"category,omitempty"`
	Course        string            `json:"course,omitempty"`
	MaterialType  string            `json:"material_type,omitempty"`
	ContentStatus string            `json:"content_status,omitempty"`
	Size          int64             `json:"size,omitempty"`
	DownloadCount int64             `json:"download_count,omitempty"`
	UploadedAt    *time.Time        `json:"uploaded_at,omitempty"`
	UpdatedAt     *time.Time        `json:"updated_at,omitempty"`
	Snippet       string            `json:"snippet,omitempty"`
	Highlights    map[string]string `json:"highlights,omitempty"`
}

func (s *SearchService) Search(ctx context.Context, input SearchInput) (*SearchResult, error) {
	page, ok := pagination.NormalizeWindow(input.Page, input.PageSize, defaultSearchPage, defaultSearchPageSize, maxSearchPageSize, defaultSearchResultWindow())
	if !ok {
		return nil, ErrSearchInvalidInput
	}

	query, err := normalizeSearchKeyword(input.Keyword)
	if err != nil {
		return nil, err
	}

	if !s.searchEngineCfg.Enabled {
		return nil, ErrSearchIndexDisabled
	}

	request := &meilisearch.SearchRequest{
		Page:             int64(page.Page),
		HitsPerPage:      int64(page.PageSize),
		MatchingStrategy: meilisearch.All,
		AttributesToRetrieve: []string{
			"type",
			"resource_id",
			"name",
			"path",
			"path_segments",
			"extension",
			"file_kind",
			"category",
			"course",
			"material_type",
			"content_status",
			"size",
			"download_count",
			"created_at",
			"updated_at",
		},
		AttributesToHighlight: []string{
			"name",
			"path",
			"course",
			"material_type",
			"description",
			"readme",
			"content_text",
		},
		AttributesToCrop: []string{
			"description",
			"readme",
			"content_text",
		},
		CropLength:       30,
		CropMarker:       "...",
		HighlightPreTag:  "[[",
		HighlightPostTag: "]]",
	}

	scopeFolderID := strings.TrimSpace(input.FolderID)
	filters, err := searchInputFilters(input)
	if err != nil {
		return nil, err
	}
	if scopeFolderID != "" {
		if s.searchRepo == nil {
			return nil, ErrSearchInvalidInput
		}
		scopeFolderIDs, err := s.searchRepo.GetDescendantFolderIDs(ctx, scopeFolderID)
		if err != nil {
			return nil, fmt.Errorf("resolve folder scope: %w", err)
		}
		filters = append([]string{searchFolderScopeFilter(scopeFolderIDs)}, filters...)
	}
	if filter := joinMeilisearchFilters(filters); filter != "" {
		request.Filter = filter
	}

	searcher, err := s.meilisearchSearcher()
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
		Page:     page.Page,
		PageSize: page.PageSize,
		Total:    searchResponseTotal(response),
	}, nil
}

func (s *SearchService) meilisearchSearcher() (meilisearchSearcher, error) {
	s.searcherMu.Lock()
	defer s.searcherMu.Unlock()

	if s.searcher != nil {
		return s.searcher, nil
	}

	searcher, err := s.newSearcher(s.searchEngineCfg)
	if err != nil {
		return nil, err
	}
	s.searcher = searcher
	return searcher, nil
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
		item := searchDocumentToResultItem(doc)
		item.Highlights = formattedSearchHitFields(hit)
		item.Snippet = searchResultSnippet(item.Highlights)
		items = append(items, item)
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
		updatedAt := searchDocumentTime(doc.UpdatedAt)
		return SearchResultItem{
			EntityType:    SearchDocumentTypeFile,
			ID:            doc.ResourceID,
			Name:          doc.Name,
			Path:          doc.Path,
			PathSegments:  doc.PathSegments,
			Extension:     doc.Extension,
			FileKind:      doc.FileKind,
			Category:      doc.Category,
			Course:        doc.Course,
			MaterialType:  doc.MaterialType,
			ContentStatus: doc.ContentStatus,
			Size:          doc.Size,
			DownloadCount: doc.DownloadCount,
			UploadedAt:    uploadedAt,
			UpdatedAt:     updatedAt,
		}
	default:
		updatedAt := searchDocumentTime(doc.UpdatedAt)
		return SearchResultItem{
			EntityType:    SearchDocumentTypeFolder,
			ID:            doc.ResourceID,
			Name:          doc.Name,
			Path:          doc.Path,
			PathSegments:  doc.PathSegments,
			Category:      doc.Category,
			Course:        doc.Course,
			MaterialType:  doc.MaterialType,
			ContentStatus: doc.ContentStatus,
			Size:          doc.Size,
			DownloadCount: doc.DownloadCount,
			UpdatedAt:     updatedAt,
		}
	}
}

func searchDocumentTime(unix int64) *time.Time {
	if unix <= 0 {
		return nil
	}
	value := time.Unix(unix, 0).UTC()
	return &value
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

func searchInputFilters(input SearchInput) ([]string, error) {
	var filters []string

	if filter, err := exactSearchFilter("type", normalizeSearchType(input.Type)); err != nil {
		return nil, err
	} else if filter != "" {
		filters = append(filters, filter)
	}
	if filter, err := exactSearchFilter("file_kind", normalizeSearchKeywordFilterValue(input.FileKind)); err != nil {
		return nil, err
	} else if filter != "" {
		filters = append(filters, filter)
	}
	if filter, err := exactSearchFilter("extension", normalizeSearchDocumentExtension(input.Extension)); err != nil {
		return nil, err
	} else if filter != "" {
		filters = append(filters, filter)
	}
	if filter, err := exactSearchFilter("category", normalizeSearchTextFilterValue(input.Category)); err != nil {
		return nil, err
	} else if filter != "" {
		filters = append(filters, filter)
	}
	if filter, err := exactSearchFilter("course", normalizeSearchTextFilterValue(input.Course)); err != nil {
		return nil, err
	} else if filter != "" {
		filters = append(filters, filter)
	}
	if filter, err := exactSearchFilter("material_type", normalizeSearchTextFilterValue(input.MaterialType)); err != nil {
		return nil, err
	} else if filter != "" {
		filters = append(filters, filter)
	}
	if filter, err := exactSearchFilter("content_status", normalizeSearchContentStatusFilter(input.ContentStatus)); err != nil {
		return nil, err
	} else if filter != "" {
		filters = append(filters, filter)
	}

	return filters, nil
}

func exactSearchFilter(field string, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if value == "\x00" {
		return "", ErrSearchInvalidInput
	}
	if len([]rune(value)) > 80 {
		return "", ErrSearchInvalidInput
	}
	return field + ` = "` + escapeMeilisearchFilterString(value) + `"`, nil
}

func normalizeSearchType(value string) string {
	value = normalizeSearchKeywordFilterValue(value)
	if value == "" {
		return ""
	}
	if value != SearchDocumentTypeFile && value != SearchDocumentTypeFolder {
		return "\x00"
	}
	return value
}

func normalizeSearchContentStatusFilter(value string) string {
	value = normalizeSearchKeywordFilterValue(value)
	if value == "" {
		return ""
	}
	switch value {
	case SearchContentStatusNone, SearchContentStatusPending, SearchContentStatusReady, SearchContentStatusFailed:
		return value
	default:
		return "\x00"
	}
}

func normalizeSearchKeywordFilterValue(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func normalizeSearchTextFilterValue(value string) string {
	return strings.TrimSpace(value)
}

func joinMeilisearchFilters(filters []string) string {
	parts := make([]string, 0, len(filters))
	for _, filter := range filters {
		filter = strings.TrimSpace(filter)
		if filter == "" {
			continue
		}
		parts = append(parts, "("+filter+")")
	}
	return strings.Join(parts, " AND ")
}

func escapeMeilisearchFilterString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}

func formattedSearchHitFields(hit meilisearch.Hit) map[string]string {
	raw, ok := hit["_formatted"]
	if !ok || len(raw) == 0 {
		return nil
	}

	var formatted map[string]json.RawMessage
	if err := json.Unmarshal(raw, &formatted); err != nil {
		return nil
	}

	result := make(map[string]string, len(formatted))
	for key, rawValue := range formatted {
		value := formattedSearchHitValue(rawValue)
		if value != "" {
			result[key] = value
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func formattedSearchHitValue(raw json.RawMessage) string {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}

	var list []string
	if err := json.Unmarshal(raw, &list); err == nil {
		return strings.TrimSpace(strings.Join(list, " / "))
	}

	return ""
}

func searchResultSnippet(highlights map[string]string) string {
	for _, field := range []string{"content_text", "description", "readme", "path", "name"} {
		if value := strings.TrimSpace(highlights[field]); value != "" {
			return value
		}
	}
	return ""
}

func defaultSearchResultWindow() int {
	return 100
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
