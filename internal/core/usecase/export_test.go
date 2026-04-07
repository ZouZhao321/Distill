package usecase

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

// tempZipPath 在 t.TempDir() 下创建一个临时 .zip 文件，返回其路径。
// 测试结束后 t.TempDir() 会自动清理，无需手动删除。
func tempZipPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "export.zip")
}

// writeTempZip 在 t.TempDir() 下创建一个包含指定内容的临时 .zip 文件，返回其路径。
// 用于模拟"已存在的输出文件"场景。
func writeTempZip(t *testing.T, content []byte) string {
	t.Helper()
	p := tempZipPath(t)
	if err := os.WriteFile(p, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return p
}

func readAll(rc io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(rc)
	return buf.Bytes(), err
}

func TestExport_Execute_CreatesValidZip(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	addUC := NewAddAssetUseCase(repo, store)
	addUC.Execute(AddAssetInput{Name: "export-test.txt", Content: []byte("zip content"), Source: "/a.txt"})

	outputPath := tempZipPath(t)
	uc := NewExportUseCase(repo, store)
	err := uc.Execute("export-test.txt", outputPath, "skip")
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	r, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 1 {
		t.Fatalf("zip should contain 1 file, got %d", len(r.File))
	}

	rc, err := r.File[0].Open()
	if err != nil {
		t.Fatalf("failed to open file in zip: %v", err)
	}
	data, err := readAll(rc)
	rc.Close()
	if err != nil {
		t.Fatalf("failed to read file in zip: %v", err)
	}

	if string(data) != "zip content" {
		t.Errorf("zip content = %q, want %q", string(data), "zip content")
	}
}

func TestExport_Execute_NotFound(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	uc := NewExportUseCase(repo, store)
	err := uc.Execute("nonexistent", tempZipPath(t), "skip")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestExport_Execute_DirectoryTree(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	hash1 := computeHash([]byte("file1"))
	hash2 := computeHash([]byte("file2"))
	store.Write(hash1, []byte("file1"))
	store.Write(hash2, []byte("file2"))

	manifest := &domain.Manifest{
		Hash: "dir-export-hash", OriginalName: "my-dir", Status: "active",
		Tree: domain.TreeNode{
			Name: "my-dir", Type: "directory",
			Children: []domain.TreeNode{
				{Name: "sub", Type: "directory", Children: []domain.TreeNode{
					{Name: "file2.txt", Type: "file", Size: 5, Object: hash2},
				}},
				{Name: "file1.txt", Type: "file", Size: 5, Object: hash1},
			},
		},
	}
	repo.SaveManifest(manifest)
	repo.CreateRef(domain.Ref{Name: "my-dir", Manifest: manifest.Hash})

	outputPath := tempZipPath(t)
	uc := NewExportUseCase(repo, store)
	uc.Execute("my-dir", outputPath, "skip")

	r, _ := zip.OpenReader(outputPath)
	defer r.Close()

	if len(r.File) != 2 {
		t.Errorf("zip should contain 2 files, got %d", len(r.File))
	}
}

// Issue #25: export 覆盖保护
func TestExport_Execute_TargetExists(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	addUC := NewAddAssetUseCase(repo, store)
	addUC.Execute(AddAssetInput{Name: "asset", Content: []byte("new content"), Source: "/a.txt"})

	t.Run("skip", func(t *testing.T) {
		outputPath := writeTempZip(t, []byte("old content"))

		uc := NewExportUseCase(repo, store)
		err := uc.Execute("asset", outputPath, "skip")
		if err != nil {
			t.Fatalf("skip strategy should not error: %v", err)
		}

		got, _ := os.ReadFile(outputPath)
		if string(got) != "old content" {
			t.Errorf("file should not be overwritten with skip strategy, got %q", string(got))
		}
	})

	t.Run("force", func(t *testing.T) {
		outputPath := writeTempZip(t, []byte("old content"))

		uc := NewExportUseCase(repo, store)
		err := uc.Execute("asset", outputPath, "force")
		if err != nil {
			t.Fatalf("force strategy should not error: %v", err)
		}

		r, err := zip.OpenReader(outputPath)
		if err != nil {
			t.Fatalf("failed to open zip after force overwrite: %v", err)
		}
		defer r.Close()

		rc, err := r.File[0].Open()
		if err != nil {
			t.Fatalf("failed to open file in zip: %v", err)
		}
		data, err := readAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("failed to read file in zip: %v", err)
		}

		if string(data) != "new content" {
			t.Errorf("zip content = %q, want %q", string(data), "new content")
		}
	})

	t.Run("ask", func(t *testing.T) {
		outputPath := writeTempZip(t, []byte("old content"))

		uc := NewExportUseCase(repo, store)
		err := uc.Execute("asset", outputPath, "ask")
		if err != domain.ErrAlreadyExists {
			t.Errorf("ask strategy should return ErrAlreadyExists when file exists, got %v", err)
		}

		got, _ := os.ReadFile(outputPath)
		if string(got) != "old content" {
			t.Errorf("file should not be overwritten by ask strategy, got %q", string(got))
		}
	})
}
