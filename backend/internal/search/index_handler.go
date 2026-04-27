package search

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SearchIndexHandler struct {
	service *SearchIndexService
}

func NewSearchIndexHandler(service *SearchIndexService) *SearchIndexHandler {
	return &SearchIndexHandler{service: service}
}

func (h *SearchIndexHandler) Status(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, h.service.Status(ctx.Request.Context()))
}

func (h *SearchIndexHandler) Rebuild(ctx *gin.Context) {
	result, err := h.service.Rebuild(ctx.Request.Context())
	if err != nil {
		switch {
		case errors.Is(err, ErrSearchIndexDisabled):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "search index is disabled"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, result)
}
