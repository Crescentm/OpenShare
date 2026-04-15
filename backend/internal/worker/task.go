package worker

import "time"

// Task is a lightweight cross-process task queue item for background workers.
type Task struct {
	ID          string     `gorm:"column:id;type:text;primaryKey"`
	WorkerName  string     `gorm:"column:worker_name;type:text;not null;uniqueIndex:ux_worker_tasks_worker_name_dedupe_key,priority:1;index:idx_worker_tasks_worker_name_available_at,priority:1"`
	Topic       string     `gorm:"column:topic;type:text;not null"`
	Payload     string     `gorm:"column:payload;type:text;not null;default:''"`
	DedupeKey   *string    `gorm:"column:dedupe_key;type:text;uniqueIndex:ux_worker_tasks_worker_name_dedupe_key,priority:2"`
	AvailableAt *time.Time `gorm:"column:available_at;type:datetime;index:idx_worker_tasks_worker_name_available_at,priority:2,sort:asc"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime;index:idx_worker_tasks_updated_at,sort:asc"`
}

func (Task) TableName() string { return "worker_tasks" }
