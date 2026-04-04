package adapter

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
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

		hash := sha256.Sum256(data)
		hashStr := hex.EncodeToString(hash[:])

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
