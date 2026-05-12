package search

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/meilisearch/meilisearch-go"

	"openshare/backend/internal/config"
	"openshare/backend/internal/model"
	"openshare/backend/internal/searchengine"
)

const (
	searchIndexBatchSize    = 500
	searchIndexTaskInterval = 200 * time.Millisecond
	searchIndexRebuildLimit = 10 * time.Minute
)

var ErrSearchIndexDisabled = errors.New("search index is disabled")

type SearchIndexService struct {
	mu              sync.Mutex
	clientMu        sync.Mutex
	cfg             config.SearchEngineConfig
	searchRepo      *SearchRepository
	profileProvider SemanticProfileProvider
	rebuildOnce     sync.Once
	rebuildRequests chan string
	client          *searchengine.MeilisearchClient
}

type SemanticProfileProvider interface {
	GetSearchProfile(ctx context.Context) (*SemanticProfile, error)
}

type SearchIndexStatus struct {
	Enabled             bool   `json:"enabled"`
	IndexName           string `json:"index_name"`
	SemanticProfileMode string `json:"semantic_profile_mode"`
	Status              string `json:"status"`
	Error               string `json:"error,omitempty"`
}

type SearchIndexRebuildResult struct {
	IndexName        string `json:"index_name"`
	IndexedDocuments int    `json:"indexed_documents"`
	IndexedFiles     int    `json:"indexed_files"`
	IndexedFolders   int    `json:"indexed_folders"`
	SkippedFiles     int    `json:"skipped_files"`
	SkippedFolders   int    `json:"skipped_folders"`
	LastTaskUID      int64  `json:"last_task_uid,omitempty"`
}

type meilisearchIndexer struct {
	client  *searchengine.MeilisearchClient
	repo    *SearchRepository
	builder *SearchDocumentBuilder
}

type searchIndexDocumentSet struct {
	Documents      []SearchDocument
	IndexedFiles   int
	IndexedFolders int
	SkippedFiles   int
	SkippedFolders int
}

type searchIndexFolderSnapshot struct {
	Folder model.Folder
	Path   string
	RootID string
}

func NewSearchIndexService(searchRepo *SearchRepository, cfg config.SearchEngineConfig, profileProvider SemanticProfileProvider) *SearchIndexService {
	return &SearchIndexService{
		cfg:             cfg,
		searchRepo:      searchRepo,
		profileProvider: profileProvider,
		rebuildRequests: make(chan string, 1),
	}
}

func (s *SearchIndexService) NotifySearchResourcesChanged(reason string) {
	if !s.cfg.Enabled {
		return
	}

	s.rebuildOnce.Do(func() {
		go s.rebuildLoop()
	})

	select {
	case s.rebuildRequests <- strings.TrimSpace(reason):
	default:
	}
}

func (s *SearchIndexService) rebuildLoop() {
	for reason := range s.rebuildRequests {
		s.rebuildAfterResourceChange(reason)

		pendingReason := ""
		for {
			select {
			case nextReason := <-s.rebuildRequests:
				pendingReason = nextReason
			default:
				if pendingReason != "" {
					select {
					case s.rebuildRequests <- pendingReason:
					default:
					}
				}
				goto next
			}
		}
	next:
	}
}

func (s *SearchIndexService) rebuildAfterResourceChange(reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), searchIndexRebuildLimit)
	defer cancel()

	result, err := s.Rebuild(ctx)
	if err != nil {
		log.Printf("[search-index] automatic rebuild failed; reason=%s error=%v", reason, err)
		return
	}

	log.Printf(
		"[search-index] automatic rebuild completed; reason=%s documents=%d files=%d folders=%d skipped_files=%d skipped_folders=%d",
		reason,
		result.IndexedDocuments,
		result.IndexedFiles,
		result.IndexedFolders,
		result.SkippedFiles,
		result.SkippedFolders,
	)
}

func (s *SearchIndexService) Status(ctx context.Context) SearchIndexStatus {
	status := SearchIndexStatus{
		Enabled:             s.cfg.Enabled,
		IndexName:           s.cfg.IndexName,
		SemanticProfileMode: "admin",
		Status:              "disabled",
	}
	if s.profileProvider == nil {
		status.SemanticProfileMode = "generic"
	}
	if !s.cfg.Enabled {
		return status
	}

	client, err := s.meilisearchClient()
	if err != nil {
		status.Status = "unavailable"
		status.Error = err.Error()
		return status
	}
	healthCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := client.HealthCheck(healthCtx); err != nil {
		status.Status = "unavailable"
		status.Error = err.Error()
		return status
	}

	status.Status = "available"
	return status
}

func (s *SearchIndexService) Rebuild(ctx context.Context) (*SearchIndexRebuildResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.cfg.Enabled {
		return nil, ErrSearchIndexDisabled
	}

	client, err := s.meilisearchClient()
	if err != nil {
		return nil, err
	}
	if err := client.HealthCheck(ctx); err != nil {
		return nil, err
	}

	builder, err := s.newDocumentBuilder(ctx)
	if err != nil {
		return nil, err
	}

	indexer := &meilisearchIndexer{
		client:  client,
		repo:    s.searchRepo,
		builder: builder,
	}
	return indexer.Rebuild(ctx)
}

func (s *SearchIndexService) meilisearchClient() (*searchengine.MeilisearchClient, error) {
	s.clientMu.Lock()
	defer s.clientMu.Unlock()

	if s.client != nil {
		return s.client, nil
	}

	client, err := searchengine.NewMeilisearchClient(s.cfg)
	if err != nil {
		return nil, err
	}
	s.client = client
	return client, nil
}

func (s *SearchIndexService) newDocumentBuilder(ctx context.Context) (*SearchDocumentBuilder, error) {
	if s.profileProvider == nil {
		return NewSearchDocumentBuilder(nil), nil
	}

	profile, err := s.profileProvider.GetSearchProfile(ctx)
	if err != nil {
		return nil, fmt.Errorf("load search semantic profile: %w", err)
	}
	return NewProfileSearchDocumentBuilder(profile), nil
}

func (i *meilisearchIndexer) Rebuild(ctx context.Context) (*SearchIndexRebuildResult, error) {
	if err := i.ensureIndex(ctx); err != nil {
		return nil, err
	}

	documents, err := i.buildDocuments(ctx)
	if err != nil {
		return nil, err
	}

	lastTaskUID, err := i.replaceDocuments(ctx, documents.Documents)
	if err != nil {
		return nil, err
	}

	return &SearchIndexRebuildResult{
		IndexName:        i.client.IndexName(),
		IndexedDocuments: len(documents.Documents),
		IndexedFiles:     documents.IndexedFiles,
		IndexedFolders:   documents.IndexedFolders,
		SkippedFiles:     documents.SkippedFiles,
		SkippedFolders:   documents.SkippedFolders,
		LastTaskUID:      lastTaskUID,
	}, nil
}

func (i *meilisearchIndexer) ensureIndex(ctx context.Context) error {
	service := i.client.Service()
	indexName := i.client.IndexName()

	if _, err := service.GetIndexWithContext(ctx, indexName); err != nil {
		if !isMeilisearchNotFound(err) {
			return fmt.Errorf("load search index: %w", err)
		}

		task, err := service.CreateIndexWithContext(ctx, &meilisearch.IndexConfig{
			Uid:        indexName,
			PrimaryKey: SearchDocumentPrimaryKey,
		})
		if err != nil {
			return fmt.Errorf("create search index: %w", err)
		}
		if err := waitMeilisearchTask(ctx, service, task); err != nil {
			return fmt.Errorf("wait search index creation: %w", err)
		}
	}

	index := service.Index(indexName)
	task, err := index.UpdateSearchableAttributesWithContext(ctx, &SearchDocumentSearchableAttributes)
	if err := waitMeilisearchTaskInfo(ctx, service, task, err); err != nil {
		return fmt.Errorf("update searchable attributes: %w", err)
	}

	filterable := make([]interface{}, 0, len(SearchDocumentFilterableAttributes))
	for _, attr := range SearchDocumentFilterableAttributes {
		filterable = append(filterable, attr)
	}
	task, err = index.UpdateFilterableAttributesWithContext(ctx, &filterable)
	if err := waitMeilisearchTaskInfo(ctx, service, task, err); err != nil {
		return fmt.Errorf("update filterable attributes: %w", err)
	}
	task, err = index.UpdateSortableAttributesWithContext(ctx, &SearchDocumentSortableAttributes)
	if err := waitMeilisearchTaskInfo(ctx, service, task, err); err != nil {
		return fmt.Errorf("update sortable attributes: %w", err)
	}
	task, err = index.UpdateRankingRulesWithContext(ctx, &SearchDocumentRankingRules)
	if err := waitMeilisearchTaskInfo(ctx, service, task, err); err != nil {
		return fmt.Errorf("update ranking rules: %w", err)
	}

	return nil
}

func (i *meilisearchIndexer) buildDocuments(ctx context.Context) (*searchIndexDocumentSet, error) {
	folders, err := i.repo.ListIndexFolders(ctx)
	if err != nil {
		return nil, err
	}
	files, err := i.repo.ListIndexFiles(ctx)
	if err != nil {
		return nil, err
	}

	folderSnapshots := buildSearchIndexFolderSnapshots(folders)
	documents := searchIndexDocumentSet{
		Documents: make([]SearchDocument, 0, len(folders)+len(files)),
	}

	for _, folder := range folders {
		snapshot, ok := folderSnapshots[folder.ID]
		folderPath := folder.Name
		rootID := folder.ID
		if ok {
			folderPath = snapshot.Path
			rootID = snapshot.RootID
		}

		doc := i.builder.BuildFolder(FolderSearchDocumentInput{
			Folder:       folder,
			RootFolderID: rootID,
			FolderPath:   folderPath,
		})
		if doc == nil {
			documents.SkippedFolders++
			continue
		}
		documents.IndexedFolders++
		documents.Documents = append(documents.Documents, *doc)
	}

	for _, file := range files {
		folderPath := ""
		rootID := ""
		if file.FolderID != nil {
			if snapshot, ok := folderSnapshots[*file.FolderID]; ok {
				folderPath = snapshot.Path
				rootID = snapshot.RootID
			}
		}

		doc := i.builder.BuildFile(FileSearchDocumentInput{
			File:         file,
			RootFolderID: rootID,
			FolderPath:   folderPath,
		})
		if doc == nil {
			documents.SkippedFiles++
			continue
		}
		documents.IndexedFiles++
		documents.Documents = append(documents.Documents, *doc)
	}

	return &documents, nil
}

func (i *meilisearchIndexer) replaceDocuments(ctx context.Context, documents []SearchDocument) (int64, error) {
	service := i.client.Service()
	index := service.Index(i.client.IndexName())

	deleteTask, err := index.DeleteAllDocumentsWithContext(ctx, &meilisearch.DocumentOptions{})
	if err != nil {
		return 0, fmt.Errorf("clear search index documents: %w", err)
	}
	if err := waitMeilisearchTask(ctx, service, deleteTask); err != nil {
		return deleteTask.TaskUID, fmt.Errorf("wait search index clear: %w", err)
	}

	lastTaskUID := deleteTask.TaskUID
	if len(documents) == 0 {
		return lastTaskUID, nil
	}

	primaryKey := SearchDocumentPrimaryKey
	tasks, err := index.AddDocumentsInBatchesWithContext(ctx, documents, searchIndexBatchSize, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return lastTaskUID, fmt.Errorf("add search documents: %w", err)
	}
	for _, task := range tasks {
		task := task
		lastTaskUID = task.TaskUID
		if err := waitMeilisearchTask(ctx, service, &task); err != nil {
			return lastTaskUID, fmt.Errorf("wait search document batch: %w", err)
		}
	}

	return lastTaskUID, nil
}

func buildSearchIndexFolderSnapshots(folders []model.Folder) map[string]searchIndexFolderSnapshot {
	byID := make(map[string]model.Folder, len(folders))
	for _, folder := range folders {
		byID[folder.ID] = folder
	}

	snapshots := make(map[string]searchIndexFolderSnapshot, len(folders))
	var resolve func(folderID string, seen map[string]struct{}) searchIndexFolderSnapshot
	resolve = func(folderID string, seen map[string]struct{}) searchIndexFolderSnapshot {
		if snapshot, ok := snapshots[folderID]; ok {
			return snapshot
		}

		folder, ok := byID[folderID]
		if !ok {
			return searchIndexFolderSnapshot{}
		}
		if _, cycle := seen[folderID]; cycle {
			return searchIndexFolderSnapshot{
				Folder: folder,
				Path:   folder.Name,
				RootID: folder.ID,
			}
		}
		seen[folderID] = struct{}{}

		snapshot := searchIndexFolderSnapshot{
			Folder: folder,
			Path:   folder.Name,
			RootID: folder.ID,
		}
		if folder.ParentID != nil {
			parent := resolve(*folder.ParentID, seen)
			if parent.RootID != "" {
				snapshot.RootID = parent.RootID
			}
			if parent.Path != "" {
				snapshot.Path = parent.Path + "/" + folder.Name
			}
		}

		snapshots[folderID] = snapshot
		return snapshot
	}

	for _, folder := range folders {
		resolve(folder.ID, make(map[string]struct{}))
	}
	return snapshots
}

func waitMeilisearchTask(ctx context.Context, service meilisearch.ServiceManager, task *meilisearch.TaskInfo) error {
	if task == nil {
		return nil
	}

	completed, err := service.WaitForTaskWithContext(ctx, task.TaskUID, searchIndexTaskInterval)
	if err != nil {
		return err
	}
	if completed.Status != meilisearch.TaskStatusSucceeded {
		return fmt.Errorf("task %d finished with status %s: %v", completed.TaskUID, completed.Status, completed.Error)
	}
	return nil
}

func waitMeilisearchTaskInfo(ctx context.Context, service meilisearch.ServiceManager, task *meilisearch.TaskInfo, err error) error {
	if err != nil {
		return err
	}
	return waitMeilisearchTask(ctx, service, task)
}

func isMeilisearchNotFound(err error) bool {
	var meiliErr *meilisearch.Error
	return errors.As(err, &meiliErr) && meiliErr.StatusCode == http.StatusNotFound
}
