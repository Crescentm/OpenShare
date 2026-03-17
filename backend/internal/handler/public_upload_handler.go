package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"openshare/backend/internal/service"
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
	request, err := h.parseSubmissionRequest(ctx)
	if err != nil {
		switch {
		case errors.Is(err, errUploadFormParse):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse upload form"})
		case errors.Is(err, errUploadManifestInvalid):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload manifest"})
		case errors.Is(err, errUploadManifestMismatch):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "upload files do not match manifest"})
		case errors.Is(err, errUploadFileRequired):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		case errors.Is(err, errUploadFileRead):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uploaded file"})
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse upload form"})
		}
		return
	}
	defer request.Close()

	result, err := h.service.CreateSubmission(ctx.Request.Context(), request.input)
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
