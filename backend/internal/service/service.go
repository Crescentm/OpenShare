package service

import (
	"github.com/openshare/backend/internal/config"
	"github.com/openshare/backend/pkg/logger"
	"gorm.io/gorm"
)

// Services 聚合所有业务服务
// 使用依赖注入模式，便于测试和解耦
type Services struct {
	Admin *AdminService
	// 后续扩展其他服务
	// File       *FileService
	// Submission *SubmissionService
	// Tag        *TagService
	// ...
}

// Options 服务初始化配置
type Options struct {
	DB     *gorm.DB
	Config *config.Config
	Logger *logger.Logger
}

// New 创建服务聚合实例
func New(opts *Options) *Services {
	return &Services{
		Admin: NewAdminService(opts.DB, opts.Config, opts.Logger),
	}
}

// baseService 基础服务，提供公共依赖
type baseService struct {
	db     *gorm.DB
	config *config.Config
	logger *logger.Logger
}

func newBaseService(db *gorm.DB, cfg *config.Config, log *logger.Logger) baseService {
	return baseService{
		db:     db,
		config: cfg,
		logger: log,
	}
}
