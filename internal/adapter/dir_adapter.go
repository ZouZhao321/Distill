package adapter

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

// DirAdapter adapts a filesystem directory into a TreeNode tree.
// It stores all file objects via the ObjectStorage during adaptation.
type DirAdapter struct {
	Store         port.ObjectStorage
	NormalizeCRLF bool
}

// NewDirAdapter creates a new DirAdapter.
func NewDirAdapter(store port.ObjectStorage, normalizeCRLF bool) *DirAdapter {
	return &DirAdapter{Store: store, NormalizeCRLF: normalizeCRLF}
}

// Adapt reads the directory at dirPath and returns a TreeNode tree.
// All file contents are stored in the ObjectStorage.
func (a *DirAdapter) Adapt(dirPath string) (*domain.TreeNode, error) {
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}
	return a.buildTree(info.Name(), dirPath)
}

func (a *DirAdapter) buildTree(name, path string) (*domain.TreeNode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if a.NormalizeCRLF {
			content = NormalizeCRLF(content)
		}
		hash := sha256.Sum256(content)
		hashStr := hex.EncodeToString(hash[:])

		// Store the object
		a.Store.Write(hashStr, content)

		return &domain.TreeNode{
			Name:   name,
			Type:   "file",
			Size:   int64(len(content)),
			Object: hashStr,
		}, nil
	}

	node := &domain.TreeNode{
		Name: name,
		Type: "directory",
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		child, err := a.buildTree(e.Name(), filepath.Join(path, e.Name()))
		if err != nil {
			continue // skip inaccessible entries
		}
		node.Children = append(node.Children, *child)
	}

	return node, nil
}

// NormalizeCRLF replaces CRLF with LF.
func NormalizeCRLF(data []byte) []byte {
	return replaceAll(data, []byte("\r\n"), []byte("\n"))
}

func replaceAll(s, old, new []byte) []byte {
	return bytesReplace(s, old, new)
}

func bytesReplace(s, old, new []byte) []byte {
	// Simple implementation using bytes package
	if len(old) == 0 {
		return s
	}
	var result []byte
	for {
		idx := bytesIndex(s, old)
		if idx < 0 {
			result = append(result, s...)
			return result
		}
		result = append(result, s[:idx]...)
		result = append(result, new...)
		s = s[idx+len(old):]
	}
}

func bytesIndex(s, sep []byte) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if equal(s[i:i+len(sep)], sep) {
			return i
		}
	}
	return -1
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
