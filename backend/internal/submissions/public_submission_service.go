package submissions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"openshare/backend/internal/model"
	"openshare/backend/internal/receipts"
	"openshare/backend/internal/resources"
)

var ErrSubmissionNotFound = errors.New("submission not found")

type PublicSubmissionService struct {
	repository *PublicSubmissionRepository
}

type PublicSubmissionLookupResult struct {
	ReceiptCode string                 `json:"receipt_code"`
	Items       []PublicSubmissionItem `json:"items"`
}

type PublicSubmissionItem struct {
	Name         string                 `json:"name"`
	RelativePath string                 `json:"relative_path"`
	Status       model.SubmissionStatus `json:"status"`
	UploadedAt   time.Time              `json:"uploaded_at"`
	ReviewReason string                 `json:"review_reason"`
}

func NewPublicSubmissionService(repository *PublicSubmissionRepository) *PublicSubmissionService {
	return &PublicSubmissionService{repository: repository}
}

func (s *PublicSubmissionService) LookupByReceiptCode(ctx context.Context, receiptCode string) (*PublicSubmissionLookupResult, error) {
	normalized, err := receipts.NormalizeReceiptCode(receiptCode)
	if err != nil {
		return nil, receipts.ErrInvalidReceiptCode
	}
	if strings.TrimSpace(normalized) == "" {
		return nil, receipts.ErrInvalidReceiptCode
	}

	rows, err := s.repository.FindAllByReceiptCode(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("lookup submission by receipt code: %w", err)
	}
	if len(rows) == 0 {
		return nil, ErrSubmissionNotFound
	}

	items := make([]PublicSubmissionItem, 0, len(rows))
	displayPathByFolder := make(map[string]string)
	for _, row := range rows {
		displayPath := resources.NormalizeRelativePathForStorage(row.RelativePath)
		if row.FolderID != nil && strings.TrimSpace(*row.FolderID) != "" {
			folderID := strings.TrimSpace(*row.FolderID)
			rootDisplayPath, exists := displayPathByFolder[folderID]
			if !exists {
				rootDisplayPath, err = s.repository.BuildFolderDisplayPath(ctx, row.FolderID)
				if err != nil {
					return nil, fmt.Errorf("build submission display path: %w", err)
				}
				displayPathByFolder[folderID] = rootDisplayPath
			}
			displayPath = resources.BuildSubmissionDisplayPath(rootDisplayPath, row.RelativePath)
		}
		items = append(items, PublicSubmissionItem{
			Name:         row.Name,
			RelativePath: displayPath,
			Status:       row.Status,
			UploadedAt:   row.CreatedAt.UTC(),
			ReviewReason: row.ReviewReason,
		})
	}

	return &PublicSubmissionLookupResult{
		ReceiptCode: normalized,
		Items:       items,
	}, nil
}
