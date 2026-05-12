package resources

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"openshare/backend/internal/admin"
	"openshare/backend/internal/httpapi"
	"openshare/backend/internal/session"
)

type ResourceManagementHandler struct {
	service       *ResourceManagementService
	authService   *admin.AdminAuthService
	indexNotifier SearchIndexNotifier
}

type SearchIndexNotifier interface {
	NotifySearchResourcesChanged(reason string)
}

type updateManagedFileRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateManagedFolderDescriptionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type deleteManagedResourceRequest struct {
	Password string `json:"password"`
}

func NewResourceManagementHandler(
	service *ResourceManagementService,
	authService *admin.AdminAuthService,
	indexNotifier SearchIndexNotifier,
) *ResourceManagementHandler {
	return &ResourceManagementHandler{
		service:       service,
		authService:   authService,
		indexNotifier: indexNotifier,
	}
}

func (h *ResourceManagementHandler) ListFiles(ctx *gin.Context) {
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

	result, err := h.service.ListFiles(ctx.Request.Context(), ListManagedFilesInput{
		Query:    ctx.Query("q"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidResourceQuery) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource query"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list resources"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (h *ResourceManagementHandler) UpdateFile(ctx *gin.Context) {
	identity, ok := session.GetAdminIdentity(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req updateManagedFileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.service.UpdateFile(ctx.Request.Context(), ctx.Param("fileID"), UpdateManagedFileInput{
		Name:        req.Name,
		Description: req.Description,
		OperatorID:  identity.AdminID,
		OperatorIP:  ctx.ClientIP(),
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidResourceEdit):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource input"})
		case errors.Is(err, ErrManagedFileConflict):
			ctx.JSON(http.StatusConflict, gin.H{"error": "file name already exists"})
		case errors.Is(err, ErrManagedFileNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update resource"})
		}
		return
	}
	h.notifySearchIndex("managed_file_update")
	ctx.Status(http.StatusNoContent)
}

func (h *ResourceManagementHandler) UpdateFolderDescription(ctx *gin.Context) {
	identity, ok := session.GetAdminIdentity(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req updateManagedFolderDescriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.service.UpdateFolderDescription(ctx.Request.Context(), ctx.Param("folderID"), UpdateManagedFolderDescriptionInput{
		Name:        req.Name,
		Description: req.Description,
		OperatorID:  identity.AdminID,
		OperatorIP:  ctx.ClientIP(),
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidResourceEdit):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder input"})
		case errors.Is(err, ErrManagedFolderConflict):
			ctx.JSON(http.StatusConflict, gin.H{"error": "folder name already exists"})
		case errors.Is(err, ErrManagedFolderNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update folder"})
		}
		return
	}
	h.notifySearchIndex("managed_folder_update")
	ctx.Status(http.StatusNoContent)
}

func (h *ResourceManagementHandler) DeleteFile(ctx *gin.Context) {
	identity, ok := session.GetAdminIdentity(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req deleteManagedResourceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.authService.VerifyPassword(ctx.Request.Context(), identity.AdminID, req.Password); err != nil {
		switch {
		case errors.Is(err, admin.ErrInvalidAdminCredentials):
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify password"})
		}
		return
	}

	err := h.service.DeleteFile(ctx.Request.Context(), ctx.Param("fileID"), identity.AdminID, ctx.ClientIP())
	if err != nil {
		switch {
		case errors.Is(err, ErrManagedFileNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete resource"})
		}
		return
	}
	h.notifySearchIndex("managed_file_delete")
	ctx.Status(http.StatusNoContent)
}

func (h *ResourceManagementHandler) DeleteFolder(ctx *gin.Context) {
	identity, ok := session.GetAdminIdentity(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req deleteManagedResourceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.authService.VerifyPassword(ctx.Request.Context(), identity.AdminID, req.Password); err != nil {
		switch {
		case errors.Is(err, admin.ErrInvalidAdminCredentials):
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify password"})
		}
		return
	}

	err := h.service.DeleteFolder(ctx.Request.Context(), ctx.Param("folderID"), identity.AdminID, ctx.ClientIP())
	if err != nil {
		switch {
		case errors.Is(err, ErrManagedFolderNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete folder"})
		}
		return
	}
	h.notifySearchIndex("managed_folder_delete")
	ctx.Status(http.StatusNoContent)
}

func (h *ResourceManagementHandler) notifySearchIndex(reason string) {
	if h.indexNotifier != nil {
		h.indexNotifier.NotifySearchResourcesChanged(reason)
	}
}
