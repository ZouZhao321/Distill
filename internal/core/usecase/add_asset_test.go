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

// --- Issue #22: 单文件导入应规范化 CRLF ---

func TestAddAsset_Execute_NormalizeCRLF(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewAddAssetUseCase(repo, store)

	crlfContent := []byte("line1\r\nline2\r\nline3")
	_, err := uc.Execute(AddAssetInput{
		Name:    "crlf-test.txt",
		Content: crlfContent,
		Source:  "/crlf-test.txt",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 验证存储的内容已规范化为 LF
	var storedHash string
	for hash, data := range store.written {
		storedHash = hash
		expected := []byte("line1\nline2\nline3")
		if string(data) != string(expected) {
			t.Errorf("stored content not normalized: got %q, want %q", string(data), string(expected))
		}
	}

	// 验证 manifest 中的 size 反映规范化后的大小
	ref, _ := repo.GetRef("crlf-test.txt")
	manifest, _ := repo.GetManifest(ref.Manifest)
	expectedSize := int64(len("line1\nline2\nline3"))
	if manifest.TotalSize != expectedSize {
		t.Errorf("TotalSize = %d, want %d (after CRLF normalization)", manifest.TotalSize, expectedSize)
	}
	if manifest.Tree.Size != expectedSize {
		t.Errorf("Tree.Size = %d, want %d (after CRLF normalization)", manifest.Tree.Size, expectedSize)
	}
	if manifest.Tree.Object != storedHash {
		t.Errorf("Tree.Object = %q, want %q", manifest.Tree.Object, storedHash)
	}
}

func TestAddAsset_Execute_CRLFDedup(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewAddAssetUseCase(repo, store)

	// 先导入 CRLF 内容
	_, err := uc.Execute(AddAssetInput{
		Name:    "a.txt",
		Content: []byte("hello\r\nworld\r\n"),
		Source:  "/a.txt",
	})
	if err != nil {
		t.Fatalf("Execute CRLF failed: %v", err)
	}

	// 再导入同样的 LF 内容（规范化后应相同）
	_, err = uc.Execute(AddAssetInput{
		Name:    "b.txt",
		Content: []byte("hello\nworld\n"),
		Source:  "/b.txt",
	})
	if err != nil {
		t.Fatalf("Execute LF failed: %v", err)
	}

	// 应只存储 1 个对象（去重生效）
	if len(store.written) != 1 {
		t.Errorf("expected 1 object (CRLF and LF should dedup after normalization), got %d", len(store.written))
	}
}

func TestAddAsset_Execute_AlreadyLFUnchanged(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewAddAssetUseCase(repo, store)

	lfContent := []byte("line1\nline2\nline3")
	_, err := uc.Execute(AddAssetInput{
		Name:    "lf-test.txt",
		Content: lfContent,
		Source:  "/lf-test.txt",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 验证 LF 内容未被修改
	for _, data := range store.written {
		if string(data) != string(lfContent) {
			t.Errorf("LF content was modified: got %q, want %q", string(data), string(lfContent))
		}
	}
}
