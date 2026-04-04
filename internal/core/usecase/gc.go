package usecase

import (
	"log/slog"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

// GCUseCase 负责对象存储的垃圾回收。
// 它识别并清理未被任何清单引用的孤立对象。
type GCUseCase struct {
	repo  port.AssetRepository
	store port.ObjectStorage
}

// NewGCUseCase 创建新的 GCUseCase 实例。
func NewGCUseCase(repo port.AssetRepository, store port.ObjectStorage) *GCUseCase {
	return &GCUseCase{repo: repo, store: store}
}

// ExecuteDryRun 返回孤立对象哈希列表，但不执行删除。
func (uc *GCUseCase) ExecuteDryRun() ([]string, error) {
	return uc.findOrphans()
}

// Execute 删除所有孤立对象，并返回清理数量。
func (uc *GCUseCase) Execute() (int, error) {
	orphans, err := uc.findOrphans()
	if err != nil {
		return 0, err
	}

	for _, hash := range orphans {
		uc.store.Delete(hash)
	}

	slog.Info("gc completed", "cleaned", len(orphans))
	return len(orphans), nil
}

// findOrphans 收集所有未被任何清单引用的孤立对象。
func (uc *GCUseCase) findOrphans() ([]string, error) {
	// 收集所有清单引用的对象哈希
	referenced := make(map[string]bool)
	manifests, err := uc.repo.ListManifests()
	if err != nil {
		return nil, err
	}
	for _, m := range manifests {
		collectObjects(m.Tree, referenced)
	}

	// 遍历所有已存储对象，找出孤立对象
	var orphans []string
	err = uc.store.Walk(func(hash string) error {
		if !referenced[hash] {
			orphans = append(orphans, hash)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return orphans, nil
}

// collectObjects 递归收集 TreeNode 树中所有对象的哈希。
func collectObjects(node domain.TreeNode, set map[string]bool) {
	if node.Type == "file" && node.Object != "" {
		set[node.Object] = true
	}
	for _, child := range node.Children {
		collectObjects(child, set)
	}
}
