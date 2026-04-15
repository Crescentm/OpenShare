package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"openshare/backend/internal/router"
	"openshare/backend/internal/worker"
)

func TestWorkerHealthReturnsUnavailableWithoutHeartbeat(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	engine := router.New(db, cfg, newRouterSessionManager(db))

	request := httptest.NewRequest(http.MethodGet, "/healthz/worker", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d, body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Status    string `json:"status"`
		LastError string `json:"last_error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode worker health response failed: %v", err)
	}
	if response.Status != "error" {
		t.Fatalf("expected status error, got %q", response.Status)
	}
	if response.LastError == "" {
		t.Fatal("expected last_error to explain missing heartbeat")
	}
}

func TestWorkerHealthReturnsOkForFreshHeartbeat(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)

	now := time.Now().UTC()
	if err := db.Create(&worker.Heartbeat{
		WorkerName: "managed-sync-worker",
		LastSeenAt: now,
		LastError:  "",
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create worker heartbeat failed: %v", err)
	}
	if err := db.Create(&worker.Task{
		ID:         "task-1",
		WorkerName: "another-worker",
		Topic:      "noop",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create unrelated worker task failed: %v", err)
	}

	engine := router.New(db, cfg, newRouterSessionManager(db))
	request := httptest.NewRequest(http.MethodGet, "/healthz/worker", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Status       string    `json:"status"`
		WorkerName   string    `json:"worker_name"`
		LastSeenAt   time.Time `json:"last_seen_at"`
		QueueBacklog int64     `json:"queue_backlog"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode worker health response failed: %v", err)
	}
	if response.Status != "ok" {
		t.Fatalf("expected status ok, got %q", response.Status)
	}
	if response.WorkerName != "managed-sync-worker" {
		t.Fatalf("expected worker name managed-sync-worker, got %q", response.WorkerName)
	}
	if response.LastSeenAt.IsZero() {
		t.Fatal("expected last_seen_at to be populated")
	}
	if response.QueueBacklog != 0 {
		t.Fatalf("expected empty backlog, got %d", response.QueueBacklog)
	}
}
