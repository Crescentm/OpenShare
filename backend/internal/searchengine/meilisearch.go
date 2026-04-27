package searchengine

import (
	"context"
	"errors"
	"fmt"

	"github.com/meilisearch/meilisearch-go"

	"openshare/backend/internal/config"
)

var ErrDisabled = errors.New("search engine is disabled")

type MeilisearchClient struct {
	service   meilisearch.ServiceManager
	indexName string
}

func NewMeilisearchClient(cfg config.SearchEngineConfig) (*MeilisearchClient, error) {
	if !cfg.Enabled {
		return nil, ErrDisabled
	}
	if err := cfg.ValidateForMeilisearch(); err != nil {
		return nil, err
	}

	return &MeilisearchClient{
		service:   meilisearch.New(cfg.Host, meilisearch.WithAPIKey(cfg.APIKey)),
		indexName: cfg.IndexName,
	}, nil
}

func (c *MeilisearchClient) IndexName() string {
	return c.indexName
}

func (c *MeilisearchClient) HealthCheck(ctx context.Context) error {
	health, err := c.service.HealthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("check meilisearch health: %w", err)
	}
	if health == nil || health.Status != "available" {
		return fmt.Errorf("meilisearch is not available")
	}
	return nil
}

func (c *MeilisearchClient) Service() meilisearch.ServiceManager {
	return c.service
}
