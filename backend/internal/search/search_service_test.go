package search

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"gorm.io/gorm"

	"openshare/backend/internal/bootstrap"
	"openshare/backend/internal/config"
	"openshare/backend/internal/model"
	"openshare/backend/pkg/database"
)

func TestSearchUsesMeilisearchAndShapesResults(t *testing.T) {
	fake := &fakeSearchEngine{
		response: &meilisearch.SearchResponse{
			Hits: meilisearch.Hits{
				searchHit(t, SearchDocument{
					Type:          SearchDocumentTypeFile,
					ResourceID:    "file-1",
					Name:          "2022年数据结构期末试卷.pdf",
					Extension:     "pdf",
					Size:          1024,
					DownloadCount: 7,
					CreatedAt:     time.Unix(1700000000, 0).Unix(),
				}),
				searchHit(t, SearchDocument{
					Type:       SearchDocumentTypeFolder,
					ResourceID: "folder-1",
					Name:       "数据结构",
				}),
			},
			TotalHits: 2,
		},
	}
	service := newFakeSearchService(nil, fake)

	result, err := service.Search(context.Background(), SearchInput{
		Keyword:  "数据结构 试卷",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if fake.query != "数据结构 试卷" {
		t.Fatalf("query = %q, want 数据结构 试卷", fake.query)
	}
	if fake.request.MatchingStrategy != meilisearch.All {
		t.Fatalf("MatchingStrategy = %q, want all", fake.request.MatchingStrategy)
	}
	if result.Total != 2 {
		t.Fatalf("Total = %d, want 2", result.Total)
	}
	if len(result.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(result.Items))
	}
	if result.Items[0].EntityType != "file" || result.Items[0].ID != "file-1" {
		t.Fatalf("first item = %+v, want file file-1", result.Items[0])
	}
	if result.Items[0].UploadedAt == nil || result.Items[0].UploadedAt.Unix() != 1700000000 {
		t.Fatalf("UploadedAt = %v, want unix 1700000000", result.Items[0].UploadedAt)
	}
	if result.Items[1].EntityType != "folder" || result.Items[1].ID != "folder-1" {
		t.Fatalf("second item = %+v, want folder folder-1", result.Items[1])
	}
}

func TestSearchBuildsMeilisearchFolderScopeFilter(t *testing.T) {
	db := newSearchTestSQLite(t)
	rootID := "folder-root"
	childID := "folder-child"
	mustCreateSearchFolder(t, db, model.Folder{ID: rootID, Name: "课程资料"})
	mustCreateSearchFolder(t, db, model.Folder{ID: childID, ParentID: ptrString(rootID), Name: "试卷"})

	fake := &fakeSearchEngine{response: &meilisearch.SearchResponse{}}
	service := newFakeSearchService(NewSearchRepository(db), fake)

	_, err := service.Search(context.Background(), SearchInput{
		Keyword:  "试卷",
		FolderID: rootID,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	want := `(type = "file" AND folder_id IN ["folder-root", "folder-child"]) OR (type = "folder" AND resource_id IN ["folder-root", "folder-child"])`
	if fake.request.Filter != want {
		t.Fatalf("Filter = %#v, want %s", fake.request.Filter, want)
	}
}

func TestSearchRejectsDisabledSearchEngine(t *testing.T) {
	service := NewSearchService(nil, config.SearchEngineConfig{})

	_, err := service.Search(context.Background(), SearchInput{Keyword: "数据结构"})
	if !errors.Is(err, ErrSearchIndexDisabled) {
		t.Fatalf("Search() error = %v, want ErrSearchIndexDisabled", err)
	}
}

func TestSearchValidatesInput(t *testing.T) {
	service := newFakeSearchService(nil, &fakeSearchEngine{response: &meilisearch.SearchResponse{}})

	if _, err := service.Search(context.Background(), SearchInput{Keyword: ""}); !errors.Is(err, ErrSearchQueryEmpty) {
		t.Fatalf("empty query error = %v, want ErrSearchQueryEmpty", err)
	}
	if _, err := service.Search(context.Background(), SearchInput{Keyword: "ok", Page: -1}); !errors.Is(err, ErrSearchInvalidInput) {
		t.Fatalf("invalid page error = %v, want ErrSearchInvalidInput", err)
	}
	if _, err := service.Search(context.Background(), SearchInput{Keyword: "ok", Page: 6, PageSize: 20}); !errors.Is(err, ErrSearchInvalidInput) {
		t.Fatalf("window error = %v, want ErrSearchInvalidInput", err)
	}
}

type fakeSearchEngine struct {
	query    string
	request  *meilisearch.SearchRequest
	response *meilisearch.SearchResponse
	err      error
}

func (f *fakeSearchEngine) Search(_ context.Context, query string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
	f.query = query
	f.request = request
	if f.err != nil {
		return nil, f.err
	}
	if f.response == nil {
		return &meilisearch.SearchResponse{}, nil
	}
	return f.response, nil
}

func newFakeSearchService(repo *SearchRepository, fake *fakeSearchEngine) *SearchService {
	service := NewSearchService(repo, config.SearchEngineConfig{
		Enabled:   true,
		Host:      "http://127.0.0.1:7700",
		APIKey:    "test-key",
		IndexName: "test_resources",
	})
	service.newSearcher = func(config.SearchEngineConfig) (meilisearchSearcher, error) {
		return fake, nil
	}
	return service
}

func searchHit(t *testing.T, doc SearchDocument) meilisearch.Hit {
	t.Helper()
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal hit failed: %v", err)
	}
	var hit meilisearch.Hit
	if err := json.Unmarshal(data, &hit); err != nil {
		t.Fatalf("unmarshal hit failed: %v", err)
	}
	return hit
}

func mustCreateSearchFolder(t *testing.T, db *gorm.DB, folder model.Folder) {
	t.Helper()
	if err := db.Create(&folder).Error; err != nil {
		t.Fatalf("create folder %q failed: %v", folder.ID, err)
	}
}

func ptrString(value string) *string {
	return &value
}

func newSearchTestSQLite(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "openshare-search-test.db")
	db, err := database.NewSQLite(database.Options{
		Path:      dbPath,
		LogLevel:  "silent",
		EnableWAL: true,
		Pragmas: []database.Pragma{
			{Name: "foreign_keys", Value: "ON"},
			{Name: "busy_timeout", Value: "5000"},
		},
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := bootstrap.EnsureSchema(db); err != nil {
		t.Fatalf("ensure schema failed: %v", err)
	}

	return db
}
