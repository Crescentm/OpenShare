package admin

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"openshare/backend/internal/model"
	"openshare/backend/internal/session"
)

type AdminAuthHandler struct {
	authService    *AdminAuthService
	sessionManager *session.Manager
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

type adminProfileResponse struct {
	ID          string                  `json:"id"`
	Username    string                  `json:"username"`
	DisplayName string                  `json:"display_name"`
	AvatarURL   string                  `json:"avatar_url"`
	Role        string                  `json:"role"`
	Status      model.AdminStatus       `json:"status"`
	Permissions []model.AdminPermission `json:"permissions"`
}

type updateProfileRequest struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type loginResponse struct {
	Admin adminProfileResponse `json:"admin"`
}

func NewAdminAuthHandler(authService *AdminAuthService, sessionManager *session.Manager) *AdminAuthHandler {
	return &AdminAuthHandler{
		authService:    authService,
		sessionManager: sessionManager,
	}
}

func (h *AdminAuthHandler) Login(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	authenticated, err := h.authService.Login(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidAdminCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid username or password",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "login failed",
		})
		return
	}

	h.sessionManager.WriteCookie(ctx.Writer, authenticated.Cookie, authenticated.Identity.ExpiresAt)
	ctx.JSON(http.StatusOK, loginResponse{
		Admin: toAdminProfileResponse(authenticated.Admin),
	})
}

func (h *AdminAuthHandler) Logout(ctx *gin.Context) {
	cookieValue, err := ctx.Cookie(h.sessionManager.CookieName())
	if err == nil {
		if destroyErr := h.authService.Logout(ctx.Request.Context(), cookieValue); destroyErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "logout failed",
			})
			return
		}
	}

	h.sessionManager.ClearCookie(ctx.Writer)
	ctx.Status(http.StatusNoContent)
}

func (h *AdminAuthHandler) ChangePassword(ctx *gin.Context) {
	identity, ok := session.GetAdminIdentity(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var req changePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if err := h.authService.ChangePassword(ctx.Request.Context(), identity.AdminID, req.NewPassword, ctx.ClientIP()); err != nil {
		switch {
		case errors.Is(err, ErrInvalidAdminPasswordChange):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "new password must be at least 8 characters"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to change password"})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *AdminAuthHandler) Me(ctx *gin.Context) {
	identity, ok := session.GetAdminIdentity(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	admin, err := h.authService.GetProfile(ctx.Request.Context(), identity.AdminID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load account profile",
		})
		return
	}
	if admin == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	ctx.JSON(http.StatusOK, loginResponse{
		Admin: toAdminProfileResponse(admin),
	})
}

func (h *AdminAuthHandler) UpdateProfile(ctx *gin.Context) {
	identity, ok := session.GetAdminIdentity(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var req updateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	admin, err := h.authService.UpdateProfile(
		ctx.Request.Context(),
		identity.AdminID,
		req.DisplayName,
		req.AvatarURL,
		ctx.ClientIP(),
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidAdminProfileUpdate):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "display name must be 1-40 characters and avatar must be a valid image"})
		case errors.Is(err, ErrAdminDisplayNameTaken):
			ctx.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update account profile"})
		}
		return
	}

	ctx.JSON(http.StatusOK, loginResponse{
		Admin: toAdminProfileResponse(admin),
	})
}

func (h *AdminAuthHandler) PermissionProbe(permission model.AdminPermission) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		identity, _ := session.GetAdminIdentity(ctx)
		ctx.JSON(http.StatusOK, gin.H{
			"admin_id":    identity.AdminID,
			"permission":  permission,
			"authorized":  true,
			"super_admin": identity.IsSuperAdmin(),
		})
	}
}

func toAdminProfileResponse(admin *model.Admin) adminProfileResponse {
	displayName := strings.TrimSpace(admin.DisplayName)
	if displayName == "" {
		displayName = strings.TrimSpace(admin.Username)
	}
	return adminProfileResponse{
		ID:          admin.ID,
		Username:    strings.TrimSpace(admin.Username),
		DisplayName: displayName,
		AvatarURL:   strings.TrimSpace(admin.AvatarURL),
		Role:        admin.Role,
		Status:      admin.Status,
		Permissions: admin.PermissionList(),
	}
}
