package router

import (
	"gorm.io/gorm"

	"openshare/backend/internal/admin"
	"openshare/backend/internal/announcements"
	"openshare/backend/internal/catalog"
	"openshare/backend/internal/config"
	"openshare/backend/internal/downloads"
	"openshare/backend/internal/feedback"
	"openshare/backend/internal/imports"
	"openshare/backend/internal/moderation"
	"openshare/backend/internal/operations"
	"openshare/backend/internal/receipts"
	"openshare/backend/internal/resources"
	"openshare/backend/internal/search"
	"openshare/backend/internal/session"
	"openshare/backend/internal/settings"
	"openshare/backend/internal/storage"
	"openshare/backend/internal/submissions"
	"openshare/backend/internal/uploads"
	"openshare/backend/internal/visits"
)

func buildRouteHandlers(
	db *gorm.DB,
	cfg config.Config,
	sessionManager *session.Manager,
	importSyncNotifier imports.ManagedRootSyncNotifier,
) *routeHandlers {
	repos := buildRouteRepositories(db)
	services := buildRouteServices(db, cfg, repos, sessionManager)
	return buildHandlers(cfg, sessionManager, services, importSyncNotifier)
}

func buildRouteRepositories(db *gorm.DB) *routeRepositories {
	return &routeRepositories{
		admin:              admin.NewAdminRepository(db),
		adminDashboard:     admin.NewAdminDashboardRepository(db),
		announcement:       announcements.NewAnnouncementRepository(db),
		feedback:           feedback.NewFeedbackRepository(db),
		imports:            imports.NewImportRepository(db),
		moderation:         moderation.NewModerationRepository(db),
		operationLog:       operations.NewOperationLogRepository(db),
		publicCatalog:      catalog.NewPublicCatalogRepository(db),
		publicDownload:     downloads.NewPublicDownloadRepository(db),
		publicSubmission:   submissions.NewPublicSubmissionRepository(db),
		resourceManagement: resources.NewResourceManagementRepository(db),
		search:             search.NewSearchRepository(db),
		siteVisit:          visits.NewSiteVisitRepository(db),
		systemSetting:      settings.NewSystemSettingRepository(db),
		upload:             uploads.NewUploadRepository(db),
		receiptCode:        receipts.NewReceiptCodeRepository(db),
	}
}

func buildRouteServices(
	db *gorm.DB,
	cfg config.Config,
	repos *routeRepositories,
	sessionManager *session.Manager,
) *routeServices {
	storageService := storage.NewService(cfg.Storage)
	receiptCodeService := receipts.NewReceiptCodeService(repos.receiptCode, cfg.Upload.ReceiptCodeLength)
	systemSettingService := settings.NewSystemSettingService(repos.systemSetting, cfg)
	adminAuthService := admin.NewAdminAuthService(db, repos.admin, sessionManager)
	searchService := search.NewSearchService(repos.search)

	return &routeServices{
		adminAuth:          adminAuthService,
		adminDashboard:     admin.NewAdminDashboardService(repos.adminDashboard),
		announcement:       announcements.NewAnnouncementService(repos.announcement, repos.admin),
		adminManagement:    admin.NewAdminManagementService(repos.admin),
		feedback:           feedback.NewFeedbackService(repos.feedback, receiptCodeService),
		imports:            imports.NewImportService(repos.imports, storageService),
		moderation:         moderation.NewModerationService(repos.moderation, storageService),
		operationLog:       operations.NewOperationLogService(repos.operationLog),
		publicCatalog:      catalog.NewPublicCatalogService(repos.publicCatalog),
		publicDownload:     downloads.NewPublicDownloadService(repos.publicDownload, storageService, cfg.Download, systemSettingService),
		publicReceipt:      receiptCodeService,
		publicSubmission:   submissions.NewPublicSubmissionService(repos.publicSubmission),
		publicUpload:       uploads.NewPublicUploadService(cfg.Upload, repos.upload, receiptCodeService, storageService, systemSettingService),
		resourceManagement: resources.NewResourceManagementService(repos.resourceManagement, storageService),
		search:             searchService,
		siteVisit:          visits.NewSiteVisitService(repos.siteVisit),
		systemSetting:      systemSettingService,
	}
}

func buildHandlers(
	_ config.Config,
	sessionManager *session.Manager,
	services *routeServices,
	importSyncNotifier imports.ManagedRootSyncNotifier,
) *routeHandlers {
	return &routeHandlers{
		adminAuth:          admin.NewAdminAuthHandler(services.adminAuth, sessionManager),
		adminDashboard:     admin.NewAdminDashboardHandler(services.adminDashboard),
		announcement:       announcements.NewAnnouncementHandler(services.announcement),
		adminManagement:    admin.NewAdminManagementHandler(services.adminManagement, services.adminAuth),
		feedback:           feedback.NewFeedbackHandler(services.feedback),
		imports:            imports.NewImportHandler(services.imports, services.adminAuth, importSyncNotifier),
		moderation:         moderation.NewModerationHandler(services.moderation),
		operationLog:       operations.NewOperationLogHandler(services.operationLog),
		publicCatalog:      catalog.NewPublicCatalogHandler(services.publicCatalog),
		publicDownload:     downloads.NewPublicDownloadHandler(services.publicDownload),
		publicReceipt:      receipts.NewPublicReceiptHandler(services.publicReceipt),
		publicSubmission:   submissions.NewPublicSubmissionHandler(services.publicSubmission),
		publicUpload:       uploads.NewPublicUploadHandler(services.publicUpload),
		resourceManagement: resources.NewResourceManagementHandler(services.resourceManagement, services.adminAuth),
		search:             search.NewSearchHandler(services.search),
		siteVisit:          visits.NewSiteVisitHandler(services.siteVisit),
		systemSetting:      settings.NewSystemSettingHandler(services.systemSetting),
	}
}
