package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"openshare/backend/internal/service"
)

type PublicSubmissionHandler struct {
	service *service.PublicSubmissionService
}

func NewPublicSubmissionHandler(service *service.PublicSubmissionService) *PublicSubmissionHandler {
	return &PublicSubmissionHandler{service: service}
}

func (h *PublicSubmissionHandler) LookupByReceiptCode(ctx *gin.Context) {
	result, err := h.service.LookupByReceiptCode(ctx.Request.Context(), ctx.Param("receiptCode"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUploadInput):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid receipt code"})
		case errors.Is(err, service.ErrSubmissionNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query submission"})
		}
		return
	}

	ctx.JSON(http.StatusOK, result)
}
