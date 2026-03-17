package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"openshare/backend/internal/model"
	"openshare/backend/internal/repository"
)

var ErrSubmissionNotFound = errors.New("submission not found")

type PublicSubmissionService struct {
	repository *repository.PublicSubmissionRepository
}

type PublicSubmissionLookupResult struct {
	ReceiptCode string                 `json:"receipt_code"`
	Items       []PublicSubmissionItem `json:"items"`
}

type PublicSubmissionItem struct {
	Title         string                 `json:"title"`
	RelativePath  string                 `json:"relative_path"`
	Status        model.SubmissionStatus `json:"status"`
	UploadedAt    time.Time              `json:"uploaded_at"`
	DownloadCount int64                  `json:"download_count"`
	RejectReason  string                 `json:"reject_reason"`
}

func NewPublicSubmissionService(repository *repository.PublicSubmissionRepository) *PublicSubmissionService {
	return &PublicSubmissionService{repository: repository}
}

func (s *PublicSubmissionService) LookupByReceiptCode(ctx context.Context, receiptCode string) (*PublicSubmissionLookupResult, error) {
	normalized, err := normalizeReceiptCode(receiptCode)
	if err != nil {
		return nil, ErrInvalidUploadInput
	}
	if strings.TrimSpace(normalized) == "" {
		return nil, ErrInvalidUploadInput
	}

	rows, err := s.repository.FindAllByReceiptCode(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("lookup submission by receipt code: %w", err)
	}
	if len(rows) == 0 {
		return nil, ErrSubmissionNotFound
	}

	items := make([]PublicSubmissionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, PublicSubmissionItem{
			Title:         row.TitleSnapshot,
			RelativePath:  row.RelativePath,
			Status:        row.Status,
			UploadedAt:    row.CreatedAt.UTC(),
			DownloadCount: row.DownloadCount,
			RejectReason:  row.RejectReason,
		})
	}

	return &PublicSubmissionLookupResult{
		ReceiptCode: normalized,
		Items:       items,
	}, nil
}
