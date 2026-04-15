package resources

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
