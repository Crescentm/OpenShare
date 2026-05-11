package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPreservesUploadDefaultsForPartialOverrides(t *testing.T) {
	defaultPath, localPath := writeTestConfigFiles(
		t,
		Default(),
		`{"upload":{"max_upload_total_bytes":123456789}}`,
	)

	cfg, err := Load(defaultPath, localPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Upload.MaxUploadTotalBytes != 123456789 {
		t.Fatalf("MaxUploadTotalBytes = %d, want %d", cfg.Upload.MaxUploadTotalBytes, int64(123456789))
	}
	if cfg.Upload.MaxDescriptionLength != Default().Upload.MaxDescriptionLength {
		t.Fatalf("MaxDescriptionLength = %d, want %d", cfg.Upload.MaxDescriptionLength, Default().Upload.MaxDescriptionLength)
	}
	if cfg.Upload.ReceiptCodeLength != Default().Upload.ReceiptCodeLength {
		t.Fatalf("ReceiptCodeLength = %d, want %d", cfg.Upload.ReceiptCodeLength, Default().Upload.ReceiptCodeLength)
	}
}

func TestLoadRejectsInvalidUploadOverride(t *testing.T) {
	defaultPath, localPath := writeTestConfigFiles(
		t,
		Default(),
		`{"upload":{"max_upload_total_bytes":0}}`,
	)

	_, err := Load(defaultPath, localPath)
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "upload.max_upload_total_bytes must be greater than 0") {
		t.Fatalf("Load() error = %v, want upload.max_upload_total_bytes validation error", err)
	}
}

func TestLoadPreservesManagedSyncDefaultsForPartialOverrides(t *testing.T) {
	defaultPath, localPath := writeTestConfigFiles(
		t,
		Default(),
		`{"managed_sync":{"refresh_interval_seconds":15}}`,
	)

	cfg, err := Load(defaultPath, localPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ManagedSync.RefreshIntervalSeconds != 15 {
		t.Fatalf("RefreshIntervalSeconds = %d, want %d", cfg.ManagedSync.RefreshIntervalSeconds, 15)
	}
	if cfg.ManagedSync.AuditIntervalSeconds != Default().ManagedSync.AuditIntervalSeconds {
		t.Fatalf(
			"AuditIntervalSeconds = %d, want %d",
			cfg.ManagedSync.AuditIntervalSeconds,
			Default().ManagedSync.AuditIntervalSeconds,
		)
	}
}

func TestLoadPreservesDatabasePoolDefaultsForPartialOverrides(t *testing.T) {
	defaultPath, localPath := writeTestConfigFiles(
		t,
		Default(),
		`{"database":{"path":"/tmp/openshare-test.db"}}`,
	)

	cfg, err := Load(defaultPath, localPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Database.MaxOpenConns != Default().Database.MaxOpenConns {
		t.Fatalf("MaxOpenConns = %d, want %d", cfg.Database.MaxOpenConns, Default().Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns != Default().Database.MaxIdleConns {
		t.Fatalf("MaxIdleConns = %d, want %d", cfg.Database.MaxIdleConns, Default().Database.MaxIdleConns)
	}
}

func TestLoadRejectsInvalidDatabasePoolConfig(t *testing.T) {
	defaultPath, localPath := writeTestConfigFiles(
		t,
		Default(),
		`{"database":{"max_open_conns":4,"max_idle_conns":5}}`,
	)

	_, err := Load(defaultPath, localPath)
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "database.max_idle_conns must be less than or equal to database.max_open_conns") {
		t.Fatalf("Load() error = %v, want database.max_idle_conns validation error", err)
	}
}

func TestLoadAppliesDatabasePoolEnvOverrides(t *testing.T) {
	t.Setenv("OPENSHARE_DATABASE_MAX_OPEN_CONNS", "12")
	t.Setenv("OPENSHARE_DATABASE_MAX_IDLE_CONNS", "6")

	defaultPath, localPath := writeTestConfigFiles(t, Default(), `{}`)

	cfg, err := Load(defaultPath, localPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Database.MaxOpenConns != 12 {
		t.Fatalf("MaxOpenConns = %d, want 12", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns != 6 {
		t.Fatalf("MaxIdleConns = %d, want 6", cfg.Database.MaxIdleConns)
	}
}

func TestLoadRejectsInvalidManagedSyncOverride(t *testing.T) {
	defaultPath, localPath := writeTestConfigFiles(
		t,
		Default(),
		`{"managed_sync":{"audit_interval_seconds":0}}`,
	)

	_, err := Load(defaultPath, localPath)
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "managed_sync.audit_interval_seconds must be greater than 0") {
		t.Fatalf("Load() error = %v, want managed_sync.audit_interval_seconds validation error", err)
	}
}

func TestLoadPreservesSearchEngineDefaultsForPartialOverrides(t *testing.T) {
	defaultPath, localPath := writeTestConfigFiles(
		t,
		Default(),
		`{"search_engine":{"enabled":true,"api_key":"test-meili-key"}}`,
	)

	cfg, err := Load(defaultPath, localPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.SearchEngine.Enabled {
		t.Fatal("SearchEngine.Enabled = false, want true")
	}
	if cfg.SearchEngine.APIKey != "test-meili-key" {
		t.Fatalf("SearchEngine.APIKey = %q, want test-meili-key", cfg.SearchEngine.APIKey)
	}
	if cfg.SearchEngine.Host != Default().SearchEngine.Host {
		t.Fatalf("SearchEngine.Host = %q, want %q", cfg.SearchEngine.Host, Default().SearchEngine.Host)
	}
	if cfg.SearchEngine.IndexName != Default().SearchEngine.IndexName {
		t.Fatalf("SearchEngine.IndexName = %q, want %q", cfg.SearchEngine.IndexName, Default().SearchEngine.IndexName)
	}
	if cfg.SearchEngine.SemanticProfilePath != Default().SearchEngine.SemanticProfilePath {
		t.Fatalf(
			"SearchEngine.SemanticProfilePath = %q, want %q",
			cfg.SearchEngine.SemanticProfilePath,
			Default().SearchEngine.SemanticProfilePath,
		)
	}
}

func TestLoadRejectsEnabledSearchEngineWithoutAPIKey(t *testing.T) {
	defaultPath, localPath := writeTestConfigFiles(
		t,
		Default(),
		`{"search_engine":{"enabled":true}}`,
	)

	_, err := Load(defaultPath, localPath)
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "search_engine.api_key must not be empty") {
		t.Fatalf("Load() error = %v, want search_engine.api_key validation error", err)
	}
}

func TestLoadAppliesSearchEngineEnvOverrides(t *testing.T) {
	t.Setenv("OPENSHARE_SEARCH_ENGINE_ENABLED", "true")
	t.Setenv("OPENSHARE_SEARCH_ENGINE_HOST", "http://meilisearch:7700")
	t.Setenv("OPENSHARE_SEARCH_ENGINE_API_KEY", "env-meili-key")
	t.Setenv("OPENSHARE_SEARCH_ENGINE_INDEX_NAME", "env_resources")
	t.Setenv("OPENSHARE_SEARCH_ENGINE_SEMANTIC_PROFILE_PATH", "config/search_semantics.custom.json")

	defaultPath, localPath := writeTestConfigFiles(t, Default(), `{}`)

	cfg, err := Load(defaultPath, localPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.SearchEngine.Enabled {
		t.Fatal("SearchEngine.Enabled = false, want true")
	}
	if cfg.SearchEngine.Host != "http://meilisearch:7700" {
		t.Fatalf("SearchEngine.Host = %q, want http://meilisearch:7700", cfg.SearchEngine.Host)
	}
	if cfg.SearchEngine.APIKey != "env-meili-key" {
		t.Fatalf("SearchEngine.APIKey = %q, want env-meili-key", cfg.SearchEngine.APIKey)
	}
	if cfg.SearchEngine.IndexName != "env_resources" {
		t.Fatalf("SearchEngine.IndexName = %q, want env_resources", cfg.SearchEngine.IndexName)
	}
	if cfg.SearchEngine.SemanticProfilePath != "config/search_semantics.custom.json" {
		t.Fatalf(
			"SearchEngine.SemanticProfilePath = %q, want config/search_semantics.custom.json",
			cfg.SearchEngine.SemanticProfilePath,
		)
	}
}

func writeTestConfigFiles(t *testing.T, cfg Config, localJSON string) (string, string) {
	t.Helper()

	dir := t.TempDir()
	cfg.Session.Secret = "test-session-secret"

	defaultPath := filepath.Join(dir, "config.default.json")
	localPath := filepath.Join(dir, "config.local.json")

	defaultData, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal default config: %v", err)
	}
	if err := os.WriteFile(defaultPath, defaultData, 0o600); err != nil {
		t.Fatalf("write default config: %v", err)
	}
	if err := os.WriteFile(localPath, []byte(localJSON), 0o600); err != nil {
		t.Fatalf("write local config: %v", err)
	}

	return defaultPath, localPath
}
