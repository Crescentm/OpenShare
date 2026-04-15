package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

const defaultQueuePollInterval = 1 * time.Second

type TaskHandler func(context.Context, Task) error

type Queue struct {
	workerName   string
	repository   *TaskRepository
	handlers     map[string]TaskHandler
	health       *HeartbeatReporter
	pollInterval time.Duration
}

func NewQueue(
	workerName string,
	repository *TaskRepository,
	health *HeartbeatReporter,
) *Queue {
	return &Queue{
		workerName:   strings.TrimSpace(workerName),
		repository:   repository,
		handlers:     make(map[string]TaskHandler),
		health:       health,
		pollInterval: defaultQueuePollInterval,
	}
}

func (w *Queue) RegisterHandler(topic string, handler TaskHandler) error {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return errors.New("queue worker topic must not be empty")
	}
	if handler == nil {
		return errors.New("queue worker handler must not be nil")
	}
	w.handlers[topic] = handler
	return nil
}

func (w *Queue) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		if _, err := w.ProcessPending(ctx); err != nil && ctx.Err() == nil {
			if w.health != nil {
				w.health.ReportError(err)
			}
			log.Printf("queue worker %s drain failed: %v", w.workerName, err)
		} else if w.health != nil {
			w.health.ReportHealthy()
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *Queue) ProcessPending(ctx context.Context) (int, error) {
	processed := 0
	for {
		task, err := w.repository.PopNextAvailableTask(ctx, w.workerName)
		if err != nil {
			return processed, err
		}
		if task == nil {
			return processed, nil
		}

		handler, ok := w.handlers[task.Topic]
		if !ok {
			log.Printf("queue worker %s ignored unknown topic: %s", w.workerName, task.Topic)
			processed++
			continue
		}
		if err := handler(ctx, *task); err != nil {
			return processed, fmt.Errorf("handle worker task %q for %q: %w", task.Topic, w.workerName, err)
		}
		processed++
	}
}
