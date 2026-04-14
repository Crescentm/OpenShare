package bootstrap

import (
	"path/filepath"
	"testing"
	"time"

	"openshare/backend/pkg/database"
)

type legacyFolder struct {
	ID            string    `gorm:"column:id;type:text;primaryKey"`
	ParentID      *string   `gorm:"column:parent_id;type:text"`
	SourcePath    *string   `gorm:"column:source_path;type:text"`
	Name          string    `gorm:"column:name;type:text;not null"`
	Description   string    `gorm:"column:description;type:text;not null;default:''"`
	FileCount     int64     `gorm:"column:file_count;type:integer;not null;default:0"`
	TotalSize     int64     `gorm:"column:total_size;type:integer;not null;default:0"`
	DownloadCount int64     `gorm:"column:download_count;type:integer;not null;default:0"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (legacyFolder) TableName() string { return "folders" }

type legacyFile struct {
	ID            string    `gorm:"column:id;type:text;primaryKey"`
	FolderID      *string   `gorm:"column:folder_id;type:text"`
	Name          string    `gorm:"column:name;type:text;not null;default:''"`
	Description   string    `gorm:"column:description;type:text;not null;default:''"`
	Extension     string    `gorm:"column:extension;type:text;not null;default:''"`
	MimeType      string    `gorm:"column:mime_type;type:text;not null;default:''"`
	Size          int64     `gorm:"column:size;type:integer;not null;default:0"`
	DownloadCount int64     `gorm:"column:download_count;type:integer;not null;default:0"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (legacyFile) TableName() string { return "files" }

func TestEnsureSchemaAddsManagedSyncColumns(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "managed-sync.db")
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

	if err := db.AutoMigrate(&legacyFolder{}, &legacyFile{}); err != nil {
		t.Fatalf("auto migrate legacy schema failed: %v", err)
	}

	now := time.Now().UTC()
	sourcePath := "/srv/share"
	folderID := "folder-1"
	if err := db.Create(&legacyFolder{
		ID:         folderID,
		SourcePath: &sourcePath,
		Name:       "资料",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("insert legacy folder failed: %v", err)
	}
	if err := db.Create(&legacyFile{
		ID:        "file-1",
		FolderID:  &folderID,
		Name:      "notes.txt",
		Extension: ".txt",
		MimeType:  "text/plain",
		Size:      12,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("insert legacy file failed: %v", err)
	}

	if err := EnsureSchema(db); err != nil {
		t.Fatalf("ensure schema failed: %v", err)
	}

	folderColumns, err := tableColumns(db, "folders")
	if err != nil {
		t.Fatalf("inspect folder columns failed: %v", err)
	}
	for _, column := range []string{"fs_dir_mtime_ns", "last_scanned_at", "sync_state", "sync_error"} {
		if !hasTableColumn(folderColumns, column) {
			t.Fatalf("expected folder column %q after migration, got %v", column, folderColumns)
		}
	}

	fileColumns, err := tableColumns(db, "files")
	if err != nil {
		t.Fatalf("inspect file columns failed: %v", err)
	}
	for _, column := range []string{"fs_file_mtime_ns", "last_verified_at"} {
		if !hasTableColumn(fileColumns, column) {
			t.Fatalf("expected file column %q after migration, got %v", column, fileColumns)
		}
	}

	var folderRow struct {
		Name      string
		SyncState string
	}
	if err := db.Raw(`SELECT name, sync_state FROM folders WHERE id = 'folder-1'`).Scan(&folderRow).Error; err != nil {
		t.Fatalf("query migrated folder failed: %v", err)
	}
	if folderRow.Name != "资料" || folderRow.SyncState != "pending" {
		t.Fatalf("unexpected migrated folder row: %+v", folderRow)
	}

	var fileRow struct {
		Name string
		Size int64
	}
	if err := db.Raw(`SELECT name, size FROM files WHERE id = 'file-1'`).Scan(&fileRow).Error; err != nil {
		t.Fatalf("query migrated file failed: %v", err)
	}
	if fileRow.Name != "notes.txt" || fileRow.Size != 12 {
		t.Fatalf("unexpected migrated file row: %+v", fileRow)
	}
}
