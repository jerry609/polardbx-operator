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

// StorageClassInfo 存储类信息
type StorageClassInfo struct {
	Name              string            `json:"name"`
	Provisioner       string            `json:"provisioner"`
	ReclaimPolicy     string            `json:"reclaimPolicy,omitempty"`
	VolumeBindingMode string            `json:"volumeBindingMode,omitempty"`
	IsDefault         bool              `json:"isDefault"`
	Parameters        map[string]string `json:"parameters,omitempty"`
}

// PolarDBXVersionInfo PolarDB-X 版本信息
type PolarDBXVersionInfo struct {
	Version     string `json:"version"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Recommended bool   `json:"recommended"`
	Deprecated  bool   `json:"deprecated"`
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

// ListStorageClasses 列出所有存储类
func (s *SystemService) ListStorageClasses(ctx context.Context) ([]StorageClassInfo, error) {
	storageClasses, err := s.repo.ListStorageClasses(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]StorageClassInfo, 0, len(storageClasses))
	for _, sc := range storageClasses {
		isDefault := false
		if sc.Annotations != nil {
			if v, ok := sc.Annotations["storageclass.kubernetes.io/is-default-class"]; ok && v == "true" {
				isDefault = true
			}
		}

		reclaimPolicy := ""
		if sc.ReclaimPolicy != nil {
			reclaimPolicy = string(*sc.ReclaimPolicy)
		}

		volumeBindingMode := ""
		if sc.VolumeBindingMode != nil {
			volumeBindingMode = string(*sc.VolumeBindingMode)
		}

		items = append(items, StorageClassInfo{
			Name:              sc.Name,
			Provisioner:       sc.Provisioner,
			ReclaimPolicy:     reclaimPolicy,
			VolumeBindingMode: volumeBindingMode,
			IsDefault:         isDefault,
			Parameters:        sc.Parameters,
		})
	}
	return items, nil
}

// GetPolarDBXVersions 返回支持的 PolarDB-X 版本列表
func (s *SystemService) GetPolarDBXVersions() []PolarDBXVersionInfo {
	// 版本信息可以从配置文件或 CRD 中读取，这里先硬编码常用版本
	return []PolarDBXVersionInfo{
		{Version: "8.0.18", Label: "8.0.18 (最新稳定版)", Recommended: true},
		{Version: "8.0.17", Label: "8.0.17"},
		{Version: "8.0.16", Label: "8.0.16"},
		{Version: "5.7.14", Label: "5.7.14 (旧版本)", Deprecated: true},
	}
}
