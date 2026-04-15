package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"openshare/backend/internal/bootstrap"
	"openshare/backend/internal/config"
	"openshare/backend/internal/imports"
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

	storageService := storage.NewService(cfg.Storage)
	workerHeartbeatRepository := worker.NewHeartbeatRepository(db)
	workerHeartbeatReporter := worker.NewHeartbeatReporter(workerHeartbeatRepository, worker.ManagedSyncWorkerName)
	importService := imports.NewImportService(imports.NewImportRepository(db), storageService)
	syncManager := imports.NewImportSyncManager(
		importService,
		imports.WithImportSyncRefreshInterval(time.Duration(cfg.ManagedSync.RefreshIntervalSeconds)*time.Second),
		imports.WithImportSyncAuditInterval(time.Duration(cfg.ManagedSync.AuditIntervalSeconds)*time.Second),
	)
	queueWorker := worker.NewQueue(
		worker.ManagedSyncWorkerName,
		worker.NewTaskRepository(db),
		workerHeartbeatReporter,
	)
	if err := queueWorker.RegisterHandler(worker.ManagedSyncTaskTopicRootsChanged, func(context.Context, worker.Task) error {
		syncManager.NotifyManagedRootsChanged()
		return nil
	}); err != nil {
		log.Fatalf("register queue handler: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := syncManager.Start(ctx); err != nil {
		log.Fatalf("start import sync manager: %v", err)
	}
	go workerHeartbeatReporter.Run(ctx)
	go queueWorker.Run(ctx)

	log.Print("OpenShare sync worker started")
	<-ctx.Done()
	log.Print("OpenShare sync worker stopped")
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
