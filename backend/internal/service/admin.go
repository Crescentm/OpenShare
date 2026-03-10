package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/openshare/backend/internal/config"
	"github.com/openshare/backend/internal/model"
	"github.com/openshare/backend/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminService 管理员业务服务
type AdminService struct {
	baseService
}

// NewAdminService 创建管理员服务实例
func NewAdminService(db *gorm.DB, cfg *config.Config, log *logger.Logger) *AdminService {
	return &AdminService{
		baseService: newBaseService(db, cfg, log),
	}
}

// InitSuperAdmin 初始化超级管理员
// 系统首次启动时自动创建，返回 (是否新创建, 初始密码, 错误)
func (s *AdminService) InitSuperAdmin() (bool, string, error) {
	// 检查是否已存在超级管理员
	var count int64
	if err := s.db.Model(&model.Admin{}).
		Where("role = ?", model.RoleSuperAdmin).
		Count(&count).Error; err != nil {
		return false, "", fmt.Errorf("failed to check super admin: %w", err)
	}

	// 已存在则跳过
	if count > 0 {
		s.logger.Info("Super admin already exists, skip initialization")
		return false, "", nil
	}

	// 生成安全的随机密码
	password, err := generateSecurePassword(16)
	if err != nil {
		return false, "", fmt.Errorf("failed to generate password: %w", err)
	}

	// 哈希密码
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return false, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建超级管理员
	admin := &model.Admin{
		Username: "admin",
		Password: hashedPassword,
		Role:     model.RoleSuperAdmin,
		Status:   model.AdminStatusActive,
	}

	if err := s.db.Create(admin).Error; err != nil {
		// 处理并发创建的情况（唯一约束冲突）
		if strings.Contains(err.Error(), "duplicate") ||
			strings.Contains(err.Error(), "unique") {
			s.logger.Info("Super admin created by another process")
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to create super admin: %w", err)
	}

	s.logger.Info("Super admin initialized successfully",
		"username", admin.Username,
		"id", admin.ID,
	)

	return true, password, nil
}

// GetByID 根据 ID 获取管理员
func (s *AdminService) GetByID(id uint) (*model.Admin, error) {
	var admin model.Admin
	if err := s.db.Preload("Permissions").First(&admin, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}
	return &admin, nil
}

// GetByUsername 根据用户名获取管理员
func (s *AdminService) GetByUsername(username string) (*model.Admin, error) {
	var admin model.Admin
	if err := s.db.Preload("Permissions").
		Where("username = ?", username).
		First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}
	return &admin, nil
}

// ValidateCredentials 验证登录凭证
// 返回 (管理员, 错误)，凭证无效时返回 nil
func (s *AdminService) ValidateCredentials(username, password string) (*model.Admin, error) {
	admin, err := s.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, nil // 用户不存在
	}

	// 检查账号状态
	if admin.Status != model.AdminStatusActive {
		return nil, nil // 账号已禁用
	}

	// 验证密码
	if !CheckPassword(password, admin.Password) {
		return nil, nil // 密码错误
	}

	return admin, nil
}

// UpdateLastLogin 更新最后登录时间
func (s *AdminService) UpdateLastLogin(id uint) error {
	now := time.Now()
	return s.db.Model(&model.Admin{}).
		Where("id = ?", id).
		Update("last_login", now).Error
}

// ChangePassword 修改密码
func (s *AdminService) ChangePassword(id uint, oldPassword, newPassword string) error {
	admin, err := s.GetByID(id)
	if err != nil {
		return err
	}
	if admin == nil {
		return errors.New("admin not found")
	}

	// 验证旧密码
	if !CheckPassword(oldPassword, admin.Password) {
		return errors.New("invalid old password")
	}

	// 哈希新密码
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.db.Model(&model.Admin{}).
		Where("id = ?", id).
		Update("password", hashedPassword).Error
}

// ResetPassword 重置密码（超级管理员操作）
// 返回新生成的随机密码
func (s *AdminService) ResetPassword(id uint) (string, error) {
	admin, err := s.GetByID(id)
	if err != nil {
		return "", err
	}
	if admin == nil {
		return "", errors.New("admin not found")
	}

	// 不允许重置超级管理员密码（需要通过 ChangePassword）
	if admin.Role == model.RoleSuperAdmin {
		return "", errors.New("cannot reset super admin password")
	}

	// 生成新密码
	password, err := generateSecurePassword(12)
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.db.Model(&model.Admin{}).
		Where("id = ?", id).
		Update("password", hashedPassword).Error; err != nil {
		return "", fmt.Errorf("failed to update password: %w", err)
	}

	return password, nil
}

// ============ 密码工具函数 ============

// HashPassword 使用 bcrypt 哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword 验证密码是否匹配
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateSecurePassword 生成加密安全的随机密码
// 使用 crypto/rand 确保密码随机性
func generateSecurePassword(length int) (string, error) {
	// 计算需要的字节数（base64 编码后会变长）
	byteLength := (length*3 + 3) / 4

	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Base64 编码并截取指定长度
	password := base64.URLEncoding.EncodeToString(bytes)

	// 移除可能造成混淆的字符，保留字母数字和部分符号
	password = strings.ReplaceAll(password, "-", "")
	password = strings.ReplaceAll(password, "_", "")

	if len(password) > length {
		password = password[:length]
	}

	return password, nil
}
