// Package store 提供 AssetRepository 和 ObjectStorage 接口的文件系统实现。
// 清单存储为 JSON 文件，对象以 SHA-256 哈希分片存储。
package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

// ManifestStore 使用文件系统实现 port.AssetRepository 接口。
// 清单以独立 JSON 文件存储在 manifestsDir 中。
// 引用以单个 JSON 映射存储在 refsPath 中。
type ManifestStore struct {
	manifestsDir string
	refsPath     string
	mu           sync.Mutex
}

// NewManifestStore 创建新的 ManifestStore 实例。
func NewManifestStore(manifestsDir, refsPath string) *ManifestStore {
	os.MkdirAll(manifestsDir, 0755)
	return &ManifestStore{
		manifestsDir: manifestsDir,
		refsPath:     refsPath,
	}
}

// ManifestsDir 返回清单目录路径。
func (s *ManifestStore) ManifestsDir() string { return s.manifestsDir }

// RefsPath 返回引用 JSON 文件路径。
func (s *ManifestStore) RefsPath() string { return s.refsPath }

// manifestPath 根据哈希计算清单文件路径。
func (s *ManifestStore) manifestPath(hash string) string {
	return filepath.Join(s.manifestsDir, hash+"manifest.json")
}

// --- 清单操作 ---

// SaveManifest 将清单写入 JSON 文件。
func (s *ManifestStore) SaveManifest(m *domain.Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.manifestPath(m.Hash), data, 0644)
}

// GetManifest 根据哈希读取并解析清单。
func (s *ManifestStore) GetManifest(hash string) (*domain.Manifest, error) {
	data, err := os.ReadFile(s.manifestPath(hash))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	var m domain.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// ListManifests 返回仓库中的所有清单。
func (s *ManifestStore) ListManifests() ([]domain.Manifest, error) {
	entries, err := os.ReadDir(s.manifestsDir)
	if err != nil {
		return nil, err
	}

	var manifests []domain.Manifest
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.manifestsDir, e.Name()))
		if err != nil {
			continue
		}
		var m domain.Manifest
		if json.Unmarshal(data, &m) == nil {
			manifests = append(manifests, m)
		}
	}
	return manifests, nil
}

// RemoveManifest 删除清单文件。
func (s *ManifestStore) RemoveManifest(hash string) error {
	return os.Remove(s.manifestPath(hash))
}

// --- 引用操作 ---

// loadRefs 从文件加载引用映射。
func (s *ManifestStore) loadRefs() (map[string]string, error) {
	refs := make(map[string]string)
	data, err := os.ReadFile(s.refsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return refs, nil
		}
		return nil, err
	}
	json.Unmarshal(data, &refs)
	return refs, nil
}

// saveRefs 将引用映射写入文件。
func (s *ManifestStore) saveRefs(refs map[string]string) error {
	data, err := json.MarshalIndent(refs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.refsPath, data, 0644)
}

// CreateRef 注册新的资产名称到清单哈希的映射。
func (s *ManifestStore) CreateRef(ref domain.Ref) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	refs, err := s.loadRefs()
	if err != nil {
		return err
	}
	if _, exists := refs[ref.Name]; exists {
		return domain.ErrAlreadyExists
	}
	refs[ref.Name] = ref.Manifest
	return s.saveRefs(refs)
}

// GetRef 根据名称查找引用。
func (s *ManifestStore) GetRef(name string) (domain.Ref, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	refs, err := s.loadRefs()
	if err != nil {
		return domain.Ref{}, err
	}
	hash, ok := refs[name]
	if !ok {
		return domain.Ref{}, domain.ErrNotFound
	}
	return domain.Ref{Name: name, Manifest: hash}, nil
}

// ListRefs 返回所有已注册的引用。
func (s *ManifestStore) ListRefs() ([]domain.Ref, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	refs, err := s.loadRefs()
	if err != nil {
		return nil, err
	}
	var result []domain.Ref
	for name, hash := range refs {
		result = append(result, domain.Ref{Name: name, Manifest: hash})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

// DeleteRef 根据名称删除引用。
func (s *ManifestStore) DeleteRef(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	refs, err := s.loadRefs()
	if err != nil {
		return err
	}
	if _, ok := refs[name]; !ok {
		return domain.ErrNotFound
	}
	delete(refs, name)
	return s.saveRefs(refs)
}

// GetRefNameByHash 根据清单哈希反查引用名称。
func (s *ManifestStore) GetRefNameByHash(hash string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	refs, err := s.loadRefs()
	if err != nil {
		return "", err
	}
	for name, h := range refs {
		if h == hash {
			return name, nil
		}
	}
	return "", domain.ErrNotFound
}
