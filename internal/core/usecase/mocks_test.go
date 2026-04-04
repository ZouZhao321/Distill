package usecase

import (
	"github.com/ZouZhao321/distill/internal/core/domain"
	"github.com/ZouZhao321/distill/internal/core/port"
)

// --- Mock ObjectStorage ---

type mockObjectStorage struct {
	written map[string][]byte
}

func newMockObjectStorage() *mockObjectStorage {
	return &mockObjectStorage{written: make(map[string][]byte)}
}

func (m *mockObjectStorage) Exists(hash string) (bool, error) {
	_, ok := m.written[hash]
	return ok, nil
}

func (m *mockObjectStorage) Read(hash string) ([]byte, error) {
	data, ok := m.written[hash]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return data, nil
}

func (m *mockObjectStorage) Write(hash string, data []byte) error {
	m.written[hash] = data
	return nil
}

func (m *mockObjectStorage) Delete(hash string) error {
	delete(m.written, hash)
	return nil
}

func (m *mockObjectStorage) Walk(fn func(hash string) error) error {
	for h := range m.written {
		if err := fn(h); err != nil {
			return err
		}
	}
	return nil
}

// --- Mock AssetRepository ---

type mockAssetRepo struct {
	manifests map[string]*domain.Manifest
	refs      map[string]string
}

func newMockAssetRepo() *mockAssetRepo {
	return &mockAssetRepo{
		manifests: make(map[string]*domain.Manifest),
		refs:      make(map[string]string),
	}
}

func (m *mockAssetRepo) SaveManifest(manifest *domain.Manifest) error {
	m.manifests[manifest.Hash] = manifest
	return nil
}

func (m *mockAssetRepo) GetManifest(hash string) (*domain.Manifest, error) {
	if mf, ok := m.manifests[hash]; ok {
		return mf, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockAssetRepo) ListManifests() ([]domain.Manifest, error) {
	var list []domain.Manifest
	for _, mf := range m.manifests {
		list = append(list, *mf)
	}
	return list, nil
}

func (m *mockAssetRepo) RemoveManifest(hash string) error {
	delete(m.manifests, hash)
	return nil
}

func (m *mockAssetRepo) CreateRef(ref domain.Ref) error {
	m.refs[ref.Name] = ref.Manifest
	return nil
}

func (m *mockAssetRepo) GetRef(name string) (domain.Ref, error) {
	if h, ok := m.refs[name]; ok {
		return domain.Ref{Name: name, Manifest: h}, nil
	}
	return domain.Ref{}, domain.ErrNotFound
}

func (m *mockAssetRepo) ListRefs() ([]domain.Ref, error) {
	var list []domain.Ref
	for name, hash := range m.refs {
		list = append(list, domain.Ref{Name: name, Manifest: hash})
	}
	return list, nil
}

func (m *mockAssetRepo) DeleteRef(name string) error {
	delete(m.refs, name)
	return nil
}

func (m *mockAssetRepo) GetRefNameByHash(hash string) (string, error) {
	for name, h := range m.refs {
		if h == hash {
			return name, nil
		}
	}
	return "", domain.ErrNotFound
}

// Compile-time interface checks
var _ port.ObjectStorage = (*mockObjectStorage)(nil)
var _ port.AssetRepository = (*mockAssetRepo)(nil)
