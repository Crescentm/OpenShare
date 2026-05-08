package search

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SearchHandler exposes the public search API.
type SearchHandler struct {
	service *SearchService
}

func NewSearchHandler(service *SearchService) *SearchHandler {
	return &SearchHandler{service: service}
}

// Search handles GET /api/public/search
//
//	Query parameters:
//	  q         – keyword
//	  folder_id – optional folder scope
//	  type      – optional resource type: file or folder
//	  file_kind – optional file kind: pdf, office, image, archive, ...
//	  extension – optional file extension
//	  category  – optional semantic category
//	  course    – optional course name
//	  material_type – optional material type
//	  content_status – optional extracted content status
//	  page      – page number (default 1)
//	  page_size – results per page (default 20, max 100)
func (h *SearchHandler) Search(ctx *gin.Context) {
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

	result, err := h.service.Search(ctx.Request.Context(), SearchInput{
		Keyword:       ctx.Query("q"),
		FolderID:      ctx.Query("folder_id"),
		Type:          ctx.Query("type"),
		FileKind:      ctx.Query("file_kind"),
		Extension:     ctx.Query("extension"),
		Category:      ctx.Query("category"),
		Course:        ctx.Query("course"),
		MaterialType:  ctx.Query("material_type"),
		ContentStatus: ctx.Query("content_status"),
		Page:          page,
		PageSize:      pageSize,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrSearchQueryEmpty):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "search query is empty"})
		case errors.Is(err, ErrSearchQueryTooLong):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "search query is too long"})
		case errors.Is(err, ErrSearchInvalidInput):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid search parameters"})
		case errors.Is(err, ErrSearchIndexDisabled):
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "search index is disabled"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
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
