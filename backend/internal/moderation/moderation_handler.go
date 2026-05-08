package moderation

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"openshare/backend/internal/session"
)

type ModerationHandler struct {
	service       *ModerationService
	indexNotifier SearchIndexNotifier
}

type SearchIndexNotifier interface {
	NotifySearchResourcesChanged(reason string)
}

type reviewSubmissionRequest struct {
	ReviewReason string `json:"review_reason"`
	RejectReason string `json:"reject_reason"`
}

func NewModerationHandler(service *ModerationService, indexNotifier SearchIndexNotifier) *ModerationHandler {
	return &ModerationHandler{service: service, indexNotifier: indexNotifier}
}

func (h *ModerationHandler) ListPendingSubmissions(ctx *gin.Context) {
	page, err := parseIntQuery(ctx.Query("page"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}
	pageSize, err := parseIntQuery(ctx.Query("page_size"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size"})
		return
	}

	result, err := h.service.ListPendingSubmissions(ctx.Request.Context(), PendingSubmissionListInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidModerationQuery) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid moderation query"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list pending submissions"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *ModerationHandler) ApproveSubmission(ctx *gin.Context) {
	identity, ok := session.GetAdminIdentity(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	result, err := h.service.ApproveSubmission(ctx.Request.Context(), ctx.Param("submissionID"), identity.AdminID, ctx.ClientIP())
	if err != nil {
		switch {
		case errors.Is(err, ErrSubmissionMissing):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		case errors.Is(err, ErrSubmissionNotPending):
			ctx.JSON(http.StatusConflict, gin.H{"error": "submission is not pending"})
		case errors.Is(err, ErrStagedFileMissing):
			ctx.JSON(http.StatusConflict, gin.H{"error": "staged file is missing"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve submission"})
		}
		return
	}

	if h.indexNotifier != nil {
		h.indexNotifier.NotifySearchResourcesChanged("submission_approved")
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *ModerationHandler) RejectSubmission(ctx *gin.Context) {
	identity, ok := session.GetAdminIdentity(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req reviewSubmissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	reviewReason := req.ReviewReason
	if reviewReason == "" {
		reviewReason = req.RejectReason
	}

	result, err := h.service.RejectSubmission(ctx.Request.Context(), ctx.Param("submissionID"), identity.AdminID, ctx.ClientIP(), reviewReason)
	if err != nil {
		switch {
		case errors.Is(err, ErrSubmissionReviewReasonRequired):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "review_reason is required"})
		case errors.Is(err, ErrSubmissionMissing):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		case errors.Is(err, ErrSubmissionNotPending):
			ctx.JSON(http.StatusConflict, gin.H{"error": "submission is not pending"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject submission"})
		}
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func parseIntQuery(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	return strconv.Atoi(raw)
}
