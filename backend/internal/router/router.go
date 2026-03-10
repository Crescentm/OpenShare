package router

import (
	"github.com/gin-gonic/gin"
	"github.com/openshare/backend/internal/config"
	"github.com/openshare/backend/internal/handler"
	"github.com/openshare/backend/internal/middleware"
	"github.com/openshare/backend/internal/model"
	"github.com/openshare/backend/pkg/jwt"
	"github.com/openshare/backend/pkg/logger"
	"gorm.io/gorm"
)

// Options 路由初始化配置
type Options struct {
	Config     *config.Config
	Logger     *logger.Logger
	Handlers   *handler.Handlers
	JWTManager *jwt.Manager
	DB         *gorm.DB
}

// Setup 初始化路由
func Setup(opts *Options) *gin.Engine {
	// 设置运行模式
	gin.SetMode(opts.Config.Server.Mode)

	r := gin.New()

	// 全局中间件
	r.Use(middleware.Recovery(opts.Logger))
	r.Use(middleware.Logger(opts.Logger))
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/health", handler.Health)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 公开接口（无需认证）
		public := v1.Group("")
		{
			// 资料相关
			public.GET("/files", handler.NotImplemented)
			public.GET("/files/:id", handler.NotImplemented)
			public.GET("/files/:id/download", handler.NotImplemented)
			public.POST("/files/upload", handler.NotImplemented)

			// 搜索
			public.GET("/search", handler.NotImplemented)

			// 投稿查询
			public.GET("/submissions", handler.NotImplemented)

			// 公告
			public.GET("/announcements", handler.NotImplemented)

			// Tag
			public.GET("/tags", handler.NotImplemented)

			// 举报
			public.POST("/reports", handler.NotImplemented)
		}

		// 管理端接口
		admin := v1.Group("/admin")
		{
			// 认证接口（无需 token）
			admin.POST("/login", opts.Handlers.Admin.Login)

			// 基础认证接口（无需权限加载）
			basicAuth := admin.Group("")
			basicAuth.Use(middleware.Auth(opts.JWTManager))
			{
				// 当前用户相关（只需要认证，不需要特定权限）
				basicAuth.GET("/me", opts.Handlers.Admin.GetCurrentAdmin)
				basicAuth.POST("/password", opts.Handlers.Admin.ChangePassword)
				basicAuth.POST("/refresh", opts.Handlers.Admin.RefreshToken)
				basicAuth.POST("/logout", opts.Handlers.Admin.Logout)
			}

			// 需要权限验证的接口
			auth := admin.Group("")
			auth.Use(middleware.AuthWithPermissions(opts.JWTManager, opts.DB))
			{
				// 审核管理 - 需要审核权限
				submissions := auth.Group("/submissions")
				submissions.Use(middleware.RequirePermission(model.PermissionReviewSubmission))
				{
					submissions.GET("", handler.NotImplemented)
					submissions.POST("/:id/approve", handler.NotImplemented)
					submissions.POST("/:id/reject", handler.NotImplemented)
				}

				// 资料管理 - 需要对应权限
				files := auth.Group("/files")
				{
					files.GET("", handler.NotImplemented) // 查看列表无需特殊权限
					files.PUT("/:id", middleware.RequirePermission(model.PermissionEditFile), handler.NotImplemented)
					files.DELETE("/:id", middleware.RequirePermission(model.PermissionDeleteFile), handler.NotImplemented)
					files.POST("/:id/offline", middleware.RequirePermission(model.PermissionEditFile), handler.NotImplemented)
				}

				// Tag 管理 - 需要 Tag 管理权限
				tags := auth.Group("/tags")
				tags.Use(middleware.RequirePermission(model.PermissionManageTag))
				{
					tags.POST("", handler.NotImplemented)
					tags.PUT("/:id", handler.NotImplemented)
					tags.DELETE("/:id", handler.NotImplemented)
				}

				// 举报管理 - 需要举报管理权限
				reports := auth.Group("/reports")
				reports.Use(middleware.RequirePermission(model.PermissionManageReport))
				{
					reports.GET("", handler.NotImplemented)
					reports.POST("/:id/approve", handler.NotImplemented)
					reports.POST("/:id/reject", handler.NotImplemented)
				}

				// 公告管理 - 需要发布公告权限
				announcements := auth.Group("/announcements")
				announcements.Use(middleware.RequirePermission(model.PermissionPublishAnnounce))
				{
					announcements.POST("", handler.NotImplemented)
					announcements.PUT("/:id", handler.NotImplemented)
					announcements.DELETE("/:id", handler.NotImplemented)
				}

				// 管理员管理 - 仅超级管理员
				admins := auth.Group("/admins")
				admins.Use(middleware.RequireSuperAdmin())
				{
					admins.GET("", opts.Handlers.Admin.ListAdmins)
					admins.POST("", opts.Handlers.Admin.CreateAdmin)
					admins.GET("/:id", opts.Handlers.Admin.GetAdmin)
					admins.PUT("/:id", opts.Handlers.Admin.UpdateAdmin)
					admins.DELETE("/:id", opts.Handlers.Admin.DeleteAdmin)
					admins.PUT("/:id/permissions", opts.Handlers.Admin.SetAdminPermissions)
					admins.POST("/:id/reset-password", opts.Handlers.Admin.ResetAdminPassword)
				}

				// 权限元数据 - 无需特殊权限（便于前端展示）
				auth.GET("/permissions", opts.Handlers.Admin.GetAllPermissions)

				// 操作日志 - 需要查看日志权限
				auth.GET("/logs", middleware.RequirePermission(model.PermissionViewLog), handler.NotImplemented)

				// 系统配置 - 仅超级管理员
				settings := auth.Group("/settings")
				settings.Use(middleware.RequireSuperAdmin())
				{
					settings.GET("", handler.NotImplemented)
					settings.PUT("", handler.NotImplemented)
				}
			}
		}
	}

	return r
}
