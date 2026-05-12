package operations

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"openshare/backend/internal/httpapi"
)

type OperationLogHandler struct {
	service *OperationLogService
}

func NewOperationLogHandler(service *OperationLogService) *OperationLogHandler {
	return &OperationLogHandler{service: service}
}

func (h *OperationLogHandler) List(ctx *gin.Context) {
	page, err := httpapi.ParseIntQuery(ctx.Query("page"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}

	pageSize, err := httpapi.ParseIntQuery(ctx.Query("page_size"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size"})
		return
	}

	result, err := h.service.List(ctx.Request.Context(), ListOperationLogsInput{
		Action:     ctx.Query("action"),
		TargetType: ctx.Query("target_type"),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidOperationLogQuery) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid operation log query"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load operation logs"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
