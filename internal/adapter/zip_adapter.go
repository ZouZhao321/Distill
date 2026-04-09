package adapter

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

// ZipAdapter 将 ZIP 文件适配为 TreeNode 树。
// 适配过程中将所有文件对象存储到 ObjectStorage。
type ZipAdapter struct {
	Store         port.ObjectStorage
	NormalizeCRLF bool
}

// NewZipAdapter 创建新的 ZipAdapter 实例。
func NewZipAdapter(store port.ObjectStorage, normalizeCRLF bool) *ZipAdapter {
	return &ZipAdapter{Store: store, NormalizeCRLF: normalizeCRLF}
}

// Adapt 读取 zipPath 指定的 ZIP 文件并返回 TreeNode 树。
// 所有文件内容会在适配过程中存储到 ObjectStorage。
// 加密或损坏的条目会被跳过。
func (a *ZipAdapter) Adapt(zipPath string) (*domain.TreeNode, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	root := &domain.TreeNode{
		Name: strings.TrimSuffix(filepath.Base(zipPath), ".zip"),
		Type: "directory",
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		// 路径安全校验：拒绝路径穿越和绝对路径
		if err := validateZipEntryName(f.Name); err != nil {
			return nil, err
		}

		rc, err := f.Open()
		if err != nil {
			continue // 跳过加密或损坏的条目
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}

		if a.NormalizeCRLF {
			data = NormalizeCRLF(data)
		}

		hashStr := domain.ComputeHash(data)

		// 存储对象
		a.Store.Write(hashStr, data)

		parts := strings.Split(filepath.ToSlash(f.Name), "/")
		current := root
		for i, part := range parts {
			if i == len(parts)-1 {
				current.Children = append(current.Children, domain.TreeNode{
					Name:   part,
					Type:   "file",
					Size:   int64(len(data)),
					Object: hashStr,
				})
			} else {
				current = a.findOrCreateDir(current, part)
			}
		}
	}

	return root, nil
}

// findOrCreateDir 在父节点中查找或创建指定名称的目录节点。
func (a *ZipAdapter) findOrCreateDir(parent *domain.TreeNode, name string) *domain.TreeNode {
	for i := range parent.Children {
		if parent.Children[i].Name == name && parent.Children[i].Type == "directory" {
			return &parent.Children[i]
		}
	}
	dir := domain.TreeNode{Name: name, Type: "directory"}
	parent.Children = append(parent.Children, dir)
	return &parent.Children[len(parent.Children)-1]
}

// validateZipEntryName 校验 ZIP 条目路径安全性。
// 拒绝包含 ".."、"." 段、绝对路径以及空文件名的条目。
func validateZipEntryName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: empty zip entry name", domain.ErrInvalidPath)
	}

	// 统一使用 / 分割路径
	cleaned := filepath.ToSlash(name)

	// 拒绝绝对路径（Unix 风格 / 开头或 Windows 风格盘符开头）
	if strings.HasPrefix(cleaned, "/") || isWindowsAbsPath(cleaned) {
		return fmt.Errorf("%w: absolute path not allowed: %s", domain.ErrInvalidPath, name)
	}

	parts := strings.Split(cleaned, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("%w: unsafe path segment %q in: %s", domain.ErrInvalidPath, part, name)
		}
	}
	return nil
}

// isWindowsAbsPath 检查路径是否为 Windows 绝对路径（如 C:/... 或 \\server\share）。
func isWindowsAbsPath(path string) bool {
	if len(path) >= 3 && path[1] == ':' && (path[2] == '/' || path[2] == '\\') {
		return true
	}
	if len(path) >= 2 && path[0] == '\\' && path[1] == '\\' {
		return true
	}
	return false
}
