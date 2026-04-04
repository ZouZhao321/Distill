package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

// AddAssetInput holds the input for adding an asset.
type AddAssetInput struct {
	Name    string
	Content []byte          // single-file mode
	Tree    *domain.TreeNode // directory/ZIP mode
	Source  string
}

// AddAssetUseCase handles importing a single file into the repository.
type AddAssetUseCase struct {
	repo  port.AssetRepository
	store port.ObjectStorage
}

// NewAddAssetUseCase creates a new AddAssetUseCase.
func NewAddAssetUseCase(repo port.AssetRepository, store port.ObjectStorage) *AddAssetUseCase {
	return &AddAssetUseCase{repo: repo, store: store}
}

// Execute imports a single file: compute hash, store object, create manifest, register ref.
func (uc *AddAssetUseCase) Execute(input AddAssetInput) (*domain.Manifest, error) {
	if len(input.Content) == 0 {
		return nil, domain.ErrEmptySource
	}

	// Check name uniqueness
	if _, err := uc.repo.GetRef(input.Name); err == nil {
		return nil, domain.ErrAlreadyExists
	}

	// Compute hash and store object
	hash := computeHash(input.Content)
	if err := uc.store.Write(hash, input.Content); err != nil {
		return nil, err
	}

	// Build manifest
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

	// Compute manifest identity hash
	manifest.Hash = computeHash([]byte(manifest.OriginalName + manifest.CreatedAt))

	// Save manifest and register ref
	if err := uc.repo.SaveManifest(manifest); err != nil {
		return nil, err
	}
	if err := uc.repo.CreateRef(domain.Ref{Name: input.Name, Manifest: manifest.Hash}); err != nil {
		return nil, err
	}

	return manifest, nil
}

// ExecuteForDirectory imports a directory tree: create manifest and register ref.
// Objects should already be stored by the caller (e.g., DirAdapter/ZipAdapter).
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

	return manifest, nil
}

// countTree counts files and total size in a TreeNode tree.
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

// computeHash returns the SHA-256 hex digest of data.
func computeHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
