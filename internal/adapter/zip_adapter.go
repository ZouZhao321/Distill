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

// ZipAdapter adapts a ZIP file into a TreeNode tree.
// It stores all file objects via the ObjectStorage during adaptation.
type ZipAdapter struct {
	Store         port.ObjectStorage
	NormalizeCRLF bool
}

// NewZipAdapter creates a new ZipAdapter.
func NewZipAdapter(store port.ObjectStorage, normalizeCRLF bool) *ZipAdapter {
	return &ZipAdapter{Store: store, NormalizeCRLF: normalizeCRLF}
}

// Adapt reads the ZIP file at zipPath and returns a TreeNode tree.
// All file contents are stored in the ObjectStorage.
// Encrypted or corrupted entries are skipped.
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
			continue // skip encrypted or corrupted
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

		// Store the object
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
