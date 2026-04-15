package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"openshare/backend/internal/config"
	"openshare/backend/internal/imports"
	"openshare/backend/internal/middleware"
	"openshare/backend/internal/session"
	"openshare/backend/internal/worker"
	webui "openshare/backend/web"
)

func New(
	db *gorm.DB,
	cfg config.Config,
	sessionManager *session.Manager,
	importSyncNotifiers ...imports.ManagedRootSyncNotifier,
) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())
	engine.Use(middleware.SessionLoader(sessionManager))

	var importSyncNotifier imports.ManagedRootSyncNotifier
	if len(importSyncNotifiers) > 0 {
		importSyncNotifier = importSyncNotifiers[0]
	}

	handlers := buildRouteHandlers(db, cfg, sessionManager, importSyncNotifier)
	workerHealthService := worker.NewHealthService(
		worker.NewHeartbeatRepository(db),
		worker.NewTaskRepository(db),
		worker.ManagedSyncWorkerName,
	)
	registerHealthRoutes(engine, func(ctx *gin.Context) {
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
	}, func(ctx *gin.Context) {
		status, code, err := workerHealthService.Status(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "failed to load worker health",
			})
			return
		}
		ctx.JSON(code, status)
	})

	api := engine.Group("/api")
	registerPublicRoutes(api, handlers)
	registerAdminRoutes(api, handlers)
	webui.Register(engine)

	return engine
}
