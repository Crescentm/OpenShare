package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"openshare/backend/internal/model"
	"openshare/backend/internal/router"
)

func TestAdminLoginCreatesSessionAndReturnsProfile(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	admin := createRouterTestAdmin(t, db, "superadmin", "s3cret-pass")
	manager := newRouterSessionManager(db)
	engine := router.New(db, cfg, manager)

	body := bytes.NewBufferString(`{"username":"superadmin","password":"s3cret-pass"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/admin/session/login", body)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Admin struct {
			ID          string   `json:"id"`
			Username    string   `json:"username"`
			DisplayName string   `json:"display_name"`
			Role        string   `json:"role"`
			Status      string   `json:"status"`
			Permissions []string `json:"permissions"`
		} `json:"admin"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if response.Admin.ID != admin.ID {
		t.Fatalf("expected admin id %q, got %q", admin.ID, response.Admin.ID)
	}
	if response.Admin.Username != admin.Username {
		t.Fatalf("expected username %q, got %q", admin.Username, response.Admin.Username)
	}
	if response.Admin.DisplayName != admin.DisplayName {
		t.Fatalf("expected display name %q, got %q", admin.DisplayName, response.Admin.DisplayName)
	}
	if len(response.Admin.Permissions) != 0 {
		t.Fatalf("expected no explicit permissions for super admin bootstrap test, got %v", response.Admin.Permissions)
	}

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "openshare_session" {
		t.Fatalf("unexpected cookie name %q", cookies[0].Name)
	}

	var count int64
	if err := db.Model(&model.AdminSession{}).Count(&count).Error; err != nil {
		t.Fatalf("count sessions failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 persisted session, got %d", count)
	}
}

func TestAdminLoginRejectsInvalidCredentials(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	createRouterTestAdmin(t, db, "superadmin", "correct-password")
	manager := newRouterSessionManager(db)
	engine := router.New(db, cfg, manager)

	body := bytes.NewBufferString(`{"username":"superadmin","password":"wrong-password"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/admin/session/login", body)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d, body=%s", recorder.Code, recorder.Body.String())
	}

	var count int64
	if err := db.Model(&model.AdminSession{}).Count(&count).Error; err != nil {
		t.Fatalf("count sessions failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 persisted sessions, got %d", count)
	}
}

func TestAdminLogoutDeletesSessionAndClearsCookie(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	admin := createRouterTestAdmin(t, db, "superadmin", "s3cret-pass")
	manager := newRouterSessionManager(db)
	engine := router.New(db, cfg, manager)

	cookieValue, identity, err := manager.Create(t.Context(), admin)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	if identity.AdminID != admin.ID {
		t.Fatalf("expected admin id %q, got %q", admin.ID, identity.AdminID)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/admin/session/logout", nil)
	request.AddCookie(&http.Cookie{
		Name:  manager.CookieName(),
		Value: cookieValue,
		Path:  "/",
	})
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", recorder.Code)
	}

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie in logout response, got %d", len(cookies))
	}
	if cookies[0].MaxAge != -1 {
		t.Fatalf("expected cleared cookie MaxAge=-1, got %d", cookies[0].MaxAge)
	}

	var count int64
	if err := db.Model(&model.AdminSession{}).Count(&count).Error; err != nil {
		t.Fatalf("count sessions failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 persisted sessions after logout, got %d", count)
	}
}

func TestAdminMeRequiresAuthentication(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	manager := newRouterSessionManager(db)
	engine := router.New(db, cfg, manager)

	request := httptest.NewRequest(http.MethodGet, "/api/admin/me", nil)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}

func TestAdminMeReturnsIdentityFromSession(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	admin := createRouterTestAdminWithAccess(t, db, adminAccess{
		username: "reviewer",
		password: "s3cret-pass",
		role:     string(model.AdminRoleAdmin),
		permissions: []model.AdminPermission{
			model.AdminPermissionReviewSubmissions,
			model.AdminPermissionManageSystem,
		},
	})
	manager := newRouterSessionManager(db)
	engine := router.New(db, cfg, manager)

	cookieValue, _, err := manager.Create(t.Context(), admin)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/admin/me", nil)
	request.AddCookie(&http.Cookie{Name: manager.CookieName(), Value: cookieValue, Path: "/"})
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Admin struct {
			ID          string   `json:"id"`
			Username    string   `json:"username"`
			DisplayName string   `json:"display_name"`
			Role        string   `json:"role"`
			Permissions []string `json:"permissions"`
		} `json:"admin"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if response.Admin.ID != admin.ID {
		t.Fatalf("expected admin id %q, got %q", admin.ID, response.Admin.ID)
	}
	if response.Admin.Role != string(model.AdminRoleAdmin) {
		t.Fatalf("expected role %q, got %q", model.AdminRoleAdmin, response.Admin.Role)
	}
	if response.Admin.DisplayName != admin.DisplayName {
		t.Fatalf("expected display name %q, got %q", admin.DisplayName, response.Admin.DisplayName)
	}
	if len(response.Admin.Permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %v", response.Admin.Permissions)
	}
}

func TestAdminChangePassword(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	admin := createRouterTestAdmin(t, db, "superadmin", "old-pass-123")
	manager := newRouterSessionManager(db)
	engine := router.New(db, cfg, manager)

	cookieValue, _, err := manager.Create(t.Context(), admin)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/admin/session/change-password", bytes.NewBufferString(`{"new_password":"new-pass-123"}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: manager.CookieName(), Value: cookieValue, Path: "/"})
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d, body=%s", recorder.Code, recorder.Body.String())
	}

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/admin/session/login", bytes.NewBufferString(`{"username":"superadmin","password":"new-pass-123"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginRecorder := httptest.NewRecorder()
	engine.ServeHTTP(loginRecorder, loginRequest)

	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("expected login with new password to succeed, got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}
}

func TestPermissionMiddlewareRejectsUnauthorizedPermission(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	admin := createRouterTestAdminWithAccess(t, db, adminAccess{
		username: "reviewer",
		password: "s3cret-pass",
		role:     string(model.AdminRoleAdmin),
		permissions: []model.AdminPermission{
			model.AdminPermissionReviewSubmissions,
		},
	})
	manager := newRouterSessionManager(db)
	engine := router.New(db, cfg, manager)

	cookieValue, _, err := manager.Create(t.Context(), admin)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/admin/_internal/system", nil)
	request.AddCookie(&http.Cookie{Name: manager.CookieName(), Value: cookieValue, Path: "/"})
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d, body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestPermissionMiddlewareAllowsGrantedPermission(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	admin := createRouterTestAdminWithAccess(t, db, adminAccess{
		username: "reviewer",
		password: "s3cret-pass",
		role:     string(model.AdminRoleAdmin),
		permissions: []model.AdminPermission{
			model.AdminPermissionReviewSubmissions,
		},
	})
	manager := newRouterSessionManager(db)
	engine := router.New(db, cfg, manager)

	cookieValue, _, err := manager.Create(t.Context(), admin)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/admin/_internal/review", nil)
	request.AddCookie(&http.Cookie{Name: manager.CookieName(), Value: cookieValue, Path: "/"})
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestPermissionMiddlewareAllowsSuperAdminBypass(t *testing.T) {
	cfg := newRouterTestConfig(t)
	db := newRouterTestDB(t)
	admin := createRouterTestAdminWithAccess(t, db, adminAccess{
		username:    "superadmin",
		password:    "s3cret-pass",
		role:        string(model.AdminRoleSuperAdmin),
		permissions: nil,
	})
	manager := newRouterSessionManager(db)
	engine := router.New(db, cfg, manager)

	cookieValue, _, err := manager.Create(t.Context(), admin)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/admin/_internal/system", nil)
	request.AddCookie(&http.Cookie{Name: manager.CookieName(), Value: cookieValue, Path: "/"})
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", recorder.Code, recorder.Body.String())
	}
}
