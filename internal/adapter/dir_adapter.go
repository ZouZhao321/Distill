package adapter

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

// DirAdapter adapts a filesystem directory into a TreeNode tree.
// It stores all file objects via the ObjectStorage during adaptation.
// Symlinks are skipped. Inaccessible files are skipped with a warning.
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
	if !info.IsDir() {
		return nil, os.ErrNotExist
	}
	return a.buildTree(info.Name(), dirPath)
}

func (a *DirAdapter) buildTree(name, path string) (*domain.TreeNode, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	// Skip symlinks
	if info.Mode()&os.ModeSymlink != 0 {
		slog.Warn("skipping symlink", "path", path)
		return nil, nil
	}

	if !info.IsDir() {
		content, err := os.ReadFile(path)
		if err != nil {
			slog.Warn("skipping inaccessible file", "path", path, "error", err)
			return nil, nil
		}
		if a.NormalizeCRLF {
			content = NormalizeCRLF(content)
		}
		hash := sha256.Sum256(content)
		hashStr := hex.EncodeToString(hash[:])

		// Store the object
		if err := a.Store.Write(hashStr, content); err != nil {
			slog.Warn("failed to store object", "hash", hashStr[:16], "error", err)
		}

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
		slog.Warn("cannot read directory", "path", path, "error", err)
		return node, nil
	}

	for _, e := range entries {
		child, err := a.buildTree(e.Name(), filepath.Join(path, e.Name()))
		if err != nil {
			slog.Warn("skipping entry", "name", e.Name(), "error", err)
			continue
		}
		if child != nil {
			node.Children = append(node.Children, *child)
		}
	}

	return node, nil
}

// NormalizeCRLF replaces CRLF with LF.
func NormalizeCRLF(data []byte) []byte {
	return bytesReplaceAll(data, []byte("\r\n"), []byte("\n"))
}

func bytesReplaceAll(s, old, new []byte) []byte {
	if len(old) == 0 {
		return s
	}
	var result []byte
	for {
		idx := bytesIndexOf(s, old)
		if idx < 0 {
			result = append(result, s...)
			return result
		}
		result = append(result, s[:idx]...)
		result = append(result, new...)
		s = s[idx+len(old):]
	}
}

func bytesIndexOf(s, sep []byte) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if bytesEqual(s[i:i+len(sep)], sep) {
			return i
		}
	}
	return -1
}

func bytesEqual(a, b []byte) bool {
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
