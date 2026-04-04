package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

func setupManifestStore(t *testing.T) (*ManifestStore, string) {
	t.Helper()
	dir := t.TempDir()
	manifestsDir := filepath.Join(dir, "manifests")
	refsPath := filepath.Join(dir, "refs.json")
	os.MkdirAll(manifestsDir, 0755)
	return NewManifestStore(manifestsDir, refsPath), dir
}

func createTestManifest(name string) *domain.Manifest {
	return &domain.Manifest{
		Hash:         ComputeHash([]byte(name)),
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

func TestManifestStore_SaveAndGetManifest(t *testing.T) {
	store, dir := setupManifestStore(t)
	m := createTestManifest("test-project")

	err := store.SaveManifest(m)
	if err != nil {
		t.Fatalf("SaveManifest failed: %v", err)
	}

	// Verify file exists
	expectedFile := filepath.Join(dir, "manifests", m.Hash+"manifest.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("manifest file not found at %s", expectedFile)
	}

	// Read and verify
	got, err := store.GetManifest(m.Hash)
	if err != nil {
		t.Fatalf("GetManifest failed: %v", err)
	}
	if got.OriginalName != "test-project" {
		t.Errorf("OriginalName = %q, want %q", got.OriginalName, "test-project")
	}
	if got.FileCount != 1 {
		t.Errorf("FileCount = %d, want 1", got.FileCount)
	}
}

func TestManifestStore_GetManifest_NotFound(t *testing.T) {
	store, _ := setupManifestStore(t)

	_, err := store.GetManifest("nonexistent00000000000000000000000000000000000000000000000000000000")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestManifestStore_ListManifests(t *testing.T) {
	store, _ := setupManifestStore(t)

	store.SaveManifest(createTestManifest("project-a"))
	store.SaveManifest(createTestManifest("project-b"))

	list, err := store.ListManifests()
	if err != nil {
		t.Fatalf("ListManifests failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("ListManifests returned %d items, want 2", len(list))
	}
}

func TestManifestStore_RemoveManifest(t *testing.T) {
	store, dir := setupManifestStore(t)
	m := createTestManifest("to-remove")
	store.SaveManifest(m)

	err := store.RemoveManifest(m.Hash)
	if err != nil {
		t.Fatalf("RemoveManifest failed: %v", err)
	}

	// File should be deleted
	expectedFile := filepath.Join(dir, "manifests", m.Hash+"manifest.json")
	if _, err := os.Stat(expectedFile); !os.IsNotExist(err) {
		t.Error("manifest file should be deleted")
	}

	// GetManifest should return error
	_, err = store.GetManifest(m.Hash)
	if err == nil {
		t.Error("GetManifest after remove should return error")
	}
}

func TestManifestStore_RefOperations(t *testing.T) {
	store, _ := setupManifestStore(t)

	// Create Ref
	err := store.CreateRef(domain.Ref{Name: "my-app", Manifest: "hash123"})
	if err != nil {
		t.Fatalf("CreateRef failed: %v", err)
	}

	// Read Ref
	ref, err := store.GetRef("my-app")
	if err != nil {
		t.Fatalf("GetRef failed: %v", err)
	}
	if ref.Manifest != "hash123" {
		t.Errorf("Ref.Manifest = %q, want %q", ref.Manifest, "hash123")
	}

	// List all Refs
	refs, _ := store.ListRefs()
	if len(refs) != 1 {
		t.Errorf("ListRefs returned %d items, want 1", len(refs))
	}

	// Delete Ref
	err = store.DeleteRef("my-app")
	if err != nil {
		t.Fatalf("DeleteRef failed: %v", err)
	}

	// Read deleted Ref should error
	_, err = store.GetRef("my-app")
	if err == nil {
		t.Error("GetRef after delete should return error")
	}
}

func TestManifestStore_CreateRef_DuplicateName(t *testing.T) {
	store, _ := setupManifestStore(t)

	store.CreateRef(domain.Ref{Name: "dup", Manifest: "hash1"})

	err := store.CreateRef(domain.Ref{Name: "dup", Manifest: "hash2"})
	if err == nil {
		t.Error("duplicate ref name should return error")
	}
}

func TestManifestStore_DeleteRef_NotFound(t *testing.T) {
	store, _ := setupManifestStore(t)

	err := store.DeleteRef("nonexistent")
	if err == nil {
		t.Error("DeleteRef for nonexistent ref should return error")
	}
}

func TestManifestStore_refsJsonPersistence(t *testing.T) {
	store, _ := setupManifestStore(t)

	store.CreateRef(domain.Ref{Name: "persist-test", Manifest: "abc123"})

	// Recreate store to simulate restart
	refsPath := store.RefsPath()
	store2 := NewManifestStore(store.ManifestsDir(), refsPath)

	ref, err := store2.GetRef("persist-test")
	if err != nil {
		t.Fatalf("ref not persisted across store instances: %v", err)
	}
	if ref.Manifest != "abc123" {
		t.Errorf("persisted Manifest = %q, want %q", ref.Manifest, "abc123")
	}
}

func TestManifestStore_SavedJsonFormat(t *testing.T) {
	store, dir := setupManifestStore(t)
	m := createTestManifest("json-test")
	store.SaveManifest(m)

	data, _ := os.ReadFile(filepath.Join(dir, "manifests", m.Hash+"manifest.json"))

	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	if parsed["original_name"] != "json-test" {
		t.Errorf("JSON original_name = %v, want %q", parsed["original_name"], "json-test")
	}
	if parsed["status"] != "active" {
		t.Errorf("JSON status = %v, want %q", parsed["status"], "active")
	}
	tree, ok := parsed["tree"].(map[string]interface{})
	if !ok || tree["name"] != "json-test" {
		t.Error("JSON tree structure incorrect")
	}
}

func TestManifestStore_ListManifests_Empty(t *testing.T) {
	store, _ := setupManifestStore(t)

	list, err := store.ListManifests()
	if err != nil {
		t.Fatalf("ListManifests failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 manifests, got %d", len(list))
	}
}

func TestManifestStore_GetRefNameByHash(t *testing.T) {
	store, _ := setupManifestStore(t)

	store.CreateRef(domain.Ref{Name: "find-me", Manifest: "abc123"})

	name, err := store.GetRefNameByHash("abc123")
	if err != nil {
		t.Fatalf("GetRefNameByHash failed: %v", err)
	}
	if name != "find-me" {
		t.Errorf("name = %q, want %q", name, "find-me")
	}
}

func TestManifestStore_GetRefNameByHash_NotFound(t *testing.T) {
	store, _ := setupManifestStore(t)

	_, err := store.GetRefNameByHash("nonexistent")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ManifestStore must implement port.AssetRepository
var _ port.AssetRepository = (*ManifestStore)(nil)
