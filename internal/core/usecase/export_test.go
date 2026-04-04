package usecase

import (
	"archive/zip"
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

func TestExport_Execute_CreatesValidZip(t *testing.T) {
	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	addUC := NewAddAssetUseCase(repo, store)
	addUC.Execute(AddAssetInput{Name: "export-test.txt", Content: []byte("zip content"), Source: "/a.txt"})

	outputPath := filepath.Join(t.TempDir(), "output.zip")
	uc := NewExportUseCase(repo, store)
	err := uc.Execute("export-test.txt", outputPath)
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
	err := uc.Execute("nonexistent", filepath.Join(t.TempDir(), "out.zip"))
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

	outputPath := filepath.Join(t.TempDir(), "dir-output.zip")
	uc := NewExportUseCase(repo, store)
	uc.Execute("my-dir", outputPath)

	r, _ := zip.OpenReader(outputPath)
	defer r.Close()

	if len(r.File) != 2 {
		t.Errorf("zip should contain 2 files, got %d", len(r.File))
	}
}

func readAll(rc io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(rc)
	return buf.Bytes(), err
}
