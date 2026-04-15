package receipts

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PublicReceiptHandler struct {
	receiptCodes *ReceiptCodeService
}

func NewPublicReceiptHandler(receiptCodes *ReceiptCodeService) *PublicReceiptHandler {
	return &PublicReceiptHandler{receiptCodes: receiptCodes}
}

func (h *PublicReceiptHandler) Ensure(ctx *gin.Context) {
	receiptCode, err := h.receiptCodes.ResolveForSession(ctx.Request.Context(), ReadPublicReceiptCode(ctx))
	if err != nil {
		switch {
		case errors.Is(err, ErrReceiptCodeGenerate):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate receipt code"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load receipt code"})
		}
		return
	}

	WritePublicReceiptCode(ctx, receiptCode)
	ctx.JSON(http.StatusOK, gin.H{"receipt_code": receiptCode})
}
