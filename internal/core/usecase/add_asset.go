// Package usecase 实现资产管理的核心业务用例。
// 包括添加、查询、检出、导出、移除和垃圾回收。
package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

// AddAssetInput 保存添加资产所需的输入参数。
type AddAssetInput struct {
	Name    string
	Content []byte           // 单文件模式
	Tree    *domain.TreeNode // 目录/ZIP 模式
	Source  string
}

// AddAssetUseCase 负责将资产导入仓库。
type AddAssetUseCase struct {
	repo  port.AssetRepository
	store port.ObjectStorage
}

// NewAddAssetUseCase 创建新的 AddAssetUseCase 实例。
func NewAddAssetUseCase(repo port.AssetRepository, store port.ObjectStorage) *AddAssetUseCase {
	return &AddAssetUseCase{repo: repo, store: store}
}

// Execute 导入单个文件：计算哈希、存储对象、创建清单、注册引用。
func (uc *AddAssetUseCase) Execute(input AddAssetInput) (*domain.Manifest, error) {
	if len(input.Content) == 0 {
		return nil, domain.ErrEmptySource
	}

	// 检查名称是否已存在
	if _, err := uc.repo.GetRef(input.Name); err == nil {
		return nil, domain.ErrAlreadyExists
	}

	// 计算哈希并存储对象
	hash := computeHash(input.Content)
	if err := uc.store.Write(hash, input.Content); err != nil {
		return nil, err
	}

	// 构建清单
	now := time.Now().UTC().Format(time.RFC3339)
	manifest := &domain.Manifest{
		OriginalName: input.Name,
		OriginalPath: input.Source,
		CreatedAt:    now,
		FileCount:    1,
		TotalSize:    int64(len(input.Content)),
		StoredSize:   int64(len(input.Content)),
		Status:       "active",
		Tree: domain.TreeNode{
			Name:   input.Name,
			Type:   "file",
			Size:   int64(len(input.Content)),
			Object: hash,
		},
	}

	// 计算清单身份哈希
	manifest.Hash = computeHash([]byte(manifest.OriginalName + manifest.CreatedAt))

	// 保存清单并注册引用
	if err := uc.repo.SaveManifest(manifest); err != nil {
		return nil, err
	}
	if err := uc.repo.CreateRef(domain.Ref{Name: input.Name, Manifest: manifest.Hash}); err != nil {
		return nil, err
	}

	slog.Info("asset added", "name", input.Name, "hash", manifest.Hash, "files", manifest.FileCount, "size", manifest.TotalSize)
	return manifest, nil
}

// ExecuteForDirectory 导入目录树：创建清单并注册引用。
// 调用方（如 DirAdapter/ZipAdapter）应已将对象存储完毕。
func (uc *AddAssetUseCase) ExecuteForDirectory(input AddAssetInput) (*domain.Manifest, error) {
	if input.Tree == nil {
		return nil, domain.ErrEmptySource
	}

	if _, err := uc.repo.GetRef(input.Name); err == nil {
		return nil, domain.ErrAlreadyExists
	}

	fileCount, totalSize := countTree(input.Tree)

	now := time.Now().UTC().Format(time.RFC3339)
	manifest := &domain.Manifest{
		OriginalName: input.Name,
		OriginalPath: input.Source,
		CreatedAt:    now,
		FileCount:    fileCount,
		TotalSize:    totalSize,
		StoredSize:   totalSize,
		Status:       "active",
		Tree:         *input.Tree,
	}

	manifest.Hash = computeHash([]byte(input.Name + now))

	if err := uc.repo.SaveManifest(manifest); err != nil {
		return nil, err
	}
	if err := uc.repo.CreateRef(domain.Ref{Name: input.Name, Manifest: manifest.Hash}); err != nil {
		return nil, err
	}

	slog.Info("directory asset added", "name", input.Name, "hash", manifest.Hash, "files", manifest.FileCount, "size", manifest.TotalSize)
	return manifest, nil
}

// countTree 递归统计 TreeNode 树中的文件数量和总大小。
func countTree(node *domain.TreeNode) (int, int64) {
	if node.Type == "file" {
		return 1, node.Size
	}
	count := 0
	var size int64
	for i := range node.Children {
		c, s := countTree(&node.Children[i])
		count += c
		size += s
	}
	return count, size
}

// computeHash 返回数据的 SHA-256 十六进制摘要。
func computeHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
