// Package port 定义领域层对外暴露的接口。
// 适配器层（adapter/infra）负责实现这些接口。
package port

import "github.com/ZouZhao321/distill/internal/core/domain"

// AssetRepository 定义清单和引用的存储操作接口。
type AssetRepository interface {
	// 清单操作
	SaveManifest(m *domain.Manifest) error
	GetManifest(hash string) (*domain.Manifest, error)
	ListManifests() ([]domain.Manifest, error)
	RemoveManifest(hash string) error

	// 引用操作
	CreateRef(ref domain.Ref) error
	GetRef(name string) (domain.Ref, error)
	ListRefs() ([]domain.Ref, error)
	DeleteRef(name string) error
	GetRefNameByHash(hash string) (string, error)
}

// ObjectStorage 定义内容寻址对象存储接口。
type ObjectStorage interface {
	Exists(hash string) (bool, error)
	Read(hash string) ([]byte, error)
	Write(hash string, data []byte) error
	Delete(hash string) error
	Walk(fn func(hash string) error) error
}
