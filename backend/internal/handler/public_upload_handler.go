package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"openshare/backend/internal/model"
	"openshare/backend/internal/service"
	"openshare/backend/internal/session"
)

type PublicUploadHandler struct {
	service         *service.PublicUploadService
	systemSetting   *service.SystemSettingService
	maxRequestBytes int64
}

func NewPublicUploadHandler(service *service.PublicUploadService, systemSetting *service.SystemSettingService, maxRequestBytes int64) *PublicUploadHandler {
	return &PublicUploadHandler{
		service:         service,
		systemSetting:   systemSetting,
		maxRequestBytes: maxRequestBytes,
	}
}

func (h *PublicUploadHandler) CreateSubmission(ctx *gin.Context) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, h.currentMaxRequestBytes(ctx.Request.Context()))

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse upload form"})
		return
	}

	var manifest []struct {
		RelativePath string `json:"relative_path"`
	}
	fileHeaders := form.File["files"]
	manifestRaw := ctx.PostForm("manifest")
	if manifestRaw != "" {
		if err := json.Unmarshal([]byte(manifestRaw), &manifest); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload manifest"})
			return
		}
		if len(fileHeaders) == 0 || len(fileHeaders) != len(manifest) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "upload files do not match manifest"})
			return
		}
	} else {
		fileHeaders = form.File["file"]
		if len(fileHeaders) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
			return
		}
		manifest = make([]struct {
			RelativePath string `json:"relative_path"`
		}, len(fileHeaders))
		for index, fileHeader := range fileHeaders {
			manifest[index].RelativePath = fileHeader.Filename
		}
	}

	files := make([]service.PublicUploadFileInput, 0, len(fileHeaders))
	closers := make([]func(), 0, len(fileHeaders))
	for index, fileHeader := range fileHeaders {
		file, openErr := fileHeader.Open()
		if openErr != nil {
			for _, closer := range closers {
				closer()
			}
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uploaded file"})
			return
		}
		closers = append(closers, func() { _ = file.Close() })
		files = append(files, service.PublicUploadFileInput{
			OriginalName: fileHeader.Filename,
			RelativePath: manifest[index].RelativePath,
			DeclaredMIME: fileHeader.Header.Get("Content-Type"),
			File:         file,
		})
	}
	defer func() {
		for _, closer := range closers {
			closer()
		}
	}()

	result, err := h.service.CreateSubmission(ctx.Request.Context(), service.PublicUploadInput{
		Description:   ctx.PostForm("description"),
		ReceiptCode:   readPublicReceiptCode(ctx),
		FolderID:      ctx.PostForm("folder_id"),
		UploaderIP:    ctx.ClientIP(),
		DirectPublish: canDirectPublish(ctx),
		Files:         files,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUploadInput):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload form"})
		case errors.Is(err, service.ErrUploadFolderRequired):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "target folder is required"})
		case errors.Is(err, service.ErrUploadFolderNotFound):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "target folder not found"})
		case errors.Is(err, service.ErrUploadReceiptExists):
			ctx.JSON(http.StatusConflict, gin.H{"error": "receipt code already exists"})
		case errors.Is(err, service.ErrUploadFileTooLarge):
			ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file is too large"})
		case errors.Is(err, service.ErrUploadEmptyFile):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is empty"})
		case errors.Is(err, service.ErrInvalidFileExtension):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "file extension is not allowed"})
		case errors.Is(err, service.ErrInvalidFileMIMEType):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "file type is not allowed"})
		case errors.Is(err, service.ErrReceiptCodeGenerate):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate receipt code"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create submission"})
		}
		return
	}

	writePublicReceiptCode(ctx, result.ReceiptCode)
	ctx.JSON(http.StatusCreated, result)
}

func canDirectPublish(ctx *gin.Context) bool {
	identity, ok := session.GetAdminIdentity(ctx)
	if !ok {
		return false
	}
	return identity.HasPermission(model.AdminPermissionSubmissionModeration)
}

func (h *PublicUploadHandler) currentMaxRequestBytes(ctx context.Context) int64 {
	limit := h.maxRequestBytes
	if h.systemSetting == nil {
		return limit
	}

	policy, err := h.systemSetting.GetPolicy(ctx)
	if err != nil || policy == nil || policy.Upload.MaxFileSizeBytes <= 0 {
		return limit
	}
	fileBound := policy.Upload.MaxFileSizeBytes + (1 << 20)
	if fileBound > limit {
		return fileBound
	}
	return limit
}
