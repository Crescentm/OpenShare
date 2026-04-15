package admin

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"openshare/backend/internal/model"
	"openshare/backend/pkg/identity"
)

type AdminBootstrapService struct {
	db        *gorm.DB
	adminRepo *AdminRepository
}

func NewAdminBootstrapService(db *gorm.DB, adminRepo *AdminRepository) *AdminBootstrapService {
	return &AdminBootstrapService{
		db:        db,
		adminRepo: adminRepo,
	}
}

func (s *AdminBootstrapService) EnsureDefaultSuperAdmin() error {
	result, err := s.ensureDefaultSuperAdmin()
	if err != nil {
		return err
	}

	if result.Created {
		log.Printf(
			"[bootstrap] super admin initialized; username=%s password=%s",
			result.Username,
			result.PlaintextPassword,
		)
	}

	return nil
}

type bootstrapResult struct {
	Created           bool
	Username          string
	PlaintextPassword string
}

func (s *AdminBootstrapService) ensureDefaultSuperAdmin() (*bootstrapResult, error) {
	result := &bootstrapResult{}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		exists, err := s.adminRepo.HasSuperAdmin(tx)
		if err != nil {
			return fmt.Errorf("check super admin existence: %w", err)
		}
		if exists {
			return nil
		}

		username, err := generateInitialUsername()
		if err != nil {
			return fmt.Errorf("generate initial username: %w", err)
		}

		password, err := generateInitialPassword()
		if err != nil {
			return fmt.Errorf("generate initial password: %w", err)
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash initial password: %w", err)
		}

		adminID, err := identity.NewID()
		if err != nil {
			return fmt.Errorf("generate super admin id: %w", err)
		}

		admin := &model.Admin{
			ID:           adminID,
			Username:     username,
			DisplayName:  "Superadmin",
			PasswordHash: string(passwordHash),
			Role:         string(model.AdminRoleSuperAdmin),
			Status:       model.AdminStatusActive,
		}
		if err := s.adminRepo.Create(tx, admin); err != nil {
			return fmt.Errorf("create default super admin: %w", err)
		}

		result.Created = true
		result.Username = admin.Username
		result.PlaintextPassword = password
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func generateInitialUsername() (string, error) {
	randomBytes := make([]byte, 5)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return "admin" + hex.EncodeToString(randomBytes), nil
}

func generateInitialPassword() (string, error) {
	randomBytes := make([]byte, 18)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	password := base64.RawURLEncoding.EncodeToString(randomBytes)
	password = strings.TrimSpace(password)
	if len(password) < 16 {
		return "", errors.New("generated password is unexpectedly short")
	}

	return password, nil
}
