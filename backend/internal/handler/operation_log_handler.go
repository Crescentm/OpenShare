package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"openshare/backend/internal/service"
)

type OperationLogHandler struct {
	service *service.OperationLogService
}

func NewOperationLogHandler(service *service.OperationLogService) *OperationLogHandler {
	return &OperationLogHandler{service: service}
}

func (h *OperationLogHandler) List(ctx *gin.Context) {
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

	result, err := h.service.List(ctx.Request.Context(), service.ListOperationLogsInput{
		Action:     ctx.Query("action"),
		TargetType: ctx.Query("target_type"),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidOperationLogQuery) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid operation log query"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load operation logs"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
