package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/gorm"

	"openshare/backend/internal/model"
)

func TestPublicSubmissionLookupByReceiptCode(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	engine := New(db, cfg, newRouterSessionManager(db))

	submission := createPendingSubmissionForTest(t, db, "RECEIPT88")
	createFileForSubmission(t, db, submission.ID, 23)

	request := httptest.NewRequest(http.MethodGet, "/api/public/submissions/RECEIPT88", nil)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		ReceiptCode   string `json:"receipt_code"`
		Title         string `json:"title"`
		Status        string `json:"status"`
		DownloadCount int64  `json:"download_count"`
		RejectReason  string `json:"reject_reason"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if response.ReceiptCode != "RECEIPT88" {
		t.Fatalf("unexpected receipt code %q", response.ReceiptCode)
	}
	if response.Title != submission.TitleSnapshot {
		t.Fatalf("unexpected title %q", response.Title)
	}
	if response.Status != string(model.SubmissionStatusPending) {
		t.Fatalf("unexpected status %q", response.Status)
	}
	if response.DownloadCount != 23 {
		t.Fatalf("unexpected download_count %d", response.DownloadCount)
	}
}

func TestPublicSubmissionLookupReturns404WhenMissing(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	engine := New(db, cfg, newRouterSessionManager(db))

	request := httptest.NewRequest(http.MethodGet, "/api/public/submissions/UNKNOWN88", nil)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d, body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestPublicSubmissionLookupRejectsInvalidReceiptCode(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	engine := New(db, cfg, newRouterSessionManager(db))

	request := httptest.NewRequest(http.MethodGet, "/api/public/submissions/invalid!*", nil)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", recorder.Code, recorder.Body.String())
	}
}

func createFileForSubmission(t *testing.T, db *gorm.DB, submissionID string, downloadCount int64) {
	t.Helper()

	file := &model.File{
		ID:            mustNewID(t),
		SubmissionID:  &submissionID,
		Title:         "existing",
		OriginalName:  "existing.pdf",
		StoredName:    mustNewID(t) + ".pdf",
		Extension:     ".pdf",
		MimeType:      "application/pdf",
		Size:          1024,
		DiskPath:      "/tmp/existing.pdf",
		Status:        model.ResourceStatusOffline,
		DownloadCount: downloadCount,
		UploaderIP:    "127.0.0.1",
		CreatedAt:     time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC),
	}
	if err := db.Create(file).Error; err != nil {
		t.Fatalf("create file for submission failed: %v", err)
	}
}
