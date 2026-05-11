package moderation

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"openshare/backend/internal/model"
	"openshare/backend/internal/resources"
	"openshare/backend/internal/storage"
)

var (
	ErrSubmissionNotPending           = errors.New("submission is not pending")
	ErrSubmissionMissing              = errors.New("submission not found")
	ErrStagedFileMissing              = errors.New("staged file not found")
	ErrSubmissionReviewReasonRequired = errors.New("submission review reason is required")
	ErrApproveNoFolder                = errors.New("cannot approve: file has no target folder")
	ErrApproveFolderMissing           = errors.New("cannot approve: target folder not found or has no source path")
	ErrInvalidModerationQuery         = errors.New("invalid moderation query")
)

const (
	defaultPendingSubmissionPage     = 1
	defaultPendingSubmissionPageSize = 20
	maxPendingSubmissionPageSize     = 100
)

type ModerationService struct {
	repository *ModerationRepository
	storage    *storage.Service
	nowFunc    func() time.Time
}

type PendingSubmissionItem struct {
	SubmissionID string                 `json:"submission_id"`
	ReceiptCode  string                 `json:"receipt_code"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	RelativePath string                 `json:"relative_path"`
	ReviewReason string                 `json:"review_reason"`
	Status       model.SubmissionStatus `json:"status"`
	UploadedAt   time.Time              `json:"uploaded_at"`
	Size         int64                  `json:"size"`
	MimeType     string                 `json:"mime_type"`
}

type PendingSubmissionListInput struct {
	Page     int
	PageSize int
}

type PendingSubmissionListResult struct {
	Items    []PendingSubmissionItem `json:"items"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
	Total    int64                   `json:"total"`
}

type ReviewResult struct {
	SubmissionID string                 `json:"submission_id"`
	Status       model.SubmissionStatus `json:"status"`
	ReviewedAt   time.Time              `json:"reviewed_at"`
	ReviewReason string                 `json:"review_reason,omitempty"`
}

func NewModerationService(repository *ModerationRepository, storageService *storage.Service) *ModerationService {
	return &ModerationService{
		repository: repository,
		storage:    storageService,
		nowFunc:    func() time.Time { return time.Now().UTC() },
	}
}

func (s *ModerationService) ListPendingSubmissions(ctx context.Context, input PendingSubmissionListInput) (*PendingSubmissionListResult, error) {
	page, pageSize, err := normalizePendingSubmissionPagination(input.Page, input.PageSize)
	if err != nil {
		return nil, err
	}

	rows, total, err := s.repository.ListPendingSubmissions(ctx, PendingSubmissionListQuery{
		Offset: (page - 1) * pageSize,
		Limit:  pageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("list pending submissions: %w", err)
	}

	items := make([]PendingSubmissionItem, 0, len(rows))
	displayPathByFolder := make(map[string]string)
	for _, row := range rows {
		displayPath := resources.NormalizeRelativePathForStorage(row.RelativePath)
		if row.FolderID != nil && strings.TrimSpace(*row.FolderID) != "" {
			folderID := strings.TrimSpace(*row.FolderID)
			rootDisplayPath, exists := displayPathByFolder[folderID]
			if !exists {
				rootDisplayPath, err = s.repository.BuildFolderDisplayPath(ctx, row.FolderID)
				if err != nil {
					return nil, fmt.Errorf("build pending submission display path: %w", err)
				}
				displayPathByFolder[folderID] = rootDisplayPath
			}
			displayPath = resources.BuildSubmissionDisplayPath(rootDisplayPath, row.RelativePath)
		}
		items = append(items, PendingSubmissionItem{
			SubmissionID: row.SubmissionID,
			ReceiptCode:  row.ReceiptCode,
			Name:         row.Name,
			Description:  row.Description,
			RelativePath: displayPath,
			ReviewReason: row.ReviewReason,
			Status:       row.Status,
			UploadedAt:   row.CreatedAt,
			Size:         row.Size,
			MimeType:     row.MimeType,
		})
	}

	return &PendingSubmissionListResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func normalizePendingSubmissionPagination(page int, pageSize int) (int, int, error) {
	if page == 0 {
		page = defaultPendingSubmissionPage
	}
	if pageSize == 0 {
		pageSize = defaultPendingSubmissionPageSize
	}
	if page < 1 || pageSize < 1 || pageSize > maxPendingSubmissionPageSize {
		return 0, 0, ErrInvalidModerationQuery
	}
	return page, pageSize, nil
}

func (s *ModerationService) ApproveSubmission(ctx context.Context, submissionID string, adminID string, operatorIP string) (*ReviewResult, error) {
	record, err := s.repository.FindPendingSubmission(ctx, strings.TrimSpace(submissionID))
	if err != nil {
		return nil, fmt.Errorf("find submission for approval: %w", err)
	}
	if record == nil {
		return nil, ErrSubmissionMissing
	}
	if record.Submission.Status != model.SubmissionStatusPending {
		return nil, ErrSubmissionNotPending
	}

	exists, err := s.storage.StagedFileExists(record.Submission.StagingPath)
	if err != nil {
		return nil, fmt.Errorf("validate staged file: %w", err)
	}
	if !exists {
		return nil, ErrStagedFileMissing
	}

	// Resolve the target folder's disk directory.
	if record.Submission.FolderID == nil {
		return nil, ErrApproveNoFolder
	}
	folder, err := s.repository.FindFolderByID(ctx, *record.Submission.FolderID)
	if err != nil {
		return nil, fmt.Errorf("find target folder: %w", err)
	}
	if folder == nil || folder.SourcePath == nil {
		return nil, ErrApproveFolderMissing
	}

	rootFolderDisplayPath, err := s.repository.BuildFolderDisplayPath(ctx, &folder.ID)
	if err != nil {
		return nil, fmt.Errorf("resolve approval folder path: %w", err)
	}
	relativePath := resources.NormalizeStoredSubmissionRelativePath(rootFolderDisplayPath, record.Submission.RelativePath)

	targetFolder, err := s.ensureApprovalTargetFolder(ctx, folder, relativePath)
	if err != nil {
		return nil, err
	}

	// Move staged file into the folder's physical directory.
	finalPath, finalName, err := s.storage.MoveStagedFileToFolder(record.Submission.StagingPath, *targetFolder.SourcePath, record.Submission.Name)
	if err != nil {
		return nil, fmt.Errorf("move staged file to folder: %w", err)
	}
	finalRelativePath := resources.ReplaceRelativePathBase(relativePath, finalName)

	reviewedAt := s.nowFunc()
	if err := s.repository.ApproveSubmission(
		ctx,
		record.Submission.ID,
		adminID,
		operatorIP,
		reviewedAt,
		targetFolder.ID,
		finalName,
		finalRelativePath,
	); err != nil {
		// Rollback: move the file back to staging.
		if _, rollbackErr := s.storage.MoveFileBackToStaging(finalPath, record.Submission.StagingPath); rollbackErr != nil {
			return nil, fmt.Errorf("approve submission failed (%v); rollback failed: %w", err, rollbackErr)
		}
		return nil, fmt.Errorf("approve submission: %w", err)
	}

	return &ReviewResult{
		SubmissionID: record.Submission.ID,
		Status:       model.SubmissionStatusApproved,
		ReviewedAt:   reviewedAt,
	}, nil
}

func (s *ModerationService) ensureApprovalTargetFolder(ctx context.Context, rootFolder *model.Folder, relativePath string) (*model.Folder, error) {
	relativeDir := resources.NormalizeRelativePathForStorage(filepath.ToSlash(filepath.Dir(strings.TrimSpace(relativePath))))
	if relativeDir == "" {
		return rootFolder, nil
	}
	if rootFolder.SourcePath == nil || strings.TrimSpace(*rootFolder.SourcePath) == "" {
		return nil, ErrApproveFolderMissing
	}
	targetPath := filepath.Join(*rootFolder.SourcePath, filepath.FromSlash(relativeDir))
	if err := s.storage.EnsureManagedDirectory(targetPath); err != nil {
		return nil, fmt.Errorf("ensure approval directory: %w", err)
	}

	var targetFolder *model.Folder
	now := s.nowFunc()
	if err := s.repository.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		leaf, ensureErr := resources.EnsureManagedFolderPathTx(tx, rootFolder, relativeDir, now)
		if ensureErr != nil {
			return ensureErr
		}
		targetFolder = leaf
		return nil
	}); err != nil {
		return nil, fmt.Errorf("ensure approval folder path: %w", err)
	}
	return targetFolder, nil
}

func (s *ModerationService) RejectSubmission(ctx context.Context, submissionID string, adminID string, operatorIP string, reviewReason string) (*ReviewResult, error) {
	reviewReason = strings.TrimSpace(reviewReason)
	if reviewReason == "" {
		return nil, ErrSubmissionReviewReasonRequired
	}

	record, err := s.repository.FindPendingSubmission(ctx, strings.TrimSpace(submissionID))
	if err != nil {
		return nil, fmt.Errorf("find submission for rejection: %w", err)
	}
	if record == nil {
		return nil, ErrSubmissionMissing
	}
	if record.Submission.Status != model.SubmissionStatusPending {
		return nil, ErrSubmissionNotPending
	}

	exists, err := s.storage.StagedFileExists(record.Submission.StagingPath)
	if err != nil {
		return nil, fmt.Errorf("validate staged file: %w", err)
	}
	if exists {
		if err := s.storage.DeleteStagedFile(record.Submission.StagingPath); err != nil {
			return nil, fmt.Errorf("delete staged file: %w", err)
		}
	}

	reviewedAt := s.nowFunc()
	if err := s.repository.RejectSubmission(ctx, record.Submission.ID, adminID, operatorIP, reviewedAt, reviewReason); err != nil {
		return nil, fmt.Errorf("reject submission: %w", err)
	}

	return &ReviewResult{
		SubmissionID: record.Submission.ID,
		Status:       model.SubmissionStatusRejected,
		ReviewedAt:   reviewedAt,
		ReviewReason: reviewReason,
	}, nil
}
