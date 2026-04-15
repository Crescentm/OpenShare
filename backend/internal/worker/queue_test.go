package worker_test

import (
	"context"
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"openshare/backend/internal/bootstrap"
	"openshare/backend/internal/config"
	"openshare/backend/internal/storage"
	"openshare/backend/internal/worker"
	"openshare/backend/pkg/database"
)

func TestManagedSyncTaskNotifierDeduplicatesTopic(t *testing.T) {
	_, db := newWorkerTestDeps(t)
	taskRepository := worker.NewTaskRepository(db)
	notifier := worker.NewManagedSyncTaskNotifier(taskRepository)

	notifier.NotifyManagedRootsChanged()
	notifier.NotifyManagedRootsChanged()

	var tasks []worker.Task
	if err := db.Find(&tasks).Error; err != nil {
		t.Fatalf("list worker tasks failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected one deduplicated worker task, got %d", len(tasks))
	}
	if tasks[0].WorkerName != worker.ManagedSyncWorkerName {
		t.Fatalf("expected worker name %q, got %q", worker.ManagedSyncWorkerName, tasks[0].WorkerName)
	}
	if tasks[0].Topic != worker.ManagedSyncTaskTopicRootsChanged {
		t.Fatalf("expected topic %q, got %q", worker.ManagedSyncTaskTopicRootsChanged, tasks[0].Topic)
	}
}

func TestQueueDispatchesRegisteredHandler(t *testing.T) {
	_, db := newWorkerTestDeps(t)
	taskRepository := worker.NewTaskRepository(db)
	if err := taskRepository.Enqueue(context.Background(), worker.TaskInput{
		WorkerName: worker.ManagedSyncWorkerName,
		Topic:      worker.ManagedSyncTaskTopicRootsChanged,
	}); err != nil {
		t.Fatalf("enqueue worker task failed: %v", err)
	}

	calls := 0
	queue := worker.NewQueue(worker.ManagedSyncWorkerName, taskRepository, nil)
	if err := queue.RegisterHandler(worker.ManagedSyncTaskTopicRootsChanged, func(context.Context, worker.Task) error {
		calls++
		return nil
	}); err != nil {
		t.Fatalf("register handler failed: %v", err)
	}

	processed, err := queue.ProcessPending(context.Background())
	if err != nil {
		t.Fatalf("process pending worker tasks failed: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected 1 processed task, got %d", processed)
	}
	if calls != 1 {
		t.Fatalf("expected handler to be called once, got %d", calls)
	}

	var remaining int64
	if err := db.Model(&worker.Task{}).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining worker tasks failed: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected queue to be drained, got %d remaining rows", remaining)
	}
}

func newWorkerTestDeps(t *testing.T) (config.Config, *gorm.DB) {
	t.Helper()

	cfg := config.Default()
	cfg.Session.Secret = "test-secret"
	cfg.Storage.Root = filepath.Join(t.TempDir(), "storage")
	cfg.Database.Path = filepath.Join(t.TempDir(), "openshare-worker.db")

	if err := storage.EnsureLayout(cfg.Storage); err != nil {
		t.Fatalf("ensure storage layout failed: %v", err)
	}

	db, err := database.NewSQLite(database.Options{
		Path:      cfg.Database.Path,
		LogLevel:  "silent",
		EnableWAL: true,
		Pragmas: []database.Pragma{
			{Name: "foreign_keys", Value: "ON"},
			{Name: "busy_timeout", Value: "5000"},
		},
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := bootstrap.EnsureSchema(db); err != nil {
		t.Fatalf("ensure schema failed: %v", err)
	}

	return cfg, db
}
