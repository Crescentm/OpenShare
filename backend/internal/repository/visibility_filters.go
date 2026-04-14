package repository

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

const hiddenManagedNamePattern = ".%"
const hiddenManagedPathPattern = "%/.%"

func applyVisibleManagedFolderFilter(db *gorm.DB, folderNameColumn string, folderSourcePathColumn string) *gorm.DB {
	if db == nil {
		return nil
	}

	db = db.Where(fmt.Sprintf("%s NOT LIKE ?", folderNameColumn), hiddenManagedNamePattern)
	if strings.TrimSpace(folderSourcePathColumn) == "" {
		return db
	}

	return db.Where(
		fmt.Sprintf("(COALESCE(%s, '') = '' OR COALESCE(%s, '') NOT LIKE ?)", folderSourcePathColumn, folderSourcePathColumn),
		hiddenManagedPathPattern,
	)
}

func applyVisibleManagedFileFilter(db *gorm.DB, fileNameColumn string, folderIDColumn string, folderSourcePathColumn string) *gorm.DB {
	if db == nil {
		return nil
	}

	db = db.Where(fmt.Sprintf("%s NOT LIKE ?", fileNameColumn), hiddenManagedNamePattern)
	if strings.TrimSpace(folderSourcePathColumn) == "" {
		return db
	}

	return db.Where(
		fmt.Sprintf("(%s IS NULL OR COALESCE(%s, '') NOT LIKE ?)", folderIDColumn, folderSourcePathColumn),
		hiddenManagedPathPattern,
	)
}
