package service

import (
	"context"
	"time"

	"polardbx-ui-backend/pkg/api/domain/platform/system/repository"
)

// NamespaceInfo 命名空间信息
type NamespaceInfo struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// ContextInfo 上下文信息
type ContextInfo struct {
	User             string `json:"user"`
	Context          string `json:"context"`
	DefaultNamespace string `json:"defaultNamespace"`
}

// SystemService 定义 system 业务逻辑层
type SystemService struct {
	repo repository.SystemRepository
}

// NewSystemService 创建新的 SystemService
func NewSystemService(repo repository.SystemRepository) *SystemService {
	return &SystemService{repo: repo}
}

// ListNamespaces 列出所有命名空间
func (s *SystemService) ListNamespaces(ctx context.Context) ([]NamespaceInfo, error) {
	namespaces, err := s.repo.ListNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]NamespaceInfo, 0, len(namespaces))
	for _, ns := range namespaces {
		items = append(items, NamespaceInfo{
			Name:      ns.Name,
			Status:    string(ns.Status.Phase),
			CreatedAt: ns.CreationTimestamp.Time,
		})
	}
	return items, nil
}
