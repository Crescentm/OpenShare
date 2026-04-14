package bootstrap

import "fmt"

import "gorm.io/gorm"

func migrateManagedSyncSchema(db *gorm.DB) error {
	if !db.Migrator().HasTable("folders") || !db.Migrator().HasTable("files") {
		return nil
	}

	if !db.Migrator().HasColumn("folders", "fs_dir_mtime_ns") {
		if err := db.Exec(`ALTER TABLE folders ADD COLUMN fs_dir_mtime_ns INTEGER NOT NULL DEFAULT 0`).Error; err != nil {
			return fmt.Errorf("add folders.fs_dir_mtime_ns: %w", err)
		}
	}
	if !db.Migrator().HasColumn("folders", "last_scanned_at") {
		if err := db.Exec(`ALTER TABLE folders ADD COLUMN last_scanned_at DATETIME`).Error; err != nil {
			return fmt.Errorf("add folders.last_scanned_at: %w", err)
		}
	}
	if !db.Migrator().HasColumn("folders", "sync_state") {
		if err := db.Exec(`ALTER TABLE folders ADD COLUMN sync_state TEXT NOT NULL DEFAULT 'pending'`).Error; err != nil {
			return fmt.Errorf("add folders.sync_state: %w", err)
		}
	}
	if !db.Migrator().HasColumn("folders", "sync_error") {
		if err := db.Exec(`ALTER TABLE folders ADD COLUMN sync_error TEXT NOT NULL DEFAULT ''`).Error; err != nil {
			return fmt.Errorf("add folders.sync_error: %w", err)
		}
	}

	if !db.Migrator().HasColumn("files", "fs_file_mtime_ns") {
		if err := db.Exec(`ALTER TABLE files ADD COLUMN fs_file_mtime_ns INTEGER NOT NULL DEFAULT 0`).Error; err != nil {
			return fmt.Errorf("add files.fs_file_mtime_ns: %w", err)
		}
	}
	if !db.Migrator().HasColumn("files", "last_verified_at") {
		if err := db.Exec(`ALTER TABLE files ADD COLUMN last_verified_at DATETIME`).Error; err != nil {
			return fmt.Errorf("add files.last_verified_at: %w", err)
		}
	}

	if !db.Migrator().HasIndex("folders", "idx_folders_sync_state") {
		if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_folders_sync_state ON folders(sync_state)`).Error; err != nil {
			return fmt.Errorf("create idx_folders_sync_state: %w", err)
		}
	}

	if err := db.Exec(`UPDATE folders SET sync_state = 'pending' WHERE COALESCE(TRIM(sync_state), '') = ''`).Error; err != nil {
		return fmt.Errorf("backfill folders.sync_state: %w", err)
	}
	if err := db.Exec(`UPDATE folders SET sync_error = '' WHERE sync_error IS NULL`).Error; err != nil {
		return fmt.Errorf("backfill folders.sync_error: %w", err)
	}

	return nil
}
