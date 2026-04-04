package usecase

import (
	"errors"

	"github.com/ZouZhao321/distill/internal/core/domain"
)

var ErrAlreadyInitialized = errors.New("repository already initialized")

// InitUseCase handles repository initialization.
// It produces a default Config without requiring external dependencies.
type InitUseCase struct {
	initialized bool
}

// NewInitUseCase creates a new InitUseCase.
func NewInitUseCase() *InitUseCase {
	return &InitUseCase{}
}

// Execute creates a default Config with the given home and trash paths.
// Returns ErrAlreadyInitialized if called more than once.
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
