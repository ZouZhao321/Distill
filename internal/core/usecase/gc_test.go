package usecase

import (
	"testing"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

func TestGC_ExecuteDryRun(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	// Add an orphan object (not referenced by any manifest)
	store.Write("orphan00000000000000000000000000000000000000000000000000000001", []byte("orphan data"))

	uc := NewGCUseCase(repo, store)
	orphans, err := uc.ExecuteDryRun()
	if err != nil {
		t.Fatalf("ExecuteDryRun failed: %v", err)
	}

	found := false
	for _, h := range orphans {
		if h == "orphan00000000000000000000000000000000000000000000000000000001" {
			found = true
		}
	}
	if !found {
		t.Error("orphan object should be detected in dry run")
	}

	// Dry run should NOT delete
	exists, _ := store.Exists("orphan00000000000000000000000000000000000000000000000000000001")
	if !exists {
		t.Error("dry run should not delete objects")
	}
}

func TestGC_Execute(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	// Add a referenced object
	hash := computeHash([]byte("referenced"))
	store.Write(hash, []byte("referenced"))
	repo.SaveManifest(&domain.Manifest{
		Hash: hash,
		Tree: domain.TreeNode{Type: "file", Object: hash},
	})

	// Add an orphan object
	orphanHash := "orphan00000000000000000000000000000000000000000000000000000001"
	store.Write(orphanHash, []byte("orphan"))

	uc := NewGCUseCase(repo, store)
	cleaned, err := uc.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if cleaned != 1 {
		t.Errorf("expected 1 object cleaned, got %d", cleaned)
	}

	// Orphan should be deleted
	exists, _ := store.Exists(orphanHash)
	if exists {
		t.Error("orphan should be deleted after GC")
	}

	// Referenced object should still exist
	exists, _ = store.Exists(hash)
	if !exists {
		t.Error("referenced object should NOT be deleted")
	}
}

func TestGC_Execute_NoOrphans(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	// Add a referenced object
	hash := computeHash([]byte("kept"))
	store.Write(hash, []byte("kept"))
	repo.SaveManifest(&domain.Manifest{
		Hash: hash,
		Tree: domain.TreeNode{Type: "file", Object: hash},
	})

	uc := NewGCUseCase(repo, store)
	cleaned, err := uc.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if cleaned != 0 {
		t.Errorf("expected 0 objects cleaned, got %d", cleaned)
	}
}

func TestGC_ExecuteWithDirectory(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	// Simulate a directory manifest with multiple files
	hash1 := computeHash([]byte("file1"))
	hash2 := computeHash([]byte("file2"))
	store.Write(hash1, []byte("file1"))
	store.Write(hash2, []byte("file2"))

	repo.SaveManifest(&domain.Manifest{
		Hash: computeHash([]byte("dir-manifest")),
		Tree: domain.TreeNode{
			Type: "directory",
			Children: []domain.TreeNode{
				{Type: "file", Object: hash1},
				{Type: "file", Object: hash2},
			},
		},
	})

	// Add orphan
	store.Write("orphan00000000000000000000000000000000000000000000000000000002", []byte("x"))

	uc := NewGCUseCase(repo, store)
	cleaned, _ := uc.Execute()
	if cleaned != 1 {
		t.Errorf("expected 1 orphan cleaned, got %d", cleaned)
	}

	// Both referenced files should still exist
	if exists, _ := store.Exists(hash1); !exists {
		t.Error("hash1 should still exist")
	}
	if exists, _ := store.Exists(hash2); !exists {
		t.Error("hash2 should still exist")
	}
}
