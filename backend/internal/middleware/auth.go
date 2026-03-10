package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openshare/backend/internal/model"
	"github.com/openshare/backend/pkg/jwt"
	"github.com/openshare/backend/pkg/response"
	"gorm.io/gorm"
)

// ============ 认证中间件 ============

// AuthConfig 认证中间件配置
type AuthConfig struct {
	JWTManager *jwt.Manager
	DB         *gorm.DB // 可选：用于加载权限
}

// Auth JWT 认证中间件
// 职责：验证 JWT Token，将用户基本信息存入上下文
// 不加载权限，保持轻量；权限由 LoadPermissions 中间件按需加载
func Auth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractToken(c)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// 验证 JWT token
		claims, err := jwtManager.ParseToken(tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrExpiredToken) {
				response.Unauthorized(c, "token has expired")
			} else {
				response.Unauthorized(c, "invalid token")
			}
			c.Abort()
			return
		}

		// 将用户基本信息存入上下文
		c.Set(CtxKeyAdminID, claims.AdminID)
		c.Set(CtxKeyUsername, claims.Username)
		c.Set(CtxKeyRole, claims.Role)
		c.Set(CtxKeyToken, tokenString)

		c.Next()
	}
}

// AuthWithPermissions 带权限加载的认证中间件
// 职责：验证 JWT + 从数据库加载用户权限
// 适用于需要进行权限校验的路由组
func AuthWithPermissions(jwtManager *jwt.Manager, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractToken(c)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// 验证 JWT token
		claims, err := jwtManager.ParseToken(tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrExpiredToken) {
				response.Unauthorized(c, "token has expired")
			} else {
				response.Unauthorized(c, "invalid token")
			}
			c.Abort()
			return
		}

		// 存入基本信息
		c.Set(CtxKeyAdminID, claims.AdminID)
		c.Set(CtxKeyUsername, claims.Username)
		c.Set(CtxKeyRole, claims.Role)
		c.Set(CtxKeyToken, tokenString)

		// 超级管理员不需要加载权限（拥有全部权限）
		if claims.Role == model.RoleSuperAdmin {
			c.Set(CtxKeyPermissions, []string{})
			c.Next()
			return
		}

		// 从数据库加载权限
		permissions, err := loadPermissions(db, claims.AdminID)
		if err != nil {
			// 权限加载失败不阻断请求，但记录日志
			// 后续权限校验会失败
			c.Set(CtxKeyPermissions, []string{})
		} else {
			c.Set(CtxKeyPermissions, permissions)
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件
// 职责：如果有 Token 则验证，没有则跳过
// 适用于游客可访问但登录用户有额外功能的接口
func OptionalAuth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractToken(c)
		if err != nil {
			// 没有 token 或格式错误，作为游客继续
			c.Next()
			return
		}

		// 尝试验证 token
		claims, err := jwtManager.ParseToken(tokenString)
		if err != nil {
			// token 无效，作为游客继续
			c.Next()
			return
		}

		// token 有效，存入用户信息
		c.Set(CtxKeyAdminID, claims.AdminID)
		c.Set(CtxKeyUsername, claims.Username)
		c.Set(CtxKeyRole, claims.Role)
		c.Set(CtxKeyToken, tokenString)

		c.Next()
	}
}

// ============ 角色中间件 ============

// RequireRole 角色检查中间件
// 职责：检查用户是否属于指定角色之一
// 用法：RequireRole(model.RoleSuperAdmin) 或 RequireRole(model.RoleAdmin, model.RoleSuperAdmin)
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(CtxKeyRole)
		if !exists {
			response.Forbidden(c, "authentication required")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			response.Forbidden(c, "invalid authentication state")
			c.Abort()
			return
		}

		// 检查是否匹配任一角色
		for _, r := range roles {
			if r == roleStr {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "insufficient role privileges")
		c.Abort()
	}
}

// RequireSuperAdmin 要求超级管理员角色
// 便捷方法，等同于 RequireRole(model.RoleSuperAdmin)
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(model.RoleSuperAdmin)
}

// RequireAdmin 要求管理员角色（包括超级管理员）
// 便捷方法，等同于 RequireRole(model.RoleAdmin, model.RoleSuperAdmin)
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(model.RoleAdmin, model.RoleSuperAdmin)
}

// ============ 权限中间件 ============

// RequirePermission 权限检查中间件
// 职责：检查用户是否拥有指定权限
// 超级管理员自动通过所有权限检查
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasPermission(c, permission) {
			response.Forbidden(c, "permission denied: "+permission)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAnyPermission 要求拥有任一权限
// 用法：RequireAnyPermission(model.PermissionEditFile, model.PermissionDeleteFile)
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasAnyPermission(c, permissions...) {
			response.Forbidden(c, "permission denied")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAllPermissions 要求拥有全部权限
// 用法：RequireAllPermissions(model.PermissionEditFile, model.PermissionDeleteFile)
func RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasAllPermissions(c, permissions...) {
			response.Forbidden(c, "permission denied")
			c.Abort()
			return
		}
		c.Next()
	}
}

// ============ 账号状态检查 ============

// RequireActiveAccount 检查账号是否为活跃状态
// 职责：防止已禁用的账号继续使用旧 token 访问
// 注意：此中间件需要数据库查询，建议只在关键操作时使用
func RequireActiveAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, exists := GetAdminID(c)
		if !exists {
			response.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}

		var admin model.Admin
		if err := db.Select("status").First(&admin, adminID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Unauthorized(c, "account not found")
			} else {
				response.InternalError(c, "failed to verify account")
			}
			c.Abort()
			return
		}

		if admin.Status != model.AdminStatusActive {
			response.Forbidden(c, "account has been disabled")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ============ 工具函数 ============

// extractToken 从请求头提取 Bearer Token
func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization format")
	}

	token := parts[1]
	if token == "" {
		return "", errors.New("empty token")
	}

	return token, nil
}

// loadPermissions 从数据库加载用户权限
func loadPermissions(db *gorm.DB, adminID uint) ([]string, error) {
	var permissions []model.AdminPermission
	if err := db.Where("admin_id = ?", adminID).Find(&permissions).Error; err != nil {
		return nil, err
	}

	result := make([]string, len(permissions))
	for i, p := range permissions {
		result[i] = p.Permission
	}
	return result, nil
}
