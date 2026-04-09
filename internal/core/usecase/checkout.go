package usecase

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

// CheckoutUseCase 负责从仓库还原资产到目标目录。
type CheckoutUseCase struct {
	repo  port.AssetRepository
	store port.ObjectStorage
}

// NewCheckoutUseCase 创建新的 CheckoutUseCase 实例。
func NewCheckoutUseCase(repo port.AssetRepository, store port.ObjectStorage) *CheckoutUseCase {
	return &CheckoutUseCase{repo: repo, store: store}
}

// Execute 将指定资产还原到 outputDir。
// overwrite 控制文件已存在时的行为："skip"、"force" 或 "ask"。
func (uc *CheckoutUseCase) Execute(name, outputDir, overwrite string) error {
	ref, err := uc.repo.GetRef(name)
	if err != nil {
		return err
	}

	manifest, err := uc.repo.GetManifest(ref.Manifest)
	if err != nil {
		return err
	}

	slog.Info("checkout started", "name", name, "output", outputDir, "overwrite", overwrite)
	return restoreTree(uc.store, manifest.Tree, outputDir, overwrite)
}

// restoreTree 递归将 TreeNode 树还原到文件系统。
func restoreTree(store port.ObjectStorage, node domain.TreeNode, targetDir, overwrite string) error {
	// 防御性校验：拒绝路径穿越节点
	if node.Name == ".." || node.Name == "." || node.Name == "" {
		return &domain.ErrWithPathTraversal{Path: node.Name}
	}

	targetPath := filepath.Join(targetDir, node.Name)

	// 确保还原路径仍在 targetDir 内（深度防御）
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(absTarget, absDir+string(filepath.Separator)) && absTarget != absDir {
		return &domain.ErrWithPathTraversal{Path: targetPath}
	}

	switch node.Type {
	case "file":
		if _, err := os.Stat(targetPath); err == nil {
			switch overwrite {
			case "skip":
				return nil
			case "force":
				// 继续覆盖
			case "ask":
				return domain.ErrAlreadyExists
			}
		}

		data, err := store.Read(node.Object)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, 0644)

	case "directory":
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return err
		}
		for _, child := range node.Children {
			if err := restoreTree(store, child, targetPath, overwrite); err != nil {
				return err
			}
		}
	}
	return nil
}
