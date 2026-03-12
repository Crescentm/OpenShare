package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"openshare/backend/internal/repository"
	"openshare/backend/internal/storage"
)

var (
	ErrDownloadFileNotFound    = errors.New("download file not found")
	ErrDownloadFileUnavailable = errors.New("download file unavailable")
	ErrPreviewUnsupported      = errors.New("preview is not supported for this file")
	ErrBatchDownloadInvalid    = errors.New("invalid batch download request")
)

type PublicDownloadService struct {
	repository *repository.PublicDownloadRepository
	storage    *storage.Service
}

type DownloadableFile struct {
	FileID       string
	OriginalName string
	MimeType     string
	Size         int64
	ModTime      time.Time
	Content      *os.File
}

type FilePreviewKind string

const (
	FilePreviewNone  FilePreviewKind = "none"
	FilePreviewPDF   FilePreviewKind = "pdf"
	FilePreviewImage FilePreviewKind = "image"
	FilePreviewText  FilePreviewKind = "text"
)

type PublicFileDetail struct {
	ID            string          `json:"id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	OriginalName  string          `json:"original_name"`
	MimeType      string          `json:"mime_type"`
	Size          int64           `json:"size"`
	Tags          []string        `json:"tags"`
	UploadedAt    time.Time       `json:"uploaded_at"`
	DownloadCount int64           `json:"download_count"`
	PreviewKind   FilePreviewKind `json:"preview_kind"`
	CanPreview    bool            `json:"can_preview"`
}

type BatchDownloadFile struct {
	FileID       string
	OriginalName string
	DiskPath     string
}

func NewPublicDownloadService(repository *repository.PublicDownloadRepository, storageService *storage.Service) *PublicDownloadService {
	return &PublicDownloadService{
		repository: repository,
		storage:    storageService,
	}
}

func (s *PublicDownloadService) PrepareDownload(ctx context.Context, fileID string) (*DownloadableFile, error) {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return nil, ErrDownloadFileNotFound
	}

	file, err := s.repository.FindActiveFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("find downloadable file: %w", err)
	}
	if file == nil {
		return nil, ErrDownloadFileNotFound
	}

	opened, err := s.storage.OpenManagedFile(file.DiskPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrDownloadFileUnavailable
		}
		return nil, fmt.Errorf("open downloadable file: %w", err)
	}

	return &DownloadableFile{
		FileID:       file.ID,
		OriginalName: file.OriginalName,
		MimeType:     file.MimeType,
		Size:         opened.Info.Size(),
		ModTime:      opened.Info.ModTime(),
		Content:      opened.File,
	}, nil
}

func (s *PublicDownloadService) PreparePreview(ctx context.Context, fileID string) (*DownloadableFile, FilePreviewKind, error) {
	download, err := s.PrepareDownload(ctx, fileID)
	if err != nil {
		return nil, FilePreviewNone, err
	}
	kind := previewKind(download.OriginalName, download.MimeType)
	if kind == FilePreviewNone {
		download.Content.Close()
		return nil, FilePreviewNone, ErrPreviewUnsupported
	}
	return download, kind, nil
}

func (s *PublicDownloadService) GetFileDetail(ctx context.Context, fileID string) (*PublicFileDetail, error) {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return nil, ErrDownloadFileNotFound
	}
	file, err := s.repository.FindActiveFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("find public file detail: %w", err)
	}
	if file == nil {
		return nil, ErrDownloadFileNotFound
	}

	tagsByFile, err := s.repository.ListTagsByFileIDs(ctx, []string{file.ID})
	if err != nil {
		return nil, fmt.Errorf("list file detail tags: %w", err)
	}
	kind := previewKind(file.OriginalName, file.MimeType)
	return &PublicFileDetail{
		ID:            file.ID,
		Title:         file.Title,
		Description:   file.Description,
		OriginalName:  file.OriginalName,
		MimeType:      file.MimeType,
		Size:          file.Size,
		Tags:          tagsByFile[file.ID],
		UploadedAt:    file.CreatedAt,
		DownloadCount: file.DownloadCount,
		PreviewKind:   kind,
		CanPreview:    kind != FilePreviewNone,
	}, nil
}

func (s *PublicDownloadService) PrepareBatchDownload(ctx context.Context, fileIDs []string) ([]BatchDownloadFile, error) {
	normalized := normalizeBatchFileIDs(fileIDs)
	if len(normalized) == 0 {
		return nil, ErrBatchDownloadInvalid
	}

	files, err := s.repository.ListActiveFilesByIDs(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("list batch download files: %w", err)
	}
	if len(files) != len(normalized) {
		return nil, ErrDownloadFileNotFound
	}

	items := make([]BatchDownloadFile, 0, len(files))
	for _, file := range files {
		opened, err := s.storage.OpenManagedFile(file.DiskPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, ErrDownloadFileUnavailable
			}
			return nil, fmt.Errorf("validate batch download file: %w", err)
		}
		opened.File.Close()

		items = append(items, BatchDownloadFile{
			FileID:       file.ID,
			OriginalName: file.OriginalName,
			DiskPath:     file.DiskPath,
		})
	}
	return items, nil
}

func (s *PublicDownloadService) RecordDownloadAsync(fileID string) {
	go func() {
		if err := s.repository.IncrementDownloadCount(context.Background(), fileID); err != nil {
			log.Printf("increment download count for file %s: %v", fileID, err)
		}
	}()
}

func (s *PublicDownloadService) RecordBatchDownloadAsync(fileIDs []string) {
	normalized := normalizeBatchFileIDs(fileIDs)
	if len(normalized) == 0 {
		return
	}

	go func() {
		if err := s.repository.IncrementDownloadCounts(context.Background(), normalized); err != nil {
			log.Printf("increment download counts for files %v: %v", normalized, err)
		}
	}()
}

func previewKind(originalName string, mimeType string) FilePreviewKind {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	switch {
	case mimeType == "application/pdf" || strings.EqualFold(filepath.Ext(originalName), ".pdf"):
		return FilePreviewPDF
	case strings.HasPrefix(mimeType, "image/"):
		return FilePreviewImage
	case strings.HasPrefix(mimeType, "text/"):
		return FilePreviewText
	}

	switch strings.ToLower(strings.TrimSpace(filepath.Ext(originalName))) {
	case ".txt", ".md", ".log", ".json", ".csv":
		return FilePreviewText
	default:
		return FilePreviewNone
	}
}

func normalizeBatchFileIDs(fileIDs []string) []string {
	normalized := make([]string, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		fileID = strings.TrimSpace(fileID)
		if fileID == "" || slices.Contains(normalized, fileID) {
			continue
		}
		normalized = append(normalized, fileID)
	}
	return normalized
}
