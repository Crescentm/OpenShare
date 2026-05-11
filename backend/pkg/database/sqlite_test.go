package database

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNewSQLiteConfiguresConnectionPool(t *testing.T) {
	db, err := NewSQLite(Options{
		Path:         filepath.Join(t.TempDir(), "test.db"),
		LogLevel:     "silent",
		EnableWAL:    true,
		MaxOpenConns: 6,
		MaxIdleConns: 3,
		Pragmas: []Pragma{
			{Name: "busy_timeout", Value: "5000"},
		},
	})
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	defer sqlDB.Close()

	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 6 {
		t.Fatalf("MaxOpenConnections = %d, want 6", stats.MaxOpenConnections)
	}
}

func TestNewSQLiteUsesConcurrentReadPoolDefaults(t *testing.T) {
	db, err := NewSQLite(Options{
		Path:      filepath.Join(t.TempDir(), "test.db"),
		LogLevel:  "silent",
		EnableWAL: true,
	})
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	defer sqlDB.Close()

	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 8 {
		t.Fatalf("MaxOpenConnections = %d, want 8", stats.MaxOpenConnections)
	}
}

func TestBuildDSNIncludesConnectionPragmas(t *testing.T) {
	dsn, err := buildDSN(Options{
		Path:      filepath.Join(t.TempDir(), "test.db"),
		EnableWAL: true,
		Pragmas: []Pragma{
			{Name: "foreign_keys", Value: "ON"},
			{Name: "busy_timeout", Value: "5000"},
		},
	})
	if err != nil {
		t.Fatalf("buildDSN() error = %v", err)
	}

	for _, expected := range []string{
		"_foreign_keys=ON",
		"_busy_timeout=5000",
		"_journal_mode=WAL",
	} {
		if !strings.Contains(dsn, expected) {
			t.Fatalf("dsn %q does not contain %q", dsn, expected)
		}
	}
}
