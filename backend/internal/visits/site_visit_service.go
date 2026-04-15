package visits

import (
	"context"
	"fmt"
	"strings"
)

type SiteVisitService struct {
	repo *SiteVisitRepository
}

func NewSiteVisitService(repo *SiteVisitRepository) *SiteVisitService {
	return &SiteVisitService{repo: repo}
}

func (s *SiteVisitService) Record(ctx context.Context, scope string, path string, ip string) error {
	if strings.TrimSpace(ip) == "" {
		return nil
	}
	if err := s.repo.Create(ctx, scope, path, ip); err != nil {
		return fmt.Errorf("record site visit: %w", err)
	}
	return nil
}
