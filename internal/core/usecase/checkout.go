package usecase

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

// CheckoutUseCase restores assets from the repository to a target directory.
type CheckoutUseCase struct {
	repo  port.AssetRepository
	store port.ObjectStorage
}

// NewCheckoutUseCase creates a new CheckoutUseCase.
func NewCheckoutUseCase(repo port.AssetRepository, store port.ObjectStorage) *CheckoutUseCase {
	return &CheckoutUseCase{repo: repo, store: store}
}

// Execute restores the named asset to outputDir.
// overwrite controls behavior when files already exist: "skip", "force", or "ask".
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

// restoreTree recursively restores a TreeNode tree to the filesystem.
func restoreTree(store port.ObjectStorage, node domain.TreeNode, targetDir, overwrite string) error {
	targetPath := filepath.Join(targetDir, node.Name)

	switch node.Type {
	case "file":
		if _, err := os.Stat(targetPath); err == nil {
			switch overwrite {
			case "skip":
				return nil
			case "force":
				// proceed to overwrite
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
