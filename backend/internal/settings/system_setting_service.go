package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"openshare/backend/internal/config"
	"openshare/backend/internal/search"
	"openshare/backend/pkg/identity"
)

const systemPolicyKey = "system_policy"
const searchProfileKey = "search_profile"

var ErrInvalidSystemPolicy = errors.New("invalid system policy")
var ErrInvalidSearchProfile = errors.New("invalid search profile")

type UploadPolicy struct {
	MaxUploadTotalBytes int64 `json:"max_upload_total_bytes"`
}

type DownloadPolicy struct {
	MaxDownloadTotalBytes int64 `json:"max_download_total_bytes"`
}

type SystemPolicy struct {
	Upload   UploadPolicy   `json:"upload"`
	Download DownloadPolicy `json:"download"`
}

type SystemSettingService struct {
	repo              *SystemSettingRepository
	defaultPolicy     SystemPolicy
	searchProfilePath string
	nowFunc           func() time.Time
}

func defaultSystemPolicy(cfg config.Config) SystemPolicy {
	return SystemPolicy{
		Upload: UploadPolicy{
			MaxUploadTotalBytes: cfg.Upload.MaxUploadTotalBytes,
		},
		Download: DownloadPolicy{
			MaxDownloadTotalBytes: cfg.Download.MaxDownloadTotalBytes,
		},
	}
}

func NewSystemSettingService(repo *SystemSettingRepository, cfg config.Config) *SystemSettingService {
	return &SystemSettingService{
		repo:              repo,
		defaultPolicy:     defaultSystemPolicy(cfg),
		searchProfilePath: cfg.SearchEngine.SemanticProfilePath,
		nowFunc:           func() time.Time { return time.Now().UTC() },
	}
}

func (s *SystemSettingService) GetPolicy(ctx context.Context) (*SystemPolicy, error) {
	item, err := s.repo.FindByKey(ctx, systemPolicyKey)
	if err != nil {
		return nil, err
	}
	if item == nil || strings.TrimSpace(item.Value) == "" {
		policy := s.defaultPolicy
		return &policy, nil
	}

	var policy SystemPolicy
	if err := json.Unmarshal([]byte(item.Value), &policy); err != nil {
		return nil, fmt.Errorf("decode system policy: %w", err)
	}
	return &policy, nil
}

func (s *SystemSettingService) GetSearchProfile(ctx context.Context) (*search.SemanticProfile, error) {
	item, err := s.repo.FindByKey(ctx, searchProfileKey)
	if err != nil {
		return nil, err
	}
	if item == nil || strings.TrimSpace(item.Value) == "" {
		if strings.TrimSpace(s.searchProfilePath) == "" {
			return &search.SemanticProfile{}, nil
		}
		return search.LoadSemanticProfile(s.searchProfilePath)
	}

	var profile search.SemanticProfile
	if err := json.Unmarshal([]byte(item.Value), &profile); err != nil {
		return nil, fmt.Errorf("decode search profile: %w", err)
	}
	if err := profile.Validate(); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *SystemSettingService) SavePolicy(ctx context.Context, policy SystemPolicy, operatorID string, operatorIP string) (*SystemPolicy, error) {
	if policy.Upload.MaxUploadTotalBytes <= 0 {
		return nil, ErrInvalidSystemPolicy
	}
	if policy.Download.MaxDownloadTotalBytes <= 0 {
		return nil, ErrInvalidSystemPolicy
	}

	payload, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("encode system policy: %w", err)
	}
	logID, err := identity.NewID()
	if err != nil {
		return nil, fmt.Errorf("generate system policy log id: %w", err)
	}
	if err := s.repo.UpsertWithLog(ctx, systemPolicyKey, string(payload), operatorID, operatorIP, logID, s.nowFunc()); err != nil {
		return nil, fmt.Errorf("save system policy: %w", err)
	}
	return &policy, nil
}

func (s *SystemSettingService) SaveSearchProfile(ctx context.Context, profile search.SemanticProfile, operatorID string, operatorIP string) (*search.SemanticProfile, error) {
	if err := profile.Validate(); err != nil {
		return nil, ErrInvalidSearchProfile
	}

	payload, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("encode search profile: %w", err)
	}
	logID, err := identity.NewID()
	if err != nil {
		return nil, fmt.Errorf("generate search profile log id: %w", err)
	}
	if err := s.repo.UpsertWithLog(ctx, searchProfileKey, string(payload), operatorID, operatorIP, logID, s.nowFunc()); err != nil {
		return nil, fmt.Errorf("save search profile: %w", err)
	}
	return &profile, nil
}
