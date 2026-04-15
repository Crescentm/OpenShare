package worker

import "time"

// Heartbeat records liveness for standalone background workers.
type Heartbeat struct {
	WorkerName string    `gorm:"column:worker_name;type:text;primaryKey"`
	LastSeenAt time.Time `gorm:"column:last_seen_at;type:datetime;not null;index:idx_worker_heartbeats_last_seen_at,sort:desc"`
	LastError  string    `gorm:"column:last_error;type:text;not null;default:''"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Heartbeat) TableName() string { return "worker_heartbeats" }
