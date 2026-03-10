package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/openshare/backend/internal/middleware"
	"github.com/openshare/backend/internal/model"
	"github.com/openshare/backend/internal/service"
	"github.com/openshare/backend/pkg/jwt"
	"github.com/openshare/backend/pkg/logger"
	"github.com/openshare/backend/pkg/response"
)

// AdminHandler 管理员相关接口
type AdminHandler struct {
	adminService *service.AdminService
	jwtManager   *jwt.Manager
	logger       *logger.Logger
}

// NewAdminHandler 创建管理员 handler
func NewAdminHandler(opts *Options) *AdminHandler {
	return &AdminHandler{
		adminService: opts.Services.Admin,
		jwtManager:   opts.JWTManager,
		logger:       opts.Logger,
	}
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=1,max=50"`
	Password string `json:"password" binding:"required,min=1,max=100"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt int64     `json:"expires_at"` // Unix 时间戳
	Admin     AdminInfo `json:"admin"`
}

// AdminInfo 管理员基本信息
type AdminInfo struct {
	ID          uint     `json:"id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions,omitempty"`
}

// Login 管理员登录
// POST /api/v1/admin/login
func (h *AdminHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	// 验证凭证
	admin, err := h.adminService.ValidateCredentials(req.Username, req.Password)
	if err != nil {
		h.logger.Error("Login validation error", "error", err, "username", req.Username)
		response.InternalError(c, "internal server error")
		return
	}

	if admin == nil {
		// 登录失败：用户不存在、密码错误或账号已禁用
		// 使用统一的错误信息避免泄露账号存在性
		h.logger.Info("Login failed", "username", req.Username, "ip", c.ClientIP())
		response.Unauthorized(c, "invalid username or password")
		return
	}

	// 生成 JWT Token
	token, expiresAt, err := h.jwtManager.GenerateToken(admin.ID, admin.Username, admin.Role)
	if err != nil {
		h.logger.Error("Failed to generate token", "error", err, "admin_id", admin.ID)
		response.InternalError(c, "failed to generate token")
		return
	}

	// 更新最后登录时间
	if err := h.adminService.UpdateLastLogin(admin.ID); err != nil {
		// 登录时间更新失败不影响主流程
		h.logger.Warn("Failed to update last login time", "error", err, "admin_id", admin.ID)
	}

	// 获取权限列表
	var permissions []string
	for _, p := range admin.Permissions {
		permissions = append(permissions, p.Permission)
	}

	h.logger.Info("Admin logged in",
		"admin_id", admin.ID,
		"username", admin.Username,
		"role", admin.Role,
		"ip", c.ClientIP(),
	)

	response.Success(c, LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
		Admin: AdminInfo{
			ID:          admin.ID,
			Username:    admin.Username,
			Role:        admin.Role,
			Permissions: permissions,
		},
	})
}

// GetCurrentAdmin 获取当前登录管理员信息
// GET /api/v1/admin/me
func (h *AdminHandler) GetCurrentAdmin(c *gin.Context) {
	adminID, exists := middleware.GetAdminID(c)
	if !exists {
		response.Unauthorized(c, "not logged in")
		return
	}

	admin, err := h.adminService.GetByID(adminID)
	if err != nil {
		h.logger.Error("Failed to get admin", "error", err, "admin_id", adminID)
		response.InternalError(c, "internal server error")
		return
	}

	if admin == nil {
		response.NotFound(c, "admin not found")
		return
	}

	var permissions []string
	for _, p := range admin.Permissions {
		permissions = append(permissions, p.Permission)
	}

	response.Success(c, AdminInfo{
		ID:          admin.ID,
		Username:    admin.Username,
		Role:        admin.Role,
		Permissions: permissions,
	})
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=1"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

// ChangePassword 修改当前管理员密码
// POST /api/v1/admin/password
func (h *AdminHandler) ChangePassword(c *gin.Context) {
	adminID, exists := middleware.GetAdminID(c)
	if !exists {
		response.Unauthorized(c, "not logged in")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	if err := h.adminService.ChangePassword(adminID, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "invalid old password" {
			response.BadRequest(c, "invalid old password")
			return
		}
		h.logger.Error("Failed to change password", "error", err, "admin_id", adminID)
		response.InternalError(c, "failed to change password")
		return
	}

	h.logger.Info("Password changed", "admin_id", adminID, "ip", c.ClientIP())
	response.Success(c, nil)
}

// RefreshToken 刷新 Token
// POST /api/v1/admin/refresh
func (h *AdminHandler) RefreshToken(c *gin.Context) {
	// 从 header 获取当前 token
	tokenString := c.GetString("token")
	if tokenString == "" {
		response.Unauthorized(c, "no token provided")
		return
	}

	// 刷新 token
	newToken, expiresAt, err := h.jwtManager.RefreshToken(tokenString)
	if err != nil {
		h.logger.Warn("Failed to refresh token", "error", err)
		response.Unauthorized(c, "invalid or expired token")
		return
	}

	response.Success(c, gin.H{
		"token":      newToken,
		"expires_at": expiresAt.Unix(),
	})
}

// Logout 退出登录（前端清除 token 即可，后端记录日志）
// POST /api/v1/admin/logout
func (h *AdminHandler) Logout(c *gin.Context) {
	adminID, _ := middleware.GetAdminID(c)
	h.logger.Info("Admin logged out", "admin_id", adminID, "ip", c.ClientIP())
	response.Success(c, nil)
}

// ============ 管理员账号管理（仅超级管理员） ============

// ListAdminsRequest 获取管理员列表请求参数
type ListAdminsRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Role     string `form:"role"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
}

// ListAdminsResponse 管理员列表响应
type ListAdminsResponse struct {
	Admins []AdminDetail `json:"admins"`
	Total  int64         `json:"total"`
	Page   int           `json:"page"`
}

// AdminDetail 管理员详细信息
type AdminDetail struct {
	ID          uint     `json:"id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Status      string   `json:"status"`
	Permissions []string `json:"permissions"`
	LastLogin   *int64   `json:"last_login,omitempty"` // Unix 时间戳
	CreatedAt   int64    `json:"created_at"`
}

// toAdminDetail 转换为 AdminDetail
func toAdminDetail(admin *model.Admin) AdminDetail {
	detail := AdminDetail{
		ID:          admin.ID,
		Username:    admin.Username,
		Role:        admin.Role,
		Status:      admin.Status,
		Permissions: admin.GetPermissionCodes(),
		CreatedAt:   admin.CreatedAt.Unix(),
	}
	if admin.LastLogin != nil {
		ts := admin.LastLogin.Unix()
		detail.LastLogin = &ts
	}
	return detail
}

// ListAdmins 获取管理员列表
// GET /api/v1/admin/admins
func (h *AdminHandler) ListAdmins(c *gin.Context) {
	var req ListAdminsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

	result, err := h.adminService.ListAdmins(&service.ListAdminsInput{
		Page:     req.Page,
		PageSize: req.PageSize,
		Role:     req.Role,
		Status:   req.Status,
		Keyword:  req.Keyword,
	})
	if err != nil {
		h.logger.Error("Failed to list admins", "error", err)
		response.InternalError(c, "failed to list admins")
		return
	}

	admins := make([]AdminDetail, 0, len(result.Admins))
	for _, admin := range result.Admins {
		admins = append(admins, toAdminDetail(admin))
	}

	response.Success(c, ListAdminsResponse{
		Admins: admins,
		Total:  result.Total,
		Page:   req.Page,
	})
}

// CreateAdminRequest 创建管理员请求
type CreateAdminRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
}

// CreateAdminResponse 创建管理员响应
type CreateAdminResponse struct {
	Admin    AdminDetail `json:"admin"`
	Password string      `json:"password"` // 初始密码，仅创建时返回
}

// CreateAdmin 创建普通管理员
// POST /api/v1/admin/admins
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	operatorID, _ := middleware.GetAdminID(c)

	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	result, err := h.adminService.CreateAdmin(&service.CreateAdminInput{
		Username: req.Username,
	})
	if err != nil {
		if err.Error() == "username already exists" {
			response.Conflict(c, "username already exists")
			return
		}
		h.logger.Error("Failed to create admin", "error", err, "operator_id", operatorID)
		response.InternalError(c, err.Error())
		return
	}

	h.logger.Info("Admin created",
		"operator_id", operatorID,
		"new_admin_id", result.Admin.ID,
		"username", result.Admin.Username,
	)

	response.Success(c, CreateAdminResponse{
		Admin:    toAdminDetail(result.Admin),
		Password: result.Password,
	})
}

// GetAdmin 获取管理员详情
// GET /api/v1/admin/admins/:id
func (h *AdminHandler) GetAdmin(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid admin id")
		return
	}

	admin, err := h.adminService.GetByID(id)
	if err != nil {
		h.logger.Error("Failed to get admin", "error", err, "admin_id", id)
		response.InternalError(c, "failed to get admin")
		return
	}

	if admin == nil {
		response.NotFound(c, "admin not found")
		return
	}

	response.Success(c, toAdminDetail(admin))
}

// UpdateAdminRequest 更新管理员请求
type UpdateAdminRequest struct {
	Username *string `json:"username,omitempty"`
	Status   *string `json:"status,omitempty"`
}

// UpdateAdmin 更新管理员信息
// PUT /api/v1/admin/admins/:id
func (h *AdminHandler) UpdateAdmin(c *gin.Context) {
	operatorID, _ := middleware.GetAdminID(c)

	id, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid admin id")
		return
	}

	var req UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	admin, err := h.adminService.UpdateAdmin(&service.UpdateAdminInput{
		ID:       id,
		Username: req.Username,
		Status:   req.Status,
	})
	if err != nil {
		switch err.Error() {
		case "admin not found":
			response.NotFound(c, err.Error())
		case "cannot modify super admin":
			response.Forbidden(c, err.Error())
		case "username already exists":
			response.Conflict(c, err.Error())
		default:
			h.logger.Error("Failed to update admin", "error", err, "admin_id", id)
			response.InternalError(c, err.Error())
		}
		return
	}

	h.logger.Info("Admin updated",
		"operator_id", operatorID,
		"admin_id", id,
	)

	response.Success(c, toAdminDetail(admin))
}

// DeleteAdmin 删除管理员（软删除）
// DELETE /api/v1/admin/admins/:id
func (h *AdminHandler) DeleteAdmin(c *gin.Context) {
	operatorID, _ := middleware.GetAdminID(c)

	id, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid admin id")
		return
	}

	// 不能删除自己
	if id == operatorID {
		response.BadRequest(c, "cannot delete yourself")
		return
	}

	if err := h.adminService.DeleteAdmin(id); err != nil {
		switch err.Error() {
		case "admin not found":
			response.NotFound(c, err.Error())
		case "cannot delete super admin":
			response.Forbidden(c, err.Error())
		default:
			h.logger.Error("Failed to delete admin", "error", err, "admin_id", id)
			response.InternalError(c, "failed to delete admin")
		}
		return
	}

	h.logger.Info("Admin deleted",
		"operator_id", operatorID,
		"deleted_admin_id", id,
	)

	response.Success(c, nil)
}

// SetPermissionsRequest 设置权限请求
type SetPermissionsRequest struct {
	Permissions []string `json:"permissions" binding:"required"`
}

// SetAdminPermissions 设置管理员权限（整体替换）
// PUT /api/v1/admin/admins/:id/permissions
func (h *AdminHandler) SetAdminPermissions(c *gin.Context) {
	operatorID, _ := middleware.GetAdminID(c)

	id, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid admin id")
		return
	}

	var req SetPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	if err := h.adminService.SetPermissions(id, req.Permissions); err != nil {
		switch err.Error() {
		case "admin not found":
			response.NotFound(c, err.Error())
		case "super admin has all permissions by default":
			response.BadRequest(c, err.Error())
		default:
			if len(err.Error()) > 20 && err.Error()[:20] == "invalid permission: " {
				response.BadRequest(c, err.Error())
				return
			}
			h.logger.Error("Failed to set permissions", "error", err, "admin_id", id)
			response.InternalError(c, "failed to set permissions")
		}
		return
	}

	h.logger.Info("Admin permissions updated",
		"operator_id", operatorID,
		"admin_id", id,
		"permissions", req.Permissions,
	)

	// 返回更新后的管理员信息
	admin, _ := h.adminService.GetByID(id)
	if admin != nil {
		response.Success(c, toAdminDetail(admin))
	} else {
		response.Success(c, nil)
	}
}

// ResetAdminPassword 重置管理员密码
// POST /api/v1/admin/admins/:id/reset-password
func (h *AdminHandler) ResetAdminPassword(c *gin.Context) {
	operatorID, _ := middleware.GetAdminID(c)

	id, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid admin id")
		return
	}

	newPassword, err := h.adminService.ResetPassword(id)
	if err != nil {
		switch err.Error() {
		case "admin not found":
			response.NotFound(c, err.Error())
		case "cannot reset super admin password":
			response.Forbidden(c, err.Error())
		default:
			h.logger.Error("Failed to reset password", "error", err, "admin_id", id)
			response.InternalError(c, "failed to reset password")
		}
		return
	}

	h.logger.Info("Admin password reset",
		"operator_id", operatorID,
		"admin_id", id,
	)

	response.Success(c, gin.H{
		"password": newPassword,
	})
}

// ============ 权限元数据 ============

// GetAllPermissions 获取所有可用权限列表
// GET /api/v1/admin/permissions
func (h *AdminHandler) GetAllPermissions(c *gin.Context) {
	groups := model.GetPermissionGroups()
	response.Success(c, groups)
}

// ============ 工具函数 ============

// parseUintParam 解析路由参数为 uint
func parseUintParam(c *gin.Context, name string) (uint, error) {
	param := c.Param(name)
	var id uint
	_, err := fmt.Sscanf(param, "%d", &id)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return id, nil
}
