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

// ManifestStore implements port.AssetRepository using the filesystem.
// Manifests are stored as individual JSON files in manifestsDir.
// Refs are stored as a single JSON map in refsPath.
type ManifestStore struct {
	manifestsDir string
	refsPath     string
	mu           sync.Mutex
}

// NewManifestStore creates a new ManifestStore.
func NewManifestStore(manifestsDir, refsPath string) *ManifestStore {
	os.MkdirAll(manifestsDir, 0755)
	return &ManifestStore{
		manifestsDir: manifestsDir,
		refsPath:     refsPath,
	}
}

// ManifestsDir returns the manifests directory path.
func (s *ManifestStore) ManifestsDir() string { return s.manifestsDir }

// RefsPath returns the refs JSON file path.
func (s *ManifestStore) RefsPath() string { return s.refsPath }

func (s *ManifestStore) manifestPath(hash string) string {
	return filepath.Join(s.manifestsDir, hash+"manifest.json")
}

// --- Manifest operations ---

// SaveManifest writes the manifest as a JSON file.
func (s *ManifestStore) SaveManifest(m *domain.Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.manifestPath(m.Hash), data, 0644)
}

// GetManifest reads and parses a manifest by hash.
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

// ListManifests returns all manifests in the store.
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

// RemoveManifest deletes the manifest file.
func (s *ManifestStore) RemoveManifest(hash string) error {
	return os.Remove(s.manifestPath(hash))
}

// --- Ref operations ---

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

func (s *ManifestStore) saveRefs(refs map[string]string) error {
	data, err := json.MarshalIndent(refs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.refsPath, data, 0644)
}

// CreateRef registers a new name-to-manifest mapping.
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

// GetRef looks up a ref by name.
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

// ListRefs returns all registered refs.
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

// DeleteRef removes a ref by name.
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

// GetRefNameByHash finds a ref name by its manifest hash.
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
