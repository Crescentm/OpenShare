package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"openshare/backend/internal/model"
	"openshare/backend/internal/repository"
)

func TestRescanManagedRootsSyncsManagedDirectoriesAndContinuesOnUnavailableRoot(t *testing.T) {
	_, db, storageService := newUploadTestDeps(t)
	importService := NewImportService(repository.NewImportRepository(db), storageService)

	firstRoot := createManagedRootFixture(t, "startup-rescan-a")
	secondRoot := createManagedRootFixture(t, "startup-rescan-b")

	if _, err := importService.ImportLocalDirectory(context.Background(), LocalImportInput{RootPath: firstRoot}); err != nil {
		t.Fatalf("import first managed root failed: %v", err)
	}
	if _, err := importService.ImportLocalDirectory(context.Background(), LocalImportInput{RootPath: secondRoot}); err != nil {
		t.Fatalf("import second managed root failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(firstRoot, "later-added.md"), []byte("startup sync"), 0o644); err != nil {
		t.Fatalf("write later-added file failed: %v", err)
	}
	if err := os.RemoveAll(secondRoot); err != nil {
		t.Fatalf("remove second managed root failed: %v", err)
	}

	results, err := importService.RescanManagedRoots(context.Background(), "")
	if err != nil {
		t.Fatalf("rescan managed roots failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 managed root results, got %d", len(results))
	}

	byPath := make(map[string]ManagedRootRescanOutcome, len(results))
	for _, result := range results {
		byPath[result.RootPath] = result
	}

	firstResult, ok := byPath[firstRoot]
	if !ok {
		t.Fatalf("expected result for first root %q, got %+v", firstRoot, results)
	}
	if firstResult.Error != "" {
		t.Fatalf("expected first root to rescan successfully, got error %q", firstResult.Error)
	}
	if firstResult.AddedFiles != 1 {
		t.Fatalf("expected first root to add 1 file, got %+v", firstResult)
	}

	secondResult, ok := byPath[secondRoot]
	if !ok {
		t.Fatalf("expected result for second root %q, got %+v", secondRoot, results)
	}
	if !strings.Contains(secondResult.Error, "托管目录不可用") {
		t.Fatalf("expected unavailable root error, got %+v", secondResult)
	}

	var addedFile model.File
	if err := db.Where("name = ?", "later-added.md").Take(&addedFile).Error; err != nil {
		t.Fatalf("expected rescanned file persisted, got %v", err)
	}
}

func TestRescanManagedPathUpdatesNestedFileWithoutWalkingCleanRoot(t *testing.T) {
	_, db, storageService := newUploadTestDeps(t)
	importService := NewImportService(repository.NewImportRepository(db), storageService)

	rootPath := createManagedRootFixture(t, "path-dirty-root")
	if _, err := importService.ImportLocalDirectory(context.Background(), LocalImportInput{RootPath: rootPath}); err != nil {
		t.Fatalf("import managed root failed: %v", err)
	}

	var rootFolder model.Folder
	if err := db.Where("source_path = ?", rootPath).Take(&rootFolder).Error; err != nil {
		t.Fatalf("find root folder failed: %v", err)
	}
	var nestedFolder model.Folder
	nestedPath := filepath.Join(rootPath, "nested")
	if err := db.Where("source_path = ?", nestedPath).Take(&nestedFolder).Error; err != nil {
		t.Fatalf("find nested folder failed: %v", err)
	}

	if _, err := importService.RescanManagedDirectory(context.Background(), rootFolder.ID, "", ""); err != nil {
		t.Fatalf("initial rescan failed: %v", err)
	}

	var before model.File
	if err := db.Where("name = ?", "chapter1.txt").Take(&before).Error; err != nil {
		t.Fatalf("find nested file failed: %v", err)
	}

	time.Sleep(20 * time.Millisecond)
	filePath := filepath.Join(nestedPath, "chapter1.txt")
	if err := os.WriteFile(filePath, []byte("chapter two"), 0o644); err != nil {
		t.Fatalf("rewrite nested file failed: %v", err)
	}

	skippedResult, err := importService.RescanManagedDirectory(context.Background(), rootFolder.ID, "", "")
	if err != nil {
		t.Fatalf("root incremental rescan failed: %v", err)
	}
	if skippedResult.UpdatedFiles != 0 {
		t.Fatalf("expected clean root rescan to skip nested file update, got %+v", skippedResult)
	}

	var skipped model.File
	if err := db.Where("id = ?", before.ID).Take(&skipped).Error; err != nil {
		t.Fatalf("reload skipped nested file failed: %v", err)
	}
	if skipped.FsFileMtimeNs != before.FsFileMtimeNs {
		t.Fatalf("expected nested file mtime unchanged without dirty path, got %d want %d", skipped.FsFileMtimeNs, before.FsFileMtimeNs)
	}

	dirtyResult, err := importService.RescanManagedPath(context.Background(), rootFolder.ID, nestedPath, "")
	if err != nil {
		t.Fatalf("dirty path rescan failed: %v", err)
	}
	if dirtyResult.UpdatedFiles != 1 {
		t.Fatalf("expected dirty path rescan to update 1 file, got %+v", dirtyResult)
	}

	var updated model.File
	if err := db.Where("id = ?", before.ID).Take(&updated).Error; err != nil {
		t.Fatalf("reload updated nested file failed: %v", err)
	}
	if updated.FsFileMtimeNs == before.FsFileMtimeNs {
		t.Fatalf("expected nested file mtime updated after dirty rescan")
	}
	if updated.LastVerifiedAt == nil {
		t.Fatalf("expected nested file last_verified_at to be populated")
	}

	var refreshedRoot model.Folder
	if err := db.Where("id = ?", rootFolder.ID).Take(&refreshedRoot).Error; err != nil {
		t.Fatalf("reload root folder failed: %v", err)
	}
	if refreshedRoot.SyncState != string(model.FolderSyncStateClean) {
		t.Fatalf("expected root sync state clean, got %q", refreshedRoot.SyncState)
	}
}

func createManagedRootFixture(t *testing.T, name string) string {
	t.Helper()

	root := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(filepath.Join(root, "nested"), 0o755); err != nil {
		t.Fatalf("create managed root dirs failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "root.pdf"), []byte("root file"), 0o644); err != nil {
		t.Fatalf("write managed root file failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "nested", "chapter1.txt"), []byte("chapter one"), 0o644); err != nil {
		t.Fatalf("write nested managed root file failed: %v", err)
	}

	return root
}
