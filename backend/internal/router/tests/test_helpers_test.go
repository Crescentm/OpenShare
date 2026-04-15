package router_test

import (
	"path/filepath"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"openshare/backend/internal/admin"
	"openshare/backend/internal/bootstrap"
	"openshare/backend/internal/config"
	"openshare/backend/internal/model"
	"openshare/backend/internal/session"
	"openshare/backend/internal/storage"
	"openshare/backend/pkg/database"
	"openshare/backend/pkg/identity"
)

func newRouterTestConfig(t *testing.T) config.Config {
	t.Helper()

	cfg := config.Default()
	cfg.Session.Secret = "test-secret"
	cfg.Storage.Root = filepath.Join(t.TempDir(), "storage")
	cfg.Database.Path = filepath.Join(t.TempDir(), "openshare-test.db")

	if err := storage.EnsureLayout(cfg.Storage); err != nil {
		t.Fatalf("ensure storage layout failed: %v", err)
	}

	return cfg
}

func newRouterTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "openshare-router-test.db")
	db, err := database.NewSQLite(database.Options{
		Path:      dbPath,
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

	return db
}

func newRouterSessionManager(db *gorm.DB) *session.Manager {
	return session.NewManager(db, config.SessionConfig{
		Name:            "openshare_session",
		Secret:          "test-secret",
		Path:            "/",
		MaxAgeSeconds:   3600,
		HTTPOnly:        true,
		Secure:          false,
		SameSite:        "lax",
		RenewWindowSecs: 300,
	}, admin.NewAdminSessionRepository())
}

func createRouterTestAdmin(t *testing.T, db *gorm.DB, username, password string) *model.Admin {
	t.Helper()
	return createRouterTestAdminWithAccess(t, db, adminAccess{
		username: username,
		password: password,
		role:     string(model.AdminRoleSuperAdmin),
	})
}

type adminAccess struct {
	username    string
	password    string
	role        string
	permissions []model.AdminPermission
}

func createRouterTestAdminWithAccess(t *testing.T, db *gorm.DB, access adminAccess) *model.Admin {
	t.Helper()

	adminID, err := identity.NewID()
	if err != nil {
		t.Fatalf("generate admin id failed: %v", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(access.password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate password hash failed: %v", err)
	}

	admin := &model.Admin{
		ID:           adminID,
		Username:     access.username,
		DisplayName:  access.username,
		PasswordHash: string(passwordHash),
		Role:         access.role,
		Permissions:  model.NormalizeAdminPermissions(access.permissions),
		Status:       model.AdminStatusActive,
	}
	if err := db.Create(admin).Error; err != nil {
		t.Fatalf("create admin failed: %v", err)
	}

	return admin
}

func mustNewID(t *testing.T) string {
	t.Helper()
	id, err := identity.NewID()
	if err != nil {
		t.Fatalf("generate id failed: %v", err)
	}
	return id
}
