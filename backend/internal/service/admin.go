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

// ============ 管理员 CRUD ============

// CreateAdminInput 创建管理员的输入参数
type CreateAdminInput struct {
	Username string
}

// CreateAdminResult 创建管理员的返回结果
type CreateAdminResult struct {
	Admin    *model.Admin
	Password string // 初始密码，仅创建时返回
}

// CreateAdmin 创建普通管理员
// 仅超级管理员可调用，返回管理员信息和初始密码
func (s *AdminService) CreateAdmin(input *CreateAdminInput) (*CreateAdminResult, error) {
	// 验证用户名
	if err := s.validateUsername(input.Username); err != nil {
		return nil, err
	}

	// 检查用户名是否已存在
	existing, err := s.GetByUsername(input.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	// 生成初始密码
	password, err := generateSecurePassword(12)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	admin := &model.Admin{
		Username: input.Username,
		Password: hashedPassword,
		Role:     model.RoleAdmin,
		Status:   model.AdminStatusActive,
	}

	if err := s.db.Create(admin).Error; err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	s.logger.Info("Admin created",
		"admin_id", admin.ID,
		"username", admin.Username,
	)

	return &CreateAdminResult{
		Admin:    admin,
		Password: password,
	}, nil
}

// validateUsername 验证用户名格式
func (s *AdminService) validateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if len(username) > 32 {
		return errors.New("username must be at most 32 characters")
	}
	// 只允许字母、数字、下划线
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_') {
			return errors.New("username can only contain letters, numbers And underscores")
		}
	}
	return nil
}

// ListAdminsInput 查询管理员列表的输入参数
type ListAdminsInput struct {
	Page     int
	PageSize int
	Role     string // 可选，筛选角色
	Status   string // 可选，筛选状态
	Keyword  string // 可选，搜索用户名
}

// ListAdminsResult 查询管理员列表的返回结果
type ListAdminsResult struct {
	Admins []*model.Admin
	Total  int64
}

// ListAdmins 获取管理员列表（分页）
func (s *AdminService) ListAdmins(input *ListAdminsInput) (*ListAdminsResult, error) {
	// 设置默认分页
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}

	query := s.db.Model(&model.Admin{})

	// 筛选条件
	if input.Role != "" {
		query = query.Where("role = ?", input.Role)
	}
	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}
	if input.Keyword != "" {
		query = query.Where("username ILIKE ?", "%"+input.Keyword+"%")
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count admins: %w", err)
	}

	// 查询列表
	var admins []*model.Admin
	offset := (input.Page - 1) * input.PageSize
	if err := query.
		Preload("Permissions").
		Order("created_at DESC").
		Offset(offset).
		Limit(input.PageSize).
		Find(&admins).Error; err != nil {
		return nil, fmt.Errorf("failed to list admins: %w", err)
	}

	return &ListAdminsResult{
		Admins: admins,
		Total:  total,
	}, nil
}

// UpdateAdminInput 更新管理员的输入参数
type UpdateAdminInput struct {
	ID       uint
	Username *string // 可选，新用户名
	Status   *string // 可选，新状态
}

// UpdateAdmin 更新管理员基本信息
func (s *AdminService) UpdateAdmin(input *UpdateAdminInput) (*model.Admin, error) {
	admin, err := s.GetByID(input.ID)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, errors.New("admin not found")
	}

	// 不允许修改超级管理员
	if admin.Role == model.RoleSuperAdmin {
		return nil, errors.New("cannot modify super admin")
	}

	updates := make(map[string]interface{})

	// 更新用户名
	if input.Username != nil && *input.Username != admin.Username {
		if err := s.validateUsername(*input.Username); err != nil {
			return nil, err
		}
		// 检查新用户名是否已存在
		existing, err := s.GetByUsername(*input.Username)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != admin.ID {
			return nil, errors.New("username already exists")
		}
		updates["username"] = *input.Username
	}

	// 更新状态
	if input.Status != nil {
		if *input.Status != model.AdminStatusActive && *input.Status != model.AdminStatusDisabled {
			return nil, errors.New("invalid status")
		}
		updates["status"] = *input.Status
	}

	if len(updates) > 0 {
		if err := s.db.Model(&model.Admin{}).
			Where("id = ?", input.ID).
			Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update admin: %w", err)
		}
	}

	// 返回更新后的数据
	return s.GetByID(input.ID)
}

// DeleteAdmin 删除管理员
// 采用软删除策略（禁用账号），超级管理员不可删除
func (s *AdminService) DeleteAdmin(id uint) error {
	admin, err := s.GetByID(id)
	if err != nil {
		return err
	}
	if admin == nil {
		return errors.New("admin not found")
	}

	// 不允许删除超级管理员
	if admin.Role == model.RoleSuperAdmin {
		return errors.New("cannot delete super admin")
	}

	// 软删除：设置状态为 disabled
	if err := s.db.Model(&model.Admin{}).
		Where("id = ?", id).
		Update("status", model.AdminStatusDisabled).Error; err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}

	// 同时清除权限
	if err := s.db.Where("admin_id = ?", id).Delete(&model.AdminPermission{}).Error; err != nil {
		s.logger.Warn("Failed to clear permissions for deleted admin",
			"admin_id", id,
			"error", err,
		)
	}

	s.logger.Info("Admin deleted",
		"admin_id", id,
		"username", admin.Username,
	)

	return nil
}

// ============ 权限管理 ============

// SetPermissions 设置管理员权限（整体替换）
// permissions 为权限代码列表，空列表表示清除所有权限
func (s *AdminService) SetPermissions(adminID uint, permissions []string) error {
	admin, err := s.GetByID(adminID)
	if err != nil {
		return err
	}
	if admin == nil {
		return errors.New("admin not found")
	}

	// 超级管理员无需设置权限
	if admin.Role == model.RoleSuperAdmin {
		return errors.New("super admin has all permissions by default")
	}

	// 验证权限代码
	for _, perm := range permissions {
		if !model.IsValidPermission(perm) {
			return fmt.Errorf("invalid permission: %s", perm)
		}
	}

	// 使用事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有权限
		if err := tx.Where("admin_id = ?", adminID).Delete(&model.AdminPermission{}).Error; err != nil {
			return fmt.Errorf("failed to clear permissions: %w", err)
		}

		// 添加新权限
		for _, perm := range permissions {
			ap := &model.AdminPermission{
				AdminID:    adminID,
				Permission: perm,
			}
			if err := tx.Create(ap).Error; err != nil {
				return fmt.Errorf("failed to add permission %s: %w", perm, err)
			}
		}

		return nil
	})
}

// AddPermission 为管理员添加单个权限
func (s *AdminService) AddPermission(adminID uint, permission string) error {
	admin, err := s.GetByID(adminID)
	if err != nil {
		return err
	}
	if admin == nil {
		return errors.New("admin not found")
	}

	if admin.Role == model.RoleSuperAdmin {
		return errors.New("super admin has all permissions by default")
	}

	if !model.IsValidPermission(permission) {
		return fmt.Errorf("invalid permission: %s", permission)
	}

	// 检查是否已有该权限
	if admin.HasPermission(permission) {
		return nil // 幂等操作，已有则跳过
	}

	ap := &model.AdminPermission{
		AdminID:    adminID,
		Permission: permission,
	}
	if err := s.db.Create(ap).Error; err != nil {
		return fmt.Errorf("failed to add permission: %w", err)
	}

	return nil
}

// RemovePermission 移除管理员的单个权限
func (s *AdminService) RemovePermission(adminID uint, permission string) error {
	admin, err := s.GetByID(adminID)
	if err != nil {
		return err
	}
	if admin == nil {
		return errors.New("admin not found")
	}

	if admin.Role == model.RoleSuperAdmin {
		return errors.New("cannot modify super admin permissions")
	}

	if err := s.db.Where("admin_id = ? AND permission = ?", adminID, permission).
		Delete(&model.AdminPermission{}).Error; err != nil {
		return fmt.Errorf("failed to remove permission: %w", err)
	}

	return nil
}

// GetPermissions 获取管理员的权限列表
func (s *AdminService) GetPermissions(adminID uint) ([]string, error) {
	admin, err := s.GetByID(adminID)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, errors.New("admin not found")
	}

	return admin.GetPermissionCodes(), nil
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
