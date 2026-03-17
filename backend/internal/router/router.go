package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"openshare/backend/internal/config"
	"openshare/backend/internal/handler"
	"openshare/backend/internal/middleware"
	"openshare/backend/internal/model"
	"openshare/backend/internal/repository"
	"openshare/backend/internal/service"
	"openshare/backend/internal/session"
	"openshare/backend/internal/storage"
)

func New(db *gorm.DB, cfg config.Config, sessionManager *session.Manager) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())
	engine.Use(middleware.SessionLoader(sessionManager))

	storageService := storage.NewService(cfg.Storage)
	receiptCodeService := service.NewReceiptCodeService(repository.NewReceiptCodeRepository(db), cfg.Upload.ReceiptCodeLength)
	adminRepo := repository.NewAdminRepository(db)
	systemSettingService := service.NewSystemSettingService(repository.NewSystemSettingRepository(db), cfg)
	adminAuthService := service.NewAdminAuthService(db, adminRepo, sessionManager)
	adminAuthHandler := handler.NewAdminAuthHandler(adminAuthService, sessionManager)
	adminDashboardHandler := handler.NewAdminDashboardHandler(
		service.NewAdminDashboardService(repository.NewAdminDashboardRepository(db)),
	)

	searchRepo := repository.NewSearchRepository(db)
	searchService := service.NewSearchService(searchRepo, systemSettingService)
	_ = searchService.RebuildAllIndexes(context.Background())
	searchHandler := handler.NewSearchHandler(searchService)
	announcementHandler := handler.NewAnnouncementHandler(
		service.NewAnnouncementService(repository.NewAnnouncementRepository(db), adminRepo),
	)
	adminManagementHandler := handler.NewAdminManagementHandler(
		service.NewAdminManagementService(adminRepo),
	)
	operationLogHandler := handler.NewOperationLogHandler(
		service.NewOperationLogService(repository.NewOperationLogRepository(db)),
	)

	importHandler := handler.NewImportHandler(
		service.NewImportService(repository.NewImportRepository(db), storageService, searchService),
		adminAuthService,
	)

	moderationHandler := handler.NewModerationHandler(
		service.NewModerationService(repository.NewModerationRepository(db), storageService, searchService),
	)
	resourceManagementHandler := handler.NewResourceManagementHandler(
		service.NewResourceManagementServiceWithSettings(repository.NewResourceManagementRepository(db), storageService, systemSettingService, searchService),
		adminAuthService,
	)
	systemSettingHandler := handler.NewSystemSettingHandler(
		systemSettingService,
	)
	publicCatalogHandler := handler.NewPublicCatalogHandler(
		service.NewPublicCatalogService(repository.NewPublicCatalogRepository(db)),
	)
	publicDownloadHandler := handler.NewPublicDownloadHandler(
		service.NewPublicDownloadService(repository.NewPublicDownloadRepository(db), storageService),
	)
	publicSubmissionHandler := handler.NewPublicSubmissionHandler(
		service.NewPublicSubmissionService(repository.NewPublicSubmissionRepository(db)),
	)
	publicUploadHandler := handler.NewPublicUploadHandler(
		service.NewPublicUploadService(
			cfg.Upload,
			repository.NewUploadRepository(db),
			receiptCodeService,
			storageService,
			systemSettingService,
		),
		systemSettingService,
		cfg.Upload.MaxBatchTotalSizeBytes+(1<<20),
	)

	reportRepo := repository.NewReportRepository(db)
	reportService := service.NewReportService(reportRepo, receiptCodeService, searchService, storageService)
	reportHandler := handler.NewReportHandler(reportService)
	siteVisitHandler := handler.NewSiteVisitHandler(
		service.NewSiteVisitService(repository.NewSiteVisitRepository(db)),
	)

	engine.GET("/healthz", func(ctx *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "database handle is unavailable",
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "error",
				"error":  "database ping failed",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := engine.Group("/api")
	api.POST("/visits", siteVisitHandler.Record)
	public := api.Group("/public")
	public.GET("/files", publicCatalogHandler.ListPublicFiles)
	public.POST("/files/batch-download", publicDownloadHandler.DownloadBatch)
	public.GET("/files/:fileID", publicDownloadHandler.GetFileDetail)
	public.PUT("/files/:fileID", resourceManagementHandler.PublicUpdateFile)
	public.DELETE("/files/:fileID", resourceManagementHandler.PublicDeleteFile)
	public.GET("/files/:fileID/download", publicDownloadHandler.DownloadFile)
	public.GET("/folders", publicCatalogHandler.ListPublicFolders)
	public.GET("/folders/:folderID", publicCatalogHandler.GetPublicFolderDetail)
	public.GET("/folders/:folderID/download", publicDownloadHandler.DownloadFolder)
	public.GET("/announcements", announcementHandler.ListPublic)
	public.GET("/system/policy", systemSettingHandler.GetPublicPolicy)
	public.GET("/search", searchHandler.Search)
	public.POST("/submissions", publicUploadHandler.CreateSubmission)
	public.GET("/submissions/:receiptCode", publicSubmissionHandler.LookupByReceiptCode)
	public.POST("/reports", reportHandler.CreateReport)
	public.GET("/reports/:reportID", reportHandler.LookupReport)

	admin := api.Group("/admin")
	admin.POST("/session/login", adminAuthHandler.Login)
	admin.POST("/session/logout", adminAuthHandler.Logout)

	adminProtected := admin.Group("")
	adminProtected.Use(middleware.RequireAdminAuth())
	adminProtected.GET("/me", adminAuthHandler.Me)
	adminProtected.GET("/dashboard/stats", adminDashboardHandler.GetStats)
	adminProtected.POST("/session/change-password", adminAuthHandler.ChangePassword)
	adminProtected.PATCH("/account/profile", adminAuthHandler.UpdateProfile)
	adminProtected.GET("/operation-logs", operationLogHandler.List)
	adminProtected.GET(
		"/announcements",
		middleware.RequireAdminPermission(model.AdminPermissionAnnouncements),
		announcementHandler.ListAdmin,
	)
	adminProtected.POST(
		"/announcements",
		middleware.RequireAdminPermission(model.AdminPermissionAnnouncements),
		announcementHandler.Create,
	)
	adminProtected.PUT(
		"/announcements/:announcementID",
		middleware.RequireAdminPermission(model.AdminPermissionAnnouncements),
		announcementHandler.Update,
	)
	adminProtected.DELETE(
		"/announcements/:announcementID",
		middleware.RequireAdminPermission(model.AdminPermissionAnnouncements),
		announcementHandler.Delete,
	)
	adminProtected.GET(
		"/submissions/pending",
		middleware.RequireAdminPermission(model.AdminPermissionSubmissionModeration),
		moderationHandler.ListPendingSubmissions,
	)
	adminProtected.POST(
		"/submissions/:submissionID/approve",
		middleware.RequireAdminPermission(model.AdminPermissionSubmissionModeration),
		moderationHandler.ApproveSubmission,
	)
	adminProtected.POST(
		"/submissions/:submissionID/reject",
		middleware.RequireAdminPermission(model.AdminPermissionSubmissionModeration),
		moderationHandler.RejectSubmission,
	)
	adminProtected.POST(
		"/imports/local",
		middleware.RequireAdminPermission(model.AdminPermissionManageSystem),
		importHandler.ImportLocalDirectory,
	)
	adminProtected.GET(
		"/imports/directories",
		middleware.RequireAdminPermission(model.AdminPermissionManageSystem),
		importHandler.ListDirectories,
	)
	adminProtected.DELETE(
		"/imports/local/:folderID",
		middleware.RequireSuperAdmin(),
		importHandler.DeleteManagedDirectory,
	)
	adminProtected.POST(
		"/search/rebuild-index",
		middleware.RequireAdminPermission(model.AdminPermissionManageSystem),
		searchHandler.RebuildIndex,
	)
	adminProtected.GET(
		"/folders/tree",
		importHandler.GetFolderTree,
	)
	adminProtected.GET(
		"/resources/files",
		resourceManagementHandler.ListFiles,
	)
	adminProtected.PUT(
		"/resources/folders/:folderID",
		middleware.RequireAdminPermission(model.AdminPermissionResourceModeration),
		resourceManagementHandler.UpdateFolderDescription,
	)
	adminProtected.PUT(
		"/resources/files/:fileID",
		middleware.RequireAdminPermission(model.AdminPermissionResourceModeration),
		resourceManagementHandler.UpdateFile,
	)
	adminProtected.POST(
		"/resources/files/:fileID/offline",
		middleware.RequireAdminPermission(model.AdminPermissionResourceModeration),
		resourceManagementHandler.OfflineFile,
	)
	adminProtected.DELETE(
		"/resources/files/:fileID",
		middleware.RequireAdminPermission(model.AdminPermissionResourceModeration),
		resourceManagementHandler.DeleteFile,
	)
	adminProtected.DELETE(
		"/resources/folders/:folderID",
		middleware.RequireAdminPermission(model.AdminPermissionResourceModeration),
		resourceManagementHandler.DeleteFolder,
	)
	// Report management routes
	adminProtected.GET(
		"/reports/pending",
		middleware.RequireAdminPermission(model.AdminPermissionResourceModeration),
		reportHandler.ListPendingReports,
	)
	adminProtected.POST(
		"/reports/:reportID/approve",
		middleware.RequireAdminPermission(model.AdminPermissionResourceModeration),
		reportHandler.ApproveReport,
	)
	adminProtected.POST(
		"/reports/:reportID/reject",
		middleware.RequireAdminPermission(model.AdminPermissionResourceModeration),
		reportHandler.RejectReport,
	)

	adminProtected.GET(
		"/admins",
		adminManagementHandler.ListAdmins,
	)
	adminProtected.POST(
		"/admins",
		middleware.RequireAdminPermission(model.AdminPermissionManageAdmins),
		adminManagementHandler.CreateAdmin,
	)
	adminProtected.PUT(
		"/admins/:adminID",
		middleware.RequireAdminPermission(model.AdminPermissionManageAdmins),
		adminManagementHandler.UpdateAdmin,
	)
	adminProtected.POST(
		"/admins/:adminID/reset-password",
		middleware.RequireAdminPermission(model.AdminPermissionManageAdmins),
		adminManagementHandler.ResetPassword,
	)
	adminProtected.DELETE(
		"/admins/:adminID",
		middleware.RequireAdminPermission(model.AdminPermissionManageAdmins),
		adminManagementHandler.DeleteAdmin,
	)
	superAdminOnly := adminProtected.Group("")
	superAdminOnly.Use(middleware.RequireSuperAdmin())
	superAdminOnly.GET("/system/settings", systemSettingHandler.GetPolicy)
	superAdminOnly.PUT("/system/settings", systemSettingHandler.SavePolicy)

	adminPermissionProbe := adminProtected.Group("/_internal")
	adminPermissionProbe.GET(
		"/review",
		middleware.RequireAdminPermission(model.AdminPermissionSubmissionModeration),
		adminAuthHandler.PermissionProbe(model.AdminPermissionSubmissionModeration),
	)
	adminPermissionProbe.GET(
		"/system",
		middleware.RequireAdminPermission(model.AdminPermissionManageSystem),
		adminAuthHandler.PermissionProbe(model.AdminPermissionManageSystem),
	)

	return engine
}
