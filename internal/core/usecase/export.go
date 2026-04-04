package usecase

import (
	"archive/zip"
	"bytes"
	"io"
	"os"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

// ExportUseCase exports an asset as a ZIP file.
type ExportUseCase struct {
	repo  port.AssetRepository
	store port.ObjectStorage
}

// NewExportUseCase creates a new ExportUseCase.
func NewExportUseCase(repo port.AssetRepository, store port.ObjectStorage) *ExportUseCase {
	return &ExportUseCase{repo: repo, store: store}
}

// Execute exports the named asset to outputPath as a ZIP file.
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

	return addTreeToZip(uc.store, manifest.Tree, "", zw)
}

// addTreeToZip recursively adds a TreeNode tree to a zip writer.
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
