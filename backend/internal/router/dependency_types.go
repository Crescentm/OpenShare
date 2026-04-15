package router

import (
	"openshare/backend/internal/admin"
	"openshare/backend/internal/announcements"
	"openshare/backend/internal/catalog"
	"openshare/backend/internal/downloads"
	"openshare/backend/internal/feedback"
	"openshare/backend/internal/imports"
	"openshare/backend/internal/moderation"
	"openshare/backend/internal/operations"
	"openshare/backend/internal/receipts"
	"openshare/backend/internal/resources"
	"openshare/backend/internal/search"
	"openshare/backend/internal/settings"
	"openshare/backend/internal/submissions"
	"openshare/backend/internal/uploads"
	"openshare/backend/internal/visits"
)

type routeHandlers struct {
	adminAuth          *admin.AdminAuthHandler
	adminDashboard     *admin.AdminDashboardHandler
	announcement       *announcements.AnnouncementHandler
	adminManagement    *admin.AdminManagementHandler
	feedback           *feedback.FeedbackHandler
	imports            *imports.ImportHandler
	moderation         *moderation.ModerationHandler
	operationLog       *operations.OperationLogHandler
	publicCatalog      *catalog.PublicCatalogHandler
	publicDownload     *downloads.PublicDownloadHandler
	publicReceipt      *receipts.PublicReceiptHandler
	publicSubmission   *submissions.PublicSubmissionHandler
	publicUpload       *uploads.PublicUploadHandler
	resourceManagement *resources.ResourceManagementHandler
	search             *search.SearchHandler
	siteVisit          *visits.SiteVisitHandler
	systemSetting      *settings.SystemSettingHandler
}

type routeRepositories struct {
	admin              *admin.AdminRepository
	adminDashboard     *admin.AdminDashboardRepository
	announcement       *announcements.AnnouncementRepository
	feedback           *feedback.FeedbackRepository
	imports            *imports.ImportRepository
	moderation         *moderation.ModerationRepository
	operationLog       *operations.OperationLogRepository
	publicCatalog      *catalog.PublicCatalogRepository
	publicDownload     *downloads.PublicDownloadRepository
	publicSubmission   *submissions.PublicSubmissionRepository
	resourceManagement *resources.ResourceManagementRepository
	search             *search.SearchRepository
	siteVisit          *visits.SiteVisitRepository
	systemSetting      *settings.SystemSettingRepository
	upload             *uploads.UploadRepository
	receiptCode        *receipts.ReceiptCodeRepository
}

type routeServices struct {
	adminAuth          *admin.AdminAuthService
	adminDashboard     *admin.AdminDashboardService
	announcement       *announcements.AnnouncementService
	adminManagement    *admin.AdminManagementService
	feedback           *feedback.FeedbackService
	imports            *imports.ImportService
	moderation         *moderation.ModerationService
	operationLog       *operations.OperationLogService
	publicCatalog      *catalog.PublicCatalogService
	publicDownload     *downloads.PublicDownloadService
	publicReceipt      *receipts.ReceiptCodeService
	publicSubmission   *submissions.PublicSubmissionService
	publicUpload       *uploads.PublicUploadService
	resourceManagement *resources.ResourceManagementService
	search             *search.SearchService
	siteVisit          *visits.SiteVisitService
	systemSetting      *settings.SystemSettingService
}
