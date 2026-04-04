package usecase

import (
	"log/slog"

	"github.com/ZouZhao321/distill/internal/core/port"
)

// RemoveUseCase 负责从仓库移除资产。
// 它删除清单和引用，但不删除底层对象，对象由 GC 统一清理。
type RemoveUseCase struct {
	repo port.AssetRepository
}

// NewRemoveUseCase 创建新的 RemoveUseCase 实例。
func NewRemoveUseCase(repo port.AssetRepository) *RemoveUseCase {
	return &RemoveUseCase{repo: repo}
}

// Execute 通过删除清单和引用来移除指定资产。
// 对象保留在存储中，等待 GC 清理。
func (uc *RemoveUseCase) Execute(name string) error {
	ref, err := uc.repo.GetRef(name)
	if err != nil {
		return err
	}

	// 删除清单
	if err := uc.repo.RemoveManifest(ref.Manifest); err != nil {
		return err
	}

	// 删除引用
	err = uc.repo.DeleteRef(name)
	if err == nil {
		slog.Info("asset removed", "name", name, "manifest_hash", ref.Manifest)
	}
	return err
}
