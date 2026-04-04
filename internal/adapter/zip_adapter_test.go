package adapter

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func createTestZip(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(content))
	}
	zw.Close()
	f.Close()
	return path
}

func TestZipAdapter_Adapt(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"main.go":    "package main",
		"README.md":  "# Hello",
		"lib/util.go": "package lib",
	})

	store := newMockTestStore()
	ada := NewZipAdapter(store, false)

	tree, err := ada.Adapt(zipPath)
	if err != nil {
		t.Fatalf("Adapt failed: %v", err)
	}

	if tree.Type != "directory" {
		t.Errorf("root type = %q, want directory", tree.Type)
	}
	if len(tree.Children) != 3 {
		t.Errorf("children count = %d, want 3", len(tree.Children))
	}

	// Verify objects were stored
	if len(store.data) != 3 {
		t.Errorf("expected 3 objects stored, got %d", len(store.data))
	}
}

func TestZipAdapter_Adapt_NestedPaths(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"src/main.go":   "package main",
		"src/lib/util.go": "package lib",
	})

	store := newMockTestStore()
	ada := NewZipAdapter(store, false)

	tree, err := ada.Adapt(zipPath)
	if err != nil {
		t.Fatalf("Adapt failed: %v", err)
	}

	// Should have a "src" directory with 2 children
	foundSrc := false
	for _, child := range tree.Children {
		if child.Name == "src" && child.Type == "directory" {
			foundSrc = true
			if len(child.Children) != 2 {
				t.Errorf("src dir should have 2 children, got %d", len(child.Children))
			}
		}
	}
	if !foundSrc {
		t.Error("src directory not found in tree")
	}
}

func TestZipAdapter_Adapt_Empty(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{})

	store := newMockTestStore()
	ada := NewZipAdapter(store, false)

	tree, err := ada.Adapt(zipPath)
	if err != nil {
		t.Fatalf("Adapt failed: %v", err)
	}

	if len(tree.Children) != 0 {
		t.Errorf("empty zip should have 0 children, got %d", len(tree.Children))
	}
}

func TestZipAdapter_Adapt_WithCRLF(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"main.go": "line1\r\nline2\r\n",
	})

	store := newMockTestStore()
	ada := NewZipAdapter(store, true) // normalizeCRLF = true

	_, err := ada.Adapt(zipPath)
	if err != nil {
		t.Fatalf("Adapt failed: %v", err)
	}

	// Check stored content is normalized
	for _, data := range store.data {
		if string(data) != "line1\nline2\n" {
			t.Errorf("CRLF not normalized, got %q", string(data))
		}
	}
}
