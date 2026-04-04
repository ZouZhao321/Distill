package usecase

import (
	"testing"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

func TestAddAsset_Execute_SingleFile(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewAddAssetUseCase(repo, store)

	input := AddAssetInput{
		Name:    "hello.txt",
		Content: []byte("hello world"),
		Source:  "/tmp/hello.txt",
	}

	manifest, err := uc.Execute(input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify object stored
	if len(store.written) != 1 {
		t.Errorf("expected 1 object stored, got %d", len(store.written))
	}

	// Verify manifest
	if manifest.OriginalName != "hello.txt" {
		t.Errorf("OriginalName = %q, want %q", manifest.OriginalName, "hello.txt")
	}
	if manifest.FileCount != 1 {
		t.Errorf("FileCount = %d, want 1", manifest.FileCount)
	}
	if manifest.TotalSize != 11 {
		t.Errorf("TotalSize = %d, want 11", manifest.TotalSize)
	}

	// Verify tree
	if manifest.Tree.Name != "hello.txt" {
		t.Errorf("Tree.Name = %q, want %q", manifest.Tree.Name, "hello.txt")
	}
	if manifest.Tree.Type != "file" {
		t.Errorf("Tree.Type = %q, want %q", manifest.Tree.Type, "file")
	}
	if manifest.Tree.Size != 11 {
		t.Errorf("Tree.Size = %d, want 11", manifest.Tree.Size)
	}
	if manifest.Tree.Object == "" {
		t.Error("Tree.Object should not be empty")
	}
	if len(manifest.Tree.Object) != 64 {
		t.Errorf("Tree.Object should be 64-char SHA-256 hex, got %d chars", len(manifest.Tree.Object))
	}
}

func TestAddAsset_Execute_Dedup(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewAddAssetUseCase(repo, store)

	_, _ = uc.Execute(AddAssetInput{Name: "a.txt", Content: []byte("same content"), Source: "/a.txt"})
	_, _ = uc.Execute(AddAssetInput{Name: "b.txt", Content: []byte("same content"), Source: "/b.txt"})

	// Should store only 1 object (dedup)
	if len(store.written) != 1 {
		t.Errorf("expected 1 object (dedup), got %d", len(store.written))
	}

	// Should have 2 manifests
	if len(repo.manifests) != 2 {
		t.Errorf("expected 2 manifests, got %d", len(repo.manifests))
	}
}

func TestAddAsset_Execute_DuplicateName(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewAddAssetUseCase(repo, store)

	_, _ = uc.Execute(AddAssetInput{Name: "dup.txt", Content: []byte("first"), Source: "/a.txt"})

	_, err := uc.Execute(AddAssetInput{Name: "dup.txt", Content: []byte("second"), Source: "/b.txt"})
	if err != domain.ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestAddAsset_Execute_EmptyContent(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewAddAssetUseCase(repo, store)

	_, err := uc.Execute(AddAssetInput{Name: "empty.txt", Content: []byte{}, Source: "/empty.txt"})
	if err != domain.ErrEmptySource {
		t.Errorf("expected ErrEmptySource, got %v", err)
	}
}

func TestAddAsset_Execute_RefRegistered(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewAddAssetUseCase(repo, store)

	manifest, _ := uc.Execute(AddAssetInput{Name: "my-asset", Content: []byte("data"), Source: "/data"})

	ref, err := repo.GetRef("my-asset")
	if err != nil {
		t.Fatalf("ref not found: %v", err)
	}
	if ref.Manifest != manifest.Hash {
		t.Errorf("ref.Manifest = %q, want %q", ref.Manifest, manifest.Hash)
	}
}

func TestAddAsset_Execute_StatusActive(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewAddAssetUseCase(repo, store)

	manifest, _ := uc.Execute(AddAssetInput{Name: "status-test", Content: []byte("x"), Source: "/x"})

	if manifest.Status != "active" {
		t.Errorf("status = %q, want %q", manifest.Status, "active")
	}
}
