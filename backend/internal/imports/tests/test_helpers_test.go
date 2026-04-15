package imports

import (
	"os"
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"openshare/backend/internal/bootstrap"
	"openshare/backend/internal/config"
	"openshare/backend/internal/storage"
	"openshare/backend/pkg/database"
)

func newUploadTestDeps(t *testing.T) (config.Config, *gorm.DB, *storage.Service) {
	t.Helper()

	cfg := config.Default()
	cfg.Session.Secret = "test-secret"
	cfg.Storage.Root = filepath.Join(t.TempDir(), "storage")
	cfg.Database.Path = filepath.Join(t.TempDir(), "openshare-upload.db")

	if err := storage.EnsureLayout(cfg.Storage); err != nil {
		t.Fatalf("ensure storage layout failed: %v", err)
	}

	db, err := database.NewSQLite(database.Options{
		Path:      cfg.Database.Path,
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

	return cfg, db, storage.NewService(cfg.Storage)
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
