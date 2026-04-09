package usecase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

func TestCheckout_Execute_SingleFile(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	addUC := NewAddAssetUseCase(repo, store)
	_, _ = addUC.Execute(AddAssetInput{
		Name:    "restore-me.txt",
		Content: []byte("checkout test content"),
		Source:  "/original/restore-me.txt",
	})

	outputDir := t.TempDir()
	checkoutUC := NewCheckoutUseCase(repo, store)
	err := checkoutUC.Execute("restore-me.txt", outputDir, "skip")
	if err != nil {
		t.Fatalf("Checkout failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(outputDir, "restore-me.txt"))
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if string(got) != "checkout test content" {
		t.Errorf("file content = %q, want %q", string(got), "checkout test content")
	}
}

func TestCheckout_Execute_NotFound(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewCheckoutUseCase(repo, store)
	err := uc.Execute("nonexistent", t.TempDir(), "skip")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCheckout_Execute_TargetExists_Skip(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	addUC := NewAddAssetUseCase(repo, store)
	addUC.Execute(AddAssetInput{Name: "skip-test.txt", Content: []byte("original"), Source: "/a.txt"})

	outputDir := t.TempDir()
	os.WriteFile(filepath.Join(outputDir, "skip-test.txt"), []byte("existing"), 0644)

	uc := NewCheckoutUseCase(repo, store)
	err := uc.Execute("skip-test.txt", outputDir, "skip")
	if err != nil {
		t.Fatalf("skip strategy should not error: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(outputDir, "skip-test.txt"))
	if string(got) != "existing" {
		t.Errorf("file should not be overwritten, got %q", string(got))
	}
}

func TestCheckout_Execute_TargetExists_Force(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	addUC := NewAddAssetUseCase(repo, store)
	addUC.Execute(AddAssetInput{Name: "force-test.txt", Content: []byte("new content"), Source: "/a.txt"})

	outputDir := t.TempDir()
	os.WriteFile(filepath.Join(outputDir, "force-test.txt"), []byte("old"), 0644)

	uc := NewCheckoutUseCase(repo, store)
	uc.Execute("force-test.txt", outputDir, "force")

	got, _ := os.ReadFile(filepath.Join(outputDir, "force-test.txt"))
	if string(got) != "new content" {
		t.Errorf("file should be overwritten, got %q", string(got))
	}
}

func TestCheckout_Execute_DirectoryTree(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	hash1 := domain.ComputeHash([]byte("file1 content"))
	hash2 := domain.ComputeHash([]byte("file2 content"))
	store.Write(hash1, []byte("file1 content"))
	store.Write(hash2, []byte("file2 content"))

	manifest := &domain.Manifest{
		Hash:         "dir-manifest-hash",
		OriginalName: "my-dir",
		Status:       "active",
		FileCount:    2,
		TotalSize:    26,
		Tree: domain.TreeNode{
			Name: "my-dir",
			Type: "directory",
			Children: []domain.TreeNode{
				{Name: "sub", Type: "directory", Children: []domain.TreeNode{
					{Name: "file2.txt", Type: "file", Size: 13, Object: hash2},
				}},
				{Name: "file1.txt", Type: "file", Size: 13, Object: hash1},
			},
		},
	}
	repo.SaveManifest(manifest)
	repo.CreateRef(domain.Ref{Name: "my-dir", Manifest: manifest.Hash})

	outputDir := t.TempDir()
	uc := NewCheckoutUseCase(repo, store)
	err := uc.Execute("my-dir", outputDir, "force")
	if err != nil {
		t.Fatalf("Checkout directory failed: %v", err)
	}

	got1, _ := os.ReadFile(filepath.Join(outputDir, "my-dir", "file1.txt"))
	if string(got1) != "file1 content" {
		t.Errorf("file1 content = %q", string(got1))
	}
	got2, _ := os.ReadFile(filepath.Join(outputDir, "my-dir", "sub", "file2.txt"))
	if string(got2) != "file2 content" {
		t.Errorf("file2 content = %q", string(got2))
	}
}

// --- Issue #57: checkout 防御性路径穿越校验 ---

func TestCheckout_Execute_RejectsPathTraversalInTree(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	hash := domain.ComputeHash([]byte("evil"))
	store.Write(hash, []byte("evil"))

	manifest := &domain.Manifest{
		Hash:         "evil-manifest-hash",
		OriginalName: "evil-asset",
		Status:       "active",
		FileCount:    1,
		TotalSize:    4,
		Tree: domain.TreeNode{
			Name: "evil-asset",
			Type: "directory",
			Children: []domain.TreeNode{
				{Name: "..", Type: "directory", Children: []domain.TreeNode{
					{Name: "escaped.txt", Type: "file", Size: 4, Object: hash},
				}},
			},
		},
	}
	repo.SaveManifest(manifest)
	repo.CreateRef(domain.Ref{Name: "evil-asset", Manifest: manifest.Hash})

	outputDir := t.TempDir()
	uc := NewCheckoutUseCase(repo, store)
	err := uc.Execute("evil-asset", outputDir, "force")
	if err == nil {
		t.Fatal("expected error for path traversal in tree, got nil")
	}
}

func TestCheckout_Execute_RejectsDotDotFile(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	hash := domain.ComputeHash([]byte("evil"))
	store.Write(hash, []byte("evil"))

	manifest := &domain.Manifest{
		Hash:         "evil2-manifest-hash",
		OriginalName: "evil2-asset",
		Status:       "active",
		FileCount:    1,
		TotalSize:    4,
		Tree: domain.TreeNode{
			Name: "evil2-asset",
			Type: "directory",
			Children: []domain.TreeNode{
				{Name: "..evil.txt", Type: "file", Size: 4, Object: hash},
			},
		},
	}
	repo.SaveManifest(manifest)
	repo.CreateRef(domain.Ref{Name: "evil2-asset", Manifest: manifest.Hash})

	outputDir := t.TempDir()
	uc := NewCheckoutUseCase(repo, store)
	err := uc.Execute("evil2-asset", outputDir, "force")
	// ..evil.txt 不是路径穿越（不以分隔符分隔的 ".."），应正常还原
	if err != nil {
		t.Fatalf("file named '..evil.txt' should be allowed: %v", err)
	}
}
