package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"openshare/backend/pkg/identity"
)

type TaskInput struct {
	WorkerName  string
	Topic       string
	Payload     string
	DedupeKey   *string
	AvailableAt *time.Time
}

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Enqueue(ctx context.Context, input TaskInput) error {
	now := time.Now().UTC()
	id, err := identity.NewID()
	if err != nil {
		return fmt.Errorf("generate worker task id: %w", err)
	}

	workerName := strings.TrimSpace(input.WorkerName)
	topic := strings.TrimSpace(input.Topic)
	if workerName == "" {
		return errors.New("worker task worker_name must not be empty")
	}
	if topic == "" {
		return errors.New("worker task topic must not be empty")
	}

	dedupeKey := normalizeOptionalString(input.DedupeKey)
	availableAt := input.AvailableAt
	if availableAt == nil {
		availableAt = &now
	} else {
		normalized := availableAt.UTC()
		availableAt = &normalized
	}

	task := &Task{
		ID:          id,
		WorkerName:  workerName,
		Topic:       topic,
		Payload:     strings.TrimSpace(input.Payload),
		DedupeKey:   dedupeKey,
		AvailableAt: availableAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := r.db.WithContext(ctx)
	if task.DedupeKey != nil {
		query = query.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "worker_name"},
				{Name: "dedupe_key"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"topic":        task.Topic,
				"payload":      task.Payload,
				"available_at": task.AvailableAt,
				"updated_at":   task.UpdatedAt,
			}),
		})
	}

	if err := query.Create(task).Error; err != nil {
		return fmt.Errorf("enqueue worker task %q for %q: %w", task.Topic, task.WorkerName, err)
	}
	return nil
}

func (r *TaskRepository) PopNextAvailableTask(ctx context.Context, workerName string) (*Task, error) {
	var task Task
	now := time.Now().UTC()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Where("worker_name = ?", strings.TrimSpace(workerName)).
			Where("(available_at IS NULL OR available_at <= ?)", now).
			Order("COALESCE(available_at, updated_at) ASC, updated_at ASC, created_at ASC").
			Limit(1).
			Find(&task)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Delete(&Task{}, "id = ?", task.ID).Error; err != nil {
			return fmt.Errorf("delete worker task %s: %w", task.ID, err)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("pop worker task for %q: %w", workerName, err)
	}
	return &task, nil
}

func (r *TaskRepository) CountPendingTasks(ctx context.Context, workerName string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&Task{})
	if strings.TrimSpace(workerName) != "" {
		query = query.Where("worker_name = ?", strings.TrimSpace(workerName))
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count worker tasks for %q: %w", workerName, err)
	}
	return count, nil
}

type HeartbeatRepository struct {
	db *gorm.DB
}

func NewHeartbeatRepository(db *gorm.DB) *HeartbeatRepository {
	return &HeartbeatRepository{db: db}
}

func (r *HeartbeatRepository) Upsert(
	ctx context.Context,
	workerName string,
	lastSeenAt time.Time,
	lastError string,
) error {
	heartbeat := &Heartbeat{
		WorkerName: strings.TrimSpace(workerName),
		LastSeenAt: lastSeenAt.UTC(),
		LastError:  strings.TrimSpace(lastError),
		UpdatedAt:  lastSeenAt.UTC(),
	}
	if heartbeat.WorkerName == "" {
		return errors.New("worker name must not be empty")
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "worker_name"}},
			DoUpdates: clause.Assignments(map[string]any{
				"last_seen_at": heartbeat.LastSeenAt,
				"last_error":   heartbeat.LastError,
				"updated_at":   heartbeat.UpdatedAt,
			}),
		}).
		Create(heartbeat).Error
	if err != nil {
		return fmt.Errorf("upsert worker heartbeat for %q: %w", heartbeat.WorkerName, err)
	}
	return nil
}

func (r *HeartbeatRepository) FindByWorkerName(
	ctx context.Context,
	workerName string,
) (*Heartbeat, error) {
	var heartbeat Heartbeat
	err := r.db.WithContext(ctx).
		Where("worker_name = ?", strings.TrimSpace(workerName)).
		Take(&heartbeat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find worker heartbeat %q: %w", workerName, err)
	}
	return &heartbeat, nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
