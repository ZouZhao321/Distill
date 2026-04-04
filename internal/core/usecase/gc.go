package usecase

import (
	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

// GCUseCase performs garbage collection on the object store.
// It identifies and optionally removes objects not referenced by any manifest.
type GCUseCase struct {
	repo  port.AssetRepository
	store port.ObjectStorage
}

// NewGCUseCase creates a new GCUseCase.
func NewGCUseCase(repo port.AssetRepository, store port.ObjectStorage) *GCUseCase {
	return &GCUseCase{repo: repo, store: store}
}

// ExecuteDryRun returns a list of orphan object hashes without deleting them.
func (uc *GCUseCase) ExecuteDryRun() ([]string, error) {
	return uc.findOrphans()
}

// Execute removes all orphan objects and returns the count of cleaned objects.
func (uc *GCUseCase) Execute() (int, error) {
	orphans, err := uc.findOrphans()
	if err != nil {
		return 0, err
	}

	for _, hash := range orphans {
		uc.store.Delete(hash)
	}

	return len(orphans), nil
}

// findOrphans collects all objects not referenced by any manifest.
func (uc *GCUseCase) findOrphans() ([]string, error) {
	// Collect all object hashes referenced by manifests
	referenced := make(map[string]bool)
	manifests, err := uc.repo.ListManifests()
	if err != nil {
		return nil, err
	}
	for _, m := range manifests {
		collectObjects(m.Tree, referenced)
	}

	// Walk all stored objects and find orphans
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

// collectObjects recursively collects all object hashes from a TreeNode tree.
func collectObjects(node domain.TreeNode, set map[string]bool) {
	if node.Type == "file" && node.Object != "" {
		set[node.Object] = true
	}
	for _, child := range node.Children {
		collectObjects(child, set)
	}
}
