package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"openshare/backend/internal/model"
	"openshare/backend/internal/repository"
)

func TestImportSyncManagerEnqueueDeduplicatesDescendants(t *testing.T) {
	manager := &ImportSyncManager{
		debounceInterval: time.Second,
		pending:          make(map[string]managedRootSyncRequest),
	}

	manager.enqueue("root-1", "资料", "/srv/share", "/srv/share/a/b", false)
	manager.enqueue("root-1", "资料", "/srv/share", "/srv/share/a", false)
	manager.enqueue("root-1", "资料", "/srv/share", "/srv/share/a/c", true)

	if len(manager.pending) != 1 {
		t.Fatalf("expected 1 pending request after dedupe, got %d", len(manager.pending))
	}

	request, ok := manager.pending["/srv/share/a"]
	if !ok {
		t.Fatalf("expected ancestor path retained, got %+v", manager.pending)
	}
	if !request.ForceFull {
		t.Fatalf("expected forceFull upgrade to be retained")
	}
}

func TestImportSyncManagerFsnotifySyncsNewFile(t *testing.T) {
	_, db, storageService := newUploadTestDeps(t)
	importService := NewImportService(repository.NewImportRepository(db), storageService)

	rootPath := createManagedRootFixture(t, "watch-root")
	if _, err := importService.ImportLocalDirectory(context.Background(), LocalImportInput{RootPath: rootPath}); err != nil {
		t.Fatalf("import managed root failed: %v", err)
	}

	manager := NewImportSyncManager(importService)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("start import sync manager failed: %v", err)
	}

	time.Sleep(150 * time.Millisecond)
	if err := os.WriteFile(filepath.Join(rootPath, "watched-new.txt"), []byte("hello watcher"), 0o644); err != nil {
		t.Fatalf("write watched file failed: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var file model.File
		err := db.Where("name = ?", "watched-new.txt").Take(&file).Error
		if err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("expected fsnotify manager to sync newly created file")
}
