package port

import "github.com/ZouZhao321/distill/internal/core/domain"

// AssetRepository defines the interface for manifest and ref storage operations.
type AssetRepository interface {
	// Manifest operations
	SaveManifest(m *domain.Manifest) error
	GetManifest(hash string) (*domain.Manifest, error)
	ListManifests() ([]domain.Manifest, error)
	RemoveManifest(hash string) error

	// Ref operations
	CreateRef(ref domain.Ref) error
	GetRef(name string) (domain.Ref, error)
	ListRefs() ([]domain.Ref, error)
	DeleteRef(name string) error
	GetRefNameByHash(hash string) (string, error)
}

// ObjectStorage defines the interface for content-addressed object storage.
type ObjectStorage interface {
	Exists(hash string) (bool, error)
	Read(hash string) ([]byte, error)
	Write(hash string, data []byte) error
	Delete(hash string) error
	Walk(fn func(hash string) error) error
}
