package usecase

import (
	"testing"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

func TestRemove_Execute(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	// Prepare data
	addUC := NewAddAssetUseCase(repo, store)
	_, _ = addUC.Execute(AddAssetInput{Name: "to-remove", Content: []byte("data"), Source: "/a.txt"})

	// Remove
	uc := NewRemoveUseCase(repo)
	err := uc.Execute("to-remove")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Ref should be deleted
	_, err = repo.GetRef("to-remove")
	if err == nil {
		t.Error("ref should be deleted after remove")
	}

	// Objects should NOT be deleted
	if len(store.written) != 1 {
		t.Error("objects should NOT be deleted on remove")
	}
}

func TestRemove_Execute_NotFound(t *testing.T) {
	repo := newMockAssetRepo()

	uc := NewRemoveUseCase(repo)
	err := uc.Execute("nonexistent")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRemove_Execute_ThenList(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	addUC := NewAddAssetUseCase(repo, store)
	_, _ = addUC.Execute(AddAssetInput{Name: "keep-me", Content: []byte("keep"), Source: "/a.txt"})
	_, _ = addUC.Execute(AddAssetInput{Name: "remove-me", Content: []byte("remove"), Source: "/b.txt"})

	// Remove one
	uc := NewRemoveUseCase(repo)
	_ = uc.Execute("remove-me")

	// List should only show "keep-me"
	listUC := NewListAssetsUseCase(repo)
	items, err := listUC.Execute()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item after remove, got %d", len(items))
	}
	if len(items) > 0 && items[0].Name != "keep-me" {
		t.Errorf("expected 'keep-me', got %q", items[0].Name)
	}
}
