package service

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"openshare/backend/internal/config"
	"openshare/backend/internal/model"
	"openshare/backend/internal/repository"
	"openshare/backend/internal/storage"
	"openshare/backend/pkg/identity"
)

var (
	ErrInvalidUploadInput   = errors.New("invalid upload input")
	ErrUploadReceiptExists  = errors.New("receipt code already exists")
	ErrUploadFileTooLarge   = errors.New("upload file too large")
	ErrUploadEmptyFile      = errors.New("upload file is empty")
	ErrInvalidFileExtension = errors.New("invalid file extension")
	ErrInvalidFileMIMEType  = errors.New("invalid file mime type")
	ErrReceiptCodeGenerate  = errors.New("failed to generate receipt code")
	ErrUploadFolderRequired = errors.New("upload target folder is required")
	ErrUploadFolderNotFound = errors.New("upload target folder not found")
)

const maxGeneratedReceiptAttempts = 5

type PublicUploadService struct {
	config        config.UploadConfig
	repository    *repository.UploadRepository
	storage       *storage.Service
	systemSetting *SystemSettingService
	nowFunc       func() time.Time
	codeGen       func(int) (string, error)
}

type PublicUploadInput struct {
	Description  string
	Tags         []string
	ReceiptCode  string
	FolderID     string
	OriginalName string
	DeclaredMIME string
	UploaderIP   string
	File         io.Reader
}

type PublicUploadResult struct {
	ReceiptCode string                 `json:"receipt_code"`
	Status      model.SubmissionStatus `json:"status"`
	Title       string                 `json:"title"`
	UploadedAt  time.Time              `json:"uploaded_at"`
}

func NewPublicUploadService(
	cfg config.UploadConfig,
	repository *repository.UploadRepository,
	storageService *storage.Service,
	systemSettingService *SystemSettingService,
) *PublicUploadService {
	return &PublicUploadService{
		config:        cfg,
		repository:    repository,
		storage:       storageService,
		systemSetting: systemSettingService,
		nowFunc:       func() time.Time { return time.Now().UTC() },
		codeGen:       generateReceiptCode,
	}
}

func (s *PublicUploadService) CreateSubmission(ctx context.Context, input PublicUploadInput) (*PublicUploadResult, error) {
	policy := s.effectivePolicy(ctx)
	normalized, err := s.normalizeInput(input, policy)
	if err != nil {
		return nil, err
	}

	if normalized.FolderID == "" {
		return nil, ErrUploadFolderRequired
	}

	folder, err := s.repository.FindActiveFolderByID(ctx, normalized.FolderID)
	if err != nil {
		return nil, fmt.Errorf("validate upload folder: %w", err)
	}
	if folder == nil || folder.SourcePath == nil || strings.TrimSpace(*folder.SourcePath) == "" {
		return nil, ErrUploadFolderNotFound
	}
	folderID := normalized.FolderID

	bufferedReader := bufio.NewReader(normalized.File)
	head, err := bufferedReader.Peek(512)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("inspect upload file: %w", err)
	}

	detectedMIME := strings.ToLower(strings.TrimSpace(http.DetectContentType(head)))
	if !s.isAllowedMIME(detectedMIME, normalized.DeclaredMIME) {
		return nil, ErrInvalidFileMIMEType
	}

	maxFileSizeBytes := s.config.MaxFileSizeBytes
	if policy.Upload.MaxFileSizeBytes > 0 {
		maxFileSizeBytes = policy.Upload.MaxFileSizeBytes
	}
	stagedFile, err := s.storage.SaveToStaging(bufferedReader, normalized.Extension, maxFileSizeBytes)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrFileTooLarge):
			return nil, ErrUploadFileTooLarge
		case strings.Contains(strings.ToLower(err.Error()), "empty file"):
			return nil, ErrUploadEmptyFile
		default:
			return nil, fmt.Errorf("save upload to staging: %w", err)
		}
	}

	defer func() {
		if err != nil {
			_ = s.storage.DeleteStagedFile(stagedFile.DiskPath)
		}
	}()

	receiptCode, err := s.resolveReceiptCode(ctx, normalized.ReceiptCode)
	if err != nil {
		return nil, err
	}

	now := s.nowFunc()
	submissionID, err := identity.NewID()
	if err != nil {
		return nil, fmt.Errorf("generate submission id: %w", err)
	}
	fileID, err := identity.NewID()
	if err != nil {
		return nil, fmt.Errorf("generate file id: %w", err)
	}

	tagsSnapshot, err := json.Marshal(normalized.Tags)
	if err != nil {
		return nil, fmt.Errorf("encode tags snapshot: %w", err)
	}

	submission := &model.Submission{
		ID:                  submissionID,
		ReceiptCode:         receiptCode,
		TitleSnapshot:       normalized.Title,
		DescriptionSnapshot: normalized.Description,
		TagsSnapshot:        string(tagsSnapshot),
		Status:              model.SubmissionStatusPending,
		UploaderIP:          normalized.UploaderIP,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	submissionRef := submissionID
	file := &model.File{
		ID:            fileID,
		FolderID:      &folderID,
		SubmissionID:  &submissionRef,
		Title:         normalized.Title,
		Description:   normalized.Description,
		OriginalName:  normalized.OriginalName,
		StoredName:    stagedFile.StoredName,
		Extension:     normalized.Extension,
		MimeType:      detectedMIME,
		Size:          stagedFile.Size,
		DiskPath:      stagedFile.DiskPath,
		Status:        model.ResourceStatusOffline,
		DownloadCount: 0,
		UploaderIP:    normalized.UploaderIP,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if policy.Guest.AllowDirectPublish {
		finalPath, finalName, moveErr := s.storage.MoveStagedFileToFolder(stagedFile.DiskPath, *folder.SourcePath, normalized.OriginalName)
		if moveErr != nil {
			err = fmt.Errorf("move direct-publish upload: %w", moveErr)
			return nil, err
		}

		submission.Status = model.SubmissionStatusApproved
		submission.ReviewedAt = &now
		file.Status = model.ResourceStatusActive
		file.DiskPath = finalPath
		file.StoredName = finalName

		if createErr := s.repository.CreateUpload(ctx, submission, file); createErr != nil {
			err = fmt.Errorf("persist direct-publish upload: %w", createErr)
			if _, rollbackErr := s.storage.MoveFileBackToStaging(finalPath, stagedFile.StoredName); rollbackErr != nil {
				err = fmt.Errorf("%w; rollback failed: %v", err, rollbackErr)
			}
			return nil, err
		}

		return &PublicUploadResult{
			ReceiptCode: receiptCode,
			Status:      model.SubmissionStatusApproved,
			Title:       normalized.Title,
			UploadedAt:  now,
		}, nil
	}

	if createErr := s.repository.CreateUpload(ctx, submission, file); createErr != nil {
		err = fmt.Errorf("persist upload submission: %w", createErr)
		return nil, err
	}

	return &PublicUploadResult{
		ReceiptCode: receiptCode,
		Status:      model.SubmissionStatusPending,
		Title:       normalized.Title,
		UploadedAt:  now,
	}, nil
}

type normalizedUploadInput struct {
	Title        string
	Description  string
	Tags         []string
	ReceiptCode  string
	FolderID     string
	OriginalName string
	DeclaredMIME string
	UploaderIP   string
	Extension    string
	File         io.Reader
}

func (s *PublicUploadService) normalizeInput(input PublicUploadInput, policy SystemPolicy) (*normalizedUploadInput, error) {
	description := strings.TrimSpace(input.Description)
	if len([]rune(description)) > s.config.MaxDescriptionLength {
		return nil, ErrInvalidUploadInput
	}

	maxTagCount := s.config.MaxTagCount
	if policy.Upload.MaxTagCount >= 0 {
		maxTagCount = policy.Upload.MaxTagCount
	}
	tags, err := normalizeTags(input.Tags, maxTagCount, s.config.MaxTagLength)
	if err != nil {
		return nil, ErrInvalidUploadInput
	}

	receiptCode, err := normalizeReceiptCode(input.ReceiptCode)
	if err != nil {
		return nil, ErrInvalidUploadInput
	}

	originalName := filepath.Base(strings.TrimSpace(input.OriginalName))
	if originalName == "" || originalName == "." {
		return nil, ErrInvalidUploadInput
	}

	extension := strings.ToLower(strings.TrimSpace(filepath.Ext(originalName)))
	if !isAllowedExtension(extension, policy.Upload.AllowedExtensions) {
		return nil, ErrInvalidFileExtension
	}

	// Title is always derived from the original filename (immutable).
	title := strings.TrimSuffix(originalName, filepath.Ext(originalName))
	if title == "" {
		title = originalName
	}

	return &normalizedUploadInput{
		Title:        title,
		Description:  description,
		Tags:         tags,
		ReceiptCode:  receiptCode,
		FolderID:     strings.TrimSpace(input.FolderID),
		OriginalName: originalName,
		DeclaredMIME: strings.ToLower(strings.TrimSpace(input.DeclaredMIME)),
		UploaderIP:   strings.TrimSpace(input.UploaderIP),
		Extension:    extension,
		File:         input.File,
	}, nil
}

func (s *PublicUploadService) effectivePolicy(ctx context.Context) SystemPolicy {
	if s.systemSetting == nil {
		return defaultSystemPolicy(s.config)
	}

	policy, err := s.systemSetting.GetPolicy(ctx)
	if err != nil || policy == nil {
		return defaultSystemPolicy(s.config)
	}

	return *policy
}

func (s *PublicUploadService) resolveReceiptCode(ctx context.Context, receiptCode string) (string, error) {
	// If a receipt code is provided (user-defined or cached from previous upload),
	// reuse it directly. Multiple submissions can share one receipt code.
	if receiptCode != "" {
		return receiptCode, nil
	}

	// Auto-generate a new receipt code
	candidate, err := s.codeGen(s.config.ReceiptCodeLength)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrReceiptCodeGenerate, err)
	}

	return candidate, nil
}

// retryCreateWithGeneratedReceipt is kept for backward compatibility but
// receipt code conflicts should no longer occur since the unique constraint
// was replaced with a regular index.

func isAllowedExtension(extension string, allowedExtensions []string) bool {
	if len(allowedExtensions) == 0 {
		return true
	}
	for _, allowed := range allowedExtensions {
		if strings.EqualFold(extension, strings.TrimSpace(allowed)) {
			return true
		}
	}
	return false
}

func (s *PublicUploadService) isAllowedMIME(detectedMIME, declaredMIME string) bool {
	if len(s.config.AllowedMIMETypes) == 0 {
		return true
	}
	for _, allowed := range s.config.AllowedMIMETypes {
		if detectedMIME == allowed || declaredMIME == allowed {
			return true
		}
	}
	return false
}

func normalizeTags(tags []string, maxCount, maxLength int) ([]string, error) {
	normalized := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, raw := range tags {
		for _, part := range strings.Split(raw, ",") {
			tag := strings.TrimSpace(part)
			if tag == "" {
				continue
			}
			if len([]rune(tag)) > maxLength {
				return nil, ErrInvalidUploadInput
			}
			key := strings.ToLower(tag)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			normalized = append(normalized, tag)
			if maxCount > 0 && len(normalized) > maxCount {
				return nil, ErrInvalidUploadInput
			}
		}
	}

	return normalized, nil
}

func normalizeReceiptCode(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if len(raw) < 6 || len(raw) > 64 {
		return "", ErrInvalidUploadInput
	}
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return "", ErrInvalidUploadInput
		}
	}

	return raw, nil
}

func generateReceiptCode(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	builder := strings.Builder{}
	builder.Grow(length)
	for _, b := range raw {
		builder.WriteByte(alphabet[int(b)%len(alphabet)])
	}

	return builder.String(), nil
}
