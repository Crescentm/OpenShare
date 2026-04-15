package worker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	defaultHeartbeatInterval     = 10 * time.Second
	defaultHeartbeatStaleTimeout = 30 * time.Second
)

type HealthStatus struct {
	WorkerName     string     `json:"worker_name"`
	Status         string     `json:"status"`
	LastSeenAt     *time.Time `json:"last_seen_at,omitempty"`
	LastError      string     `json:"last_error,omitempty"`
	QueueBacklog   int64      `json:"queue_backlog"`
	TimeoutSeconds int64      `json:"timeout_seconds"`
}

type HeartbeatReporter struct {
	repository *HeartbeatRepository
	workerName string
	interval   time.Duration
	nowFunc    func() time.Time

	mu        sync.RWMutex
	lastError string
}

func NewHeartbeatReporter(repository *HeartbeatRepository, workerName string) *HeartbeatReporter {
	return &HeartbeatReporter{
		repository: repository,
		workerName: strings.TrimSpace(workerName),
		interval:   defaultHeartbeatInterval,
		nowFunc:    func() time.Time { return time.Now().UTC() },
	}
}

func (r *HeartbeatReporter) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		if err := r.flush(ctx); err != nil && ctx.Err() == nil {
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *HeartbeatReporter) ReportHealthy() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastError = ""
}

func (r *HeartbeatReporter) ReportError(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err == nil {
		r.lastError = ""
		return
	}
	r.lastError = strings.TrimSpace(err.Error())
}

func (r *HeartbeatReporter) flush(ctx context.Context) error {
	r.mu.RLock()
	lastError := r.lastError
	r.mu.RUnlock()

	return r.repository.Upsert(ctx, r.workerName, r.nowFunc(), lastError)
}

type HealthService struct {
	heartbeatRepository *HeartbeatRepository
	taskRepository      *TaskRepository
	workerName          string
	staleTimeout        time.Duration
	nowFunc             func() time.Time
}

func NewHealthService(
	heartbeatRepository *HeartbeatRepository,
	taskRepository *TaskRepository,
	workerName string,
) *HealthService {
	return &HealthService{
		heartbeatRepository: heartbeatRepository,
		taskRepository:      taskRepository,
		workerName:          strings.TrimSpace(workerName),
		staleTimeout:        defaultHeartbeatStaleTimeout,
		nowFunc:             func() time.Time { return time.Now().UTC() },
	}
}

func (s *HealthService) Status(ctx context.Context) (HealthStatus, int, error) {
	backlog, err := s.taskRepository.CountPendingTasks(ctx, s.workerName)
	if err != nil {
		return HealthStatus{}, 0, fmt.Errorf("count worker queue backlog: %w", err)
	}

	status := HealthStatus{
		WorkerName:     s.workerName,
		Status:         "error",
		QueueBacklog:   backlog,
		TimeoutSeconds: int64(s.staleTimeout / time.Second),
	}

	heartbeat, err := s.heartbeatRepository.FindByWorkerName(ctx, s.workerName)
	if err != nil {
		return HealthStatus{}, 0, fmt.Errorf("load worker heartbeat: %w", err)
	}
	if heartbeat == nil {
		status.LastError = "worker heartbeat not observed yet"
		return status, 503, nil
	}

	lastSeenAt := heartbeat.LastSeenAt.UTC()
	status.LastSeenAt = &lastSeenAt
	status.LastError = strings.TrimSpace(heartbeat.LastError)

	if s.nowFunc().Sub(lastSeenAt) > s.staleTimeout {
		if status.LastError == "" {
			status.LastError = "worker heartbeat timed out"
		}
		return status, 503, nil
	}

	if status.LastError != "" {
		return status, 503, nil
	}

	status.Status = "ok"
	return status, 200, nil
}
