package usecase

import (
	"archive/zip"
	"bytes"
	"io"
	"log/slog"
	"os"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

// ExportUseCase 负责将资产导出为 ZIP 压缩包。
type ExportUseCase struct {
	repo  port.AssetRepository
	store port.ObjectStorage
}

// NewExportUseCase 创建新的 ExportUseCase 实例。
func NewExportUseCase(repo port.AssetRepository, store port.ObjectStorage) *ExportUseCase {
	return &ExportUseCase{repo: repo, store: store}
}

// Execute 将指定资产导出为 ZIP 文件并保存到 outputPath。
func (uc *ExportUseCase) Execute(name, outputPath string) error {
	ref, err := uc.repo.GetRef(name)
	if err != nil {
		return err
	}

	manifest, err := uc.repo.GetManifest(ref.Manifest)
	if err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	slog.Info("export started", "name", name, "output", outputPath)
	return addTreeToZip(uc.store, manifest.Tree, "", zw)
}

// addTreeToZip 递归将 TreeNode 树添加到 zip 写入器中。
func addTreeToZip(store port.ObjectStorage, node domain.TreeNode, prefix string, zw *zip.Writer) error {
	path := prefix + node.Name

	switch node.Type {
	case "file":
		w, err := zw.Create(path)
		if err != nil {
			return err
		}
		data, err := store.Read(node.Object)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, bytes.NewReader(data))
		return err

	case "directory":
		for _, child := range node.Children {
			if err := addTreeToZip(store, child, path+"/", zw); err != nil {
				return err
			}
		}
	}
	return nil
}
