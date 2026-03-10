package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/openshare/backend/internal/model"
)

// ============ 上下文键定义 ============
// 统一管理所有中间件使用的上下文键，避免硬编码和键名冲突

const (
	CtxKeyAdminID     = "admin_id"
	CtxKeyUsername    = "username"
	CtxKeyRole        = "role"
	CtxKeyToken       = "token"
	CtxKeyPermissions = "permissions"
	CtxKeyAdmin       = "admin" // 完整的 Admin 对象（可选加载）
)

// ============ 上下文帮助函数 ============
// 提供类型安全的上下文访问方法

// AuthContext 认证上下文，封装已认证用户的信息
type AuthContext struct {
	AdminID     uint
	Username    string
	Role        string
	Permissions []string
}

// GetAuthContext 从 Gin 上下文获取认证信息
// 返回 nil 表示未认证
func GetAuthContext(c *gin.Context) *AuthContext {
	adminID, exists := c.Get(CtxKeyAdminID)
	if !exists {
		return nil
	}

	id, ok := adminID.(uint)
	if !ok {
		return nil
	}

	ctx := &AuthContext{AdminID: id}

	if username, exists := c.Get(CtxKeyUsername); exists {
		if s, ok := username.(string); ok {
			ctx.Username = s
		}
	}

	if role, exists := c.Get(CtxKeyRole); exists {
		if s, ok := role.(string); ok {
			ctx.Role = s
		}
	}

	if perms, exists := c.Get(CtxKeyPermissions); exists {
		if p, ok := perms.([]string); ok {
			ctx.Permissions = p
		}
	}

	return ctx
}

// GetAdminID 从上下文获取管理员ID（便捷方法）
func GetAdminID(c *gin.Context) (uint, bool) {
	ctx := GetAuthContext(c)
	if ctx == nil {
		return 0, false
	}
	return ctx.AdminID, true
}

// GetRole 从上下文获取角色（便捷方法）
func GetRole(c *gin.Context) (string, bool) {
	ctx := GetAuthContext(c)
	if ctx == nil {
		return "", false
	}
	return ctx.Role, true
}

// GetUsername 从上下文获取用户名（便捷方法）
func GetUsername(c *gin.Context) (string, bool) {
	ctx := GetAuthContext(c)
	if ctx == nil {
		return "", false
	}
	return ctx.Username, true
}

// IsSuperAdmin 检查是否为超级管理员
func IsSuperAdmin(c *gin.Context) bool {
	role, ok := GetRole(c)
	return ok && role == model.RoleSuperAdmin
}

// IsAdmin 检查是否为管理员（包括超级管理员）
func IsAdmin(c *gin.Context) bool {
	role, ok := GetRole(c)
	return ok && (role == model.RoleAdmin || role == model.RoleSuperAdmin)
}

// HasPermission 检查当前用户是否拥有指定权限
func HasPermission(c *gin.Context, perm string) bool {
	// 超级管理员拥有所有权限
	if IsSuperAdmin(c) {
		return true
	}

	perms, exists := c.Get(CtxKeyPermissions)
	if !exists {
		return false
	}

	permissions, ok := perms.([]string)
	if !ok {
		return false
	}

	for _, p := range permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// HasAnyPermission 检查是否拥有任一权限
func HasAnyPermission(c *gin.Context, perms ...string) bool {
	if IsSuperAdmin(c) {
		return true
	}
	for _, p := range perms {
		if HasPermission(c, p) {
			return true
		}
	}
	return false
}

// HasAllPermissions 检查是否拥有所有权限
func HasAllPermissions(c *gin.Context, perms ...string) bool {
	if IsSuperAdmin(c) {
		return true
	}
	for _, p := range perms {
		if !HasPermission(c, p) {
			return false
		}
	}
	return true
}

// SetAuthContext 设置认证上下文（供中间件内部使用）
func SetAuthContext(c *gin.Context, adminID uint, username, role string, permissions []string) {
	c.Set(CtxKeyAdminID, adminID)
	c.Set(CtxKeyUsername, username)
	c.Set(CtxKeyRole, role)
	c.Set(CtxKeyPermissions, permissions)
}
