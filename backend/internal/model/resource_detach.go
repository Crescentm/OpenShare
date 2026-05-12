package model

import (
	"fmt"

	"gorm.io/gorm"
)

func DetachDeletedResourcesTx(tx *gorm.DB, fileIDs []string, folderIDs []string) error {
	if len(fileIDs) > 0 {
		if err := tx.Model(&Submission{}).
			Where("file_id IN ?", fileIDs).
			Update("file_id", nil).Error; err != nil {
			return fmt.Errorf("clear submission file links: %w", err)
		}
		if err := tx.Model(&Feedback{}).
			Where("file_id IN ?", fileIDs).
			Update("file_id", nil).Error; err != nil {
			return fmt.Errorf("clear feedback file links: %w", err)
		}
	}

	if len(folderIDs) > 0 {
		if err := tx.Model(&Submission{}).
			Where("folder_id IN ?", folderIDs).
			Update("folder_id", nil).Error; err != nil {
			return fmt.Errorf("clear submission folder links: %w", err)
		}
		if err := tx.Model(&Feedback{}).
			Where("folder_id IN ?", folderIDs).
			Update("folder_id", nil).Error; err != nil {
			return fmt.Errorf("clear feedback folder links: %w", err)
		}
	}

	return nil
}
