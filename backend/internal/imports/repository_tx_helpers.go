package imports

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"openshare/backend/internal/model"
)

func createOperationLogTx(
	tx *gorm.DB,
	logID string,
	adminID string,
	action string,
	targetType string,
	targetID string,
	detail string,
	operatorIP string,
	now time.Time,
) error {
	var adminRef *string
	if adminID != "" {
		adminRef = &adminID
	}

	entry := &model.OperationLog{
		ID:         logID,
		AdminID:    adminRef,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
		IP:         operatorIP,
		CreatedAt:  now,
	}
	if err := tx.Create(entry).Error; err != nil {
		return fmt.Errorf("create operation log: %w", err)
	}
	return nil
}

func detachDeletedResourcesTx(tx *gorm.DB, fileIDs []string, folderIDs []string) error {
	if len(fileIDs) > 0 {
		if err := tx.Model(&model.Submission{}).
			Where("file_id IN ?", fileIDs).
			Update("file_id", nil).Error; err != nil {
			return fmt.Errorf("clear submission file links: %w", err)
		}
		if err := tx.Model(&model.Feedback{}).
			Where("file_id IN ?", fileIDs).
			Update("file_id", nil).Error; err != nil {
			return fmt.Errorf("clear feedback file links: %w", err)
		}
	}

	if len(folderIDs) > 0 {
		if err := tx.Model(&model.Submission{}).
			Where("folder_id IN ?", folderIDs).
			Update("folder_id", nil).Error; err != nil {
			return fmt.Errorf("clear submission folder links: %w", err)
		}
		if err := tx.Model(&model.Feedback{}).
			Where("folder_id IN ?", folderIDs).
			Update("folder_id", nil).Error; err != nil {
			return fmt.Errorf("clear feedback folder links: %w", err)
		}
	}

	return nil
}
