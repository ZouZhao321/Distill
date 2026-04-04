package usecase

import (
	"errors"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

// ErrAlreadyInitialized 表示仓库已经初始化过。
var ErrAlreadyInitialized = errors.New("repository already initialized")

// InitUseCase 负责仓库初始化。
// 它生成默认配置，不依赖外部资源。
type InitUseCase struct {
	initialized bool
}

// NewInitUseCase 创建新的 InitUseCase 实例。
func NewInitUseCase() *InitUseCase {
	return &InitUseCase{}
}

// Execute 使用给定的仓库路径和回收站路径创建默认配置。
// 如果已初始化过，返回 ErrAlreadyInitialized。
func (uc *InitUseCase) Execute(homePath, trashPath string) (*domain.Config, error) {
	if uc.initialized {
		return nil, ErrAlreadyInitialized
	}

	config := &domain.Config{}
	config.Core.Version = "1"
	config.Core.ObjectsFormat = "plain"
	config.Store.Home = homePath
	config.Store.TrashPath = trashPath
	config.Checkout.Overwrite = "ask"
	config.Log.Format = "text"
	config.Log.Level = "info"
	config.Normalize.CRLFToLF = true

	uc.initialized = true
	return config, nil
}
