package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminDashboardHandler struct {
	service *AdminDashboardService
}

func NewAdminDashboardHandler(service *AdminDashboardService) *AdminDashboardHandler {
	return &AdminDashboardHandler{service: service}
}

func (h *AdminDashboardHandler) GetStats(ctx *gin.Context) {
	stats, err := h.service.GetStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dashboard stats"})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}
