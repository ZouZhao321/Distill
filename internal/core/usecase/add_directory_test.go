package usecase

import (
	"testing"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

func TestAddAsset_ExecuteForDirectory(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewAddAssetUseCase(repo, store)

	tree := domain.TreeNode{
		Name: "my-project",
		Type: "directory",
		Children: []domain.TreeNode{
			{Name: "main.go", Type: "file", Size: 10, Object: "ab123"},
			{Name: "README.md", Type: "file", Size: 5, Object: "de456"},
			{
				Name: "utils",
				Type: "directory",
				Children: []domain.TreeNode{
					{Name: "helper.go", Type: "file", Size: 8, Object: "gh789"},
				},
			},
		},
	}

	input := AddAssetInput{
		Name:   "my-project",
		Tree:   &tree,
		Source: "/data/my-project",
	}

	manifest, err := uc.ExecuteForDirectory(input)
	if err != nil {
		t.Fatalf("ExecuteForDirectory failed: %v", err)
	}

	if manifest.FileCount != 3 {
		t.Errorf("FileCount = %d, want 3", manifest.FileCount)
	}
	if manifest.TotalSize != 23 {
		t.Errorf("TotalSize = %d, want 23", manifest.TotalSize)
	}
	if len(manifest.Tree.Children) != 3 {
		t.Errorf("root children = %d, want 3", len(manifest.Tree.Children))
	}
}

func TestAddAsset_ExecuteForDirectory_DuplicateName(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewAddAssetUseCase(repo, store)

	tree := domain.TreeNode{Name: "dup", Type: "directory"}
	_, _ = uc.ExecuteForDirectory(AddAssetInput{Name: "dup", Tree: &tree, Source: "/a"})

	_, err := uc.ExecuteForDirectory(AddAssetInput{Name: "dup", Tree: &tree, Source: "/b"})
	if err != domain.ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestAddAsset_ExecuteForDirectory_NilTree(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewAddAssetUseCase(repo, store)

	_, err := uc.ExecuteForDirectory(AddAssetInput{Name: "empty", Source: "/a"})
	if err != domain.ErrEmptySource {
		t.Errorf("expected ErrEmptySource, got %v", err)
	}
}
