package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"openshare/backend/internal/repository"
)

var ErrInvalidOperationLogQuery = errors.New("invalid operation log query")

const (
	defaultOperationLogPage     = 1
	defaultOperationLogPageSize = 20
	maxOperationLogPageSize     = 100
)

type OperationLogService struct {
	repo *repository.OperationLogRepository
}

type ListOperationLogsInput struct {
	Action     string
	TargetType string
	Page       int
	PageSize   int
}

type OperationLogItem struct {
	ID         string    `json:"id"`
	AdminID    *string   `json:"admin_id,omitempty"`
	AdminName  string    `json:"admin_name"`
	Action     string    `json:"action"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Detail     string    `json:"detail"`
	IP         string    `json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
}

type OperationLogListResult struct {
	Items    []OperationLogItem `json:"items"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int64              `json:"total"`
}

func NewOperationLogService(repo *repository.OperationLogRepository) *OperationLogService {
	return &OperationLogService{repo: repo}
}

func (s *OperationLogService) List(ctx context.Context, input ListOperationLogsInput) (*OperationLogListResult, error) {
	page := input.Page
	if page == 0 {
		page = defaultOperationLogPage
	}
	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = defaultOperationLogPageSize
	}
	if page < 1 || pageSize < 1 || pageSize > maxOperationLogPageSize {
		return nil, ErrInvalidOperationLogQuery
	}

	rows, total, err := s.repo.List(ctx, strings.TrimSpace(input.Action), strings.TrimSpace(input.TargetType), page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list operation logs: %w", err)
	}

	items := make([]OperationLogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, OperationLogItem{
			ID:         row.ID,
			AdminID:    row.AdminID,
			AdminName:  row.AdminName,
			Action:     row.Action,
			TargetType: row.TargetType,
			TargetID:   row.TargetID,
			Detail:     row.Detail,
			IP:         row.IP,
			CreatedAt:  row.CreatedAt,
		})
	}

	return &OperationLogListResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}
