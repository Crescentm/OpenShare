package submissions

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"openshare/backend/internal/receipts"
)

type PublicSubmissionHandler struct {
	service *PublicSubmissionService
}

func NewPublicSubmissionHandler(service *PublicSubmissionService) *PublicSubmissionHandler {
	return &PublicSubmissionHandler{service: service}
}

func (h *PublicSubmissionHandler) LookupByReceiptCode(ctx *gin.Context) {
	result, err := h.service.LookupByReceiptCode(ctx.Request.Context(), ctx.Param("receiptCode"))
	if err != nil {
		switch {
		case errors.Is(err, receipts.ErrInvalidReceiptCode):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid receipt code"})
		case errors.Is(err, ErrSubmissionNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query submission"})
		}
		return
	}

	ctx.JSON(http.StatusOK, result)
}
