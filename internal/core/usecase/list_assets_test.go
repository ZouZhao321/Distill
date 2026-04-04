package usecase

import (
	"testing"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

func createTestManifest(name string) *domain.Manifest {
	return &domain.Manifest{
		Hash:         computeHash([]byte(name)),
		OriginalName: name,
		OriginalPath: "/test/" + name,
		CreatedAt:    "2026-04-04T00:00:00Z",
		FileCount:    1,
		TotalSize:    100,
		StoredSize:   100,
		Status:       "active",
		Tree: domain.TreeNode{
			Name: name,
			Type: "directory",
			Children: []domain.TreeNode{
				{Name: "file.txt", Type: "file", Size: 100, Object: "ab/cdef1234"},
			},
		},
	}
}

func TestListAssets_Execute_Empty(t *testing.T) {
	repo := newMockAssetRepo()

	uc := NewListAssetsUseCase(repo)
	result, err := uc.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d items", len(result))
	}
}

func TestListAssets_Execute_WithItems(t *testing.T) {
	repo := newMockAssetRepo()

	m1 := createTestManifest("project-a")
	m2 := createTestManifest("project-b")
	repo.SaveManifest(m1)
	repo.SaveManifest(m2)
	repo.CreateRef(domain.Ref{Name: "project-a", Manifest: m1.Hash})
	repo.CreateRef(domain.Ref{Name: "project-b", Manifest: m2.Hash})

	uc := NewListAssetsUseCase(repo)
	result, err := uc.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}

	// Verify item fields
	names := map[string]bool{}
	for _, item := range result {
		names[item.Name] = true
		if item.FileCount != 1 {
			t.Errorf("item %s FileCount = %d, want 1", item.Name, item.FileCount)
		}
	}
	if !names["project-a"] || !names["project-b"] {
		t.Errorf("expected both project-a and project-b in results, got %v", names)
	}
}

func TestListAssets_Execute_SkipsBroken(t *testing.T) {
	repo := newMockAssetRepo()

	// Register a ref pointing to a nonexistent manifest
	repo.CreateRef(domain.Ref{Name: "broken", Manifest: "nonexistent_hash"})

	uc := NewListAssetsUseCase(repo)
	result, err := uc.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	// Should skip broken refs without error
	if len(result) != 0 {
		t.Errorf("expected 0 items (broken ref skipped), got %d", len(result))
	}
}
