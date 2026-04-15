package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"openshare/backend/internal/admin"
	"openshare/backend/internal/bootstrap"
	"openshare/backend/internal/config"
	"openshare/backend/internal/router"
	"openshare/backend/internal/session"
	"openshare/backend/internal/storage"
	"openshare/backend/internal/worker"
	"openshare/backend/pkg/database"
)

func main() {
	cfg, err := config.Load("config/config.default.json", "config/config.local.json")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.NewSQLite(toDatabaseOptions(cfg.Database))
	if err != nil {
		log.Fatalf("init database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	if err := storage.EnsureLayout(cfg.Storage); err != nil {
		log.Fatalf("init storage layout: %v", err)
	}

	if err := bootstrap.EnsureSchema(db); err != nil {
		log.Fatalf("init schema: %v", err)
	}

	adminBootstrap := admin.NewAdminBootstrapService(db, admin.NewAdminRepository(db))
	if err := adminBootstrap.EnsureDefaultSuperAdmin(); err != nil {
		log.Fatalf("init default super admin: %v", err)
	}

	serverCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sessionManager := session.NewManager(db, cfg.Session, admin.NewAdminSessionRepository())
	managedSyncNotifier := worker.NewManagedSyncTaskNotifier(worker.NewTaskRepository(db))
	engine := router.New(db, cfg, sessionManager, managedSyncNotifier)
	server := &http.Server{
		Addr:    cfg.Server.Address(),
		Handler: engine,
	}

	log.Printf("OpenShare server listening on :%d", cfg.Server.Port)
	go func() {
		<-serverCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown server: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("run server: %v", err)
	}
}

func toDatabaseOptions(cfg config.DatabaseConfig) database.Options {
	pragmas := make([]database.Pragma, len(cfg.Pragmas))
	for i, p := range cfg.Pragmas {
		pragmas[i] = database.Pragma{Name: p.Name, Value: p.Value}
	}
	return database.Options{
		Path:      cfg.Path,
		LogLevel:  cfg.LogLevel,
		EnableWAL: cfg.EnableWAL,
		Pragmas:   pragmas,
	}
}
