package usecase

import (
	"runtime"
	"testing"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

func TestPerformance_LargeFileAdd_LowMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	// Generate 10MB of data
	size := 10 * 1024 * 1024 // 10MB
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	uc := NewAddAssetUseCase(repo, store)
	_, err := uc.Execute(AddAssetInput{
		Name:    "large-file",
		Content: data,
		Source:  "/large-file.bin",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	heapGrowth := int64(m2.HeapAlloc) - int64(m1.HeapAlloc)
	t.Logf("Heap growth: %d bytes for %d bytes file", heapGrowth, size)

	if heapGrowth > int64(size*3) {
		t.Errorf("heap growth %d is too large for input size %d", heapGrowth, size)
	}
}

func TestPerformance_LargeTreeAdd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	repo := newMockAssetRepo()
	store := newMockObjectStorage()

	// Create a tree with 100 files
	tree := createLargeTestTree(100)

	uc := NewAddAssetUseCase(repo, store)
	manifest, err := uc.ExecuteForDirectory(AddAssetInput{
		Name:   "large-tree",
		Tree:   tree,
		Source: "/large-tree",
	})
	if err != nil {
		t.Fatalf("ExecuteForDirectory failed: %v", err)
	}

	if manifest.FileCount != 100 {
		t.Errorf("expected 100 files, got %d", manifest.FileCount)
	}
}

func createLargeTestTree(fileCount int) *domain.TreeNode {
	root := &domain.TreeNode{
		Name: "large-tree",
		Type: "directory",
	}
	for i := 0; i < fileCount; i++ {
		name := string(rune('a'+i%26)) + string(rune('0'+i/26)) + ".txt"
		root.Children = append(root.Children, domain.TreeNode{
			Name:   name,
			Type:   "file",
			Size:   1024,
			Object: "hash_placeholder_" + name,
		})
	}
	return root
}
