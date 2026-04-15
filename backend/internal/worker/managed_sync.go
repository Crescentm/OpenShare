package worker

import (
	"context"
	"log"
	"time"
)

const (
	ManagedSyncWorkerName            = "managed-sync-worker"
	ManagedSyncTaskTopicRootsChanged = "managed_roots_changed"
)

type ManagedSyncTaskNotifier struct {
	repository *TaskRepository
	timeout    time.Duration
}

func NewManagedSyncTaskNotifier(repository *TaskRepository) *ManagedSyncTaskNotifier {
	return &ManagedSyncTaskNotifier{
		repository: repository,
		timeout:    5 * time.Second,
	}
}

func (n *ManagedSyncTaskNotifier) NotifyManagedRootsChanged() {
	ctx, cancel := context.WithTimeout(context.Background(), n.timeout)
	defer cancel()

	dedupeKey := ManagedSyncTaskTopicRootsChanged
	if err := n.repository.Enqueue(ctx, TaskInput{
		WorkerName: ManagedSyncWorkerName,
		Topic:      ManagedSyncTaskTopicRootsChanged,
		DedupeKey:  &dedupeKey,
	}); err != nil {
		log.Printf("managed sync enqueue failed: %v", err)
	}
}
