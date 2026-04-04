// Package adapter 提供将外部数据源（文件系统目录、ZIP 文件）适配为 TreeNode 树的适配器。
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

// DirAdapter 将文件系统目录适配为 TreeNode 树。
// 适配过程中将所有文件对象存储到 ObjectStorage。
// 符号链接会被跳过，不可访问的文件会记录警告后跳过。
type DirAdapter struct {
	Store         port.ObjectStorage
	NormalizeCRLF bool
}

// NewDirAdapter 创建新的 DirAdapter 实例。
func NewDirAdapter(store port.ObjectStorage, normalizeCRLF bool) *DirAdapter {
	return &DirAdapter{Store: store, NormalizeCRLF: normalizeCRLF}
}

// Adapt 读取 dirPath 指定的目录并返回 TreeNode 树。
// 所有文件内容会在适配过程中存储到 ObjectStorage。
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

// buildTree 递归构建目录树。
func (a *DirAdapter) buildTree(name, path string) (*domain.TreeNode, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	// 跳过符号链接
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

		// 存储对象
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

// NormalizeCRLF 将 CRLF 换行替换为 LF。
func NormalizeCRLF(data []byte) []byte {
	return bytesReplaceAll(data, []byte("\r\n"), []byte("\n"))
}

// bytesReplaceAll 将 s 中所有 old 替换为 new。
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

// bytesIndexOf 返回 sep 在 s 中首次出现的位置，未找到返回 -1。
func bytesIndexOf(s, sep []byte) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if bytesEqual(s[i:i+len(sep)], sep) {
			return i
		}
	}
	return -1
}

// bytesEqual 比较两个字节切片是否相等。
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
