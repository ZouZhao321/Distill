package adapter

import (
	"os"
	"path/filepath"
	"testing"
)

// --- DirAdapter tests ---

func TestDirAdapter_Adapt_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	store := newMockTestStore()
	ada := NewDirAdapter(store, false)

	tree, err := ada.Adapt(dir)
	if err != nil {
		t.Fatalf("Adapt failed: %v", err)
	}
	if tree.Type != "directory" {
		t.Errorf("root type = %q, want directory", tree.Type)
	}
	if len(tree.Children) != 0 {
		t.Errorf("empty dir should have 0 children, got %d", len(tree.Children))
	}
}

func TestDirAdapter_Adapt_WithFiles(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("world"), 0644)

	store := newMockTestStore()
	ada := NewDirAdapter(store, false)

	tree, err := ada.Adapt(dir)
	if err != nil {
		t.Fatalf("Adapt failed: %v", err)
	}

	if len(tree.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(tree.Children))
	}

	if len(store.data) != 2 {
		t.Errorf("expected 2 objects stored, got %d", len(store.data))
	}
}

func TestDirAdapter_Adapt_NestedDir(t *testing.T) {
	dir := t.TempDir()

	subDir := filepath.Join(dir, "sub")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested content"), 0644)

	store := newMockTestStore()
	ada := NewDirAdapter(store, false)

	tree, err := ada.Adapt(dir)
	if err != nil {
		t.Fatalf("Adapt failed: %v", err)
	}

	foundSub := false
	for _, child := range tree.Children {
		if child.Name == "sub" && child.Type == "directory" {
			foundSub = true
			if len(child.Children) != 1 {
				t.Errorf("sub dir should have 1 child, got %d", len(child.Children))
			}
		}
	}
	if !foundSub {
		t.Error("sub directory not found in tree")
	}
}

func TestDirAdapter_Adapt_SkipsSymlinks(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "regular.txt"), []byte("hello"), 0644)

	// Create a symlink (may fail on Windows without privileges)
	linkPath := filepath.Join(dir, "link.txt")
	_ = os.Symlink(filepath.Join(dir, "regular.txt"), linkPath)

	store := newMockTestStore()
	ada := NewDirAdapter(store, false)

	tree, err := ada.Adapt(dir)
	if err != nil {
		t.Fatalf("Adapt failed: %v", err)
	}

	foundRegular := false
	for _, child := range tree.Children {
		if child.Name == "regular.txt" {
			foundRegular = true
		}
	}
	if !foundRegular {
		t.Error("regular.txt should be in the tree")
	}

	// Symlinks should not appear as regular files
	for _, child := range tree.Children {
		if child.Name == "link.txt" && child.Type == "file" {
			t.Error("symlink should be skipped, not listed as file")
		}
	}
}

func TestDirAdapter_Adapt_NonexistentDir(t *testing.T) {
	store := newMockTestStore()
	ada := NewDirAdapter(store, false)

	_, err := ada.Adapt("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("should return error for nonexistent directory")
	}
}

func TestDirAdapter_Adapt_FileAsDir(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "notadir.txt")
	os.WriteFile(filePath, []byte("content"), 0644)

	store := newMockTestStore()
	ada := NewDirAdapter(store, false)

	_, err := ada.Adapt(filePath)
	if err == nil {
		t.Error("should return error when given a file instead of directory")
	}
}

// --- NormalizeCRLF tests ---

func TestNormalizeCRLF_CRLFToLF(t *testing.T) {
	input := []byte("line1\r\nline2\r\nline3")
	got := NormalizeCRLF(input)
	expected := []byte("line1\nline2\nline3")
	if string(got) != string(expected) {
		t.Errorf("got %q, want %q", string(got), string(expected))
	}
}

func TestNormalizeCRLF_AlreadyLF(t *testing.T) {
	input := []byte("line1\nline2\nline3")
	got := NormalizeCRLF(input)
	if string(got) != string(input) {
		t.Errorf("should not modify already-LF content, got %q", string(got))
	}
}

func TestNormalizeCRLF_Empty(t *testing.T) {
	input := []byte{}
	got := NormalizeCRLF(input)
	if len(got) != 0 {
		t.Errorf("empty input should return empty, got %q", string(got))
	}
}

// --- mock test store ---

type mockTestStore struct {
	data map[string][]byte
}

func newMockTestStore() *mockTestStore {
	return &mockTestStore{data: make(map[string][]byte)}
}

func (m *mockTestStore) Exists(hash string) (bool, error) {
	_, ok := m.data[hash]
	return ok, nil
}

func (m *mockTestStore) Read(hash string) ([]byte, error) {
	d, ok := m.data[hash]
	if !ok {
		return nil, os.ErrNotExist
	}
	return d, nil
}

func (m *mockTestStore) Write(hash string, data []byte) error {
	m.data[hash] = data
	return nil
}

func (m *mockTestStore) Delete(hash string) error {
	delete(m.data, hash)
	return nil
}

func (m *mockTestStore) Walk(fn func(hash string) error) error {
	for h := range m.data {
		if err := fn(h); err != nil {
			return err
		}
	}
	return nil
}
