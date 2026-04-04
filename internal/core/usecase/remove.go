package usecase

import (
	"log/slog"

	"github.com/ZouZhao321/distill/internal/core/port"
)

// RemoveUseCase removes an asset from the repository.
// It deletes the manifest and ref, but does NOT delete the underlying objects.
// Objects will be cleaned up by GC.
type RemoveUseCase struct {
	repo port.AssetRepository
}

// NewRemoveUseCase creates a new RemoveUseCase.
func NewRemoveUseCase(repo port.AssetRepository) *RemoveUseCase {
	return &RemoveUseCase{repo: repo}
}

// Execute removes the named asset by deleting its manifest and ref.
// Objects remain in storage until GC is run.
func (uc *RemoveUseCase) Execute(name string) error {
	ref, err := uc.repo.GetRef(name)
	if err != nil {
		return err
	}

	// Delete manifest
	if err := uc.repo.RemoveManifest(ref.Manifest); err != nil {
		return err
	}

	// Delete ref
	err = uc.repo.DeleteRef(name)
	if err == nil {
		slog.Info("asset removed", "name", name, "manifest_hash", ref.Manifest)
	}
	return err
}
