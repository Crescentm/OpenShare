package router

import (
	"path/filepath"
	"testing"

	"openshare/backend/internal/config"
	"openshare/backend/internal/storage"
)

func newRouterTestConfig(t *testing.T) config.Config {
	t.Helper()

	cfg := config.Default()
	cfg.Session.Secret = "test-secret"
	cfg.Storage.Root = filepath.Join(t.TempDir(), "storage")
	cfg.Database.Path = filepath.Join(t.TempDir(), "openshare-test.db")

	if err := storage.EnsureLayout(cfg.Storage); err != nil {
		t.Fatalf("ensure storage layout failed: %v", err)
	}

	return cfg
}
